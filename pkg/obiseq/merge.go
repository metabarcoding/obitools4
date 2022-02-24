package obiseq

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

type StatsOnValues map[string]int

func (sequence *BioSequence) HasStatsOn(key string) bool {
	if !sequence.HasAnnotation() {
		return false
	}

	mkey := "merged_" + key
	annotations := sequence.Annotations()
	_, ok := annotations[mkey]

	return ok
}

func (sequence *BioSequence) StatsOn(key string, na string) StatsOnValues {
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

	if newstat && sequence.StatsPlusOne(key, sequence, na) {
		delete(sequence.Annotations(), key)
	}

	return stats
}

func (sequence *BioSequence) StatsPlusOne(key string, toAdd *BioSequence, na string) bool {
	sval := na
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
	stats[sval] = old + 1

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

func (sequence *BioSequence) Merge(tomerge *BioSequence, na string, inplace bool, statsOn ...string) *BioSequence {
	if !inplace {
		sequence = sequence.Copy()
	}

	if sequence.HasQualities() {
		sequence.SetQualities(nil)
	}

	annotation := sequence.Annotations()

	count := sequence.Count() + tomerge.Count()

	for _, key := range statsOn {
		if tomerge.HasStatsOn(key) {
			smk := sequence.StatsOn(key, na)
			mmk := tomerge.StatsOn(key, na)
			smk.Merge(mmk)
		} else {
			sequence.StatsPlusOne(key, tomerge, na)
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

func (sequences BioSequenceSlice) Merge(na string, statsOn []string) *BioSequence {
	seq := sequences[0]
	//sequences[0] = nil
	seq.SetQualities(nil)

	if len(sequences) == 1 {
		seq.Annotations()["count"] = 1
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
