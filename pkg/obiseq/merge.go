package obiseq

import (
	"fmt"
	"math"
	"reflect"
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

type StatsOnValues map[string]int
type StatsOnWeights func(sequence *BioSequence) int
type StatsOnDescription struct {
	Name   string
	Key    string
	Weight StatsOnWeights
}
type StatsOnDescriptions map[string]StatsOnDescription

func BioseqCount(sequence *BioSequence) int {
	return sequence.Count()
}

func MakeStatsOnDescription(descriptor string) StatsOnDescription {
	parts := strings.SplitN(descriptor, ":", 2)
	var ff StatsOnWeights
	switch len(parts) {
	case 1:
		ff = func(s *BioSequence) int {
			return s.Count()
		}

	case 2:
		ff = func(s *BioSequence) int {
			v, ok := s.GetIntAttribute(parts[1])
			if !ok {
				return 0
			}
			return v
		}
	}

	return StatsOnDescription{
		Name:   descriptor,
		Key:    parts[0],
		Weight: ff,
	}
}

var _merge_prefix = "merged_"

// StatsOnSlotName returns the name of the slot that summarizes statistics of occurrence for a given attribute.
//
// Parameters:
// - key: the attribute key (string)
//
// Return type:
// - string
func StatsOnSlotName(key string) string {
	return _merge_prefix + key
}

// HasStatsOn tests if the sequence has already a slot summarizing statistics of occurrence for a given attribute.
//
// Parameters:
// - key: the attribute key (string)
//
// Return type:
// - bool
func (sequence *BioSequence) HasStatsOn(key string) bool {
	if !sequence.HasAnnotation() {
		return false
	}

	mkey := StatsOnSlotName(key)
	annotations := sequence.Annotations()
	_, ok := annotations[mkey]

	return ok
}

// StatsOn returns the slot summarizing statistics of occurrence for a given attribute.
//
// Parameters:
// - key: the attribute key (string) to be summarized
// - na: the value to be used if the attribute is not present
//
// Return type:
// - StatsOnValues
func (sequence *BioSequence) StatsOn(desc StatsOnDescription, na string) StatsOnValues {
	mkey := StatsOnSlotName(desc.Name)
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
				stats[k], err = obiutils.InterfaceToInt(v)
				if err != nil {
					log.Panicf("In sequence %s : %s stat tag not only containing integer values %s",
						sequence.Id(), mkey, istat)
				}
			}
		default:
			stats = make(StatsOnValues, 10)
			annotations[mkey] = stats
			newstat = true
		}
	} else {
		stats = make(StatsOnValues, 10)
		annotations[mkey] = stats
		newstat = true
	}

	if newstat && sequence.StatsPlusOne(desc, sequence, na) {
		delete(sequence.Annotations(), desc.Key)
	}

	return stats
}

// StatsPlusOne adds the count of the sequence toAdd to the count of the key in the stats.
//
// Parameters:
// - key: the attribute key (string) to be summarized
// - toAdd: the BioSequence to add to the stats
// - na: the value to be used if the attribute is not present
// Return type:
// - bool
func (sequence *BioSequence) StatsPlusOne(desc StatsOnDescription, toAdd *BioSequence, na string) bool {
	sval := na
	annotations := sequence.Annotations()
	stats := sequence.StatsOn(desc, na)
	retval := false

	if toAdd.HasAnnotation() {
		value, ok := toAdd.Annotations()[desc.Key]

		if ok {

			switch value := value.(type) {
			case string:
				sval = value
			case int,
				uint8, uint16, uint32, uint64,
				int8, int16, int32, int64, bool:
				sval = fmt.Sprint(value)
			case float64:
				if math.Floor(value) == value {
					sval = fmt.Sprint(int(value))
				} else {
					log.Fatalf("Trying to make stats on a float value (%v : %T)", value, value)
				}
			default:
				log.Fatalf("Trying to make stats not on a string, integer or boolean value (%v : %T)", value, value)
			}
			retval = true
		}

	}

	old, ok := stats[sval]
	if !ok {
		old = 0
	}
	stats[sval] = old + desc.Weight(toAdd)
	annotations[StatsOnSlotName(desc.Name)] = stats // TODO: check if this is necessary
	return retval
}

// Merge merges the given StatsOnValues with the current StatsOnValues.
//
// It takes a parameter `toMerged` of type StatsOnValues, which represents the StatsOnValues to be merged.
// It returns a value of type StatsOnValues, which represents the merged StatsOnValues.
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

// Merge merges two sequences into a single sequence.
//
// Parameters:
// - tomerge: the sequence to be merged (BioSequence)
// - na: the value to be used if the attribute is not present (string)
// - inplace: a boolean indicating whether to merge in place or not (bool)
// - statsOn: a variadic string parameter representing the attributes to be summarized (string)
//
// Return type:
// - *BioSequence: the merged sequence (BioSequence)
func (sequence *BioSequence) Merge(tomerge *BioSequence, na string, inplace bool, statsOn StatsOnDescriptions) *BioSequence {
	if !inplace {
		sequence = sequence.Copy()
	}

	if sequence.HasQualities() {
		sequence.SetQualities(nil)
	}

	annotations := sequence.Annotations()

	count := sequence.Count() + tomerge.Count()

	for key, desc := range statsOn {
		if tomerge.HasStatsOn(key) {
			smk := sequence.StatsOn(desc, na)
			mmk := tomerge.StatsOn(desc, na)

			annotations[StatsOnSlotName(key)] = smk.Merge(mmk)
		} else {
			sequence.StatsPlusOne(desc, tomerge, na)
		}
	}

	if tomerge.HasAnnotation() {
		ma := tomerge.Annotations()
		for k, va := range annotations {
			if !strings.HasPrefix(k, _merge_prefix) {
				vm, ok := ma[k]
				if ok {
					switch vm := vm.(type) {
					case int, float64, string, bool:
						if va != vm {
							delete(annotations, k)
						}
					default:
						if !reflect.DeepEqual(va, vm) {
							delete(annotations, k)
						}
					}

				} else {
					delete(annotations, k)
				}
			}
		}
	} else {
		for k := range annotations {
			if !strings.HasPrefix(k, _merge_prefix) {
				delete(annotations, k)
			}
		}
	}

	annotations["count"] = count
	return sequence
}

// Merge merges the given sequences into a single sequence.
//
// Parameters:
// - sequences: a slice of BioSequence objects to be merged (BioSequenceSlice)
// - na: the value to be used if the attribute is not present (string)
// - statsOn: a slice of strings representing the attributes to be summarized ([]string)
//
// Return type:
// - *BioSequence: the merged sequence (BioSequence)
func (sequences BioSequenceSlice) Merge(na string, statsOn StatsOnDescriptions) *BioSequence {
	seq := sequences[0]
	//sequences[0] = nil
	seq.SetQualities(nil)

	if len(sequences) == 1 {
		seq.Annotations()["count"] = seq.Count()
		for _, desc := range statsOn {
			seq.StatsOn(desc, na)
		}
	} else {
		for k, toMerge := range sequences[1:] {
			seq.Merge(toMerge, na, true, statsOn)
			toMerge.Recycle()
			sequences[1+k] = nil
		}
	}

	sequences.Recycle(false)
	return seq

}
