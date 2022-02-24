package obichunk

import (
	"sync"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func IUniqueSequence(iterator obiiter.IBioSequenceBatch,
	options ...WithOption) (obiiter.IBioSequenceBatch, error) {

	var err error
	opts := MakeOptions(options)
	nworkers := opts.ParallelWorkers()

	iUnique := obiiter.MakeIBioSequenceBatch(opts.BufferSize())

	if opts.SortOnDisk() {
		nworkers = 1
		iterator, err = ISequenceChunkOnDisk(iterator,
			obiseq.HashClassifier(opts.BatchCount()),
			0)

		if err != nil {
			return obiiter.NilIBioSequenceBatch, err
		}

	} else {
		iterator, err = ISequenceChunk(iterator,
			obiseq.HashClassifier(opts.BatchCount()),
			opts.BufferSize())

		if err != nil {
			return obiiter.NilIBioSequenceBatch, err
		}
	}

	iUnique.Add(nworkers)

	go func() {
		iUnique.Wait()
		iUnique.Close()
	}()

	omutex := sync.Mutex{}
	order := 0

	nextOrder := func() int {
		omutex.Lock()
		neworder := order
		order++
		omutex.Unlock()
		return neworder
	}

	var ff func(obiiter.IBioSequenceBatch, *obiseq.BioSequenceClassifier, int)

	cat := opts.Categories()
	na := opts.NAValue()

	ff = func(input obiiter.IBioSequenceBatch,
		classifier *obiseq.BioSequenceClassifier,
		icat int) {
		icat--
		input, err = ISequenceSubChunk(input,
			classifier,
			1,
			opts.BufferSize())

		var next obiiter.IBioSequenceBatch
		if icat >= 0 {
			next = obiiter.MakeIBioSequenceBatch(opts.BufferSize())

			iUnique.Add(1)
			go ff(next,
				obiseq.AnnotationClassifier(cat[icat], na),
				icat)
		}

		o := 0
		for input.Next() {
			batch := input.Get()

			if icat < 0 || len(batch.Slice()) == 1 {
				iUnique.Push(batch.Reorder(nextOrder()))
			} else {
				next.Push(batch.Reorder(o))
				o++
			}
		}

		if icat >= 0 {
			next.Close()
		}

		iUnique.Done()
	}

	for i := 0; i < nworkers-1; i++ {
		go ff(iterator.Split(),
			obiseq.SequenceClassifier(),
			len(cat))
	}
	go ff(iterator,
		obiseq.SequenceClassifier(),
		len(cat))

	iMerged := iUnique.IMergeSequenceBatch(opts.NAValue(),
		opts.StatsOn(),
		opts.BufferSize(),
	)

	return iMerged.Speed(), nil
}
