package obiiter

import (
	"sync"
	"sync/atomic"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
)

func (iter IBioSequence) NumberSequences(start int, forceReordering bool) IBioSequence {

	next_first := &atomic.Int64{}
	next_first.Store(int64(start))
	lock := &sync.Mutex{}

	w := obidefault.ParallelWorkers()
	if forceReordering {
		iter = iter.SortBatches()
		w = 1
	}

	newIter := MakeIBioSequence()
	newIter.Add(w)

	is_paired := false

	if iter.IsPaired() {
		is_paired = true
		newIter.MarkAsPaired()
	}

	number := func(iter IBioSequence) {
		for iter.Next() {
			batch := iter.Get()
			seqs := batch.Slice()
			lock.Lock()
			start := int(next_first.Load())
			next_first.Store(int64(start + len(seqs)))
			lock.Unlock()
			for i, seq := range seqs {
				num := start + i
				seq.SetAttribute("seq_number", num)
				if is_paired {
					seq.PairedWith().SetAttribute("seq_number", num)
				}
			}
			newIter.Push(batch)
		}
		newIter.Done()
	}

	go func() {
		newIter.WaitAndClose()
	}()

	for i := 1; i < w; i++ {
		go number(iter.Split())
	}
	go number(iter)

	return newIter
}
