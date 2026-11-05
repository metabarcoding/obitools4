package obiformats

import (
	"strings"
	"testing"
)

const tagMismatchesBody = `experiment,sample,sample_tag,forward_primer,reverse_primer
exp,s1,aaaaaaa:cccccct,GGGCAATCCTGAGCCAAATC,CTTCCRTTGAGTCTCTGCACCT
exp,s2,tttttta:ccccccc,GGGCAATCCTGAGCCAAATC,CTTCCRTTGAGTCTCTGCACCT
`

// TestTagMismatchesParameters: the @tag_mismatches parameter, global, for the
// forward or the reverse primer only, or for one given primer, as the other
// parameters of the same kind; the default is no limit.
func TestTagMismatchesParameters(t *testing.T) {
	const forwardPrimer, reversePrimer = "GGGCAATCCTGAGCCAAATC", "CTTCCRTTGAGTCTCTGCACCT"
	for _, tc := range []struct {
		name          string
		params        string
		forward, back int
	}{
		{"default", "", -1, -1},
		{"global", "@param,tag_mismatches,1\n", 1, 1},
		{"forward", "@param,forward_tag_mismatches,2\n", 2, -1},
		{"reverse", "@param,reverse_tag_mismatches,0\n", -1, 0},
		{"forward primer", "@param,tag_mismatches," + forwardPrimer + ",3\n", 3, -1},
		{"reverse primer", "@param,tag_mismatches," + reversePrimer + ",1\n", -1, 1},
		{"global then reverse", "@param,tag_mismatches,2\n@param,reverse_tag_mismatches,0\n", 2, 0},
		{"negative is no limit", "@param,tag_mismatches,-1\n", -1, -1},
	} {
		library, err := ReadCSVNGSFilter(strings.NewReader(tc.params + tagMismatchesBody))
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if len(library.Markers) != 1 {
			t.Fatalf("%s: %d markers", tc.name, len(library.Markers))
		}
		for _, m := range library.Markers {
			if m.Forward_tag_mismatches != tc.forward || m.Reverse_tag_mismatches != tc.back {
				t.Errorf("%s: forward %d reverse %d, want %d and %d", tc.name, m.Forward_tag_mismatches, m.Reverse_tag_mismatches, tc.forward, tc.back)
			}
		}
	}
}

// TestPrimerParameters: the parameters setting the number of mismatches allowed for
// the primers and whether they allow indels, as the documentation names them:
// global, forward or reverse only, or for one given primer.
func TestPrimerParameters(t *testing.T) {
	const forwardPrimer, reversePrimer = "GGGCAATCCTGAGCCAAATC", "CTTCCRTTGAGTCTCTGCACCT"
	for _, tc := range []struct {
		name                   string
		params                 string
		forwardErr, reverseErr int  // the default is 2
		forwardIndel, revIndel bool // the default is false
	}{
		{"default", "", 2, 2, false, false},
		{"primer_mismatches", "@param,primer_mismatches,5\n", 5, 5, false, false},
		{"forward and reverse", "@param,forward_primer_mismatches,4\n@param,reverse_primer_mismatches,1\n", 4, 1, false, false},
		{"one primer", "@param,primer_mismatches," + reversePrimer + ",3\n", 2, 3, false, false},
		{"global then forward", "@param,primer_mismatches,5\n@param,forward_primer_mismatches,1\n", 1, 5, false, false},
		{"indels", "@param,indels,true\n", 2, 2, true, true},
		{"forward and reverse indels", "@param,forward_indels,true\n@param,reverse_indels,false\n", 2, 2, true, false},
		{"indels of one primer", "@param,indels," + forwardPrimer + ",true\n", 2, 2, true, false},
	} {
		library, err := ReadCSVNGSFilter(strings.NewReader(tc.params + tagMismatchesBody))
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		for _, m := range library.Markers {
			if m.Forward_error != tc.forwardErr || m.Reverse_error != tc.reverseErr {
				t.Errorf("%s: mismatches %d and %d, want %d and %d", tc.name, m.Forward_error, m.Reverse_error, tc.forwardErr, tc.reverseErr)
			}
			if m.Forward_allows_indels != tc.forwardIndel || m.Reverse_allows_indels != tc.revIndel {
				t.Errorf("%s: indels %v and %v, want %v and %v", tc.name, m.Forward_allows_indels, m.Reverse_allows_indels, tc.forwardIndel, tc.revIndel)
			}
		}
	}
}
