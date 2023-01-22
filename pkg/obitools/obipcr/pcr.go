package obipcr

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiapat"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obioptions"
)

// PCR iterates over sequences provided by a obiseq.IBioSequenceBatch
// and returns an other obiseq.IBioSequenceBatch distributing
// obiseq.BioSequenceBatch containing the selected amplicon sequences.
func PCR(iterator obiiter.IBioSequence) (obiiter.IBioSequence, error) {

	opts := make([]obiapat.WithOption, 0, 10)

	opts = append(opts,
		obiapat.OptionForwardPrimer(
			ForwardPrimer(),
			AllowedMismatch(),
		),
		obiapat.OptionReversePrimer(
			ReversePrimer(),
			AllowedMismatch(),
		),
	)

	if MinLength() > 0 {
		opts = append(opts, obiapat.OptionMinLength(MinLength()))
	}

	if MaxLength() > 0 {
		opts = append(opts, obiapat.OptionMaxLength(MaxLength()))
	}

	if Circular() {
		opts = append(opts, obiapat.OptionCircular(Circular()))
	}

	worker := obiapat.PCRSliceWorker(opts...)

	return iterator.MakeISliceWorker(worker, obioptions.CLIParallelWorkers(), 0), nil
}
