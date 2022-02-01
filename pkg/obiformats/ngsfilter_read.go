package obiformats

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obingslibrary"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
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
			Forward: tags[0],
			Reverse: tags[0],
		}
	}

	if tags[0] == "-" {
		tags[0] = ""
	}

	if tags[1] == "-" {
		tags[1] = ""
	}

	return obingslibrary.TagPair{
		Forward: tags[0],
		Reverse: tags[1],
	}
}

func _parseMainNGSFilter(text string) (obingslibrary.PrimerPair, obingslibrary.TagPair, string, string, bool) {
	fields := strings.Fields(text)

	tags := _parseMainNGSFilterTags(fields[2])
	partial := fields[5] == "T" || fields[5] == "t"

	return obingslibrary.PrimerPair{
			Forward: fields[3],
			Reverse: fields[4],
		},
		tags,
		fields[0],
		fields[1],
		partial
}

func ReadNGSFilter(reader io.Reader) (obingslibrary.NGSLibrary, error) {
	ngsfilter := obingslibrary.MakeNGSLibrary()

	lines := _readLines(reader)

	for i, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		}

		split := strings.SplitN(line, "@", 2)

		primers, tags, experiment, sample, partial := _parseMainNGSFilter(split[0])

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
