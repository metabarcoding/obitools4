package obiprimer

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// maxSeqLen bounds the length of a sequence: the position of a primer site and
// its strand share 32 bits (see seqMatch).
const maxSeqLen = 1 << 31

// Counts gives the number of sequences by role.
type Counts struct {
	Examples        int
	CounterExamples int
	Others          int
}

// Corpus is the set of sequences of the primer search, in input order, each
// with its role: the examples (the targets), the counterexamples, and all the
// other sequences, which are loaded as well since the amplicons they give
// decide whether a target is identified.
//
// The primer search needs several passes over the sequences (words are
// searched approximately in every one of them), so the corpus lives in memory.
type Corpus struct {
	Seqs  []*obiseq.BioSequence // nil entries once the corpus is packed
	Roles []Role                // Roles[i] is the role of Seqs[i]

	packed []packedSeq // see Pack
	ids    []string
}

// Counts returns the number of sequences by role.
func (c *Corpus) Counts() Counts {
	var n Counts
	for _, r := range c.Roles {
		switch r {
		case Example:
			n.Examples++
		case CounterExample:
			n.CounterExamples++
		default:
			n.Others++
		}
	}
	return n
}

// IndexOfExample returns the index of the example sequence with the given
// identifier. It is an error if there is none, or if the sequence is a
// counterexample.
func (c *Corpus) IndexOfExample(id string) (int, error) {
	for i := range c.Roles {
		if c.ID(i) == id {
			if c.Roles[i] != Example {
				return -1, fmt.Errorf("obiprimer: reference sequence %q is a counterexample, it must be an example", id)
			}
			return i, nil
		}
	}
	return -1, fmt.Errorf("obiprimer: reference sequence %q is not among the selected example sequences", id)
}

// Len returns the number of sequences.
func (c *Corpus) Len() int { return len(c.Roles) }

// SeqLen returns the length of sequence i.
func (c *Corpus) SeqLen(i int) int {
	if c.Seqs[i] != nil {
		return c.Seqs[i].Len()
	}
	return c.packed[i].n
}

// ID returns the identifier of sequence i.
func (c *Corpus) ID(i int) string {
	if c.Seqs[i] != nil {
		return c.Seqs[i].Id()
	}
	return c.ids[i]
}

// bases returns the bases of sequence i, lower case. A packed sequence is
// unpacked into buf, grown if need be, which the caller keeps for the next call
// (the second result); an unpacked one is returned as it is. The result must
// not be modified, and is valid until the next call with the same buffer.
func (c *Corpus) bases(i int, buf []byte) (raw, newBuf []byte) {
	if c.Seqs[i] != nil {
		return c.Seqs[i].Sequence(), buf
	}
	buf = c.packed[i].unpack(buf)
	return buf, buf
}

// Pack replaces the sequences by their packed form, on 2 bits per base, and
// frees the sequence objects (their annotations, their slack): the primer search
// only needs the bases, the identifiers, and the taxa, which must have been
// assigned before (AssignTaxa). A packed corpus holds about a quarter of the
// memory of the bases themselves.
func (c *Corpus) Pack() {
	if c.packed != nil {
		return
	}
	c.packed = make([]packedSeq, len(c.Seqs))
	c.ids = make([]string, len(c.Seqs))
	parallelFor(len(c.Seqs), func(_, i int) {
		c.packed[i] = packBases(c.Seqs[i].Sequence())
		c.ids[i] = c.Seqs[i].Id()
		c.Seqs[i] = nil
	})
}
