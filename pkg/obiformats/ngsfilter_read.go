package obiformats

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obingslibrary"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"github.com/gabriel-vasile/mimetype"
)

func _readLines(reader io.Reader) []string {
	r := bufio.NewReader(reader)
	bytes := []byte{}
	lines := []string{}
	for {
		line, isPrefix, err := r.ReadLine()
		if err != nil {
			break
		}
		bytes = append(bytes, line...)
		if !isPrefix {
			str := strings.TrimSpace(string(bytes))
			if len(str) > 0 {
				lines = append(lines, str)
				bytes = []byte{}
			}
		}
	}
	if len(bytes) > 0 {
		lines = append(lines, string(bytes))
	}
	return lines
}

func _parseMainNGSFilterTags(text string) obingslibrary.TagPair {

	tags := strings.Split(text, ":")

	if len(tags) == 1 {
		return obingslibrary.TagPair{
			Forward: strings.ToLower(tags[0]),
			Reverse: strings.ToLower(tags[0]),
		}
	}

	if tags[0] == "-" {
		tags[0] = ""
	}

	if tags[1] == "-" {
		tags[1] = ""
	}

	return obingslibrary.TagPair{
		Forward: strings.ToLower(tags[0]),
		Reverse: strings.ToLower(tags[1]),
	}
}

func _parseMainNGSFilter(text string) (obingslibrary.PrimerPair, obingslibrary.TagPair, string, string, bool, bool) {
	fields := strings.Fields(text)

	if len(fields) != 6 {
		return obingslibrary.PrimerPair{}, obingslibrary.TagPair{}, "", "", false, false
	}

	tags := _parseMainNGSFilterTags(fields[2])
	partial := fields[5] == "T" || fields[5] == "t"

	return obingslibrary.PrimerPair{
			Forward: fields[3],
			Reverse: fields[4],
		},
		tags,
		fields[0],
		fields[1],
		partial,
		true
}

func OBIMimeNGSFilterTypeGuesser(stream io.Reader) (*mimetype.MIME, io.Reader, error) {

	// Create a buffer to store the read data
	buf := make([]byte, 1024*128)
	n, err := io.ReadFull(stream, buf)

	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, nil, err
	}

	// Detect the MIME type using the mimetype library
	mimeType := mimetype.Detect(buf[:n])
	if mimeType == nil {
		return nil, nil, err
	}

	// Create a new reader based on the read data
	newReader := io.Reader(bytes.NewReader(buf[:n]))

	if err == nil {
		newReader = io.MultiReader(newReader, stream)
	}

	return mimeType, newReader, nil
}

func ReadNGSFilter(reader io.Reader) (obingslibrary.NGSLibrary, error) {
	mimetype, newReader, err := OBIMimeNGSFilterTypeGuesser(reader)

	if err != nil {
		return nil, err
	}

	log.Infof("NGSFilter configuration mimetype: %s", mimetype.String())

	if mimetype.String() == "text/csv" {
		return ReadCSVNGSFilter(newReader)
	}

	return ReadOldNGSFilter(newReader)
}
func ReadOldNGSFilter(reader io.Reader) (obingslibrary.NGSLibrary, error) {
	ngsfilter := obingslibrary.MakeNGSLibrary()

	lines := _readLines(reader)

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		}

		split := strings.SplitN(line, "@", 2)

		if len(split) < 1 {
			return nil, fmt.Errorf("line %d : invalid format", i+1)
		}

		primers, tags, experiment, sample, partial, ok := _parseMainNGSFilter(split[0])

		if !ok {
			return nil, fmt.Errorf("line %d : invalid format : \n%s", i+1, line)
		}

		marker, _ := ngsfilter.GetMarker(primers.Forward, primers.Reverse)
		pcr, ok := marker.GetPCR(tags.Forward, tags.Reverse)

		if ok {
			return ngsfilter,
				fmt.Errorf("line %d : tag pair (%s,%s) used more than once with marker (%s,%s)",
					i, tags.Forward, tags.Reverse, primers.Forward, primers.Reverse)
		}

		pcr.Experiment = experiment
		pcr.Sample = sample
		pcr.Partial = partial

		if len(split) > 1 && len(split[1]) > 0 {
			pcr.Annotations = make(obiseq.Annotation)
			ParseOBIFeatures(split[1], pcr.Annotations)
		}

	}

	return ngsfilter, nil
}

func ReadCSVNGSFilter(reader io.Reader) (obingslibrary.NGSLibrary, error) {
	ngsfilter := obingslibrary.MakeNGSLibrary()
	file := csv.NewReader(reader)

	file.Comma = ','
	file.Comment = '#'
	file.TrimLeadingSpace = true
	file.ReuseRecord = true

	records, err := file.ReadAll()

	if err != nil {
		return nil, err
	}

	log.Info("Read ", len(records), " records")
	log.Infof("First record: %s", records[0])

	header := records[0]
	data := records[1:]

	// Find the index of the column named "sample"
	experimentColIndex := -1
	sampleColIndex := -1
	sample_tagColIndex := -1
	forward_primerColIndex := -1
	reverse_primerColIndex := -1

	extraColumns := make([]int, 0)

	for i, colName := range header {
		switch colName {
		case "experiment":
			experimentColIndex = i
		case "sample":
			sampleColIndex = i
		case "sample_tag":
			sample_tagColIndex = i
		case "forward_primer":
			forward_primerColIndex = i
		case "reverse_primer":
			reverse_primerColIndex = i
		default:
			extraColumns = append(extraColumns, i)
		}
	}

	if experimentColIndex == -1 {
		return nil, fmt.Errorf("column 'experiment' not found in the CSV file")
	}

	if sampleColIndex == -1 {
		return nil, fmt.Errorf("column 'sample' not found in the CSV file")
	}

	if sample_tagColIndex == -1 {
		return nil, fmt.Errorf("column 'sample_tag' not found in the CSV file")
	}

	if forward_primerColIndex == -1 {
		return nil, fmt.Errorf("column 'forward_primer' not found in the CSV file")
	}

	if reverse_primerColIndex == -1 {
		return nil, fmt.Errorf("column 'reverse_primer' not found in the CSV file")
	}

	for i, fields := range data {
		if len(fields) != len(header) {
			return nil, fmt.Errorf("row %d has %d columns, expected %d", len(data), len(fields), len(header))
		}

		forward_primer := fields[forward_primerColIndex]
		reverse_primer := fields[reverse_primerColIndex]
		tags := _parseMainNGSFilterTags(fields[sample_tagColIndex])

		marker, _ := ngsfilter.GetMarker(forward_primer, reverse_primer)
		pcr, ok := marker.GetPCR(tags.Forward, tags.Reverse)

		if ok {
			return ngsfilter,
				fmt.Errorf("line %d : tag pair (%s,%s) used more than once with marker (%s,%s)",
					i, tags.Forward, tags.Reverse, forward_primer, reverse_primer)
		}

		pcr.Experiment = fields[experimentColIndex]
		pcr.Sample = fields[sampleColIndex]
		pcr.Partial = false

		if extraColumns != nil {
			pcr.Annotations = make(obiseq.Annotation)
			for _, colIndex := range extraColumns {
				pcr.Annotations[header[colIndex]] = fields[colIndex]
			}
		}

	}

	return ngsfilter, nil
}
