package obiconsensus

import (
	"fmt"
	"os"
	"path"
	"slices"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obisuffix"
)

func BuildConsensus(seqs obiseq.BioSequenceSlice,
	consensus_id string,
	kmer_size int,
	save_graph bool, dirname string) (*obiseq.BioSequence, error) {

	if save_graph {
		if dirname == "" {
			dirname = "."
		}

		if stat, err := os.Stat(dirname); err != nil || !stat.IsDir() {
			// path does not exist or is not directory
			os.RemoveAll(dirname)
			err := os.Mkdir(dirname, 0755)

			if err != nil {
				log.Panicf("Cannot create directory %s for saving graphs", dirname)
			}
		}

		fasta, err := os.Create(path.Join(dirname, fmt.Sprintf("%s_consensus.fasta", consensus_id)))

		if err == nil {
			defer fasta.Close()
			fasta.Write(obiformats.FormatFastaBatch(obiiter.MakeBioSequenceBatch(0, seqs), obiformats.FormatFastSeqJsonHeader, false))
			fasta.Close()
		}

	}

	log.Printf("Number of reads : %d\n", len(seqs))

	if kmer_size < 0 {
		longest := make([]int, len(seqs))

		for i, seq := range seqs {
			s := obiseq.BioSequenceSlice{seq}
			sa := obisuffix.BuildSuffixArray(&s)
			longest[i] = slices.Max(sa.CommonSuffix())
		}

		kmer_size = slices.Max(longest) + 1
		log.Printf("estimated kmer size : %d", kmer_size)
	}

	var graph *obikmer.DeBruijnGraph
	for {
		graph = obikmer.MakeDeBruijnGraph(kmer_size)

		for _, s := range seqs {
			graph.Push(s)
		}

		if !graph.HasCycle() {
			break
		}

		kmer_size++
		log.Infof("Cycle detected, increasing kmer size to %d\n", kmer_size)
	}

	if save_graph {

		file, err := os.Create(path.Join(dirname,
			fmt.Sprintf("%s_consensus.gml", consensus_id)))

		if err != nil {
			fmt.Println(err)
		} else {
			file.WriteString(graph.Gml())
			file.Close()
		}
	}

	log.Printf("Graph size : %d\n", graph.Len())
	total_kmer := graph.Len()

	seq, err := graph.LongestConsensus(consensus_id)

	sumCount := 0

	if seq != nil {
		for _, s := range seqs {
			sumCount += s.Count()
		}

		seq.SetCount(sumCount)
		seq.SetAttribute("seq_length", seq.Len())
		seq.SetAttribute("kmer_size", kmer_size)
		seq.SetAttribute("kmer_max_occur", graph.MaxWeight())
		seq.SetAttribute("filtered_graph_size", graph.Len())
		seq.SetAttribute("full_graph_size", total_kmer)
	}
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

			sequences := seqs.Slice()

			id := sequences[0].Source()
			if id == "" {
				id = sequences[0].Id()
			}
			consensus, err := BuildConsensus(sequences,
				id,
				CLIKmerSize(),
				CLISaveGraphToFiles(),
				CLIGraphFilesDirectory(),
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
