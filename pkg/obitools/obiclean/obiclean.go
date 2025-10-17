package obiclean

import (
	"encoding/csv"
	"fmt"
	"maps"
	"os"
	"sort"
	"strconv"
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
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
	IsHead    bool
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
				IsHead:    false,
			})
		}
	}

	return samples
}

func annotateOBIClean(source string, dataset obiseq.BioSequenceSlice) obiiter.IBioSequence {
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

func NotAlwaysChimera(tag string) obiseq.SequencePredicate {
	descriptor := obiseq.MakeStatsOnDescription(tag)
	predicat := func(sequence *obiseq.BioSequence) bool {

		chimera, ok := sequence.GetStringMap("chimera")
		if !ok || len(chimera) == 0 {
			return true
		}
		samples := maps.Keys(sequence.StatsOn(descriptor, "NA"))

		for s := range samples {
			if _, ok := chimera[s]; !ok {
				return true
			}
		}

		return false
	}

	return predicat
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
	var err error
	annotation := sequence.Annotations()
	iobistatus, ok := annotation["obiclean_status"]
	var obistatus map[string]string

	if ok {
		switch iobistatus := iobistatus.(type) {
		case map[string]string:
			obistatus = iobistatus
		case map[string]interface{}:
			obistatus, err = obiutils.InterfaceToStringMap(obistatus)
			if err != nil {
				log.Panicf("obiclean_status attribute of sequence %s must be castable to a map[string]string", sequence.Id())
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

// DumpOutputTable writes the merged CSV output across all samples.
func DumpOutputTable(filename string, samples map[string]*[]*seqPCR) error {
	// Create file and CSV writer
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	// Collect and sort sample names
	sampleNames := make([]string, 0, len(samples))
	for name := range samples {
		sampleNames = append(sampleNames, name)
	}
	sort.Strings(sampleNames)

	// Determine sequences that are head in at least one sample
	headSeqs := make(map[string]*obiseq.BioSequence)
	for sname, seqs := range samples {
		for _, pcr := range *seqs {
			st := Status(pcr.Sequence)
			if v, ok := st[sname]; ok && v == "h" {
				headSeqs[pcr.Sequence.Id()] = pcr.Sequence
			}
		}
	}

	// Build a map of chimera status per sample for each sequence
	// chimericInSample[seqID][sampleName] = true if sequence is chimeric in that sample
	chimericInSample := make(map[string]map[string]bool)
	for id, seq := range headSeqs {
		if cm, ok := seq.GetStringMap("chimera"); ok && len(cm) > 0 {
			chimericInSample[id] = make(map[string]bool)
			for sname, v := range cm {
				if v != "0" && v != "" {
					chimericInSample[id][sname] = true
				}
			}
		}
	}

	// Build header: first column is sequences (nucleotide string), then one column per sample
	header := make([]string, 0, len(sampleNames)+1)
	header = append(header, "sequences")
	header = append(header, sampleNames...)
	if err := w.Write(header); err != nil {
		return err
	}

	// Prepare sorted list of sequences by sequence string
	type seqPair struct {
		id  string
		seq string
	}

	seqList := make([]seqPair, 0, len(headSeqs))
	for id, s := range headSeqs {
		seqList = append(seqList, seqPair{id: id, seq: strings.ToUpper(string(s.Sequence()))})
	}

	// Sort lexicographically by sequence string
	sort.Slice(seqList, func(i, j int) bool {
		return seqList[i].seq < seqList[j].seq
	})

	// Get minimum sample count threshold
	minSampleCount := MinSampleCount()

	// Write one row per sequence
	for _, pair := range seqList {
		seq := headSeqs[pair.id]
		seqStr := pair.seq

		statusMap := Status(seq)
		weightMap := Weight(seq)

		// Build row and count valid samples simultaneously
		row := make([]string, 0, len(sampleNames)+1)
		row = append(row, seqStr)
		validSampleCount := 0

		for _, sname := range sampleNames {
			val := 0
			// Check if sequence is head in this sample
			if st, ok := statusMap[sname]; ok && st == "h" {
				// Check if sequence is chimeric in this specific sample
				isChimericHere := false
				if sampleChimeras, ok := chimericInSample[pair.id]; ok {
					isChimericHere = sampleChimeras[sname]
				}
				// Only include weight if not chimeric in this sample
				if !isChimericHere {
					if wv, ok := weightMap[sname]; ok {
						val = wv
						validSampleCount++
					}
				}
			}
			row = append(row, strconv.Itoa(val))
		}

		// Skip sequences that don't meet minimum sample count
		if validSampleCount < minSampleCount {
			continue
		}

		if err := w.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func CLIOBIClean(itertator obiiter.IBioSequence) obiiter.IBioSequence {

	source, db := itertator.Load()

	log.Infof("Sequence dataset of %d sequeences loaded\n", len(db))

	samples := buildSamples(db, SampleAttribute(), "NA")

	log.Infof("Dataset composed of %d samples\n", len(samples))

	BuildSeqGraph(samples,
		DistStepMax(),
		obidefault.ParallelWorkers())

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

	if DetectChimera() {
		AnnotateChimera(samples)
	}

	if SaveGraphToFiles() {
		SaveGMLGraphs(GraphFilesDirectory(), samples, MinCountToEvalMutationRate())
	}

	if IsSaveRatioTable() {
		all_ratio := EstimateRatio(samples, MinCountToEvalMutationRate())
		EmpiricalDistCsv(RatioTableFilename(), all_ratio, obidefault.CompressOutput())
	}

	if IsOutputTable() {
		err := DumpOutputTable(OutputTable(), samples)
		if err != nil {
			log.Errorf("cannot write output table: %v", err)
		} else {
			log.Infof("Output table written to %s", OutputTable())
		}
	}

	iter := annotateOBIClean(source, db)

	if OnlyHead() {
		iter = iter.FilterOn(IsHead,
			obidefault.BatchSize()).FilterOn(NotAlwaysChimera(SampleAttribute()),
			obidefault.BatchSize())
	}

	if MinSampleCount() > 1 {
		sc := obiseq.OccurInAtleast(SampleAttribute(), MinSampleCount())
		iter = iter.FilterOn(sc, obidefault.BatchSize())
	}

	return iter
}
