package obiseq

import (
	"fmt"
	"log"
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

	annotation := sequence.Annotations()

	annotation["count"] = tomerge.Count() + sequence.Count()

	for _, key := range keys {
		if tomerge.HasStatsOn(key) {
			smk := sequence.StatsOn(key)
			mmk := tomerge.StatsOn(key)
			smk.Merge(mmk)
		} else {
			sequence.StatsPlusOne(key, tomerge)
		}
	}

	return sequence
}
