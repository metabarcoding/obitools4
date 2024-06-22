package obiformats

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

func _EndOfLastFastqEntry(buffer []byte) int {
	var i int

	imax := len(buffer)
	state := 0
	restart := imax - 1
	cut := imax

	for i = imax - 1; i >= 0 && state < 7; i-- {
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
			} else if (C >= 'a' && C <= 'z') || (C >= 'A' && C <= 'Z') || C == '-' || C == '.' || C == '[' || C == ']' {
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
			} else if (C >= 'a' && C <= 'z') || (C >= 'A' && C <= 'Z') || C == '-' || C == '.' || C == '[' || C == ']' {
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

	if i == 0 || state != 7 {
		return -1
	}

	return cut
}

func _storeSequenceQuality(bytes *bytes.Buffer, out *obiseq.BioSequence, quality_shift byte) {
	q := bytes.Bytes()
	if len(q) == 0 {
		log.Fatalf("@%s[%s] : sequence quality is empty", out.Id(), out.Source())
	}

	if len(q) != out.Len() {
		log.Fatalf("%s[%s] : sequence data and quality lenght not equal (%d <> %d)",
			out.Id(), out.Source(), len(q), out.Len())
	}

	for i := 0; i < len(q); i++ {
		q[i] -= quality_shift
	}
	out.SetQualities(q)
}

func _ParseFastqFile(source string,
	input ChannelSeqFileChunk,
	out obiiter.IBioSequence,
	quality_shift byte,
	no_order bool,
	batch_size int,
	chunck_order func() int,
) {

	var identifier string
	var definition string

	idBytes := bytes.Buffer{}
	defBytes := bytes.Buffer{}
	qualBytes := bytes.Buffer{}
	seqBytes := bytes.Buffer{}

	for chunks := range input {
		state := 0
		scanner := bufio.NewReader(chunks.raw)
		sequences := make(obiseq.BioSequenceSlice, 0, 100)
		previous := byte(0)

		for C, err := scanner.ReadByte(); err != io.EOF; C, err = scanner.ReadByte() {

			is_end_of_line := C == '\r' || C == '\n'
			is_space := C == ' ' || C == '\t'
			is_sep := is_space || is_end_of_line

			switch state {
			case 0: // Beginning of sequence chunk must start with @

				if C == '@' {
					// Beginning of sequence
					state = 1
				} else {
					log.Fatalf("%s : sequence entry is not starting with @", source)
				}
			case 1: // Beginning of identifier (Mandatory)
				if is_sep {
					// No identifier -> ERROR
					log.Fatalf("%s : sequence identifier is empty", source)
				} else {
					// Beginning of identifier
					state = 2
					idBytes.Reset()
					idBytes.WriteByte(C)
				}
			case 2: // Following of the identifier
				if is_sep {
					// End of identifier
					identifier = idBytes.String()
					state = 3
				} else {
					idBytes.WriteByte(C)
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
					defBytes.Reset()
					defBytes.WriteByte(C)
					state = 4
				}
			case 4: // Following of the definition
				if is_end_of_line {
					definition = defBytes.String()
					state = 5
				} else {
					defBytes.WriteByte(C)
				}
			case 5: // Beginning of sequence
				if !is_end_of_line {
					// Beginning of sequence
					if C >= 'A' && C <= 'Z' {
						C = C + 'a' - 'A'
					}
					seqBytes.Reset()
					seqBytes.WriteByte(C)
					state = 6
				}
			case 6:
				if is_end_of_line {
					// End of sequence
					rawseq := seqBytes.Bytes()
					if len(rawseq) == 0 {
						log.Fatalf("@%s[%s] : sequence is empty", identifier, source)
					}
					s := obiseq.NewBioSequence(identifier, rawseq, definition)
					s.SetSource(source)
					sequences = append(sequences, s)
					state = 7
				} else {
					if C >= 'A' && C <= 'Z' {
						C = C + 'a' - 'A'
					}
					if (C >= 'a' && C <= 'z') || C == '-' || C == '.' || C == '[' || C == ']' {
						seqBytes.WriteByte(C)
					} else {
						context, _ := scanner.Peek(30)
						context = append(
							append([]byte{previous}, C),
							context...)
						log.Fatalf("%s [%s]: sequence contains invalid character %c (%s)",
							source, identifier, C, string(context))
					}
				}
			case 7:
				if is_end_of_line {
					state = 7
				} else if C == '+' {
					state = 8
				} else {
					log.Fatalf("@%s[%s] : sequence data not followed by a line starting with + but a %c", identifier, source, C)
				}
			case 8:
				// State consuming the + internal header line
				if is_end_of_line {
					state = 9
				}
			case 9:
				if is_end_of_line {
					state = 9
				} else {
					// beginning of quality
					state = 10
					qualBytes.Reset()
					qualBytes.WriteByte(C)
				}
			case 10:
				if is_end_of_line {
					_storeSequenceQuality(&qualBytes, sequences[len(sequences)-1], quality_shift)

					if no_order {
						if len(sequences) == batch_size {
							out.Push(obiiter.MakeBioSequenceBatch(chunck_order(), sequences))
							sequences = make(obiseq.BioSequenceSlice, 0, batch_size)
						}
					}

					state = 11
				} else {
					qualBytes.WriteByte(C)
				}
			case 11:
				if is_end_of_line {
					state = 11
				} else if C == '@' {
					state = 1
				} else {
					log.Fatalf("%s[%s] : sequence record not followed by a line starting with @", identifier, source)
				}

			}

			previous = C
		}

		if len(sequences) > 0 {
			if state == 10 {
				_storeSequenceQuality(&qualBytes, sequences[len(sequences)-1], quality_shift)
				state = 1
			}

			co := chunks.order
			if no_order {
				co = chunck_order()
			}
			out.Push(obiiter.MakeBioSequenceBatch(co, sequences))
		}

	}

	out.Done()

}

func ReadFastq(reader io.Reader, options ...WithOption) (obiiter.IBioSequence, error) {
	opt := MakeOptions(options)
	out := obiiter.MakeIBioSequence()

	nworker := opt.ParallelWorkers()
	chunkorder := obiutils.AtomicCounter()

	buff := make([]byte, 1024*1024*1024)

	chkchan := ReadSeqFileChunk(reader, buff, _EndOfLastFastqEntry)

	for i := 0; i < nworker; i++ {
		out.Add(1)
		go _ParseFastqFile(opt.Source(),
			chkchan,
			out,
			byte(obioptions.InputQualityShift()),
			opt.NoOrder(),
			opt.BatchSize(),
			chunkorder)
	}

	go func() {
		out.WaitAndClose()
	}()

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
