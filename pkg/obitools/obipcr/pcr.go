package obipcr

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiapat"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

// PCR iterates over sequences provided by a obiseq.IBioSequenceBatch
// and returns an other obiseq.IBioSequenceBatch distributing
// obiseq.BioSequenceBatch containing the selected amplicon sequences.
func PCR(iterator obiseq.IBioSequenceBatch) (obiseq.IBioSequenceBatch, error) {

	forward := ForwardPrimer()
	reverse := ReversePrimer()
	opts := make([]obiapat.WithOption, 0, 10)

	opts = append(opts, obiapat.OptionForwardError(AllowedMismatch()),
		obiapat.OptionReverseError(AllowedMismatch()))

	if MinLength() > 0 {
		opts = append(opts, obiapat.OptionMinLength(MinLength()))
	}

	if MaxLength() > 0 {
		opts = append(opts, obiapat.OptionMaxLength(MaxLength()))
	}

	if Circular() {
		opts = append(opts, obiapat.OptionCircular(Circular()))
	}

	worker := obiapat.PCRSliceWorker(forward, reverse, opts...)

	return iterator.MakeISliceWorker(worker), nil
}
