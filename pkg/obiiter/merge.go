package obiiter

import "git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"

func (iterator IBioSequenceBatch) IMergeSequenceBatch(na string, statsOn []string, sizes ...int) IBioSequenceBatch {
	batchsize := 100
	buffsize := iterator.BufferSize()

	if len(sizes) > 0 {
		batchsize = sizes[0]
	}
	if len(sizes) > 1 {
		buffsize = sizes[1]
	}

	newIter := MakeIBioSequenceBatch(buffsize)

	newIter.Add(1)

	go func() {
		newIter.WaitAndClose()
	}()

	go func() {
		for j := 0; !iterator.Finished(); j++ {
			batch := BioSequenceBatch{
				slice: obiseq.MakeBioSequenceSlice(),
				order: j}
			for i := 0; i < batchsize && iterator.Next(); i++ {
				seqs := iterator.Get()
				batch.slice = append(batch.slice, seqs.slice.Merge(na, statsOn))
			}
			if batch.Length() > 0 {
				newIter.Push(batch)
			}
		}
		newIter.Done()
	}()

	return newIter
}


func MergePipe(na string, statsOn []string, sizes ...int) Pipeable {
	f := func(iterator IBioSequenceBatch) IBioSequenceBatch {
		return iterator.IMergeSequenceBatch(na,statsOn,sizes...)
	}

	return f
}