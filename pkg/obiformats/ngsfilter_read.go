package obiformats

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"errors"

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

func _parseMainNGSFilter(text string) (obingslibrary.PrimerPair, obingslibrary.TagPair, string, string, bool) {
	fields := strings.Fields(text)

	if len(fields) != 6 {
		return obingslibrary.PrimerPair{}, obingslibrary.TagPair{}, "", "", false
	}

	tags := _parseMainNGSFilterTags(fields[2])

	return obingslibrary.PrimerPair{
			Forward: strings.ToLower(fields[3]),
			Reverse: strings.ToLower(fields[4]),
		},
		tags,
		fields[0],
		fields[1],
		true
}

func NGSFilterCsvDetector(raw []byte, limit uint32) bool {
	r := csv.NewReader(bytes.NewReader(dropLastLine(raw, limit)))
	r.Comma = ','
	r.ReuseRecord = true
	r.LazyQuotes = true
	r.FieldsPerRecord = -1
	r.Comment = '#'

	nfields := 0

	lines := 0
	for {
		rec, err := r.Read()
		if len(rec) > 0 && rec[0] == "@param" {
			continue
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return false
		}

		if nfields == 0 {
			nfields = len(rec)
		} else if nfields != len(rec) {
			return false
		}
		lines++
	}

	return nfields > 1 && lines > 1

}

func dropLastLine(b []byte, readLimit uint32) []byte {
	if readLimit == 0 || uint32(len(b)) < readLimit {
		return b
	}
	for i := len(b) - 1; i > 0; i-- {
		if b[i] == '\n' {
			return b[:i]
		}
	}
	return b
}

func OBIMimeNGSFilterTypeGuesser(stream io.Reader) (*mimetype.MIME, io.Reader, error) {

	// Create a buffer to store the read data
	buf := make([]byte, 1024*128)
	n, err := io.ReadFull(stream, buf)

	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, nil, err
	}

	mimetype.Lookup("text/plain").Extend(NGSFilterCsvDetector, "text/ngsfilter-csv", ".csv")

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

func ReadNGSFilter(reader io.Reader) (*obingslibrary.NGSLibrary, error) {
	var ngsfilter *obingslibrary.NGSLibrary
	var err error
	var mimetype *mimetype.MIME
	var newReader io.Reader

	mimetype, newReader, err = OBIMimeNGSFilterTypeGuesser(reader)

	if err != nil {
		return nil, err
	}

	log.Infof("NGSFilter configuration mimetype: %s", mimetype.String())

	if mimetype.String() == "text/ngsfilter-csv" || mimetype.String() == "text/csv" {
		ngsfilter, err = ReadCSVNGSFilter(newReader)
	} else {
		ngsfilter, err = ReadOldNGSFilter(newReader)
	}

	if err != nil {
		return nil, err
	}

	ngsfilter.CheckPrimerUnicity()
	ngsfilter.CheckTagLength()

	return ngsfilter, nil
}

func ReadOldNGSFilter(reader io.Reader) (*obingslibrary.NGSLibrary, error) {
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

		primers, tags, experiment, sample, ok := _parseMainNGSFilter(split[0])

		if !ok {
			return nil, fmt.Errorf("line %d : invalid format : \n%s", i+1, line)
		}

		marker, _ := ngsfilter.GetMarker(primers.Forward, primers.Reverse)
		pcr, ok := marker.GetPCR(tags.Forward, tags.Reverse)

		if ok {
			return &ngsfilter,
				fmt.Errorf("line %d : tag pair (%s,%s) used more than once with marker (%s,%s)",
					i, tags.Forward, tags.Reverse, primers.Forward, primers.Reverse)
		}

		pcr.Experiment = experiment
		pcr.Sample = sample

		if len(split) > 1 && len(split[1]) > 0 {
			pcr.Annotations = make(obiseq.Annotation)
			ParseOBIFeatures(split[1], pcr.Annotations)
		}

	}

	return &ngsfilter, nil
}

var library_parameter = map[string]func(library *obingslibrary.NGSLibrary, values ...string){
	"spacer": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @spacer parameter")
		case 1:
			spacer, err := strconv.Atoi(values[0])

			if err != nil {
				log.Fatalln("Invalid value for @spacer parameter")
			}

			log.Infof("Set global spacer to %d bp", spacer)
			library.SetTagSpacer(spacer)
		case 2:
			primer := values[0]
			spacer, err := strconv.Atoi(values[1])

			if err != nil {
				log.Fatalln("Invalid value for @spacer parameter")
			}

			log.Infof("Set spacer for primer %s to %d bp", primer, spacer)
			library.SetTagSpacerFor(primer, spacer)
		default:
			log.Fatalln("Invalid value for @spacer parameter")
		}
	},
	"forward_spacer": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @forward_spacer parameter")
		case 1:
			spacer, err := strconv.Atoi(values[0])

			if err != nil {
				log.Fatalln("Invalid value for @forward_spacer parameter")
			}

			log.Infof("Set spacer for forward primer to %d bp", spacer)
			library.SetForwardTagSpacer(spacer)
		default:
			log.Fatalln("Invalid value for @forward_spacer parameter")
		}
	},
	"reverse_spacer": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @reverse_spacer parameter")
		case 1:
			spacer, err := strconv.Atoi(values[0])

			if err != nil {
				log.Fatalln("Invalid value for @reverse_spacer parameter")
			}

			log.Infof("Set spacer for reverse primer to %d bp", spacer)
			library.SetReverseTagSpacer(spacer)
		default:
			log.Fatalln("Invalid value for @reverse_spacer parameter")
		}
	},
	"tag_delimiter": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @tag_delimiter parameter")
		case 1:
			value := []byte(values[0])[0]

			log.Infof("Set global tag delimiter to %c", value)
			library.SetTagDelimiter(value)
		case 2:
			value := []byte(values[1])[0]

			log.Infof("Set tag delimiter for primer %s to %c", values[0], value)
			library.SetTagDelimiterFor(values[0], value)
		default:
			log.Fatalln("Invalid value for @tag_delimiter parameter")
		}
	},
	"forward_tag_delimiter": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @forward_tag_delimiter parameter")
		case 1:
			value := []byte(values[0])[0]

			log.Infof("Set tag delimiter for forward primer to %c", value)
			library.SetForwardTagDelimiter(value)
		default:
			log.Fatalln("Invalid value for @forward_tag_delimiter parameter")
		}
	},
	"reverse_tag_delimiter": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @reverse_tag_delimiter parameter")
		case 1:
			value := []byte(values[0])[0]

			log.Infof("Set tag delimiter for reverse primer to %c", value)
			library.SetReverseTagDelimiter(value)
		default:
			log.Fatalln("Invalid value for @reverse_tag_delimiter parameter")
		}
	},
	"matching": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @matching parameter")
		case 1:
			if err := library.SetMatching(values[0]); err != nil {
				log.Fatalf("Invalid value %s for @matching parameter", values[0])
			}
			log.Infof("Set tag matching mode to %s", values[0])
		default:
			log.Fatalln("Invalid value for @matching parameter")
		}
	},
	"primer_mismatches": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @primer_error parameter")
		case 1:
			dist, err := strconv.Atoi(values[0])

			if err != nil {
				log.Fatalf("Invalid value %s for @primer_error parameter", values[0])
			}

			log.Infof("Set global allowed primer mismatches to %d", dist)
			library.SetAllowedMismatches(dist)
		case 2:
			primer := values[0]
			dist, err := strconv.Atoi(values[1])

			if err != nil {
				log.Fatalf("Invalid value %s for @primer_error parameter", values[1])
			}

			log.Infof("Set allowed primer mismatches for primer %s to %d", primer, dist)
			library.SetAllowedMismatchesFor(primer, dist)
		default:
			log.Fatalln("Invalid value for @primer_error parameter")
		}
	},
	"forward_mismatches": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @forward_primer_error parameter")
		case 1:
			dist, err := strconv.Atoi(values[0])

			if err != nil {
				log.Fatalf("Invalid value %s for @forward_primer_error parameter", values[0])
			}

			log.Infof("Set allowed mismatches for forward primer to %d", dist)
			library.SetForwardAllowedMismatches(dist)
		default:
			log.Fatalln("Invalid value for @forward_primer_error parameter")
		}
	},
	"reverse_mismatches": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @reverse_primer_error parameter")
		case 1:
			dist, err := strconv.Atoi(values[0])

			if err != nil {
				log.Fatalf("Invalid value %s for @reverse_primer_error parameter", values[0])
			}

			log.Infof("Set allowed mismatches for reverse primer to %d", dist)
			library.SetReverseAllowedMismatches(dist)
		default:
			log.Fatalln("Invalid value for @reverse_primer_error parameter")
		}
	},
	"tag_indels": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @tag_indels parameter")
		case 1:
			indels, err := strconv.Atoi(values[0])

			if err != nil {
				log.Fatalf("Invalid value %s for @tag_indels parameter", values[0])
			}

			log.Infof("Set global maximum tag indels to %d", indels)
			library.SetTagIndels(indels)
		case 2:
			indels, err := strconv.Atoi(values[1])

			if err != nil {
				log.Fatalf("Invalid value %s for @tag_indels parameter", values[1])
			}

			log.Infof("Set maximum tag indels for primer %s to %d", values[0], indels)
			library.SetTagIndelsFor(values[0], indels)
		}
	},

	"forward_tag_indels": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @forward_tag_indels parameter")
		case 1:
			indels, err := strconv.Atoi(values[0])

			if err != nil {
				log.Fatalf("Invalid value %s for @forward_tag_indels parameter", values[0])
			}

			log.Infof("Set maximum tag indels for forward primer to %d", indels)
			library.SetForwardTagIndels(indels)
		default:
			log.Fatalln("Invalid value for @forward_tag_indels parameter")
		}
	},
	"reverse_tag_indels": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @reverse_tag_indels parameter")
		case 1:
			indels, err := strconv.Atoi(values[0])

			if err != nil {
				log.Fatalf("Invalid value %s for @reverse_tag_indels parameter", values[0])
			}

			log.Infof("Set maximum tag indels for reverse primer to %d", indels)
			library.SetReverseTagIndels(indels)
		default:
			log.Fatalln("Invalid value for @reverse_tag_indels parameter")
		}
	},
	"indels": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @indels parameter")
		case 1:

			if values[0] == "true" {
				log.Info("Allows indels for primer matching")
			} else {
				log.Info("Disallows indels for primer matching")
			}

			library.SetAllowsIndels(values[0] == "true")
		case 2:

			if values[1] == "true" {
				log.Infof("Allows indels for primer matching %s", values[0])
			} else {
				log.Infof("Disallows indels for primer matching %s", values[0])
			}

			library.SetAllowsIndelsFor(values[0], values[1] == "true")
		default:
			log.Fatalln("Invalid value for @indels parameter")
		}
	},

	"forward_indels": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @forward_indels parameter")
		case 1:
			if values[0] == "true" {
				log.Info("Allows indels for forward primer matching")
			} else {
				log.Info("Disallows indels for forward primer matching")
			}

			library.SetForwardAllowsIndels(values[0] == "true")
		default:
			log.Fatalln("Invalid value for @forward_indels parameter")
		}
	},
	"reverse_indels": func(library *obingslibrary.NGSLibrary, values ...string) {
		switch len(values) {
		case 0:
			log.Fatalln("Missing value for @reverse_indels parameter")
		case 1:
			if values[0] == "true" {
				log.Info("Allows indels for reverse primer matching")
			} else {
				log.Info("Disallows indels for reverse primer matching")
			}
			library.SetReverseAllowsIndels(values[0] == "true")
		default:
			log.Fatalln("Invalid value for @reverse_indels parameter")
		}
	},
}

func ReadCSVNGSFilter(reader io.Reader) (*obingslibrary.NGSLibrary, error) {
	ngsfilter := obingslibrary.MakeNGSLibrary()
	file := csv.NewReader(reader)

	file.Comma = ','
	file.ReuseRecord = true
	file.LazyQuotes = true
	file.Comment = '#'
	file.FieldsPerRecord = -1
	file.TrimLeadingSpace = true

	records, err := file.ReadAll()

	if err != nil {
		return nil, err
	}

	i := 0
	for i = 0; i < len(records) && records[i][0] == "@param"; i++ {
	}

	params := records[0:i]
	records = records[i:]

	header := records[0]
	data := records[1:]

	log.Info("Read ", len(records), " records")
	log.Infof("First record: %s", header)

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
			return &ngsfilter,
				fmt.Errorf("line %d : tag pair (%s,%s) used more than once with marker (%s,%s)",
					i, tags.Forward, tags.Reverse, forward_primer, reverse_primer)
		}

		pcr.Experiment = fields[experimentColIndex]
		pcr.Sample = fields[sampleColIndex]

		if extraColumns != nil {
			pcr.Annotations = make(obiseq.Annotation)
			for _, colIndex := range extraColumns {
				pcr.Annotations[header[colIndex]] = fields[colIndex]
			}
		}

	}

	ngsfilter.CheckPrimerUnicity()

	for i := 0; i < len(params); i++ {
		param := params[i][1]
		if len(params[i]) < 3 {
			log.Fatalf("At line %d: Missing value for parameter %s", i, param)
		}
		data := params[i][2:]
		setparam, ok := library_parameter[param]

		if ok {
			setparam(&ngsfilter, data...)
		} else {
			log.Warnf("At line %d: Skipping unknown parameter %s: %v", i, param, data)
		}
	}

	return &ngsfilter, nil
}
