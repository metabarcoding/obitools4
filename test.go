package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"regexp"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"github.com/gabriel-vasile/mimetype"
	"golang.org/x/exp/slices"
)

func OBIMimeTypeGuesser(stream io.Reader) (*mimetype.MIME, io.Reader, error) {
	fastaDetector := func(raw []byte, limit uint32) bool {
		ok, err := regexp.Match("^>[^ ]", raw)
		return ok && err == nil
	}

	fastqDetector := func(raw []byte, limit uint32) bool {
		ok, err := regexp.Match("^@[^ ]", raw)
		return ok && err == nil
	}

	ecoPCR2Detector := func(raw []byte, limit uint32) bool {
		ok := bytes.HasPrefix(raw, []byte("#@ecopcr-v2"))
		return ok
	}

	genbankDetector := func(raw []byte, limit uint32) bool {
		ok2 := bytes.HasPrefix(raw, []byte("LOCUS       "))
		ok1, err := regexp.Match("^[^ ]* +Genetic Sequence Data Bank *\n", raw)
		return ok2 || (ok1 && err == nil)
	}

	emblDetector := func(raw []byte, limit uint32) bool {
		ok := bytes.HasPrefix(raw, []byte("ID   "))
		return ok
	}

	mimetype.Lookup("text/plain").Extend(fastaDetector, "text/fasta", ".fasta")
	mimetype.Lookup("text/plain").Extend(fastqDetector, "text/fastq", ".fastq")
	mimetype.Lookup("text/plain").Extend(ecoPCR2Detector, "text/ecopcr2", ".ecopcr")
	mimetype.Lookup("text/plain").Extend(genbankDetector, "text/genbank", ".seq")
	mimetype.Lookup("text/plain").Extend(emblDetector, "text/embl", ".dat")

	// Create a buffer to store the read data
	buf := make([]byte, 1024*128)
	n, err := stream.Read(buf)

	if err != nil && err != io.EOF {
		return nil, nil, err
	}

	// Detect the MIME type using the mimetype library
	mimeType := mimetype.Detect(buf)
	if mimeType == nil {
		return nil, nil, err
	}

	// Create a new reader based on the read data
	newReader := io.MultiReader(bytes.NewReader(buf[:n]), stream)

	return mimeType, newReader, nil
}

var xxx1 = `00422_612GNAAXX:7:73:6614:3284#0/1  
ccgaatatcttagataccccactatgcttagccctaaacataaacattcaataaacaaga
atgttcgccagagtactactagcaacagcctgaaactcaaagcacttg
>HELIUM_000100422_612GNAAXX:7:13:11063:8138#0/1  
ccgcctcctttagataccccactatgcttagccctaaacacaagtaattattataacaaa
attattcgccagagtactaccggcaatagcttaaaactcacagaactt
>HELIUM_000100422_612GNAAXX:7:2:7990:17026#0/1  
ccgaatatctagaacaggctcctctagagggatgtaaagcaccgccaagtcctttgcttt
tcaggctgttgctagtagtactctggcgaccattcttgtttattgatt
>HELIUM_000100422_612GNAAXX:7:3:19649:11224#0/1  
ccgaatatctagaacaggctcctctagagggatgtaaagcaccgccaagtcctttgagtt
tcaggctgttgctagtagtactctggcgaacattcttgtttattgaat
>HELIUM_000100422_612GNAAXX:7:3:8446:7884#0/1  
ccgaatatctagaacaggctcctctagagggatgtaaagcaccgccaagtcctttgagtt
tcaggctgttgctagtagtactctggcgaacattcttgtttattgaat
>HELIUM_000100422_612GNAAXX:7:108:8714:2464#0/1  
ccgcctcctttagataccccactatgcttagccctaaacacaagtaattaatataacaaa
attattcgccagagtactaccggcaatagcttaaaactcaaaggactt
>HELIUM_000100422_612GNAAXX:7:28:3969:15209#0/1  
ccaattaacttagataccccactatgcctagccttaaacacaaatagttatgcaaacaaa
actattcgccagagtactaccggcaatagcttaaaactcaacgcactg
>HELIUM_000100422_612GNAAXX:7:44:3269:3608#0/1  
gaagtagtagaacaggctcctctagaagggt`

var xxx2 = `>HELIUM_000100422_612GNAAXX:7:13:11063:8138#0/1  
ccgcctcctttagataccccactatgcttagccctaaacacaagtaattattataacaaa
attattcgccagagtactaccggcaatagcttaaaactcacagaactt
>HELIUM_000100422_612GNAAXX:7:2:7990:17026#0/1  
ccgaatatctagaacaggctcctctagagggatgtaaagcaccgccaagtcctttgcttt
tcaggctgttgctagtagtactctggcgaccattcttgtttattgatt
>HELIUM_000100422_612GNAAXX:7:3:19649:11224#0/1  
ccgaatatctagaacaggctcctctagagggatgtaaagcaccgccaagtcctttgagtt
tcaggctgttgctagtagtactctggcgaacattcttgtttattgaat
>HELIUM_000100422_612GNAAXX:7:3:8446:7884#0/1  
ccgaatatctagaacaggctcctctagagggatgtaaagcaccgccaagtcctttgagtt
tcaggctgttgctagtagtactctggcgaacattcttgtttattgaat
>HELIUM_000100422_612GNAAXX:7:108:8714:2464#0/1  
ccgcctcctttagataccccactatgcttagccctaaacacaagtaattaatataacaaa
attattcgccagagtactaccggcaatagcttaaaactcaaaggactt
>HELIUM_000100422_612GNAAXX:7:28:3969:15209#0/1  
ccaattaacttagataccccactatgcctagccttaaacacaaatagttatgcaaacaaa
actattcgccagagtactaccggcaatagcttaaaactcaacgcactg
>HELIUM_000100422_612GNAAXX:7:44:3269:3608#0/1  
gaagtagtagaacaggctcctctagaagggt`

var xxx3 = `00422_612GNAAXX:7:73:6614:3284#0/1  
ccgaatatcttagataccccactatgcttagccctaaacataaacattcaataaacaaga
atgttcgccagagtactactagcaacagcctgaaactcaaagcacttg
>HELIUM_000100422_612GNAAXX:7:13:11063:8138#0/1  
ccgcctcctttagataccccactatgcttagccctaaacacaagtaattattataacaaa
attattcgccagagtactaccggcaatagcttaaaactcacagaactt
>HELIUM_000100422_612GNAAXX:7:2:7990:17026#0/1  
ccgaatatctagaacaggctcctctagagggatgtaaagcaccgccaagtcctttgcttt
tcaggctgttgctagtagtactctggcgaccattcttgtttattgatt
>HELIUM_000100422_612GNAAXX:7:3:19649:11224#0/1  
ccgaatatctagaacaggctcctctagagggatgtaaagcaccgccaagtcctttgagtt
tcaggctgttgctagtagtactctggcgaacattcttgtttattgaat
>HELIUM_000100422_612GNAAXX:7:3:8446:7884#0/1  
ccgaatatctagaacaggctcctctagagggatgtaaagcaccgccaagtcctttgagtt
tcaggctgttgctagtagtactctggcgaacattcttgtttattgaat
>HELIUM_000100422_612GNAAXX:7:108:8714:2464#0/1  
ccgcctcctttagataccccactatgcttagccctaaacacaagtaattaatataacaaa
attattcgccagagtactaccggcaatagcttaaaactcaaaggactt
>HELIUM_000100422_612GNAAXX:7:28:3969:15209#0/1  
ccaattaacttagataccccactatgcctagccttaaacacaaatagttatgcaaacaaa
actattcgccagagtactaccggcaatagcttaaaactcaacgcactg`

var yyy1 = `@HELIUM_000100422_612GNAAXX:7:1:9007:3289#0/1 {"demultiplex_error":"cannot assign the sequence to a sample"} 
ccatctctcttagataccccactatgcttagccctaaacacaagtaattaatataacaaaattattcgccagagtactaccggcaatagcttaaaactcaaagaactc
+
CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCBCCACCCCCCCCCBCCACC779?############################################
@HELIUM_000100422_612GNAAXX:7:1:8849:9880#0/1 {"demultiplex_error":"cannot match any primer pair"} 
gatcggaagagcggttcagcaggaatgccgagaccgatatcgtatgccgtcttctgcttgaaaaaaaaaacaaaataggagagtagactcactgccagtggtcgtcag
`

func LastFastqCut(buffer []byte) ([]byte, []byte) {
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

func LastSequenceCut(buffer []byte) ([]byte, []byte) {
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

func FirstSequenceCut(buffer []byte) ([]byte, []byte) {
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

func FullSequenceCut(buffer []byte) ([]byte, []byte, []byte) {
	before, buffer := FirstSequenceCut(buffer)

	if len(buffer) == 0 {
		return before, []byte{}, []byte{}
	}

	buffer, after := LastSequenceCut(buffer)
	return before, buffer, after
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

		begin, buff := FirstSequenceCut(buff)

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
				buff, end = LastSequenceCut(buff)
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
					fmt.Println("***** Empty buff *****")
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

func ParseFastaChunk(ch FastxChunk) *obiiter.BioSequenceBatch {
	fmt.Println(string(ch.Bytes))
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
				current = i
				state = 6
			}
		case 6:
			if C == '>' {
				// End of sequence
				s := obiseq.NewBioSequence(identifier, bytes.Clone(ch.Bytes[start:current]), definition)
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

	fmt.Printf("Index = %d, State = %d\n", ch.index, state)
	slice = append(slice, obiseq.NewBioSequence(identifier, bytes.Clone(ch.Bytes[start:current]), definition))
	batch := obiiter.MakeBioSequenceBatch(ch.index, slice)
	return &batch
}

func ReadFastaSequence(reader io.Reader) obiiter.IBioSequence {
	out := obiiter.MakeIBioSequence()

	nworker := obioptions.CLIReadParallelWorkers()
	out.Add(nworker)

	chkchan, err := FastaChunkReader(reader, 1024*500, false)

	if err != nil {
		log.Panicln("Error:", err)
	}

	go func() {
		out.WaitAndClose()
	}()

	parser := func() {
		defer out.Done()
		for chk := range chkchan {
			seqs := ParseFastaChunk(chk)
			if seqs != nil {
				out.Push(*seqs)
			}
		}
	}

	for i := 0; i < nworker; i++ {
		go parser()
	}

	return out.SortBatches().Rebatch(obioptions.CLIBatchSize())
}

func main() {
	// if len(os.Args) != 2 {
	// 	fmt.Println("Usage: go run main.go <filename>")
	// 	return
	// }

	// filename := os.Args[1]
	// filename := "100.fasta"
	// file, err := os.Open(filename)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }
	// defer file.Close()

	// mimeType, input, err := OBIMimeTypeGuesser(file)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }

	// fmt.Println("Detected MIME Type:", mimeType.String())

	// ch, err := FastaChunkReader(input, 1024, false)

	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }

	// for chk := range ch {
	// 	fmt.Printf("--------------------\n")
	// 	b := ParseFastaChunk(chk)
	// 	fmt.Printf("-------- %d --------\n", b.Order())
	// 	for _, b := range b.Slice() {
	// 		fmt.Printf("--%s--\t--%s--\t--%s--\n", b.Id(), b.Definition(), b.String())
	// 	}
	// }

	d1, f1 := LastFastqCut([]byte(yyy1))
	// d2, f2 := LastSequenceCut([]byte(xxx2))
	// d3, f3 := LastSequenceCut([]byte(xxx3))

	fmt.Println("Last Sequence Cut 1:", string(d1), "---", string(f1))
	// fmt.Println("Last Sequence Cut 2:", string(d2), "---", string(f2))
	// fmt.Println("Last Sequence Cut 3:", string(d3), "---", string(f3))

	// d1, b1, f1 := FullSequenceCut([]byte(xxx1))
	// d2, b2, f2 := FullSequenceCut([]byte(xxx2))
	// d3, b3, f3 := FullSequenceCut([]byte(xxx3))

	// fmt.Println("Last Sequence Cut:", string(d1), "---", string(b1), "---", string(f1))
	// fmt.Println("Last Sequence Cut:", string(d2), "---", string(b2), "---", string(f2))
	// fmt.Println("Last Sequence Cut:", string(d3), "---", string(b3), "---", string(f3))

	// Now you can use "extractedData" to access the read data with the associated MIME type.
	// For example, you can copy the data into a buffer for further manipulation.
}
