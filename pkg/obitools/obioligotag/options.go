// obioligotag function utility package.
//
// The obitools/obioligotag package contains every
// functions specificaly required by the obioligotag utility.
package obioligotag

import (
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"github.com/DavidGamba/go-getoptions"
)

var __oligo_size__ = 0
var __oligo_list__ = ""
var __family_size__ = 0
var __distance__ = 1
var __distance_type__ = "hamming"
var __alphabet__ = "ACGT"
var __gc_max__ = -1
var __accepted__ = ""
var __rejected__ = ""
var __homopolymer__ = -1
var __homopolymer_min__ = 0
var __timeout__ = 60

// OligoTagOptionSet registers the options specific to obioligotag.
func OligoTagOptionSet(options *getoptions.GetOpt) {
	options.IntVar(&__oligo_size__, "oligo-size", __oligo_size__,
		options.Alias("s"),
		options.ArgName("SIZE"),
		options.Description("Length of the oligonucleotides to generate (1 to 12). "+
			"Cannot be combined with --oligo-list."))

	options.StringVar(&__oligo_list__, "oligo-list", __oligo_list__,
		options.Alias("L"),
		options.ArgName("FILENAME"),
		options.Description("File of candidate oligonucleotides, in any sequence file format "+
			"or as a text file with one oligonucleotide per line (format detected automatically). "+
			"The family is selected among them. Cannot be combined with --oligo-size."))

	options.IntVar(&__family_size__, "family-size", __family_size__,
		options.Alias("f"),
		options.ArgName("SIZE"),
		options.Description("Minimum number of oligonucleotides of the family to build."))

	options.IntVar(&__distance__, "distance", __distance__,
		options.Alias("d"),
		options.ArgName("DISTANCE"),
		options.Description("Minimum distance between any two oligonucleotides of the family."))

	options.StringVar(&__distance_type__, "distance-type", __distance_type__,
		options.ArgName("TYPE"),
		options.Description("Distance between oligonucleotides: 'hamming' (number of differing positions) "+
			"or 'lcs' (alignment length minus the length of the longest common subsequence)."))

	options.StringVar(&__alphabet__, "alphabet", __alphabet__,
		options.ArgName("LETTERS"),
		options.Description("Nucleotides allowed in the oligonucleotides, as IUPAC codes "+
			"(for instance ACT, or H, to forbid G)."))

	options.IntVar(&__gc_max__, "gc-max", __gc_max__,
		options.Alias("g"),
		options.ArgName("COUNT"),
		options.Description("Maximum number of G or C per oligonucleotide."))

	options.StringVar(&__accepted__, "accepted", __accepted__,
		options.Alias("a"),
		options.ArgName("IUPAC_PATTERN"),
		options.Description("Selected oligonucleotides must match this IUPAC pattern "+
			"(of the same length as the oligonucleotides)."))

	options.StringVar(&__rejected__, "rejected", __rejected__,
		options.Alias("r"),
		options.ArgName("IUPAC_PATTERN"),
		options.Description("Selected oligonucleotides must not match this IUPAC pattern "+
			"(of the same length as the oligonucleotides)."))

	options.IntVar(&__homopolymer__, "homopolymer", __homopolymer__,
		options.Alias("p"),
		options.ArgName("LENGTH"),
		options.Description("Excludes the oligonucleotides containing a homopolymer longer than LENGTH."))

	options.IntVar(&__homopolymer_min__, "homopolymer-min", __homopolymer_min__,
		options.Alias("P"),
		options.ArgName("LENGTH"),
		options.Description("Keeps only the oligonucleotides containing a homopolymer of at least LENGTH."))

	options.IntVar(&__timeout__, "timeout", __timeout__,
		options.Alias("T"),
		options.ArgName("SECONDS"),
		options.Description("Time limit, in seconds, for finding a family of the required size "+
			"(0 for no limit). When it is reached, the best family found is returned with a warning."))
}

// OptionSet registers every option of obioligotag.
func OptionSet(options *getoptions.GetOpt) {
	OligoTagOptionSet(options)
	obiconvert.InputOptionSet(options)
	obiconvert.OutputOptionSet(options)
}

// CLIOligoSize returns the size of the oligonucleotides to generate.
func CLIOligoSize() int { return __oligo_size__ }

// CLIOligoList returns the name of the file of candidate oligonucleotides.
func CLIOligoList() string { return __oligo_list__ }

// CLIFamilySize returns the minimum size of the family to build.
func CLIFamilySize() int { return __family_size__ }

// CLIDistance returns the minimum distance between two oligonucleotides.
func CLIDistance() int { return __distance__ }

// CLIDistanceType returns the name of the distance used.
func CLIDistanceType() string { return __distance_type__ }

// CLIAlphabet returns the IUPAC codes of the allowed nucleotides.
func CLIAlphabet() string { return __alphabet__ }

// CLIGCMax returns the maximum number of G or C (negative if unlimited).
func CLIGCMax() int { return __gc_max__ }

// CLIAccepted returns the IUPAC pattern the oligonucleotides must match.
func CLIAccepted() string { return __accepted__ }

// CLIRejected returns the IUPAC pattern the oligonucleotides must not match.
func CLIRejected() string { return __rejected__ }

// CLIHomopolymerMax returns the longest homopolymer allowed (<= 0 if unlimited).
func CLIHomopolymerMax() int { return __homopolymer__ }

// CLIHomopolymerMin returns the length of the homopolymer required (<= 1 if none).
func CLIHomopolymerMin() int { return __homopolymer_min__ }

// CLITimeout returns the time limit in seconds (0 if none).
func CLITimeout() int { return __timeout__ }

// Setters, mainly for testing.

func SetOligoSize(v int)       { __oligo_size__ = v }
func SetOligoList(v string)    { __oligo_list__ = v }
func SetFamilySize(v int)      { __family_size__ = v }
func SetDistance(v int)        { __distance__ = v }
func SetDistanceType(v string) { __distance_type__ = v }
func SetAlphabet(v string)     { __alphabet__ = v }
func SetGCMax(v int)           { __gc_max__ = v }
func SetAccepted(v string)     { __accepted__ = v }
func SetRejected(v string)     { __rejected__ = v }
func SetHomopolymerMax(v int)  { __homopolymer__ = v }
func SetHomopolymerMin(v int)  { __homopolymer_min__ = v }
func SetTimeout(v int)         { __timeout__ = v }
