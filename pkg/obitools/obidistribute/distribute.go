package obidistribute

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
)

func CLIDistributeSequence(sequences obiiter.IBioSequence) {

	opts := make([]obiformats.WithOption, 0, 10)

	switch obiconvert.CLIOutputFastHeaderFormat() {
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

	nworkers := obidefault.ParallelWorkers() / 4
	if nworkers < 2 {
		nworkers = 2
	}

	opts = append(opts, obiformats.OptionsParallelWorkers(nworkers),
		obiformats.OptionsBatchSize(obidefault.BatchSize()),
		obiformats.OptionsAppendFile(CLIAppendSequences()),
		obiformats.OptionsCompressed(obidefault.CompressOutput()))

	var formater obiformats.SequenceBatchWriterToFile

	switch obiconvert.CLIOutputFormat() {
	case "fastq":
		formater = obiformats.WriteFastqToFile
	case "fasta":
		formater = obiformats.WriteFastaToFile
	default:
		formater = obiformats.WriteSequencesToFile
	}

	dispatcher := sequences.Distribute(CLISequenceClassifier(),
		obidefault.BatchSize())

	obiformats.WriterDispatcher(CLIFileNamePattern(),
		dispatcher, formater, opts...,
	)

}
