package obicsv

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"github.com/tevino/abool/v2"

	log "github.com/sirupsen/logrus"
)

type CSVHeader []string
type CSVRecord map[string]interface{}
type CSVRecordBatch struct {
	source string
	data   []CSVRecord
	order  int
}

var NilCSVRecordBatch = CSVRecordBatch{"", nil, -1}

// Structure implementing an iterator over bioseq.BioSequenceBatch
// based on a channel.
type ICSVRecord struct {
	channel         chan CSVRecordBatch
	current         CSVRecordBatch
	pushBack        *abool.AtomicBool
	all_done        *sync.WaitGroup
	lock            *sync.RWMutex
	buffer_size     int32
	batch_size      int32
	sequence_format string
	finished        *abool.AtomicBool
	header          CSVHeader
}

var NilIBioSequenceBatch = (*ICSVRecord)(nil)

func NewICSVRecord() *ICSVRecord {

	i := ICSVRecord{
		channel:         make(chan CSVRecordBatch),
		current:         NilCSVRecordBatch,
		pushBack:        abool.New(),
		batch_size:      -1,
		sequence_format: "",
		finished:        abool.New(),
		header:          make(CSVHeader, 0),
	}

	waiting := sync.WaitGroup{}
	i.all_done = &waiting
	lock := sync.RWMutex{}
	i.lock = &lock

	obiiter.RegisterAPipe()

	return &i
}

func MakeCSVRecordBatch(source string, order int, data []CSVRecord) CSVRecordBatch {
	return CSVRecordBatch{
		source: source,
		order:  order,
		data:   data,
	}
}

func (batch *CSVRecordBatch) Order() int {
	return batch.order
}

func (batch *CSVRecordBatch) Source() string {
	return batch.source
}

func (batch *CSVRecordBatch) Slice() []CSVRecord {
	return batch.data
}

// NotEmpty returns whether the BioSequenceBatch is empty or not.
//
// It checks if the BioSequenceSlice contained within the BioSequenceBatch is not empty.
//
// Returns:
// - bool: True if the BioSequenceBatch is not empty, false otherwise.
func (batch *CSVRecordBatch) NotEmpty() bool {
	return len(batch.data) > 0
}

// IsNil checks if the BioSequenceBatch's slice is nil.
//
// This function takes a BioSequenceBatch as a parameter and returns a boolean value indicating whether the slice of the BioSequenceBatch is nil or not.
//
// Parameters:
// - batch: The BioSequenceBatch to check for nil slice.
//
// Returns:
// - bool: True if the BioSequenceBatch's slice is nil, false otherwise.
func (batch *CSVRecordBatch) IsNil() bool {
	return batch.data == nil
}

func (iterator *ICSVRecord) Add(n int) {
	if iterator == nil {
		log.Panic("call of ICSVRecord.Add method on NilIBioSequenceBatch")
	}

	iterator.all_done.Add(n)
}

func (iterator *ICSVRecord) Done() {
	if iterator == nil {
		log.Panic("call of ICSVRecord.Done method on NilIBioSequenceBatch")
	}

	iterator.all_done.Done()
}

func (iterator *ICSVRecord) Unlock() {
	if iterator == nil {
		log.Panic("call of ICSVRecord.Unlock method on NilIBioSequenceBatch")
	}

	iterator.lock.Unlock()
}

func (iterator *ICSVRecord) Lock() {
	if iterator == nil {
		log.Panic("call of ICSVRecord.Lock method on NilIBioSequenceBatch")
	}

	iterator.lock.Lock()
}

func (iterator *ICSVRecord) RLock() {
	if iterator == nil {
		log.Panic("call of ICSVRecord.RLock method on NilIBioSequenceBatch")
	}

	iterator.lock.RLock()
}

func (iterator *ICSVRecord) RUnlock() {
	if iterator == nil {
		log.Panic("call of ICSVRecord.RUnlock method on NilIBioSequenceBatch")
	}

	iterator.lock.RUnlock()
}

func (iterator *ICSVRecord) Wait() {
	if iterator == nil {
		log.Panic("call of ICSVRecord.Wait method on NilIBioSequenceBatch")
	}

	iterator.all_done.Wait()
}

func (iterator *ICSVRecord) Channel() chan CSVRecordBatch {
	if iterator == nil {
		log.Panic("call of ICSVRecord.Channel method on NilIBioSequenceBatch")
	}

	return iterator.channel
}

func (iterator *ICSVRecord) IsNil() bool {
	if iterator == nil {
		log.Panic("call of ICSVRecord.IsNil method on NilIBioSequenceBatch")
	}

	return iterator == nil
}

func (iterator *ICSVRecord) BatchSize() int {
	if iterator == nil {
		log.Panic("call of ICSVRecord.BatchSize method on NilIBioSequenceBatch")
	}

	return int(atomic.LoadInt32(&iterator.batch_size))
}

func (iterator *ICSVRecord) SetBatchSize(size int) error {
	if size >= 0 {
		atomic.StoreInt32(&iterator.batch_size, int32(size))
		return nil
	}

	return fmt.Errorf("size (%d) cannot be negative", size)
}

func (iterator *ICSVRecord) Split() *ICSVRecord {
	iterator.lock.RLock()
	defer iterator.lock.RUnlock()
	i := ICSVRecord{
		channel:         iterator.channel,
		current:         NilCSVRecordBatch,
		pushBack:        abool.New(),
		all_done:        iterator.all_done,
		buffer_size:     iterator.buffer_size,
		batch_size:      iterator.batch_size,
		sequence_format: iterator.sequence_format,
		finished:        iterator.finished,
		header:          iterator.header,
	}
	lock := sync.RWMutex{}
	i.lock = &lock

	return &i
}

func (iterator *ICSVRecord) Header() CSVHeader {
	return iterator.header
}

func (iterator *ICSVRecord) SetHeader(header CSVHeader) {
	iterator.header = header
}

func (iterator *ICSVRecord) AppendField(field string) {
	iterator.header = append(iterator.header, field)
}

func (iterator *ICSVRecord) Next() bool {
	if iterator.pushBack.IsSet() {
		iterator.pushBack.UnSet()
		return true
	}

	if iterator.finished.IsSet() {
		return false
	}

	next, ok := (<-iterator.channel)

	if ok {
		iterator.current = next
		return true
	}

	iterator.current = NilCSVRecordBatch
	iterator.finished.Set()
	return false
}

func (iterator *ICSVRecord) PushBack() {
	if !iterator.current.IsNil() {
		iterator.pushBack.Set()
	}
}

// The 'Get' method returns the instance of BioSequenceBatch
// currently pointed by the iterator. You have to use the
// 'Next' method to move to the next entry before calling
// 'Get' to retreive the following instance.
func (iterator *ICSVRecord) Get() CSVRecordBatch {
	return iterator.current
}

func (iterator *ICSVRecord) Push(batch CSVRecordBatch) {
	if batch.IsNil() {
		log.Panicln("A Nil batch is pushed on the channel")
	}
	// if batch.Len() == 0 {
	// 	log.Panicln("An empty batch is pushed on the channel")
	// }

	iterator.channel <- batch
}

func (iterator *ICSVRecord) Close() {
	close(iterator.channel)
	obiiter.UnregisterPipe()
}

func (iterator *ICSVRecord) WaitAndClose() {
	iterator.Wait()

	for len(iterator.Channel()) > 0 {
		time.Sleep(time.Millisecond)
	}

	iterator.Close()
}

// Finished returns 'true' value if no more data is available
// from the iterator.
func (iterator *ICSVRecord) Finished() bool {
	return iterator.finished.IsSet()
}

// Sorting the batches of sequences.
func (iterator *ICSVRecord) SortBatches(sizes ...int) *ICSVRecord {

	newIter := NewICSVRecord()

	newIter.Add(1)

	go func() {
		newIter.WaitAndClose()
	}()

	next_to_send := 0
	//log.Println("wait for batch #", next_to_send)
	received := make(map[int]CSVRecordBatch)
	go func() {
		for iterator.Next() {
			batch := iterator.Get()
			// log.Println("\nPushd seq #\n", batch.order, next_to_send)

			if batch.order == next_to_send {
				newIter.channel <- batch
				next_to_send++
				//log.Println("\nwait for batch #\n", next_to_send)
				batch, ok := received[next_to_send]
				for ok {
					newIter.channel <- batch
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

func (iterator *ICSVRecord) Consume() {
	for iterator.Next() {
		iterator.Get()
	}
}
