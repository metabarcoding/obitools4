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
	Ignored         int
}

// Corpus is the set of sequences used by the primer search: the examples and
// the counterexamples, in input order. Sequences with the Ignored role are not
// kept (as in ecoPrimers), only counted.
//
// The primer search needs several passes over the sequences (words are
// searched approximately in every one of them), so the corpus lives in memory.
type Corpus struct {
	Seqs    []*obiseq.BioSequence
	Roles   []Role // Roles[i] is Example or CounterExample
	Ignored int    // number of sequences dropped by the selection
}

// Counts returns the number of sequences by role.
func (c *Corpus) Counts() Counts {
	n := Counts{Ignored: c.Ignored}
	for _, r := range c.Roles {
		if r == Example {
			n.Examples++
		} else {
			n.CounterExamples++
		}
	}
	return n
}

type corpusEntry struct {
	batch, index int
	seq          *obiseq.BioSequence
	role         Role
}

// LoadCorpus reads every sequence of seqs, once, and keeps those selected as
// examples or counterexamples. The order of the input is preserved whatever
// the number of reader goroutines.
func LoadCorpus(seqs obiiter.IBioSequence, sel Selection) (*Corpus, error) {
	var mu sync.Mutex
	var entries []corpusEntry
	ignored := 0

	consumer := func(it obiiter.IBioSequence, wg *sync.WaitGroup) {
		defer wg.Done()
		var local []corpusEntry
		localIgnored := 0
		for it.Next() {
			batch := it.Get()
			for i, seq := range batch.Slice() {
				if role := sel.Role(seq); role == Ignored {
					localIgnored++
				} else {
					local = append(local, corpusEntry{batch.Order(), i, seq, role})
				}
			}
		}
		mu.Lock()
		entries = append(entries, local...)
		ignored += localIgnored
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
		Seqs:    make([]*obiseq.BioSequence, len(entries)),
		Roles:   make([]Role, len(entries)),
		Ignored: ignored,
	}
	for i, e := range entries {
		c.Seqs[i], c.Roles[i] = e.seq, e.role
	}
	if c.Counts().Examples == 0 {
		return nil, fmt.Errorf("obiprimer: no example sequence selected")
	}
	return c, nil
}
