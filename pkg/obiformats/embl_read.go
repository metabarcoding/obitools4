package obiformats

import (
	"bufio"
	"bytes"
	"io"
	"path"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

// EndOfLastFlatFileEntry finds the index of the last entry in the given byte slice 'buff'
// using a pattern match of the form:
// <CR>?<LF>//<CR>?<LF>
// where <CR> and <LF> are the ASCII codes for carriage return and line feed,
// respectively. The function returns the index of the end of the last entry
// or -1 if no match is found.
//
// Arguments:
// buff []byte - a byte slice to search for the end of the last entry
//
// Returns:
// int - the index of the end of the last entry or -1 if no match is found.
func EndOfLastFlatFileEntry(buff []byte) int {
	//  6    5  43 2    1
	// <CR>?<LF>//<CR>?<LF>
	var i int
	var state = 0
	var start = 0
	for i = len(buff) - 1; i >= 0 && state < 5; i-- {
		switch state {
		case 0: // outside of the pattern
			if buff[i] == '\n' {
				state = 1
			}
		case 1: // a \n have been matched
			start = i + 2
			switch buff[i] {
			case '\r':
				state = 2
			case '/':
				state = 3
			case '\n':
				state = 1
			default:
				state = 0
			}
		case 2: // a \r have been matched
			switch buff[i] {
			case '/':
				state = 3
			case '\n':
				state = 1
			default:
				state = 0
			}
		case 3: // the first / have been matched
			switch buff[i] {
			case '/':
				state = 4
			case '\n':
				state = 1
			default:
				state = 0
			}
		case 4: // the second / have been matched
			switch buff[i] {
			case '\n':
				state = 5
			default:
				state = 0
			}
		}

	}

	if i > 0 {
		return start
	}

	return -1
}

func EmblChunkParser(withFeatureTable, UtoT bool) func(string, io.Reader) (obiseq.BioSequenceSlice, error) {
	parser := func(source string, input io.Reader) (obiseq.BioSequenceSlice, error) {
		scanner := bufio.NewScanner(input)
		sequences := make(obiseq.BioSequenceSlice, 0, 100)
		id := ""
		scientificName := ""
		defBytes := new(bytes.Buffer)
		featBytes := new(bytes.Buffer)
		seqBytes := new(bytes.Buffer)
		taxid := 1
		for scanner.Scan() {

			line := scanner.Text()

			switch {
			case strings.HasPrefix(line, "ID   "):
				id = strings.SplitN(line[5:], ";", 2)[0]
			case strings.HasPrefix(line, "OS   "):
				scientificName = strings.TrimSpace(line[5:])
			case strings.HasPrefix(line, "DE   "):
				if defBytes.Len() > 0 {
					defBytes.WriteByte(' ')
				}
				defBytes.WriteString(strings.TrimSpace(line[5:]))
			case withFeatureTable && strings.HasPrefix(line, "FH   "):
				featBytes.WriteString(line)
			case withFeatureTable && line == "FH":
				featBytes.WriteByte('\n')
				featBytes.WriteString(line)
			case strings.HasPrefix(line, "FT   "):
				if withFeatureTable {
					featBytes.WriteByte('\n')
					featBytes.WriteString(line)
				}
				if strings.HasPrefix(line, `FT                   /db_xref="taxon:`) {
					taxid, _ = strconv.Atoi(strings.SplitN(line[37:], `"`, 2)[0])
				}
			case strings.HasPrefix(line, "     "):
				parts := strings.SplitN(line[5:], " ", 7)
				np := len(parts) - 1
				for i := 0; i < np; i++ {
					if UtoT {
						parts[i] = strings.ReplaceAll(parts[i], "u", "t")
					}
					seqBytes.WriteString(parts[i])
				}
			case line == "//":
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
				sequences = append(sequences, sequence)
				defBytes = new(bytes.Buffer)
				featBytes = new(bytes.Buffer)
				seqBytes = new(bytes.Buffer)
			}
		}

		return sequences, nil

	}

	return parser
}

// extractEmblSeq scans the sequence section of an EMBL record directly on the
// rope. EMBL sequence lines start with 5 spaces followed by bases in groups of
// 10, separated by spaces, with a position number at the end. The section ends
// with "//".
func (s *ropeScanner) extractEmblSeq(dest []byte, UtoT bool) []byte {
	// We use ReadLine and scan each line for bases (skip digits, spaces, newlines).
	for {
		line := s.ReadLine()
		if line == nil {
			break
		}
		if len(line) >= 2 && line[0] == '/' && line[1] == '/' {
			break
		}
		// Lines start with 5 spaces; bases follow separated by single spaces.
		// Digits at the end are the position counter — skip them.
		// Simplest: take every byte that is a letter.
		for _, b := range line {
			if b >= 'A' && b <= 'Z' {
				b += 'a' - 'A'
			}
			if UtoT && b == 'u' {
				b = 't'
			}
			if b >= 'a' && b <= 'z' {
				dest = append(dest, b)
			}
		}
	}
	return dest
}

// EmblChunkParserRope parses an EMBL chunk directly from a rope without Pack().
func EmblChunkParserRope(source string, rope *PieceOfChunk, withFeatureTable, UtoT bool) (obiseq.BioSequenceSlice, error) {
	scanner := newRopeScanner(rope)
	sequences := obiseq.MakeBioSequenceSlice(100)[:0]

	var id string
	var scientificName string
	defBytes := make([]byte, 0, 256)
	featBytes := make([]byte, 0, 1024)
	var taxid int
	inSeq := false

	for {
		line := scanner.ReadLine()
		if line == nil {
			break
		}

		if inSeq {
			// Should not happen — extractEmblSeq consumed up to "//"
			inSeq = false
			continue
		}

		switch {
		case bytes.HasPrefix(line, []byte("ID   ")):
			id = string(bytes.SplitN(line[5:], []byte(";"), 2)[0])
		case bytes.HasPrefix(line, []byte("OS   ")):
			scientificName = string(bytes.TrimSpace(line[5:]))
		case bytes.HasPrefix(line, []byte("DE   ")):
			if len(defBytes) > 0 {
				defBytes = append(defBytes, ' ')
			}
			defBytes = append(defBytes, bytes.TrimSpace(line[5:])...)
		case withFeatureTable && bytes.HasPrefix(line, []byte("FH   ")):
			featBytes = append(featBytes, line...)
		case withFeatureTable && bytes.Equal(line, []byte("FH")):
			featBytes = append(featBytes, '\n')
			featBytes = append(featBytes, line...)
		case bytes.HasPrefix(line, []byte("FT   ")):
			if withFeatureTable {
				featBytes = append(featBytes, '\n')
				featBytes = append(featBytes, line...)
			}
			if bytes.HasPrefix(line, []byte(`FT                   /db_xref="taxon:`)) {
				rest := line[37:]
				end := bytes.IndexByte(rest, '"')
				if end > 0 {
					taxid, _ = strconv.Atoi(string(rest[:end]))
				}
			}
		case bytes.HasPrefix(line, []byte("     ")):
			// First sequence line: extract all bases via extractEmblSeq,
			// which also consumes this line's remaining content.
			// But ReadLine already consumed this line — we need to process it
			// plus subsequent lines. Process this line inline then call helper.
			seqDest := make([]byte, 0, 4096)
			for _, b := range line {
				if b >= 'A' && b <= 'Z' {
					b += 'a' - 'A'
				}
				if UtoT && b == 'u' {
					b = 't'
				}
				if b >= 'a' && b <= 'z' {
					seqDest = append(seqDest, b)
				}
			}
			seqDest = scanner.extractEmblSeq(seqDest, UtoT)

			seq := obiseq.NewBioSequenceOwning(id, seqDest, string(defBytes))
			seq.SetSource(source)
			if withFeatureTable {
				seq.SetFeatures(featBytes)
			}
			annot := seq.Annotations()
			annot["scientific_name"] = scientificName
			annot["taxid"] = taxid
			sequences = append(sequences, seq)

			// Reset state
			id = ""
			scientificName = ""
			defBytes = defBytes[:0]
			featBytes = featBytes[:0]
			taxid = 1

		case bytes.Equal(line, []byte("//")):
			// record ended without SQ/sequence section (e.g. WGS entries)
			if id != "" {
				seq := obiseq.NewBioSequenceOwning(id, []byte{}, string(defBytes))
				seq.SetSource(source)
				if withFeatureTable {
					seq.SetFeatures(featBytes)
				}
				annot := seq.Annotations()
				annot["scientific_name"] = scientificName
				annot["taxid"] = taxid
				sequences = append(sequences, seq)
			}
			id = ""
			scientificName = ""
			defBytes = defBytes[:0]
			featBytes = featBytes[:0]
			taxid = 1
		}
	}

	return sequences, nil
}

func _ParseEmblFile(
	input ChannelFileChunk,
	out obiiter.IBioSequence,
	withFeatureTable, UtoT bool,
) {

	parser := EmblChunkParser(withFeatureTable, UtoT)

	for chunks := range input {
		order := chunks.Order
		var sequences obiseq.BioSequenceSlice
		var err error

		if chunks.Rope != nil {
			sequences, err = EmblChunkParserRope(chunks.Source, chunks.Rope, withFeatureTable, UtoT)
		} else {
			sequences, err = parser(chunks.Source, chunks.Raw)
		}

		if err != nil {
			log.Fatalf("%s : Cannot parse the embl file : %v", chunks.Source, err)
		}

		out.Push(obiiter.MakeBioSequenceBatch(chunks.Source, order, sequences))
	}

	out.Done()

}

//	6    5  43 2    1
//
// <CR>?<LF>//<CR>?<LF>
func ReadEMBL(reader io.Reader, options ...WithOption) (obiiter.IBioSequence, error) {
	opt := MakeOptions(options)

	entry_channel := ReadFileChunk(
		opt.Source(),
		reader,
		1024*1024*128,
		EndOfLastFlatFileEntry,
		"\nID   ",
		false,
	)

	newIter := obiiter.MakeIBioSequence()

	nworkers := opt.ParallelWorkers()

	// for j := 0; j < opt.ParallelWorkers(); j++ {
	for j := 0; j < nworkers; j++ {
		newIter.Add(1)
		go _ParseEmblFile(
			entry_channel,
			newIter,
			opt.WithFeatureTable(),
			opt.UtoT(),
		)
	}

	go func() {
		newIter.WaitAndClose()
	}()

	if opt.pointer.full_file_batch {
		newIter = newIter.CompleteFileIterator()
	}

	return newIter, nil
}

func ReadEMBLFromFile(filename string, options ...WithOption) (obiiter.IBioSequence, error) {
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

	return ReadEMBL(reader, options...)
}
