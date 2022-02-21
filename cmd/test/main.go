package main

import (
	"fmt"
	"log"
	"os"
	"runtime/trace"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiapat"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiformats"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func main() {

	ftrace, err := os.Create("cpu.trace")
	if err != nil {
		log.Fatal(err)
	}
	trace.Start(ftrace)
	defer trace.Stop()

	// option_parser := obioptions.GenerateOptionParser(
	// 	obiconvert.InputOptionSet,
	// )

	//_, args, _ := option_parser(os.Args)

	// fs, _ := obiconvert.ReadBioSequences(args...)
	// buffer := make([]byte, 0)
	// fs.Next()
	// s := fs.Get()
	// index := obikmer.Index4mer(s, nil, nil)
	// for fs.Next() {
	// 	s := fs.Get()
	// 	if s.IsNil() {
	// 		log.Panicln("Read sequence is nil")
	// 	}
	// 	maxshift, maxcount := obikmer.FastShiftFourMer(index, s, buffer)

	// 	fmt.Printf("Shift : %d   Score : %d\n", maxshift, maxcount)
	// }

	A := []byte("ccgcctccttagaacaggctcctctagaaaaccatagtgggatatctaaagaaggcggagatagaaagagcggttcagcaggaatgccgagatggacggcgtgtgacg")
	// B := []byte("cgccaccaccgagatctacactctttccctacacgacgctcttccgatctccgcctccttagaacaggctcctctagaaaagcatagtggggtatctaaaggaggcgg")
	sA := obiseq.NewBioSequence("A", A, "")
	// sB := obiseq.MakeBioSequence("B", B, "")

	pat, _ := obiapat.MakeApatPattern("TCCTTCCAACAGGCTCCTC", 3)
	as, _ := obiapat.MakeApatSequence(sA, false)
	fmt.Println(pat.FindAllIndex(as))

	file, _ := os.Open("sample/wolf_diet_ngsfilter.txt")
	xxx, _ := obiformats.ReadNGSFilter(file)
	xxx.Compile(2)
	fmt.Printf("%v\n==================\n", xxx)

	for pp, m := range xxx {
		fmt.Printf("%v %v\n", pp, *m)
	}

	seqfile, _ := obiformats.ReadFastSeqFromFile("xxxx.fastq")

	for seqfile.Next() {
		seq := seqfile.Get()
		barcode, _ := xxx.ExtractBarcode(seq, true)
		fmt.Println(obiformats.FormatFasta(barcode, obiformats.FormatFastSeqOBIHeader))
	}
}
