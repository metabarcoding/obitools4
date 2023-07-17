package obiconvert

import (
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiformats"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
)

func BuildPairedFileNames(filename string) (string, string) {

	dir, name := filepath.Split(filename)
	parts := strings.SplitN(name, ".", 2)

	forward := parts[0] + "_R1"
	reverse := parts[0] + "_R2"

	if parts[1] != "" {
		suffix := "." + parts[1]
		forward += suffix
		reverse += suffix
	}

	if dir != "" {
		forward = filepath.Join(dir, forward)
		reverse = filepath.Join(dir, reverse)
	}

	return forward, reverse
}
func CLIWriteBioSequences(iterator obiiter.IBioSequence,
	terminalAction bool, filenames ...string) (obiiter.IBioSequence, error) {

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
	opts = append(opts, obiformats.OptionsBatchSize(obioptions.CLIBatchSize()))

	opts = append(opts, obiformats.OptionsQualityShift(CLIOutputQualityShift()))

	opts = append(opts, obiformats.OptionsCompressed(CLICompressed()))

	var err error

	// No file names are specified or it is "-" : the output is done on stdout

	if CLIOutPutFileName() != "-" || (len(filenames) > 0 && filenames[0] != "-") {
		var fn string

		if len(filenames) == 0 {
			fn = CLIOutPutFileName()
		} else {
			fn = filenames[0]
		}

		if iterator.IsPaired() {
			var reverse string
			fn, reverse = BuildPairedFileNames(fn)
			opts = append(opts, obiformats.WritePairedReadsTo(reverse))
		} else {
			opts = append(opts, obiformats.OptionsSkipEmptySequence(CLISkipEmpty()))
		}

		switch CLIOutputFormat() {
		case "fastq":
			newIter, err = obiformats.WriteFastqToFile(iterator, fn, opts...)
		case "fasta":
			newIter, err = obiformats.WriteFastaToFile(iterator, fn, opts...)
		default:
			newIter, err = obiformats.WriteSequencesToFile(iterator, fn, opts...)
		}
	} else {
		opts = append(opts, obiformats.OptionsSkipEmptySequence(CLISkipEmpty()))
		switch CLIOutputFormat() {
		case "fastq":
			newIter, err = obiformats.WriteFastqToStdout(iterator, opts...)
		case "fasta":
			newIter, err = obiformats.WriteFastaToStdout(iterator, opts...)
		default:
			newIter, err = obiformats.WriteSequencesToStdout(iterator, opts...)
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
