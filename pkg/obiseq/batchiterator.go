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
		buffer_size: buffsize,
		finished:    false,
		p_finished:  nil}
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
		all_done:    iterator.pointer.all_done,
		buffer_size: iterator.pointer.buffer_size,
		finished:    false,
		p_finished:  iterator.pointer.p_finished}
	new_iter := IBioSequenceBatch{&i}
	return new_iter
}

func (iterator IBioSequenceBatch) Next() bool {
	if *(iterator.pointer.p_finished) {
		return false
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

	new_iter := MakeIBioSequence(buffsize)

	new_iter.Add(1)

	go func() {
		new_iter.Wait()
		close(new_iter.pointer.channel)
	}()

	go func() {
		for iterator.Next() {
			batch := iterator.Get()

			for _, s := range batch.slice {
				new_iter.pointer.channel <- s
			}
		}
		new_iter.Done()
	}()

	return new_iter
}

func (iterator IBioSequenceBatch) SortBatches(sizes ...int) IBioSequenceBatch {
	buffsize := iterator.BufferSize()

	if len(sizes) > 0 {
		buffsize = sizes[0]
	}

	new_iter := MakeIBioSequenceBatch(buffsize)

	new_iter.Add(1)

	go func() {
		new_iter.Wait()
		close(new_iter.pointer.channel)
	}()

	next_to_send := 0
	received := make(map[int]BioSequenceBatch)
	go func() {
		for iterator.Next() {
			batch := iterator.Get()
			if batch.order == next_to_send {
				new_iter.pointer.channel <- batch
				next_to_send++
				batch, ok := received[next_to_send]
				for ok {
					new_iter.pointer.channel <- batch
					delete(received, next_to_send)
					next_to_send++
					batch, ok = received[next_to_send]
				}
			} else {
				received[batch.order] = batch
			}
		}
		new_iter.Done()
	}()

	return new_iter

}

func (iterator IBioSequenceBatch) Concat(iterators ...IBioSequenceBatch) IBioSequenceBatch {

	if len(iterators) == 0 {
		return iterator
	}

	buffsize := iterator.BufferSize()
	new_iter := MakeIBioSequenceBatch(buffsize)

	new_iter.Add(1)

	go func() {
		new_iter.Wait()
		close(new_iter.Channel())
	}()

	go func() {
		previous_max := 0
		max_order := 0

		for iterator.Next() {
			s := iterator.Get()
			if s.order > max_order {
				max_order = s.order
			}
			new_iter.Channel() <- MakeBioSequenceBatch(s.order+previous_max, s.slice...)
		}

		previous_max = max_order + 1
		for _, iter := range iterators {
			for iter.Next() {
				s := iter.Get()
				if (s.order + previous_max) > max_order {
					max_order = s.order + previous_max
				}

				new_iter.Channel() <- MakeBioSequenceBatch(s.order+previous_max, s.slice...)
			}
			previous_max = max_order + 1
		}
		new_iter.Done()
	}()

	return new_iter
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

	new_iter := MakeIBioSequenceBatch(buffsize)

	new_iter.Add(1)

	go func() {
		new_iter.Wait()
		close(new_iter.pointer.channel)
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
					new_iter.Channel() <- MakeBioSequenceBatch(order, buffer...)
					order++
					buffer = make(BioSequenceSlice, 0, size)
				}
			}
		}

		if len(buffer) > 0 {
			new_iter.Channel() <- MakeBioSequenceBatch(order, buffer...)
		}

		new_iter.Done()

	}()

	return new_iter
}

func (iterator IBioSequenceBatch) Destroy() {

	log.Println("Start recycling of Bioseq objects")

	for iterator.Next() {
		batch := iterator.Get()
		for _, seq := range batch.Slice() {
			(&seq).Destroy()
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

	new_iter := MakeIPairedBioSequenceBatch(buffsize)

	new_iter.Add(1)

	go func() {
		new_iter.Wait()
		close(new_iter.pointer.channel)
		log.Println("End of association of paired reads")
	}()

	log.Println("Start association of paired reads")
	go func() {
		for iterator.Next() {
			if !reverse.Next() {
				log.Panicln("Etrange reverse pas prêt")
			}
			new_iter.Channel() <- MakePairedBioSequenceBatch(iterator.Get(),
				reverse.Get())
		}

		new_iter.Done()
	}()

	return new_iter
}
