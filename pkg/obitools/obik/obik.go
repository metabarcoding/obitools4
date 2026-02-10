package obik

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

// OptionSet registers all obik subcommands on the root GetOpt.
func OptionSet(opt *getoptions.GetOpt) {
	// index: build or extend a kmer index from sequence files
	indexCmd := opt.NewCommand("index", "Build a disk-based kmer index from sequence files")
	obiconvert.InputOptionSet(indexCmd)
	obiconvert.OutputModeOptionSet(indexCmd, false)
	KmerIndexOptionSet(indexCmd)
	indexCmd.StringMapVar(&_setMetaTags, "tag", 1, 1,
		indexCmd.Alias("T"),
		indexCmd.ArgName("KEY=VALUE"),
		indexCmd.Description("Per-set metadata tag (repeatable)."))
	indexCmd.SetCommandFn(runIndex)

	// ls: list sets in a kmer index
	lsCmd := opt.NewCommand("ls", "List sets in a kmer index")
	OutputFormatOptionSet(lsCmd)
	SetSelectionOptionSet(lsCmd)
	lsCmd.SetCommandFn(runLs)

	// summary: detailed statistics
	summaryCmd := opt.NewCommand("summary", "Show detailed statistics of a kmer index")
	OutputFormatOptionSet(summaryCmd)
	summaryCmd.BoolVar(&_jaccard, "jaccard", false,
		summaryCmd.Description("Compute and display pairwise Jaccard distance matrix."))
	summaryCmd.SetCommandFn(runSummary)

	// cp: copy sets between indices
	cpCmd := opt.NewCommand("cp", "Copy sets between kmer indices")
	SetSelectionOptionSet(cpCmd)
	ForceOptionSet(cpCmd)
	cpCmd.SetCommandFn(runCp)

	// mv: move sets between indices
	mvCmd := opt.NewCommand("mv", "Move sets between kmer indices")
	SetSelectionOptionSet(mvCmd)
	ForceOptionSet(mvCmd)
	mvCmd.SetCommandFn(runMv)

	// rm: remove sets from an index
	rmCmd := opt.NewCommand("rm", "Remove sets from a kmer index")
	SetSelectionOptionSet(rmCmd)
	rmCmd.SetCommandFn(runRm)

	// spectrum: output k-mer frequency spectrum as CSV
	spectrumCmd := opt.NewCommand("spectrum", "Output k-mer frequency spectrum as CSV")
	SetSelectionOptionSet(spectrumCmd)
	obiconvert.OutputModeOptionSet(spectrumCmd, false)
	spectrumCmd.SetCommandFn(runSpectrum)

	// super: extract super k-mers from sequences
	superCmd := opt.NewCommand("super", "Extract super k-mers from sequence files")
	obiconvert.InputOptionSet(superCmd)
	obiconvert.OutputOptionSet(superCmd)
	SuperKmerOptionSet(superCmd)
	superCmd.SetCommandFn(runSuper)

	// lowmask: mask low-complexity regions
	lowmaskCmd := opt.NewCommand("lowmask", "Mask low-complexity regions in sequences using entropy")
	obiconvert.InputOptionSet(lowmaskCmd)
	obiconvert.OutputOptionSet(lowmaskCmd)
	LowMaskOptionSet(lowmaskCmd)
	lowmaskCmd.SetCommandFn(runLowmask)

	// match: annotate sequences with k-mer match positions from an index
	matchCmd := opt.NewCommand("match", "Annotate sequences with k-mer match positions from an index")
	IndexDirectoryOptionSet(matchCmd)
	obiconvert.InputOptionSet(matchCmd)
	obiconvert.OutputOptionSet(matchCmd)
	SetSelectionOptionSet(matchCmd)
	matchCmd.SetCommandFn(runMatch)

	// filter: filter an index to remove low-complexity k-mers
	filterCmd := opt.NewCommand("filter", "Filter a kmer index to remove low-complexity k-mers")
	obiconvert.OutputModeOptionSet(filterCmd, false)
	EntropyFilterOptionSet(filterCmd)
	SetSelectionOptionSet(filterCmd)
	filterCmd.SetCommandFn(runFilter)
}
