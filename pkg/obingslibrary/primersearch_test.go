package obingslibrary_test

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiformats"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obingslibrary"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

func sortedMatches(m []obingslibrary.PrimerMatch) []obingslibrary.PrimerMatch {
	m = slices.Clone(m)
	slices.SortFunc(m, func(a, b obingslibrary.PrimerMatch) int {
		for _, d := range []int{a.Begin - b.Begin, a.End - b.End, a.Marker - b.Marker} {
			if d != 0 {
				return d
			}
		}
		if a.Forward != b.Forward {
			if a.Forward {
				return -1
			}
			return 1
		}
		return 0
	})
	return m
}

// TestPrimerSearchAgainstSeparate compares the matches of the joint search with
// those of the separate ones on real files (OBIDNAPAT_FASTQ, OBIDNAPAT_NGSFILTER,
// OBIDNAPAT_READS), and reports how many reads differ.
func TestPrimerSearchAgainstSeparate(t *testing.T) {
	path, ngs := os.Getenv("OBIDNAPAT_FASTQ"), os.Getenv("OBIDNAPAT_NGSFILTER")
	if path == "" || ngs == "" {
		t.Skip("OBIDNAPAT_FASTQ and OBIDNAPAT_NGSFILTER not set")
	}
	n := 20000
	if v, err := strconv.Atoi(os.Getenv("OBIDNAPAT_READS")); err == nil && v > 0 {
		n = v
	}
	f, err := os.Open(ngs)
	if err != nil {
		t.Fatal(err)
	}
	library, err := obiformats.ReadNGSFilter(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if err := library.Compile2(); err != nil {
		t.Fatal(err)
	}
	r, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<26)
	reads, differ, shown, positive := 0, 0, 0, 0
	for line := 0; sc.Scan() && reads < n; line++ {
		if line%4 != 1 {
			continue
		}
		reads++
		s := obiseq.NewBioSequence(fmt.Sprint(reads), append([]byte(nil), sc.Bytes()...), "")
		want := sortedMatches(library.OldPrimerMatches(s))
		got := sortedMatches(library.PrimerMatchesOf(s))
		if !slices.Equal(want, got) {
			differ++
			// Expected: only matches of the primer of the second end differ (a
			// window of the truncated region that a better one, starting before the
			// region, absorbs in the search of the whole read).
			for _, m := range slices.Concat(diffMatches(want, got), diffMatches(got, want)) {
				if m.Marker > 0 {
					positive++
					t.Errorf("read %d: a match of a first end differs: %v\n separate %v\n joint    %v", reads, m, want, got)
				}
			}
			if shown < 3 {
				shown++
				t.Logf("read %d (%d bases)\n separate %v\n joint    %v", reads, s.Len(), want, got)
			}
		}
	}
	t.Logf("%d reads, %d with different matches", reads, differ)
	_ = positive
}

// diffMatches returns the matches of a that b does not have.
func diffMatches(a, b []obingslibrary.PrimerMatch) []obingslibrary.PrimerMatch {
	var d []obingslibrary.PrimerMatch
	for _, m := range a {
		if !slices.Contains(b, m) {
			d = append(d, m)
		}
	}
	return d
}
