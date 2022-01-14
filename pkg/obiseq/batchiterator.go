package obiseq

import (
	"log"
	"sync"
)

type BioSequenceBatch struct {
	slice BioSequenceSlice
	order int
}

var NilBioSequenceBatch = BioSequenceBatch{nil, -1}

func MakeBioSequenceBatch(order int, sequences ...BioSequence) BioSequenceBatch {
	return BioSequenceBatch{
		slice: sequences,
		order: order,
	}
}

func (batch BioSequenceBatch) Order() int {
	return batch.order
}

func (batch BioSequenceBatch) Slice() BioSequenceSlice {
	return batch.slice
}

func (batch BioSequenceBatch) Length() int {
	return len(batch.slice)
}
func (batch BioSequenceBatch) IsNil() bool {
	return batch.slice == nil
}

// Structure implementing an iterator over bioseq.BioSequenceBatch
// based on a channel.
type __ibiosequencebatch__ struct {
	channel     chan BioSequenceBatch
	current     BioSequenceBatch
	pushBack    bool
	all_done    *sync.WaitGroup
	buffer_size int
	finished    bool
	p_finished  *bool
}

type IBioSequenceBatch struct {
	pointer *__ibiosequencebatch__
}

var NilIBioSequenceBatch = IBioSequenceBatch{pointer: nil}

func MakeIBioSequenceBatch(sizes ...int) IBioSequenceBatch {
	buffsize := 1

	if len(sizes) > 0 {
		buffsize = sizes[0]
	}

	i := __ibiosequencebatch__{
		channel:     make(chan BioSequenceBatch, buffsize),
		current:     NilBioSequenceBatch,
		pushBack:    false,
		buffer_size: buffsize,
		finished:    false,
		p_finished:  nil,
	}
	i.p_finished = &i.finished
	waiting := sync.WaitGroup{}
	i.all_done = &waiting
	ii := IBioSequenceBatch{&i}
	return ii
}

func (iterator IBioSequenceBatch) Add(n int) {
	iterator.pointer.all_done.Add(n)
}

func (iterator IBioSequenceBatch) Done() {
	iterator.pointer.all_done.Done()
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
	return iterator.pointer.buffer_size
}

func (iterator IBioSequenceBatch) Split() IBioSequenceBatch {
	i := __ibiosequencebatch__{
		channel:     iterator.pointer.channel,
		current:     NilBioSequenceBatch,
		pushBack:    false,
		all_done:    iterator.pointer.all_done,
		buffer_size: iterator.pointer.buffer_size,
		finished:    false,
		p_finished:  iterator.pointer.p_finished}
	newIter := IBioSequenceBatch{&i}
	return newIter
}

func (iterator IBioSequenceBatch) Next() bool {
	if *(iterator.pointer.p_finished) {
		return false
	}

	if iterator.pointer.pushBack {
		iterator.pointer.pushBack = false
		return true
	}

	next, ok := (<-iterator.pointer.channel)

	if ok {
		iterator.pointer.current = next
		return true
	}

	iterator.pointer.current = NilBioSequenceBatch
	*iterator.pointer.p_finished = true
	return false
}

func (iterator IBioSequenceBatch) PushBack() {
	if !iterator.pointer.current.IsNil() {
		iterator.pointer.pushBack = true
	}
}

// The 'Get' method returns the instance of BioSequenceBatch
// currently pointed by the iterator. You have to use the
// 'Next' method to move to the next entry before calling
// 'Get' to retreive the following instance.
func (iterator IBioSequenceBatch) Get() BioSequenceBatch {
	return iterator.pointer.current
}

// Finished returns 'true' value if no more data is available
// from the iterator.
func (iterator IBioSequenceBatch) Finished() bool {
	return *iterator.pointer.p_finished
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

			for _, s := range batch.slice {
				newIter.pointer.channel <- s
			}
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
			newIter.Channel() <- MakeBioSequenceBatch(s.order+previous_max, s.slice...)
		}

		previous_max = max_order + 1
		for _, iter := range iterators {
			for iter.Next() {
				s := iter.Get()
				if (s.order + previous_max) > max_order {
					max_order = s.order + previous_max
				}

				newIter.Channel() <- MakeBioSequenceBatch(s.order+previous_max, s.slice...)
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
		buffer := make(BioSequenceSlice, 0, size)

		for iterator.Next() {
			seqs := iterator.Get()
			for _, s := range seqs.slice {
				buffer = append(buffer, s)
				if len(buffer) == size {
					newIter.Channel() <- MakeBioSequenceBatch(order, buffer...)
					order++
					buffer = make(BioSequenceSlice, 0, size)
				}
			}
		}

		if len(buffer) > 0 {
			newIter.Channel() <- MakeBioSequenceBatch(order, buffer...)
		}

		newIter.Done()

	}()

	return newIter
}

func (iterator IBioSequenceBatch) Recycle() {

	log.Println("Start recycling of Bioseq objects")

	for iterator.Next() {
		batch := iterator.Get()
		for _, seq := range batch.Slice() {
			(&seq).Recycle()
		}
	}
	log.Println("End of the recycling of Bioseq objects")
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
