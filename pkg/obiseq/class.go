package obiseq

import (
	"fmt"
	"hash/crc32"
	"strconv"
	"sync"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
	log "github.com/sirupsen/logrus"
)

// Defines an object able to classify a sequence in classes defined by an integer index.
//
// The first function is the classifier itself. It takes a BioSequence and returns
// an integer. The integer is the class of the BioSequence.
//
// The second function is the classifier's value function. It takes an integer and
// returns a string. The string is the original value used to define the class of the sequence.
//
// Moreover a third function resets the classifier, and fourth one
// returns a clone of the classifier.
type BioSequenceClassifier struct {
	Code  func(*BioSequence) int
	Value func(int) string
	Reset func()
	Clone func() *BioSequenceClassifier
	Type  string
}

// It creates a classifier that returns the value of the annotation key as an integer. If the
// annotation key is not present, it returns the integer value of the string na
func AnnotationClassifier(key string, na string) *BioSequenceClassifier {
	encode := make(map[string]int, 1000)
	decode := make([]string, 0, 1000)
	locke := sync.RWMutex{}
	maxcode := 0

	code := func(sequence *BioSequence) int {
		var val = na
		var ok bool
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

	c := BioSequenceClassifier{code, value, reset, clone, "AnnotationClassifier"}
	return &c
}

// It creates a classifier that returns the value of the annotation key as an integer. If the
// annotation key is not present, it returns the integer value of the string na
func DualAnnotationClassifier(key1, key2 string, na string) *BioSequenceClassifier {
	encode := make(map[string]int, 1000)
	decode := make([]string, 0, 1000)
	locke := sync.RWMutex{}
	maxcode := 0

	code := func(sequence *BioSequence) int {
		var val1 = na
		var val2 = ""
		var ok bool
		if sequence.HasAnnotation() {
			value, ok := sequence.Annotations()[key1]
			if ok {
				switch value := value.(type) {
				case string:
					val1 = value
				default:
					val1 = fmt.Sprint(value)
				}
			}

			if key2 != "" {
				value, ok := sequence.Annotations()[key2]
				if ok {
					switch value := value.(type) {
					case string:
						val2 = value
					default:
						val2 = fmt.Sprint(value)
					}
				} else {
					val2 = na
				}
			}
		}

		locke.Lock()
		defer locke.Unlock()

		jb, _ := goutils.JsonMarshal([2]string{val1, val2})
		json := string(jb)
		k, ok := encode[json]

		if !ok {
			k = maxcode
			maxcode++
			encode[json] = k
			decode = append(decode, json)
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
		return DualAnnotationClassifier(key1, key2, na)
	}

	c := BioSequenceClassifier{code, value, reset, clone, "DualAnnotationClassifier"}
	return &c
}

// It takes a predicate function and returns a classifier that returns 1 if the predicate is true and 0
// otherwise
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

	c := BioSequenceClassifier{code, value, reset, clone, "PredicateClassifier"}
	return &c
}

// Builds a classifier function based on CRC32 of the sequence
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

	c := BioSequenceClassifier{code, value, reset, clone, "HashClassifier"}
	return &c
}

// Builds a classifier function based on the sequence
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

	c := BioSequenceClassifier{code, value, reset, clone, "SequenceClassifier"}
	return &c
}

// It returns a classifier that assigns each sequence to a different class, cycling through the classes
// in order
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

	c := BioSequenceClassifier{code, value, reset, clone, "RotateClassifier"}
	return &c
}
