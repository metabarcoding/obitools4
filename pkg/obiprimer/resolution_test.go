package obiprimer

import (
	"math/rand"
	"testing"
)

// rec builds an amplicon record: barcode, taxon (-1: none at the rank), role.
func rec(barcode uint64, taxon int32, role Role) ampRecord {
	return ampRecord{hash: barcode, taxon: taxon, role: role}
}

func TestResolutionUserExample(t *testing.T) {
	// A has the barcodes A1 and A2, B has A2 and B1. A2 is shared: neither A
	// nor B is identified, but A1 and B1 lead to one species each.
	const A, B = 0, 1
	const A1, A2, B1 = 11, 12, 21
	res := resolve([]ampRecord{
		rec(A1, A, Example), rec(A2, A, Example),
		rec(A2, B, Example), rec(B1, B, Example),
	})
	want := resolution{targetTaxa: 2, identified: 0, barcodes: 3, discriminant: 2}
	if res != want {
		t.Fatalf("got %+v, want %+v", res, want)
	}
}

func TestResolutionAllUnique(t *testing.T) {
	res := resolve([]ampRecord{rec(1, 0, Example), rec(2, 1, Example), rec(3, 2, Example)})
	if res.identified != 3 || res.discriminant != 3 || res.barcodes != 3 || res.targetTaxa != 3 {
		t.Errorf("all barcodes distinct: %+v", res)
	}
}

func TestResolutionSameTaxonSharesItsBarcodeFreely(t *testing.T) {
	// several sequences of one species with the same barcode are not ambiguous,
	// and several barcodes for one species are fine when none is shared
	res := resolve([]ampRecord{
		rec(1, 0, Example), rec(1, 0, Example), rec(1, 0, Example),
		rec(2, 0, Example),
	})
	if res.identified != 1 || res.barcodes != 2 || res.discriminant != 2 {
		t.Errorf("%+v", res)
	}
}

// The barcode of a target that is unique among the targets but shared with a
// non-target, or with a counterexample, does not identify it.
func TestResolutionNonTargetsCount(t *testing.T) {
	for _, role := range []Role{Other, CounterExample} {
		res := resolve([]ampRecord{
			rec(1, 0, Example), // unique among the targets...
			rec(1, 5, role),    // ...but shared with a non-target taxon
			rec(2, 1, Example), // identified
			rec(3, 6, role),    // a barcode of non-targets only
		})
		if res.targetTaxa != 2 || res.identified != 1 {
			t.Errorf("role %d: %+v, want 2 target taxa and 1 identified", role, res)
		}
		// barcodes carried by the targets: 1 (ambiguous) and 2; barcode 3 leads to a non-target
		if res.barcodes != 2 || res.discriminant != 1 {
			t.Errorf("role %d: Bd counts %d/%d, want 1/2", role, res.discriminant, res.barcodes)
		}
	}
}

func TestResolutionAmplifiedTaxaCounts(t *testing.T) {
	res := resolve([]ampRecord{
		rec(1, 0, Example), rec(2, 1, CounterExample), rec(3, 1, CounterExample),
		rec(4, 2, CounterExample), rec(5, 3, Other),
	})
	if res.targetTaxa != 1 || res.counterTaxa != 2 {
		t.Errorf("target taxa %d (want 1), counterexample taxa %d (want 2: others are not counted)",
			res.targetTaxa, res.counterTaxa)
	}
}

// A sequence with no taxon at the rank is an anonymous taxon, distinct from all
// the others: a barcode shared with it is ambiguous.
func TestResolutionAnonymousTaxon(t *testing.T) {
	res := resolve([]ampRecord{
		rec(1, 0, Example), rec(1, -1, Other), // shared with an unknown: ambiguous
		rec(2, 1, Example),
	})
	if res.identified != 1 || res.barcodes != 2 || res.discriminant != 1 {
		t.Errorf("target sharing with an anonymous sequence: %+v", res)
	}
	// two anonymous sequences with the same barcode never count: no target carries it
	res = resolve([]ampRecord{rec(7, -1, Other), rec(7, -1, Other), rec(8, 0, Example)})
	if res.barcodes != 1 || res.discriminant != 1 || res.identified != 1 {
		t.Errorf("anonymous barcodes alone: %+v", res)
	}
	// a target with no taxon is not a target taxon, and is anonymous too
	res = resolve([]ampRecord{rec(1, -1, Example), rec(1, 0, Example), rec(2, 0, Example)})
	if res.targetTaxa != 1 || res.identified != 0 {
		t.Errorf("an anonymous target sequence sharing a barcode must make the taxon ambiguous: %+v", res)
	}
}

func TestResolutionSeveralPartsAndOrder(t *testing.T) {
	// the amplicons come as one slice per worker, in any order
	a := []ampRecord{rec(1, 0, Example), rec(2, 1, Example)}
	b := []ampRecord{rec(2, 2, Other), rec(3, 1, Example)}
	if res := resolve(a, b); res.targetTaxa != 2 || res.identified != 1 {
		t.Errorf("taxon 1 shares barcode 2 with taxon 2: %+v", res)
	}
	if res := resolve(b, a); res.targetTaxa != 2 || res.identified != 1 {
		t.Errorf("order must not matter: %+v", res)
	}
}

// ecoPrimers' lenient rule on the user's example: A and B each own one barcode,
// so both count (the new rule counts none), and the bound is what lets a pair
// be rejected early.
func TestLenientIdentifiedUserExample(t *testing.T) {
	recs := []ampRecord{
		rec(11, 0, Example), rec(12, 0, Example),
		rec(12, 1, Example), rec(21, 1, Example),
	}
	if targets, identified := lenientIdentified(recs); targets != 2 || identified != 2 {
		t.Errorf("lenient rule: %d of %d, want 2 of 2", identified, targets)
	}
	if res := resolve(recs); res.identified != 0 {
		t.Errorf("strict rule: %d identified, want 0", res.identified)
	}
}

// The pruning of phase 2 is only exact if the bound never falls under the
// true count, with the same number of target taxa.
func TestLenientIdentifiedIsAnUpperBound(t *testing.T) {
	rng := rand.New(rand.NewSource(3))
	for trial := 0; trial < 3000; trial++ {
		var recs []ampRecord
		n := 1 + rng.Intn(40)
		for range n {
			role := []Role{Example, Example, Example, CounterExample, Other}[rng.Intn(5)]
			recs = append(recs, rec(uint64(rng.Intn(12)), int32(rng.Intn(8))-1, role)) // taxon -1: none
		}
		// split in a few parts, as the workers do
		cut := rng.Intn(len(recs) + 1)
		res := resolve(recs[:cut], recs[cut:])
		targets, lenient := lenientIdentified(recs[:cut], recs[cut:])
		if targets != res.targetTaxa {
			t.Fatalf("trial %d: %d target taxa for the bound, %d for resolve", trial, targets, res.targetTaxa)
		}
		if res.identified > lenient {
			t.Fatalf("trial %d: identified %d exceeds the bound %d for %v", trial, res.identified, lenient, recs)
		}
		if res.discriminant > res.barcodes || res.identified > res.targetTaxa {
			t.Fatalf("trial %d: inconsistent %+v", trial, res)
		}
	}
}
