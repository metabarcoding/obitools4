// It takes a slice of BioSequence objects, and returns an iterator that will return batches of
// BioSequence objects
package obiiter

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/tevino/abool/v2"
)

var globalLocker sync.WaitGroup
var globalLockerCounter = 0

func RegisterAPipe() {
	globalLocker.Add(1)
	globalLockerCounter++
	log.Debugln(globalLockerCounter, " Pipes are registered now")
}

func UnregisterPipe() {
	globalLocker.Done()
	globalLockerCounter--
	log.Debugln(globalLockerCounter, "are still registered")
}

func WaitForLastPipe() {
	globalLocker.Wait()
}

// Structure implementing an iterator over bioseq.BioSequenceBatch
// based on a channel.
type _IBioSequence struct {
	channel         chan BioSequenceBatch
	current         BioSequenceBatch
	pushBack        *abool.AtomicBool
	all_done        *sync.WaitGroup
	lock            *sync.RWMutex
	buffer_size     int32
	batch_size      int32
	sequence_format string
	finished        *abool.AtomicBool
	paired          bool
}

type IBioSequence struct {
	pointer *_IBioSequence
}

// NilIBioSequence nil instance for IBioSequenceBatch
//
// NilIBioSequence is the nil instance for the
// IBioSequenceBatch type.
var NilIBioSequence = IBioSequence{pointer: nil}

func MakeIBioSequence() IBioSequence {

	i := _IBioSequence{
		channel:         make(chan BioSequenceBatch),
		current:         NilBioSequenceBatch,
		pushBack:        abool.New(),
		batch_size:      -1,
		sequence_format: "",
		finished:        abool.New(),
		paired:          false,
	}

	waiting := sync.WaitGroup{}
	i.all_done = &waiting
	lock := sync.RWMutex{}
	i.lock = &lock
	ii := IBioSequence{&i}

	RegisterAPipe()

	return ii
}

func (iterator IBioSequence) Add(n int) {
	if iterator.pointer == nil {
		log.Panic("call of IBioSequenceBatch.Add method on NilIBioSequenceBatch")
	}

	iterator.pointer.all_done.Add(n)
}

func (iterator IBioSequence) Done() {
	if iterator.pointer == nil {
		log.Panic("call of IBioSequenceBatch.Done method on NilIBioSequenceBatch")
	}

	iterator.pointer.all_done.Done()
}

func (iterator IBioSequence) Unlock() {
	if iterator.pointer == nil {
		log.Panic("call of IBioSequenceBatch.Unlock method on NilIBioSequenceBatch")
	}

	iterator.pointer.lock.Unlock()
}

func (iterator IBioSequence) Lock() {
	if iterator.pointer == nil {
		log.Panic("call of IBioSequenceBatch.Lock method on NilIBioSequenceBatch")
	}

	iterator.pointer.lock.Lock()
}

func (iterator IBioSequence) RLock() {
	if iterator.pointer == nil {
		log.Panic("call of IBioSequenceBatch.RLock method on NilIBioSequenceBatch")
	}

	iterator.pointer.lock.RLock()
}

func (iterator IBioSequence) RUnlock() {
	if iterator.pointer == nil {
		log.Panic("call of IBioSequenceBatch.RUnlock method on NilIBioSequenceBatch")
	}

	iterator.pointer.lock.RUnlock()
}

func (iterator IBioSequence) Wait() {
	if iterator.pointer == nil {
		log.Panic("call of IBioSequenceBatch.Wait method on NilIBioSequenceBatch")
	}

	iterator.pointer.all_done.Wait()
}

func (iterator IBioSequence) Channel() chan BioSequenceBatch {
	if iterator.pointer == nil {
		log.Panic("call of IBioSequenceBatch.Channel method on NilIBioSequenceBatch")
	}

	return iterator.pointer.channel
}

func (iterator IBioSequence) IsNil() bool {
	if iterator.pointer == nil {
		log.Panic("call of IBioSequenceBatch.IsNil method on NilIBioSequenceBatch")
	}

	return iterator.pointer == nil
}

func (iterator IBioSequence) BatchSize() int {
	if iterator.pointer == nil {
		log.Panic("call of IBioSequenceBatch.BatchSize method on NilIBioSequenceBatch")
	}

	return int(atomic.LoadInt32(&iterator.pointer.batch_size))
}

func (iterator IBioSequence) SetBatchSize(size int) error {
	if size >= 0 {
		atomic.StoreInt32(&iterator.pointer.batch_size, int32(size))
		return nil
	}

	return fmt.Errorf("size (%d) cannot be negative", size)
}

func (iterator IBioSequence) Split() IBioSequence {
	iterator.pointer.lock.RLock()
	defer iterator.pointer.lock.RUnlock()
	i := _IBioSequence{
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

	newIter := IBioSequence{&i}

	if iterator.IsPaired() {
		newIter.MarkAsPaired()
	}

	return newIter
}

func (iterator IBioSequence) Next() bool {
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

func (iterator IBioSequence) PushBack() {
	if !iterator.pointer.current.IsNil() {
		iterator.pointer.pushBack.Set()
	}
}

// The 'Get' method returns the instance of BioSequenceBatch
// currently pointed by the iterator. You have to use the
// 'Next' method to move to the next entry before calling
// 'Get' to retreive the following instance.
func (iterator IBioSequence) Get() BioSequenceBatch {
	return iterator.pointer.current
}

func (iterator IBioSequence) Push(batch BioSequenceBatch) {
	if batch.IsNil() {
		log.Panicln("A Nil batch is pushed on the channel")
	}
	// if batch.Len() == 0 {
	// 	log.Panicln("An empty batch is pushed on the channel")
	// }

	iterator.pointer.channel <- batch
}

func (iterator IBioSequence) Close() {
	close(iterator.pointer.channel)
	UnregisterPipe()
}

func (iterator IBioSequence) WaitAndClose() {
	iterator.Wait()

	for len(iterator.Channel()) > 0 {
		time.Sleep(time.Millisecond)
	}

	iterator.Close()
}

// Finished returns 'true' value if no more data is available
// from the iterator.
func (iterator IBioSequence) Finished() bool {
	return iterator.pointer.finished.IsSet()
}

// Sorting the batches of sequences.
func (iterator IBioSequence) SortBatches(sizes ...int) IBioSequence {

	newIter := MakeIBioSequence()

	newIter.Add(1)

	go func() {
		newIter.WaitAndClose()
	}()

	next_to_send := 0
	//log.Println("wait for batch #", next_to_send)
	received := make(map[int]BioSequenceBatch)
	go func() {
		for iterator.Next() {
			batch := iterator.Get()
			// log.Println("\nPushd seq #\n", batch.order, next_to_send)

			if batch.order == next_to_send {
				newIter.pointer.channel <- batch
				next_to_send++
				//log.Println("\nwait for batch #\n", next_to_send)
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

	if iterator.IsPaired() {
		newIter.MarkAsPaired()
	}

	return newIter

}

func (iterator IBioSequence) Concat(iterators ...IBioSequence) IBioSequence {

	if len(iterators) == 0 {
		return iterator
	}

	allPaired := iterator.IsPaired()
	for _, i := range iterators {
		allPaired = allPaired && i.IsPaired()
	}

	newIter := MakeIBioSequence()

	newIter.Add(1)

	go func() {
		newIter.WaitAndClose()
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

	if allPaired {
		newIter.MarkAsPaired()
	}

	return newIter
}

func (iterator IBioSequence) Pool(iterators ...IBioSequence) IBioSequence {

	niterator := len(iterators) + 1

	if niterator == 1 {
		return iterator
	}

	allPaired := iterator.IsPaired()

	for _, i := range iterators {
		allPaired = allPaired && i.IsPaired()
	}

	nextCounter := obiutils.AtomicCounter()
	newIter := MakeIBioSequence()

	newIter.Add(niterator)

	go func() {
		newIter.WaitAndClose()
	}()

	ff := func(iterator IBioSequence) {

		for iterator.Next() {
			s := iterator.Get()
			newIter.Push(s.Reorder(nextCounter()))
		}
		newIter.Done()
	}

	go ff(iterator)
	for _, i := range iterators {
		go ff(i)
	}

	if allPaired {
		newIter.MarkAsPaired()
	}

	return newIter
}

// Redistributes sequences from a IBioSequenceBatch into a new
// IBioSequenceBatch with every batches having the same size
// indicated in parameter. Rebatching implies to sort the
// source IBioSequenceBatch.
func (iterator IBioSequence) Rebatch(size int) IBioSequence {

	newIter := MakeIBioSequence()

	newIter.Add(1)

	go func() {
		newIter.WaitAndClose()
	}()

	go func() {
		order := 0
		iterator = iterator.SortBatches()
		buffer := obiseq.MakeBioSequenceSlice()

		for iterator.Next() {
			seqs := iterator.Get()
			lc := seqs.Len()
			remains := lc
			i := 0
			for remains > 0 {
				space := size - len(buffer)
				to_push := min(lc-i, space)
				remains = lc - to_push - i
				buffer = append(buffer, seqs.Slice()[i:(i+to_push)]...)
				if len(buffer) == size {
					newIter.Push(MakeBioSequenceBatch(order, buffer))
					order++
					buffer = obiseq.MakeBioSequenceSlice()
				}
				i += to_push
			}
			seqs.Recycle(false)
		}

		if len(buffer) > 0 {
			newIter.Push(MakeBioSequenceBatch(order, buffer))
		}

		newIter.Done()

	}()

	if iterator.IsPaired() {
		newIter.MarkAsPaired()
	}

	return newIter
}

func (iterator IBioSequence) Recycle() {

	log.Debugln("Start recycling of Bioseq objects")
	recycled := 0
	for iterator.Next() {
		// iterator.Get()
		batch := iterator.Get()
		log.Debugln("Recycling batch #", batch.Order())
		recycled += batch.Len()
		batch.Recycle(true)
	}
	log.Debugf("End of the recycling of %d Bioseq objects", recycled)
}

func (iterator IBioSequence) Consume() {
	for iterator.Next() {
		batch := iterator.Get()
		batch.Recycle(false)
	}
}

func (iterator IBioSequence) Count(recycle bool) (int, int, int) {
	variants := 0
	reads := 0
	nucleotides := 0

	log.Debugln("Start counting of Bioseq objects")
	for iterator.Next() {
		// iterator.Get()
		batch := iterator.Get()
		for _, seq := range batch.Slice() {
			variants++
			reads += seq.Count()
			nucleotides += seq.Len()
		}
		batch.Recycle(recycle)
	}
	log.Debugf("End of the counting of %d Bioseq objects", variants)
	return variants, reads, nucleotides
}

// A function that takes a predicate and returns two IBioSequenceBatch iterators.
// Sequences extracted from the input iterator are distributed among both the
// iterator following the predicate value.
func (iterator IBioSequence) DivideOn(predicate obiseq.SequencePredicate,
	size int, sizes ...int) (IBioSequence, IBioSequence) {

	trueIter := MakeIBioSequence()
	falseIter := MakeIBioSequence()

	if iterator.IsPaired() {
		trueIter.MarkAsPaired()
		falseIter.MarkAsPaired()
	}

	trueIter.Add(1)
	falseIter.Add(1)

	go func() {
		trueOrder := 0
		falseOrder := 0
		iterator = iterator.SortBatches()

		trueSlice := obiseq.MakeBioSequenceSlice()
		falseSlice := obiseq.MakeBioSequenceSlice()

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
					trueSlice = obiseq.MakeBioSequenceSlice()
				}

				if len(falseSlice) == size {
					falseIter.Push(MakeBioSequenceBatch(falseOrder, falseSlice))
					falseOrder++
					falseSlice = obiseq.MakeBioSequenceSlice()
				}
			}
			seqs.Recycle(false)
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

	go func() {
		trueIter.WaitAndClose()
		falseIter.WaitAndClose()
	}()

	return trueIter, falseIter
}

// Filtering a batch of sequences.
// A function that takes a predicate and a batch of sequences and returns a filtered batch of sequences.
func (iterator IBioSequence) FilterOn(predicate obiseq.SequencePredicate,
	size int, sizes ...int) IBioSequence {
	nworkers := obioptions.CLIReadParallelWorkers()

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	trueIter := MakeIBioSequence()

	trueIter.Add(nworkers)

	go func() {
		trueIter.WaitAndClose()
	}()

	ff := func(iterator IBioSequence) {
		// iterator = iterator.SortBatches()

		for iterator.Next() {
			seqs := iterator.Get()
			slice := seqs.slice
			j := 0
			for _, s := range slice {
				if predicate(s) {
					slice[j] = s
					j++
				} else {
					s.Recycle()
				}
			}

			seqs.slice = slice[:j]
			trueIter.pointer.channel <- seqs
		}

		trueIter.Done()
	}

	for i := 1; i < nworkers; i++ {
		go ff(iterator.Split())
	}

	go ff(iterator)

	if iterator.IsPaired() {
		trueIter.MarkAsPaired()
	}

	return trueIter.Rebatch(size)
}

func (iterator IBioSequence) FilterAnd(predicate obiseq.SequencePredicate,
	size int, sizes ...int) IBioSequence {
	nworkers := obioptions.CLIReadParallelWorkers()

	if len(sizes) > 0 {
		nworkers = sizes[0]
	}

	trueIter := MakeIBioSequence()

	trueIter.Add(nworkers)

	go func() {
		trueIter.WaitAndClose()
	}()

	ff := func(iterator IBioSequence) {
		// iterator = iterator.SortBatches()

		for iterator.Next() {
			seqs := iterator.Get()
			slice := seqs.slice
			j := 0
			for _, s := range slice {
				good := predicate(s)
				if s.IsPaired() {
					good = good && predicate(s.PairedWith())
				}
				if good {
					slice[j] = s
					j++
				} else {
					s.Recycle()
				}
			}

			seqs.slice = slice[:j]
			trueIter.pointer.channel <- seqs
		}

		trueIter.Done()
	}

	for i := 1; i < nworkers; i++ {
		go ff(iterator.Split())
	}

	go ff(iterator)

	if iterator.IsPaired() {
		trueIter.MarkAsPaired()
	}

	return trueIter.Rebatch(size)
}

// Load all sequences availables from an IBioSequenceBatch iterator into
// a large obiseq.BioSequenceSlice.
func (iterator IBioSequence) Load() obiseq.BioSequenceSlice {

	chunck := obiseq.MakeBioSequenceSlice()
	for iterator.Next() {
		b := iterator.Get()
		log.Debugf("append %d sequences", b.Len())
		chunck = append(chunck, b.Slice()...)
		b.Recycle(false)
	}

	return chunck
}

// CompleteFileIterator generates a new iterator for reading a complete file.
//
// This iterator reads all the remaining sequences in the file, and returns them as a
// single obiseq.BioSequenceSlice.
//
// The function takes no parameters.
// It returns an IBioSequence object.
func (iterator IBioSequence) CompleteFileIterator() IBioSequence {

	newIter := MakeIBioSequence()
	log.Debug("Stream is read in full file mode")

	newIter.Add(1)

	go func() {
		newIter.WaitAndClose()
	}()

	go func() {
		slice := iterator.Load()
		log.Printf("A batch of  %d sequence is read", len(slice))
		if len(slice) > 0 {
			newIter.Push(MakeBioSequenceBatch(0, slice))
		}
		newIter.Done()
	}()

	if iterator.IsPaired() {
		newIter.MarkAsPaired()
	}

	return newIter
}

// It takes a slice of BioSequence objects, and returns an iterator that will return batches of
// BioSequence objects
func IBatchOver(data obiseq.BioSequenceSlice,
	size int, sizes ...int) IBioSequence {

	newIter := MakeIBioSequence()

	newIter.Add(1)

	go func() {
		newIter.WaitAndClose()
	}()

	go func() {
		ldata := len(data)
		batchid := 0
		next := 0
		for i := 0; i < ldata; i = next {
			next = i + size
			if next > ldata {
				next = ldata
			}
			newIter.Push(MakeBioSequenceBatch(batchid, data[i:next]))
			batchid++
		}

		newIter.Done()
	}()

	if data.IsPaired() {
		newIter.MarkAsPaired()
	}
	return newIter
}

// func IBatchOverClasses(data obiseq.BioSequenceSlice,
// 	classifier *obiseq.BioSequenceClassifier) IBioSequence {

// 	newIter := MakeIBioSequence()
// 	classMap := make(map[string]int)
// }
