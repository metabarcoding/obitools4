package obidistribute

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiformats"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obitools/obiconvert"
)

func DistributeSequence(sequences obiiter.IBioSequenceBatch) {

	opts := make([]obiformats.WithOption, 0, 10)

	switch obiconvert.OutputFastHeaderFormat() {
	case "json":
		log.Println("On output use JSON headers")
		opts = append(opts, obiformats.OptionsFastSeqHeaderFormat(obiformats.FormatFastSeqJsonHeader))
	case "obi":
		log.Println("On output use OBI headers")
		opts = append(opts, obiformats.OptionsFastSeqHeaderFormat(obiformats.FormatFastSeqOBIHeader))
	default:
		log.Println("On output use JSON headers")
		opts = append(opts, obiformats.OptionsFastSeqHeaderFormat(obiformats.FormatFastSeqJsonHeader))
	}

	nworkers := obioptions.CLIParallelWorkers() / 4
	if nworkers < 2 {
		nworkers = 2
	}

	opts = append(opts, obiformats.OptionsParallelWorkers(nworkers))
	opts = append(opts, obiformats.OptionsBufferSize(obioptions.CLIBufferSize()))
	opts = append(opts, obiformats.OptionsBatchSize(obioptions.CLIBatchSize()))

	opts = append(opts, obiformats.OptionsQualityShift(obiconvert.OutputQualityShift()))

	var formater obiformats.SequenceBatchWriterToFile

	switch obiconvert.OutputFormat() {
	case "fastq":
		formater = obiformats.WriteFastqBatchToFile
	case "fasta":
		formater = obiformats.WriteFastaBatchToFile
	default:
		formater = obiformats.WriteSequencesBatchToFile
	}

	dispatcher := sequences.Distribute(CLISequenceClassifier(),
		obioptions.CLIBatchSize())

	obiformats.WriterDispatcher(CLIFileNamePattern(),
		dispatcher, formater, opts...,
	)

}