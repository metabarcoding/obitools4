package obiconsensus

import (
	"fmt"
	"os"
	"path"
	"slices"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obigraph"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obisuffix"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiannotate"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiuniq"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

func BuildConsensus(seqs obiseq.BioSequenceSlice,
	consensus_id string,
	kmer_size int,
	filter_out float64,
	save_graph bool, dirname string) (*obiseq.BioSequence, error) {

	if seqs.Len() == 0 {
		return nil, fmt.Errorf("no sequence provided")
	}

	if seqs.Len() == 1 {
		seq := seqs[0].Copy()
		seq.SetAttribute("obiconsensus_consensus", false)
		return seq, nil
	}

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
			fasta.Write(obiformats.FormatFastaBatch(obiiter.MakeBioSequenceBatch(
				fmt.Sprintf("%s_consensus", consensus_id),
				0,
				seqs,
			),
				obiformats.FormatFastSeqJsonHeader, false).Bytes())
			fasta.Close()
		}

	}

	log.Debugf("Number of reads : %d\n", len(seqs))

	if kmer_size < 0 {
		longest := make([]int, len(seqs))

		for i, seq := range seqs {
			s := obiseq.BioSequenceSlice{seq}
			sa := obisuffix.BuildSuffixArray(&s)
			longest[i] = slices.Max(sa.CommonSuffix())
		}

		kmer_size = slices.Max(longest) + 1
		log.Debugf("estimated kmer size : %d", kmer_size)
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
		log.Debugf("Cycle detected, increasing kmer size to %d\n", kmer_size)
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

	log.Debugf("Graph size : %d\n", graph.Len())
	total_kmer := graph.Len()

	seq, err := graph.LongestConsensus(consensus_id, filter_out)

	sumCount := 0

	if seq != nil {
		for _, s := range seqs {
			sumCount += s.Count()
		}
		seq.SetAttribute("obiconsensus_consensus", true)
		seq.SetAttribute("obiconsensus_weight", sumCount)
		seq.SetAttribute("obiconsensus_seq_length", seq.Len())
		seq.SetAttribute("obiconsensus_kmer_size", kmer_size)
		seq.SetAttribute("obiconsensus_kmer_max_occur", graph.MaxWeight())
		seq.SetAttribute("obiconsensus_filtered_graph_size", graph.Len())
		seq.SetAttribute("obiconsensus_full_graph_size", total_kmer)
	}
	return seq, err
}

// SampleWeight calculates the weight of a sample based on the statistics of a sequence.
//
// Parameters:
// - seqs: a pointer to BioSequenceSlice representing the sequences (*BioSequenceSlice)
// - sample: the sample for which the weight is calculated (string)
// - sample_key: the key used to access the sample's statistics (string)
// Return type: a function that takes an integer index and returns the weight of the sample at that index (func(int) int)
func SampleWeight(seqs *obiseq.BioSequenceSlice, sample, sample_key string) func(int) float64 {

	f := func(i int) float64 {

		stats := (*seqs)[i].StatsOn(obiseq.MakeStatsOnDescription(sample_key), "NA")

		if stats == nil {
			log.Panicf("Sample %s not found in sequence %d", sample, i)
		}

		if value, ok := stats.Get(sample); ok {
			return float64(value)
		}

		return 0
	}

	return f
}

// SeqBySamples sorts the sequences by samples.
//
// Parameters:
// - seqs: a pointer to BioSequenceSlice representing the sequences (*BioSequenceSlice)
// - sample_key: a string representing the sample key (string)
//
// Return type:
// - map[string]BioSequenceSlice: a map indexed by sample names, each containing a slice of BioSequence objects (map[string]BioSequenceSlice)
func SeqBySamples(seqs obiseq.BioSequenceSlice, sample_key string) map[string]*obiseq.BioSequenceSlice {

	samples := make(map[string]*obiseq.BioSequenceSlice)

	for _, s := range seqs {
		if s.HasStatsOn(sample_key) {
			stats := s.StatsOn(obiseq.MakeStatsOnDescription(sample_key), "NA")
			stats.RLock()
			for k := range stats.Map() {
				if seqset, ok := samples[k]; ok {
					*seqset = append(*seqset, s)
					samples[k] = seqset
				} else {
					samples[k] = &obiseq.BioSequenceSlice{s}
				}
			}
			stats.RUnlock()
		} else {
			if k, ok := s.GetStringAttribute(sample_key); ok {
				if seqset, ok := samples[k]; ok {
					*seqset = append(*seqset, s)
					samples[k] = seqset
				} else {
					samples[k] = &obiseq.BioSequenceSlice{s}
				}
			}
		}
	}

	return samples

}

type Mutation struct {
	Position int
	SeqA     byte
	SeqB     byte
	Ratio    float64
}

func BuildDiffSeqGraph(name, name_key string,
	seqs *obiseq.BioSequenceSlice,
	distmax, nworkers int) *obigraph.Graph[*obiseq.BioSequence, Mutation] {
	graph := obigraph.NewGraphBuffer[*obiseq.BioSequence, Mutation](name, (*[]*obiseq.BioSequence)(seqs))
	iseq := make(chan int)
	defer graph.Close()

	ls := len(*seqs)

	sw := SampleWeight(seqs, name, name_key)
	graph.Graph.VertexWeight = sw

	waiting := sync.WaitGroup{}
	waiting.Add(nworkers)

	bar := (*progressbar.ProgressBar)(nil)
	if obiconvert.CLIProgressBar() {

		pbopt := make([]progressbar.Option, 0, 5)
		pbopt = append(pbopt,
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionSetWidth(15),
			progressbar.OptionShowIts(),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionSetDescription(fmt.Sprintf("[Build graph] on %s", name)),
		)

		bar = progressbar.NewOptions(len(*seqs), pbopt...)
	}

	computeEdges := func() {
		defer waiting.Done()
		for i := range iseq {
			s1 := (*seqs)[i]
			for j := i + 1; j < ls; j++ {
				s2 := (*seqs)[j]
				ratio := sw(i) / sw(j)
				ok, pos, a1, a2 := obialign.D1Or0(s1, s2)
				if ok >= 0 {
					graph.AddEdge(i, j, &Mutation{pos, a1, a2, ratio})
				} else if distmax > 1 {
					lcs, lali := obialign.FastLCSScore(s1, s2, distmax, nil)
					dist := lali - lcs
					if lcs > 0 && dist <= distmax {
						// log.Infof("Seq %s and %s: LCSScore: %d, dist: %d\n", s1.Id(), s2.Id(), lcs, dist)
						graph.AddEdge(i, j, &Mutation{pos, a1, a2, ratio})
					}
				}
			}

			if bar != nil {
				bar.Add(1)
			}
		}
	}

	for i := 0; i < nworkers; i++ {
		go computeEdges()
	}

	for i := 0; i < ls; i++ {
		iseq <- i
	}
	close(iseq)

	waiting.Wait()
	return graph.Graph
}

func MinionDenoise(graph *obigraph.Graph[*obiseq.BioSequence, Mutation],
	sample_key string, kmer_size int) obiseq.BioSequenceSlice {
	denoised := obiseq.MakeBioSequenceSlice(len(*graph.Vertices))

	bar := (*progressbar.ProgressBar)(nil)
	if obiconvert.CLIProgressBar() {

		pbopt := make([]progressbar.Option, 0, 5)
		pbopt = append(pbopt,
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionSetWidth(15),
			progressbar.OptionShowIts(),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionSetDescription(fmt.Sprintf("[Build consensus] on %s", graph.Name)),
		)

		bar = progressbar.NewOptions(len(*graph.Vertices), pbopt...)
	}

	wg := &sync.WaitGroup{}
	seqidxchan := make(chan int)
	build := func() {
		var err error
		var clean *obiseq.BioSequence

		for i := range seqidxchan {
			v := (*graph.Vertices)[i]

			degree := graph.Degree(i)
			if degree > 4 {
				pack := obiseq.MakeBioSequenceSlice(degree + 1)
				for k, j := range graph.Neighbors(i) {
					pack[k] = (*graph.Vertices)[j]
				}
				pack[degree] = v
				clean, err = BuildConsensus(pack,
					fmt.Sprintf("%s_consensus", v.Id()),
					kmer_size, CLILowCoverage(),
					CLISaveGraphToFiles(), CLIGraphFilesDirectory())

				if err != nil {
					log.Warning(err)
					clean = (*graph.Vertices)[i].Copy()
					clean.SetAttribute("obiconsensus_consensus", false)

				}

			} else {
				clean = obiseq.NewBioSequence(v.Id(), v.Sequence(), v.Definition())
				clean.SetAttribute("obiconsensus_consensus", false)
			}

			clean.SetCount(int(graph.VertexWeight(i)))
			clean.SetAttribute(sample_key, graph.Name)

			if !clean.HasAttribute("obiconsensus_weight") {
				clean.SetAttribute("obiconsensus_weight", int(1))
			}

			annotations := v.Annotations()

			staton := obiseq.StatsOnSlotName(sample_key)
			for k, v := range annotations {
				if !clean.HasAttribute(k) && k != staton {
					clean.SetAttribute(k, v)
				}
			}

			denoised[i] = clean

			if bar != nil {
				bar.Add(1)
			}
		}

		wg.Done()
	}

	nworkers := obidefault.ParallelWorkers()
	wg.Add(nworkers)

	for j := 0; j < nworkers; j++ {
		go build()
	}

	for i := range *graph.Vertices {
		seqidxchan <- i
	}

	close(seqidxchan)

	wg.Wait()

	return denoised
}

func MinionClusterDenoise(graph *obigraph.Graph[*obiseq.BioSequence, Mutation],
	sample_key string, kmer_size int) obiseq.BioSequenceSlice {
	denoised := obiseq.MakeBioSequenceSlice()
	seqs := (*obiseq.BioSequenceSlice)(graph.Vertices)
	weight := SampleWeight(seqs, graph.Name, sample_key)
	seqWeights := make([]float64, len(*seqs))

	// Compute weights for each vertex as the sum of the weights of its neighbors

	log.Info("")
	log.Infof("Sample %s: Computing weights", graph.Name)
	for i := range *seqs {
		w := weight(i)
		for _, j := range graph.Neighbors(i) {
			w += weight(j)
		}

		seqWeights[i] = w
	}

	log.Infof("Sample %s: Done computing weights", graph.Name)

	log.Infof("Sample %s: Clustering", graph.Name)
	// Look for vertex not having a neighbor with a higher weight
	for i := range *seqs {
		v := (*seqs)[i]
		head := true
		neighbors := graph.Neighbors(i)
		for _, j := range neighbors {
			if seqWeights[i] < seqWeights[j] {
				head = false
				continue
			}
		}

		if head {
			pack := obiseq.MakeBioSequenceSlice(len(neighbors) + 1)
			for k, j := range neighbors {
				pack[k] = (*seqs)[j]
			}
			pack[len(neighbors)] = v

			clean, err := BuildConsensus(pack,
				fmt.Sprintf("%s_consensus", v.Id()),
				kmer_size, CLILowCoverage(),
				CLISaveGraphToFiles(), CLIGraphFilesDirectory())

			if err != nil {
				log.Warning(err)
				clean = (*graph.Vertices)[i].Copy()
				clean.SetAttribute("obiconsensus_consensus", false)
			}

			clean.SetAttribute(sample_key, graph.Name)

			annotations := v.Annotations()
			clean.SetCount(int(weight(i)))

			staton := obiseq.StatsOnSlotName(sample_key)

			for k, v := range annotations {
				if !clean.HasAttribute(k) && k != staton {
					clean.SetAttribute(k, v)
				}
			}

			denoised = append(denoised, clean)
		}
	}

	log.Infof("Sample %s: Done clustering", graph.Name)

	return denoised
}

func CLIOBIMinion(itertator obiiter.IBioSequence) obiiter.IBioSequence {
	dirname := CLIGraphFilesDirectory()
	newIter := obiiter.MakeIBioSequence()

	source, db := itertator.Load()

	log.Infof("Sequence dataset of %d sequeences loaded\n", len(db))

	samples := SeqBySamples(db, CLISampleAttribute())

	log.Infof("Dataset composed of %d samples\n", len(samples))

	if CLISaveGraphToFiles() {
		if stat, err := os.Stat(dirname); err != nil || !stat.IsDir() {
			// path does not exist or is not directory
			os.RemoveAll(dirname)
			err := os.Mkdir(dirname, 0755)

			if err != nil {
				log.Panicf("Cannot create directory %s for saving graphs", dirname)
			}
		}
	}

	bar := (*progressbar.ProgressBar)(nil)
	if obiconvert.CLIProgressBar() {

		pbopt := make([]progressbar.Option, 0, 5)
		pbopt = append(pbopt,
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionSetWidth(15),
			progressbar.OptionShowIts(),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionSetDescription("[Filter graph on abundance ratio]"),
		)

		bar = progressbar.NewOptions(len(samples), pbopt...)
	}

	newIter.Add(1)

	go func() {
		sample_order := 0
		for sample, seqs := range samples {
			graph := BuildDiffSeqGraph(sample,
				CLISampleAttribute(),
				seqs,
				CLIDistStepMax(),
				obidefault.ParallelWorkers())
			if bar != nil {
				bar.Add(1)
			}

			if CLISaveGraphToFiles() {
				graph.WriteGmlFile(fmt.Sprintf("%s/%s.gml",
					CLIGraphFilesDirectory(),
					sample),
					false, 1, 0, 3)
			}

			var denoised obiseq.BioSequenceSlice

			if CLICluterDenoise() {
				denoised = MinionClusterDenoise(graph,
					CLISampleAttribute(),
					CLIKmerSize())
			} else {
				denoised = MinionDenoise(graph,
					CLISampleAttribute(),
					CLIKmerSize())
			}

			newIter.Push(obiiter.MakeBioSequenceBatch(source, sample_order, denoised))

			sample_order++
		}

		newIter.Done()
	}()

	go func() {
		newIter.WaitAndClose()
	}()

	res := newIter
	if CLIUnique() {
		obiuniq.AddStatsOn(CLISampleAttribute())
		// obiuniq.AddStatsOn("sample:obiconsensus_weight")
		obiuniq.SetUniqueInMemory(false)
		obiuniq.SetNoSingleton(CLINoSingleton())
		res = obiuniq.CLIUnique(newIter)
	}

	return res.Pipe(obiiter.WorkerPipe(obiannotate.AddSeqLengthWorker(), false))
}
