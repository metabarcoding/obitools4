package obiseq

import (
	"sync"
)

var __bioseq__pool__ = sync.Pool{
	New: func() interface{} {
		var bs __sequence__
		bs.annotations = make(Annotation, 50)
		return &bs
	},
}

func MakeEmptyBioSequence() BioSequence {
	bs := BioSequence{__bioseq__pool__.Get().(*__sequence__)}
	bs.Reset()
	return bs
}

func MakeBioSequence(id string,
	sequence []byte,
	definition string) BioSequence {
	bs := MakeEmptyBioSequence()
	bs.SetId(id)
	bs.SetSequence(sequence)
	bs.SetDefinition(definition)
	return bs
}

func (sequence *BioSequence) Destroy() {
	__bioseq__pool__.Put(sequence.sequence)
	sequence.sequence = nil
}
