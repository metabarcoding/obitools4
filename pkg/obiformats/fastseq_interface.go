package obiformats

import "git.metabarcoding.org/lecasofts/go/oa2/pkg/obiseq"

type FormatHeader func(sequence obiseq.BioSequence) string
