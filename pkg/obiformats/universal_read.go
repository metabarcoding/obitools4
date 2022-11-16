package obiformats

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
)

func GuessSeqFileType(firstline string) string {
	switch {
	case strings.HasPrefix(firstline, "#@ecopcr-v2"):
		return "ecopcr"

	case strings.HasPrefix(firstline, "#"):
		return "ecopcr"

	case strings.HasPrefix(firstline, ">"):
		return "fasta"

	case strings.HasPrefix(firstline, "@"):
		return "fastq"

	case strings.HasPrefix(firstline, "ID   "):
		return "embl"

	case strings.HasPrefix(firstline, "LOCUS       "):
		return "genbank"

	// Special case for genbank release files
	// I hope it is enougth stringeant
	case strings.HasSuffix(firstline, " Genetic Se"):
		return "genbank"

	default:
		return "unknown"
	}
}

func ReadSequencesBatchFromFile(filename string,
	options ...WithOption) (obiiter.IBioSequenceBatch, error) {
	var file *os.File
	var reader io.Reader
	var greader io.Reader
	var err error

	file, err = os.Open(filename)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return obiiter.NilIBioSequenceBatch, err
	}

	reader = file

	// Test if the flux is compressed by gzip
	greader, err = gzip.NewReader(reader)
	if err != nil {
		file.Seek(0, 0)
	} else {
		log.Debugf("File %s is gz compressed ", filename)
		reader = greader
	}

	breader := bufio.NewReader(reader)

	tag, _ := breader.Peek(30)

	if len(tag) < 30 {
		newIter := obiiter.MakeIBioSequenceBatch()
		newIter.Close()
		return newIter, nil
	}

	filetype := GuessSeqFileType(string(tag))
	log.Debug("File guessed format : %s (tag: %s)",
		filetype, (strings.Split(string(tag), "\n"))[0])
	reader = breader

	switch filetype {
	case "fastq", "fasta":
		file.Close()
		is, _ := ReadFastSeqBatchFromFile(filename, options...)
		return is, nil
	case "ecopcr":
		return ReadEcoPCRBatch(reader, options...), nil
	case "embl":
		return ReadEMBLBatch(reader, options...), nil
	case "genbank":
		return ReadGenbankBatch(reader, options...), nil
	default:
		log.Fatalf("File %s has guessed format %s which is not yet implemented",
			filename, filetype)
	}

	return obiiter.NilIBioSequenceBatch, nil
}
