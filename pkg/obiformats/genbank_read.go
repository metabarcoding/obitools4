package obiformats

import (
	"bufio"
	"bytes"
	"io"
	"path"
	"regexp"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

type gbstate int

const (
	inHeader     gbstate = 0
	inEntry      gbstate = 1
	inDefinition gbstate = 2
	inFeature    gbstate = 3
	inSequence   gbstate = 4
	inContig     gbstate = 5
)

var _seqlenght_rx = regexp.MustCompile(" +([0-9]+) bp")

func GenbankChunkParser(withFeatureTable bool) func(string, io.Reader) (obiseq.BioSequenceSlice, error) {
	return func(source string, input io.Reader) (obiseq.BioSequenceSlice, error) {
		state := inHeader
		scanner := bufio.NewReader(input)
		sequences := obiseq.MakeBioSequenceSlice(100)[:0]
		id := ""
		lseq := -1
		scientificName := ""
		defBytes := new(bytes.Buffer)
		featBytes := new(bytes.Buffer)
		seqBytes := new(bytes.Buffer)
		taxid := 1
		nl := 0
		sl := 0
		var line string
		for bline, is_prefix, err := scanner.ReadLine(); err != io.EOF; bline, is_prefix, err = scanner.ReadLine() {
			nl++
			line = string(bline)
			if is_prefix || len(line) > 100 {
				log.Fatalf("From %s:Line too long: %s", source, line)
			}
			processed := false
			for !processed {
				switch {

				case strings.HasPrefix(line, "LOCUS       "):
					if state != inHeader {
						log.Fatalf("Unexpected state %d while reading LOCUS: %s", state, line)
					}
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
					state = inEntry
					processed = true

				case strings.HasPrefix(line, "DEFINITION  "):
					if state != inEntry {
						log.Fatalf("Unexpected state %d while reading DEFINITION: %s", state, line)
					}
					defBytes.WriteString(strings.TrimSpace(line[12:]))
					state = inDefinition
					processed = true

				case state == inDefinition:
					if strings.HasPrefix(line, "            ") {
						defBytes.WriteByte(' ')
						defBytes.WriteString(strings.TrimSpace(line[12:]))
						processed = true
					} else {
						state = inEntry
					}

				case strings.HasPrefix(line, "SOURCE      "):
					if state != inEntry {
						log.Fatalf("Unexpected state %d while reading SOURCE: %s", state, line)
					}
					scientificName = strings.TrimSpace(line[12:])
					processed = true

				case strings.HasPrefix(line, "FEATURES    "):
					if state != inEntry {
						log.Fatalf("Unexpected state %d while reading FEATURES: %s", state, line)
					}
					featBytes.WriteString(line)
					state = inFeature
					processed = true

				case strings.HasPrefix(line, "ORIGIN"):
					if state != inFeature {
						log.Fatalf("Unexpected state %d while reading ORIGIN: %s", state, line)
					}
					state = inSequence
					processed = true

				case strings.HasPrefix(line, "CONTIG"):
					if state != inFeature && state != inContig {
						log.Fatalf("Unexpected state %d while reading ORIGIN: %s", state, line)
					}
					state = inContig
					processed = true

				case line == "//":

					if state != inSequence && state != inContig {
						log.Fatalf("Unexpected state %d while reading end of record %s", state, id)
					}
					// log.Debugln("Total lines := ", nl)
					if id == "" {
						log.Warn("Empty id when parsing genbank file")
					}

					// log.Debugf("End of sequence %s: %dbp ", id, seqBytes.Len())

					sequence := obiseq.NewBioSequence(id,
						seqBytes.Bytes(),
						defBytes.String())
					sequence.SetSource(source)

					if withFeatureTable {
						sequence.SetFeatures(featBytes.Bytes())
					}

					annot := sequence.Annotations()
					annot["scientific_name"] = scientificName
					annot["taxid"] = taxid
					// log.Println(FormatFasta(sequence, FormatFastSeqJsonHeader))
					// log.Debugf("Read sequences %s: %dbp (%d)", sequence.Id(),
					//	sequence.Len(), seqBytes.Len())

					sequences = append(sequences, sequence)

					defBytes = bytes.NewBuffer(obiseq.GetSlice(200))
					featBytes = new(bytes.Buffer)
					nl = 0
					sl = 0

					state = inHeader
					processed = true

				case state == inSequence:
					// log.Debugf("Chunk %d : Genbank: line %d, state = %d : %s", chunks.order, nl, state, line)

					sl++
					parts := strings.SplitN(line[10:], " ", 6)
					lparts := len(parts)
					for i := 0; i < lparts; i++ {
						seqBytes.WriteString(parts[i])
					}
					processed = true

				default:
					switch state {
					case inFeature:
						if withFeatureTable {
							featBytes.WriteByte('\n')
							featBytes.WriteString(line)
						}
						if strings.HasPrefix(line, `                     /db_xref="taxon:`) {
							taxid, _ = strconv.Atoi(strings.SplitN(line[37:], `"`, 2)[0])
						}
						processed = true
					case inHeader:
						processed = true
					case inEntry:
						processed = true
					case inContig:
						processed = true
					default:
						log.Fatalf("Unexpected state %d while reading: %s", state, line)
					}
				}
			}

		}

		return sequences, nil
	}
}

func _ParseGenbankFile(input ChannelSeqFileChunk,
	out obiiter.IBioSequence,
	withFeatureTable bool) {

	parser := GenbankChunkParser(withFeatureTable)

	for chunks := range input {
		sequences, err := parser(chunks.Source, chunks.Raw)

		if err != nil {
			log.Fatalf("File %s : Cannot parse the genbank file : %v", chunks.Source, err)
		}

		out.Push(obiiter.MakeBioSequenceBatch(chunks.Source, chunks.Order, sequences))
	}

	log.Debug("End of the Genbank thread")
	out.Done()

}

func ReadGenbank(reader io.Reader, options ...WithOption) (obiiter.IBioSequence, error) {
	opt := MakeOptions(options)
	// entry_channel := make(chan _FileChunk)

	buff := make([]byte, 1024*1024*512)

	entry_channel := ReadSeqFileChunk(
		opt.Source(),
		reader,
		buff,
		EndOfLastFlatFileEntry,
	)

	newIter := obiiter.MakeIBioSequence()

	nworkers := opt.ParallelWorkers()

	// for j := 0; j < opt.ParallelWorkers(); j++ {
	for j := 0; j < nworkers; j++ {
		newIter.Add(1)
		go _ParseGenbankFile(
			entry_channel,
			newIter,
			opt.WithFeatureTable(),
		)
	}

	// go _ReadFlatFileChunk(reader, entry_channel)

	go func() {
		newIter.WaitAndClose()
		log.Debug("End of the genbank file ", opt.Source())
	}()

	if opt.FullFileBatch() {
		newIter = newIter.CompleteFileIterator()
	}

	return newIter, nil
}

func ReadGenbankFromFile(filename string, options ...WithOption) (obiiter.IBioSequence, error) {
	var reader io.Reader
	var err error

	options = append(options, OptionsSource(obiutils.RemoveAllExt((path.Base(filename)))))

	reader, err = Ropen(filename)

	if err == ErrNoContent {
		log.Infof("file %s is empty", filename)
		return ReadEmptyFile(options...)
	}

	if err != nil {
		log.Printf("open file error: %+v", err)
		return obiiter.NilIBioSequence, err
	}

	return ReadGenbank(reader, options...)
}
