package obiformats

import "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"

type IBatchReader func(string, ...WithOption) (obiiter.IBioSequence, error)
