package obiseq

type SequencePredicate func(BioSequence) bool


func (predicate1 SequencePredicate) And(predicate2 SequencePredicate) SequencePredicate {
	f := func(sequence BioSequence) bool {
		return predicate1(sequence) && predicate2(sequence)
	}

	return f
}

func (predicate1 SequencePredicate) Or(predicate2 SequencePredicate) SequencePredicate {
	f := func(sequence BioSequence) bool {
		return predicate1(sequence) || predicate2(sequence)
	}

	return f
}

func (predicate1 SequencePredicate) Xor(predicate2 SequencePredicate) SequencePredicate {
	f := func(sequence BioSequence) bool {
		p1 := predicate1(sequence)
		p2 := predicate2(sequence)
		return (p1 && !p2) || (p2 && !p1)
	}

	return f
}

func (predicate1 SequencePredicate) Not() SequencePredicate {
	f := func(sequence BioSequence) bool {
		return !predicate1(sequence)
	}

	return f
}

func HasAttribute(name string) SequencePredicate {

	f := func(sequence BioSequence) bool {
		if sequence.HasAnnotation() {
			_, ok := (sequence.Annotations())[name]
			return ok
		}

		return false
	}

	return f
}

func MoreAbundantThan(count int) SequencePredicate {
	f := func(sequence BioSequence) bool {
		return sequence.Count() > count
	}

	return f
}

func IsLongerOrEqualTo(length int) SequencePredicate {
	f := func(sequence BioSequence) bool {
		return sequence.Length() >= length
	}

	return f
}

func IsShorterOrEqualTo(length int) SequencePredicate {
	f := func(sequence BioSequence) bool {
		return sequence.Length() <= length
	}

	return f
}

