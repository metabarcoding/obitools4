package obialign

import (
	"fmt"
	"math/rand"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

// randomDNA returns a random ACGT sequence of length n.
func randomDNA(rng *rand.Rand, n int) []byte {
	const nuc = "ACGT"
	s := make([]byte, n)
	for i := range s {
		s[i] = nuc[rng.Intn(4)]
	}
	return s
}

// mutate returns a copy of s in which each position is, with probability
// rate, substituted, deleted or followed by an inserted base (equiprobable).
func mutate(rng *rand.Rand, s []byte, rate float64) []byte {
	const nuc = "ACGT"
	out := make([]byte, 0, len(s)+len(s)/10)
	for _, c := range s {
		if rng.Float64() >= rate {
			out = append(out, c)
			continue
		}
		switch rng.Intn(3) {
		case 0:
			out = append(out, nuc[rng.Intn(4)])
		case 1: // deletion
		case 2:
			out = append(out, c, nuc[rng.Intn(4)])
		}
	}
	return out
}

// benchPairs builds a fixed set of pairs. "similar" pairs differ by about
// 2% of mutated positions, "random" pairs are unrelated sequences.
func benchPairs(n, length int, similar bool) [][2]*obiseq.BioSequence {
	rng := rand.New(rand.NewSource(42))
	pairs := make([][2]*obiseq.BioSequence, n)
	for i := range pairs {
		a := randomDNA(rng, length)
		var b []byte
		if similar {
			b = mutate(rng, a, 0.02)
		} else {
			b = randomDNA(rng, length)
		}
		pairs[i] = [2]*obiseq.BioSequence{
			obiseq.NewBioSequence("a", a, ""),
			obiseq.NewBioSequence("b", b, ""),
		}
	}
	return pairs
}

// BenchmarkLegacyLCS measures the historical implementation (the one
// FastLCSScore used before BandedLCS), kept as the "before" reference.
func BenchmarkLegacyLCS(b *testing.B) {
	for _, similar := range []bool{true, false} {
		kind := "random"
		if similar {
			kind = "similar"
		}
		for _, length := range []int{100, 500, 2000} {
			pairs := benchPairs(64, length, similar)
			for _, maxErr := range []int{-1, 5, length / 20, length / 5} {
				name := fmt.Sprintf("%s/len=%d/maxErr=%d", kind, length, maxErr)
				b.Run(name, func(b *testing.B) {
					var buffer []uint64
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						p := pairs[i%len(pairs)]
						FastLCSEGFScoreByte(p[0].Sequence(), p[1].Sequence(), maxErr, false, &buffer)
					}
				})
			}
		}
	}
}

// BenchmarkBandedLCS runs on the same pairs as BenchmarkLegacyLCS.
func BenchmarkBandedLCS(b *testing.B) {
	for _, similar := range []bool{true, false} {
		kind := "random"
		if similar {
			kind = "similar"
		}
		for _, length := range []int{100, 500, 2000} {
			pairs := benchPairs(64, length, similar)
			for _, maxErr := range []int{-1, 5, length / 20, length / 5} {
				name := fmt.Sprintf("%s/len=%d/maxErr=%d", kind, length, maxErr)
				b.Run(name, func(b *testing.B) {
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						p := pairs[i%len(pairs)]
						BandedLCS(p[0].Sequence(), p[1].Sequence(), maxErr)
					}
				})
			}
		}
	}
}
