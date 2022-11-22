package obiseq

import (
	"context"
	"fmt"
	"regexp"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obieval"
	log "github.com/sirupsen/logrus"
)

type SequencePredicate func(*BioSequence) bool

func (predicate1 SequencePredicate) And(predicate2 SequencePredicate) SequencePredicate {
	switch {
	case predicate1 == nil:
		return predicate2
	case predicate2 == nil:
		return predicate1
	default:
		return func(sequence *BioSequence) bool {
			if !predicate1(sequence) {
				return false
			}

			return predicate2(sequence)
		}
	}
}

func (predicate1 SequencePredicate) Or(predicate2 SequencePredicate) SequencePredicate {
	switch {
	case predicate1 == nil:
		return predicate2
	case predicate2 == nil:
		return predicate1
	default:
		return func(sequence *BioSequence) bool {
			if predicate1(sequence) {
				return true
			}
			return predicate2(sequence)
		}
	}
}

func (predicate1 SequencePredicate) Xor(predicate2 SequencePredicate) SequencePredicate {
	switch {
	case predicate1 == nil:
		return predicate2
	case predicate2 == nil:
		return predicate1
	default:
		return func(sequence *BioSequence) bool {
			p1 := predicate1(sequence)
			p2 := predicate2(sequence)
			return (p1 && !p2) || (p2 && !p1)
		}
	}
}

func (predicate1 SequencePredicate) Not() SequencePredicate {
	switch {
	case predicate1 == nil:
		return nil
	default:
		return func(sequence *BioSequence) bool {
			return !predicate1(sequence)
		}
	}
}

func HasAttribute(name string) SequencePredicate {

	f := func(sequence *BioSequence) bool {
		if sequence.HasAnnotation() {
			_, ok := (sequence.Annotations())[name]
			return ok
		}

		return false
	}

	return f
}

func IsAttributeMatch(name string, pattern string) SequencePredicate {
	pat, err := regexp.Compile(pattern)

	if err != nil {
		log.Fatalf("error in atribute %s regular pattern syntax : %v", name, err)
	}

	f := func(sequence *BioSequence) bool {
		if sequence.HasAnnotation() {
			val, ok := (sequence.Annotations())[name]
			if ok {
				switch val := val.(type) {
				case string:
					return pat.MatchString(val)
				default:
					return pat.MatchString(fmt.Sprint(val))
				}
			}
		}

		return false
	}

	return f
}

func IsMoreAbundantOrEqualTo(count int) SequencePredicate {
	f := func(sequence *BioSequence) bool {
		return sequence.Count() >= count
	}

	return f
}

func IsLessAbundantOrEqualTo(count int) SequencePredicate {
	f := func(sequence *BioSequence) bool {
		return sequence.Count() <= count
	}

	return f
}

func IsLongerOrEqualTo(length int) SequencePredicate {
	f := func(sequence *BioSequence) bool {
		return sequence.Len() >= length
	}

	return f
}

func IsShorterOrEqualTo(length int) SequencePredicate {
	f := func(sequence *BioSequence) bool {
		return sequence.Len() <= length
	}

	return f
}

func IsSequenceMatch(pattern string) SequencePredicate {
	pat, err := regexp.Compile("(?i)" + pattern)

	if err != nil {
		log.Fatalf("error in sequence regular pattern syntax : %v", err)
	}

	f := func(sequence *BioSequence) bool {
		return pat.Match(sequence.Sequence())
	}

	return f
}

func IsDefinitionMatch(pattern string) SequencePredicate {
	pat, err := regexp.Compile(pattern)

	if err != nil {
		log.Fatalf("error in definition regular pattern syntax : %v", err)
	}

	f := func(sequence *BioSequence) bool {
		return pat.MatchString(sequence.Definition())
	}

	return f
}

func IsIdMatch(pattern string) SequencePredicate {
	pat, err := regexp.Compile(pattern)

	if err != nil {
		log.Fatalf("error in identifier regular pattern syntax : %v", err)
	}

	f := func(sequence *BioSequence) bool {
		return pat.MatchString(sequence.Id())
	}

	return f
}

func IsIdIn(ids ...string) SequencePredicate {
	idset := make(map[string]bool)

	for _, v := range ids {
		idset[v] = true
	}

	f := func(sequence *BioSequence) bool {
		_, ok := idset[sequence.Id()]
		return ok
	}

	return f
}

func ExpressionPredicat(expression string) SequencePredicate {

	exp, err := obieval.OBILang.NewEvaluable(expression)
	if err != nil {
		log.Fatalf("Error in the expression : %s", expression)
	}

	f := func(sequence *BioSequence) bool {
		value, err := exp.EvalBool(context.Background(),
			map[string]interface{}{
				"annotations":     sequence.Annotations(),
				"sequence":  sequence,
			},
		)

		if err != nil {
			log.Fatalf("Expression '%s' cannot be evaluated on sequence %s",
				expression,
				sequence.Id())
		}

		return value
	}

	return f
}
