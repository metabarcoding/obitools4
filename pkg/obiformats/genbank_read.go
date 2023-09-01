package obiformats

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	gzip "github.com/klauspost/pgzip"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
)

type gbstate int

const (
	inHeader     gbstate = 0
	inEntry      gbstate = 1
	inDefinition gbstate = 2
	inFeature    gbstate = 3
	inSequence   gbstate = 4
)

var _seqlenght_rx = regexp.MustCompile(" +([0-9]+) bp")

func _ParseGenbankFile(source string,
	input <-chan _FileChunk, out obiiter.IBioSequence,
	chunck_order func() int) {
	var err error
	state := inHeader

	for chunks := range input {
		// log.Debugln("Chunk size", (chunks.raw.(*bytes.Buffer)).Len())
		scanner := bufio.NewScanner(chunks.raw)
		sequences := make(obiseq.BioSequenceSlice, 0, 100)
		sumlength := 0
		id := ""
		lseq := -1
		scientificName := ""
		defBytes := new(bytes.Buffer)
		featBytes := new(bytes.Buffer)
		seqBytes := new(bytes.Buffer)
		taxid := 1
		nl := 0
		sl := 0
		for scanner.Scan() {
			nl++
			line := scanner.Text()
			switch {
			case state == inDefinition && !strings.HasPrefix(line, "            "):
				state = inEntry
				fallthrough
			case strings.HasPrefix(line, "LOCUS       "):
				state = inEntry
				id = strings.SplitN(line[12:], " ", 2)[0]
				match_length := _seqlenght_rx.FindStringSubmatch(line)
				if len(match_length) > 0 {
					lseq, err = strconv.Atoi(match_length[1])
					if err != nil {
						lseq = -1
					}
				}
				if lseq > 0 {
					seqBytes = bytes.NewBuffer(obiseq.GetSlice(lseq + 20))
				} else {
					seqBytes = new(bytes.Buffer)
				}
			case strings.HasPrefix(line, "SOURCE      "):
				scientificName = strings.TrimSpace(line[12:])
			case strings.HasPrefix(line, "DEFINITION  "):
				defBytes.WriteString(strings.TrimSpace(line[12:]))
				state = inDefinition
			case strings.HasPrefix(line, "FEATURES    "):
				featBytes.WriteString(line)
				state = inFeature
			case strings.HasPrefix(line, "ORIGIN"):
				state = inSequence
			case line == "//":
				// log.Debugln("Total lines := ", nl)
				sequence := obiseq.NewBioSequence(id,
					seqBytes.Bytes(),
					defBytes.String())
				sequence.SetSource(source)
				state = inHeader

				sequence.SetFeatures(featBytes.Bytes())

				annot := sequence.Annotations()
				annot["scientific_name"] = scientificName
				annot["taxid"] = taxid
				// log.Println(FormatFasta(sequence, FormatFastSeqJsonHeader))
				// log.Debugf("Read sequences %s: %dbp (%d)", sequence.Id(),
				//	sequence.Len(), seqBytes.Len())

				sequences = append(sequences, sequence)
				sumlength += sequence.Len()

				if len(sequences) == 100 || sumlength > 1e7 {
					out.Push(obiiter.MakeBioSequenceBatch(chunck_order(), sequences))
					sequences = make(obiseq.BioSequenceSlice, 0, 100)
					sumlength = 0
				}
				defBytes = bytes.NewBuffer(obiseq.GetSlice(200))
				featBytes = new(bytes.Buffer)
				nl = 0
				sl = 0
			default:
				switch state {
				case inDefinition:
					defBytes.WriteByte(' ')
					defBytes.WriteString(strings.TrimSpace(line[5:]))
				case inFeature:
					featBytes.WriteByte('\n')
					featBytes.WriteString(line)
					if strings.HasPrefix(line, `                     /db_xref="taxon:`) {
						taxid, _ = strconv.Atoi(strings.SplitN(line[37:], `"`, 2)[0])
					}
				case inSequence:
					sl++
					parts := strings.SplitN(line[10:], " ", 7)
					lparts := len(parts)
					for i := 0; i < lparts; i++ {
						seqBytes.WriteString(parts[i])
					}
				}
			}

		}
		if len(sequences) > 0 {
			out.Push(obiiter.MakeBioSequenceBatch(chunck_order(), sequences))
		}
	}

	out.Done()

}

func ReadGenbank(reader io.Reader, options ...WithOption) obiiter.IBioSequence {
	opt := MakeOptions(options)
	entry_channel := make(chan _FileChunk)

	newIter := obiiter.MakeIBioSequence()

	nworkers := opt.ParallelWorkers()
	chunck_order := obiutils.AtomicCounter()
	newIter.Add(nworkers)

	go func() {
		newIter.WaitAndClose()
	}()

	// for j := 0; j < opt.ParallelWorkers(); j++ {
	for j := 0; j < nworkers; j++ {
		go _ParseGenbankFile(opt.Source(), entry_channel, newIter, chunck_order)
	}

	go _ReadFlatFileChunk(reader, entry_channel)

	if opt.pointer.full_file_batch {
		newIter = newIter.CompleteFileIterator()
	}

	return newIter
}

func ReadGenbankFromFile(filename string, options ...WithOption) (obiiter.IBioSequence, error) {
	var reader io.Reader
	var greader io.Reader
	var err error

	options = append(options, OptionsSource(obiutils.RemoveAllExt((path.Base(filename)))))

	reader, err = os.Open(filename)
	if err != nil {
		log.Printf("open file error: %+v", err)
		return obiiter.NilIBioSequence, err
	}

	// Test if the flux is compressed by gzip
	//greader, err = gzip.NewReader(reader)
	greader, err = gzip.NewReaderN(reader, 1<<24, 2)
	if err == nil {
		reader = greader
	}

	return ReadGenbank(reader, options...), nil
}
