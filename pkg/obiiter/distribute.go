package obiiter

import (
	"fmt"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// IDistribute represents a distribution mechanism for biosequences.
// It manages the outputs of biosequences, provides a channel for
// new data notifications, and maintains a classifier for sequence
// classification. It is designed to facilitate the distribution
// of biosequences to various processing components.
//
// Fields:
//   - outputs: A map that associates integer keys with corresponding
//     biosequence outputs (IBioSequence).
//   - news: A channel that sends notifications of new data available
//     for processing, represented by integer identifiers.
//   - classifier: A pointer to a BioSequenceClassifier used to classify
//     the biosequences during distribution.
//   - lock: A mutex for synchronizing access to the outputs and other
//     shared resources to ensure thread safety.
type IDistribute struct {
	outputs    map[int]IBioSequence
	news       chan int
	classifier *obiseq.BioSequenceClassifier
	lock       *sync.Mutex
}

func (dist *IDistribute) Outputs(key int) (IBioSequence, error) {
	dist.lock.Lock()
	iter, ok := dist.outputs[key]
	dist.lock.Unlock()

	if !ok {
		return NilIBioSequence, fmt.Errorf("code %d unknown", key)
	}

	return iter, nil
}

// News returns a channel that provides notifications of new data
// available for processing. The channel sends integer identifiers
// representing the new data.
func (dist *IDistribute) News() chan int {
	return dist.news
}

// Classifier returns a pointer to the BioSequenceClassifier
// associated with the distribution mechanism. This classifier
// is used to classify biosequences during the distribution process.
func (dist *IDistribute) Classifier() *obiseq.BioSequenceClassifier {
	return dist.classifier
}

// Distribute organizes the biosequences from the iterator into batches
// based on the provided classifier and batch sizes. It returns an
// IDistribute instance that manages the distribution of the sequences.
//
// Parameters:
//   - class: A pointer to a BioSequenceClassifier used to classify
//     the biosequences during distribution.
//   - sizes: Optional integer values specifying the batch size. If
//     no sizes are provided, a default batch size of 5000 is used.
//
// Returns:
// An IDistribute instance that contains the outputs of the
// classified biosequences, a channel for new data notifications,
// and the classifier used for distribution. The method operates
// asynchronously, processing the sequences in separate goroutines.
// It ensures that the outputs are closed and cleaned up once
// processing is complete.
func (iterator IBioSequence) Distribute(class *obiseq.BioSequenceClassifier, sizes ...int) IDistribute {
	batchsize := obidefault.BatchSize()

	outputs := make(map[int]IBioSequence, 100)
	slices := make(map[int]*obiseq.BioSequenceSlice, 100)
	orders := make(map[int]int, 100)
	news := make(chan int)

	if len(sizes) > 0 {
		batchsize = sizes[0]
	}

	jobDone := sync.WaitGroup{}
	lock := sync.Mutex{}

	jobDone.Add(1)

	go func() {
		jobDone.Wait()
		close(news)
		for _, i := range outputs {
			i.Close()
		}
	}()

	go func() {
		iterator = iterator.SortBatches()
		source := ""

		for iterator.Next() {
			seqs := iterator.Get()
			source = seqs.Source()

			for _, s := range seqs.Slice() {
				key := class.Code(s)
				slice, ok := slices[key]

				if !ok {
					s := obiseq.MakeBioSequenceSlice()
					slice = &s
					slices[key] = slice
					orders[key] = 0

					lock.Lock()
					outputs[key] = MakeIBioSequence()
					lock.Unlock()

					news <- key
				}

				*slice = append(*slice, s)

				if len(*slice) == batchsize {
					outputs[key].Push(MakeBioSequenceBatch(source, orders[key], *slice))
					orders[key]++
					s := obiseq.MakeBioSequenceSlice()
					slices[key] = &s
				}
			}
		}

		for key, slice := range slices {
			if len(*slice) > 0 {
				outputs[key].Push(MakeBioSequenceBatch(source, orders[key], *slice))
			}
		}

		jobDone.Done()

	}()

	return IDistribute{
		outputs,
		news,
		class,
		&lock}

}
