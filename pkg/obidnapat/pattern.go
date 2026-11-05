package obidnapat

import (
	"fmt"
	"strings"
)

// MaxCoreLength is the longest pattern accepted, the wildcards at both ends
// (NNNN...) set aside (but not with indels): the state of the search is one 64 bit word, with a guard
// bit. obiapat announced 64, but its search finds nothing at 64 symbols and
// returns negative positions beyond.
const MaxCoreLength = 63

// maxLevels is the number of error levels of the widest search.
const maxLevels = MaxCoreLength + 1

// compiled is the immutable result of the compilation of a pattern.
type compiled struct {
	text      string
	raw       string // the pattern as given, in upper case: what obialign.LocatePattern realigns (BestMatch)
	positions []position
	maxErr    int

	// The pattern is its core, framed by left and right wildcards. Only the
	// core is searched; the frame only decides which windows fit in the sequence.
	left, right int
	coreLen     int
	levels      int  // error levels searched: min(maxErr, coreLen)
	bndm        bool // the strict search of a linear sequence uses BNDM (see strict.go)

	// indel: insertions and deletions are allowed (the pattern was compiled with
	// allowsIndel and at least one error). The whole pattern is then the core: no
	// wildcard is set aside at its ends (see scan_indel.go).
	indel       bool
	allowsIndel bool // the flag given to MakeApatPattern, kept for the reverse complement and BestMatch

	mask  [256]uint64 // mask[b] has bit i when byte b matches the core position i
	end   uint64      // bit of the last core position
	allow uint64      // bit i when an error is allowed at the core position i
}

// ApatPattern is a compiled pattern. It is immutable and can be used by several
// goroutines at once.
type ApatPattern struct {
	c *compiled
}

// MakeApatPattern compiles a pattern with at most errormax substitutions. The
// pattern is read as described in the package documentation: IUPAC letters, [AG]
// classes, ! negations, # to forbid an error after a symbol. With allowsIndel the
// errors may also be insertions and deletions (Manber's search, as obiapat's
// ManberIndel; see scan_indel.go), and the pattern is searched whole: the limit of
// MaxCoreLength positions then counts its wildcards at the ends too.
func MakeApatPattern(pattern string, errormax int, allowsIndel bool) (ApatPattern, error) {
	if errormax < 0 {
		return ApatPattern{}, fmt.Errorf("obidnapat: the number of errors cannot be negative")
	}
	ps, err := parse(pattern)
	if err != nil {
		return ApatPattern{}, err
	}

	c := &compiled{raw: strings.ToUpper(pattern), positions: ps, maxErr: errormax, allowsIndel: allowsIndel, indel: allowsIndel && errormax > 0}
	for _, p := range ps {
		c.text += p.text
	}
	if !c.indel {
		for c.left < len(ps) && ps[c.left].wildcard {
			c.left++
		}
		if c.left < len(ps) {
			for ps[len(ps)-1-c.right].wildcard {
				c.right++
			}
		}
	}
	core := ps[c.left : len(ps)-c.right]
	c.coreLen = len(core)
	if c.coreLen > MaxCoreLength {
		return ApatPattern{}, fmt.Errorf("obidnapat: pattern %q: %d positions once the wildcards at both ends are set aside, the limit is %d",
			pattern, c.coreLen, MaxCoreLength)
	}
	c.levels = min(errormax, c.coreLen)
	if c.indel { // insertions can bring the errors above the length of the pattern
		c.levels = min(errormax, maxLevels-1)
	}
	c.bndm = c.levels == 0 && c.advance() >= minAdvance // (false when indel: levels > 0)

	if c.coreLen > 0 {
		for b := range 256 {
			for i, p := range core {
				if p.wildcard || letterSet[b]&p.set != 0 {
					c.mask[b] |= 1 << uint(i)
				}
			}
		}
		c.end = 1 << uint(c.coreLen-1)
		for i, p := range core {
			if !p.oblig {
				c.allow |= 1 << uint(i)
			}
		}
	}
	return ApatPattern{c}, nil
}

// String returns the pattern, in upper case.
func (p ApatPattern) String() string { return p.c.text }

// Len returns the number of positions of the pattern, wildcards at both ends
// included: it is the length of the region a match covers.
func (p ApatPattern) Len() int { return len(p.c.positions) }

// MaxErrors returns the number of errors allowed.
func (p ApatPattern) MaxErrors() int { return p.c.maxErr }

// ReverseComplement returns the pattern matching the opposite strand.
func (p ApatPattern) ReverseComplement() (ApatPattern, error) {
	return MakeApatPattern(reverseComplementText(p.c.positions), p.c.maxErr, p.c.allowsIndel)
}
