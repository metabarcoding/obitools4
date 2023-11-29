package obingslibrary

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
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

type Marker struct {
	forward   obiapat.ApatPattern
	cforward  obiapat.ApatPattern
	reverse   obiapat.ApatPattern
	creverse  obiapat.ApatPattern
	taglength int
	samples   map[TagPair]*PCR
}
type NGSLibrary map[PrimerPair]*Marker

func MakeNGSLibrary() NGSLibrary {
	return make(NGSLibrary, 10)
}

func (library *NGSLibrary) GetMarker(forward, reverse string) (*Marker, bool) {
	pair := PrimerPair{forward, reverse}
	marker, ok := (*library)[pair]

	if ok {
		return marker, true
	}

	m := Marker{samples: make(map[TagPair]*PCR, 1000)}
	(*library)[pair] = &m

	return &m, false
}

func (marker *Marker) GetPCR(forward, reverse string) (*PCR, bool) {
	pair := TagPair{forward, reverse}
	pcr, ok := marker.samples[pair]

	if ok {
		return pcr, ok
	}

	ipcr := PCR{}
	marker.samples[pair] = &ipcr

	return &ipcr, false
}
