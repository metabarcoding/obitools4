package obiformats

import "git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"

func ReadEmptyFile(options ...WithOption) (obiiter.IBioSequence, error) {
	out := obiiter.MakeIBioSequence()
	out.Close()
	return out, nil
}
