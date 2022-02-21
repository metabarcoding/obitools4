package obichunk

import (
	"sync"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func IUniqueSequence(iterator obiseq.IBioSequenceBatch,
	options ...WithOption) (obiseq.IBioSequenceBatch, error) {

	var err error
	opts := MakeOptions(options)
	nworkers := opts.ParallelWorkers()

	iUnique := obiseq.MakeIBioSequenceBatch(opts.BufferSize())

	if opts.SortOnDisk() {
		nworkers = 1
		iterator, err = ISequenceChunkOnDisk(iterator,
			obiseq.HashClassifier(opts.BatchCount()),
			opts.BufferSize())

		if err != nil {
			return obiseq.NilIBioSequenceBatch, err
		}

	} else {
		iterator, err = ISequenceChunk(iterator,
			obiseq.HashClassifier(opts.BatchCount()),
			opts.BufferSize())

		if err != nil {
			return obiseq.NilIBioSequenceBatch, err
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

	var ff func(obiseq.IBioSequenceBatch, *obiseq.BioSequenceClassifier, int)

	cat := opts.Categories()
	na := opts.NAValue()

	ff = func(input obiseq.IBioSequenceBatch,
		classifier *obiseq.BioSequenceClassifier,
		icat int) {
		icat--
		input, err = ISequenceSubChunk(input,
			classifier,
			1,
			opts.BufferSize())

		var next obiseq.IBioSequenceBatch
		if icat >= 0 {
			next = obiseq.MakeIBioSequenceBatch(opts.BufferSize())

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
