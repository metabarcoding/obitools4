package obiseq

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func Expression(expression string) func(*BioSequence) (interface{}, error) {

	exp, err := OBILang.NewEvaluable(expression)
	if err != nil {
		log.Fatalf("Error in the expression : %s", expression)
	}

	f := func(sequence *BioSequence) (interface{}, error) {
		return exp(context.Background(),
			map[string]interface{}{
				"annotations": sequence.Annotations(),
				"sequence":    sequence,
			},
		)
	}

	return f
}

func EditIdWorker(expression string) SeqWorker {
	e := Expression(expression)
	f := func(sequence *BioSequence) *BioSequence {
		v, err := e(sequence)

		if err != nil {
			log.Fatalf("Expression '%s' cannot be evaluated on sequence %s",
				expression,
				sequence.Id())
		}
		sequence.SetId(fmt.Sprintf("%v", v))
		return sequence
	}

	return f
}

func EditAttributeWorker(key string, expression string) SeqWorker {
	e := Expression(expression)
	f := func(sequence *BioSequence) *BioSequence {
		v, err := e(sequence)

		if err != nil {
			log.Fatalf("Expression '%s' cannot be evaluated on sequence %s",
				expression,
				sequence.Id())
		}
		sequence.SetAttribute(key, v)
		return sequence
	}

	return f
}
