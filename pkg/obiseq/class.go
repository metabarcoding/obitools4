package obiseq

import (
	"fmt"
	"hash/crc32"
	"strconv"
)

type SequenceClassifier func(sequence BioSequence) string

func AnnotationClassifier(key string) SequenceClassifier {
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
		return ""
	}

	return f
}

var SampleClassifier = AnnotationClassifier("sample")

func PredicateClassifier(predicate SequencePredicate) SequenceClassifier {
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
func HashClassifier(size int) SequenceClassifier {
	f := func(sequence BioSequence) string {
		h := crc32.ChecksumIEEE(sequence.Sequence()) % uint32(size)
		return strconv.Itoa(int(h))
	}

	return f
}

func RotateClassifier(size int) SequenceClassifier {
	n := 0
	f := func(sequence BioSequence) string {
		h := n % size
		n++
		return strconv.Itoa(int(h))
	}

	return f
}
