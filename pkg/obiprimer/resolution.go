package obiprimer

// The three quality indices of a pair, over the sequences it amplifies.
//
// Every sequence the pair amplifies gives a barcode (the amplicon between the
// primers); all of them are taken into account, whatever their role: examples
// (the targets), counterexamples and non-targets.
//
//   - Bc, coverage: target taxa amplified / target taxa.
//   - Bs, specificity: target taxa identified / target taxa amplified. A target
//     taxon is identified when ALL its barcodes belong to it alone: as soon as a
//     barcode is shared with another taxon, target or not, counterexample or
//     not, both taxa are badly identified.
//   - Bd, barcode discrimination: the share of the barcodes carried by the
//     targets that lead without ambiguity to one taxon.
//
// A sequence with no taxon at the chosen rank is an anonymous taxon, distinct
// from every other: a barcode it shares with a taxon is ambiguous.

// barcodeState is what is known of one barcode while the amplicons are scanned.
type barcodeState struct {
	taxon int32 // 1 + the first taxon met, 0 while none
	flags uint8
}

const (
	barcodeAmbiguous = 1 << iota // two distinct taxa carry it
	barcodeAnonymous             // a sequence with no taxon carries it
	barcodeTarget                // a target sequence with a taxon carries it
)

// ambiguous reports whether the barcode does not lead to a single taxon.
func (b barcodeState) ambiguous() bool {
	return b.flags&barcodeAmbiguous != 0 || (b.flags&barcodeAnonymous != 0 && b.taxon != 0)
}

// resolution is the outcome of the identification of the taxa of one pair.
type resolution struct {
	targetTaxa   int // target taxa amplified
	counterTaxa  int // counterexample taxa amplified
	identified   int // target taxa whose barcodes all belong to them alone
	barcodes     int // distinct barcodes carried by the targets
	discriminant int // of which lead to a single taxon
}

// resolve computes the resolution of one pair from its amplicons. The slices
// need not be sorted; they are only read.
func resolve(parts ...[]ampRecord) resolution {
	var res resolution

	state := map[uint64]barcodeState{}
	targets := map[int32]struct{}{}
	counters := map[int32]struct{}{}

	for _, part := range parts {
		for _, r := range part {
			st := state[r.hash]
			switch {
			case r.taxon < 0:
				st.flags |= barcodeAnonymous
			case st.taxon == 0:
				st.taxon = r.taxon + 1
			case st.taxon != r.taxon+1:
				st.flags |= barcodeAmbiguous
			}
			if r.taxon >= 0 {
				switch r.role {
				case Example:
					st.flags |= barcodeTarget
					targets[r.taxon] = struct{}{}
				case CounterExample:
					counters[r.taxon] = struct{}{}
				}
			}
			state[r.hash] = st
		}
	}

	bad := map[int32]struct{}{}
	for _, part := range parts {
		for _, r := range part {
			if r.role == Example && r.taxon >= 0 && state[r.hash].ambiguous() {
				bad[r.taxon] = struct{}{}
			}
		}
	}

	for _, st := range state {
		if st.flags&barcodeTarget != 0 {
			res.barcodes++
			if !st.ambiguous() {
				res.discriminant++
			}
		}
	}
	res.targetTaxa, res.counterTaxa = len(targets), len(counters)
	res.identified = len(targets) - len(bad)
	return res
}

// lenientIdentified is an upper bound of resolution.identified, computed on
// the target sequences alone and with the lenient rule of ecoPrimers (a taxon
// counts as soon as ONE of its barcodes belongs to it alone, among the targets).
//
// A taxon identified by resolve has all its barcodes alone among every amplified
// sequence, hence among the targets: it has at least one such barcode and is
// counted here too. So identified <= the second result, with the same number of
// target taxa, which lets a pair be rejected on the specificity threshold without
// gathering the amplicons of the non-target sequences.
func lenientIdentified(parts ...[]ampRecord) (targetTaxa, identified int) {
	owner := map[uint64]int32{} // barcode -> its taxon, or -1 when shared
	targets := map[int32]struct{}{}
	for _, part := range parts {
		for _, r := range part {
			if r.role != Example || r.taxon < 0 {
				continue
			}
			targets[r.taxon] = struct{}{}
			if prev, seen := owner[r.hash]; !seen {
				owner[r.hash] = r.taxon
			} else if prev != r.taxon {
				owner[r.hash] = -1
			}
		}
	}
	alone := map[int32]struct{}{}
	for _, t := range owner {
		if t >= 0 {
			alone[t] = struct{}{}
		}
	}
	return len(targets), len(alone)
}
