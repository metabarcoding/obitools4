package obiconvert

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiformats"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func __expand_list_of_files__(check_ext bool, filenames ...string) ([]string, error) {
	var err error
	list_of_files := make([]string, 0, 100)
	for _, fn := range filenames {

		err = filepath.Walk(fn,
			func(path string, info os.FileInfo, err error) error {
				var e error
				for info.Mode()&os.ModeSymlink == os.ModeSymlink {
					path, e = filepath.EvalSymlinks(path)
					if e != nil {
						return e
					}

					info, e = os.Stat(path)
					if e != nil {
						return e
					}
				}

				if info.IsDir() {
					if path != fn {
						subdir, e := __expand_list_of_files__(true, path)
						if e != nil {
							return e
						}
						list_of_files = append(list_of_files, subdir...)
					} else {
						check_ext = true
					}
				} else {
					if !check_ext ||
						strings.HasSuffix(path, "fasta") ||
						strings.HasSuffix(path, "fasta.gz") ||
						strings.HasSuffix(path, "fastq") ||
						strings.HasSuffix(path, "fastq.gz") ||
						strings.HasSuffix(path, "dat") ||
						strings.HasSuffix(path, "dat.gz") ||
						strings.HasSuffix(path, "ecopcr") ||
						strings.HasSuffix(path, "ecopcr.gz") {
						log.Printf("Appending %s file\n", path)
						list_of_files = append(list_of_files, path)
					}
				}
				return nil
			})

		if err != nil {
			return nil, err
		}
	}

	return list_of_files, nil
}

func ReadBioSequencesBatch(filenames ...string) (obiseq.IBioSequenceBatch, error) {
	var iterator obiseq.IBioSequenceBatch
	var reader func(string, ...obiformats.WithOption) (obiseq.IBioSequenceBatch, error)

	opts := make([]obiformats.WithOption, 0, 10)

	switch InputFastHeaderFormat() {
	case "json":
		opts = append(opts, obiformats.OptionsFastSeqHeaderParser(obiformats.ParseFastSeqJsonHeader))
	case "obi":
		opts = append(opts, obiformats.OptionsFastSeqHeaderParser(obiformats.ParseFastSeqOBIHeader))
	default:
		opts = append(opts, obiformats.OptionsFastSeqHeaderParser(obiformats.ParseGuessedFastSeqHeader))
	}

	opts = append(opts, obiformats.OptionsQualityShift(InputQualityShift()))

	if len(filenames) == 0 {

		switch InputFormat() {
		case "ecopcr":
			iterator = obiformats.ReadEcoPCRBatch(os.Stdin, opts...)
		case "embl":
			iterator = obiformats.ReadEMBLBatch(os.Stdin, opts...)
		default:
			iterator = obiformats.ReadFastSeqBatchFromStdin(opts...)
		}
	} else {

		list_of_files, err := __expand_list_of_files__(false, filenames...)
		if err != nil {
			return obiseq.NilIBioSequenceBatch, err
		}

		switch InputFormat() {
		case "ecopcr":
			reader = obiformats.ReadEcoPCRBatchFromFile
		case "embl":
			reader = obiformats.ReadEMBLBatchFromFile
		default:
			reader = obiformats.ReadSequencesBatchFromFile
		}

		iterator, err = reader(list_of_files[0], opts...)

		if err != nil {
			return obiseq.NilIBioSequenceBatch, err
		}

		list_of_files = list_of_files[1:]
		others := make([]obiseq.IBioSequenceBatch, 0, len(list_of_files))

		for _, fn := range list_of_files {
			r, err := reader(fn, opts...)
			if err != nil {
				return obiseq.NilIBioSequenceBatch, err
			}
			others = append(others, r)
		}

		if len(others) > 0 {
			iterator = iterator.Concat(others...)
		}

	}

	// if SequencesToSkip() > 0 {
	// 	iterator = iterator.Skip(SequencesToSkip())
	// }

	// if AnalyzeOnly() > 0 {
	// 	iterator = iterator.Head(AnalyzeOnly())
	// }

	return iterator, nil
}

func ReadBioSequences(filenames ...string) (obiseq.IBioSequence, error) {
	ib, err := ReadBioSequencesBatch(filenames...)
	return ib.SortBatches().IBioSequence(), err

}
