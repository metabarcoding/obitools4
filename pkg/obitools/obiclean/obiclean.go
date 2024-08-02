package obiclean

import (
	"fmt"
	"os"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

type seqPCR struct {
	Count     int                 // number of reads associated to a sequence in a PCR
	Weight    int                 // Number of reads associated to a sequence after clustering
	Sequence  *obiseq.BioSequence // pointer to the corresponding sequence
	SonCount  int
	AddedSons int
	Edges     []Edge
	Cluster   map[int]bool // used as the set of head sequences associated to that sequence
}

// buildSamples sorts the sequences by samples
//
// The sequences are distributed according to their sample association.
// The function returns a map indexed by sample names.
// Each sample is represented by a vector where each element associates
// a sequence to its count of occurrences in the considered sample.
func buildSamples(dataset obiseq.BioSequenceSlice,
	tag, NAValue string) map[string]*([]*seqPCR) {
	samples := make(map[string]*([]*seqPCR))

	for _, s := range dataset {
		stats := s.StatsOn(obiseq.MakeStatsOnDescription(tag), NAValue)

		for k, v := range stats {
			pcr, ok := samples[k]

			if !ok {
				p := make([]*seqPCR, 0, 10)
				pcr = &p
				samples[k] = pcr
			}

			*pcr = append(*pcr, &seqPCR{
				Count:     v,
				Sequence:  s,
				SonCount:  0,
				AddedSons: 0,
			})
		}
	}

	return samples
}

func annotateOBIClean(source string, dataset obiseq.BioSequenceSlice,
	sample map[string]*([]*seqPCR),
	tag, NAValue string) obiiter.IBioSequence {
	batchsize := 1000
	var annot = func(data obiseq.BioSequenceSlice) (obiseq.BioSequenceSlice, error) {

		for _, s := range data {
			status := Status(s)
			head := 0
			internal := 0
			singleton := 0
			for _, v := range status {
				switch v {
				case "i":
					internal++
				case "h":
					head++
				case "s":
					singleton++
				}
			}
			is_head := (head + singleton) > 0

			annotation := s.Annotations()
			annotation["obiclean_head"] = is_head
			annotation["obiclean_singletoncount"] = singleton
			annotation["obiclean_internalcount"] = internal
			annotation["obiclean_headcount"] = head
			annotation["obiclean_samplecount"] = head + internal + singleton

		}
		return data, nil
	}

	iter := obiiter.IBatchOver(source, dataset, batchsize)
	riter := iter.MakeISliceWorker(annot, false)

	return riter
}

func IsHead(sequence *obiseq.BioSequence) bool {
	annotation := sequence.Annotations()
	iishead, ok := annotation["obiclean_head"]
	ishead := true

	if ok {
		switch iishead := iishead.(type) {
		case bool:
			ishead = iishead
		default:
			log.Panicf("obiclean_head attribute of sequence %s must be a boolean not : %v", sequence.Id(), iishead)
		}
	}

	return ishead
}

func HeadCount(sequence *obiseq.BioSequence) int {
	var err error
	annotation := sequence.Annotations()
	ivalue, ok := annotation["obiclean_headcount"]
	value := 0

	if ok {
		value, err = obiutils.InterfaceToInt(value)
		if err != nil {
			log.Panicf("obiclean_headcount attribute of sequence %s must be an integer value not : %v", sequence.Id(), ivalue)
		}
	}

	return value
}

func InternalCount(sequence *obiseq.BioSequence) int {
	var err error
	annotation := sequence.Annotations()
	ivalue, ok := annotation["obiclean_internalcount"]
	value := 0

	if ok {
		value, err = obiutils.InterfaceToInt(value)
		if err != nil {
			log.Panicf("obiclean_internalcount attribute of sequence %s must be an integer value not : %v", sequence.Id(), ivalue)
		}
	}

	return value
}

func SingletonCount(sequence *obiseq.BioSequence) int {
	var err error
	annotation := sequence.Annotations()
	ivalue, ok := annotation["obiclean_samplecount"]
	value := 0

	if ok {
		value, err = obiutils.InterfaceToInt(value)
		if err != nil {
			log.Panicf("obiclean_samplecount attribute of sequence %s must be an integer value not : %v", sequence.Id(), ivalue)
		}
	}

	return value
}

func GetMutation(sequence *obiseq.BioSequence) map[string]string {
	annotation := sequence.Annotations()
	imutation, ok := annotation["obiclean_mutation"]
	var mutation map[string]string

	if ok {
		switch imutation := imutation.(type) {
		case map[string]string:
			mutation = imutation
		case map[string]interface{}:
			mutation = make(map[string]string)
			for k, v := range imutation {
				mutation[k] = fmt.Sprint(v)
			}
		}
	} else {
		mutation = make(map[string]string)
		annotation["obiclean_mutation"] = mutation
	}

	return mutation
}

func GetCluster(sequence *obiseq.BioSequence) map[string]string {
	annotation := sequence.Annotations()
	icluster, ok := annotation["obiclean_cluster"]
	var cluster map[string]string

	if ok {
		switch icluster := icluster.(type) {
		case map[string]string:
			cluster = icluster
		case map[string]interface{}:
			cluster = make(map[string]string)
			for k, v := range icluster {
				cluster[k] = fmt.Sprint(v)
			}
		}
	} else {
		cluster = make(map[string]string)
		annotation["obiclean_cluster"] = cluster
	}

	return cluster
}

// func Cluster(sample map[string]*([]*seqPCR)) {
// 	for _, graph := range sample {
// 		for _, s := range *graph {
// 			cluster := GetCluster(s.Sequence)
// 			if len(s.Edges) > 0 {
// 				for _, f := range s.Edges {

// 				}
// 			} else {
// 				cluster
// 			}

// 		}
// 	}
// }

func Mutation(sample map[string]*([]*seqPCR)) {
	for _, graph := range sample {
		for _, s := range *graph {
			for _, f := range s.Edges {
				id := (*graph)[f.Father].Sequence.Id()
				GetMutation(s.Sequence)[id] = fmt.Sprintf("(%c)->(%c)@%d",
					f.From, f.To, f.Pos+1)
			}
		}
	}
}

func Status(sequence *obiseq.BioSequence) map[string]string {
	annotation := sequence.Annotations()
	iobistatus, ok := annotation["obiclean_status"]
	var obistatus map[string]string

	if ok {
		switch iobistatus := iobistatus.(type) {
		case map[string]string:
			obistatus = iobistatus
		case map[string]interface{}:
			obistatus = make(map[string]string)
			for k, v := range iobistatus {
				obistatus[k] = fmt.Sprint(v)
			}
		}
	} else {
		obistatus = make(map[string]string)
		annotation["obiclean_status"] = obistatus
	}

	return obistatus
}

func Weight(sequence *obiseq.BioSequence) map[string]int {
	annotation := sequence.Annotations()
	iobistatus, ok := annotation["obiclean_weight"]
	var weight map[string]int
	var err error

	if ok {
		switch iobistatus := iobistatus.(type) {
		case map[string]int:
			weight = iobistatus
		case map[string]interface{}:
			weight = make(map[string]int)
			for k, v := range iobistatus {
				weight[k], err = obiutils.InterfaceToInt(v)
				if err != nil {
					log.Panicf("Weight value %v cannnot be casted to an integer value\n", v)
				}
			}
		}
	} else {
		weight = make(map[string]int)
		annotation["obiclean_weight"] = weight
	}

	return weight
}

func CLIOBIClean(itertator obiiter.IBioSequence) obiiter.IBioSequence {

	source, db := itertator.Load()

	log.Infof("Sequence dataset of %d sequeences loaded\n", len(db))

	samples := buildSamples(db, SampleAttribute(), "NA")

	log.Infof("Dataset composed of %d samples\n", len(samples))

	BuildSeqGraph(samples,
		DistStepMax(),
		obioptions.CLIParallelWorkers())

	if RatioMax() < 1.0 {
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

		for _, seqs := range samples {
			FilterGraphOnRatio(seqs, RatioMax())
			if bar != nil {
				bar.Add(1)
			}
		}
	}

	Mutation(samples)

	bar := (*progressbar.ProgressBar)(nil)

	if obiconvert.CLIProgressBar() {
		pbopt := make([]progressbar.Option, 0, 5)
		pbopt = append(pbopt,
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionSetWidth(15),
			progressbar.OptionShowIts(),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionSetDescription("[Annotate sequence status]"),
		)

		bar = progressbar.NewOptions(len(samples), pbopt...)
	}

	for name, seqs := range samples {
		for _, pcr := range *seqs {
			obistatus := Status(pcr.Sequence)
			obistatus[name] = ObicleanStatus(pcr)

			obiweight := Weight(pcr.Sequence)
			obiweight[name] = pcr.Weight
		}

		if bar != nil {
			bar.Add(1)
		}
	}

	if SaveGraphToFiles() {
		SaveGMLGraphs(GraphFilesDirectory(), samples, MinCountToEvalMutationRate())
	}

	if IsSaveRatioTable() {
		all_ratio := EstimateRatio(samples, MinCountToEvalMutationRate())
		EmpiricalDistCsv(RatioTableFilename(), all_ratio)
	}

	iter := annotateOBIClean(source, db, samples, SampleAttribute(), "NA")

	if OnlyHead() {
		iter = iter.FilterOn(IsHead, 1000)
	}

	return iter
}
