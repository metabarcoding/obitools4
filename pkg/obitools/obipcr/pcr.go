package obipcr

import (
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiapat"
	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"
)

func PCR(iterator obiseq.IBioSequenceBatch) (obiseq.IBioSequence, error) {

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

	return iterator.MakeISliceWorker(worker).IBioSequence(), nil
}
