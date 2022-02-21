package obiformats

import (
	"bufio"
	"compress/gzip"
	"io"
	"log"
	"os"
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
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
		return "genebank"

	default:
		return "unknown"
	}
}

func ReadSequencesBatchFromFile(filename string, options ...WithOption) (obiseq.IBioSequenceBatch, error) {
	var file *os.File
	var reader io.Reader
	var greader io.Reader
	var err error

	file, err = os.Open(filename)

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return obiseq.NilIBioSequenceBatch, err
	}

	reader = file

	// Test if the flux is compressed by gzip
	greader, err = gzip.NewReader(reader)
	if err != nil {
		file.Seek(0, 0)
	} else {
		log.Printf("File %s is gz compressed ", filename)
		reader = greader
	}

	breader := bufio.NewReader(reader)

	tag, _ := breader.Peek(30)

	if len(tag) < 30 {
		newIter := obiseq.MakeIBioSequenceBatch()
		newIter.Close()
		return newIter, nil
	}

	filetype := GuessSeqFileType(string(tag))
	log.Printf("File guessed format : %s (tag: %s)",
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
	default:
		log.Fatalf("File %s has guessed format %s which is not yet implemented",
			filename, filetype)
	}

	return obiseq.NilIBioSequenceBatch, nil
}

func ReadSequencesFromFile(filename string, options ...WithOption) (obiseq.IBioSequence, error) {
	ib, err := ReadSequencesBatchFromFile(filename, options...)
	return ib.SortBatches().IBioSequence(), err

}
