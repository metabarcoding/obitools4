package obiseq

import (
	"fmt"
	"hash/crc32"
	"strconv"
)

type BioSequenceClassifier func(sequence BioSequence) string

func AnnotationClassifier(key string, na string) BioSequenceClassifier {
	f := func(sequence BioSequence) string {
		if sequence.HasAnnotation() {
			value, ok := sequence.Annotations()[key]

			if ok {
				switch value := value.(type) {
				case string:
					return value
				default:
					return fmt.Sprint(value)
				}
			} 
		}
		return na
	}

	return f
}

func PredicateClassifier(predicate SequencePredicate) BioSequenceClassifier {
	f := func(sequence BioSequence) string {
		if predicate(sequence) {
			return "true"
		} else {
			return "false"
		}
	}

	return f
}

// Builds a classifier function based on CRC32 of the sequence
//
func HashClassifier(size int) BioSequenceClassifier {
	f := func(sequence BioSequence) string {
		h := crc32.ChecksumIEEE(sequence.Sequence()) % uint32(size)
		return strconv.Itoa(int(h))
	}

	return f
}

// Builds a classifier function based on the sequence
//
func SequenceClassifier() BioSequenceClassifier {
	f := func(sequence BioSequence) string {
		return sequence.String()
	}

	return f
}

func RotateClassifier(size int) BioSequenceClassifier {
	n := 0
	f := func(sequence BioSequence) string {
		h := n % size
		n++
		return strconv.Itoa(int(h))
	}

	return f
}
