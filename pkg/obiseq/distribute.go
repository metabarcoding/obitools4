package obiseq

import (
	"fmt"
	"sync"
)

type IDistribute struct {
	outputs map[string]IBioSequenceBatch
	news    chan string
	lock    *sync.Mutex
}

func (dist *IDistribute) Outputs(key string) (IBioSequenceBatch, error) {
	dist.lock.Lock()
	iter, ok := dist.outputs[key]
	dist.lock.Unlock()

	if !ok {
		return NilIBioSequenceBatch, fmt.Errorf("key %s unknown", key)
	}

	return iter, nil
}

func (dist *IDistribute) News() chan string {
	return dist.news
}

func (iterator IBioSequenceBatch) Distribute(class SequenceClassifier, sizes ...int) IDistribute {
	batchsize := 5000
	buffsize := 2

	outputs := make(map[string]IBioSequenceBatch, 100)
	slices := make(map[string]*BioSequenceSlice, 100)
	orders := make(map[string]int, 100)
	news := make(chan string)

	if len(sizes) > 0 {
		batchsize = sizes[0]
	}

	if len(sizes) > 1 {
		buffsize = sizes[1]
	}

	jobDone := sync.WaitGroup{}
	lock := sync.Mutex{}

	jobDone.Add(1)

	go func() {
		jobDone.Wait()
		close(news)
		for _, i := range outputs {
			close(i.Channel())
		}
	}()

	go func() {
		iterator = iterator.SortBatches()

		for iterator.Next() {
			seqs := iterator.Get()
			for _, s := range seqs.Slice() {
				key := class(s)
				slice, ok := slices[key]

				if !ok {
					s := make(BioSequenceSlice, 0, batchsize)
					slice = &s
					slices[key] = slice
					orders[key] = 0

					lock.Lock()
					outputs[key] = MakeIBioSequenceBatch(buffsize)
					lock.Unlock()

					news <- key
				}

				*slice = append(*slice, s)
				
				if len(*slice) == batchsize {
					outputs[key].Channel() <- MakeBioSequenceBatch(orders[key], *slice...)
					orders[key]++
					s := make(BioSequenceSlice, 0, batchsize)
					slices[key] = &s
				}
			}
		}

		for key, slice := range slices {
			if len(*slice) > 0 {
				outputs[key].Channel() <- MakeBioSequenceBatch(orders[key], *slice...)
			}
		}

		jobDone.Done()

	}()

	return IDistribute{
		outputs,
		news,
		&lock}

}
