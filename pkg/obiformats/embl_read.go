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
	"time"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

var __FILE_CHUNK_SIZE__ = 1 << 20

type __file_chunk__ struct {
	raw   io.Reader
	order int
}

func __end_of_last_entry__(buff []byte) int {
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
	} else {
		return -1
	}
}

func __parse_embl_file__(input <-chan __file_chunk__, out obiseq.IBioSequenceBatch) {

	for chunks := range input {
		scanner := bufio.NewScanner(chunks.raw)
		order := chunks.order
		sequences := make(obiseq.BioSequenceSlice, 0, 100)
		id := ""
		scientific_name := ""
		def_bytes := new(bytes.Buffer)
		feat_bytes := new(bytes.Buffer)
		seq_bytes := new(bytes.Buffer)
		taxid := 1
		for scanner.Scan() {

			line := scanner.Text()

			switch {
			case strings.HasPrefix(line, "ID   "):
				id = strings.SplitN(line[5:], ";", 2)[0]
			case strings.HasPrefix(line, "OS   "):
				scientific_name = strings.TrimSpace(line[5:])
			case strings.HasPrefix(line, "DE   "):
				if def_bytes.Len() > 0 {
					def_bytes.WriteByte(' ')
				}
				def_bytes.WriteString(strings.TrimSpace(line[5:]))
			case strings.HasPrefix(line, "FH   "):
				feat_bytes.WriteString(line)
			case line == "FH":
				feat_bytes.WriteByte('\n')
				feat_bytes.WriteString(line)
			case strings.HasPrefix(line, "FT   "):
				feat_bytes.WriteByte('\n')
				feat_bytes.WriteString(line)
				if strings.HasPrefix(line, `FT                   /db_xref="taxon:`) {
					taxid, _ = strconv.Atoi(strings.SplitN(line[37:], `"`, 2)[0])
				}
			case strings.HasPrefix(line, "     "):
				parts := strings.SplitN(line[5:], " ", 7)
				for i := 0; i < 6; i++ {
					seq_bytes.WriteString(parts[i])
				}
			case line == "//":
				sequence := obiseq.MakeBioSequence(id,
					seq_bytes.Bytes(),
					def_bytes.String())

				sequence.SetFeatures(feat_bytes.String())

				annot := sequence.Annotations()
				annot["scientific_name"] = scientific_name
				annot["taxid"] = taxid
				// log.Println(FormatFasta(sequence, FormatFastSeqJsonHeader))
				sequences = append(sequences, sequence)
				def_bytes = new(bytes.Buffer)
				feat_bytes = new(bytes.Buffer)
				seq_bytes = new(bytes.Buffer)
			}
		}
		out.Channel() <- obiseq.MakeBioSequenceBatch(order, sequences...)

	}

	out.Done()

}

func __read_flat_file_chunk__(reader io.Reader, readers chan __file_chunk__) {
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
		end := __end_of_last_entry__(buff)
		remains := buff[end:]
		buff = buff[:end]
		io := bytes.NewBuffer(buff)
		readers <- __file_chunk__{io, i}
		i++
		buff = make([]byte, __FILE_CHUNK_SIZE__)
		copy(buff, remains)
		l = len(remains)
	}

	close(readers)
}

//  6    5  43 2    1
// <CR>?<LF>//<CR>?<LF>
func ReadEMBLBatch(reader io.Reader, options ...WithOption) obiseq.IBioSequenceBatch {
	opt := MakeOptions(options)
	entry_channel := make(chan __file_chunk__, opt.BufferSize())

	new_iter := obiseq.MakeIBioSequenceBatch(opt.BufferSize())

	// new_iter.Add(opt.ParallelWorkers())
	new_iter.Add(2)

	go func() {
		new_iter.Wait()
		for len(new_iter.Channel()) > 0 {
			time.Sleep(time.Millisecond)
		}
		close(new_iter.Channel())
	}()

	// for j := 0; j < opt.ParallelWorkers(); j++ {
	for j := 0; j < 2; j++ {
		go __parse_embl_file__(entry_channel, new_iter)
	}

	go __read_flat_file_chunk__(reader, entry_channel)

	return new_iter
}

func ReadEMBL(reader io.Reader, options ...WithOption) obiseq.IBioSequence {
	ib := ReadEMBLBatch(reader, options...)
	return ib.SortBatches().IBioSequence()
}

func ReadEMBLBatchFromFile(filename string, options ...WithOption) (obiseq.IBioSequenceBatch, error) {
	var reader io.Reader
	var greader io.Reader
	var err error

	reader, err = os.Open(filename)
	if err != nil {
		log.Printf("open file error: %+v", err)
		return obiseq.NilIBioSequenceBatch, err
	}

	// Test if the flux is compressed by gzip
	greader, err = gzip.NewReader(reader)
	if err == nil {
		reader = greader
	}

	return ReadEMBLBatch(reader, options...), nil
}

func ReadEMBLFromFile(filename string, options ...WithOption) (obiseq.IBioSequence, error) {
	ib, err := ReadEMBLBatchFromFile(filename, options...)
	return ib.SortBatches().IBioSequence(), err

}
