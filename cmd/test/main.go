package main

import (
	"fmt"
	"log"
	"os"
	"runtime/trace"

	"cloudeng.io/algo/lcs"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obialign"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func SESStat(script *lcs.EditScript[byte]) (int, int) {
	llcs := 0
	gaps := 0
	extra := 0
	i := 0
	ls := len(*script)
	for i < ls {
		e := (*script)[i]
		// log.Println(i,e,e.Op)
		switch e.Op {
		case lcs.Identical: // 2
			if gaps > 0 {
				extra += gaps
			}
			llcs++
			gaps = 0
			i++
		case lcs.Delete: // 1
			if i+1 < ls {
				en := (*script)[i+1]
				if en.Op == lcs.Identical && e.Val == en.Val {
					log.Println("Swap delete")
					(*script)[i], (*script)[i+1] = (*script)[i+1], (*script)[i]
					continue
				}
			}
			gaps--
			i++
		case lcs.Insert: // 0
			if i+1 < ls {
				en := (*script)[i+1]
				if en.Op == lcs.Identical && e.Val == en.Val {
					log.Println("Swap Insert")
					(*script)[i], (*script)[i+1] = (*script)[i+1], (*script)[i]
					continue
				}
			}
			gaps++
			i++
		}
	}

	if gaps > 0 {
		extra += gaps
	}

	return llcs, extra
}

func main() {

	// Creating a file called cpu.trace.
	ftrace, err := os.Create("cpu.trace")
	if err != nil {
		log.Fatal(err)
	}
	trace.Start(ftrace)
	defer trace.Stop()

	// "---CACGATCGTGC-CAGTCAT-GGCTAT"
	// "CCCCA-GATCGTGCG-AGTCATGGGCTAT"

	//  00 0 00000000 1111111 111222
	//  01 2 34567889 0123456 789012
	// "CA-C-GATCGTGC-CAGTCAT-GGCTAT"
	// "CCCCAGATCGTGCG-AGTCATGGGCTAT"

	//A := "CACGATCGTGCCCCCAGTCATGGCTAT"
	A := "AAATGCCCCAGATCGTGC"
	B := "TGCCCCAGAT"

	//A = "aaaggaacttggactgaagatttccacagaggttgcgaatgaaaaacacgtattcgaatgcctcaaatacggaatcgatcttgtctg"
	A = "aaaggaacttggactgaagatttccacagaggttgcgaatgaaaaacacgtattcgaatgcctcaaatacggaatcgatcttgtctg"
	B = "atccggttttacgaaaatgcgtgtggtaggggcttatgaaaacgataatcgaataaaaaagggtaggtgcagagactcaacggaagatgttctaacaaatgg"
	// A = "aataa"
	// B = "ttttt"
	sA := obiseq.NewBioSequence("A", []byte(A), "")
	sB := obiseq.NewBioSequence("A", []byte(B), "")
	// M := lcs.NewMyers([]byte(A), []byte(B))
	// fmt.Println(M.LCS())
	// fmt.Println(M.SES())
	// fmt.Println(len(M.LCS()))
	// M.SES().FormatHorizontal(os.Stdout, []byte(B))
	// llcs, extra := SESStat(M.SES())
	// nlcs, nali := obialign.LCSScore(sA, sB, sB.Length(), nil)
	// fmt.Println(llcs, extra, len(A)+extra)
	// fmt.Println(nlcs, nali)
	nlcs, nali := obialign.FastLCSScore(sA, sB, sB.Length(), nil)
	fmt.Println(nlcs, nali)

	// option_parser := obioptions.GenerateOptionParser(
	// 	obiconvert.InputOptionSet,
	// )

	// _, args, _ := option_parser(os.Args)

	// fs, _ := obiconvert.ReadBioSequencesBatch(args...)

	// obiclean.IOBIClean(fs)

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

	// A := []byte("ccgcctccttagaacaggctcctctagaaaaccatgtgggatatctaaagaaggcggagatagaaagagcggttcagcaggaatgccgagatggacggcgtgtgacg")
	// B := []byte("ccgcctccttagaacaggctcctctagaaaaaccatgtgggatatctaaagaaggcggagatagaaagagcggttcagcaggaatgccgagatggacggcgtgtgacg")
	// B := []byte("cgccaccaccgagatctacactctttccctacacgacgctcttccgatctccgcctccttagaacaggctcctctagaaaagcatagtggggtatctaaaggaggcgg")
	// sA := obiseq.NewBioSequence("A", A, "")
	// sB := obiseq.MakeBioSequence("B", B, "")

	// s, l := obialign.LCSScore(sA, &sB, 2, nil)

	// fmt.Printf("score : %d  length : %d  error : %d\n", s, l, l-s)

	// s, l = obialign.LCSScore(&sB, &sB, 2, nil)

	// fmt.Printf("score : %d  length : %d  error : %d\n", s, l, l-s)

	// pat, _ := obiapat.MakeApatPattern("TCCTTCCAACAGGCTCCTC", 3)
	// as, _ := obiapat.MakeApatSequence(sA, false)
	// fmt.Println(pat.FindAllIndex(as))

	// file, _ := os.Open("sample/wolf_diet_ngsfilter.txt")
	// xxx, _ := obiformats.ReadNGSFilter(file)
	// xxx.Compile(2)
	// fmt.Printf("%v\n==================\n", xxx)

	// for pp, m := range xxx {
	// 	fmt.Printf("%v %v\n", pp, *m)
	// }

	// seqfile, _ := obiformats.ReadFastSeqFromFile("xxxx.fastq")

	// for seqfile.Next() {
	// 	seq := seqfile.Get()
	// 	barcode, _ := xxx.ExtractBarcode(seq, true)
	// 	fmt.Println(obiformats.FormatFasta(barcode, obiformats.FormatFastSeqOBIHeader))
	// }
}
