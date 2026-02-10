package obik

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

// runSpectrum implements the "obik spectrum" subcommand.
// It outputs k-mer frequency spectra as CSV with one column per set.
func runSpectrum(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: obik spectrum [options] <index_directory>")
	}

	ksg, err := obikmer.OpenKmerSetGroup(args[0])
	if err != nil {
		return fmt.Errorf("failed to open kmer index: %w", err)
	}

	// Determine which sets to include
	patterns := CLISetPatterns()
	var indices []int
	if len(patterns) > 0 {
		indices, err = ksg.MatchSetIDs(patterns)
		if err != nil {
			return fmt.Errorf("failed to match set patterns: %w", err)
		}
		if len(indices) == 0 {
			return fmt.Errorf("no sets match the given patterns")
		}
	} else {
		// All sets
		indices = make([]int, ksg.Size())
		for i := range indices {
			indices[i] = i
		}
	}

	// Read spectra for selected sets
	spectraMaps := make([]map[int]uint64, len(indices))
	maxFreq := 0
	for i, idx := range indices {
		spectrum, err := ksg.Spectrum(idx)
		if err != nil {
			return fmt.Errorf("failed to read spectrum for set %d: %w", idx, err)
		}
		if spectrum == nil {
			log.Warnf("No spectrum data for set %d (%s)", idx, ksg.SetIDOf(idx))
			spectraMaps[i] = make(map[int]uint64)
			continue
		}
		spectraMaps[i] = spectrum.ToMap()
		if mf := spectrum.MaxFrequency(); mf > maxFreq {
			maxFreq = mf
		}
	}

	if maxFreq == 0 {
		return fmt.Errorf("no spectrum data found in any selected set")
	}

	// Determine output destination
	outFile := obiconvert.CLIOutPutFileName()
	var w *csv.Writer
	if outFile == "" || outFile == "-" {
		w = csv.NewWriter(os.Stdout)
	} else {
		f, err := os.Create(outFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer f.Close()
		w = csv.NewWriter(f)
	}
	defer w.Flush()

	// Build header: frequency, set_id_1, set_id_2, ...
	header := make([]string, 1+len(indices))
	header[0] = "frequency"
	for i, idx := range indices {
		id := ksg.SetIDOf(idx)
		if id == "" {
			id = fmt.Sprintf("set_%d", idx)
		}
		header[i+1] = id
	}
	if err := w.Write(header); err != nil {
		return err
	}

	// Write rows for each frequency from 1 to maxFreq
	record := make([]string, 1+len(indices))
	for freq := 1; freq <= maxFreq; freq++ {
		record[0] = strconv.Itoa(freq)
		hasData := false
		for i := range indices {
			count := spectraMaps[i][freq]
			record[i+1] = strconv.FormatUint(count, 10)
			if count > 0 {
				hasData = true
			}
		}
		// Only write rows where at least one set has a non-zero count
		if hasData {
			if err := w.Write(record); err != nil {
				return err
			}
		}
	}

	return nil
}
