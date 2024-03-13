package obisplit

import (
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiapat"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
)

type SplitSequence struct {
	pattern         string
	name            string
	forward_pattern obiapat.ApatPattern
	reverse_pattern obiapat.ApatPattern
}

type Pattern_match struct {
	name    string
	pattern string
	match   string
	begin   int
	end     int
	nerrors int
	forward bool
}

func LocatePatterns(sequence *obiseq.BioSequence,
	patterns []SplitSequence) []Pattern_match {

	aseq, err := obiapat.MakeApatSequence(sequence, false)

	if err != nil {
		log.Fatalf("Cannot index sequence %s for patern matching", sequence.Id())
	}

	res := make([]Pattern_match, 0, 10)

	for _, split := range patterns {
		ms := split.forward_pattern.AllMatches(aseq, 0, aseq.Len())
		for _, m := range ms {
			m[0] = max(0, m[0])
			m[1] = min(sequence.Len(), m[1])
			match, err := sequence.Subsequence(m[0], m[1], false)

			if err != nil {
				log.Fatalf("Cannot extract pattern %s from sequence %s", split.pattern, sequence.Id())
			}

			res = append(res, Pattern_match{
				name:    split.name,
				pattern: split.pattern,
				match:   match.String(),
				begin:   m[0],
				end:     m[1],
				nerrors: m[2],
				forward: true,
			})
		}

		ms = split.reverse_pattern.AllMatches(aseq, 0, aseq.Len())
		for _, m := range ms {
			m[0] = max(0, m[0])
			m[1] = min(sequence.Len(), m[1])
			match, err := sequence.Subsequence(m[0], m[1], false)

			if err != nil {
				log.Fatalf("Cannot extract reverse pattern %s from sequence %s", split.pattern, sequence.Id())
			}

			match = match.ReverseComplement(true)

			res = append(res, Pattern_match{
				name:    split.name,
				pattern: split.pattern,
				match:   match.String(),
				begin:   m[0],
				end:     m[1],
				nerrors: m[2],
				forward: false,
			})
		}

	}

	sort.Slice(res, func(i, j int) bool {
		a := res[i].begin
		b := res[j].begin
		return a < b
	})

	log.Debugf("Sequence %s Raw match : %v", sequence.Id(), res)
	if len(res) > 1 {
		j := 0
		m1 := res[0]
		for _, m2 := range res[1:] {
			if m2.begin < m1.end {
				if m2.nerrors < m1.nerrors {
					m1 = m2
				}
				continue
			}
			res[j] = m1
			m1 = m2
			j++
		}

		res[j] = m1
		res = res[:j+1]
	}

	log.Debugf("Sequence %s No overlap match : %v", sequence.Id(), res)

	return res
}

func SplitPattern(sequence *obiseq.BioSequence,
	patterns []SplitSequence) (obiseq.BioSequenceSlice, error) {

	matches := LocatePatterns(sequence, patterns)

	from := Pattern_match{
		name:    "extremity",
		pattern: "",
		match:   "",
		begin:   0,
		end:     0,
		nerrors: 0,
		forward: true,
	}

	res := obiseq.MakeBioSequenceSlice(10)
	nfrag := 0
	res = res[:nfrag]

	for i, to := range matches {
		log.Debugf("from : %v  to : %v", from, to)
		start := from.end
		end := to.begin

		if i == 0 && end <= 0 {
			from = to
			continue
		}

		if end > start {
			log.Debugf("Extracting fragment %d from sequence %s [%d:%d]",
				nfrag+1, sequence.Id(),
				start, end,
			)

			sub, err := sequence.Subsequence(start, end, false)

			if err != nil {
				return res[:nfrag],
					fmt.Errorf("cannot extract fragment %d from sequence %s [%d:%d]",
						nfrag+1, sequence.Id(),
						start, end,
					)
			}

			nfrag++
			sub.SetAttribute("obisplit_frg", nfrag)

			if from.name == to.name {
				sub.SetAttribute("obisplit_group", from.name)
			} else {
				fname := from.name
				tname := to.name
				if tname == "extremity" {
					fname, tname = tname, fname
				} else {
					if tname < fname && fname != "extremity" {
						fname, tname = tname, fname
					}
				}
				sub.SetAttribute("obisplit_group", fmt.Sprintf("%s-%s", fname, tname))
			}

			sub.SetAttribute("obisplit_location", fmt.Sprintf("%d..%d", start+1, end))

			sub.SetAttribute("obisplit_right_error", to.nerrors)
			sub.SetAttribute("obisplit_left_error", from.nerrors)

			sub.SetAttribute("obisplit_right_pattern", to.pattern)
			sub.SetAttribute("obisplit_left_pattern", from.pattern)

			sub.SetAttribute("obisplit_left_match", from.match)
			sub.SetAttribute("obisplit_right_match", to.match)

			res = append(res, sub)

		}
		from = to
	}

	if from.end < sequence.Len() {
		to := Pattern_match{
			name:    "extremity",
			pattern: "",
			match:   "",
			begin:   sequence.Len(),
			end:     sequence.Len(),
			nerrors: 0,
			forward: true,
		}

		start := from.end
		end := to.begin

		sub, err := sequence.Subsequence(start, end, false)

		if err != nil {
			return res[:nfrag],
				fmt.Errorf("cannot extract last fragment %d from sequence %s [%d:%d]",
					nfrag+1, sequence.Id(),
					start, end,
				)
		}

		nfrag++
		sub.SetAttribute("obisplit_frg", nfrag)
		if from.name == to.name {
			sub.SetAttribute("obisplit_group", from.name)
		} else {
			fname := from.name
			tname := to.name
			if tname == "extremity" {
				fname, tname = tname, fname
			} else {
				if tname < fname && fname != "extremity" {
					fname, tname = tname, fname
				}
			}
			sub.SetAttribute("obisplit_group", fmt.Sprintf("%s-%s", fname, tname))
		}
		sub.SetAttribute("obisplit_location", fmt.Sprintf("%d..%d", start+1, end))

		sub.SetAttribute("obisplit_right_error", to.nerrors)
		sub.SetAttribute("obisplit_left_error", from.nerrors)

		sub.SetAttribute("obisplit_right_pattern", to.pattern)
		sub.SetAttribute("obisplit_left_pattern", from.pattern)

		sub.SetAttribute("obisplit_left_match", from.match)
		sub.SetAttribute("obisplit_right_match", to.match)

		res = append(res, sub)

	}

	for i := 0; i < nfrag; i++ {
		res[i].SetAttribute("obisplit_nfrg", nfrag)
	}

	return res[:nfrag], nil
}

func SplitPatternWorker(patterns []SplitSequence) obiseq.SeqWorker {
	f := func(sequence *obiseq.BioSequence) (obiseq.BioSequenceSlice, error) {
		return SplitPattern(sequence, patterns)
	}

	return f
}

func CLISlitPipeline() obiiter.Pipeable {

	worker := SplitPatternWorker(CLIConfig())

	annotator := obiseq.SeqToSliceWorker(worker, false)
	f := obiiter.SliceWorkerPipe(annotator, false, obioptions.CLIParallelWorkers())

	return f
}
