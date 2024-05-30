package obiformats

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path"
	"slices"

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

	state := 0

	idBytes := new(bytes.Buffer)
	defBytes := new(bytes.Buffer)
	qualBytes := new(bytes.Buffer)
	seqBytes := new(bytes.Buffer)

	for chunks := range input {
		scanner := bufio.NewReader(chunks.raw)
		sequences := make(obiseq.BioSequenceSlice, 0, 100)
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
					log.Errorf("%s : sequence entry is not starting with @", source)
				}
			case 1: // Beginning of identifier (Mandatory)
				if is_sep {
					// No identifier -> ERROR
					log.Errorf("%s : sequence identifier is empty", source)
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
					s := obiseq.NewBioSequence(identifier, slices.Clone(seqBytes.Bytes()), definition)
					s.SetSource(source)
					sequences = append(sequences, s)
					state = 7
				} else {
					if C >= 'A' && C <= 'Z' {
						C = C + 'a' - 'A'
					}
					seqBytes.WriteByte(C)
				}
			case 7:
				if is_end_of_line {
					state = 7
				} else if C == '+' {
					state = 8
				} else {
					log.Errorf("@%s[%s] : sequence data not followed by a line starting with + but a %c", identifier, source, C)
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
					qualBytes.Reset()
					qualBytes.WriteByte(C)
				}
			case 10:
				if is_end_of_line {
					// End of quality
					q := qualBytes.Bytes()
					if len(q) != sequences[len(sequences)-1].Len() {
						log.Errorf("%s[%s] : sequence data and quality lenght not equal (%d/%d)",
							identifier, source, len(q), sequences[len(sequences)-1].Len())
					}
					for i := 0; i < len(q); i++ {
						q[i] = q[i] - quality_shift
					}
					sequences[len(sequences)-1].SetQualities(q)

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
					log.Errorf("%s[%s] : sequence record not followed by a line starting with @", identifier, source)
				}

			}
		}

		if len(sequences) > 0 {
			if no_order {
				out.Push(obiiter.MakeBioSequenceBatch(chunck_order(), sequences))
			} else {
				out.Push(obiiter.MakeBioSequenceBatch(chunks.order, sequences))
			}
		}

	}

	out.Done()

}

func ReadFastq(reader io.Reader, options ...WithOption) (obiiter.IBioSequence, error) {
	opt := MakeOptions(options)
	out := obiiter.MakeIBioSequence()

	nworker := opt.ParallelWorkers()
	chunkorder := obiutils.AtomicCounter()

	chkchan := ReadSeqFileChunk(reader, _EndOfLastFastqEntry)

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
