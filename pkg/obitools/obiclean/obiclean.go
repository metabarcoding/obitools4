package obiclean

import (
	"fmt"
	"os"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

type seqPCR struct {
	Count    int                 // number of reads associated to a sequence in a PCR
	Sequence *obiseq.BioSequence // pointer to the corresponding sequence
	SonCount int
	Fathers  []int
	Dist     []int
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
		stats := s.StatsOn(tag, NAValue)

		for k, v := range stats {
			pcr, ok := samples[k]

			if !ok {
				p := make([]*seqPCR, 0, 10)
				pcr = &p
				samples[k] = pcr
			}

			*pcr = append(*pcr, &seqPCR{
				Count:    v,
				Sequence: s,
				SonCount: 0,
			})
		}
	}

	return samples
}

func annotateOBIClean(dataset obiseq.BioSequenceSlice,
	sample map[string]*([]*seqPCR),
	tag, NAValue string) obiiter.IBioSequenceBatch {
	batchsize := 1000
	var annot = func(data obiseq.BioSequenceSlice) obiseq.BioSequenceSlice {

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
		return data
	}

	iter := obiiter.IBatchOver(dataset, batchsize)
	riter := iter.MakeISliceWorker(annot)

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
			log.Panic("obiclean_head attribute of sequence %s must be a boolean not : %v", sequence.Id(), iishead)
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
		value, err = goutils.InterfaceToInt(value)
		if err != nil {
			log.Panic("obiclean_headcount attribute of sequence %s must be an integer value not : %v", sequence.Id(), ivalue)
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
		value, err = goutils.InterfaceToInt(value)
		if err != nil {
			log.Panic("obiclean_internalcount attribute of sequence %s must be an integer value not : %v", sequence.Id(), ivalue)
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
		value, err = goutils.InterfaceToInt(value)
		if err != nil {
			log.Panic("obiclean_samplecount attribute of sequence %s must be an integer value not : %v", sequence.Id(), ivalue)
		}
	}

	return value
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

func IOBIClean(itertator obiiter.IBioSequenceBatch) obiiter.IBioSequenceBatch {

	db := itertator.Load()

	log.Infof("Sequence dataset of %d sequeences loaded\n", len(db))

	samples := buildSamples(db, SampleAttribute(), "NA")

	log.Infof("Dataset composed of %d samples\n", len(samples))

	all_ratio := BuildSeqGraph(samples,
		DistStepMax(),
		MinCountToEvalMutationRate(),
		obioptions.CLIParallelWorkers())

	if RatioMax() < 1.0 {
		pbopt := make([]progressbar.Option, 0, 5)
		pbopt = append(pbopt,
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionSetWidth(15),
			progressbar.OptionShowIts(),
			progressbar.OptionSetPredictTime(true),
			progressbar.OptionSetDescription("[Filter graph on abundance ratio]"),
		)

		bar := progressbar.NewOptions(len(samples), pbopt...)

		for _, seqs := range samples {
			FilterGraphOnRatio(seqs, RatioMax())
			bar.Add(1)
		}
	}

	if IsSaveRatioTable() {
		EmpiricalDistCsv(RatioTableFilename(), all_ratio)
	}

	if SaveGraphToFiles() {
		SaveGMLGraphs(GraphFilesDirectory(), samples, MinCountToEvalMutationRate())
	}

	pbopt := make([]progressbar.Option, 0, 5)
	pbopt = append(pbopt,
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionSetDescription("[Annotate sequence status]"),
	)

	bar := progressbar.NewOptions(len(samples), pbopt...)

	for name, seqs := range samples {
		for _, pcr := range *seqs {
			obistatus := Status(pcr.Sequence)
			obistatus[name] = ObicleanStatus(pcr)
		}
		bar.Add(1)
	}

	iter := annotateOBIClean(db, samples, SampleAttribute(), "NA")

	if OnlyHead() {
		iter = iter.FilterOn(IsHead, 1000)
	}

	return iter
}