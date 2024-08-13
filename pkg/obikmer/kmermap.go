package obikmer

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obifp"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

type KmerMap[T obifp.FPUint[T]] struct {
	index    map[T][]*obiseq.BioSequence
	Kmersize uint
	kmermask T

	leftMask   T
	rightMask  T
	sparseMask T

	SparseAt int
}

type KmerMatch map[*obiseq.BioSequence]int

func (k *KmerMap[T]) KmerSize() uint {
	return k.Kmersize
}

func (k *KmerMap[T]) Len() int {
	return len(k.index)
}

func (k *KmerMap[T]) KmerAsString(kmer T) string {
	buff := make([]byte, k.Kmersize)
	ks := int(k.Kmersize)

	if k.SparseAt >= 0 {
		ks--
	}

	for i, j := 0, int(k.Kmersize)-1; i < ks; i++ {
		code := kmer.And(obifp.From64[T](3)).AsUint64()
		buff[j] = decode[code]
		j--
		if k.SparseAt >= 0 && j == k.SparseAt {
			buff[j] = '#'
			j--
		}
		kmer = kmer.RightShift(2)
	}

	return string(buff)
}

func (k *KmerMap[T]) NormalizedKmerSlice(sequence *obiseq.BioSequence, buff *[]T) []T {

	makeSparseAt := func(kmer T) T {
		if k.SparseAt == -1 {
			return kmer
		}

		return kmer.And(k.leftMask).RightShift(2).Or(kmer.And(k.rightMask))
	}

	normalizedKmer := func(fw, rv T) T {

		if k.SparseAt >= 0 {
			fw = makeSparseAt(fw)
			rv = makeSparseAt(rv)
		}

		if fw.LessThan(rv) {
			return fw
		}

		return rv
	}

	current := obifp.ZeroUint[T]()
	ccurrent := obifp.ZeroUint[T]()
	lshift := uint(2 * (k.Kmersize - 1))

	sup := sequence.Len() - int(k.Kmersize) + 1

	var rep []T
	if buff == nil {
		rep = make([]T, 0, sup)
	} else {
		rep = (*buff)[:0]
	}

	nuc := sequence.Sequence()

	size := 0
	for i := 0; i < len(nuc); i++ {
		current = current.LeftShift(2)
		ccurrent = ccurrent.RightShift(2)

		code := iupac[nuc[i]]
		ccode := iupac[revcompnuc[nuc[i]]]

		if len(code) != 1 {
			current = obifp.ZeroUint[T]()
			ccurrent = obifp.ZeroUint[T]()
			size = 0
			continue
		}

		current = current.Or(obifp.From64[T](uint64(code[0])))
		ccurrent = ccurrent.Or(obifp.From64[T](uint64(ccode[0])).LeftShift(lshift))

		size++

		if size == int(k.Kmersize) {

			kmer := normalizedKmer(current, ccurrent)
			rep = append(rep, kmer)
			size--
		}
	}

	return rep
}

func (k *KmerMap[T]) Push(sequence *obiseq.BioSequence) {
	kmers := k.NormalizedKmerSlice(sequence, nil)
	for _, kmer := range kmers {
		k.index[kmer] = append(k.index[kmer], sequence)
	}
}

func (k *KmerMap[T]) Query(sequence *obiseq.BioSequence) KmerMatch {
	kmers := k.NormalizedKmerSlice(sequence, nil)
	rep := make(KmerMatch)

	for _, kmer := range kmers {
		if _, ok := k.index[kmer]; ok {
			for _, seq := range k.index[kmer] {
				if seq != sequence {
					if _, ok := rep[seq]; !ok {
						rep[seq] = 0
					}
					rep[seq]++
				}
			}
		}
	}

	return rep
}

func (k *KmerMatch) FilterMinCount(mincount int) {
	for seq, count := range *k {
		if count < mincount {
			delete(*k, seq)
		}
	}
}

func (k *KmerMatch) Len() int {
	return len(*k)
}

func (k *KmerMatch) Sequences() obiseq.BioSequenceSlice {
	ks := make([]*obiseq.BioSequence, 0, len(*k))

	for seq := range *k {
		ks = append(ks, seq)
	}

	return ks
}

func (k *KmerMatch) Max() *obiseq.BioSequence {
	max := 0
	var maxseq *obiseq.BioSequence
	for seq, n := range *k {
		if max < n {
			max = n
			maxseq = seq
		}
	}
	return maxseq
}

func NewKmerMap[T obifp.FPUint[T]](
	sequences obiseq.BioSequenceSlice,
	kmersize uint,
	sparse bool) *KmerMap[T] {
	idx := make(map[T][]*obiseq.BioSequence)

	sparseAt := -1

	if sparse && kmersize%2 == 0 {
		log.Warnf("Kmer size must be odd when using sparse mode")
		kmersize++
	}

	if !sparse && kmersize%2 == 1 {
		log.Warnf("Kmer size must be even when not using sparse mode")
		kmersize--

	}

	if sparse {
		sparseAt = int(kmersize / 2)
	}

	kmermask := obifp.OneUint[T]().LeftShift(kmersize * 2).Sub(obifp.OneUint[T]())
	leftMask := obifp.ZeroUint[T]()
	rightMask := obifp.ZeroUint[T]()

	if sparseAt >= 0 {
		if sparseAt >= int(kmersize) {
			sparseAt = -1
		} else {
			pos := kmersize - 1 - uint(sparseAt)
			left := uint(sparseAt) * 2
			right := pos * 2

			leftMask = obifp.OneUint[T]().LeftShift(left).Sub(obifp.OneUint[T]()).LeftShift(right + 2)
			rightMask = obifp.OneUint[T]().LeftShift(right).Sub(obifp.OneUint[T]())
		}
	}

	kmap := &KmerMap[T]{
		Kmersize:  kmersize,
		kmermask:  kmermask,
		leftMask:  leftMask,
		rightMask: rightMask,
		index:     idx,
		SparseAt:  sparseAt,
	}

	n := len(sequences)
	pbopt := make([]progressbar.Option, 0, 5)
	pbopt = append(pbopt,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetDescription("Indexing kmers"),
	)

	bar := progressbar.NewOptions(n, pbopt...)

	for i, sequence := range sequences {
		kmap.Push(sequence)
		if i%100 == 0 {
			bar.Add(100)
		}
	}

	return kmap
}
