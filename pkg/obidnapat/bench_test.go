package obidnapat

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

const benchCore = "ACGTTGCATGCAAGTCGATC"

func benchSeq(n int) []byte {
	return randomSequence(rand.New(rand.NewSource(3)), "acgt", n, "")
}

// BenchmarkScan reads a sequence of 1 Mb for a pattern of 20 positions, with 0 to
// 5 errors allowed, with obiapat (the sequence is encoded once, outside of the
// measure) and with this package. MB/s is millions of bases per second.
func BenchmarkScan(b *testing.B) {
	seq := benchSeq(1 << 20)
	bs := newSeq(seq)
	ls, _ := obiapat.MakeApatSequence(newSeq(append([]byte(nil), seq...)), false)
	for _, e := range []int{0, 1, 2, 3, 4, 5, 6, 8} {
		lp, _ := obiapat.MakeApatPattern(benchCore, e, false)
		p, _ := MakeApatPattern(benchCore, e, false)
		b.Run(fmt.Sprintf("obiapat/e=%d", e), func(b *testing.B) {
			b.SetBytes(int64(len(seq)))
			for i := 0; i < b.N; i++ {
				lp.FindAllIndex(ls, 0, -1)
			}
		})
		b.Run(fmt.Sprintf("new/e=%d", e), func(b *testing.B) {
			b.SetBytes(int64(len(seq)))
			var dst [][3]int
			for i := 0; i < b.N; i++ {
				dst = p.AppendAllIndex(dst[:0], bs, 0, -1)
			}
		})
	}
}

// BenchmarkReads is the cost per sequence for short sequences (reads of 150
// bases), where obiapat pays its encoding and its cgo call for every one.
func BenchmarkReads(b *testing.B) {
	rng := rand.New(rand.NewSource(5))
	var reads [][]byte
	for i := 0; i < 1000; i++ {
		reads = append(reads, randomSequence(rng, "acgt", 150, ""))
	}
	for _, e := range []int{0, 3} {
		lp, _ := obiapat.MakeApatPattern(benchCore, e, false)
		p, _ := MakeApatPattern(benchCore, e, false)
		b.Run(fmt.Sprintf("obiapat/e=%d", e), func(b *testing.B) {
			b.SetBytes(150 * 1000)
			for i := 0; i < b.N; i++ {
				for _, r := range reads {
					ls, _ := obiapat.MakeApatSequence(newSeq(r), false)
					lp.FindAllIndex(ls, 0, -1)
				}
			}
		})
		b.Run(fmt.Sprintf("new/e=%d", e), func(b *testing.B) {
			b.SetBytes(150 * 1000)
			var dst [][3]int
			for i := 0; i < b.N; i++ {
				for _, r := range reads {
					dst = p.AppendAllIndex(dst[:0], newSeq(r), 0, -1)
				}
			}
		})
	}
}

// BenchmarkFlanks is the same pattern framed by wildcards (the NNNN...pattern
// NNNN used to get the context of a site), 1 Mb, e=3.
func BenchmarkFlanks(b *testing.B) {
	seq := benchSeq(1 << 20)
	bs := newSeq(seq)
	ls, _ := obiapat.MakeApatSequence(newSeq(append([]byte(nil), seq...)), false)
	framed := "NNNNNNNNNNN" + benchCore + "NNNNNNNNNNNNN" // 44 positions
	lp, _ := obiapat.MakeApatPattern(framed, 3, false)
	p, _ := MakeApatPattern(framed, 3, false)
	b.Run("obiapat", func(b *testing.B) {
		b.SetBytes(int64(len(seq)))
		for i := 0; i < b.N; i++ {
			lp.FindAllIndex(ls, 0, -1)
		}
	})
	b.Run("new", func(b *testing.B) {
		b.SetBytes(int64(len(seq)))
		var dst [][3]int
		for i := 0; i < b.N; i++ {
			dst = p.AppendAllIndex(dst[:0], bs, 0, -1)
		}
	})
}

func withBNDM(p ApatPattern, on bool) ApatPattern {
	q := ApatPattern{c: new(compiled)}
	*q.c = *p.c
	q.c.bndm = on
	return q
}

// BenchmarkStrict compares, for no error allowed, obiapat, the Shift-And and the
// BNDM, on 1 Mb, by length of the core and by number of degenerate positions.
func BenchmarkStrict(b *testing.B) {
	rng := rand.New(rand.NewSource(9))
	seq := benchSeq(1 << 20)
	bs := newSeq(seq)
	ls, _ := obiapat.MakeApatSequence(newSeq(append([]byte(nil), seq...)), false)
	random := func(n int) string {
		out := make([]byte, n)
		for i := range out {
			out[i] = "ACGT"[rng.Intn(4)]
		}
		return string(out)
	}
	degenerate := func(core string, k int, code string) string {
		out := []byte(core)
		for _, i := range rng.Perm(len(out))[:k] {
			out[i] = code[rng.Intn(len(code))]
		}
		return string(out)
	}
	cases := map[string]string{}
	var names []string
	add := func(name, pat string) { cases[name] = pat; names = append(names, name) }
	for _, m := range []int{4, 6, 8, 10, 12, 16, 20, 30, 50, 63} {
		add(fmt.Sprintf("len=%02d", m), random(m))
	}
	for _, k := range []int{3, 6, 10} {
		add(fmt.Sprintf("len=20/%02d two-base codes", k), degenerate(random(20), k, "RYSWKM"))
	}
	add("len=20/03 three-base codes", degenerate(random(20), 3, "BDHV"))
	add("len=20/02 N inside", degenerate(random(20), 2, "N"))
	add("len=20/12 N inside", degenerate(random(20), 12, "N"))
	add("len=20/16 two-base codes", degenerate(random(20), 16, "RYSWKM"))
	add("len=20/20 two-base codes", degenerate(random(20), 20, "RYSWKM"))
	add("len=20/10 N+three-base", degenerate(degenerate(random(20), 10, "N"), 6, "BDHV"))
	add("len=10/03 two-base codes", degenerate(random(10), 3, "RYSWKM"))
	add("len=12/06 two-base codes", degenerate(random(12), 6, "RYSWKM"))
	for _, name := range names {
		pat := cases[name]
		lp, _ := obiapat.MakeApatPattern(pat, 0, false)
		p, _ := MakeApatPattern(pat, 0, false)
		b.Run(name+"/obiapat", func(b *testing.B) {
			b.SetBytes(int64(len(seq)))
			for i := 0; i < b.N; i++ {
				lp.FindAllIndex(ls, 0, -1)
			}
		})
		for _, v := range []struct {
			name string
			on   bool
		}{{"shiftand", false}, {"bndm", true}} {
			q := withBNDM(p, v.on)
			b.Run(name+"/"+v.name, func(b *testing.B) {
				b.SetBytes(int64(len(seq)))
				var dst [][3]int
				for i := 0; i < b.N; i++ {
					dst = q.AppendAllIndex(dst[:0], bs, 0, -1)
				}
			})
		}
	}
}

// BenchmarkIndel reads 1 Mb for a pattern of 20 positions with 1 to 5 errors
// where insertions and deletions count, with obiapat (ManberIndel) and with this package.
func BenchmarkIndel(b *testing.B) {
	seq := benchSeq(1 << 20)
	bs := newSeq(seq)
	ls, _ := obiapat.MakeApatSequence(newSeq(append([]byte(nil), seq...)), false)
	for _, e := range []int{1, 2, 3, 5} {
		lp, _ := obiapat.MakeApatPattern(benchCore, e, true)
		p, _ := MakeApatPattern(benchCore, e, true)
		b.Run(fmt.Sprintf("obiapat/e=%d", e), func(b *testing.B) {
			b.SetBytes(int64(len(seq)))
			for i := 0; i < b.N; i++ {
				lp.FindAllIndex(ls, 0, -1)
			}
		})
		b.Run(fmt.Sprintf("new/e=%d", e), func(b *testing.B) {
			b.SetBytes(int64(len(seq)))
			var dst [][3]int
			for i := 0; i < b.N; i++ {
				dst = p.AppendAllIndex(dst[:0], bs, 0, -1)
			}
		})
	}
}

// BenchmarkRefine is the cost per read of BestMatch with indels (1000 reads of
// 150 bases, the pattern planted with up to two indels, 3 errors allowed): the
// forward search alone, then with the realignment by obialign.LocatePattern, then
// with the backward bit-parallel search for the start.
func BenchmarkRefine(b *testing.B) {
	rng := rand.New(rand.NewSource(5))
	p, _ := MakeApatPattern(benchCore, 3, true)
	var reads []*obiseq.BioSequence
	for i := 0; i < 1000; i++ {
		r := randomSequence(rng, "acgt", 150, "")
		m := []byte(mutate(rng, strings.ToLower(benchCore)))
		copy(r[40+rng.Intn(60):], m)
		reads = append(reads, newSeq(r))
	}
	b.Run("forward only", func(b *testing.B) {
		var dst [][3]int
		for i := 0; i < b.N; i++ {
			for _, r := range reads {
				dst = p.AppendAllIndex(dst[:0], r, 0, -1)
			}
		}
	})
	b.Run("LocatePattern", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, r := range reads {
				p.BestMatch(r, 0, -1)
			}
		}
	})
	b.Run("backward", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, r := range reads {
				p.BestMatchBackward(r, 0, -1)
			}
		}
	})
	lp, _ := obiapat.MakeApatPattern(benchCore, 3, true)
	b.Run("AllMatches obiapat (LocatePattern)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, r := range reads {
				ls, _ := obiapat.MakeApatSequence(r, false)
				lp.AllMatches(ls, 0, -1)
			}
		}
	})
	b.Run("AllMatches backward", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, r := range reads {
				p.AllMatches(r, 0, -1)
			}
		}
	})
}
