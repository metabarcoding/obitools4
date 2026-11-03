package obiprimer

import (
	"cmp"
	"fmt"
	"slices"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
)

// corpusEntry is a sequence read by one of the loading goroutines: either the
// sequence itself, or, when it is packed as it is read, what is kept of it.
type corpusEntry struct {
	batch, index int
	role         Role

	seq *obiseq.BioSequence // unpacked loading

	packed packedSeq // packed loading
	id     string
	taxon  *obitax.Taxon // at the chosen rank, nil when there is none
}

// loadEntries reads every sequence of seqs, once, on all the cores, and returns
// the entries in the order of the input, whatever the number of reader
// goroutines. keep fills the entry of a sequence (besides its position and its
// role); it runs in the reading goroutines.
func loadEntries(seqs obiiter.IBioSequence, sel Selection, keep func(*obiseq.BioSequence, *corpusEntry)) []corpusEntry {
	var mu sync.Mutex
	var entries []corpusEntry

	consumer := func(it obiiter.IBioSequence, wg *sync.WaitGroup) {
		defer wg.Done()
		var local []corpusEntry
		for it.Next() {
			batch := it.Get()
			for i, seq := range batch.Slice() {
				e := corpusEntry{batch: batch.Order(), index: i, role: sel.Role(seq)}
				keep(seq, &e)
				local = append(local, e)
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
	return entries
}

func checkSeqLen(id string, n int) error {
	if n >= maxSeqLen {
		return fmt.Errorf("obiprimer: sequence %q is too long (%d bases, the limit is %d)", id, n, maxSeqLen-1)
	}
	return nil
}

// LoadCorpus reads every sequence of seqs, once, and keeps all of them with the
// role given by the selection. The order of the input is preserved whatever
// the number of reader goroutines.
func LoadCorpus(seqs obiiter.IBioSequence, sel Selection) (*Corpus, error) {
	entries := loadEntries(seqs, sel, func(seq *obiseq.BioSequence, e *corpusEntry) { e.seq = seq })

	c := &Corpus{
		Seqs:  make([]*obiseq.BioSequence, len(entries)),
		Roles: make([]Role, len(entries)),
	}
	for i, e := range entries {
		if err := checkSeqLen(e.seq.Id(), e.seq.Len()); err != nil {
			return nil, err
		}
		c.Seqs[i], c.Roles[i] = e.seq, e.role
	}
	if c.Counts().Examples == 0 {
		return nil, fmt.Errorf("obiprimer: no example sequence selected")
	}
	return c, nil
}

// LoadPackedCorpus is LoadCorpus followed by Pack and AssignTaxa, without ever
// holding the sequences themselves: each one is packed, and its taxon at rank
// looked up, as it is read, and then dropped. The corpus of a large database
// would otherwise first take, in memory, the sequences with all their
// annotations, four times the size of their bases and more.
func LoadPackedCorpus(seqs obiiter.IBioSequence, sel Selection, taxonomy *obitax.Taxonomy, rank string) (*Corpus, *Taxa, error) {
	if err := checkRank(taxonomy, rank); err != nil {
		return nil, nil, err
	}
	entries := loadEntries(seqs, sel, func(seq *obiseq.BioSequence, e *corpusEntry) {
		e.packed = packBases(seq.Sequence())
		e.id = seq.Id()
		e.taxon = seq.Taxon(taxonomy).TaxonAtRank(rank)
	})

	c := &Corpus{
		Seqs:   make([]*obiseq.BioSequence, len(entries)),
		Roles:  make([]Role, len(entries)),
		packed: make([]packedSeq, len(entries)),
		ids:    make([]string, len(entries)),
	}
	taxons := make([]*obitax.Taxon, len(entries))
	for i := range entries {
		e := &entries[i]
		if err := checkSeqLen(e.id, e.packed.n); err != nil {
			return nil, nil, err
		}
		c.Roles[i], c.packed[i], c.ids[i], taxons[i] = e.role, e.packed, e.id, e.taxon
		*e = corpusEntry{} // frees the entry as it goes
	}
	if c.Counts().Examples == 0 {
		return nil, nil, fmt.Errorf("obiprimer: no example sequence selected")
	}
	taxa, err := buildTaxa(rank, taxons, c.Roles)
	if err != nil {
		return nil, nil, err
	}
	return c, taxa, nil
}
