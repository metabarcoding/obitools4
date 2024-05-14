package obiminion

import (
	"fmt"
	"os"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obialign"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obigraph"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiannotate"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconsensus"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiuniq"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

// SampleWeight calculates the weight of a sample based on the statistics of a sequence.
//
// Parameters:
// - seqs: a pointer to BioSequenceSlice representing the sequences (*BioSequenceSlice)
// - sample: the sample for which the weight is calculated (string)
// - sample_key: the key used to access the sample's statistics (string)
// Return type: a function that takes an integer index and returns the weight of the sample at that index (func(int) int)
func SampleWeight(seqs *obiseq.BioSequenceSlice, sample, sample_key string) func(int) float64 {

	f := func(i int) float64 {

		stats := (*seqs)[i].StatsOn(sample_key, "NA")

		if value, ok := stats[sample]; ok {
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
			stats := s.StatsOn(sample_key, "NA")
			for k := range stats {
				if seqset, ok := samples[k]; ok {
					*seqset = append(*seqset, s)
					samples[k] = seqset
				} else {
					samples[k] = &obiseq.BioSequenceSlice{s}
				}
			}
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

	for i, v := range *graph.Vertices {
		var err error
		var clean *obiseq.BioSequence
		degree := graph.Degree(i)
		if degree > 4 {
			pack := obiseq.MakeBioSequenceSlice(degree + 1)
			for k,j := range graph.Neighbors(i) {
				pack[k] = (*graph.Vertices)[j]
			}
			pack[degree] = v
			clean, err = obiconsensus.BuildConsensus(pack,
				fmt.Sprintf("%s_consensus", v.Id()),
				kmer_size,
				CLISaveGraphToFiles(), CLIGraphFilesDirectory())

			if err != nil {
				log.Warning(err)
				clean = (*graph.Vertices)[i]
				clean.SetAttribute("obiminion_consensus", false)
			} else {
				clean.SetAttribute("obiminion_consensus", true)
			}
			pack.Recycle(false)
		} else {
			clean = obiseq.NewBioSequence(v.Id(), v.Sequence(), v.Definition())
			clean.SetAttribute("obiminion_consensus", false)
		}

		clean.SetCount(int(graph.VertexWeight(i)))
		clean.SetAttribute(sample_key, graph.Name)

		denoised[i] = clean
	}

	return denoised
}
func CLIOBIMinion(itertator obiiter.IBioSequence) obiiter.IBioSequence {
	dirname := CLIGraphFilesDirectory()
	newIter := obiiter.MakeIBioSequence()

	db := itertator.Load()

	log.Infof("Sequence dataset of %d sequeences loaded\n", len(db))

	samples := SeqBySamples(db, CLISampleAttribute())
	db.Recycle(false)

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
				obioptions.CLIParallelWorkers())
			if bar != nil {
				bar.Add(1)
			}

			if CLISaveGraphToFiles() {
				graph.WriteGmlFile(fmt.Sprintf("%s/%s.gml",
					CLIGraphFilesDirectory(),
					sample),
					false, 1, 0, 3)
			}

			denoised := MinionDenoise(graph,
				CLISampleAttribute(),
				CLIKmerSize())

			newIter.Push(obiiter.MakeBioSequenceBatch(sample_order, denoised))

			sample_order++
		}

		newIter.Done()
	}()

	go func() {
		newIter.WaitAndClose()
	}()

	obiuniq.AddStatsOn(CLISampleAttribute())
	obiuniq.SetUniqueInMemory(false)
	obiuniq.SetNoSingleton(CLINoSingleton())
	return obiuniq.CLIUnique(newIter).Pipe(obiiter.WorkerPipe(obiannotate.AddSeqLengthWorker(), false))
}
