package obidnapat

import (
	"fmt"
	"math/rand"
	"testing"
)

// BenchmarkSet compares a PatternSet (N motifs of 20 positions, both strands)
// with the 2N scans one after the other, on 1 Mb. MB/s is the megabases of the
// sequence read per second for the whole set.
func BenchmarkSet(b *testing.B) {
	seq := benchSeq(1 << 20)
	bs := newSeq(seq)
	rng := rand.New(rand.NewSource(9))
	for _, e := range []int{0, 2} {
		for _, n := range []int{1, 2, 4, 8, 32, 200} {
			var ps, all []ApatPattern
			for i := 0; i < n; i++ {
				p, _ := MakeApatPattern(exactPattern(rng, 20), e, false)
				rc, _ := p.ReverseComplement()
				ps, all = append(ps, p), append(all, p, rc)
			}
			set, _ := MakePatternSet(ps, true)
			b.Run(fmt.Sprintf("separate/e=%d/n=%d", e, n), func(b *testing.B) {
				b.SetBytes(int64(len(seq)))
				var dst [][3]int
				for i := 0; i < b.N; i++ {
					for _, p := range all {
						dst = p.AppendAllIndex(dst[:0], bs, 0, -1)
					}
				}
			})
			b.Run(fmt.Sprintf("set/e=%d/n=%d", e, n), func(b *testing.B) {
				b.SetBytes(int64(len(seq)))
				var dst []Occurrence
				for i := 0; i < b.N; i++ {
					dst = set.AppendAll(dst[:0], bs, 0, -1)
				}
			})
		}
	}
}

// BenchmarkSetLen is BenchmarkSet for 8 motifs (16 entries with both strands) of
// growing length: the number of cores a word holds falls from 3 (20) to 1 (> 32).
func BenchmarkSetLen(b *testing.B) {
	seq := benchSeq(1 << 20)
	bs := newSeq(seq)
	rng := rand.New(rand.NewSource(9))
	for _, e := range []int{0, 2} {
		for _, l := range []int{10, 20, 32, 33, 40, 63} {
			var ps, all []ApatPattern
			for i := 0; i < 8; i++ {
				p, _ := MakeApatPattern(exactPattern(rng, l), e, false)
				rc, _ := p.ReverseComplement()
				ps, all = append(ps, p), append(all, p, rc)
			}
			set, _ := MakePatternSet(ps, true)
			b.Run(fmt.Sprintf("separate/e=%d/len=%d", e, l), func(b *testing.B) {
				b.SetBytes(int64(len(seq)))
				var dst [][3]int
				for i := 0; i < b.N; i++ {
					for _, p := range all {
						dst = p.AppendAllIndex(dst[:0], bs, 0, -1)
					}
				}
			})
			b.Run(fmt.Sprintf("set/e=%d/len=%d/words=%d", e, l, set.Words()), func(b *testing.B) {
				b.SetBytes(int64(len(seq)))
				var dst []Occurrence
				for i := 0; i < b.N; i++ {
					dst = set.AppendAll(dst[:0], bs, 0, -1)
				}
			})
		}
	}
}

// BenchmarkSetVector compares, for patterns of growing length with 2 errors
// (8 patterns, both strands), the scalar words with the words run on SIMD
// vectors. Without GOEXPERIMENT=simd both are the scalar ones.
func BenchmarkSetVector(b *testing.B) {
	seq := benchSeq(1 << 20)
	bs := newSeq(seq)
	rng := rand.New(rand.NewSource(9))
	for _, l := range []int{10, 20, 33, 63} {
		for _, n := range []int{8, 64} {
			var ps []ApatPattern
			for i := 0; i < n; i++ {
				p, _ := MakeApatPattern(exactPattern(rng, l), 2, false)
				ps = append(ps, p)
			}
			for _, vec := range []bool{false, true} {
				set, _ := makePatternSet(ps, true, vec)
				b.Run(fmt.Sprintf("vector=%v/len=%d/n=%d/words=%d/groups=%d", vec, l, n, set.Words(), len(set.groups)), func(b *testing.B) {
					b.SetBytes(int64(len(seq)))
					var dst []Occurrence
					for i := 0; i < b.N; i++ {
						dst = set.AppendAll(dst[:0], bs, 0, -1)
					}
				})
			}
		}
	}
}

// BenchmarkBatch compares the number of words advanced in one loop (1 to 3), for
// 8 or 64 patterns of growing length, both strands.
func BenchmarkBatch(b *testing.B) {
	seq := benchSeq(1 << 20)
	bs := newSeq(seq)
	defer func(old int) { forcedBatch = old }(forcedBatch)
	for _, e := range []int{1, 2, 3, 4, 6, 8} {
		for _, l := range []int{20, 63} {
			for _, n := range []int{8} {
				for k := 1; k <= 3; k++ {
					forcedBatch = k
					rng := rand.New(rand.NewSource(9))
					var ps []ApatPattern
					for i := 0; i < n; i++ {
						p, _ := MakeApatPattern(exactPattern(rng, l), e, false)
						ps = append(ps, p)
					}
					set, _ := MakePatternSet(ps, true)
					b.Run(fmt.Sprintf("e=%d/len=%d/n=%d/k=%d", e, l, n, k), func(b *testing.B) {
						b.SetBytes(int64(len(seq)))
						var dst []Occurrence
						for i := 0; i < b.N; i++ {
							dst = set.AppendAll(dst[:0], bs, 0, -1)
						}
					})
				}
			}
		}
	}
}
