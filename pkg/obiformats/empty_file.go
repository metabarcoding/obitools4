package obiformats

import "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"

func ReadEmptyFile(options ...WithOption) (obiiter.IBioSequence, error) {
	out := obiiter.MakeIBioSequence()
	out.Close()
	return out, nil
}
