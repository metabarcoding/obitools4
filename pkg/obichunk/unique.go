package obichunk

import (
	"sync"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// Runs dereplication algorithm on a  obiiter.IBioSequenceBatch
// iterator.

func IUniqueSequence(iterator obiiter.IBioSequence,
	options ...WithOption) (obiiter.IBioSequence, error) {

	var err error
	opts := MakeOptions(options)
	nworkers := opts.ParallelWorkers()

	iUnique := obiiter.MakeIBioSequence()

	iterator = iterator.Speed("Splitting data set")

	log.Infoln("Starting data splitting")

	if opts.SortOnDisk() {
		nworkers = 1
		iterator, err = ISequenceChunkOnDisk(iterator,
			obiseq.HashClassifier(opts.BatchCount()))

		if err != nil {
			return obiiter.NilIBioSequence, err
		}

	} else {
		iterator, err = ISequenceChunkOnMemory(iterator,
			obiseq.HashClassifier(opts.BatchCount()))

		if err != nil {
			return obiiter.NilIBioSequence, err
		}
	}

	log.Infoln("End of the data splitting")

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

	var ff func(obiiter.IBioSequence,
		*obiseq.BioSequenceClassifier,
		int)

	cat := opts.Categories()
	na := opts.NAValue()

	ff = func(input obiiter.IBioSequence,
		classifier *obiseq.BioSequenceClassifier,
		icat int) {
		icat--
		input, err = ISequenceSubChunk(input,
			classifier,
			1)

		var next obiiter.IBioSequence
		if icat >= 0 {
			next = obiiter.MakeIBioSequence()

			iUnique.Add(1)

			go ff(next,
				obiseq.AnnotationClassifier(cat[icat], na),
				icat)
		}

		o := 0
		for input.Next() {
			batch := input.Get()

			if icat < 0 || len(batch.Slice()) == 1 {
				// No more sub classification of sequence or only a single sequence
				if !(opts.NoSingleton() && len(batch.Slice()) == 1 && batch.Slice()[0].Count() == 1) {
					iUnique.Push(batch.Reorder(nextOrder()))
				}
			} else {
				// A new step of classification must du realized
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
	)

	return iMerged, nil
}
