package obiseq

import (
	"fmt"
	"log"
	"strings"
)

type StatsOnValues map[string]int

func (sequence BioSequence) HasStatsOn(key string) bool {
	if !sequence.HasAnnotation() {
		return false
	}

	mkey := "merged_" + key
	annotations := sequence.Annotations()
	_, ok := annotations[mkey]

	return ok
}

func (sequence BioSequence) StatsOn(key string) StatsOnValues {
	mkey := "merged_" + key
	annotations := sequence.Annotations()
	istat, ok := annotations[mkey]

	var stats StatsOnValues
	var newstat bool

	if ok {
		switch istat := istat.(type) {
		case StatsOnValues:
			stats = istat
			newstat = false
		default:
			stats = make(StatsOnValues, 100)
			annotations[mkey] = stats
			newstat = true
		}
	} else {
		stats = make(StatsOnValues, 100)
		annotations[mkey] = stats
		newstat = true
	}

	if newstat && sequence.StatsPlusOne(key, sequence) {
		delete(sequence.Annotations(), key)
	}

	return stats
}

func (sequence BioSequence) StatsPlusOne(key string, toAdd BioSequence) bool {
	if toAdd.HasAnnotation() {
		stats := sequence.StatsOn(key)
		value, ok := toAdd.Annotations()[key]

		if ok {
			var sval string

			switch value := value.(type) {
			case string:
				sval = value
			case int,
				uint8, uint16, uint32, uint64,
				int8, int16, int32, int64, bool:
				sval = fmt.Sprint(value)
			default:
				log.Fatalf("Trying to make stats on a none string, integer or boolean value (%v)", value)
			}
			old, ok := stats[sval]
			if !ok {
				old = 0
			}
			stats[sval] = old + 1

			return true
		}
	}

	return false
}

func (stats StatsOnValues) Merge(toMerged StatsOnValues) StatsOnValues {
	for k, val := range toMerged {
		old, ok := stats[k]
		if !ok {
			old = 0
		}
		stats[k] = old + val
	}

	return stats
}

func (sequence BioSequence) Merge(tomerge BioSequence, inplace bool, keys ...string) BioSequence {
	if !inplace {
		sequence = sequence.Copy()
	}

	if sequence.HasQualities() {
		sequence.SetQualities(nil)
	}

	annotation := sequence.Annotations()

	count := tomerge.Count() + sequence.Count()

	for _, key := range keys {
		if tomerge.HasStatsOn(key) {
			smk := sequence.StatsOn(key)
			mmk := tomerge.StatsOn(key)
			smk.Merge(mmk)
		} else {
			sequence.StatsPlusOne(key, tomerge)
		}
	}

	if tomerge.HasAnnotation() {
		ma := tomerge.Annotations()
		for k, va := range annotation {
			if !strings.HasPrefix(k, "merged_") {
				vm, ok := ma[k]
				if !ok || vm != va {
					delete(annotation, k)
				}
			}
		}
	} else {
		for k := range annotation {
			if !strings.HasPrefix(k, "merged_") {
				delete(annotation, k)
			}
		}
	}

	annotation["count"] = count

	return sequence
}

func (sequences BioSequenceSlice) Unique(statsOn []string, keys ...string) BioSequenceSlice {
	uniq := make(map[string]*BioSequenceSlice, len(sequences))
	nVariant := 0

	for _, seq := range sequences {

		sstring := seq.String()
		pgroup, ok := uniq[sstring]

		if !ok {
			group := make(BioSequenceSlice, 0, 10)
			pgroup = &group
			uniq[sstring] = pgroup
		}

		ok = false
		i := 0
		var s BioSequence

		for i, s = range *pgroup {
			ok = true
			switch {
			case seq.HasAnnotation() && s.HasAnnotation():
				for _, k := range keys {
					seqV, seqOk := seq.Annotations()[k]
					sV, sOk := s.Annotations()[k]

					ok = ok && ((!seqOk && !sOk) || ((seqOk && sOk) && (seqV == sV)))

					if !ok {
						break
					}
				}
			case seq.HasAnnotation() && !s.HasAnnotation():
				for _, k := range keys {
					_, seqOk := seq.Annotations()[k]
					ok = ok && !seqOk
					if !ok {
						break
					}
				}
			case !seq.HasAnnotation() && s.HasAnnotation():
				for _, k := range keys {
					_, sOk := s.Annotations()[k]
					ok = ok && !sOk
					if !ok {
						break
					}
				}
			default:
				ok = true
			}

			if ok {
				break
			}
		}

		if ok {
			(*pgroup)[i] = s.Merge(seq, true, statsOn...)
		} else {
			seq.SetQualities(nil)
			if seq.Count() == 1 {
				seq.Annotations()["count"] = 1
			}
			*pgroup = append(*pgroup, seq)
			nVariant++
		}

	}

	output := make(BioSequenceSlice, 0, nVariant)
	for _, seqs := range uniq {
		output = append(output, *seqs...)
	}

	return output
}

func UniqueSliceWorker(statsOn []string, keys ...string) SeqSliceWorker {

	worker := func(sequences BioSequenceSlice) BioSequenceSlice {
		return sequences.Unique(statsOn, keys...)
	}

	return worker
}
