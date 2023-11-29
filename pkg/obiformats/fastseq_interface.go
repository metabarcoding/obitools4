package obiformats

import "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"

type FormatHeader func(sequence *obiseq.BioSequence) string
