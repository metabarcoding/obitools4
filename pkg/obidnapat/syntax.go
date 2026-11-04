// Package obidnapat searches DNA patterns, with IUPAC codes, in obitools
// sequences. It is a pure Go, DNA specialised replacement of obiapat (the cgo
// wrapper of apat, an agrep derivative), written to be faster and to fix its
// quirks:
//
//   - a pattern has at most MaxCoreLength (63) positions once the wildcards at
//     both ends (NNNN...) are set aside; longer ones are refused, not silently
//     mis-searched;
//   - a letter of the SEQUENCE that stands for several bases (n, r, y...)
//     matches a pattern position as soon as one of the bases it stands for is
//     accepted there (obiapat counted it as an error);
//   - no sequence copy or encoding: the bytes of the sequence are scanned in
//     place, and results are returned per call, so patterns and sequences are
//     safe for concurrent use.
//
// This first version searches with substitutions only (Hamming distance), the
// only mode used by obipcr.
package obidnapat

import (
	"fmt"
	"strings"
)

// The four bases, as bits of a set.
const (
	baseA uint8 = 1 << iota
	baseC
	baseG
	baseT
	anyBase = baseA | baseC | baseG | baseT
)

// letterSet gives, for every byte, the bases it stands for: the IUPAC codes, in
// either case, with u read as t. A byte that is not an IUPAC letter stands for
// no base, and so matches nothing but a wildcard position.
var letterSet = func() (t [256]uint8) {
	for letter, set := range map[byte]uint8{
		'A': baseA, 'C': baseC, 'G': baseG, 'T': baseT, 'U': baseT,
		'R': baseA | baseG, 'Y': baseC | baseT, 'S': baseC | baseG, 'W': baseA | baseT,
		'K': baseG | baseT, 'M': baseA | baseC,
		'B': baseC | baseG | baseT, 'D': baseA | baseG | baseT,
		'H': baseA | baseC | baseT, 'V': baseA | baseC | baseG,
		'N': anyBase,
	} {
		t[letter], t[letter|0x20] = set, set
	}
	return
}()

// complementLetter is the complement of an IUPAC code (0 for the other bytes).
var complementLetter = func() (t [256]byte) {
	for from, to := range map[byte]byte{
		'A': 'T', 'C': 'G', 'G': 'C', 'T': 'A', 'U': 'A',
		'R': 'Y', 'Y': 'R', 'S': 'S', 'W': 'W', 'K': 'M', 'M': 'K',
		'B': 'V', 'V': 'B', 'D': 'H', 'H': 'D', 'N': 'N', 'X': 'X',
	} {
		t[from] = to
	}
	return
}()

// position is one position of a pattern.
type position struct {
	set      uint8  // bases accepted (0: none, as for !N)
	wildcard bool   // accepts any byte: N, X, or any set holding the four bases
	oblig    bool   // # : no error allowed here
	text     string // the symbol as written (!, class and # included)
}

// patternSet returns the bases of a pattern letter. X is read as N.
func patternSet(c byte) (uint8, bool) {
	if c == 'X' {
		return anyBase, true
	}
	s := letterSet[c]
	return s, s != 0
}

// parse reads a pattern: IUPAC letters, classes [AG], negations ! (of a letter
// or of a class), and # after a symbol to forbid an error there. The pattern is
// read in upper case.
func parse(pattern string) ([]position, error) {
	pat := strings.ToUpper(pattern)
	var out []position
	for i := 0; i < len(pat); {
		start := i
		negated := false
		for i < len(pat) && pat[i] == '!' {
			negated = !negated
			i++
		}
		if i >= len(pat) {
			return nil, fmt.Errorf("obidnapat: pattern %q: '!' with nothing to negate", pattern)
		}
		var set uint8
		if pat[i] == '[' {
			end := strings.IndexByte(pat[i:], ']')
			if end < 0 {
				return nil, fmt.Errorf("obidnapat: pattern %q: '[' is not closed", pattern)
			}
			body := pat[i+1 : i+end]
			if body == "" {
				return nil, fmt.Errorf("obidnapat: pattern %q: empty class", pattern)
			}
			for j := 0; j < len(body); j++ {
				s, ok := patternSet(body[j])
				if !ok {
					return nil, fmt.Errorf("obidnapat: pattern %q: %q is not allowed in a class", pattern, body[j])
				}
				set |= s
			}
			i += end + 1
		} else {
			s, ok := patternSet(pat[i])
			if !ok {
				return nil, fmt.Errorf("obidnapat: pattern %q: unexpected %q", pattern, pat[i])
			}
			set = s
			i++
		}
		if negated {
			set = ^set & anyBase
		}
		oblig := false
		if i < len(pat) && pat[i] == '#' {
			oblig = true
			i++
		}
		out = append(out, position{set: set, wildcard: set == anyBase, oblig: oblig, text: pat[start:i]})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("obidnapat: empty pattern")
	}
	return out, nil
}

// reverseComplementText writes the pattern that matches the opposite strand:
// the symbols in the reverse order, each one complemented (a class and a
// negation stay what they are).
func reverseComplementText(ps []position) string {
	var b strings.Builder
	for i := len(ps) - 1; i >= 0; i-- {
		for j := 0; j < len(ps[i].text); j++ {
			c := ps[i].text[j]
			if cc := complementLetter[c]; cc != 0 {
				c = cc
			}
			b.WriteByte(c)
		}
	}
	return b.String()
}
