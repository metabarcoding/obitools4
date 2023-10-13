package obiformats

import (
	"bytes"
	"io"
	"os"
	"path"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

func lastFastqCut(buffer []byte) ([]byte, []byte) {
	imax := len(buffer)
	cut := imax
	state := 0
	restart := imax - 1
	for i := restart; i >= 0 && state < 7; i-- {
		C := buffer[i]
		is_end_of_line := C == '\r' || C == '\n'
		is_space := C == ' ' || C == '\t'
		is_sep := is_space || is_end_of_line

		switch state {
		case 0:
			if C == '+' {
				// Potential start of quality part step 1
				state = 1
				restart = i
			}
		case 1:
			if is_end_of_line {
				// Potential start of quality part step 2
				state = 2
			} else {
				// it was not the start of quality part
				state = 0
				i = restart
			}
		case 2:
			if is_sep {
				// Potential start of quality part step 2 (stay in the same state)
				state = 2
			} else if (C >= 'a' && C <= 'z') || C == '-' || C == '.' {
				// End of the sequence
				state = 3
			} else {
				// it was not the start of quality part
				state = 0
				i = restart
			}
		case 3:
			if is_end_of_line {
				// Entrering in the header line
				state = 4
			} else if (C >= 'a' && C <= 'z') || C == '-' || C == '.' {
				// progressing along of the sequence
				state = 3
			} else {
				// it was not the sequence part
				state = 0
				i = restart
			}
		case 4:
			if is_end_of_line {
				state = 4
			} else {
				state = 5
			}
		case 5:
			if is_end_of_line {
				// It was not the header line
				state = 0
				i = restart
			} else if C == '@' {
				state = 6
				cut = i
			}
		case 6:
			if is_end_of_line {
				state = 7
			} else {
				state = 0
				i = restart
			}
		}
	}
	if state == 7 {
		return buffer[:cut], bytes.Clone(buffer[cut:])
	}
	return []byte{}, buffer
}

func FastqChunkReader(r io.Reader, size int) (chan FastxChunk, error) {
	out := make(chan FastxChunk)
	buff := make([]byte, size)

	n, err := io.ReadFull(r, buff)

	if err == io.ErrUnexpectedEOF {
		err = nil
	}

	if n > 0 && err == nil {
		if n < size {
			buff = buff[:n]
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
				buff, end = lastFastqCut(buff)
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
				} else {
					size = size * 2
				}

				buff = slices.Grow(buff[:0], size)[0:size]
				n, err = io.ReadFull(r, buff)
				if n < size {
					buff = buff[:n]
				}

				if err == io.ErrUnexpectedEOF {
					err = nil
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

func ParseFastqChunk(source string, ch FastxChunk, quality_shift byte) *obiiter.BioSequenceBatch {
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

		// log.Infof("%s : state = %d pos = %d character = %c (%d)", source, state, i, C, C)

		switch state {
		case 0: // Beginning of sequence chunk must start with @

			if C == '@' {
				// Beginning of sequence
				state = 1
			} else {
				log.Errorf("%s : sequence entry is not starting with @", source)
				return nil
			}
		case 1: // Beginning of identifier (Mandatory)
			if is_sep {
				// No identifier -> ERROR
				log.Errorf("%s : sequence identifier is empty", source)
				return nil
			} else {
				// Beginning of identifier
				state = 2
				start = i
			}
		case 2: // Following of the identifier
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
		case 3: // Beginning of definition
			if is_end_of_line {
				// Definition empty
				definition = ""
				state = 5
			} else if !is_space {
				// Beginning of definition
				start = i
				state = 4
			}
		case 4: // Following of the definition
			if is_end_of_line {
				definition = string(ch.Bytes[start:i])
				state = 5
			}
		case 5: // Beginning of sequence
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
			if is_end_of_line {
				// End of sequence
				s := obiseq.NewBioSequence(identifier, bytes.Clone(ch.Bytes[start:current]), definition)
				s.SetSource(source)
				slice = append(slice, s)
				state = 7
			} else {
				if C >= 'A' && C <= 'Z' {
					ch.Bytes[current] = C + 'a' - 'A'
				}
				current = i + 1
			}
		case 7:
			if is_end_of_line {
				state = 7
			} else if C == '+' {
				state = 8
			} else {
				log.Info(ch.Bytes[0:i])
				log.Info(string(ch.Bytes[0:i]))
				log.Info(C)
				log.Errorf("@%s[%s] : sequence data not followed by a line starting with +", identifier, source)

				return nil // Error
			}
		case 8:
			if is_end_of_line {
				state = 9
			}
		case 9:
			if is_end_of_line {
				state = 9
			} else {
				// beginning of quality
				state = 10
				start = i
			}
		case 10:
			if is_end_of_line {
				// End of quality
				q := ch.Bytes[start:i]
				if len(q) != slice[len(slice)-1].Len() {
					log.Errorf("%s[%s] : sequence data and quality lenght not equal (%d/%d)",
						identifier, source, len(q), slice[len(slice)-1].Len())
					return nil // Error quality lenght not equal to sequence length
				}
				for i := 0; i < len(q); i++ {
					q[i] = q[i] - quality_shift
				}
				slice[len(slice)-1].SetQualities(q)
				state = 11
			}
		case 11:
			if is_end_of_line {
				state = 11
			} else if C == '@' {
				state = 1
			} else {
				log.Errorf("%s[%s] : sequence record not followed by a line starting with @", identifier, source)
				return nil
			}

		}
	}

	batch := obiiter.MakeBioSequenceBatch(ch.index, slice)
	return &batch
}

func ReadFastq(reader io.Reader, options ...WithOption) (obiiter.IBioSequence, error) {
	opt := MakeOptions(options)
	out := obiiter.MakeIBioSequence()

	source := opt.Source()

	nworker := obioptions.CLIReadParallelWorkers()
	out.Add(nworker)

	chkchan, err := FastqChunkReader(reader, 1024*500)

	if err != nil {
		return obiiter.NilIBioSequence, err
	}

	go func() {
		out.WaitAndClose()
	}()

	parser := func() {
		defer out.Done()
		for chk := range chkchan {
			seqs := ParseFastqChunk(source, chk, byte(opt.QualityShift()))
			if seqs != nil {
				out.Push(*seqs)
			} else {
				log.Fatalf("error parsing %s", source)
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

func ReadFastqFromFile(filename string, options ...WithOption) (obiiter.IBioSequence, error) {
	options = append(options, OptionsSource(obiutils.RemoveAllExt((path.Base(filename)))))

	file, err := Ropen(filename)

	if err == ErrNoContent {
		log.Infof("file %s is empty", filename)
		return ReadEmptyFile(options...)
	}

	if err != nil {
		return obiiter.NilIBioSequence, err
	}

	return ReadFastq(file, options...)
}

func ReadFastqFromStdin(reader io.Reader, options ...WithOption) (obiiter.IBioSequence, error) {
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

	return ReadFastq(input, options...)
}
