package obimicrosat

import (
	"fmt"
	"sort"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"github.com/dlclark/regexp2"
)

// MakeMicrosatWorker creates a SeqWorker that finds microsatellite regions in a BioSequence.
//
// The function takes three integer parameters: minLength, maxLength, and minUnits. minLength specifies the
// minimum length of the microsatellite region, maxLength specifies the maximum length, and minUnits specifies
// the minimum number of repeating units. The function returns an obiseq.SeqWorker, which is a Go function that
// takes a BioSequence as input and returns a BioSequenceSlice and an error. The SeqWorker performs the following
// steps:
// 1. It defines two helper functions: min_unit and normalizedUnit.
// 2. It defines a regular expression pattern based on the input parameters.
// 3. It defines the SeqWorker function w.
// 4. The w function searches for a match of the regular expression pattern in the sequence string.
// 5. If no match is found, it returns an empty BioSequenceSlice and nil error.
// 6. It calculates the length of the matching unit.
// 7. It checks if the unit length is less than minLength.
// 8. It creates a new regular expression pattern based on the unit length.
// 9. It extracts the matching unit from the sequence string.
// 10. It sets various attributes on the sequence.
// 11. It returns a BioSequenceSlice containing the original sequence and nil error.
//
// Parameters:
// - minLength: the minimum length of the microsatellite region.
// - maxLength: the maximum length of the microsatellite region.
// - minUnits: the minimum number of repeating units.
//
// Return type:
// - obiseq.SeqWorker: a Go function that takes a BioSequence as input and returns a BioSequenceSlice and an error.
func MakeMicrosatWorker(minUnitLength, maxUnitLength, minUnits, minLength, minflankLength int, reoriented bool) obiseq.SeqWorker {

	min_unit := func(microsat string) int {
		for i := 1; i < len(microsat); i++ {
			s1 := microsat[0 : len(microsat)-i]
			s2 := microsat[i:]

			if s1 == s2 {
				return i
			}
		}

		return 0
	}

	normalizedUnit := func(unit string) (string, bool) {
		all := make([]struct {
			unit   string
			direct bool
		}, 0, len(unit)*2)

		for i := 0; i < len(unit); i++ {
			rotate := unit[i:] + unit[:i]
			revcomp_rotate := obiseq.NewBioSequence("", []byte(rotate), "").ReverseComplement(true).String()
			all = append(all, struct {
				unit   string
				direct bool
			}{unit: rotate, direct: true},
				struct {
					unit   string
					direct bool
				}{unit: revcomp_rotate, direct: false})
		}

		sort.Slice(all, func(i, j int) bool {
			return all[i].unit < all[j].unit
		})

		return all[0].unit, all[0].direct
	}

	build_regexp := func(minLength, maxLength, minUnits int) *regexp2.Regexp {
		return regexp2.MustCompile(
			fmt.Sprintf("([acgt]{%d,%d})\\1{%d,}",
				minLength,
				maxLength,
				minUnits-1,
			),
			regexp2.RE2)
	}

	regexp := build_regexp(minUnitLength, maxUnitLength, minUnits)

	w := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {

		match, _ := regexp.FindStringMatch(sequence.String())

		if match == nil {
			return obiseq.BioSequenceSlice{}, nil
		}

		unit_length := min_unit(match.String())

		if unit_length < minUnitLength {
			return obiseq.BioSequenceSlice{}, nil
		}

		pattern := build_regexp(unit_length, unit_length, minUnits)

		match, _ = pattern.FindStringMatch(sequence.String())

		if match.Length < minLength {
			return obiseq.BioSequenceSlice{}, nil
		}

		unit := match.String()[0:unit_length]
		normalized, direct := normalizedUnit(unit)
		matchFrom := match.Index

		if !direct && reoriented {
			sequence = sequence.ReverseComplement(true)
			sequence.SetId(sequence.Id() + "_cmp")
			matchFrom = sequence.Len() - match.Index - match.Length
		}

		matchTo := matchFrom + match.Length

		microsat := sequence.String()[matchFrom:matchTo]
		unit = microsat[0:unit_length]

		left := sequence.String()[0:matchFrom]
		right := sequence.String()[matchTo:]

		if len(left) < minflankLength || len(right) < minflankLength {
			return obiseq.BioSequenceSlice{}, nil
		}

		sequence.SetAttribute("microsat_unit_length", unit_length)
		sequence.SetAttribute("microsat_unit_count", match.Length/unit_length)
		sequence.SetAttribute("seq_length", sequence.Len())
		sequence.SetAttribute("microsat", microsat)
		sequence.SetAttribute("microsat_from", matchFrom+1)
		sequence.SetAttribute("microsat_to", matchTo)

		sequence.SetAttribute("microsat_unit", unit)
		sequence.SetAttribute("microsat_unit_normalized", normalized)

		sequence.SetAttribute("microsat_unit_orientation", map[bool]string{true: "direct", false: "reverse"}[direct])

		sequence.SetAttribute("microsat_left", left)
		sequence.SetAttribute("microsat_right", right)

		return obiseq.BioSequenceSlice{sequence}, nil
	}

	return obiseq.SeqWorker(w)
}

// CLIAnnotateMicrosat is a function that annotates microsatellites in a given sequence.
//
// It takes an iterator of type `obiiter.IBioSequence` as a parameter.
// The function returns an iterator of type `obiiter.IBioSequence`.
func CLIAnnotateMicrosat(iterator obiiter.IBioSequence) obiiter.IBioSequence {
	var newIter obiiter.IBioSequence

	worker := MakeMicrosatWorker(CLIMinUnitLength(),
		CLIMaxUnitLength(),
		CLIMinUnitCount(),
		CLIMinLength(),
		CLIMinFlankLength(),
		CLIReoriented())

	newIter = iterator.MakeIWorker(worker, false, obioptions.CLIParallelWorkers())

	return newIter.FilterEmpty()

}
