package obiseq

import (
	"fmt"
	"hash/crc32"
	log "github.com/sirupsen/logrus"
	"strconv"
	"sync"
)

type BioSequenceClassifier struct {
	Code  func(*BioSequence) int
	Value func(int) string
	Reset func()
	Clone func() *BioSequenceClassifier
}

func AnnotationClassifier(key string, na string) *BioSequenceClassifier {
	encode := make(map[string]int, 1000)
	decode := make([]string, 0, 1000)
	locke := sync.RWMutex{}
	maxcode := 0

	code := func(sequence *BioSequence) int {
		var val string
		if sequence.HasAnnotation() {
			value, ok := sequence.Annotations()[key]

			if ok {
				switch value := value.(type) {
				case string:
					val = value
				default:
					val = fmt.Sprint(value)
				}
			}
		}
		val = na

		locke.Lock()
		defer locke.Unlock()

		k, ok := encode[val]

		if !ok {
			k = maxcode
			maxcode++
			encode[val] = k
			decode = append(decode, val)
		}

		return k
	}

	value := func(k int) string {

		locke.RLock()
		defer locke.RUnlock()
		if k >= maxcode {
			log.Fatalf("value %d not register")
		}
		return decode[k]
	}

	reset := func() {
		locke.Lock()
		defer locke.Unlock()

		for k := range encode {
			delete(encode, k)
		}
		decode = decode[:0]
	}

	clone := func() *BioSequenceClassifier {
		return AnnotationClassifier(key, na)
	}

	c := BioSequenceClassifier{code, value, reset, clone}
	return &c
}

func PredicateClassifier(predicate SequencePredicate) *BioSequenceClassifier {
	code := func(sequence *BioSequence) int {
		if predicate(sequence) {
			return 1
		} else {
			return 0
		}

	}

	value := func(k int) string {
		if k == 0 {
			return "false"
		} else {
			return "true"
		}

	}

	reset := func() {

	}

	clone := func() *BioSequenceClassifier {
		return PredicateClassifier(predicate)
	}

	c := BioSequenceClassifier{code, value, reset, clone}
	return &c
}

// Builds a classifier function based on CRC32 of the sequence
//
func HashClassifier(size int) *BioSequenceClassifier {
	code := func(sequence *BioSequence) int {
		return int(crc32.ChecksumIEEE(sequence.Sequence()) % uint32(size))
	}

	value := func(k int) string {
		return strconv.Itoa(k)
	}

	reset := func() {

	}

	clone := func() *BioSequenceClassifier {
		return HashClassifier(size)
	}

	c := BioSequenceClassifier{code, value, reset, clone}
	return &c
}

// Builds a classifier function based on the sequence
//
func SequenceClassifier() *BioSequenceClassifier {
	encode := make(map[string]int, 1000)
	decode := make([]string, 0, 1000)
	locke := sync.RWMutex{}
	maxcode := 0

	code := func(sequence *BioSequence) int {
		val := sequence.String()

		locke.Lock()
		defer locke.Unlock()

		k, ok := encode[val]

		if !ok {
			k = maxcode
			maxcode++
			encode[val] = k
			decode = append(decode, val)
		}

		return k
	}

	value := func(k int) string {
		locke.RLock()
		defer locke.RUnlock()

		if k >= maxcode {
			log.Fatalf("value %d not register")
		}
		return decode[k]
	}

	reset := func() {
		locke.Lock()
		defer locke.Unlock()

		// for k := range encode {
		// 	delete(encode, k)
		// }
		encode = make(map[string]int)
		decode = decode[:0]
		maxcode = 0
	}

	clone := func() *BioSequenceClassifier {
		return SequenceClassifier()
	}

	c := BioSequenceClassifier{code, value, reset, clone}
	return &c
}

func RotateClassifier(size int) *BioSequenceClassifier {
	n := 0
	lock := sync.Mutex{}

	code := func(sequence *BioSequence) int {
		lock.Lock()
		defer lock.Unlock()
		n = n % size
		n++
		return n
	}

	value := func(k int) string {
		return strconv.Itoa(k)
	}

	reset := func() {

	}

	clone := func() *BioSequenceClassifier {
		return RotateClassifier(size)
	}

	c := BioSequenceClassifier{code, value, reset, clone}
	return &c
}
