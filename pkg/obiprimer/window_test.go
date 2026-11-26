package obiprimer

import "testing"

func TestPairParamsWindow(t *testing.T) {
	p := PairParams{MinLength: 20, MaxLength: 100}

	// defaults: the targets keep ]20,100[, the counterexamples have no lower bound
	if lo, hi := p.window(Example); lo != 20 || hi != 100 {
		t.Errorf("example window ]%d,%d[", lo, hi)
	}
	if lo, hi := p.window(Other); lo != 20 || hi != 100 {
		t.Errorf("other window ]%d,%d[", lo, hi)
	}
	if lo, hi := p.window(CounterExample); lo != 0 || hi != 100 {
		t.Errorf("counterexample window ]%d,%d[", lo, hi)
	}

	p.MinLengthNeg, p.MaxLengthNeg = 5, 1000
	if lo, hi := p.window(CounterExample); lo != 5 || hi != 1000 {
		t.Errorf("counterexample window ]%d,%d[", lo, hi)
	}

	// a counterexample maximum below -L is raised to it
	p.MaxLengthNeg = 50
	if _, hi := p.window(CounterExample); hi != 100 {
		t.Errorf("counterexample maximum %d, want 100", hi)
	}
}
