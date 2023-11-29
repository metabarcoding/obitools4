package obiiter

import "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"

func (iterator IBioSequence) IMergeSequenceBatch(na string, statsOn []string, sizes ...int) IBioSequence {
	batchsize := 100

	if len(sizes) > 0 {
		batchsize = sizes[0]
	}

	newIter := MakeIBioSequence()

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
			if batch.Len() > 0 {
				newIter.Push(batch)
			}
		}
		newIter.Done()
	}()

	return newIter
}

func MergePipe(na string, statsOn []string, sizes ...int) Pipeable {
	f := func(iterator IBioSequence) IBioSequence {
		return iterator.IMergeSequenceBatch(na, statsOn, sizes...)
	}

	return f
}
