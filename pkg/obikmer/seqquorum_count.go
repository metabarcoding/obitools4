package obikmer

import (
	"bufio"
	"cmp"
	"encoding/binary"
	"io"
	"os"
	"slices"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
)

// kmerOcc is one occurrence of a canonical k-mer in a sequence.
type kmerOcc struct {
	kmer uint64
	seq  uint32
}

func cmpOcc(a, b kmerOcc) int {
	if c := cmp.Compare(a.kmer, b.kmer); c != 0 {
		return c
	}
	return cmp.Compare(a.seq, b.seq)
}

// CountAtLeast returns, sorted by k-mer, every canonical k-mer found in at
// least minSeq distinct sequences, with its exact sequence count.
//
// The result is held in memory: it is meant for quorum-filtered sets (the
// conserved k-mers), not for the complete k-mer spectrum. The first call
// seals the counter; AddSequence must not be called afterwards. The count
// phase uses obidefault.ParallelWorkers() goroutines over the partitions.
func (c *SeqKmerCounter) CountAtLeast(minSeq int) ([]SeqKmerCount, error) {
	if err := c.seal(); err != nil {
		return nil, err
	}
	if minSeq < 1 {
		minSeq = 1
	}

	workers := min(obidefault.ParallelWorkers(), c.p)

	// the largest partitions first: they are the hardest to fit in the budget
	// and the last ones to finish, if they come late
	order := make([]int, c.p)
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int { return cmp.Compare(c.parts[b].kmers, c.parts[a].kmers) })
	budget := newOccurrenceBudget(c.budget)

	jobs := make(chan int)
	results := make([][]SeqKmerCount, c.p)

	var wg sync.WaitGroup
	var errMu sync.Mutex
	var firstErr error

	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var occ []kmerOcc // reused between partitions of similar sizes
			for p := range jobs {
				n := c.parts[p].kmers
				held := budget.acquire(n)
				if cap(occ) < n || cap(occ) > 2*n+1024 {
					occ = make([]kmerOcc, 0, n)
				}
				var err error
				occ, err = c.loadPartition(p, occ[:0])
				if err == nil {
					results[p] = reduceOccurrences(occ, minSeq)
				}
				budget.release(held)
				if err != nil {
					errMu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					errMu.Unlock()
				}
			}
		}()
	}
	for _, p := range order {
		jobs <- p
	}
	close(jobs)
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}

	total := 0
	for _, r := range results {
		total += len(r)
	}
	all := make([]SeqKmerCount, 0, total)
	for _, r := range results {
		all = append(all, r...)
	}
	slices.SortFunc(all, func(a, b SeqKmerCount) int { return cmp.Compare(a.Kmer, b.Kmer) })
	return all, nil
}

// loadPartition appends to occ every (canonical k-mer, sequence id) of the
// partition file.
func (c *SeqKmerCounter) loadPartition(p int, occ []kmerOcc) ([]kmerOcc, error) {
	f, err := os.Open(c.parts[p].path)
	if err != nil {
		return occ, err
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, 64*1024)

	var packed, bases []byte
	for {
		id, err := binary.ReadUvarint(r)
		if err == io.EOF {
			return occ, nil
		}
		if err != nil {
			return occ, err
		}
		n, err := binary.ReadUvarint(r)
		if err != nil {
			return occ, err
		}

		nb := (int(n) + 3) / 4
		packed = slices.Grow(packed[:0], nb)[:nb]
		if _, err := io.ReadFull(r, packed); err != nil {
			return occ, err
		}
		bases = slices.Grow(bases[:0], int(n))[:n]
		unpackBases(bases, packed)

		for kmer := range IterCanonicalKmers(bases, c.k) {
			occ = append(occ, kmerOcc{kmer, uint32(id)})
		}
	}
}

// reduceOccurrences sorts occ and returns the k-mers present in at least
// minSeq distinct sequences. occ is reordered in place.
func reduceOccurrences(occ []kmerOcc, minSeq int) []SeqKmerCount {
	slices.SortFunc(occ, cmpOcc)

	var out []SeqKmerCount
	for i := 0; i < len(occ); {
		kmer := occ[i].kmer
		var seqs uint32
		repeated := false
		j := i
		for ; j < len(occ) && occ[j].kmer == kmer; j++ {
			if j == i || occ[j].seq != occ[j-1].seq {
				seqs++
			} else {
				repeated = true
			}
		}
		if int(seqs) >= minSeq {
			out = append(out, SeqKmerCount{Kmer: kmer, Sequences: seqs, Repeated: repeated})
		}
		i = j
	}
	return out
}

// occurrenceBudget bounds the k-mer occurrences held in memory at once by the
// goroutines that load the partitions.
type occurrenceBudget struct {
	mu    sync.Mutex
	freed *sync.Cond
	total int
	free  int
}

func newOccurrenceBudget(total int) *occurrenceBudget {
	b := &occurrenceBudget{total: total, free: total}
	b.freed = sync.NewCond(&b.mu)
	return b
}

// acquire waits until n occurrences fit in the budget and takes them; a request
// above the whole budget waits for the budget to be free and takes all of it.
// It returns what was taken, to give back with release.
func (b *occurrenceBudget) acquire(n int) int {
	n = min(n, b.total)
	b.mu.Lock()
	for b.free < n {
		b.freed.Wait()
	}
	b.free -= n
	b.mu.Unlock()
	return n
}

func (b *occurrenceBudget) release(n int) {
	b.mu.Lock()
	b.free += n
	b.mu.Unlock()
	b.freed.Broadcast()
}
