package obiprimer

import (
	"cmp"
	"fmt"
	"slices"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

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
	Seqs  []*obiseq.BioSequence
	Roles []Role // Roles[i] is the role of Seqs[i]
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

type corpusEntry struct {
	batch, index int
	seq          *obiseq.BioSequence
	role         Role
}

// LoadCorpus reads every sequence of seqs, once, and keeps all of them with the
// role given by the selection. The order of the input is preserved whatever
// the number of reader goroutines.
func LoadCorpus(seqs obiiter.IBioSequence, sel Selection) (*Corpus, error) {
	var mu sync.Mutex
	var entries []corpusEntry

	consumer := func(it obiiter.IBioSequence, wg *sync.WaitGroup) {
		defer wg.Done()
		var local []corpusEntry
		for it.Next() {
			batch := it.Get()
			for i, seq := range batch.Slice() {
				local = append(local, corpusEntry{batch.Order(), i, seq, sel.Role(seq)})
			}
		}
		mu.Lock()
		entries = append(entries, local...)
		mu.Unlock()
	}

	var wg sync.WaitGroup
	for i := 1; i < obidefault.ParallelWorkers(); i++ {
		wg.Add(1)
		go consumer(seqs.Split(), &wg)
	}
	wg.Add(1)
	go consumer(seqs, &wg)
	wg.Wait()

	slices.SortFunc(entries, func(a, b corpusEntry) int {
		if c := cmp.Compare(a.batch, b.batch); c != 0 {
			return c
		}
		return cmp.Compare(a.index, b.index)
	})

	c := &Corpus{
		Seqs:  make([]*obiseq.BioSequence, len(entries)),
		Roles: make([]Role, len(entries)),
	}
	for i, e := range entries {
		c.Seqs[i], c.Roles[i] = e.seq, e.role
	}
	if c.Counts().Examples == 0 {
		return nil, fmt.Errorf("obiprimer: no example sequence selected")
	}
	return c, nil
}

// IndexOfExample returns the index of the example sequence with the given
// identifier. It is an error if there is none, or if the sequence is a
// counterexample.
func (c *Corpus) IndexOfExample(id string) (int, error) {
	for i, s := range c.Seqs {
		if s.Id() == id {
			if c.Roles[i] != Example {
				return -1, fmt.Errorf("obiprimer: reference sequence %q is a counterexample, it must be an example", id)
			}
			return i, nil
		}
	}
	return -1, fmt.Errorf("obiprimer: reference sequence %q is not among the selected example sequences", id)
}
