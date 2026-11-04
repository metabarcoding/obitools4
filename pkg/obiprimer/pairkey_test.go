package obiprimer

import "testing"

// A pair of sites and the same two sites read on the complementary strand are
// one amplicon, so they must give one pair key. The mirror of an amplicon
// between the sites at pu < pd of a sequence of length n is the amplicon
// between the sites at n-lp-pd < n-lp-pu of its reverse complement, each site
// now reading the other way round.
func TestPairKeyIsStrandInvariant(t *testing.T) {
	const lp, n, pu, pd = 18, 300, 100, 160
	ps := &primerSet{primers: []Primer{
		{Word: 5, Good: true},
		{Word: 9, Good: true},
	}}
	p := PairParams{PrimerLength: lp, MinLength: 0, MaxLength: 200}

	keyOf := func(ms []seqMatch) []pairKey {
		var keys []pairKey
		ps.forEachOccurrenceIn(nil, ms, p, p.MinLength, p.MaxLength, func(o occurrence) { keys = append(keys, o.key) })
		return keys
	}

	for upPrimer := int32(0); upPrimer < 2; upPrimer++ {
		for downPrimer := int32(0); downPrimer < 2; downPrimer++ {
			for _, ru := range []bool{false, true} {
				for _, rd := range []bool{false, true} {
					here := []seqMatch{newSeqMatch(upPrimer, pu, ru), newSeqMatch(downPrimer, pd, rd)}
					mirror := []seqMatch{
						newSeqMatch(downPrimer, n-lp-pd, !rd),
						newSeqMatch(upPrimer, n-lp-pu, !ru),
					}
					k1, k2 := keyOf(here), keyOf(mirror)
					if len(k1) != 1 || len(k2) != 1 {
						t.Fatalf("primers %d,%d reverse %v,%v: %d and %d amplicons, want 1 and 1",
							upPrimer, downPrimer, ru, rd, len(k1), len(k2))
					}
					if k1[0] != k2[0] {
						t.Errorf("primers %d,%d reverse %v,%v: key %+v on the strand, %+v on the other",
							upPrimer, downPrimer, ru, rd, k1[0], k2[0])
					}
				}
			}
		}
	}
}

func reverseComplementString(s string) string {
	comp := map[byte]byte{'a': 't', 'c': 'g', 'g': 'c', 't': 'a'}
	out := make([]byte, len(s))
	for i := range s {
		out[len(s)-1-i] = comp[s[i]]
	}
	return string(out)
}

// With the same primer P at both ends, the pair (P, rc(P)) is its own mirror:
// P BARCODE rc(P) read on the other strand is P rc(BARCODE) rc(P). It is one
// barcode, so it must give one hash whatever the strand the sequence is given
// on; otherwise the two orientations count as two barcodes, a barcode shared
// by two taxa goes unseen, and Bs is overestimated.
func TestBarcodeHashOfAPairOfOnePrimer(t *testing.T) {
	const lp = 18
	primer := "acgtacgtacgtaaccgg"
	barcode := "aacctgcatgcaagtcgatcgggtacc"
	if barcode == reverseComplementString(barcode) {
		t.Fatal("the barcode must not be its own reverse complement")
	}
	seqA := primer + barcode + reverseComplementString(primer)
	seqB := reverseComplementString(seqA) // primer + rc(barcode) + rc(primer)

	ps := &primerSet{primers: []Primer{{Word: 7, Good: true}}}
	p := PairParams{PrimerLength: lp, MaxLength: 200}
	hashOf := func(seq string) uint64 {
		// the primer reads direct at the start and reversed at the end, in both sequences
		ms := []seqMatch{newSeqMatch(0, 0, false), newSeqMatch(0, uint32(len(seq)-lp), true)}
		var hashes []uint64
		ps.forEachOccurrenceIn(nil, ms, p, 0, p.MaxLength, func(o occurrence) {
			hashes = append(hashes, o.barcodeHash([]byte(seq), lp))
		})
		if len(hashes) != 1 {
			t.Fatalf("%d amplicons, want 1", len(hashes))
		}
		return hashes[0]
	}
	if a, b := hashOf(seqA), hashOf(seqB); a != b {
		t.Errorf("the same barcode on the two strands hashes to %x and %x", a, b)
	}
}

// The barcode of any other pair is read from its first primer, so the two
// strands already agree; the hash must not depend on the strand there either.
func TestBarcodeHashOfAPairOfTwoPrimers(t *testing.T) {
	const lp = 18
	f, r := "acgtacgtacgtaaccgg", "ttgcaagctagctagcta"
	barcode := "aacctgcatgcaagtcgatcgggtacc"
	seqA := f + barcode + reverseComplementString(r)
	seqB := reverseComplementString(seqA)

	// canonical order: the primer 0 is f, the primer 1 is r
	ps := &primerSet{primers: []Primer{{Word: 3, Good: true}, {Word: 9, Good: true}}}
	p := PairParams{PrimerLength: lp, MaxLength: 200}
	hashOf := func(ms []seqMatch, seq string) uint64 {
		var hashes []uint64
		ps.forEachOccurrenceIn(nil, ms, p, 0, p.MaxLength, func(o occurrence) {
			hashes = append(hashes, o.barcodeHash([]byte(seq), lp))
		})
		if len(hashes) != 1 {
			t.Fatalf("%d amplicons, want 1", len(hashes))
		}
		return hashes[0]
	}
	n := uint32(len(seqA))
	// on seqA: f direct then r reversed; on seqB the sites are mirrored
	a := hashOf([]seqMatch{newSeqMatch(0, 0, false), newSeqMatch(1, n-lp, true)}, seqA)
	b := hashOf([]seqMatch{newSeqMatch(1, 0, false), newSeqMatch(0, n-lp, true)}, seqB)
	if a != b {
		t.Errorf("the same barcode on the two strands hashes to %x and %x", a, b)
	}
}
