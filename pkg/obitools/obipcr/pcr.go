package obipcr

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidefault"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obidnapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	log "github.com/sirupsen/logrus"
)

// CLIPCR iterates over sequences provided by a obiseq.IBioSequenceBatch
// and returns an other obiseq.IBioSequenceBatch distributing
// obiseq.BioSequenceBatch containing the selected amplicon sequences.
func CLIPCR(iterator obiiter.IBioSequence) (obiiter.IBioSequence, error) {

	opts := []obidnapat.PCROption{
		obidnapat.PCRPrimers(
			CLIForwardPrimer(), CLIAllowedMismatch(),
			CLIReversePrimer(), CLIAllowedMismatch(),
		),
		obidnapat.PCRFullExtension(CLIOnlyFull()),
		obidnapat.PCRMinLength(CLIMinLength()),
		obidnapat.PCRMaxLength(CLIMaxLength()),
		obidnapat.PCRCircular(CLICircular()),
	}

	if CLIWithExtension() {
		opts = append(opts, obidnapat.PCRExtension(CLIExtension()))
	}

	pcr, err := obidnapat.NewPCR(opts...)
	if err != nil {
		return obiiter.NilIBioSequence, err
	}

	worker := pcr.SliceWorker()

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
