package obikmer

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// SeqKmerCounter counts, for every canonical k-mer, in how many distinct input
// sequences it occurs. It answers "which k-mers are present in at least q of
// the N sequences" without one k-mer set per sequence: a KmerSetGroup needs
// N*P partition files, which is out of reach for thousands of genomes, whereas
// a SeqKmerCounter uses a single set of P files whatever N is.
//
// Construction is two-phase, like KmerSetGroupBuilder:
//
//  1. AddSequence extracts super-kmers (grouped by minimizer, so a k-mer and its
//     reverse complement always land in the same partition) and appends them,
//     tagged with the sequence id, to one of P temporary files. It is safe to
//     call concurrently.
//  2. CountAtLeast processes the partitions in parallel: each one is loaded,
//     expanded to (canonical k-mer, sequence id) pairs, sorted, and reduced.
//
// Memory during phase 2 is 16 bytes per k-mer occurrence of the partitions
// being processed at the same time. Their number is bounded by an occurrence
// budget (see SetOccurrenceBudget), so that the memory does not depend on how
// the k-mers happen to be spread over the partitions; P only has to be large
// enough for the largest partition to be small against the budget (total bases
// / P is the occurrence count of an average partition, see DefaultPartitions).
//
// Bases other than A, C, G, T, U break a sequence into independent runs: no
// k-mer ever contains an ambiguous base.
type SeqKmerCounter struct {
	dir     string
	k, m, p int

	parts []*seqPartWriter

	mu       sync.Mutex
	firstErr error
	sealed   bool

	budget int // k-mer occurrences held at once in phase 2
}

type seqPartWriter struct {
	mu    sync.Mutex
	w     *bufio.Writer
	file  *os.File
	path  string
	kmers int // k-mer occurrences written, so that the partition is loaded in an array of the exact size
}

// SeqKmerCount is the result for one k-mer.
type SeqKmerCount struct {
	Kmer      uint64 // canonical k-mer
	Sequences uint32 // number of distinct sequences containing it
	Repeated  bool   // occurs more than once (either strand) in at least one sequence
}

// NewSeqKmerCounter creates a counter whose temporary files live in a new
// directory created under parentDir (use "" for the system temporary
// directory). k is the k-mer size (2..31), m the minimizer size (<=0 for the
// default) and partitions the number of partition files P.
func NewSeqKmerCounter(parentDir string, k, m, partitions int) (*SeqKmerCounter, error) {
	if k < 2 || k > 31 {
		return nil, fmt.Errorf("obikmer: k must be between 2 and 31, got %d", k)
	}
	if partitions < 1 {
		return nil, fmt.Errorf("obikmer: partitions must be >= 1, got %d", partitions)
	}
	if m <= 0 {
		m = DefaultMinimizerSize(k)
	}
	if m >= k {
		return nil, fmt.Errorf("obikmer: minimizer size %d must be smaller than k=%d", m, k)
	}

	dir, err := os.MkdirTemp(parentDir, "obikmer-seqcount-")
	if err != nil {
		return nil, err
	}

	c := &SeqKmerCounter{dir: dir, k: k, m: m, p: partitions, budget: DefaultOccurrenceBudget}
	c.parts = make([]*seqPartWriter, partitions)
	for i := range c.parts {
		path := filepath.Join(dir, fmt.Sprintf("part_%04d.sqk", i))
		f, err := os.Create(path)
		if err != nil {
			c.Close()
			return nil, err
		}
		c.parts[i] = &seqPartWriter{w: bufio.NewWriterSize(f, 32*1024), file: f, path: path}
	}
	return c, nil
}

// DefaultOccurrenceBudget is the number of k-mer occurrences phase 2 holds at
// once by default: 32 million, 512 MB. A partition larger than the budget is
// processed alone. Measured on 16 cores, 526 million occurrences to count: with
// 32M the count is as fast as without any limit; 16M make it 15% slower, 8M
// 60% slower.
const DefaultOccurrenceBudget = 32 << 20

// SetOccurrenceBudget bounds the k-mer occurrences (16 bytes each) that
// CountAtLeast holds in memory at once. It must be called before CountAtLeast.
func (c *SeqKmerCounter) SetOccurrenceBudget(occurrences int) {
	if occurrences > 0 {
		c.budget = occurrences
	}
}

// partitionOccurrences is the occurrence count aimed at for an average
// partition by DefaultPartitions (a partition of 1M occurrences is 16 MB).
const partitionOccurrences = 1 << 20

// DefaultPartitions returns a partition count suited to a total of kmers
// k-mer occurrences: 256 at least, and more, as a power of two up to 1024, when
// the average partition would otherwise exceed about a million occurrences.
// The largest partitions hold several times the average. The limit keeps the
// number of simultaneously open files reasonable.
func DefaultPartitions(kmers int) int {
	p := 256
	for p < 1024 && kmers/p > partitionOccurrences {
		p *= 2
	}
	return p
}

// K returns the k-mer size.
func (c *SeqKmerCounter) K() int { return c.k }

var unambiguousBase = func() (t [256]bool) {
	for _, b := range "ACGTUacgtu" {
		t[b] = true
	}
	return
}()

// AddSequence registers sequence id. The same id must be used for every call
// concerning the same sequence; ids need not be contiguous. If circular is
// true the k-1 first bases are appended so that k-mers spanning the junction
// are counted. AddSequence is safe for concurrent use and must not be called
// after CountAtLeast.
func (c *SeqKmerCounter) AddSequence(id uint32, seq []byte, circular bool) {
	if len(seq) < c.k {
		return
	}
	if circular {
		ext := make([]byte, 0, len(seq)+c.k-1)
		ext = append(ext, seq...)
		seq = append(ext, seq[:c.k-1]...)
	}

	var scratch []byte
	start := -1
	for i, b := range seq {
		switch {
		case unambiguousBase[b]:
			if start < 0 {
				start = i
			}
		case start >= 0:
			scratch = c.addRun(id, seq[start:i], scratch)
			start = -1
		}
	}
	if start >= 0 {
		c.addRun(id, seq[start:], scratch)
	}
}

// addRun writes the super-kmers of one ambiguity-free run.
func (c *SeqKmerCounter) addRun(id uint32, run []byte, scratch []byte) []byte {
	if len(run) < c.k {
		return scratch
	}
	for sk := range IterSuperKmers(run, c.k, c.m) {
		scratch = binary.AppendUvarint(scratch[:0], uint64(id))
		scratch = binary.AppendUvarint(scratch, uint64(len(sk.Sequence)))
		scratch = appendPacked(scratch, sk.Sequence)

		pw := c.parts[int(sk.Minimizer%uint64(c.p))]
		pw.mu.Lock()
		_, err := pw.w.Write(scratch)
		pw.kmers += len(sk.Sequence) - c.k + 1
		pw.mu.Unlock()
		if err != nil {
			c.setErr(err)
		}
	}
	return scratch
}

func (c *SeqKmerCounter) setErr(err error) {
	c.mu.Lock()
	if c.firstErr == nil {
		c.firstErr = err
	}
	c.mu.Unlock()
}

// seal flushes and closes the partition files; no more sequences can be added.
func (c *SeqKmerCounter) seal() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sealed {
		return c.firstErr
	}
	c.sealed = true
	for _, pw := range c.parts {
		if err := pw.w.Flush(); err != nil && c.firstErr == nil {
			c.firstErr = err
		}
		if err := pw.file.Close(); err != nil && c.firstErr == nil {
			c.firstErr = err
		}
	}
	return c.firstErr
}

// Close removes the temporary files. It is safe to call more than once.
func (c *SeqKmerCounter) Close() error {
	_ = c.seal()
	return os.RemoveAll(c.dir)
}
