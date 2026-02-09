package obik

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"github.com/DavidGamba/go-getoptions"
)

func runMv(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: obik mv [--set PATTERN]... [--force] <source_index> <dest_index>")
	}

	srcDir := args[0]
	destDir := args[1]

	ksg, err := obikmer.OpenKmerSetGroup(srcDir)
	if err != nil {
		return fmt.Errorf("failed to open source kmer index: %w", err)
	}

	// Resolve set patterns
	patterns := CLISetPatterns()
	var ids []string
	if len(patterns) > 0 {
		indices, err := ksg.MatchSetIDs(patterns)
		if err != nil {
			return err
		}
		if len(indices) == 0 {
			return fmt.Errorf("no sets match the given patterns")
		}
		ids = make([]string, len(indices))
		for i, idx := range indices {
			ids[i] = ksg.SetIDOf(idx)
		}
	} else {
		// Move all sets
		ids = ksg.SetsIDs()
	}

	log.Infof("Moving %d set(s) from %s to %s", len(ids), srcDir, destDir)

	// Copy first
	dest, err := ksg.CopySetsByIDTo(ids, destDir, CLIForce())
	if err != nil {
		return err
	}

	// Remove from source (in reverse order to avoid renumbering issues)
	for i := len(ids) - 1; i >= 0; i-- {
		if err := ksg.RemoveSetByID(ids[i]); err != nil {
			return fmt.Errorf("failed to remove set %q from source after copy: %w", ids[i], err)
		}
	}

	log.Infof("Destination now has %d set(s), source has %d set(s)", dest.Size(), ksg.Size())
	return nil
}
