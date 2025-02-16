package obiseq

import log "github.com/sirupsen/logrus"

func (s *BioSequence) IsPaired() bool {
	return s.paired != nil
}

func (s *BioSequence) PairedWith() *BioSequence {
	return s.paired
}

func (s *BioSequence) PairTo(p *BioSequence) {
	s.paired = p
	if p != nil {
		p.paired = s
	}
}

func (s *BioSequence) UnPair() {
	if s.paired != nil {
		s.paired.paired = nil
	}
	s.paired = nil
}

func (s *BioSequenceSlice) IsPaired() bool {
	return s != nil && s.Len() > 0 && (*s)[0].paired != nil
}

func (s *BioSequenceSlice) PairedWith() *BioSequenceSlice {
	ps := NewBioSequenceSlice(len(*s))
	for i, seq := range *s {
		(*ps)[i] = seq.PairedWith()
	}

	return ps
}

func (s *BioSequenceSlice) PairTo(p *BioSequenceSlice) {

	if len(*s) != len(*p) {
		log.Fatalf("Pairing of iterators: both batches have not the same length : (%d,%d)",
			len(*s), len(*p),
		)
	}

	for i, seq := range *s {
		seq.PairTo((*p)[i])
	}

}

func (s *BioSequenceSlice) UnPair() {
	for _, seq := range *s {
		seq.UnPair()
	}
}
