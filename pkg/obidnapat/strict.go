package obidnapat

import (
	"math"
	"math/bits"
)

// The strict search (no error allowed) reads fewer bases than the text holds.
//
// BNDM (backward nondeterministic DAWG matching, Navarro and Raffinot) reads
// each window of the text from its right end towards its left end, and stops as
// soon as what has been read cannot be a factor of the pattern. Then the window
// jumps forward, by up to the length of the pattern: on DNA, a pattern of 20
// positions is checked by reading about three bases per 17 positions. It uses
// the same masks as the Shift-And of the kernels (bit i of mask[b] is set when
// the position i of the core accepts the byte b), so that the sets of bases of
// the IUPAC codes, the classes and the wildcards need nothing more.
//
// F is the set of the starts s such that the part of the window read so far
// (of length r) equals the factor of the core that starts at s. Reading one more
// byte c to the left gives F' = (F >> 1) & mask[c]. When the start 0 is in F, the
// part read is a prefix of the core: the next window must not start further than
// where this prefix starts (last), and when the whole window has been read, the
// core is there.

// minAdvance is the number of bases BNDM must advance per base read, at least,
// for it to be used (see advance): below it, a step costs more than it brings
// and the Shift-And, which reads every base but with a cheaper step, is faster.
// Fixed on the measures of BenchmarkStrict.
const minAdvance = 3.6

// advance estimates, for a text of uniform random bases, how many positions BNDM
// moves forward per base it reads: the length of the core divided by the
// expected number of bases read per window. A base is read as long as what has
// been read can still be a factor of the core; the chance that it can after r
// bases is bounded by the sum, over the factors of r positions, of the chance
// that r random bases fit them (a position accepting k of the 4 bases is fitted
// with the chance k/4).
func (c *compiled) advance() float64 {
	m := c.coreLen
	if m == 0 {
		return 0
	}
	core := c.positions[c.left : len(c.positions)-c.right]
	accept := make([]float64, m)
	for i, p := range core {
		accept[i] = float64(bits.OnesCount8(p.set)) / 4
		if p.wildcard {
			accept[i] = 1
		}
	}
	fits := make([]float64, m+1) // fits[r]: sum over the factors of r positions
	for s := 0; s < m; s++ {
		prod := 1.0
		for r := 1; s+r <= m; r++ {
			prod *= accept[s+r-1]
			fits[r] += prod
		}
	}
	reads := 1.0
	for r := 1; r < m; r++ {
		reads += math.Min(1, fits[r])
	}
	return float64(m) / reads
}

// runBNDM appends the matches of the core in text, whose first byte is the
// position base of the sequence, by increasing position. The core must lie in
// the text, which the caller guarantees by cutting it to the region.
func (c *compiled) runBNDM(dst [][3]int, text []byte, base int) [][3]int {
	m := c.coreLen
	all := (uint64(1)<<uint(m) - 1) << 1 // the starts 1..m: one more byte to read
	mask := &c.mask
	off, n := c.offset(base), len(c.positions)
	for pos := 0; pos <= len(text)-m; {
		f, j, last := all, m, m
		for {
			f = (f >> 1) & mask[text[pos+j-1]]
			j--
			if f&1 != 0 {
				if j == 0 { // the whole window has been read
					dst = append(dst, [3]int{off + pos + c.coreLen - 1, off + pos + c.coreLen - 1 + n, 0})
					break
				}
				last = j
			}
			if f == 0 || j == 0 {
				break
			}
		}
		pos += last
	}
	return dst
}
