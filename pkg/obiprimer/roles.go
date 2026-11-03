// Package obiprimer implements the core of obiprimers, the OBITools4
// re-implementation of ecoPrimers: finding conserved words usable as primers
// in a set of example sequences, and the primer pairs they form.
package obiprimer

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
)

// Role tells how a sequence takes part in the primer search.
type Role uint8

const (
	// Other sequences are non-targets: nothing is required of them, so they
	// count neither in the sensitivity nor in the false positive quorums, but
	// the amplicons they give are taken into account to decide whether a target
	// is identified (see Bs and Bd).
	Other Role = iota
	// Example sequences are the targets: the primers must amplify them.
	Example
	// CounterExample sequences are those the primers must not amplify.
	CounterExample
)

// Selection defines the roles from the taxonomy, with the ecoPrimers
// semantics of -r (restrict), -i (ignore) and -E (exception):
//
//   - a sequence is an Example when it belongs to Restrict (all sequences when
//     Restrict is nil) and not to Except;
//   - otherwise it is a CounterExample when it belongs to Ignore or to Except;
//   - otherwise it is Other.
//
// Nil predicates mean "no such constraint".
type Selection struct {
	Restrict obiseq.SequencePredicate
	Ignore   obiseq.SequencePredicate
	Except   obiseq.SequencePredicate
}

// Role returns the role of seq.
func (s Selection) Role(seq *obiseq.BioSequence) Role {
	inRestrict := s.Restrict == nil || s.Restrict(seq)
	inExcept := s.Except != nil && s.Except(seq)

	switch {
	case inRestrict && !inExcept:
		return Example
	case inExcept, s.Ignore != nil && s.Ignore(seq):
		return CounterExample
	default:
		return Other
	}
}

// BelongingTo returns a predicate true for the sequences whose taxon is a
// sub-clade of at least one of the given taxids. It returns nil (no
// constraint) when taxids is empty.
func BelongingTo(taxonomy *obitax.Taxonomy, taxids []string) (obiseq.SequencePredicate, error) {
	var pred obiseq.SequencePredicate
	for _, id := range taxids {
		taxon, _, err := taxonomy.Taxon(id)
		if err != nil {
			return nil, fmt.Errorf("obiprimer: taxid %q: %w", id, err)
		}
		if taxon == nil {
			return nil, fmt.Errorf("obiprimer: taxid %q is unknown from the taxonomy", id)
		}
		p := obiseq.IsSubCladeOf(taxonomy, taxon)
		if pred == nil {
			pred = p
		} else {
			pred = pred.Or(p)
		}
	}
	return pred, nil
}
