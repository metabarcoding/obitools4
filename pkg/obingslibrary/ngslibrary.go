package obingslibrary

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
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

type NGSLibrary struct {
	Matching string
	Primers  map[string]PrimerPair
	Markers  map[PrimerPair]*Marker
}

func MakeNGSLibrary() NGSLibrary {
	return NGSLibrary{
		Primers: make(map[string]PrimerPair, 10),
		Markers: make(map[PrimerPair]*Marker, 10),
	}
}

func (library *NGSLibrary) GetMarker(forward, reverse string) (*Marker, bool) {
	pair := PrimerPair{forward, reverse}
	marker, ok := (library.Markers)[pair]

	if ok {
		return marker, true
	}

	m := Marker{
		Forward_tag_length:    0,
		Reverse_tag_length:    0,
		Forward_spacer:        0,
		Reverse_spacer:        0,
		Forward_tag_delimiter: 0,
		Reverse_tag_delimiter: 0,
		samples:               make(map[TagPair]*PCR, 1000),
	}

	(library.Markers)[pair] = &m

	return &m, false
}

func (library *NGSLibrary) SetForwardTagSpacer(spacer int) {
	for _, marker := range library.Markers {
		marker.SetForwardTagSpacer(spacer)
	}
}

func (library *NGSLibrary) SetReverseTagSpacer(spacer int) {
	for _, marker := range library.Markers {
		marker.SetReverseTagSpacer(spacer)
	}
}

func (library *NGSLibrary) SetTagSpacer(spacer int) {
	library.SetForwardTagSpacer(spacer)
	library.SetReverseTagSpacer(spacer)
}

func (library *NGSLibrary) SetTagSpacerFor(primer string, spacer int) {
	primers, ok := library.Primers[primer]

	if ok {
		marker, ok := library.Markers[primers]

		if ok {
			if primer == primers.Forward {
				marker.SetForwardTagSpacer(spacer)
			} else {
				marker.SetReverseTagSpacer(spacer)
			}
		}
	}
}

func (library *NGSLibrary) SetForwardTagDelimiter(delim byte) {
	for _, marker := range library.Markers {
		marker.SetForwardTagDelimiter(delim)
	}
}

func (library *NGSLibrary) SetReverseTagDelimiter(delim byte) {
	for _, marker := range library.Markers {
		marker.SetReverseTagDelimiter(delim)
	}
}

func (library *NGSLibrary) SetTagDelimiter(delim byte) {
	library.SetForwardTagDelimiter(delim)
	library.SetReverseTagDelimiter(delim)
}

func (library *NGSLibrary) CheckTagLength() {

	for _, marker := range library.Markers {
		marker.CheckTagLength()
	}
}

func (library *NGSLibrary) CheckPrimerUnicity() error {
	for primers := range library.Markers {
		if _, ok := library.Primers[primers.Forward]; ok {
			return fmt.Errorf("forward primer %s is used more than once", primers.Forward)
		}
		if _, ok := library.Primers[primers.Reverse]; ok {
			return fmt.Errorf("reverse primer %s is used more than once", primers.Reverse)
		}
		if primers.Forward == primers.Reverse {
			return fmt.Errorf("forward and reverse primers are the same : %s", primers.Forward)
		}
		library.Primers[primers.Forward] = primers
		library.Primers[primers.Reverse] = primers
	}
	return nil
}
