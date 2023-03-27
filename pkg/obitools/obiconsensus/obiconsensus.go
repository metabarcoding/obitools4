package obiconsensus

import (
	"sort"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obikmer"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obisuffix"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
)

func BuildConsensus(seqs obiseq.BioSequenceSlice, quorum float64) (*obiseq.BioSequence, error) {

	log.Printf("Number of reads : %d\n", len(seqs))

	longest := make([]int, len(seqs))

	for i := range seqs {
		s := seqs[i : i+1]
		sa := obisuffix.BuildSuffixArray(&s)
		longest[i] = obiutils.MaxSlice(sa.CommonSuffix())
	}

	o := obiutils.Order(sort.IntSlice(longest))
	i := int(float64(len(seqs)) * quorum)

	kmersize := longest[o[i]] + 1
	log.Printf("estimated kmer size : %d", kmersize)

	graph := obikmer.MakeDeBruijnGraph(kmersize)

	for _, s := range seqs {
		graph.Push(s)
	}

	log.Printf("Graph size : %d\n", graph.Len())
	total_kmer := graph.Len()
	spectrum := graph.LinkSpectrum()
	cum := make(map[int]int)

	spectrum[1] = 0
	for i := 2; i < len(spectrum); i++ {
		spectrum[i] += spectrum[i-1]
		cum[spectrum[i]]++
	}

	max := 0
	kmax := 0
	for k, obs := range cum {
		if obs > max {
			max = obs
			kmax = k
		}
	}

	threshold := 0
	for i, total := range spectrum {
		if total == kmax {
			threshold = i
			break
		}
	}
	threshold /= 2
	graph.FilterMin(threshold)
	log.Printf("Graph size : %d\n", graph.Len())

	// file, err := os.Create(
	// 	fmt.Sprintf("%s.gml", seqs[0].Source()))

	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	file.WriteString(graph.GML())
	// 	file.Close()
	// }

	seq, err := graph.LongestConsensus(seqs[0].Source())

	seq.SetCount(len(seqs))
	seq.SetAttribute("seq_length", seq.Len())
	seq.SetAttribute("kmer_size", kmersize)
	seq.SetAttribute("kmer_min_occur", threshold)
	seq.SetAttribute("kmer_max_occur", graph.MaxLink())
	seq.SetAttribute("filtered_graph_size", graph.Len())
	seq.SetAttribute("full_graph_size", total_kmer)

	return seq, err
}

func Consensus(iterator obiiter.IBioSequence, quorum float64) obiiter.IBioSequence {
	newIter := obiiter.MakeIBioSequence()
	size := 10

	newIter.Add(1)

	go func() {
		newIter.WaitAndClose()
	}()

	go func() {
		order := 0
		iterator = iterator.SortBatches()
		buffer := obiseq.MakeBioSequenceSlice()

		for iterator.Next() {
			seqs := iterator.Get()
			consensus, err := BuildConsensus(seqs.Slice(), quorum)

			if err == nil {
				buffer = append(buffer, consensus)
			}

			if len(buffer) == size {
				newIter.Push(obiiter.MakeBioSequenceBatch(order, buffer))
				order++
				buffer = obiseq.MakeBioSequenceSlice()
			}
			seqs.Recycle()
		}

		if len(buffer) > 0 {
			newIter.Push(obiiter.MakeBioSequenceBatch(order, buffer))
		}

		newIter.Done()

	}()

	return newIter
}
