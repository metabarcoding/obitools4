package obiconvert

import (
	"log"

	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obiformats"
	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obiseq"
)

func WriteBioSequences(iterator obiseq.IBioSequence, filenames ...string) error {

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
