package main_test

import (
	"fmt"
	"testing"

	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obiannot"
	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obiformats"
	"git.metabarcoding.org/lecasofts/go/oa2/pkg/obiseq"
)

func TestParseOBIFasta(t *testing.T) {
	f := "/Users/coissac/travail/Adeline/Soumission_data/Zonation/euka03/euka03.ecotag.fasta.gz"

	var nseq, nread int
	nseq = 0
	nread = 0

	fs := obiformats.ReaderFromIlluminaFile(f)

	fmt.Println(f)

	for i := range obiannot.ExtractHeaderChannel(fs, fastseq.ParseOBIHeader) {
		for _, s := range i {
			nseq++
			nread += s.Count()
		}
	}
	fmt.Println(nseq, nread)

}

func ExtractHeaderChannel(fs fastseq.IFastSeq, sequence func(sequence obiseq.Sequence)) {
	panic("unimplemented")
}

// Performance test of an ADEXP message parsing
func BenchmarkParseOBIFasta(t *testing.B) {

	f := "/Users/coissac/travail/Adeline/Soumission_data/Zonation/euka03/euka03.ecotag.fasta.gz"

	var nseq, nread int
	nseq = 0
	nread = 0

	fs := fastseq.ReaderFromIlluminaFile(f)

	fmt.Println(f)

	for i := range obiannot.ExtractHeaderChannel(fs, fastseq.ParseOBIHeader) {
		for _, s := range i {
			nseq++
			nread += s.Count()
		}
	}
	fmt.Println(nseq, nread)

}
