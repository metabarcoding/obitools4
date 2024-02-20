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
)

var _seqlenght_rx = regexp.MustCompile(" +([0-9]+) bp")

func _ParseGenbankFile(source string,
	input <-chan _FileChunk, out obiiter.IBioSequence,
	chunck_order func() int,
	withFeatureTable bool,
	batch_size int,
	total_seq_size int) {
	state := inHeader
	previous_chunk := -1

	for chunks := range input {

		if state != inHeader {
			log.Fatalf("Unexpected state %d starting new chunk (id = %d, previous_chunk = %d)",
				state, chunks.order, previous_chunk)
		}

		previous_chunk = chunks.order
		scanner := bufio.NewReader(chunks.raw)
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
		var line string
		for bline, is_prefix, err := scanner.ReadLine(); err != io.EOF; bline, is_prefix, err = scanner.ReadLine() {
			nl++
			line = string(bline)
			if is_prefix || len(line) > 100 {
				log.Fatalf("Chunk %d : Line too long: %s", chunks.order, line)
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

				case line == "//":

					if state != inSequence {
						log.Fatalf("Unexpected state %d while reading end of record %s", state, id)
					}
					// log.Debugln("Total lines := ", nl)
					if id == "" {
						log.Warn("Empty id when parsing genbank file")
					}
					if seqBytes.Len() == 0 {
						log.Warn("Empty sequence when parsing genbank file")
					}

					log.Debugf("End of sequence %s: %dbp ", id, seqBytes.Len())

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
					sumlength += sequence.Len()

					if len(sequences) == batch_size || sumlength > total_seq_size {
						log.Debugln("Pushing sequences")
						out.Push(obiiter.MakeBioSequenceBatch(chunck_order(), sequences))
						sequences = make(obiseq.BioSequenceSlice, 0, 100)
						sumlength = 0
					}

					defBytes = bytes.NewBuffer(obiseq.GetSlice(200))
					featBytes = new(bytes.Buffer)
					nl = 0
					sl = 0

					state = inHeader
					processed = true

				case state == inSequence:
					log.Debugf("Chunk %d : Genbank: line %d, state = %d : %s", chunks.order, nl, state, line)

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
					}
				}
			}

		}

		log.Debugf("End of chunk %d : %s", chunks.order, line)
		if len(sequences) > 0 {
			log.Debugln("Pushing sequences")
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
		go _ParseGenbankFile(opt.Source(), entry_channel, newIter, chunck_order,
			opt.WithFeatureTable(), opt.BatchSize(), opt.TotalSeqSize())
	}

	go _ReadFlatFileChunk(reader, entry_channel)

	if opt.pointer.full_file_batch {
		newIter = newIter.CompleteFileIterator()
	}

	return newIter
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

	return ReadGenbank(reader, options...), nil
}
