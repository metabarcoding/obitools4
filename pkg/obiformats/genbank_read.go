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

// gbRopeScanner reads lines from a PieceOfChunk rope without heap allocation.
// The carry buffer (stack) handles lines that span two rope nodes.
type gbRopeScanner struct {
	current *PieceOfChunk
	pos     int
	carry   [256]byte // max GenBank line = 80 chars; 256 gives ample margin
	carryN  int
}

func newGbRopeScanner(rope *PieceOfChunk) *gbRopeScanner {
	return &gbRopeScanner{current: rope}
}

// ReadLine returns the next line without the trailing \n (or \r\n).
// Returns nil at end of rope. The returned slice aliases carry[] or the node
// data and is valid only until the next ReadLine call.
func (s *gbRopeScanner) ReadLine() []byte {
	for {
		if s.current == nil {
			if s.carryN > 0 {
				n := s.carryN
				s.carryN = 0
				return s.carry[:n]
			}
			return nil
		}

		data := s.current.data[s.pos:]
		idx := bytes.IndexByte(data, '\n')

		if idx >= 0 {
			var line []byte
			if s.carryN == 0 {
				line = data[:idx]
			} else {
				n := copy(s.carry[s.carryN:], data[:idx])
				s.carryN += n
				line = s.carry[:s.carryN]
				s.carryN = 0
			}
			s.pos += idx + 1
			if s.pos >= len(s.current.data) {
				s.current = s.current.Next()
				s.pos = 0
			}
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			return line
		}

		// No \n in this node: accumulate into carry and advance
		n := copy(s.carry[s.carryN:], data)
		s.carryN += n
		s.current = s.current.Next()
		s.pos = 0
	}
}

// extractSequence scans the ORIGIN section byte-by-byte directly on the rope,
// appending compacted bases to dest. Returns the extended slice.
// Stops and returns when "//" is found at the start of a line.
// The scanner is left positioned after the "//" line.
func (s *gbRopeScanner) extractSequence(dest []byte, UtoT bool) []byte {
	lineStart := true
	skipDigits := true

	for s.current != nil {
		data := s.current.data[s.pos:]
		for i, b := range data {
			if lineStart {
				if b == '/' {
					// End-of-record marker "//"
					s.pos += i + 1
					if s.pos >= len(s.current.data) {
						s.current = s.current.Next()
						s.pos = 0
					}
					s.skipToNewline()
					return dest
				}
				lineStart = false
				skipDigits = true
			}
			switch {
			case b == '\n':
				lineStart = true
			case b == '\r':
				// skip
			case skipDigits:
				if b != ' ' && (b < '0' || b > '9') {
					skipDigits = false
					if UtoT && b == 'u' {
						b = 't'
					}
					dest = append(dest, b)
				}
			case b != ' ':
				if UtoT && b == 'u' {
					b = 't'
				}
				dest = append(dest, b)
			}
		}
		s.current = s.current.Next()
		s.pos = 0
	}
	return dest
}

// skipToNewline advances the scanner past the next '\n'.
func (s *gbRopeScanner) skipToNewline() {
	for s.current != nil {
		data := s.current.data[s.pos:]
		idx := bytes.IndexByte(data, '\n')
		if idx >= 0 {
			s.pos += idx + 1
			if s.pos >= len(s.current.data) {
				s.current = s.current.Next()
				s.pos = 0
			}
			return
		}
		s.current = s.current.Next()
		s.pos = 0
	}
}

// parseLseqFromLocus extracts the declared sequence length from a LOCUS line.
// Format: "LOCUS       <id> <length> bp ..."
// Returns -1 if not found or parse error.
func parseLseqFromLocus(line []byte) int {
	if len(line) < 13 {
		return -1
	}
	i := 12
	for i < len(line) && line[i] != ' ' {
		i++
	}
	for i < len(line) && line[i] == ' ' {
		i++
	}
	start := i
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	if i == start {
		return -1
	}
	n, err := strconv.Atoi(string(line[start:i]))
	if err != nil {
		return -1
	}
	return n
}

// Prefix constants for GenBank section headers (byte slices for zero-alloc comparison).
var (
	gbPfxLocus      = []byte("LOCUS       ")
	gbPfxDefinition = []byte("DEFINITION  ")
	gbPfxContinue   = []byte("            ")
	gbPfxSource     = []byte("SOURCE      ")
	gbPfxFeatures   = []byte("FEATURES    ")
	gbPfxOrigin     = []byte("ORIGIN")
	gbPfxContig     = []byte("CONTIG")
	gbPfxEnd        = []byte("//")
	gbPfxDbXref     = []byte(`                     /db_xref="taxon:`)
)

// GenbankChunkParserRope parses a GenBank FileChunk directly from the rope
// (PieceOfChunk linked list) without calling Pack(). This eliminates the large
// contiguous allocation required for chromosomal-scale sequences.
func GenbankChunkParserRope(source string, rope *PieceOfChunk,
	withFeatureTable, UtoT bool) (obiseq.BioSequenceSlice, error) {

	state := inHeader
	scanner := newGbRopeScanner(rope)
	sequences := obiseq.MakeBioSequenceSlice(100)[:0]

	id := ""
	lseq := -1
	scientificName := ""
	defBytes := new(bytes.Buffer)
	featBytes := new(bytes.Buffer)
	var seqDest []byte
	taxid := 1
	nl := 0

	for bline := scanner.ReadLine(); bline != nil; bline = scanner.ReadLine() {
		nl++
		processed := false
		for !processed {
			switch {

			case bytes.HasPrefix(bline, gbPfxLocus):
				if state != inHeader {
					log.Fatalf("Line %d - Unexpected state %d while reading LOCUS: %s", nl, state, bline)
				}
				rest := bline[12:]
				sp := bytes.IndexByte(rest, ' ')
				if sp < 0 {
					id = string(rest)
				} else {
					id = string(rest[:sp])
				}
				lseq = parseLseqFromLocus(bline)
				cap0 := lseq + 20
				if cap0 < 1024 {
					cap0 = 1024
				}
				seqDest = make([]byte, 0, cap0)
				state = inEntry
				processed = true

			case bytes.HasPrefix(bline, gbPfxDefinition):
				if state != inEntry {
					log.Fatalf("Line %d - Unexpected state %d while reading DEFINITION: %s", nl, state, bline)
				}
				defBytes.Write(bytes.TrimSpace(bline[12:]))
				state = inDefinition
				processed = true

			case state == inDefinition:
				if bytes.HasPrefix(bline, gbPfxContinue) {
					defBytes.WriteByte(' ')
					defBytes.Write(bytes.TrimSpace(bline[12:]))
					processed = true
				} else {
					state = inEntry
				}

			case bytes.HasPrefix(bline, gbPfxSource):
				if state != inEntry {
					log.Fatalf("Line %d - Unexpected state %d while reading SOURCE: %s", nl, state, bline)
				}
				scientificName = string(bytes.TrimSpace(bline[12:]))
				processed = true

			case bytes.HasPrefix(bline, gbPfxFeatures):
				if state != inEntry {
					log.Fatalf("Line %d - Unexpected state %d while reading FEATURES: %s", nl, state, bline)
				}
				if withFeatureTable {
					featBytes.Write(bline)
				}
				state = inFeature
				processed = true

			case bytes.HasPrefix(bline, gbPfxOrigin):
				if state != inFeature && state != inContig {
					log.Fatalf("Line %d - Unexpected state %d while reading ORIGIN: %s", nl, state, bline)
				}
				// Use fast byte-scan to extract sequence and consume through "//"
				seqDest = scanner.extractSequence(seqDest, UtoT)
				// Emit record
				if id == "" {
					log.Warn("Empty id when parsing genbank file")
				}
				sequence := obiseq.NewBioSequenceOwning(id, seqDest, defBytes.String())
				sequence.SetSource(source)
				if withFeatureTable {
					sequence.SetFeatures(featBytes.Bytes())
				}
				annot := sequence.Annotations()
				annot["scientific_name"] = scientificName
				annot["taxid"] = taxid
				sequences = append(sequences, sequence)

				defBytes = bytes.NewBuffer(obiseq.GetSlice(200))
				featBytes = new(bytes.Buffer)
				nl = 0
				taxid = 1
				seqDest = nil
				state = inHeader
				processed = true

			case bytes.HasPrefix(bline, gbPfxContig):
				if state != inFeature && state != inContig {
					log.Fatalf("Line %d - Unexpected state %d while reading CONTIG: %s", nl, state, bline)
				}
				state = inContig
				processed = true

			case bytes.Equal(bline, gbPfxEnd):
				// Reached for CONTIG records (no ORIGIN section)
				if state != inContig {
					log.Fatalf("Line %d - Unexpected state %d while reading end of record %s", nl, state, id)
				}
				if id == "" {
					log.Warn("Empty id when parsing genbank file")
				}
				sequence := obiseq.NewBioSequenceOwning(id, seqDest, defBytes.String())
				sequence.SetSource(source)
				if withFeatureTable {
					sequence.SetFeatures(featBytes.Bytes())
				}
				annot := sequence.Annotations()
				annot["scientific_name"] = scientificName
				annot["taxid"] = taxid
				sequences = append(sequences, sequence)

				defBytes = bytes.NewBuffer(obiseq.GetSlice(200))
				featBytes = new(bytes.Buffer)
				nl = 0
				taxid = 1
				seqDest = nil
				state = inHeader
				processed = true

			default:
				switch state {
				case inFeature:
					if withFeatureTable {
						featBytes.WriteByte('\n')
						featBytes.Write(bline)
					}
					if bytes.HasPrefix(bline, gbPfxDbXref) {
						rest := bline[len(gbPfxDbXref):]
						q := bytes.IndexByte(rest, '"')
						if q >= 0 {
							taxid, _ = strconv.Atoi(string(rest[:q]))
						}
					}
					processed = true
				case inHeader, inEntry, inContig:
					processed = true
				default:
					log.Fatalf("Unexpected state %d while reading: %s", state, bline)
				}
			}
		}
	}

	return sequences, nil
}

func GenbankChunkParser(withFeatureTable, UtoT bool) func(string, io.Reader) (obiseq.BioSequenceSlice, error) {
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
						log.Fatalf("Line %d - Unexpected state %d while reading LOCUS: %s", nl, state, line)
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
						log.Fatalf("Line %d - Unexpected state %d while reading DEFINITION: %s", nl, state, line)
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
						log.Fatalf("Line %d - Unexpected state %d while reading SOURCE: %s", nl, state, line)
					}
					scientificName = strings.TrimSpace(line[12:])
					processed = true

				case strings.HasPrefix(line, "FEATURES    "):
					if state != inEntry {
						log.Fatalf("Line %d - Unexpected state %d while reading FEATURES: %s", nl, state, line)
					}
					featBytes.WriteString(line)
					state = inFeature
					processed = true

				case strings.HasPrefix(line, "ORIGIN"):
					if state != inFeature && state != inContig {
						log.Fatalf("Line %d - Unexpected state %d while reading ORIGIN: %s", nl, state, line)
					}
					state = inSequence
					processed = true

				case strings.HasPrefix(line, "CONTIG"):
					if state != inFeature && state != inContig {
						log.Fatalf("Line %d - Unexpected state %d while reading ORIGIN: %s", nl, state, line)
					}
					state = inContig
					processed = true

				case line == "//":

					if state != inSequence && state != inContig {
						log.Fatalf("Line %d - Unexpected state %d while reading end of record %s", nl, state, id)
					}
					if id == "" {
						log.Warn("Empty id when parsing genbank file")
					}

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

					sequences = append(sequences, sequence)

					defBytes = bytes.NewBuffer(obiseq.GetSlice(200))
					featBytes = new(bytes.Buffer)
					nl = 0
					sl = 0

					state = inHeader
					processed = true

				case state == inSequence:
					sl++
					cleanline := strings.TrimSpace(line)
					parts := strings.SplitN(cleanline, " ", 7)
					lparts := len(parts)
					for i := 1; i < lparts; i++ {
						if UtoT {
							parts[i] = strings.ReplaceAll(parts[i], "u", "t")
						}
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

		_ = sl
		return sequences, nil
	}
}

func _ParseGenbankFile(input ChannelFileChunk,
	out obiiter.IBioSequence,
	withFeatureTable, UtoT bool) {

	for chunks := range input {
		var sequences obiseq.BioSequenceSlice
		var err error

		if chunks.Rope != nil {
			sequences, err = GenbankChunkParserRope(chunks.Source, chunks.Rope, withFeatureTable, UtoT)
		} else {
			parser := GenbankChunkParser(withFeatureTable, UtoT)
			sequences, err = parser(chunks.Source, chunks.Raw)
		}

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

	entry_channel := ReadFileChunk(
		opt.Source(),
		reader,
		1024*1024*128,
		EndOfLastFlatFileEntry,
		"\nLOCUS       ",
		false, // do not pack: rope-based parser avoids contiguous allocation
	)

	newIter := obiiter.MakeIBioSequence()

	nworkers := opt.ParallelWorkers()

	for j := 0; j < nworkers; j++ {
		newIter.Add(1)
		go _ParseGenbankFile(
			entry_channel,
			newIter,
			opt.WithFeatureTable(),
			opt.UtoT(),
		)
	}

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

	reader, err = obiutils.Ropen(filename)

	if err == obiutils.ErrNoContent {
		log.Infof("file %s is empty", filename)
		return ReadEmptyFile(options...)
	}

	if err != nil {
		log.Printf("open file error: %+v", err)
		return obiiter.NilIBioSequence, err
	}

	return ReadGenbank(reader, options...)
}
