package obiseq

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	"github.com/goccy/go-json"

	log "github.com/sirupsen/logrus"
)

type StatsOnValues struct {
	counts map[string]int
	lock   sync.RWMutex
}

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

func NewStatOnValues() *StatsOnValues {
	v := StatsOnValues{
		counts: make(map[string]int),
		lock:   sync.RWMutex{},
	}

	return &v
}

func MapAsStatsOnValues(m map[string]int) *StatsOnValues {
	v := StatsOnValues{
		counts: m,
		lock:   sync.RWMutex{},
	}

	return &v

}
func (sov *StatsOnValues) RLock() {
	sov.lock.RLock()
}

func (sov *StatsOnValues) RUnlock() {
	sov.lock.RUnlock()
}

func (sov *StatsOnValues) Lock() {
	sov.lock.Lock()
}

func (sov *StatsOnValues) Unlock() {
	sov.lock.Unlock()
}

func (sov *StatsOnValues) Get(key string) (int, bool) {
	if sov == nil {
		return -1, false
	}

	sov.RLock()
	defer sov.RUnlock()
	v, ok := sov.counts[key]
	if !ok {
		v = 0
	}
	return v, ok
}

func (sov *StatsOnValues) Map() map[string]int {
	return sov.counts
}

func (sov *StatsOnValues) Max() int {
	data, err := obiutils.Max(sov.counts)
	if err != nil {
		return -1
	}

	return data.(int)
}

func (sov *StatsOnValues) Min() int {
	data, err := obiutils.Min(sov.counts)
	if err != nil {
		return -1
	}

	return data.(int)
}

func (sov *StatsOnValues) Set(key string, value int) {
	if sov == nil {
		return
	}

	sov.Lock()
	defer sov.Unlock()
	sov.counts[key] = value
}

func (sov *StatsOnValues) Add(key string, value int) int {
	if sov == nil {
		return -1
	}

	sov.Lock()
	defer sov.Unlock()

	v, ok := sov.counts[key]
	if !ok {
		v = 0
	}
	v += value
	sov.counts[key] = v

	return v
}

func (sov *StatsOnValues) Len() int {
	sov.RLock()
	defer sov.RUnlock()
	return len(sov.counts)
}

func (sov *StatsOnValues) Keys() []string {
	v := make([]string, 0, sov.Len())
	sov.RLock()
	defer sov.RUnlock()
	for k := range sov.counts {
		v = append(v, k)
	}
	return v
}

func (sov *StatsOnValues) MarshalJSON() ([]byte, error) {
	sov.RLock()
	defer sov.RUnlock()
	return json.Marshal(sov.Map())
}

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
func (sequence *BioSequence) StatsOn(desc StatsOnDescription, na string) *StatsOnValues {
	var stats *StatsOnValues
	var newstat bool

	mkey := StatsOnSlotName(desc.Name)
	istat, ok := sequence.GetAttribute(mkey)

	if !ok {
		stats = NewStatOnValues()
		sequence.SetAttribute(mkey, stats)
		newstat = true
	} else {
		stats, ok = istat.(*StatsOnValues)

		if !ok {
			log.Panicf("In sequence %s : %s is not a StatsOnValues type %T", sequence.Id(), mkey, istat)
		}
		newstat = false
	}

	if newstat {
		sequence.StatsPlusOne(desc, sequence, na)
	}

	return stats
}

// StatsPlusOne updates the statistics on the given attribute (desc) on the receiver BioSequence
// with the value of the attribute on the toAdd BioSequence.
//
// Parameters:
// - desc: StatsOnDescription of the attribute to be updated
// - toAdd: the BioSequence containing the attribute to be updated
// - na: the value to be used if the attribute is not present
//
// Return type:
// - bool: true if the update was successful, false otherwise
func (sequence *BioSequence) StatsPlusOne(desc StatsOnDescription, toAdd *BioSequence, na string) bool {
	sval := na
	stats := sequence.StatsOn(desc, na)
	retval := false

	if toAdd.HasAnnotation() {
		value, ok := toAdd.GetAttribute(desc.Key)

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

	stats.Add(sval, desc.Weight(toAdd))
	return retval
}

// Merge merges the given StatsOnValues with the current StatsOnValues.
//
// It takes a parameter `toMerged` of type StatsOnValues, which represents the StatsOnValues to be merged.
// It returns a value of type StatsOnValues, which represents the merged StatsOnValues.
func (stats *StatsOnValues) Merge(toMerged *StatsOnValues) *StatsOnValues {
	toMerged.RLock()
	defer toMerged.RUnlock()
	stats.Lock()
	defer stats.Unlock()

	for k, val := range toMerged.counts {
		old, ok := stats.counts[k]
		if !ok {
			old = 0
		}
		stats.counts[k] = old + val
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

			sequence.annot_lock.Lock()
			annotations[StatsOnSlotName(key)] = smk.Merge(mmk)
			sequence.annot_lock.Unlock()
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

	sequence.SetCount(count)
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
		seq.SetCount(seq.Count())
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

	return seq
}
