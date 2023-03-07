package obiiter

import (
	"fmt"
	"sync"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

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

func (dist *IDistribute) News() chan int {
	return dist.news
}

func (dist *IDistribute) Classifier() *obiseq.BioSequenceClassifier {
	return dist.classifier
}

func (iterator IBioSequence) Distribute(class *obiseq.BioSequenceClassifier, sizes ...int) IDistribute {
	batchsize := 5000

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

		for iterator.Next() {
			seqs := iterator.Get()
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
					outputs[key].Push(MakeBioSequenceBatch(orders[key], *slice))
					orders[key]++
					s := obiseq.MakeBioSequenceSlice()
					slices[key] = &s
				}
			}
			seqs.Recycle()
		}

		for key, slice := range slices {
			if len(*slice) > 0 {
				outputs[key].Push(MakeBioSequenceBatch(orders[key], *slice))
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
