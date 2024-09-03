package obingslibrary

import (
	"fmt"
	"strings"

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
	Annotations obiseq.Annotation
}

type NGSLibrary struct {
	Primers map[string]PrimerPair
	Markers map[PrimerPair]*Marker
}

func MakeNGSLibrary() NGSLibrary {
	return NGSLibrary{
		Primers: make(map[string]PrimerPair, 10),
		Markers: make(map[PrimerPair]*Marker, 10),
	}
}

func (library *NGSLibrary) GetMarker(forward, reverse string) (*Marker, bool) {
	forward = strings.ToLower(forward)
	reverse = strings.ToLower(reverse)
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
		Forward_error:         2,
		Reverse_error:         2,
		Forward_matching:      "strict",
		Reverse_matching:      "strict",
		Forward_allows_indels: false,
		Reverse_allows_indels: false,
		Forward_tag_indels:    0,
		Reverse_tag_indels:    0,
		samples:               make(map[TagPair]*PCR, 1000),
	}

	(library.Markers)[pair] = &m

	return &m, false
}

func (library *NGSLibrary) Compile(maxError int, allowsIndel bool) error {
	for primers, marker := range library.Markers {
		err := marker.Compile(primers.Forward,
			primers.Reverse,
			maxError, allowsIndel)
		if err != nil {
			return err
		}
	}
	return nil
}

func (library *NGSLibrary) Compile2() error {
	for primers, marker := range library.Markers {
		err := marker.Compile2(primers.Forward,
			primers.Reverse)
		if err != nil {
			return err
		}
	}
	return nil
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
	primer = strings.ToLower(primer)
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

func (library *NGSLibrary) SetForwardTagIndels(indels int) {
	for _, marker := range library.Markers {
		marker.SetForwardTagIndels(indels)
	}
}

func (library *NGSLibrary) SetReverseTagIndels(indels int) {
	for _, marker := range library.Markers {
		marker.SetReverseTagIndels(indels)
	}
}

func (library *NGSLibrary) SetTagIndels(indels int) {
	library.SetForwardTagIndels(indels)
	library.SetReverseTagIndels(indels)
}

func (library *NGSLibrary) SetTagIndelsFor(primer string, indels int) {
	primer = strings.ToLower(primer)
	primers, ok := library.Primers[primer]

	if ok {
		marker, ok := library.Markers[primers]

		if ok {
			if primer == primers.Forward {
				marker.SetForwardTagIndels(indels)
			} else {
				marker.SetReverseTagIndels(indels)
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

func (library *NGSLibrary) SetTagDelimiterFor(primer string, delim byte) {
	primer = strings.ToLower(primer)
	primers, ok := library.Primers[primer]

	if ok {
		marker, ok := library.Markers[primers]

		if ok {
			if primer == primers.Forward {
				marker.SetForwardTagDelimiter(delim)
			} else {
				marker.SetReverseTagDelimiter(delim)
			}
		}
	}
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

func (library *NGSLibrary) SetForwardMatching(matching string) error {

	for _, marker := range library.Markers {
		err := marker.SetForwardMatching(matching)
		if err != nil {
			return err
		}
	}

	return nil
}

func (library *NGSLibrary) SetReverseMatching(matching string) error {
	for _, marker := range library.Markers {
		err := marker.SetReverseMatching(matching)
		if err != nil {
			return err
		}
	}
	return nil
}

func (library *NGSLibrary) SetMatching(matching string) error {
	err := library.SetForwardMatching(matching)

	if err != nil {
		return err
	}

	err = library.SetReverseMatching(matching)

	return err
}

func (library *NGSLibrary) SetMatchingFor(primer string, matching string) error {
	primer = strings.ToLower(primer)
	primers, ok := library.Primers[primer]

	if ok {
		marker, ok := library.Markers[primers]

		if ok {
			if primer == primers.Forward {
				err := marker.SetForwardMatching(matching)
				if err != nil {
					return err
				}
			} else {
				err := marker.SetReverseMatching(matching)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// SetAllowsIndels sets whether the library allows indels.
// The value of the argument allows_indels is directly assigned to the library's Allows_indels field.
func (library *NGSLibrary) SetForwardAllowsIndels(allows_indels bool) {
	for _, marker := range library.Markers {
		marker.SetForwardAllowsIndels(allows_indels)
	}
}

func (library *NGSLibrary) SetReverseAllowsIndels(allows_indels bool) {
	for _, marker := range library.Markers {
		marker.SetReverseAllowsIndels(allows_indels)
	}
}

func (library *NGSLibrary) SetAllowsIndels(allows_indels bool) {
	library.SetForwardAllowsIndels(allows_indels)
	library.SetReverseAllowsIndels(allows_indels)
}

func (library *NGSLibrary) SetAllowsIndelsFor(primer string, allows_indel bool) {
	primer = strings.ToLower(primer)
	primers, ok := library.Primers[primer]

	if ok {
		marker, ok := library.Markers[primers]

		if ok {
			if primer == primers.Forward {
				marker.SetForwardAllowsIndels(allows_indel)
			} else {
				marker.SetReverseAllowsIndels(allows_indel)
			}
		}
	}
}

func (library *NGSLibrary) SetForwardAllowedMismatches(allowed_mismatches int) {
	for _, marker := range library.Markers {
		marker.SetForwardAllowedMismatches(allowed_mismatches)
	}
}

func (library *NGSLibrary) SetReverseAllowedMismatches(allowed_mismatches int) {
	for _, marker := range library.Markers {
		marker.SetReverseAllowedMismatches(allowed_mismatches)
	}
}

func (library *NGSLibrary) SetAllowedMismatches(allowed_mismatches int) {
	library.SetForwardAllowedMismatches(allowed_mismatches)
	library.SetReverseAllowedMismatches(allowed_mismatches)
}

func (library *NGSLibrary) SetAllowedMismatchesFor(primer string, allowed_mismatches int) {
	primer = strings.ToLower(primer)
	primers, ok := library.Primers[primer]

	if ok {
		marker, ok := library.Markers[primers]

		if ok {
			if primer == primers.Forward {
				marker.SetForwardAllowedMismatches(allowed_mismatches)
			} else {
				marker.SetReverseAllowedMismatches(allowed_mismatches)
			}
		}
	}
}
