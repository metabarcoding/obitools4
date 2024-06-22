package obiformats

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

	log "github.com/sirupsen/logrus"
)

func _EndOfLastFastaEntry(buffer []byte) int {
	var i int

	imax := len(buffer)
	last := 0
	state := 0

	for i = imax - 1; i >= 0 && state < 2; i-- {
		C := buffer[i]
		if C == '>' && state == 0 {
			state = 1
			last = i
		} else if state == 1 && (C == '\n' || C == '\r') {
			state = 2
		} else {
			state = 0
		}
	}

	if i == 0 || state != 2 {
		return -1
	}
	return last
}

func _ParseFastaFile(source string,
	input ChannelSeqFileChunk,
	out obiiter.IBioSequence,
	no_order bool,
	batch_size int,
	chunck_order func() int,
) {

	var identifier string
	var definition string

	idBytes := bytes.Buffer{}
	defBytes := bytes.Buffer{}
	seqBytes := bytes.Buffer{}

	for chunks := range input {
		state := 0
		scanner := bufio.NewReader(chunks.raw)
		start, _ := scanner.Peek(20)
		if start[0] != '>' {
			log.Fatalf("%s : first character is not '>'", string(start))
		}
		if start[1] == ' ' {
			log.Fatalf("%s :Strange", string(start))
		}
		sequences := make(obiseq.BioSequenceSlice, 0, batch_size)

		previous := byte(0)

		for C, err := scanner.ReadByte(); err != io.EOF; C, err = scanner.ReadByte() {

			is_end_of_line := C == '\r' || C == '\n'
			is_space := C == ' ' || C == '\t'
			is_sep := is_space || is_end_of_line

			switch state {
			case 0:
				if C == '>' {
					// Beginning of sequence
					state = 1
				} else {
					// ERROR
					log.Fatalf("%s : sequence entry does not start with '>'", source)
				}
			case 1:
				if is_sep {
					// No identifier -> ERROR

					context, _ := scanner.Peek(30)
					context = append([]byte{C}, context...)
					log.Fatalf("%s [%s]: sequence entry does not have an identifier",
						source, string(context))
				} else {
					// Beginning of identifier
					idBytes.Reset()
					state = 2
					idBytes.WriteByte(C)
				}
			case 2:
				if is_sep {
					// End of identifier
					identifier = idBytes.String()
					idBytes.Reset()
					state = 3
				} else {
					idBytes.WriteByte(C)
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
					defBytes.Reset()
					defBytes.WriteByte(C)
					state = 4
				}
			case 4:
				if is_end_of_line {
					definition = defBytes.String()
					state = 5
				} else {
					defBytes.WriteByte(C)
				}
			case 5:
				if !is_end_of_line {
					// Beginning of sequence
					seqBytes.Reset()
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
					state = 6
				}
			case 6:
				if C == '>' {
					if previous == '\r' || previous == '\n' {
						// End of sequence
						rawseq := seqBytes.Bytes()
						if len(rawseq) == 0 {
							log.Fatalf("@%s[%s] : sequence is empty", identifier, source)
						}
						s := obiseq.NewBioSequence(identifier, rawseq, definition)
						s.SetSource(source)
						sequences = append(sequences, s)
						if no_order {
							if len(sequences) == batch_size {
								out.Push(obiiter.MakeBioSequenceBatch(chunck_order(), sequences))
								sequences = make(obiseq.BioSequenceSlice, 0, batch_size)
							}
						}
						state = 1
					} else {
						// Error
						context, _ := scanner.Peek(30)
						context = append(
							append([]byte{previous}, C),
							context...)
						log.Fatalf("%s [%s]: sequence cannot contain '>' in the middle (%s)",
							source, identifier, string(context))
					}

				} else if !is_sep {
					if C >= 'A' && C <= 'Z' {
						C = C + 'a' - 'A'
					}
					// Removing white space from the sequence
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

			}

			previous = C
		}

		if state == 6 {
			rawseq := seqBytes.Bytes()
			if len(rawseq) == 0 {
				log.Fatalf("@%s[%s] : sequence is empty", identifier, source)
			}
			s := obiseq.NewBioSequence(identifier, rawseq, definition)
			s.SetSource(source)
			sequences = append(sequences, s)
		}

		if len(sequences) > 0 {
			co := chunks.order
			if no_order {
				co = chunck_order()
			}
			out.Push(obiiter.MakeBioSequenceBatch(co, sequences))
		}
	}

	out.Done()

}

func ReadFasta(reader io.Reader, options ...WithOption) (obiiter.IBioSequence, error) {
	opt := MakeOptions(options)
	out := obiiter.MakeIBioSequence()

	nworker := opt.ParallelWorkers()

	buff := make([]byte, 1024*1024*1024)

	chkchan := ReadSeqFileChunk(reader, buff, _EndOfLastFastaEntry)
	chunck_order := obiutils.AtomicCounter()

	for i := 0; i < nworker; i++ {
		out.Add(1)
		go _ParseFastaFile(opt.Source(),
			chkchan,
			out,
			opt.NoOrder(),
			opt.BatchSize(),
			chunck_order)
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
