package obik

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/DavidGamba/go-getoptions"
)

// runSuper implements the "obik super" subcommand.
// It extracts super k-mers from DNA sequences.
func runSuper(ctx context.Context, opt *getoptions.GetOpt, args []string) error {
	k := CLIKmerSize()
	m := CLIMinimizerSize()

	if k < 2 || k > 31 {
		return fmt.Errorf("invalid k-mer size: %d (must be between 2 and 31)", k)
	}

	if m < 1 || m >= k {
		return fmt.Errorf("invalid parameters: minimizer size (%d) must be between 1 and k-1 (%d)", m, k-1)
	}

	log.Printf("Extracting super k-mers with k=%d, m=%d", k, m)

	sequences, err := obiconvert.CLIReadBioSequences(args...)
	if err != nil {
		return fmt.Errorf("failed to open sequence files: %w", err)
	}

	worker := obikmer.SuperKmerWorker(k, m)

	superkmers := sequences.MakeIWorker(
		worker,
		false,
		obidefault.ParallelWorkers(),
	)

	obiconvert.CLIWriteBioSequences(superkmers, true)
	obiutils.WaitForLastPipe()

	return nil
}
