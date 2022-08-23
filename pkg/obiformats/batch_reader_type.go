package obiformats

import "git.metabarcoding.org/lecasofts/go/obitools/pkg/obiiter"

type IBatchReader func(string, ...WithOption) (obiiter.IBioSequenceBatch, error)