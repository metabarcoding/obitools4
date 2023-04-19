package obiconsensus

import (
	"fmt"
	"os"
	"path"
	"sort"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obikmer"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obisuffix"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
)

func BuildConsensus(seqs obiseq.BioSequenceSlice,
	kmer_size int, quorum float64,
	min_depth float64,
	save_graph bool, dirname string) (*obiseq.BioSequence, error) {

	log.Printf("Number of reads : %d\n", len(seqs))

	if kmer_size < 0 {
		longest := make([]int, len(seqs))

		for i := range seqs {
			s := seqs[i : i+1]
			sa := obisuffix.BuildSuffixArray(&s)
			longest[i] = obiutils.MaxSlice(sa.CommonSuffix())
		}

		o := obiutils.Order(sort.IntSlice(longest))
		i := int(float64(len(seqs)) * quorum)

		kmer_size = longest[o[i]] + 1
		log.Printf("estimated kmer size : %d", kmer_size)
	}

	graph := obikmer.MakeDeBruijnGraph(kmer_size)

	for _, s := range seqs {
		graph.Push(s)
	}

	log.Printf("Graph size : %d\n", graph.Len())
	total_kmer := graph.Len()

	threshold := 0

	switch {
	case min_depth < 0:
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

		for i, total := range spectrum {
			if total == kmax {
				threshold = i
				break
			}
		}
		threshold /= 2
	case min_depth >= 1:
		threshold = int(min_depth)
	default:
		threshold = int(float64(len(seqs)) * min_depth)
	}

	graph.FilterMin(threshold)
	
	log.Printf("Graph size : %d\n", graph.Len())

	if save_graph {
	
		file, err := os.Create(path.Join(dirname,
			fmt.Sprintf("%s.gml", seqs[0].Source())))
	
		if err != nil {
			fmt.Println(err)
		} else {
			file.WriteString(graph.Gml())
			file.Close()
		}	
	}

	seq, err := graph.LongestConsensus(seqs[0].Source())

	seq.SetCount(len(seqs))
	seq.SetAttribute("seq_length", seq.Len())
	seq.SetAttribute("kmer_size", kmer_size)
	seq.SetAttribute("kmer_min_occur", threshold)
	seq.SetAttribute("kmer_max_occur", graph.MaxLink())
	seq.SetAttribute("filtered_graph_size", graph.Len())
	seq.SetAttribute("full_graph_size", total_kmer)

	return seq, err
}

func Consensus(iterator obiiter.IBioSequence) obiiter.IBioSequence {
	newIter := obiiter.MakeIBioSequence()
	size := 10

	if CLISaveGraphToFiles() {
		dirname := CLIGraphFilesDirectory()
		if stat, err := os.Stat(dirname); err != nil || !stat.IsDir() {
			// path does not exist or is not directory
			os.RemoveAll(dirname)
			err := os.Mkdir(dirname, 0755)
	
			if err != nil {
				log.Panicf("Cannot create directory %s for saving graphs", dirname)
			}
		}	
	}

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

			consensus, err := BuildConsensus(seqs.Slice(),
				CLIKmerSize(), CLIThreshold(),
				CLIKmerDepth(),
				CLISaveGraphToFiles(), CLIGraphFilesDirectory(),
			)

			if err == nil {
				buffer = append(buffer, consensus)
			}

			if len(buffer) == size {
				newIter.Push(obiiter.MakeBioSequenceBatch(order, buffer))
				order++
				buffer = obiseq.MakeBioSequenceSlice()
			}
			seqs.Recycle(true)
		}

		if len(buffer) > 0 {
			newIter.Push(obiiter.MakeBioSequenceBatch(order, buffer))
		}

		newIter.Done()

	}()

	return newIter
}
