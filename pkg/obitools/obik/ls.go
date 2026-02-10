package obik

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"github.com/DavidGamba/go-getoptions"
	"gopkg.in/yaml.v3"
)

type setEntry struct {
	Index int    `json:"index" yaml:"index"`
	ID    string `json:"id" yaml:"id"`
	Count uint64 `json:"count" yaml:"count"`
}

func runLs(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: obik ls [options] <index_directory>")
	}

	ksg, err := obikmer.OpenKmerSetGroup(args[0])
	if err != nil {
		return fmt.Errorf("failed to open kmer index: %w", err)
	}

	// Determine which sets to show
	patterns := CLISetPatterns()
	var indices []int
	if len(patterns) > 0 {
		indices, err = ksg.MatchSetIDs(patterns)
		if err != nil {
			return err
		}
	} else {
		indices = make([]int, ksg.Size())
		for i := range indices {
			indices[i] = i
		}
	}

	entries := make([]setEntry, len(indices))
	for i, idx := range indices {
		entries[i] = setEntry{
			Index: idx,
			ID:    ksg.SetIDOf(idx),
			Count: ksg.Len(idx),
		}
	}

	format := CLIOutFormat()
	switch format {
	case "json":
		return outputLsJSON(entries)
	case "yaml":
		return outputLsYAML(entries)
	case "csv":
		return outputLsCSV(entries)
	default:
		return outputLsCSV(entries)
	}
}

func outputLsCSV(entries []setEntry) error {
	fmt.Println("index,id,count")
	for _, e := range entries {
		// Escape commas in ID if needed
		id := e.ID
		if strings.ContainsAny(id, ",\"") {
			id = "\"" + strings.ReplaceAll(id, "\"", "\"\"") + "\""
		}
		fmt.Printf("%d,%s,%d\n", e.Index, id, e.Count)
	}
	return nil
}

func outputLsJSON(entries []setEntry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputLsYAML(entries []setEntry) error {
	data, err := yaml.Marshal(entries)
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}
