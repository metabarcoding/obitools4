package obiseq

import (
	"fmt"
	"sync"
)

type IDistribute struct {
	outputs map[int]IBioSequenceBatch
	news    chan int
	lock    *sync.Mutex
}

func (dist *IDistribute) Outputs(key int) (IBioSequenceBatch, error) {
	dist.lock.Lock()
	iter, ok := dist.outputs[key]
	dist.lock.Unlock()

	if !ok {
		return NilIBioSequenceBatch, fmt.Errorf("code %d unknown", key)
	}

	return iter, nil
}

func (dist *IDistribute) News() chan int {
	return dist.news
}

func (iterator IBioSequenceBatch) Distribute(class *BioSequenceClassifier, sizes ...int) IDistribute {
	batchsize := 5000
	buffsize := 2

	outputs := make(map[int]IBioSequenceBatch, 100)
	slices := make(map[int]*BioSequenceSlice, 100)
	orders := make(map[int]int, 100)
	news := make(chan int)

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
				key := class.Code(s)
				slice, ok := slices[key]

				if !ok {
					s := GetBioSequenceSlice()
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
					s := GetBioSequenceSlice()
					slices[key] = &s
				}
			}
			seqs.Recycle()
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
