package obidnapat

import (
	"bufio"
	"compress/gzip"
	"os"
	"strconv"
	"strings"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
)

// loadRealReads reads the sequences of the FASTQ file OBIDNAPAT_FASTQ (gzip or
// not; OBIDNAPAT_READS of them, 20000 by default) and the primers of
// OBIDNAPAT_PRIMERS ("forward,reverse"). It skips the test without the file.
func loadRealReads(t *testing.T) (reads [][]byte, primers []string) {
	path := os.Getenv("OBIDNAPAT_FASTQ")
	if path == "" {
		t.Skip("OBIDNAPAT_FASTQ not set")
	}
	primers = strings.Split(os.Getenv("OBIDNAPAT_PRIMERS"), ",")
	if len(primers) != 2 {
		t.Fatal("OBIDNAPAT_PRIMERS must be forward,reverse")
	}
	nReads := 20000
	if v, err := strconv.Atoi(os.Getenv("OBIDNAPAT_READS")); err == nil && v > 0 {
		nReads = v
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	var sc *bufio.Scanner
	if strings.HasSuffix(path, ".gz") {
		z, err := gzip.NewReader(f)
		if err != nil {
			t.Fatal(err)
		}
		sc = bufio.NewScanner(z)
	} else {
		sc = bufio.NewScanner(f)
	}
	sc.Buffer(make([]byte, 1<<20), 1<<26)
	for line := 0; sc.Scan() && len(reads) < nReads; line++ {
		if line%4 == 1 {
			reads = append(reads, append([]byte(nil), sc.Bytes()...))
		}
	}
	return reads, primers
}

// TestAmbiguityOnRealReads measures, on the reads of a FASTQ file
// (OBIDNAPAT_FASTQ, gzip or not; OBIDNAPAT_READS reads, 20000 by default), how
// many primer matches have an undetermined start or end, as in
// TestAmbiguityOfTheStart. The primers are those of OBIDNAPAT_PRIMERS
// ("forward,reverse"); the four patterns of obimultiplex are searched: each primer
// and its reverse complement. Skipped without OBIDNAPAT_FASTQ.
func TestAmbiguityOnRealReads(t *testing.T) {
	reads, primers := loadRealReads(t)
	t.Logf("%d reads", len(reads))

	for _, e := range []int{2, 3, 4} {
		fwd, err := MakeApatPattern(primers[0], e, true)
		if err != nil {
			t.Fatal(err)
		}
		rev, _ := MakeApatPattern(primers[1], e, true)
		cfwd, _ := fwd.ReverseComplement()
		crev, _ := rev.ReverseComplement()
		for _, pc := range []struct {
			name string
			p    ApatPattern
		}{{"forward", fwd}, {"reverse complement of reverse", crev}, {"reverse", rev}, {"reverse complement of forward", cfwd}} {
			var found, withErr, ambStart, ambEnd, ambAny, widthSum, noStart int
			for _, r := range reads {
				start, end, nerr, ok := pc.p.BestMatch(newSeq(r), 0, -1)
				_ = start
				if !ok {
					continue
				}
				found++
				if nerr == 0 {
					continue
				}
				withErr++
				earliest, latest, okr := startRange(pc.p, r, end-1, nerr, max(0, end-1-(pc.p.Len()+nerr)))
				if !okr {
					noStart++
					continue
				}
				sa := earliest != latest
				ea := false
				for _, h := range pc.p.c.search(nil, r, 0, -1, false) {
					if h[2] == nerr && h[1] != end && h[1] >= end-3 && h[1] <= end+3 {
						ea = true
					}
				}
				if sa {
					ambStart++
					widthSum += latest - earliest
				}
				if ea {
					ambEnd++
				}
				if sa || ea {
					ambAny++
				}
			}
			pc100 := func(n, d int) float64 {
				if d == 0 {
					return 0
				}
				return 100 * float64(n) / float64(d)
			}
			t.Logf("e=%d %-30s found %5d (%.1f%%), with errors %5d (%.1f%% of found) | start ambiguous %4d (%.1f%% of found, %.1f%% of those with errors, mean width %.1f) | end ambiguous %4d (%.1f%% of found) | either %.1f%% | no start %d",
				e, pc.name, found, pc100(found, len(reads)), withErr, pc100(withErr, found), ambStart, pc100(ambStart, found), pc100(ambStart, withErr),
				float64(widthSum)/float64(max(ambStart, 1)), ambEnd, pc100(ambEnd, found), pc100(ambAny, found), noStart)
		}
	}
}

// TestAllOccurrencesOnRealReads looks at every window FilterBestMatch-style of the
// four patterns in the reads, with the nominal start (not what obimultiplex gets,
// whose obiapat AllMatches realigns the windows: see
// TestAllMatchesAgainstObiapatOnRealReads), and tells: how many windows a
// read holds, how far the nominal start is from the true one (the interval of the
// starts of minimal cost), and, for the forward primer, how often the context of
// the tag ("GG" just before the start) singles out one start of that interval.
// Same environment variables as TestAmbiguityOnRealReads.
func TestAllOccurrencesOnRealReads(t *testing.T) {
	reads, primers := loadRealReads(t)

	const e = 3
	fwd, _ := MakeApatPattern(primers[0], e, true)
	rev, _ := MakeApatPattern(primers[1], e, true)
	cfwd, _ := fwd.ReverseComplement()
	crev, _ := rev.ReverseComplement()
	isGG := func(r []byte, s int) bool {
		return s >= 2 && (r[s-1]|32) == 'g' && (r[s-2]|32) == 'g'
	}
	for _, pc := range []struct {
		name string
		p    ApatPattern
	}{{"forward", fwd}, {"reverse complement of reverse", crev}, {"reverse", rev}, {"reverse complement of forward", cfwd}} {
		perRead := map[int]int{}
		var windows, withErr, nominalOut, ambig, ctxNone, ctxOne, ctxMany int
		offsets := map[int]int{}
		for _, r := range reads {
			ws := pc.p.FilterBestMatch(newSeq(r), 0, -1) // the nominal windows
			perRead[len(ws)]++
			for _, w := range ws {
				windows++
				if w[2] == 0 {
					continue
				}
				withErr++
				pos := w[1] - 1
				earliest, latest, ok := startRange(pc.p, r, pos, w[2], max(0, pos-(pc.p.Len()+w[2])))
				if !ok {
					continue
				}
				if w[0] < earliest || w[0] > latest {
					nominalOut++
					off := latest - w[0]
					if off < 0 {
						off = -off
					}
					offsets[min(off, 4)]++
				}
				if earliest != latest {
					ambig++
					n := 0
					for s := earliest; s <= latest; s++ {
						if isGG(r, s) {
							n++
						}
					}
					switch {
					case n == 0:
						ctxNone++
					case n == 1:
						ctxOne++
					default:
						ctxMany++
					}
				}
			}
		}
		t.Logf("%-30s windows per read %v | %d windows, %d with errors | nominal start outside the true interval: %d (%.1f%% of those with errors; distance to the latest start 1:%d 2:%d 3:%d 4+:%d) | ambiguous starts: %d, of which 'GG' before: no candidate %d, exactly one %d, several %d",
			pc.name, perRead, windows, withErr, nominalOut, 100*float64(nominalOut)/float64(max(withErr, 1)),
			offsets[1], offsets[2], offsets[3], offsets[4], ambig, ctxNone, ctxOne, ctxMany)
	}
}

// tagAnchoredStart looks for the tag in front of the primer window that starts
// at s: going left, the run of delimiters (the G's of the delimiter and of the
// primer itself), then the tag, a stretch without delimiter, then a delimiter. If
// that stretch is about tagLen bases long, the primer starts delimLen bases after
// its end.
func tagAnchoredStart(r []byte, s int, delim byte, delimLen, tagLen int) (start int, ok bool) {
	i := s - 1
	for i >= 0 && r[i]|32 == delim|32 {
		i--
	}
	end := i + 1 // end of the tag
	for i >= 0 && r[i]|32 != delim|32 {
		i--
	}
	if i < 0 { // no delimiter on the left of the tag
		return 0, false
	}
	if l := end - (i + 1); l < tagLen-1 || l > tagLen+1 {
		return 0, false
	}
	return end + delimLen, true
}

// TestTagAnchoredStartOnRealReads checks, on the primers that carry their tag on
// their left (forward, and reverse in reverse reads), whether the position of the
// tag gives the start of the primer: first on the windows whose start is not
// ambiguous (how often the tag predicts the true start), then on the ambiguous ones
// (how often the prediction singles out a start of the interval).
// OBIDNAPAT_TAGLEN (7), OBIDNAPAT_DELIM ("G") and OBIDNAPAT_DELIMLEN (2) describe the tags.
func TestTagAnchoredStartOnRealReads(t *testing.T) {
	reads, primers := loadRealReads(t)
	tagLen, delimLen, delim := 7, 2, byte('G')
	if v, err := strconv.Atoi(os.Getenv("OBIDNAPAT_TAGLEN")); err == nil {
		tagLen = v
	}
	if v, err := strconv.Atoi(os.Getenv("OBIDNAPAT_DELIMLEN")); err == nil {
		delimLen = v
	}
	if v := os.Getenv("OBIDNAPAT_DELIM"); v != "" {
		delim = v[0]
	}
	const e = 3
	for _, pc := range []struct{ name, primer string }{{"forward", primers[0]}, {"reverse", primers[1]}} {
		p, _ := MakeApatPattern(pc.primer, e, true)
		var windows, exact, exactTag, exactHit int
		var amb, ambTag, ambIn, ambInEarliest, ambInLatest int
		offsets := map[int]int{}
		for _, r := range reads {
			for _, w := range p.AllMatches(newSeq(r), 0, -1) {
				pos := w[1] - 1
				earliest, latest := w[0], w[0]
				if w[2] > 0 {
					var ok bool
					earliest, latest, ok = startRange(p, r, pos, w[2], max(0, pos-(p.Len()+w[2])))
					if !ok {
						continue
					}
				}
				windows++
				pred, found := tagAnchoredStart(r, latest, delim, delimLen, tagLen)
				if earliest == latest {
					exact++
					if found {
						exactTag++
						offsets[max(-3, min(3, pred-latest))]++
						if pred == latest {
							exactHit++
						}
					}
					continue
				}
				amb++
				if !found {
					continue
				}
				ambTag++
				if pred >= earliest && pred <= latest {
					ambIn++
					if pred == earliest {
						ambInEarliest++
					}
					if pred == latest {
						ambInLatest++
					}
				}
			}
		}
		t.Logf("%-8s %d windows. Start not ambiguous: %d, tag found for %d (%.0f%%), predicted start = true start for %d (%.0f%% of those); prediction minus true start (<=-3..>=3): %v. Start ambiguous: %d, tag found for %d, prediction inside the interval for %d (earliest %d, latest %d, between %d)",
			pc.name, windows, exact, exactTag, 100*float64(exactTag)/float64(max(exact, 1)), exactHit, 100*float64(exactHit)/float64(max(exactTag, 1)),
			[]int{offsets[-3], offsets[-2], offsets[-1], offsets[0], offsets[1], offsets[2], offsets[3]},
			amb, ambTag, ambIn, ambInEarliest, ambInLatest, ambIn-ambInEarliest-ambInLatest)
	}
}

// TestAllMatchesAgainstObiapatOnRealReads compares AllMatches of this package (the
// end found by the search, the latest start of that cost) with obiapat's, which
// realigns every window with errors by obialign.LocatePattern, on the reads of the
// FASTQ file. Windows are paired when they overlap.
func TestAllMatchesAgainstObiapatOnRealReads(t *testing.T) {
	reads, primers := loadRealReads(t)
	const e = 3
	for _, name := range []string{"forward", "reverse"} {
		primer := primers[0]
		if name == "reverse" {
			primer = primers[1]
		}
		base, _ := MakeApatPattern(primer, e, true)
		for _, rc := range []bool{false, true} {
			p := base
			lp, _ := obiapat.MakeApatPattern(primer, e, true)
			label := name
			if rc {
				p, _ = base.ReverseComplement()
				lp, _ = obiapat.MakeApatPattern(primer, e, true)
				lp, _ = lp.ReverseComplement()
				label += " (reverse complement)"
			}
			var ours, theirs, paired, sameAll, sameEnd, sameEndStartDiff, endDiff, errDiff, onlyOurs, onlyTheirs int
			var startLater, startEarlier, inInterval, startDiffTotal int
			var worse, equal, better, endLater, endEarlier int
			for _, r := range reads {
				ls, err := obiapat.MakeApatSequence(newSeq(append([]byte(nil), r...)), false)
				if err != nil {
					t.Fatal(err)
				}
				mine := p.AllMatches(newSeq(r), 0, -1)
				their := lp.AllMatches(ls, 0, -1)
				ours += len(mine)
				theirs += len(their)
				used := make([]bool, len(their))
				for _, w := range mine {
					found := -1
					for j, x := range their {
						if !used[j] && x[0] < w[1] && w[0] < x[1] {
							found = j
							break
						}
					}
					if found < 0 {
						onlyOurs++
						continue
					}
					used[found] = true
					x := their[found]
					paired++
					switch {
					case x == w:
						sameAll++
					case x[1] == w[1]:
						sameEnd++
						if x[0] != w[0] {
							sameEndStartDiff++
							startDiffTotal++
							if x[0] < w[0] {
								startEarlier++
							} else {
								startLater++
							}
							if w[2] > 0 {
								earliest, latest, ok := startRange(p, r, w[1]-1, w[2], max(0, w[1]-1-(p.Len()+w[2])))
								if ok && x[0] >= earliest && x[0] <= latest {
									inInterval++
								}
							}
						}
						if x[2] != w[2] {
							errDiff++
						}
					default:
						endDiff++
						// the cost of obiapat's window under this package's costs, against ours
						c := globalCost(p, r[x[0]:x[1]])
						switch {
						case c > w[2]:
							worse++
						case c == w[2]:
							equal++
						default:
							better++
						}
						if x[1] > w[1] {
							endLater++
						} else {
							endEarlier++
						}
					}
				}
				for _, u := range used {
					if !u {
						onlyTheirs++
					}
				}
			}
			t.Logf("%-30s windows: ours %d, obiapat %d; paired %d: identical %d, same end %d (start differs %d: obiapat's earlier %d, later %d, inside our interval of minimal cost %d; errors differ %d), different end %d (obiapat's window costs more than ours: %d, the same: %d, less: %d; its end later %d, earlier %d); only ours %d, only obiapat's %d",
				label, ours, theirs, paired, sameAll, sameEnd, sameEndStartDiff, startEarlier, startLater, inInterval, errDiff, endDiff, worse, equal, better, endLater, endEarlier, onlyOurs, onlyTheirs)
		}
	}
}
