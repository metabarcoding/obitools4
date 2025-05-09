package obipcr

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	log "github.com/sirupsen/logrus"
)

// CLIPCR iterates over sequences provided by a obiseq.IBioSequenceBatch
// and returns an other obiseq.IBioSequenceBatch distributing
// obiseq.BioSequenceBatch containing the selected amplicon sequences.
func CLIPCR(iterator obiiter.IBioSequence) (obiiter.IBioSequence, error) {

	opts := make([]obiapat.WithOption, 0, 10)

	opts = append(opts,
		obiapat.OptionForwardPrimer(
			CLIForwardPrimer(),
			CLIAllowedMismatch(),
		),
		obiapat.OptionReversePrimer(
			CLIReversePrimer(),
			CLIAllowedMismatch(),
		),
		obiapat.OptionOnlyFullExtension(CLIOnlyFull()),
	)

	if CLIMinLength() > 0 {
		opts = append(opts, obiapat.OptionMinLength(CLIMinLength()))
	}

	if CLIWithExtension() {
		opts = append(opts, obiapat.OptionWithExtension(CLIExtension()))
	}

	opts = append(opts, obiapat.OptionMaxLength(CLIMaxLength()))

	if CLICircular() {
		opts = append(opts, obiapat.OptionCircular(CLICircular()))
	}

	worker := obiapat.PCRSliceWorker(opts...)

	if CLIFragmented() {
		frags := obiiter.IFragments(
			CLIMaxLength()*1000,
			CLIMaxLength()*100,
			CLIMaxLength()+max(len(CLIForwardPrimer()),
				len(CLIReversePrimer()))+min(len(CLIForwardPrimer()),
				len(CLIReversePrimer()))/2,
			100,
			obidefault.ParallelWorkers(),
		)
		log.Infof("Fragmenting sequence longer than %dbp into chuncks of %dbp",
			CLIMaxLength()*1000,
			CLIMaxLength()*100,
		)
		iterator = iterator.Pipe(frags)
	}

	return iterator.LimitMemory(0.5).MakeISliceWorker(worker, false, obidefault.ParallelWorkers()), nil
}
