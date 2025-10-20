package obilowmask

import (
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"

	log "github.com/sirupsen/logrus"
)

var __kmer_size__ = 31
var __level_max__ = 6
var __threshold__ = 0.5
var __split_mode__ = false
var __mask__ = "."

func LowMaskOptionSet(options *getoptions.GetOpt) {

	options.IntVar(&__kmer_size__, "kmer-size", __kmer_size__,
		options.Description("Size of the kmer considered to estimate entropy."),
	)

	options.IntVar(&__level_max__, "entropy_size", __level_max__,
		options.Description("Maximum word size considered for entropy estimate"),
	)

	options.Float64Var(&__threshold__, "threshold", __threshold__,
		options.Description("entropy theshold used to mask a kmer"),
	)

	options.BoolVar(&__split_mode__, "--split-mode", __split_mode__,
		options.Description("in split mode, input sequences are splitted to remove masked regions"),
	)

	options.StringVar(&__mask__, "--masking-char", __mask__,
		options.Description("Character used to mask low complexity region"),
	)
}

func OptionSet(options *getoptions.GetOpt) {
	LowMaskOptionSet(options)
	obiconvert.InputOptionSet(options)
}

func CLIKmerSize() int {
	return __kmer_size__
}

func CLILevelMax() int {
	return __level_max__
}

func CLIThreshold() float64 {
	return __threshold__
}

func CLIMaskingMode() MaskingMode {
	if __split_mode__ {
		return Split
	} else {
		return Mask
	}
}

func CLIMaskingChar() byte {
	mask := strings.TrimSpace(__mask__)
	if len(mask) != 1 {
		log.Fatalf("--masking-char option accept a single character, not %s", mask)
	}
	return []byte(mask)[0]
}
