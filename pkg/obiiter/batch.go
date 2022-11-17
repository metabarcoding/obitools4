package obiiter

import "git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"

type BioSequenceBatch struct {
	slice obiseq.BioSequenceSlice
	order int
}

var NilBioSequenceBatch = BioSequenceBatch{nil, -1}

func MakeBioSequenceBatch(order int,
	sequences obiseq.BioSequenceSlice) BioSequenceBatch {

	return BioSequenceBatch{
		slice: sequences,
		order: order,
	}
}

func (batch BioSequenceBatch) Order() int {
	return batch.order
}

func (batch BioSequenceBatch) Reorder(newOrder int) BioSequenceBatch {
	batch.order = newOrder
	return batch
}

func (batch BioSequenceBatch) Slice() obiseq.BioSequenceSlice {
	return batch.slice
}

func (batch BioSequenceBatch) Len() int {
	return len(batch.slice)
}

func (batch BioSequenceBatch) NotEmpty() bool {
	return batch.slice.NotEmpty()
}

func (batch BioSequenceBatch) Pop0() *obiseq.BioSequence {
	return batch.slice.Pop0()
}

func (batch BioSequenceBatch) IsNil() bool {
	return batch.slice == nil
}

func (batch BioSequenceBatch) Recycle() {
	batch.slice.Recycle()
	batch.slice = nil
}
