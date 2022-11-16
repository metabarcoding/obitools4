package obiconvert

import (
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiformats"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
)

func _ExpandListOfFiles(check_ext bool, filenames ...string) ([]string, error) {
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
						subdir, e := _ExpandListOfFiles(true, path)
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

func ReadBioSequences(filenames ...string) (obiiter.IBioSequenceBatch, error) {
	var iterator obiiter.IBioSequenceBatch
	var reader func(string, ...obiformats.WithOption) (obiiter.IBioSequenceBatch, error)

	opts := make([]obiformats.WithOption, 0, 10)

	switch CLIInputFastHeaderFormat() {
	case "json":
		opts = append(opts, obiformats.OptionsFastSeqHeaderParser(obiformats.ParseFastSeqJsonHeader))
	case "obi":
		opts = append(opts, obiformats.OptionsFastSeqHeaderParser(obiformats.ParseFastSeqOBIHeader))
	default:
		opts = append(opts, obiformats.OptionsFastSeqHeaderParser(obiformats.ParseGuessedFastSeqHeader))
	}

	nworkers := obioptions.CLIParallelWorkers() // / 4
	if nworkers < 2 {
		nworkers = 2
	}

	opts = append(opts, obiformats.OptionsParallelWorkers(nworkers))
	opts = append(opts, obiformats.OptionsBufferSize(obioptions.CLIBufferSize()))
	opts = append(opts, obiformats.OptionsBatchSize(obioptions.CLIBatchSize()))

	opts = append(opts, obiformats.OptionsQualityShift(CLIInputQualityShift()))

	if len(filenames) == 0 {
		log.Printf("Reading sequences from stdin in %s\n", CLIInputFormat())
		switch CLIInputFormat() {
		case "ecopcr":
			iterator = obiformats.ReadEcoPCR(os.Stdin, opts...)
		case "embl":
			iterator = obiformats.ReadEMBL(os.Stdin, opts...)
		case "genbank":
			iterator = obiformats.ReadGenbank(os.Stdin, opts...)
		default:
			iterator = obiformats.ReadFastSeqFromStdin(opts...)
		}
	} else {

		list_of_files, err := _ExpandListOfFiles(false, filenames...)
		if err != nil {
			return obiiter.NilIBioSequenceBatch, err
		}

		switch CLIInputFormat() {
		case "ecopcr":
			reader = obiformats.ReadEcoPCRBatchFromFile
		case "embl":
			reader = obiformats.ReadEMBLFromFile
		case "genbank":
			reader = obiformats.ReadGenbankFromFile
		default:
			reader = obiformats.ReadSequencesFromFile
		}

		if len(list_of_files) > 1 {
			nreader := 1

			if CLINoInputOrder() {
				nreader = obioptions.CLIParallelWorkers()
			}

			iterator = obiformats.ReadSequencesBatchFromFiles(
				filenames,
				reader,
				nreader,
				opts...,
			)
		} else {
			iterator, err = reader(list_of_files[0], opts...)

			if err != nil {
				return obiiter.NilIBioSequenceBatch, err
			}
		}

		// list_of_files = list_of_files[1:]
		// others := make([]obiiter.IBioSequenceBatch, 0, len(list_of_files))

		// for _, fn := range list_of_files {
		// 	r, err := reader(fn, opts...)
		// 	if err != nil {
		// 		return obiiter.NilIBioSequenceBatch, err
		// 	}
		// 	others = append(others, r)
		// }

		// if len(others) > 0 {
		// 	if CLINoInputOrder() {
		// 		iterator = iterator.Pool(others...)
		// 	} else {
		// 		iterator = iterator.Concat(others...)
		// 	}
		// }

	}

	// if SequencesToSkip() > 0 {
	// 	iterator = iterator.Skip(SequencesToSkip())
	// }

	// if AnalyzeOnly() > 0 {
	// 	iterator = iterator.Head(AnalyzeOnly())
	// }

	return iterator, nil
}
