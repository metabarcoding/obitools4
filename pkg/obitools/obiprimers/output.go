package obiprimers

import (
	"math"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiitercsv"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiprimer"
)

// Columns of the output, in the order of the ecoPrimers result table.
var csvColumns = []string{
	"index",              // 1  serial number
	"primer1", "primer2", // 2-3
	"tm1", "tm1_min", "tm2", "tm2_min", // 4-7 melting temperatures: perfect match, lowest over the examples
	"gc1", "gc2", // 8-9
	"good1", "good2", // 10 primer conserved in the examples and rare in the counterexamples
	"examples", "counterexamples", // 11-12 sequences amplified
	"yule",                                // 13
	"example_taxa", "counterexample_taxa", // 14-15 taxa amplified
	"bc",                                      // 16 coverage
	"identified_taxa",                         // 17
	"bs",                                      // 18 specificity
	"min_length", "max_length", "mean_length", // 19-21
}

// round returns x rounded to the given number of decimals; the CSV writer
// prints it with its shortest representation.
func round(x float64, decimals int) float64 {
	p := math.Pow10(decimals)
	return math.Round(x*p) / p
}

// pairRecord converts a pair to a CSV record. Unavailable values (NaN) are
// left out, which the writer prints as the NA value.
func pairRecord(index int, p obiprimer.PairStats) obiitercsv.CSVRecord {
	r := obiitercsv.CSVRecord{
		"index": index, "primer1": p.Primer1, "primer2": p.Primer2,
		"gc1": p.GC1, "gc2": p.GC2, "good1": p.Good1, "good2": p.Good2,
		"examples": p.Examples, "counterexamples": p.CounterExamples,
		"yule":         round(float64(p.Yule), 4),
		"example_taxa": p.ExampleTaxa, "counterexample_taxa": p.CounterTaxa,
		"bc": round(float64(p.Bc), 4), "identified_taxa": p.Identified, "bs": round(float64(p.Bs), 4),
		"min_length": p.MinLen, "max_length": p.MaxLen, "mean_length": round(p.MeanLen, 2),
	}
	for key, v := range map[string]float64{"tm1": p.Tm1, "tm1_min": p.MinTm1, "tm2": p.Tm2, "tm2_min": p.MinTm2} {
		if !math.IsNaN(v) {
			r[key] = round(v, 1)
		}
	}
	return r
}

// PairsToCSVIterator streams the pairs, in order, as CSV records.
func PairsToCSVIterator(pairs []obiprimer.PairStats) *obiitercsv.ICSVRecord {
	iterator := obiitercsv.NewICSVRecord()
	iterator.Add(1)
	for _, column := range csvColumns {
		iterator.AppendField(column)
	}

	go func() {
		defer iterator.Done()
		batchSize := obidefault.BatchSize()
		order := 0
		for start := 0; start < len(pairs) || order == 0; start += batchSize {
			end := min(start+batchSize, len(pairs))
			data := make([]obiitercsv.CSVRecord, 0, end-start)
			for i := start; i < end; i++ {
				data = append(data, pairRecord(i+1, pairs[i]))
			}
			iterator.Push(obiitercsv.MakeCSVRecordBatch("obiprimers", order, data))
			order++
		}
		iterator.Close()
	}()
	return iterator
}
