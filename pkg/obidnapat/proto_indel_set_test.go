package obidnapat

import (
	"math/rand"
	"testing"
)

// A throw-away prototype: several patterns with indels packed in one 64 bit word,
// run with the recurrence of scan_indel.go where the start bit is that of every
// core, compared hit by hit with the single pattern search.
type protoMember struct {
	c     *compiled
	shift uint
	end   uint64
}

func TestProtoPackedIndel(t *testing.T) {
	rng := rand.New(rand.NewSource(51))
	hitsChecked, multi := 0, 0
	for iter := 0; iter < iterations(8000); iter++ {
		// patterns with indels, packed while they fit
		var members []protoMember
		var word struct {
			mask              [256]uint64
			start, end, allow uint64
		}
		used, levels := 0, 0
		for tries := 0; tries < 6; tries++ {
			tx := randomPattern(rng, 1+rng.Intn(22), true)
			p, err := MakeApatPattern(tx, 1+rng.Intn(5), true)
			if err != nil || used+len(p.c.positions) > 64 {
				continue
			}
			c := p.c
			m := protoMember{c: c, shift: uint(used), end: c.end << uint(used)}
			for b := range word.mask {
				word.mask[b] |= c.mask[b] << m.shift
			}
			word.start |= 1 << m.shift
			word.end |= m.end
			word.allow |= c.allow << m.shift
			members = append(members, m)
			used += len(c.positions)
			levels = max(levels, c.levels)
		}
		if len(members) == 0 {
			continue
		}
		if len(members) > 1 {
			multi++
		}
		alphabet := []string{"acgt", "acgt", "acgtn", "acgtrykmn-"}[rng.Intn(4)]
		seq := randomSequence(rng, alphabet, rng.Intn(200), "")
		for k := 0; k < 3 && len(seq) > 30; k++ {
			c := members[rng.Intn(len(members))].c
			site := []byte(mutate(rng, string(concreteFor(rng, c))))
			if len(site) < len(seq) {
				copy(seq[rng.Intn(len(seq)-len(site)+1):], site)
			}
		}

		// the packed recurrence
		var o [maxLevels]uint64
		for j := 0; j <= levels; j++ {
			for _, m := range members {
				full := m.c.end | (m.c.end - 1)
				o[j] |= ((1<<uint(j) - 1) & full) << m.shift
			}
		}
		got := make([][][3]int, len(members))
		for pos, b := range seq {
			mk := word.mask[b]
			var n [maxLevels]uint64
			n[0] = ((o[0] << 1) | word.start) & mk
			for j := 1; j <= levels; j++ {
				n[j] = (((o[j] << 1) | word.start) & mk) |
					((o[j-1] | ((o[j-1] << 1) | word.start) | (n[j-1] << 1)) & word.allow)
			}
			for k, m := range members {
				for j := 0; j <= m.c.levels; j++ {
					if n[j]&m.end != 0 {
						got[k] = append(got[k], [3]int{pos - len(m.c.positions) + 1, pos + 1, j})
						break
					}
				}
			}
			o = n
		}
		for k, m := range members {
			want := m.c.indelHits(nil, seq, 0, -1, false)
			if !sameMatches(got[k], nilIfEmpty(want)) {
				t.Fatalf("member %d of %d, pattern %q (%d levels), seq %q\n packed %v\n single %v", k, len(members), m.c.text, m.c.levels, seq, got[k], want)
			}
			hitsChecked += len(want)
		}
	}
	t.Logf("%d hits compared, %d words with several patterns", hitsChecked, multi)
	if multi < 1000 || hitsChecked < 5000 {
		t.Errorf("not varied enough")
	}
}

func concreteFor(rng *rand.Rand, c *compiled) []byte {
	p := ApatPattern{c: c}
	return []byte(concrete(rng, p))
}
