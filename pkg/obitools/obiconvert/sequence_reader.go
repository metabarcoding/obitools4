package obiconvert

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goombaio/orderedset"
	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
)

func ExpandListOfFiles(check_ext bool, filenames ...string) ([]string, error) {
	var err error
	list_of_files := orderedset.NewOrderedSet()
	for _, fn := range filenames {

		if strings.HasPrefix(fn, "http://") ||
			strings.HasPrefix(fn, "https://") ||
			strings.HasPrefix(fn, "ftp://") {
			list_of_files.Add(fn)
			continue
		}

		err = filepath.Walk(fn,
			func(path string, info os.FileInfo, err error) error {
				var e error
				if info == nil {
					return fmt.Errorf("cannot open path")
				}
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
						subdir, e := ExpandListOfFiles(true, path)
						if e != nil {
							return e
						}
						for _, f := range subdir {
							list_of_files.Add(f)
						}
					} else {
						check_ext = true
					}
				} else {
					if !check_ext ||
						strings.HasSuffix(path, "fasta") ||
						strings.HasSuffix(path, "fasta.gz") ||
						strings.HasSuffix(path, "fastq") ||
						strings.HasSuffix(path, "fastq.gz") ||
						strings.HasSuffix(path, "fq") ||
						strings.HasSuffix(path, "fq.gz") ||
						strings.HasSuffix(path, "seq") ||
						strings.HasSuffix(path, "seq.gz") ||
						strings.HasSuffix(path, "gb") ||
						strings.HasSuffix(path, "gb.gz") ||
						strings.HasSuffix(path, "dat") ||
						strings.HasSuffix(path, "dat.gz") ||
						strings.HasSuffix(path, "ecopcr") ||
						strings.HasSuffix(path, "ecopcr.gz") {
						log.Debugf("Appending %s file\n", path)
						list_of_files.Add(path)
					}
				}
				return nil
			})

		if err != nil {
			return nil, err
		}
	}

	res := make([]string, 0, list_of_files.Size())
	for _, v := range list_of_files.Values() {
		res = append(res, v.(string))
	}

	log.Infof("Found %d files to process", len(res))
	return res, nil
}

func CLIReadBioSequences(filenames ...string) (obiiter.IBioSequence, error) {
	var iterator obiiter.IBioSequence
	var reader func(string, ...obiformats.WithOption) (obiiter.IBioSequence, error)

	opts := make([]obiformats.WithOption, 0, 10)

	switch CLIInputFastHeaderFormat() {
	case "json":
		opts = append(opts, obiformats.OptionsFastSeqHeaderParser(obiformats.ParseFastSeqJsonHeader))
	case "obi":
		opts = append(opts, obiformats.OptionsFastSeqHeaderParser(obiformats.ParseFastSeqOBIHeader))
	default:
		opts = append(opts, obiformats.OptionsFastSeqHeaderParser(obiformats.ParseGuessedFastSeqHeader))
	}

	opts = append(opts, obiformats.OptionsReadQualities(obidefault.ReadQualities()))

	nworkers := obidefault.ReadParallelWorkers()
	if nworkers < 2 {
		nworkers = 2
	}

	opts = append(opts, obiformats.OptionsParallelWorkers(nworkers))
	opts = append(opts, obiformats.OptionsBatchSize(obidefault.BatchSize()))

	opts = append(opts, obiformats.OptionsFullFileBatch(FullFileBatch()))
	opts = append(opts, obiformats.OptionsUtoT(CLIUtoT()))

	if len(filenames) == 0 {
		log.Printf("Reading sequences from stdin in %s\n", CLIInputFormat())
		opts = append(opts, obiformats.OptionsSource("stdin"))

		var err error

		switch CLIInputFormat() {
		case "ecopcr":
			iterator, err = obiformats.ReadEcoPCR(os.Stdin, opts...)
		case "embl":
			iterator, err = obiformats.ReadEMBL(os.Stdin, opts...)
		case "genbank":
			iterator, err = obiformats.ReadGenbank(os.Stdin, opts...)
		case "fasta":
			iterator, err = obiformats.ReadFasta(os.Stdin, opts...)
		case "fastq":
			iterator, err = obiformats.ReadFastq(os.Stdin, opts...)
		case "csv":
			iterator, err = obiformats.ReadCSV(os.Stdin, opts...)
		default:
			iterator, err = obiformats.ReadSequencesFromStdin(opts...)
		}

		if err != nil {
			return obiiter.NilIBioSequence, err
		}

	} else {

		list_of_files, err := ExpandListOfFiles(false, filenames...)
		if err != nil {
			return obiiter.NilIBioSequence, err
		}
		switch CLIInputFormat() {
		case "fastq", "fq":
			reader = obiformats.ReadFastqFromFile
		case "fasta":
			reader = obiformats.ReadFastaFromFile
		case "csv":
			reader = obiformats.ReadCSVFromFile
		case "ecopcr":
			reader = obiformats.ReadEcoPCRFromFile
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
				nreader = obidefault.ParallelFilesRead()
			}

			iterator = obiformats.ReadSequencesBatchFromFiles(
				list_of_files,
				reader,
				nreader,
				opts...,
			)

		} else {
			if len(list_of_files) > 0 {

				iterator, err = reader(list_of_files[0], opts...)

				if err != nil {
					return obiiter.NilIBioSequence, err
				}

				if CLIPairedFileName() != "" {
					ip, err := reader(CLIPairedFileName(), opts...)

					if err != nil {
						return obiiter.NilIBioSequence, err
					}

					iterator = iterator.PairTo(ip)
				}
			} else {
				iterator = obiiter.NilIBioSequence
			}
		}

	}

	iterator = iterator.Speed("Reading sequences")

	return iterator, nil
}

func OpenSequenceDataErrorMessage(args []string, err error) {
	if err != nil {
		switch len(args) {
		case 0:
			log.Errorf("Cannot open stdin (%v)", err)
		case 1:
			log.Errorf("Cannot open file %s: %v", args[0], err)
		default:
			log.Errorf("Cannot open one of the data files: %v", err)
		}
		os.Exit(1)
	}
}
