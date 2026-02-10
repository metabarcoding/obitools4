package obik

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"github.com/DavidGamba/go-getoptions"
)

func runRm(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: obik rm --set PATTERN [--set PATTERN]... <index_directory>")
	}

	patterns := CLISetPatterns()
	if len(patterns) == 0 {
		return fmt.Errorf("--set is required (specify which sets to remove)")
	}

	indexDir := args[0]

	ksg, err := obikmer.OpenKmerSetGroup(indexDir)
	if err != nil {
		return fmt.Errorf("failed to open kmer index: %w", err)
	}

	indices, err := ksg.MatchSetIDs(patterns)
	if err != nil {
		return err
	}
	if len(indices) == 0 {
		return fmt.Errorf("no sets match the given patterns")
	}

	// Collect IDs before removal (indices shift as we remove)
	ids := make([]string, len(indices))
	for i, idx := range indices {
		ids[i] = ksg.SetIDOf(idx)
	}

	log.Infof("Removing %d set(s) from %s", len(ids), indexDir)

	// Remove in reverse order to avoid renumbering issues
	for i := len(ids) - 1; i >= 0; i-- {
		if err := ksg.RemoveSetByID(ids[i]); err != nil {
			return fmt.Errorf("failed to remove set %q: %w", ids[i], err)
		}
		log.Infof("Removed set %q", ids[i])
	}

	log.Infof("Index now has %d set(s)", ksg.Size())
	return nil
}
