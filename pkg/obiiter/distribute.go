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
// based on the provided classifier. It returns an IDistribute instance
// that manages the distribution of the sequences.
//
// Batches are flushed when either BatchSizeMax() sequences or BatchMem()
// bytes are accumulated per key, mirroring the RebatchBySize strategy.
func (iterator IBioSequence) Distribute(class *obiseq.BioSequenceClassifier) IDistribute {
	maxCount := obidefault.BatchSizeMax()
	maxBytes := obidefault.BatchMem()

	outputs := make(map[int]IBioSequence, 100)
	slices := make(map[int]*obiseq.BioSequenceSlice, 100)
	bufBytes := make(map[int]int, 100)
	orders := make(map[int]int, 100)
	news := make(chan int)

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
					bufBytes[key] = 0

					lock.Lock()
					outputs[key] = MakeIBioSequence()
					lock.Unlock()

					news <- key
				}

				sz := s.MemorySize()
				countFull := maxCount > 0 && len(*slice) >= maxCount
				memFull := maxBytes > 0 && bufBytes[key]+sz > maxBytes && len(*slice) > 0
				if countFull || memFull {
					outputs[key].Push(MakeBioSequenceBatch(source, orders[key], *slice))
					orders[key]++
					s := obiseq.MakeBioSequenceSlice()
					slices[key] = &s
					slice = &s
					bufBytes[key] = 0
				}

				*slice = append(*slice, s)
				bufBytes[key] += sz
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
