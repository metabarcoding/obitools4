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

// resolver computes the indices of pairs one after the other, reusing its maps
// (a pair needs a few thousand entries, and there are thousands of pairs).
type resolver struct {
	state    map[uint64]barcodeState
	targets  map[int32]struct{}
	counters map[int32]struct{}
	bad      map[int32]struct{}
	owner    map[uint64]int32
}

func newResolver() *resolver {
	return &resolver{
		state:    map[uint64]barcodeState{},
		targets:  map[int32]struct{}{},
		counters: map[int32]struct{}{},
		bad:      map[int32]struct{}{},
		owner:    map[uint64]int32{},
	}
}

// resolve computes the resolution of one pair from its amplicons, in any order.
func resolve(a ampSlice) resolution { return newResolver().resolve(a) }

func (rs *resolver) resolve(a ampSlice) resolution {
	var res resolution
	clear(rs.state)
	clear(rs.targets)
	clear(rs.counters)
	clear(rs.bad)

	for i, h := range a.hash {
		taxon, role := unpackTaxon(a.tax[i])
		st := rs.state[h]
		switch {
		case taxon < 0:
			st.flags |= barcodeAnonymous
		case st.taxon == 0:
			st.taxon = taxon + 1
		case st.taxon != taxon+1:
			st.flags |= barcodeAmbiguous
		}
		if taxon >= 0 {
			switch role {
			case Example:
				st.flags |= barcodeTarget
				rs.targets[taxon] = struct{}{}
			case CounterExample:
				rs.counters[taxon] = struct{}{}
			}
		}
		rs.state[h] = st
	}

	for i, h := range a.hash {
		taxon, role := unpackTaxon(a.tax[i])
		if role == Example && taxon >= 0 && rs.state[h].ambiguous() {
			rs.bad[taxon] = struct{}{}
		}
	}

	for _, st := range rs.state {
		if st.flags&barcodeTarget != 0 {
			res.barcodes++
			if !st.ambiguous() {
				res.discriminant++
			}
		}
	}
	res.targetTaxa, res.counterTaxa = len(rs.targets), len(rs.counters)
	res.identified = len(rs.targets) - len(rs.bad)
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
func lenientIdentified(a ampSlice) (targetTaxa, identified int) {
	return newResolver().lenientIdentified(a)
}

func (rs *resolver) lenientIdentified(a ampSlice) (targetTaxa, identified int) {
	clear(rs.owner) // barcode -> its taxon, or -1 when shared
	clear(rs.targets)
	for i, h := range a.hash {
		taxon, role := unpackTaxon(a.tax[i])
		if role != Example || taxon < 0 {
			continue
		}
		rs.targets[taxon] = struct{}{}
		if prev, seen := rs.owner[h]; !seen {
			rs.owner[h] = taxon
		} else if prev != taxon {
			rs.owner[h] = -1
		}
	}
	clear(rs.bad) // here: the taxa owning a barcode
	for _, t := range rs.owner {
		if t >= 0 {
			rs.bad[t] = struct{}{}
		}
	}
	return len(rs.targets), len(rs.bad)
}
