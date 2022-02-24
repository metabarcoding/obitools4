package obiformats

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

var _FileChunkSize = 1 << 20

type _FileChunk struct {
	raw   io.Reader
	order int
}

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

func _ParseEmblFile(input <-chan _FileChunk, out obiiter.IBioSequenceBatch) {

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

func _ReadFlatFileChunk(reader io.Reader, readers chan _FileChunk) {
	var err error
	var buff []byte

	size := 0
	l := 0
	i := 0

	buff = make([]byte, 1<<20)
	for err == nil {
		for ; err == nil && l < len(buff); l += size {
			size, err = reader.Read(buff[l:])
		}
		buff = buff[:l]
		end := _EndOfLastEntry(buff)
		remains := buff[end:]
		buff = buff[:end]
		io := bytes.NewBuffer(buff)
		readers <- _FileChunk{io, i}
		i++
		buff = make([]byte, _FileChunkSize)
		copy(buff, remains)
		l = len(remains)
	}

	close(readers)
}

//  6    5  43 2    1
// <CR>?<LF>//<CR>?<LF>
func ReadEMBLBatch(reader io.Reader, options ...WithOption) obiiter.IBioSequenceBatch {
	opt := MakeOptions(options)
	entry_channel := make(chan _FileChunk, opt.BufferSize())

	newIter := obiiter.MakeIBioSequenceBatch(opt.BufferSize())

	nworkers := opt.ParallelWorkers()
	newIter.Add(nworkers)

	go func() {
		newIter.WaitAndClose()
	}()

	// for j := 0; j < opt.ParallelWorkers(); j++ {
	for j := 0; j < nworkers; j++ {
		go _ParseEmblFile(entry_channel, newIter)
	}

	go _ReadFlatFileChunk(reader, entry_channel)

	return newIter
}

func ReadEMBL(reader io.Reader, options ...WithOption) obiiter.IBioSequence {
	ib := ReadEMBLBatch(reader, options...)
	return ib.SortBatches().IBioSequence()
}

func ReadEMBLBatchFromFile(filename string, options ...WithOption) (obiiter.IBioSequenceBatch, error) {
	var reader io.Reader
	var greader io.Reader
	var err error

	reader, err = os.Open(filename)
	if err != nil {
		log.Printf("open file error: %+v", err)
		return obiiter.NilIBioSequenceBatch, err
	}

	// Test if the flux is compressed by gzip
	greader, err = gzip.NewReader(reader)
	if err == nil {
		reader = greader
	}

	return ReadEMBLBatch(reader, options...), nil
}

func ReadEMBLFromFile(filename string, options ...WithOption) (obiiter.IBioSequence, error) {
	ib, err := ReadEMBLBatchFromFile(filename, options...)
	return ib.SortBatches().IBioSequence(), err

}
