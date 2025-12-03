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

	cat := opts.Categories()
	na := opts.NAValue()

	var classifier *obiseq.BioSequenceClassifier

	if len(cat) > 0 {
		cls := make([]*obiseq.BioSequenceClassifier, len(cat)+1)
		for i, c := range cat {
			cls[i+1] = obiseq.AnnotationClassifier(c, na)
		}
		cls[0] = obiseq.HashClassifier(opts.BatchCount())
		classifier = obiseq.CompositeClassifier(cls...)
	} else {
		classifier = obiseq.HashClassifier(opts.BatchCount())
	}

	if opts.SortOnDisk() {
		nworkers = 1
		iterator, err = ISequenceChunkOnDisk(iterator, classifier, true, na, opts.StatsOn())

		if err != nil {
			return obiiter.NilIBioSequence, err
		}

	} else {
		iterator, err = ISequenceChunkOnMemory(iterator, classifier)

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

	ff := func(input obiiter.IBioSequence,
		classifier *obiseq.BioSequenceClassifier) {
		input, err = ISequenceSubChunk(input,
			classifier,
			1)

		for input.Next() {
			batch := input.Get()
			if !(opts.NoSingleton() && len(batch.Slice()) == 1 && batch.Slice()[0].Count() == 1) {
				iUnique.Push(batch.Reorder(nextOrder()))
			}
		}
		iUnique.Done()
	}

	for i := 0; i < nworkers-1; i++ {
		go ff(iterator.Split(), obiseq.SequenceClassifier())
	}
	go ff(iterator, obiseq.SequenceClassifier())

	iMerged := iUnique.IMergeSequenceBatch(opts.NAValue(),
		opts.StatsOn(),
	)

	return iMerged, nil
}
