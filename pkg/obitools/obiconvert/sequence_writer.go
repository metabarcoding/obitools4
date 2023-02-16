package obiconvert

import (
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiformats"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
)

func CLIWriteBioSequences(iterator obiiter.IBioSequence,
	terminalAction bool, filenames ...string) (obiiter.IBioSequence, error) {

	if CLIProgressBar() {
		iterator = iterator.Speed()
	}
	var newIter obiiter.IBioSequence

	opts := make([]obiformats.WithOption, 0, 10)

	switch CLIOutputFastHeaderFormat() {
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

	opts = append(opts, obiformats.OptionsQualityShift(CLIOutputQualityShift()))

	var err error

	if len(filenames) == 0 {
		switch CLIOutputFormat() {
		case "fastq":
			newIter, err = obiformats.WriteFastqToStdout(iterator, opts...)
		case "fasta":
			newIter, err = obiformats.WriteFastaToStdout(iterator, opts...)
		default:
			newIter, err = obiformats.WriteSequencesToStdout(iterator, opts...)
		}
	} else {
		switch CLIOutputFormat() {
		case "fastq":
			newIter, err = obiformats.WriteFastqToFile(iterator, filenames[0], opts...)
		case "fasta":
			newIter, err = obiformats.WriteFastaToFile(iterator, filenames[0], opts...)
		default:
			newIter, err = obiformats.WriteSequencesToFile(iterator, filenames[0], opts...)
		}
	}

	if err != nil {
		log.Fatalf("Write file error: %v", err)
		return obiiter.NilIBioSequence, err
	}

	if terminalAction {
		newIter.Recycle()
		return obiiter.NilIBioSequence, nil
	}

	return newIter, nil
}
