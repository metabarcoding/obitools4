package obiformats

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	gzip "github.com/klauspost/pgzip"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
)

var _FileChunkSize = 1 << 26

type _FileChunk struct {
	raw   io.Reader
	order int
}

// _EndOfLastEntry finds the index of the last entry in the given byte slice 'buff'
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
func _EndOfLastEntry(buff []byte) int {
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

func _ParseEmblFile(source string, input <-chan _FileChunk, out obiiter.IBioSequence) {

	for chunks := range input {
		scanner := bufio.NewScanner(chunks.raw)
		order := chunks.order
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
			case strings.HasPrefix(line, "FH   "):
				featBytes.WriteString(line)
			case line == "FH":
				featBytes.WriteByte('\n')
				featBytes.WriteString(line)
			case strings.HasPrefix(line, "FT   "):
				featBytes.WriteByte('\n')
				featBytes.WriteString(line)
				if strings.HasPrefix(line, `FT                   /db_xref="taxon:`) {
					taxid, _ = strconv.Atoi(strings.SplitN(line[37:], `"`, 2)[0])
				}
			case strings.HasPrefix(line, "     "):
				parts := strings.SplitN(line[5:], " ", 7)
				for i := 0; i < 6; i++ {
					seqBytes.WriteString(parts[i])
				}
			case line == "//":
				sequence := obiseq.NewBioSequence(id,
					seqBytes.Bytes(),
					defBytes.String())
				sequence.SetSource(source)

				sequence.SetFeatures(featBytes.Bytes())

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
		out.Push(obiiter.MakeBioSequenceBatch(order, sequences))
	}

	out.Done()

}

// _ReadFlatFileChunk reads a chunk of data from the given 'reader' and sends it to the
// 'readers' channel as a _FileChunk struct. The function reads from the reader until
// the end of the last entry is found, then sends the chunk to the channel. If the end
// of the last entry is not found in the current chunk, the function reads from the reader
// in 1 MB increments until the end of the last entry is found. The function repeats this
// process until the end of the file is reached.
//
// Arguments:
// reader io.Reader - an io.Reader to read data from
// readers chan _FileChunk - a channel to send the data as a _FileChunk struct
//
// Returns:
// None
func _ReadFlatFileChunk(reader io.Reader, readers chan _FileChunk) {
	var err error
	var buff []byte

	size := 0
	l := 0
	i := 0

	// Initialize the buffer to the size of a chunk of data
	buff = make([]byte, _FileChunkSize)

	// Read from the reader until the end of the last entry is found or the end of the file is reached
	for err == nil {

		// Read from the reader until the buffer is full or the end of the file is reached
		for ; err == nil && l < len(buff); l += size {
			size, err = reader.Read(buff[l:])
		}

		// Create an extended buffer to read from if the end of the last entry is not found in the current buffer
		extbuff := make([]byte, 1<<20)
		buff = buff[:l]
		end := 0
		ic := 0

		// Read from the reader in 1 MB increments until the end of the last entry is found
		for end = _EndOfLastEntry(buff); err == nil && end < 0; end = _EndOfLastEntry(extbuff[:size]) {
			ic++
			size, err = reader.Read(extbuff)
			buff = append(buff, extbuff[:size]...)
		}

		end = _EndOfLastEntry(buff)

		// If an extension was read, log the size and number of extensions

		if len(buff) > 0 {
			remains := buff[end:]
			buff = buff[:end]

			// Send the chunk of data as a _FileChunk struct to the readers channel
			io := bytes.NewBuffer(buff)

			log.Debugf("Flat File chunck : final buff size %d bytes (%d) (%d extensions) -> end = %d\n",
				len(buff),
				io.Cap(),
				ic,
				end,
			)

			readers <- _FileChunk{io, i}
			i++

			// Set the buffer to the size of a chunk of data and copy any remaining data to the new buffer
			buff = make([]byte, _FileChunkSize)
			copy(buff, remains)
			l = len(remains)
		}
	}

	// Close the readers channel when the end of the file is reached
	close(readers)

}

//	6    5  43 2    1
//
// <CR>?<LF>//<CR>?<LF>
func ReadEMBL(reader io.Reader, options ...WithOption) obiiter.IBioSequence {
	opt := MakeOptions(options)
	entry_channel := make(chan _FileChunk)

	newIter := obiiter.MakeIBioSequence()

	nworkers := opt.ParallelWorkers()
	newIter.Add(nworkers)

	go func() {
		newIter.WaitAndClose()
	}()

	// for j := 0; j < opt.ParallelWorkers(); j++ {
	for j := 0; j < nworkers; j++ {
		go _ParseEmblFile(opt.Source(),entry_channel, newIter)
	}

	go _ReadFlatFileChunk(reader, entry_channel)

	if opt.pointer.full_file_batch {
		newIter = newIter.FullFileIterator()
	}

	return newIter
}

func ReadEMBLFromFile(filename string, options ...WithOption) (obiiter.IBioSequence, error) {
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

	return ReadEMBL(reader, options...), nil
}
