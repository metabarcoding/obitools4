package obiprimer

import (
	"fmt"
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// span writes the 1-based GenBank range of the lp bases starting at the 0-based
// position pos. On a circular sequence a range running over the junction is
// split in two.
func span(pos, lp, n int, circular bool) string {
	end := pos + lp
	if end > n && circular {
		return fmt.Sprintf("%d..%d,1..%d", pos+1, n, end-n)
	}
	return fmt.Sprintf("%d..%d", pos+1, end)
}

// referenceSites describes an amplicon of a pair on the reference sequence raw.
//
// The location lists the two primer sites left to right as join(a..b,c..d),
// inside complement() when the first primer of the pair (Primer1 once the pair
// is in string order) binds the reverse strand. The amplicon is the region
// between the sites, read 5'->3' from the first primer, in lower case.
//
// swappedByOrder says whether putting the pair in string order exchanged the
// two primers of the key.
func referenceSites(raw []byte, h refHit, p PairParams, swappedByOrder bool) (location, amplicon string) {
	lp, n := p.PrimerLength, len(raw)

	// The key slot 1 is the upstream primer unless the match order was swapped;
	// the exchange due to the string order flips it again.
	firstIsUpstream := h.swapped == swappedByOrder

	location = "join(" + span(int(h.up.pos), lp, n, p.Circular) + "," + span(int(h.down.pos), lp, n, p.Circular) + ")"
	region := []byte(strings.ToLower(string(raw[int(h.up.pos)+lp : int(h.down.pos)])))
	if firstIsUpstream {
		return location, string(region)
	}
	rc := obiseq.NewBioSequence("", region, "").ReverseComplement(true)
	return "complement(" + location + ")", string(rc.Sequence())
}
