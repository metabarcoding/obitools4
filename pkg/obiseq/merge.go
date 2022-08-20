package obiseq

import (
	"fmt"
	"strings"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
	log "github.com/sirupsen/logrus"
)

type StatsOnValues map[string]int

func StatsOnSlotName(key string) string {
	return "merged_" + key
}

/*
 	Tests if the sequence has already a slot summarizing statistics
	 of occurrence for a given attribute.
*/
func (sequence *BioSequence) HasStatsOn(key string) bool {
	if !sequence.HasAnnotation() {
		return false
	}

	mkey := StatsOnSlotName(key)
	annotations := sequence.Annotations()
	_, ok := annotations[mkey]

	return ok
}

func (sequence *BioSequence) StatsOn(key string, na string) StatsOnValues {
	mkey := StatsOnSlotName(key)
	annotations := sequence.Annotations()
	istat, ok := annotations[mkey]

	var stats StatsOnValues
	var newstat bool

	if ok {
		switch istat := istat.(type) {
		case StatsOnValues:
			stats = istat
			newstat = false
		case map[string]int:
			stats = istat
			newstat = false
		case map[string]interface{}:
			stats = make(StatsOnValues, len(istat))
			newstat = false
			var err error
			for k, v := range istat {
				stats[k], err = goutils.InterfaceToInt(v)
				if err != nil {
					log.Panicf("In sequence %s : %s stat tag not only containing integer values %s",
						sequence.Id(), mkey, istat)
				}
			}
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

	if newstat && sequence.StatsPlusOne(key, sequence, na) {
		delete(sequence.Annotations(), key)
	}

	return stats
}

func (sequence *BioSequence) StatsPlusOne(key string, toAdd *BioSequence, na string) bool {
	sval := na
	annotations := sequence.Annotations()
	stats := sequence.StatsOn(key, na)
	retval := false

	if toAdd.HasAnnotation() {
		value, ok := toAdd.Annotations()[key]

		if ok {

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
			retval = true
		}

	}

	old, ok := stats[sval]
	if !ok {
		old = 0
	}
	stats[sval] = old + toAdd.Count()
	annotations[StatsOnSlotName(key)] = stats
	return retval
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

/*

	Merges two sequences
*/
func (sequence *BioSequence) Merge(tomerge *BioSequence, na string, inplace bool, statsOn ...string) *BioSequence {
	if !inplace {
		sequence = sequence.Copy()
	}

	if sequence.HasQualities() {
		sequence.SetQualities(nil)
	}

	annotations := sequence.Annotations()

	count := sequence.Count() + tomerge.Count()

	for _, key := range statsOn {
		if tomerge.HasStatsOn(key) {
			smk := sequence.StatsOn(key, na)
			mmk := tomerge.StatsOn(key, na)

			annotations[StatsOnSlotName(key)] = smk.Merge(mmk)
		} else {
			sequence.StatsPlusOne(key, tomerge, na)
		}
	}

	if tomerge.HasAnnotation() {
		ma := tomerge.Annotations()
		for k, va := range annotations {
			if !strings.HasPrefix(k, "merged_") {
				vm, ok := ma[k]
				if !ok || vm != va {
					delete(annotations, k)
				}
			}
		}
	} else {
		for k := range annotations {
			if !strings.HasPrefix(k, "merged_") {
				delete(annotations, k)
			}
		}
	}

	annotations["count"] = count
	return sequence
}

/**
  Merges a set of sequence into a single sequence.

  The function assumes that every sequence in the batch is
  identical in term of sequence. Actually the function only
  aggregates the annotations of the different sequences to be merged

  Quality information is lost during the merge procedure.
*/
func (sequences BioSequenceSlice) Merge(na string, statsOn []string) *BioSequence {
	seq := sequences[0]
	//sequences[0] = nil
	seq.SetQualities(nil)

	if len(sequences) == 1 {
		seq.Annotations()["count"] = seq.Count()
		for _, v := range statsOn {
			seq.StatsOn(v, na)
		}
	} else {
		for k, toMerge := range sequences[1:] {
			seq.Merge(toMerge, na, true, statsOn...)
			toMerge.Recycle()
			sequences[1+k] = nil
		}
	}

	sequences.Recycle()
	return seq

}
