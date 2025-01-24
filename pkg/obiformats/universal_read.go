package obiformats

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"errors"
	"io"
	"path"
	"regexp"

	"github.com/gabriel-vasile/mimetype"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

type SequenceReader func(reader io.Reader, options ...WithOption) (obiiter.IBioSequence, error)

// OBIMimeTypeGuesser is a function that takes an io.Reader as input and guesses the MIME type of the data.
// It uses several detectors to identify specific file formats, such as FASTA, FASTQ, ecoPCR2, GenBank, and EMBL.
// The function reads data from the input stream and analyzes it using the mimetype library.
// It then returns the detected MIME type, a modified reader with the read data, and any error encountered during the process.
//
// The following file types are recognized:
// - "text/ecopcr": if the first line starts with "#@ecopcr-v2".
// - "text/fasta": if the first line starts with ">".
// - "text/fastq": if the first line starts with "@".
// - "text/embl": if the first line starts with "ID   ".
// - "text/genbank": if the first line starts with "LOCUS       ".
// - "text/genbank" (special case): if the first line "Genetic Sequence Data Bank" (for genbank release files).
// - "text/csv"
//
// Parameters:
// - stream: An io.Reader representing the input stream to read data from.
//
// Returns:
// - *mimetype.MIME: The detected MIME type of the data.
// - io.Reader: A modified reader with the read data.
// - error: Any error encountered during the process.
func OBIMimeTypeGuesser(stream io.Reader) (*mimetype.MIME, io.Reader, error) {
	csv := func(in []byte, limit uint32) bool {
		in = dropLastLine(in, limit)

		br := bytes.NewReader(in)
		r := csv.NewReader(br)
		r.Comma = ','
		r.ReuseRecord = true
		r.LazyQuotes = true
		r.Comment = '#'

		lines := 0
		for {
			_, err := r.Read()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return false
			}
			lines++
		}

		return r.FieldsPerRecord > 1 && lines > 1
	}

	fastaDetector := func(raw []byte, limit uint32) bool {
		ok, err := regexp.Match("^>[^ ]", raw)
		return ok && err == nil
	}

	fastqDetector := func(raw []byte, limit uint32) bool {
		ok, err := regexp.Match("^@[^ ].*\n[^ ]+\n\\+", raw)
		return ok && err == nil
	}

	ecoPCR2Detector := func(raw []byte, limit uint32) bool {
		ok := bytes.HasPrefix(raw, []byte("#@ecopcr-v2"))
		return ok
	}

	genbankDetector := func(raw []byte, limit uint32) bool {
		ok2 := bytes.HasPrefix(raw, []byte("LOCUS       "))
		ok1, err := regexp.Match("^[^ ]* +Genetic Sequence Data Bank *\n", raw)
		return ok2 || (ok1 && err == nil)
	}

	emblDetector := func(raw []byte, limit uint32) bool {
		ok := bytes.HasPrefix(raw, []byte("ID   "))
		return ok
	}

	mimetype.Lookup("text/plain").Extend(fastaDetector, "text/fasta", ".fasta")
	mimetype.Lookup("text/plain").Extend(fastqDetector, "text/fastq", ".fastq")
	mimetype.Lookup("text/plain").Extend(ecoPCR2Detector, "text/ecopcr2", ".ecopcr")
	mimetype.Lookup("text/plain").Extend(genbankDetector, "text/genbank", ".seq")
	mimetype.Lookup("text/plain").Extend(emblDetector, "text/embl", ".dat")
	mimetype.Lookup("text/plain").Extend(csv, "text/csv", ".csv")

	mimetype.Lookup("application/octet-stream").Extend(fastaDetector, "text/fasta", ".fasta")
	mimetype.Lookup("application/octet-stream").Extend(fastqDetector, "text/fastq", ".fastq")
	mimetype.Lookup("application/octet-stream").Extend(ecoPCR2Detector, "text/ecopcr2", ".ecopcr")
	mimetype.Lookup("application/octet-stream").Extend(genbankDetector, "text/genbank", ".seq")
	mimetype.Lookup("application/octet-stream").Extend(emblDetector, "text/embl", ".dat")
	mimetype.Lookup("application/octet-stream").Extend(csv, "text/csv", ".csv")

	// Create a buffer to store the read data
	buf := make([]byte, 1024*1024)
	n, err := io.ReadFull(stream, buf)

	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, nil, err
	}

	// Detect the MIME type using the mimetype library
	mimeType := mimetype.Detect(buf)

	if mimeType == nil {
		return nil, nil, err
	}

	// Create a new reader based on the read data
	newReader := io.Reader(bytes.NewReader(buf[:n]))

	if err == nil {
		newReader = io.MultiReader(newReader, stream)
	}

	return mimeType, newReader, nil
}

// func ReadSequences(reader io.Reader,
// 	options ...WithOption) (obiiter.IBioSequence, error) {

// 	mime, reader, err := OBIMimeTypeGuesser(reader)

// 	if err != nil {
// 		return obiiter.NilIBioSequence, err
// 	}

// 	reader = bufio.NewReader(reader)

// 	switch mime.String() {
// 	case "text/fasta", "text/fastq":
// 		file.Close()
// 		is, err := ReadFastSeqFromFile(filename, options...)
// 		return is, err
// 	case "text/ecopcr2":
// 		return ReadEcoPCR(reader, options...), nil
// 	case "text/embl":
// 		return ReadEMBL(reader, options...), nil
// 	case "text/genbank":
// 		return ReadGenbank(reader, options...), nil
// 	default:
// 		log.Fatalf("File %s has guessed format %s which is not yet implemented",
// 			filename, mime.String())
// 	}

// 	return obiiter.NilIBioSequence, nil
// }

// ReadSequencesFromFile reads sequences from a file and returns an iterator of bio sequences and an error.
//
// Parameters:
// - filename: The name of the file to read the sequences from.
// - options: Optional parameters to customize the reading process.
//
// Returns:
// - obiiter.IBioSequence: An iterator of bio sequences.
// - error: An error if any occurred during the reading process.
func ReadSequencesFromFile(filename string,
	options ...WithOption) (obiiter.IBioSequence, error) {
	var file *obiutils.Reader
	var reader io.Reader
	var err error

	options = append(options, OptionsSource(obiutils.RemoveAllExt((path.Base(filename)))))

	file, err = obiutils.Ropen(filename)

	if err == obiutils.ErrNoContent {
		log.Infof("file %s is empty", filename)
		return ReadEmptyFile(options...)
	}

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return obiiter.NilIBioSequence, err
	}

	mime, reader, err := OBIMimeTypeGuesser(file)

	if err != nil {
		return obiiter.NilIBioSequence, err
	}
	log.Infof("%s mime type: %s", filename, mime.String())
	reader = bufio.NewReader(reader)

	switch mime.String() {
	case "text/fastq":
		return ReadFastq(reader, options...)
	case "text/fasta":
		return ReadFasta(reader, options...)
	case "text/ecopcr2":
		return ReadEcoPCR(reader, options...)
	case "text/embl":
		return ReadEMBL(reader, options...)
	case "text/genbank":
		return ReadGenbank(reader, options...)
	case "text/csv":
		return ReadCSV(reader, options...)
	default:
		log.Fatalf("File %s has guessed format %s which is not yet implemented",
			filename, mime.String())
	}

	return obiiter.NilIBioSequence, nil
}

// func ReadSequencesFromStdin(options ...WithOption) obiiter.IBioSequence {

// 	options = append(options, OptionsSource("stdin"))

// }
