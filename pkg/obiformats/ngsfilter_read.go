package obiformats

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

type PrimerPair struct {
	Forward string
	Reverse string
}

type TagPair struct {
	Forward string
	Reverse string
}

type PCR struct {
	Experiment  string
	Sample      string
	Partial     bool
	Annotations obiseq.Annotation
}

type PCRs map[TagPair]PCR
type NGSFilter map[PrimerPair]PCRs

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

func _parseMainNGSFilterTags(text string) TagPair {

	tags := strings.Split(text, ":")

	if len(tags) == 1 {
		return TagPair{tags[0], tags[0]}
	}

	if tags[0] == "-" {
		tags[0] = ""
	}

	if tags[1] == "-" {
		tags[1] = ""
	}

	return TagPair{tags[0], tags[1]}
}

func _parseMainNGSFilter(text string) (PrimerPair, TagPair, string, string, bool) {
	fields := strings.Fields(text)

	tags := _parseMainNGSFilterTags(fields[2])
	partial := fields[5] == "T" || fields[5] == "t"

	return PrimerPair{fields[3], fields[4]},
		tags,
		fields[0],
		fields[1],
		partial
}

func ReadNGSFilter(reader io.Reader) (NGSFilter, error) {
	ngsfilter := make(NGSFilter, 10)

	lines := _readLines(reader)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "#") || len(line) == 0 {
			continue
		}

		split := strings.SplitN(line, "@", 2)

		primers, tags, experiment, sample, partial := _parseMainNGSFilter(split[0])
		newPCR := PCR{
			Experiment:  experiment,
			Sample:      sample,
			Partial:     partial,
			Annotations: nil,
		}

		if len(split) > 1 && len(split[1]) > 0 {
			newPCR.Annotations = obiseq.GetAnnotation()
			ParseOBIFeatures(split[1], newPCR.Annotations)
		}

		samples, ok := ngsfilter[primers]

		if ok {
			pcr, ok := samples[tags]

			if ok {
				return nil, fmt.Errorf("pair of tags %v used for samples %s in %s and %s in %s",
					tags, sample, experiment, pcr.Sample, pcr.Experiment)
			}

			samples[tags] = newPCR
		} else {
			ngsfilter[primers] = make(PCRs, 1000)
			ngsfilter[primers][tags] = newPCR
		}
	}

	return ngsfilter, nil
}
