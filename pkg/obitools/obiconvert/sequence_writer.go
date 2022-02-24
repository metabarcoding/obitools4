package obiconvert

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiformats"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
)

func WriteBioSequences(iterator obiiter.IBioSequence, filenames ...string) error {

	opts := make([]obiformats.WithOption, 0, 10)

	switch OutputFastHeaderFormat() {
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

	opts = append(opts, obiformats.OptionsQualityShift(OutputQualityShift()))

	var err error

	if len(filenames) == 0 {
		switch OutputFormat() {
		case "fastq":
			err = obiformats.WriteFastqToStdout(iterator, opts...)
		case "fasta":
			err = obiformats.WriteFastaToStdout(iterator, opts...)
		default:
			err = obiformats.WriteSequencesToStdout(iterator, opts...)
		}
	} else {
		switch OutputFormat() {
		case "fastq":
			err = obiformats.WriteFastqToFile(iterator, filenames[0], opts...)
		case "fasta":
			err = obiformats.WriteFastaToFile(iterator, filenames[0], opts...)
		default:
			err = obiformats.WriteSequencesToFile(iterator, filenames[0], opts...)
		}
	}

	if err != nil {
		log.Fatalf("Write file error: %v", err)
		return err
	}

	return nil
}

func WriteBioSequencesBatch(iterator obiiter.IBioSequenceBatch,
	terminalAction bool, filenames ...string) (obiiter.IBioSequenceBatch, error) {

	var newIter obiiter.IBioSequenceBatch

	opts := make([]obiformats.WithOption, 0, 10)

	switch OutputFastHeaderFormat() {
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

	opts = append(opts, obiformats.OptionsQualityShift(OutputQualityShift()))

	var err error

	if len(filenames) == 0 {
		switch OutputFormat() {
		case "fastq":
			newIter, err = obiformats.WriteFastqBatchToStdout(iterator, opts...)
		case "fasta":
			newIter, err = obiformats.WriteFastaBatchToStdout(iterator, opts...)
		default:
			newIter, err = obiformats.WriteSequencesBatchToStdout(iterator, opts...)
		}
	} else {
		switch OutputFormat() {
		case "fastq":
			newIter, err = obiformats.WriteFastqBatchToFile(iterator, filenames[0], opts...)
		case "fasta":
			newIter, err = obiformats.WriteFastaBatchToFile(iterator, filenames[0], opts...)
		default:
			newIter, err = obiformats.WriteSequencesBatchToFile(iterator, filenames[0], opts...)
		}
	}

	if err != nil {
		log.Fatalf("Write file error: %v", err)
		return obiiter.NilIBioSequenceBatch, err
	}

	if terminalAction {
		newIter.Recycle()
		return obiiter.NilIBioSequenceBatch, nil
	}

	return newIter, nil
}
