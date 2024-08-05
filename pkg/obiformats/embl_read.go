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

func EmblChunkParser(withFeatureTable bool) func(string, io.Reader) (obiseq.BioSequenceSlice, error) {
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

func _ParseEmblFile(
	input ChannelSeqFileChunk,
	out obiiter.IBioSequence,
	withFeatureTable bool,
) {

	parser := EmblChunkParser(withFeatureTable)

	for chunks := range input {
		order := chunks.Order
		sequences, err := parser(chunks.Source, chunks.Raw)

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
		go _ParseEmblFile(
			entry_channel,
			newIter,
			opt.WithFeatureTable(),
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

	reader, err = Ropen(filename)

	if err == ErrNoContent {
		log.Infof("file %s is empty", filename)
		return ReadEmptyFile(options...)
	}

	if err != nil {
		log.Printf("open file error: %+v", err)
		return obiiter.NilIBioSequence, err
	}

	return ReadEMBL(reader, options...)
}
