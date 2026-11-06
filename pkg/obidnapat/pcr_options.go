package obidnapat

import "fmt"

// PCROptions are the parameters of an in silico PCR. They are set with the PCROption
// functions given to NewPCR.
type PCROptions struct {
	forwardPrimer, reversePrimer string
	forwardErrors, reverseErrors int
	minLength                    int  // shortest amplicon, the primers excluded; 0: no limit
	maxLength                    int  // longest amplicon, the primers excluded; 0 or less: no limit
	extension                    int  // bases kept around the primers; negative: the amplicon is what lies between them
	fullExtension                bool // with an extension, only the amplicons whose flanks are complete
	circular                     bool // the sequences are circular
}

// PCROption sets one parameter of an in silico PCR.
type PCROption func(*PCROptions)

// PCRPrimers sets the two primers and the number of errors allowed for each. They are
// searched with substitutions only, on both strands.
func PCRPrimers(forward string, forwardErrors int, reverse string, reverseErrors int) PCROption {
	return func(o *PCROptions) {
		o.forwardPrimer, o.forwardErrors = forward, forwardErrors
		o.reversePrimer, o.reverseErrors = reverse, reverseErrors
	}
}

// PCRMinLength sets the shortest amplicon kept, the primers excluded.
func PCRMinLength(length int) PCROption { return func(o *PCROptions) { o.minLength = length } }

// PCRMaxLength sets the longest amplicon kept, the primers excluded; 0 or less: no limit.
func PCRMaxLength(length int) PCROption { return func(o *PCROptions) { o.maxLength = length } }

// PCRExtension keeps this many bases of flank on each side of the amplicon, the primers
// included in it (0: the primers themselves). A negative value, the default, keeps only
// what lies between the primers. Flanks that go past the end of a linear sequence are cut
// unless PCRFullExtension is set. On a circular sequence they cross the junction, and an
// amplicon with its flanks longer than the sequence is not reported.
func PCRExtension(extension int) PCROption { return func(o *PCROptions) { o.extension = extension } }

// PCRFullExtension keeps, with an extension, only the amplicons whose flanks are complete.
func PCRFullExtension(full bool) PCROption { return func(o *PCROptions) { o.fullExtension = full } }

// PCRCircular tells that the sequences are circular: an amplicon can cross the junction.
func PCRCircular(circular bool) PCROption { return func(o *PCROptions) { o.circular = circular } }

func makePCROptions(setters []PCROption) PCROptions {
	o := PCROptions{extension: -1}
	for _, set := range setters {
		set(&o)
	}
	return o
}

func (o PCROptions) hasExtension() bool { return o.extension >= 0 }

func (o PCROptions) validate() error {
	if o.forwardPrimer == "" || o.reversePrimer == "" {
		return fmt.Errorf("obidnapat: a PCR needs a forward and a reverse primer")
	}
	return nil
}
