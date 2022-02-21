package obiformats

import "git.metabarcoding.org/lecasofts/go/obitools/pkg/obiseq"

type FormatHeader func(sequence *obiseq.BioSequence) string
