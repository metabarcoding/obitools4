package main

import (
	"fmt"
	"log"
	"os"
	"runtime/trace"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obialign"
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
	B := []byte("cgccaccaccgagatctacactctttccctacacgacgctcttccgatctccgcctccttagaacaggctcctctagaaaagcatagtggggtatctaaaggaggcgg")
	sA := obiseq.MakeBioSequence("A", A, "")
	sB := obiseq.MakeBioSequence("B", B, "")

	fmt.Println(string(sA.Sequence()))
	fmt.Println(sA.Qualities())
	fmt.Println(string(sB.Sequence()))
	fmt.Println(sB.Qualities())

	score, path := obialign.PELeftAlign(sA, sB, 2, obialign.NilPEAlignArena)
	fmt.Printf("Score : %d Path : %v\n", score, path)
	score, path = obialign.PERightAlign(sA, sB, 2, obialign.NilPEAlignArena)
	fmt.Printf("Score : %d Path : %v\n", score, path)

	fmt.Println(string(sA.Sequence()))
	sA.ReverseComplement(true)
	fmt.Println(string(sA.Sequence()))
	fmt.Println(string(sA.Id()))
}
