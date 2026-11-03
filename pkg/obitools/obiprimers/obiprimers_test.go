package obiprimers

import (
	"math"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiprimer"
)

func samplePairs() []obiprimer.PairStats {
	return []obiprimer.PairStats{
		{
			Primer1: "ACACCGCCCGTCACTCTC", Primer2: "CCAAGTGCACCTTCCGGT", Good1: true, Good2: false,
			Examples: 3837, CounterExamples: 2, Yule: 0.98249996, ExampleTaxa: 3805, CounterTaxa: 1,
			Bc: 0.98169, Identified: 2650, Bs: 0.69647, MinLen: 54, MaxLen: 140, MeanLen: 74.75123,
			Tm1: 62.46, Tm2: 60.74, MinTm1: 36.81, MinTm2: math.NaN(), GC1: 12, GC2: 11,
		},
		{Primer1: "AAAA", Primer2: "TTTT", Tm1: math.NaN(), Tm2: math.NaN(), MinTm1: math.NaN(), MinTm2: math.NaN()},
	}
}

func TestPairsToCSVIterator(t *testing.T) {
	it := PairsToCSVIterator(samplePairs())

	header := it.Header()
	if len(header) != len(csvColumns) || header[0] != "index" || header[1] != "primer1" {
		t.Fatalf("unexpected header %v", header)
	}

	var records []map[string]interface{}
	for it.Next() {
		batch := it.Get()
		for _, r := range batch.Slice() {
			records = append(records, r)
		}
	}
	if len(records) != 2 {
		t.Fatalf("%d records, want 2", len(records))
	}

	r := records[0]
	if r["index"] != 1 || r["primer1"] != "ACACCGCCCGTCACTCTC" || r["good1"] != true || r["good2"] != false {
		t.Errorf("unexpected record %v", r)
	}
	if r["yule"] != 0.9825 || r["bc"] != 0.9817 || r["bs"] != 0.6965 || r["mean_length"] != 74.75 {
		t.Errorf("indices are not rounded as expected: %v %v %v %v", r["yule"], r["bc"], r["bs"], r["mean_length"])
	}
	if r["tm1"] != 62.5 || r["tm2"] != 60.7 || r["tm1_min"] != 36.8 {
		t.Errorf("temperatures: %v %v %v", r["tm1"], r["tm2"], r["tm1_min"])
	}
	if _, present := r["tm2_min"]; present {
		t.Error("an unavailable temperature must be left out, to be written as NA")
	}
	if _, present := records[1]["tm1"]; present {
		t.Error("NaN temperatures must be left out")
	}
	if records[1]["index"] != 2 {
		t.Errorf("second pair index %v", records[1]["index"])
	}
}

func TestEmptyResultStillHasAHeader(t *testing.T) {
	it := PairsToCSVIterator(nil)
	if len(it.Header()) != len(csvColumns) {
		t.Fatalf("header %v", it.Header())
	}
	n := 0
	for it.Next() {
		batch := it.Get()
		n += len(batch.Slice())
	}
	if n != 0 {
		t.Errorf("%d records for no pairs", n)
	}
}

func TestEveryColumnIsProduced(t *testing.T) {
	full := pairRecord(1, samplePairs()[0])
	full["tm2_min"] = 1.0 // the only column left out for this pair
	for _, c := range csvColumns {
		if _, ok := full[c]; !ok {
			t.Errorf("column %q is never filled", c)
		}
	}
	if len(full) != len(csvColumns) {
		t.Errorf("record has %d fields for %d columns", len(full), len(csvColumns))
	}
}

func TestCLISearchParams(t *testing.T) {
	p, err := CLISearchParams()
	if err != nil {
		t.Fatal(err)
	}
	d := obiprimer.DefaultSearchParams()
	if p.PrimerLength != d.PrimerLength || p.MaxErrors != d.MaxErrors || p.MaxLength != d.MaxLength ||
		p.StrictQuorum != d.StrictQuorum || p.Sensitivity != d.Sensitivity ||
		p.FalsePositive != d.FalsePositive || p.MinSpecificity != d.MinSpecificity || p.Model == nil {
		t.Errorf("option defaults %+v differ from the ecoPrimers defaults %+v", p, d)
	}

	_saltMethod = "owczarzy"
	if p, err = CLISearchParams(); err != nil || p.Model == nil {
		t.Errorf("owczarzy must be accepted: %v", err)
	}
	_saltMethod = "nope"
	if _, err = CLISearchParams(); err == nil {
		t.Error("an unknown salt correction must be refused")
	}
	_saltMethod = "santalucia"
	_salt = -1
	if _, err = CLISearchParams(); err == nil {
		t.Error("a negative salt concentration must be refused")
	}
}
