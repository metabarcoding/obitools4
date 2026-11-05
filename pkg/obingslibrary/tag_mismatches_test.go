package obingslibrary

import (
	"strings"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func tagLibrary(t *testing.T, matching string, forwardMismatches, reverseMismatches int) (*NGSLibrary, PrimerPair) {
	library := MakeNGSLibrary()
	const fp, rp = "ggg", "ccc"
	marker, _ := library.GetMarker(fp, rp)
	for _, tags := range [][2]string{{"aaaaaaa", "cccccct"}, {"tttttta", "ccccccc"}} {
		pcr, _ := marker.GetPCR(tags[0], tags[1])
		pcr.Sample = strings.ToUpper(tags[0])
	}
	library.CheckPrimerUnicity()
	library.CheckTagLength()
	if err := library.SetMatching(matching); err != nil {
		t.Fatal(err)
	}
	marker.SetForwardTagMismatches(forwardMismatches)
	marker.SetReverseTagMismatches(reverseMismatches)
	return &library, PrimerPair{fp, rp}
}

// TestSampleIdentifierTagMismatches: with the hamming matching the closest tag of
// the library is used whatever its distance, unless tag_mismatches says the tag is
// not valid beyond a number of errors; the strict matching is not affected.
func TestSampleIdentifierTagMismatches(t *testing.T) {
	for _, tc := range []struct {
		name              string
		matching          string
		forwardMismatches int
		reverseMismatches int
		forward, reverse  string // the tags found in the sequence
		sample            string // "" : no sample
		errorContains     string
	}{
		{"exact tags", "hamming", 1, 1, "aaaaaaa", "cccccct", "AAAAAAA", ""},
		{"one error, one allowed", "hamming", 1, 1, "aaaaaat", "cccccct", "AAAAAAA", ""},
		{"two errors, no limit", "hamming", -1, -1, "aaaaatt", "cccccct", "AAAAAAA", ""},
		{"two errors, one allowed", "hamming", 1, 1, "aaaaatt", "cccccct", "", "forward tag aaaaatt is at distance 2"},
		{"reverse too far only", "hamming", 1, 1, "aaaaaaa", "ccccgtt", "", "reverse tag ccccgtt is at distance 2"},
		{"limit only on the forward tag", "hamming", 0, -1, "aaaaaat", "ccccccg", "", "forward tag aaaaaat is at distance 1"},
		{"limit zero keeps exact tags", "hamming", 0, 0, "tttttta", "ccccccc", "TTTTTTA", ""},
		{"indel matching, one error allowed", "indel", 1, 1, "aaaaaa", "cccccct", "AAAAAAA", ""},
		{"strict is not affected", "strict", 0, 0, "aaaaaaa", "cccccct", "AAAAAAA", ""},
	} {
		library, pair := tagLibrary(t, tc.matching, tc.forwardMismatches, tc.reverseMismatches)
		annotations := make(obiseq.Annotation)
		pcr := library.SampleIdentifier(pair, &TagPair{tc.forward, tc.reverse}, annotations)
		got := ""
		if pcr != nil {
			got = pcr.Sample
		}
		if got != tc.sample {
			t.Errorf("%s: sample %q, want %q (%v)", tc.name, got, tc.sample, annotations)
		}
		if tc.errorContains != "" {
			msg, _ := annotations["obimultiplex_error"].(string)
			if !strings.Contains(msg, tc.errorContains) || !strings.HasPrefix(msg, "Invalid tag") {
				t.Errorf("%s: error %q should say that the tag is invalid (%q)", tc.name, msg, tc.errorContains)
			}
		}
	}
}
