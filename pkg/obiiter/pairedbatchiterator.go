package obiiter

import (
	"log"
	"sync"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

type PairedBioSequenceBatch struct {
	forward obiseq.BioSequenceSlice
	reverse obiseq.BioSequenceSlice
	order   int
}

var NilPairedBioSequenceBatch = PairedBioSequenceBatch{nil, nil, -1}

func MakePairedBioSequenceBatch(forward, reverse BioSequenceBatch) PairedBioSequenceBatch {
	if forward.order != reverse.order {
		log.Fatalf("Forward order : %d and reverse order : %d are not matching",
			forward.order, reverse.order)
	}

	for i := range reverse.slice {
		reverse.slice[i].ReverseComplement(true)
	}

	return PairedBioSequenceBatch{
		forward: forward.slice,
		reverse: reverse.slice,
		order:   forward.order,
	}
}

func (batch PairedBioSequenceBatch) Order() int {
	return batch.order
}

func (batch PairedBioSequenceBatch) Reorder(newOrder int) PairedBioSequenceBatch {
	batch.order = newOrder
	return batch
}


func (batch PairedBioSequenceBatch) Length() int {
	return len(batch.forward)
}

func (batch PairedBioSequenceBatch) Forward() obiseq.BioSequenceSlice {
	return batch.forward
}

func (batch PairedBioSequenceBatch) Reverse() obiseq.BioSequenceSlice {
	return batch.reverse
}

func (batch PairedBioSequenceBatch) IsNil() bool {
	return batch.forward == nil
}

// Structure implementing an iterator over bioseq.BioSequenceBatch
// based on a channel.
type __ipairedbiosequencebatch__ struct {
	channel     chan PairedBioSequenceBatch
	current     PairedBioSequenceBatch
	pushBack    bool
	all_done    *sync.WaitGroup
	buffer_size int
	finished    bool
	p_finished  *bool
}

type IPairedBioSequenceBatch struct {
	pointer *__ipairedbiosequencebatch__
}

var NilIPairedBioSequenceBatch = IPairedBioSequenceBatch{pointer: nil}

func MakeIPairedBioSequenceBatch(sizes ...int) IPairedBioSequenceBatch {
	buffsize := 1

	if len(sizes) > 0 {
		buffsize = sizes[0]
	}

	i := __ipairedbiosequencebatch__{
		channel:     make(chan PairedBioSequenceBatch, buffsize),
		current:     NilPairedBioSequenceBatch,
		pushBack:    false,
		buffer_size: buffsize,
		finished:    false,
		p_finished:  nil,
	}

	i.p_finished = &i.finished
	waiting := sync.WaitGroup{}
	i.all_done = &waiting
	ii := IPairedBioSequenceBatch{&i}
	return ii
}

func (iterator IPairedBioSequenceBatch) Add(n int) {
	iterator.pointer.all_done.Add(n)
}

func (iterator IPairedBioSequenceBatch) Done() {
	iterator.pointer.all_done.Done()
}

func (iterator IPairedBioSequenceBatch) Wait() {
	iterator.pointer.all_done.Wait()
}

func (iterator IPairedBioSequenceBatch) Channel() chan PairedBioSequenceBatch {
	return iterator.pointer.channel
}

func (iterator IPairedBioSequenceBatch) IsNil() bool {
	return iterator.pointer == nil
}

func (iterator IPairedBioSequenceBatch) BufferSize() int {
	return iterator.pointer.buffer_size
}

func (iterator IPairedBioSequenceBatch) Split() IPairedBioSequenceBatch {
	i := __ipairedbiosequencebatch__{
		channel:     iterator.pointer.channel,
		current:     NilPairedBioSequenceBatch,
		pushBack:    false,
		all_done:    iterator.pointer.all_done,
		buffer_size: iterator.pointer.buffer_size,
		finished:    false,
		p_finished:  iterator.pointer.p_finished}
	newIter := IPairedBioSequenceBatch{&i}
	return newIter
}

func (iterator IPairedBioSequenceBatch) Next() bool {
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

	iterator.pointer.current = NilPairedBioSequenceBatch
	*iterator.pointer.p_finished = true
	return false
}

func (iterator IPairedBioSequenceBatch) PushBack() {
	if !iterator.pointer.current.IsNil() {
		iterator.pointer.pushBack = true
	}
}

// The 'Get' method returns the instance of BioSequenceBatch
// currently pointed by the iterator. You have to use the
// 'Next' method to move to the next entry before calling
// 'Get' to retreive the following instance.
func (iterator IPairedBioSequenceBatch) Get() PairedBioSequenceBatch {
	return iterator.pointer.current
}

// Finished returns 'true' value if no more data is available
// from the iterator.
func (iterator IPairedBioSequenceBatch) Finished() bool {
	return *iterator.pointer.p_finished
}

func (iterator IPairedBioSequenceBatch) SortBatches(sizes ...int) IPairedBioSequenceBatch {
	buffsize := iterator.BufferSize()

	if len(sizes) > 0 {
		buffsize = sizes[0]
	}

	newIter := MakeIPairedBioSequenceBatch(buffsize)

	newIter.Add(1)

	go func() {
		newIter.Wait()
		close(newIter.pointer.channel)
	}()

	next_to_send := 0
	received := make(map[int]PairedBioSequenceBatch)
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
