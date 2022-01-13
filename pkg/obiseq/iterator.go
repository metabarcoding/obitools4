package obiseq

import (
	"sync"
	"time"
)

// Private structure implementing an iterator over
// bioseq.BioSequence based on a channel.
type __ibiosequence__ struct {
	channel     chan BioSequence
	current     BioSequence
	all_done    *sync.WaitGroup
	buffer_size int
	finished    bool
	p_finished  *bool
}

type IBioSequence struct {
	pointer *__ibiosequence__
}

var NilIBioSequence = IBioSequence{pointer: nil}

func (iterator IBioSequence) IsNil() bool {
	return iterator.pointer == nil
}

func (iterator IBioSequence) Add(n int) {
	iterator.pointer.all_done.Add(n)
}

func (iterator IBioSequence) Done() {
	iterator.pointer.all_done.Done()
}

func (iterator IBioSequence) Wait() {
	iterator.pointer.all_done.Wait()
}

func (iterator IBioSequence) Channel() chan BioSequence {
	return iterator.pointer.channel
}
func (iterator IBioSequence) PChannel() *chan BioSequence {
	return &(iterator.pointer.channel)
}

func MakeIBioSequence(sizes ...int) IBioSequence {
	buffsize := 1

	if len(sizes) > 0 {
		buffsize = sizes[0]
	}

	i := __ibiosequence__{
		channel:     make(chan BioSequence, buffsize),
		current:     NilBioSequence,
		buffer_size: buffsize,
		finished:    false,
		p_finished:  nil}
	i.p_finished = &i.finished
	waiting := sync.WaitGroup{}
	i.all_done = &waiting
	ii := IBioSequence{&i}
	return ii
}

func (iterator IBioSequence) Split() IBioSequence {
	i := __ibiosequence__{
		channel:     iterator.pointer.channel,
		current:     NilBioSequence,
		finished:    false,
		all_done:    iterator.pointer.all_done,
		buffer_size: iterator.pointer.buffer_size,
		p_finished:  iterator.pointer.p_finished}
	new_iter := IBioSequence{&i}
	return new_iter
}

func (iterator IBioSequence) Next() bool {
	if iterator.IsNil() || *(iterator.pointer.p_finished) {
		iterator.pointer.current = NilBioSequence
		return false
	}

	next, ok := (<-iterator.pointer.channel)

	if ok {
		iterator.pointer.current = next
		return true
	}

	iterator.pointer.current = NilBioSequence
	*iterator.pointer.p_finished = true
	return false
}

// The 'Get' method returns the instance of BioSequence
// currently pointed by the iterator. You have to use the
// 'Next' method to move to the next entry before calling
// 'Get' to retreive the following instance.
func (iterator IBioSequence) Get() BioSequence {
	return iterator.pointer.current
}

// Finished returns 'true' value if no more data is available
// from the iterator.
func (iterator IBioSequence) Finished() bool {
	return *iterator.pointer.p_finished
}

func (iterator IBioSequence) BufferSize() int {
	return iterator.pointer.buffer_size
}

// The IBioSequenceBatch converts a IBioSequence iterator
// into an iterator oveer batches oof sequences. By default
// the size of a batch is of 100 sequences and the iterator
// implements a buffer equal to that of the source iterator.
// These defaults can be overriden by specifying one or two
// optional parametters at the method call. The first one
// indicates the batch size. The second optional parametter
// indicates the size of the buffer.
func (iterator IBioSequence) IBioSequenceBatch(sizes ...int) IBioSequenceBatch {
	batchsize := 100
	buffsize := iterator.BufferSize()

	if len(sizes) > 0 {
		batchsize = sizes[0]
	}
	if len(sizes) > 1 {
		buffsize = sizes[1]
	}

	new_iter := MakeIBioSequenceBatch(buffsize)

	new_iter.Add(1)

	go func() {
		new_iter.Wait()
		for len(new_iter.Channel()) > 0 {
			time.Sleep(time.Millisecond)
		}
		close(new_iter.pointer.channel)
	}()

	go func() {
		for j := 0; !iterator.Finished(); j++ {
			batch := BioSequenceBatch{
				slice: make(BioSequenceSlice, 0, batchsize),
				order: j}
			for i := 0; i < batchsize && iterator.Next(); i++ {
				seq := iterator.Get()
				batch.slice = append(batch.slice, seq)
			}
			new_iter.pointer.channel <- batch
		}
		new_iter.Done()
	}()

	return new_iter
}

func (iterator IBioSequence) IBioSequence(sizes ...int) IBioSequence {
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
			s := iterator.Get()
			new_iter.pointer.channel <- s
		}
		new_iter.Done()
	}()

	return new_iter
}

func (iterator IBioSequence) Skip(n int, sizes ...int) IBioSequence {
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
		for i := 0; iterator.Next(); i++ {
			if i >= n {
				s := iterator.Get()
				new_iter.pointer.channel <- s
			}
		}
		new_iter.Done()
	}()

	return new_iter
}

func (iterator IBioSequence) Head(n int, sizes ...int) IBioSequence {
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
		not_done := true
		for i := 0; iterator.Next(); i++ {
			if i < n {
				s := iterator.Get()
				new_iter.pointer.channel <- s
			} else {
				if not_done {
					new_iter.Done()
					not_done = false
				}
			}
		}
	}()

	return new_iter
}

// The 'Tail' method discard every data from the source iterator
// except the 'n' last ones.
func (iterator IBioSequence) Tail(n int, sizes ...int) IBioSequence {
	buffsize := iterator.BufferSize()

	if len(sizes) > 0 {
		buffsize = sizes[0]
	}

	new_iter := MakeIBioSequence(buffsize)
	buffseq := make(BioSequenceSlice, n)

	new_iter.Add(1)

	go func() {
		new_iter.Wait()
		close(new_iter.pointer.channel)
	}()

	go func() {
		var i int
		for i = 0; iterator.Next(); i++ {
			buffseq[i%n] = iterator.Get()
		}
		if i > n {
			for j := 0; j < n; j++ {
				new_iter.Channel() <- buffseq[(i+j)%n]
			}

		} else {
			for j := 0; j < i; j++ {
				new_iter.Channel() <- buffseq[j]
			}
		}
		new_iter.Done()
	}()

	return new_iter
}

func (iterator IBioSequence) Concat(iterators ...IBioSequence) IBioSequence {

	if len(iterators) == 0 {
		return iterator
	}

	buffsize := iterator.BufferSize()
	new_iter := MakeIBioSequence(buffsize)

	new_iter.Add(1)

	go func() {
		new_iter.Wait()
		close(new_iter.pointer.channel)
	}()

	go func() {
		for iterator.Next() {
			s := iterator.Get()
			new_iter.pointer.channel <- s
		}

		for _, iter := range iterators {
			for iter.Next() {
				s := iter.Get()
				new_iter.pointer.channel <- s
			}
		}
		new_iter.Done()
	}()

	return new_iter
}
