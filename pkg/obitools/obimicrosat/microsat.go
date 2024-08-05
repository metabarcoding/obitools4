package obimicrosat

import (
	"fmt"
	"sort"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"github.com/dlclark/regexp2"
)

func MakeMicrosatWorker(minLength, maxLength, minUnits int) obiseq.SeqWorker {

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

	normalizedUnit := func(unit string) string {
		all := make([]string, 0, len(unit)*2)

		for i := 0; i < len(unit); i++ {
			rotate := unit[i:] + unit[:i]
			revcomp_rotate := obiseq.NewBioSequence("", []byte(rotate), "").ReverseComplement(true).String()
			all = append(all, rotate, revcomp_rotate)
		}

		sort.Slice(all, func(i, j int) bool {
			return all[i] < all[j]
		})

		return all[0]
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

	regexp := build_regexp(minLength, maxLength, minUnits)

	w := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {

		match, _ := regexp.FindStringMatch(sequence.String())

		if match == nil {
			return obiseq.BioSequenceSlice{}, nil
		}

		unit_length := min_unit(match.String())

		if unit_length < minLength {
			return obiseq.BioSequenceSlice{}, nil
		}

		pattern := build_regexp(unit_length, unit_length, minUnits)

		match, _ = pattern.FindStringMatch(sequence.String())

		unit := match.String()[0:unit_length]

		sequence.SetAttribute("microsat_unit_length", unit_length)
		sequence.SetAttribute("microsat_unit_count", match.Length/unit_length)
		sequence.SetAttribute("seq_length", sequence.Len())
		sequence.SetAttribute("microsat", match.String())
		sequence.SetAttribute("microsat_from", match.Index)
		sequence.SetAttribute("microsat_to", match.Index+match.Length-1)

		sequence.SetAttribute("microsat_unit", unit)
		sequence.SetAttribute("microsat_unit_normalized", normalizedUnit(unit))

		sequence.SetAttribute("microsat_left", sequence.String()[0:match.Index])
		sequence.SetAttribute("microsat_right", sequence.String()[match.Index+match.Length:])

		return obiseq.BioSequenceSlice{sequence}, nil
	}

	return obiseq.SeqWorker(w)
}

func CLIAnnotateMicrosat(iterator obiiter.IBioSequence) obiiter.IBioSequence {
	var newIter obiiter.IBioSequence

	worker := MakeMicrosatWorker(CLIMinUnitLength(), CLIMaxUnitLength(), CLIMinUnitCount())

	newIter = iterator.MakeIWorker(worker, false, obioptions.CLIParallelWorkers())

	return newIter.FilterEmpty()

}
