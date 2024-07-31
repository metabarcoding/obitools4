package obiblackboard

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

	"github.com/gabriel-vasile/mimetype"
	"github.com/goombaio/orderedset"
	log "github.com/sirupsen/logrus"
)

func ExpandListOfFiles(check_ext bool, filenames ...string) ([]string, error) {
	res, err := _ExpandListOfFiles(check_ext, filenames...)

	if err != nil {
		log.Infof("Found %d files to process", len(res))
	}

	return res, err
}

func _ExpandListOfFiles(check_ext bool, filenames ...string) ([]string, error) {
	var err error
	list_of_files := orderedset.NewOrderedSet()
	for _, fn := range filenames {
		// Special case for stdin
		if fn == "-" {
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
						strings.HasSuffix(path, "csv") ||
						strings.HasSuffix(path, "csv.gz") ||
						strings.HasSuffix(path, "fasta") ||
						strings.HasSuffix(path, "fasta.gz") ||
						strings.HasSuffix(path, "fastq") ||
						strings.HasSuffix(path, "fastq.gz") ||
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

	return res, nil
}

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

	mimetype.Lookup("application/octet-stream").Extend(fastaDetector, "text/fasta", ".fasta")
	mimetype.Lookup("application/octet-stream").Extend(fastqDetector, "text/fastq", ".fastq")
	mimetype.Lookup("application/octet-stream").Extend(ecoPCR2Detector, "text/ecopcr2", ".ecopcr")
	mimetype.Lookup("application/octet-stream").Extend(genbankDetector, "text/genbank", ".seq")
	mimetype.Lookup("application/octet-stream").Extend(emblDetector, "text/embl", ".dat")

	// Create a buffer to store the read data
	buf := make([]byte, 1024*128)
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

func TextChunkParser(parser obiformats.SeqFileChunkParser, target string) DoTask {

	return func(bb *Blackboard, task *Task) *Task {
		chunk := task.Body.(obiformats.SeqFileChunk)
		sequences, err := parser(chunk.Source, chunk.Raw)

		if err != nil {
			return nil
		}

		nt := task.GetNext(target, false, false)
		nt.Body = obiiter.MakeBioSequenceBatch(
			chunk.Source,
			chunk.Order,
			sequences)

		return nt
	}
}

func SeqAnnotParser(parser obiseq.SeqAnnotator, target string) DoTask {
	worker := obiseq.SeqToSliceWorker(obiseq.AnnotatorToSeqWorker(parser), false)

	return func(bb *Blackboard, task *Task) *Task {
		batch := task.Body.(obiiter.BioSequenceBatch)
		sequences, err := worker(batch.Slice())

		if err != nil {
			log.Errorf("SeqAnnotParser on %s[%d]: %v", batch.Source(), batch.Order(), err)
			return nil
		}

		nt := task.GetNext(target, false, false)
		nt.Body = obiiter.MakeBioSequenceBatch(
			batch.Source(),
			batch.Order(),
			sequences,
		)
		return nt
	}

}

// OpenStream opens a file specified by the given filename and returns a reader for the file,
// the detected MIME type of the file, and any error encountered during the process.
//
// Parameters:
//   - filename: A string representing the path to the file to be opened. If the filename is "-",
//     the function opens the standard input stream.
//
// Returns:
// - io.Reader: A reader for the file.
// - *mimetype.MIME: The detected MIME type of the file.
// - error: Any error encountered during the process.
func OpenStream(filename string) (io.Reader, *mimetype.MIME, error) {
	var stream io.Reader
	var err error
	if filename == "-" {
		stream, err = obiformats.Buf(os.Stdin)
	} else {
		stream, err = obiformats.Ropen(filename)
	}

	if err != nil {
		return nil, nil, err
	}

	// Detect the MIME type using the mimetype library
	mimeType, newReader, err := OBIMimeTypeGuesser(stream)
	if err != nil {
		return nil, nil, err
	}

	log.Infof("%s mime type: %s", filename, mimeType.String())

	return bufio.NewReader(newReader), mimeType, nil
}

type OpenedStreamBody struct {
	Stream     io.Reader
	Filename   string
	Source     string
	Mime       *mimetype.MIME
	ToBeClosed bool
}

func FilenameToStream(target string) DoTask {

	return func(bb *Blackboard, task *Task) *Task {
		filename := task.Body.(Iteration[string]).Value
		stream, mimetype, err := OpenStream(filename)

		if err != nil {
			log.Errorf("Error opening %s: %v", filename, err)
			return nil
		}

		tobeclosed := filename != "-"

		switch mimetype.String() {
		case "text/fasta", "text/fastq", "text/ecopcr2", "text/genbank", "text/embl", "text/csv":
			nt := task.GetNext(target+":"+mimetype.String(), false, false)
			nt.Body = OpenedStreamBody{
				Stream:     stream,
				Mime:       mimetype,
				Filename:   filename,
				Source:     obiutils.RemoveAllExt((path.Base(filename))),
				ToBeClosed: tobeclosed,
			}

			return nt

		default:
			log.Errorf("File %s (mime type %s) is an unsupported format", filename, mimetype.String())
			return nil
		}
	}
}

type TextChunkIteratorBody struct {
	Chunks     obiformats.ChannelSeqFileChunk
	Stream     io.Reader
	Source     string
	ToBeClosed bool
}

func StreamToTextChunkReader(lastEntry obiformats.LastSeqRecord, target string) DoTask {
	return func(bb *Blackboard, task *Task) *Task {

		body := task.Body.(OpenedStreamBody)
		iterator := obiformats.ReadSeqFileChunk(
			body.Source,
			body.Stream,
			make([]byte, 64*1024*1024),
			lastEntry,
		)

		nt := task.GetNext(target, false, false)
		nt.Body = TextChunkIteratorBody{
			Chunks:     iterator,
			Stream:     body.Stream,
			Source:     body.Source,
			ToBeClosed: body.ToBeClosed,
		}

		return nt
	}
}

func TextChuckIterator(endTask *Task, target string) DoTask {
	return func(bb *Blackboard, task *Task) *Task {
		body := task.Body.(TextChunkIteratorBody)

		chunk, ok := <-body.Chunks

		if !ok {
			return endTask
		}

		var nt *Task

		if bb.Len() > bb.TargetSize {
			nt = task.GetNext(target, false, true)
		} else {
			nt = task.GetNext(target, false, false)
			bb.PushTask(task)
		}

		nt.Body = chunk
		return nt
	}
}

type SequenceIteratorBody struct {
	Iterator   obiiter.IBioSequence
	Stream     io.Reader
	Source     string
	ToBeClosed bool
}

func StreamToSequenceReader(
	reader obiformats.SequenceReader,
	options []obiformats.WithOption,
	target string) DoTask {
	return func(bb *Blackboard, task *Task) *Task {
		body := task.Body.(OpenedStreamBody)
		iterator, err := reader(body.Stream, options...)

		if err != nil {
			log.Errorf("Error opening %s: %v", body.Filename, err)
			return nil
		}

		nt := task.GetNext(target, false, false)
		nt.Body = SequenceIteratorBody{
			Iterator:   iterator,
			Stream:     body.Stream,
			Source:     body.Source,
			ToBeClosed: body.ToBeClosed,
		}

		return nt
	}
}

func SequenceIterator(endTask *Task, target string) DoTask {
	return func(bb *Blackboard, task *Task) *Task {
		body := task.Body.(SequenceIteratorBody)

		if body.Iterator.Next() {
			batch := body.Iterator.Get()

			var nt *Task
			if bb.Len() > bb.TargetSize {
				nt = task.GetNext(target, false, true)
			} else {
				nt = task.GetNext(target, false, false)
				bb.PushTask(task)
			}

			nt.Body = batch

			return nt
		} else {
			return endTask
		}
	}
}

func (bb *Blackboard) ReadSequences(filepath []string, options ...obiformats.WithOption) {

	var err error

	opts := obiformats.MakeOptions(options)

	if len(filepath) == 0 {
		filepath = []string{"-"}
	}

	filepath, err = ExpandListOfFiles(false, filepath...)

	if err != nil {
		log.Fatalf("Cannot expand list of files : %v", err)
	}

	bb.RegisterRunner(
		"initial",
		DoIterateSlice(filepath, "filename"),
	)

	bb.RegisterRunner(
		"filename",
		FilenameToStream("stream"),
	)

	bb.RegisterRunner("stream:text/fasta",
		StreamToTextChunkReader(
			obiformats.EndOfLastFastaEntry,
			"fasta_text_reader",
		))

	bb.RegisterRunner("fasta_text_reader",
		TextChuckIterator(NewInitialTask(), "fasta_text_chunk"),
	)

	bb.RegisterRunner(
		"fasta_text_chunk",
		TextChunkParser(
			obiformats.FastaChunkParser(),
			"unannotated_sequences",
		),
	)

	bb.RegisterRunner("stream:text/fastq",
		StreamToTextChunkReader(obiformats.EndOfLastFastqEntry,
			"fastq_text_reader"))

	bb.RegisterRunner("fastq_text_reader",
		TextChuckIterator(NewInitialTask(), "fastq_text_chunk"),
	)

	bb.RegisterRunner(
		"fastq_text_chunk",
		TextChunkParser(
			obiformats.FastqChunkParser(obioptions.InputQualityShift()),
			"unannotated_sequences",
		),
	)

	bb.RegisterRunner("stream:text/embl",
		StreamToTextChunkReader(obiformats.EndOfLastFlatFileEntry,
			"embl_text_reader"))

	bb.RegisterRunner("embl_text_reader",
		TextChuckIterator(NewInitialTask(), "embl_text_chunk"),
	)

	bb.RegisterRunner(
		"embl_text_chunk",
		TextChunkParser(
			obiformats.EmblChunkParser(opts.WithFeatureTable()),
			"sequences",
		),
	)

	bb.RegisterRunner("stream:text/genbank",
		StreamToTextChunkReader(obiformats.EndOfLastFlatFileEntry,
			"genbank_text_reader"))

	bb.RegisterRunner("genbank_text_reader",
		TextChuckIterator(NewInitialTask(), "genbank_text_chunk"),
	)

	bb.RegisterRunner(
		"genbank_text_chunk",
		TextChunkParser(
			obiformats.GenbankChunkParser(opts.WithFeatureTable()),
			"sequences",
		),
	)

	bb.RegisterRunner(
		"unannotated_sequences",
		SeqAnnotParser(
			opts.ParseFastSeqHeader(),
			"sequences",
		),
	)

	bb.RegisterRunner("stream:text/csv",
		StreamToSequenceReader(obiformats.ReadCSV, options, "sequence_reader"))

	bb.RegisterRunner("stream:text/ecopcr2",
		StreamToSequenceReader(obiformats.ReadEcoPCR, options, "sequence_reader"))

	bb.RegisterRunner("sequence_reader",
		SequenceIterator(NewInitialTask(), "sequences"),
	)

}
