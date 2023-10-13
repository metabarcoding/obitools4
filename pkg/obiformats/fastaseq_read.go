package obiformats

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
	"golang.org/x/exp/slices"

	log "github.com/sirupsen/logrus"
)

// lastFastaCut extracts the up to the last sequence cut from a given buffer.
//
// It takes a parameter:
//   - buffer []byte: the buffer to extract the sequence cut from.
//
// It returns two values:
//   - []byte: the extracted sequences.
//   - []byte: the remaining buffer after the sequence cut (the last sequence).
func lastFastaCut(buffer []byte) ([]byte, []byte) {
	imax := len(buffer)
	last := 0
	state := 0
	for i := imax - 1; i >= 0 && state < 2; i-- {
		if state == 0 && buffer[i] == '>' {
			state = 1
			last = i
		} else if state == 1 && (buffer[i] == '\r' || buffer[i] == '\n') {
			state = 2
		} else {
			state = 0
		}
	}

	if state == 2 {
		return buffer[:last], bytes.Clone(buffer[last:])
	}
	return []byte{}, buffer
}

// firstFastaCut cuts the input buffer at the first occurrence of a ">" character
// following a sequence of "\r" or "\n" characters.
//
// It takes a byte slice as input, representing the buffer to be cut.
// It returns two byte slices: the first slice contains the part of the buffer before the cut,
// and the second slice contains the part of the buffer after the cut.
func firstFastaCut(buffer []byte) ([]byte, []byte) {
	imax := len(buffer)
	last := 0
	state := 0
	for i := 0; i < imax && state < 2; i++ {
		if (state == 0 || state == 1) && (buffer[i] == '\r' || buffer[i] == '\n') {
			state = 1
		} else if (state == 1 || i == 0) && buffer[i] == '>' {
			state = 2
			last = i
		} else {
			state = 0
		}
	}

	if state == 2 {
		return bytes.Clone(buffer[:last]), buffer[last:]
	}
	return buffer, []byte{}

}

func Concatenate[S ~[]E, E any](s1, s2 S) S {
	if len(s1) > 0 {
		if len(s2) > 0 {
			return append(s1[:len(s1):len(s1)], s2...)
		}
		return s1
	}
	return s2
}

type FastxChunk struct {
	Bytes []byte
	index int
}

func FastaChunkReader(r io.Reader, size int, cutHead bool) (chan FastxChunk, error) {
	out := make(chan FastxChunk)
	buff := make([]byte, size)

	n, err := r.Read(buff)
	if n > 0 && err == nil {
		if n < size {
			buff = buff[:n]
		}

		begin, buff := firstFastaCut(buff)

		if len(begin) > 0 && !cutHead {
			return out, fmt.Errorf("begin is not empty : %s", string(begin))
		}

		go func(buff []byte) {
			idx := 0
			end := []byte{}

			for err == nil && n > 0 {
				// fmt.Println("============end=========================")
				// fmt.Println(string(end))
				// fmt.Println("------------buff------------------------")
				// fmt.Println(string(buff))
				buff = Concatenate(end, buff)
				// fmt.Println("------------buff--pasted----------------")
				// fmt.Println(string(buff))
				buff, end = lastFastaCut(buff)
				// fmt.Println("----------------buff--cutted------------")
				// fmt.Println(string(buff))
				// fmt.Println("------------------end-------------------")
				// fmt.Println(string(end))
				// fmt.Println("========================================")
				if len(buff) > 0 {
					out <- FastxChunk{
						Bytes: bytes.Clone(buff),
						index: idx,
					}
					idx++
				}

				buff = slices.Grow(buff[:0], size)[0:size]
				n, err = r.Read(buff)
				if n < size {
					buff = buff[:n]
				}
				// fmt.Printf("n = %d, err = %v\n", n, err)
			}

			if len(end) > 0 {
				out <- FastxChunk{
					Bytes: bytes.Clone(end),
					index: idx,
				}
			}

			close(out)
		}(buff)
	}

	return out, nil
}

func ParseFastaChunk(source string, ch FastxChunk) *obiiter.BioSequenceBatch {
	slice := make(obiseq.BioSequenceSlice, 0, obioptions.CLIBatchSize())

	state := 0
	start := 0
	current := 0
	var identifier string
	var definition string

	for i := 0; i < len(ch.Bytes); i++ {
		C := ch.Bytes[i]
		is_end_of_line := C == '\r' || C == '\n'
		is_space := C == ' ' || C == '\t'
		is_sep := is_space || is_end_of_line

		switch state {
		case 0:
			if C == '>' {
				// Beginning of sequence
				state = 1
			}
		case 1:
			if is_sep {
				// No identifier -> ERROR
				log.Errorf("%s : sequence entry does not have an identifier", source)
				return nil
			} else {
				// Beginning of identifier
				state = 2
				start = i
			}
		case 2:
			if is_sep {
				// End of identifier
				identifier = string(ch.Bytes[start:i])
				state = 3
			}
			if is_end_of_line {
				// Definition empty
				definition = ""
				state = 5
			}
		case 3:
			if is_end_of_line {
				// Definition empty
				definition = ""
				state = 5
			} else if !is_space {
				// Beginning of definition
				start = i
				state = 4
			}
		case 4:
			if is_end_of_line {
				definition = string(ch.Bytes[start:i])
				state = 5

			}
		case 5:
			if !is_end_of_line {
				// Beginning of sequence
				start = i
				if C >= 'A' && C <= 'Z' {
					ch.Bytes[current] = C + 'a' - 'A'
				}
				current = i + 1
				state = 6
			}
		case 6:
			if C == '>' {
				// End of sequence
				s := obiseq.NewBioSequence(identifier, bytes.Clone(ch.Bytes[start:current]), definition)
				s.SetSource(source)
				slice = append(slice, s)
				state = 1

			} else if !is_sep {
				if C >= 'A' && C <= 'Z' {
					C = C + 'a' - 'A'
				}
				// Removing white space from the sequence
				if (C >= 'a' && C <= 'z') || C == '-' || C == '.' {
					ch.Bytes[current] = C
					current++
				}
			}
		}
	}

	slice = append(slice, obiseq.NewBioSequence(identifier, bytes.Clone(ch.Bytes[start:current]), definition))
	batch := obiiter.MakeBioSequenceBatch(ch.index, slice)
	return &batch
}

func ReadFasta(reader io.Reader, options ...WithOption) (obiiter.IBioSequence, error) {
	opt := MakeOptions(options)
	out := obiiter.MakeIBioSequence()

	source := opt.Source()

	nworker := obioptions.CLIReadParallelWorkers()
	out.Add(nworker)

	chkchan, err := FastaChunkReader(reader, 1024*500, false)

	if err != nil {
		return obiiter.NilIBioSequence, err
	}

	go func() {
		out.WaitAndClose()
	}()

	parser := func() {
		defer out.Done()
		for chk := range chkchan {
			seqs := ParseFastaChunk(source, chk)
			if seqs != nil {
				out.Push(*seqs)
			}
		}
	}

	for i := 0; i < nworker; i++ {
		go parser()
	}

	newIter := out.SortBatches().Rebatch(opt.BatchSize())

	log.Debugln("Full file batch mode : ", opt.FullFileBatch())
	if opt.FullFileBatch() {
		newIter = newIter.CompleteFileIterator()
	}

	annotParser := opt.ParseFastSeqHeader()

	if annotParser != nil {
		return IParseFastSeqHeaderBatch(newIter, options...), nil
	}

	return newIter, nil
}

func ReadFastaFromFile(filename string, options ...WithOption) (obiiter.IBioSequence, error) {
	options = append(options, OptionsSource(obiutils.RemoveAllExt((path.Base(filename)))))

	file, err := Ropen(filename)

	if err == ErrNoContent {
		log.Infof("file %s is empty", filename)
		return ReadEmptyFile(options...)
	}

	if err != nil {
		return obiiter.NilIBioSequence, err
	}

	return ReadFasta(file, options...)
}

func ReadFastaFromStdin(reader io.Reader, options ...WithOption) (obiiter.IBioSequence, error) {
	options = append(options, OptionsSource(obiutils.RemoveAllExt("stdin")))
	input, err := Buf(os.Stdin)

	if err == ErrNoContent {
		log.Infof("stdin is empty")
		return ReadEmptyFile(options...)
	}

	if err != nil {
		log.Fatalf("open file error: %v", err)
		return obiiter.NilIBioSequence, err
	}

	return ReadFasta(input, options...)
}
