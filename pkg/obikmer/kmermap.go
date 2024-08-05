package obikmer

import (
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"github.com/schollz/progressbar/v3"
)

type KmerMap struct {
	index    map[KmerIdx64][]*obiseq.BioSequence
	kmersize int
	kmermask KmerIdx64
}

type KmerMatch map[*obiseq.BioSequence]int

func (k *KmerMap) KmerSize() int {
	return k.kmersize
}

func (k *KmerMap) Len() int {
	return len(k.index)
}

func (k *KmerMap) Push(sequence *obiseq.BioSequence) {
	current := KmerIdx64(0)
	ccurrent := KmerIdx64(0)
	lshift := uint(2 * (k.kmersize - 1))

	nuc := sequence.Sequence()
	size := 0
	for i := 0; i < len(nuc)-k.kmersize+1; i++ {
		current <<= 2
		ccurrent >>= 2
		code := iupac[nuc[i]]
		ccode := iupac[revcompnuc[nuc[i]]]

		if len(code) != 1 {
			current = KmerIdx64(0)
			ccurrent = KmerIdx64(0)
			size = 0
			continue
		}

		current |= KmerIdx64(code[0])
		ccurrent |= KmerIdx64(ccode[0]) << lshift
		size++

		if size == k.kmersize {

			kmer := min(k.kmermask&current, k.kmermask&ccurrent)
			k.index[kmer] = append(k.index[kmer], sequence)
			size--
		}
	}
}

func (k *KmerMap) Query(sequence *obiseq.BioSequence) KmerMatch {
	current := KmerIdx64(0)
	ccurrent := KmerIdx64(0)

	rep := make(KmerMatch)

	nuc := sequence.Sequence()
	size := 0
	for i := 0; i < len(nuc)-k.kmersize+1; i++ {
		current <<= 2
		ccurrent >>= 2

		code := iupac[nuc[i]]
		ccode := iupac[revcompnuc[nuc[i]]]

		if len(code) != 1 {
			current = KmerIdx64(0)
			ccurrent = KmerIdx64(0)
			size = 0
			continue
		}

		current |= KmerIdx64(code[0])
		ccurrent |= KmerIdx64(ccode[0]) << uint(2*(k.kmersize-1))
		size++

		if size == k.kmersize {
			kmer := min(k.kmermask&current, k.kmermask&ccurrent)
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
			size--
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

func NewKmerMap(sequences obiseq.BioSequenceSlice, kmersize int) *KmerMap {
	idx := make(map[KmerIdx64][]*obiseq.BioSequence)

	kmermask := KmerIdx64(^(^uint64(0) << (uint64(kmersize) * 2)))

	kmap := &KmerMap{kmersize: kmersize, kmermask: kmermask, index: idx}

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

func (k *KmerMap) MakeCountMatchWorker(minKmerCount int) obiseq.SeqWorker {
	return func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		matches := k.Query(sequence)
		matches.FilterMinCount(minKmerCount)
		n := matches.Len()

		sequence.SetAttribute("obikmer_match_count", n)
		return obiseq.BioSequenceSlice{sequence}, nil
	}
}
