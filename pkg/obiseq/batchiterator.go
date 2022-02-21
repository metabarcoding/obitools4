package obiseq

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tevino/abool/v2"
)

type BioSequenceBatch struct {
	slice BioSequenceSlice
	order int
}

var NilBioSequenceBatch = BioSequenceBatch{nil, -1}

func MakeBioSequenceBatch(order int, sequences BioSequenceSlice) BioSequenceBatch {
	return BioSequenceBatch{
		slice: sequences,
		order: order,
	}
}

func (batch BioSequenceBatch) Order() int {
	return batch.order
}

func (batch BioSequenceBatch) Reorder(newOrder int) BioSequenceBatch {
	batch.order = newOrder
	return batch
}

func (batch BioSequenceBatch) Slice() BioSequenceSlice {
	return batch.slice
}

func (batch BioSequenceBatch) Length() int {
	return len(batch.slice)
}

func (batch BioSequenceBatch) NotEmpty() bool {
	return batch.slice.NotEmpty()
}

func (batch BioSequenceBatch) Pop0() *BioSequence {
	return batch.slice.Pop0()
}

func (batch BioSequenceBatch) IsNil() bool {
	return batch.slice == nil
}

func (batch BioSequenceBatch) Recycle() {
	batch.slice.Recycle()
	batch.slice = nil
}

// Structure implementing an iterator over bioseq.BioSequenceBatch
// based on a channel.
type _IBioSequenceBatch struct {
	channel         chan BioSequenceBatch
	current         BioSequenceBatch
	pushBack        *abool.AtomicBool
	all_done        *sync.WaitGroup
	lock            *sync.RWMutex
	buffer_size     int32
	batch_size      int32
	sequence_format string
	finished        *abool.AtomicBool
}

type IBioSequenceBatch struct {
	pointer *_IBioSequenceBatch
}

var NilIBioSequenceBatch = IBioSequenceBatch{pointer: nil}

func MakeIBioSequenceBatch(sizes ...int) IBioSequenceBatch {
	buffsize := int32(1)

	if len(sizes) > 0 {
		buffsize = int32(sizes[0])
	}

	i := _IBioSequenceBatch{
		channel:         make(chan BioSequenceBatch, buffsize),
		current:         NilBioSequenceBatch,
		pushBack:        abool.New(),
		buffer_size:     buffsize,
		batch_size:      -1,
		sequence_format: "",
		finished:        abool.New(),
	}

	waiting := sync.WaitGroup{}
	i.all_done = &waiting
	lock := sync.RWMutex{}
	i.lock = &lock
	ii := IBioSequenceBatch{&i}
	return ii
}

func (iterator IBioSequenceBatch) Add(n int) {
	iterator.pointer.all_done.Add(n)
}

func (iterator IBioSequenceBatch) Done() {
	iterator.pointer.all_done.Done()
}

func (iterator IBioSequenceBatch) Unlock() {
	iterator.pointer.lock.Unlock()
}

func (iterator IBioSequenceBatch) Lock() {
	iterator.pointer.lock.Lock()
}

func (iterator IBioSequenceBatch) RLock() {
	iterator.pointer.lock.RLock()
}

func (iterator IBioSequenceBatch) RUnlock() {
	iterator.pointer.lock.RUnlock()
}

func (iterator IBioSequenceBatch) Wait() {
	iterator.pointer.all_done.Wait()
}

func (iterator IBioSequenceBatch) Channel() chan BioSequenceBatch {
	return iterator.pointer.channel
}

func (iterator IBioSequenceBatch) IsNil() bool {
	return iterator.pointer == nil
}

func (iterator IBioSequenceBatch) BufferSize() int {
	return int(atomic.LoadInt32(&iterator.pointer.buffer_size))
}

func (iterator IBioSequenceBatch) BatchSize() int {
	return int(atomic.LoadInt32(&iterator.pointer.batch_size))
}

func (iterator IBioSequenceBatch) SetBatchSize(size int) error {
	if size >= 0 {
		atomic.StoreInt32(&iterator.pointer.batch_size, int32(size))
		return nil
	}

	return fmt.Errorf("size (%d) cannot be negative", size)
}

func (iterator IBioSequenceBatch) Split() IBioSequenceBatch {
	iterator.pointer.lock.RLock()
	defer iterator.pointer.lock.RUnlock()
	i := _IBioSequenceBatch{
		channel:         iterator.pointer.channel,
		current:         NilBioSequenceBatch,
		pushBack:        abool.New(),
		all_done:        iterator.pointer.all_done,
		buffer_size:     iterator.pointer.buffer_size,
		batch_size:      iterator.pointer.batch_size,
		sequence_format: iterator.pointer.sequence_format,
		finished:        iterator.pointer.finished}
	lock := sync.RWMutex{}
	i.lock = &lock

	newIter := IBioSequenceBatch{&i}
	return newIter
}

func (iterator IBioSequenceBatch) Next() bool {
	if iterator.pointer.pushBack.IsSet() {
		iterator.pointer.pushBack.UnSet()
		return true
	}

	if iterator.pointer.finished.IsSet() {
		return false
	}

	next, ok := (<-iterator.pointer.channel)

	if ok {
		iterator.pointer.current = next
		return true
	}

	iterator.pointer.current = NilBioSequenceBatch
	iterator.pointer.finished.Set()
	return false
}

func (iterator IBioSequenceBatch) PushBack() {
	if !iterator.pointer.current.IsNil() {
		iterator.pointer.pushBack.Set()
	}
}

// The 'Get' method returns the instance of BioSequenceBatch
// currently pointed by the iterator. You have to use the
// 'Next' method to move to the next entry before calling
// 'Get' to retreive the following instance.
func (iterator IBioSequenceBatch) Get() BioSequenceBatch {
	return iterator.pointer.current
}

func (iterator IBioSequenceBatch) Push(batch BioSequenceBatch) {
	if batch.IsNil() {
		log.Panicln("An Nil batch is pushed on the channel")
	}
	if batch.Length() == 0 {
		log.Panicln("An empty batch is pushed on the channel")
	}

	iterator.pointer.channel <- batch
}

func (iterator IBioSequenceBatch) Close() {
	close(iterator.pointer.channel)
}

func (iterator IBioSequenceBatch) WaitAndClose() {
	iterator.Wait()

	for len(iterator.Channel()) > 0 {
		time.Sleep(time.Millisecond)
	}
	iterator.Close()
}

// Finished returns 'true' value if no more data is available
// from the iterator.
func (iterator IBioSequenceBatch) Finished() bool {
	return iterator.pointer.finished.IsSet()
}

func (iterator IBioSequenceBatch) IBioSequence(sizes ...int) IBioSequence {
	buffsize := iterator.BufferSize()

	if len(sizes) > 0 {
		buffsize = sizes[0]
	}

	newIter := MakeIBioSequence(buffsize)

	newIter.Add(1)

	go func() {
		newIter.Wait()
		close(newIter.pointer.channel)
	}()

	go func() {
		for iterator.Next() {
			batch := iterator.Get()

			for batch.NotEmpty() {
				newIter.pointer.channel <- batch.Pop0()
			}
			batch.Recycle()
		}
		newIter.Done()
	}()

	return newIter
}

func (iterator IBioSequenceBatch) SortBatches(sizes ...int) IBioSequenceBatch {
	buffsize := iterator.BufferSize()

	if len(sizes) > 0 {
		buffsize = sizes[0]
	}

	newIter := MakeIBioSequenceBatch(buffsize)

	newIter.Add(1)

	go func() {
		newIter.Wait()
		close(newIter.pointer.channel)
	}()

	next_to_send := 0
	received := make(map[int]BioSequenceBatch)
	go func() {
		for iterator.Next() {
			batch := iterator.Get()
			if batch.order == next_to_send {
				newIter.pointer.channel <- batch
				next_to_send++
				batch, ok := received[next_to_send]
				for ok {
					newIter.pointer.channel <- batch
					delete(received, next_to_send)
					next_to_send++
					batch, ok = received[next_to_send]
				}
			} else {
				received[batch.order] = batch
			}
		}
		newIter.Done()
	}()

	return newIter

}

func (iterator IBioSequenceBatch) Concat(iterators ...IBioSequenceBatch) IBioSequenceBatch {

	if len(iterators) == 0 {
		return iterator
	}

	buffsize := iterator.BufferSize()
	newIter := MakeIBioSequenceBatch(buffsize)

	newIter.Add(1)

	go func() {
		newIter.Wait()
		close(newIter.Channel())
	}()

	go func() {
		previous_max := 0
		max_order := 0

		for iterator.Next() {
			s := iterator.Get()
			if s.order > max_order {
				max_order = s.order
			}
			newIter.Push(s.Reorder(s.order + previous_max))
		}

		previous_max = max_order + 1
		for _, iter := range iterators {
			for iter.Next() {
				s := iter.Get()
				if (s.order + previous_max) > max_order {
					max_order = s.order + previous_max
				}

				newIter.Push(s.Reorder(s.order + previous_max))
			}
			previous_max = max_order + 1
		}
		newIter.Done()
	}()

	return newIter
}

// Redistributes sequences from a IBioSequenceBatch into a new
// IBioSequenceBatch with every batches having the same size
// indicated in parameter. Rebatching implies to sort the
// source IBioSequenceBatch.
func (iterator IBioSequenceBatch) Rebatch(size int, sizes ...int) IBioSequenceBatch {
	buffsize := iterator.BufferSize()

	if len(sizes) > 0 {
		buffsize = sizes[0]
	}

	newIter := MakeIBioSequenceBatch(buffsize)

	newIter.Add(1)

	go func() {
		newIter.Wait()
		close(newIter.pointer.channel)
	}()

	go func() {
		order := 0
		iterator = iterator.SortBatches()
		buffer := MakeBioSequenceSlice()

		for iterator.Next() {
			seqs := iterator.Get()
			for _, s := range seqs.slice {
				buffer = append(buffer, s)
				if len(buffer) == size {
					newIter.Push(MakeBioSequenceBatch(order, buffer))
					order++
					buffer = MakeBioSequenceSlice()
				}
			}
			seqs.Recycle()
		}

		if len(buffer) > 0 {
			newIter.Push(MakeBioSequenceBatch(order, buffer))
		}

		newIter.Done()

	}()

	return newIter
}

func (iterator IBioSequenceBatch) Recycle() {

	log.Println("Start recycling of Bioseq objects")
	recycled := 0
	for iterator.Next() {
		// iterator.Get()
		batch := iterator.Get()
		for _, seq := range batch.Slice() {
			seq.Recycle()
			recycled++
		}
		batch.Recycle()
	}
	log.Printf("End of the recycling of %d Bioseq objects", recycled)
}

func (iterator IBioSequenceBatch) PairWith(reverse IBioSequenceBatch, sizes ...int) IPairedBioSequenceBatch {
	buffsize := iterator.BufferSize()
	batchsize := 5000

	if len(sizes) > 0 {
		batchsize = sizes[0]
	}

	if len(sizes) > 1 {
		buffsize = sizes[1]
	}

	iterator = iterator.Rebatch(batchsize)
	reverse = reverse.Rebatch(batchsize)

	newIter := MakeIPairedBioSequenceBatch(buffsize)

	newIter.Add(1)

	go func() {
		newIter.Wait()
		close(newIter.pointer.channel)
		log.Println("End of association of paired reads")
	}()

	log.Println("Start association of paired reads")
	go func() {
		for iterator.Next() {
			if !reverse.Next() {
				log.Panicln("Etrange reverse pas prêt")
			}
			newIter.Channel() <- MakePairedBioSequenceBatch(iterator.Get(),
				reverse.Get())
		}

		newIter.Done()
	}()

	return newIter
}

func (iterator IBioSequenceBatch) DivideOn(predicate SequencePredicate,
	size int, sizes ...int) (IBioSequenceBatch, IBioSequenceBatch) {
	buffsize := iterator.BufferSize()

	if len(sizes) > 0 {
		buffsize = sizes[0]
	}

	trueIter := MakeIBioSequenceBatch(buffsize)
	falseIter := MakeIBioSequenceBatch(buffsize)

	trueIter.Add(1)
	falseIter.Add(1)

	go func() {
		trueIter.WaitAndClose()
		falseIter.WaitAndClose()
	}()

	go func() {
		trueOrder := 0
		falseOrder := 0
		iterator = iterator.SortBatches()

		trueSlice := MakeBioSequenceSlice()
		falseSlice := MakeBioSequenceSlice()

		for iterator.Next() {
			seqs := iterator.Get()
			for _, s := range seqs.slice {
				if predicate(s) {
					trueSlice = append(trueSlice, s)
				} else {
					falseSlice = append(falseSlice, s)
				}

				if len(trueSlice) == size {
					trueIter.Push(MakeBioSequenceBatch(trueOrder, trueSlice))
					trueOrder++
					trueSlice = MakeBioSequenceSlice()
				}

				if len(falseSlice) == size {
					falseIter.Push(MakeBioSequenceBatch(falseOrder, falseSlice))
					falseOrder++
					falseSlice = MakeBioSequenceSlice()
				}
			}
			seqs.Recycle()
		}

		if len(trueSlice) > 0 {
			trueIter.Push(MakeBioSequenceBatch(trueOrder, trueSlice))
		}

		if len(falseSlice) > 0 {
			falseIter.Push(MakeBioSequenceBatch(falseOrder, falseSlice))
		}

		trueIter.Done()
		falseIter.Done()
	}()

	return trueIter, falseIter
}
