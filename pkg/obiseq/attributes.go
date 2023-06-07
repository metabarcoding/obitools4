package obiseq

import (
	"fmt"
	"strconv"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

func (s *BioSequence) HasAttribute(key string) bool {
	ok := s.annotations != nil

	if ok {
		_, ok = s.annotations[key]
	}

	return ok
}

// A method that returns the value of the key in the annotation map.
func (s *BioSequence) GetAttribute(key string) (interface{}, bool) {
	var val interface{}
	ok := s.annotations != nil

	if ok {
		val, ok = s.annotations[key]
	}

	return val, ok
}

// A method that sets the value of the key in the annotation map.
func (s *BioSequence) SetAttribute(key string, value interface{}) {
	annot := s.Annotations()
	annot[key] = value
}

// A method that returns the value of the key in the annotation map.
func (s *BioSequence) GetIntAttribute(key string) (int, bool) {
	var val int
	var err error

	v, ok := s.GetAttribute(key)

	if ok {
		val, err = obiutils.InterfaceToInt(v)
		ok = err == nil
	}

	return val, ok
}

// Deleting the key from the annotation map.
func (s *BioSequence) DeleteAttribute(key string) {
	delete(s.Annotations(), key)
}

// Renaming the key in the annotation map.
func (s *BioSequence) RenameAttribute(newName, oldName string) {
	val, ok := s.GetAttribute(oldName)

	if ok {
		s.SetAttribute(newName, val)
		s.DeleteAttribute(oldName)
	}
}

// A method that returns the value of the key in the annotation map.
func (s *BioSequence) GetNumericAttribute(key string) (float64, bool) {
	var val float64
	var err error

	v, ok := s.GetAttribute(key)

	if ok {
		val, err = obiutils.InterfaceToFloat64(v)
		ok = err == nil
	}

	return val, ok
}

// A method that returns the value of the key in the annotation map.
func (s *BioSequence) GetStringAttribute(key string) (string, bool) {
	var val string
	v, ok := s.GetAttribute(key)

	if ok {
		val = fmt.Sprint(v)
	}

	return val, ok
}

// A method that returns the value of the key in the annotation map.
func (s *BioSequence) GetBoolAttribute(key string) (bool, bool) {
	var val bool
	var err error

	v, ok := s.GetAttribute(key)

	if ok {
		val, err = obiutils.InterfaceToBool(v)
		ok = err == nil
	}

	return val, ok
}

func (s *BioSequence) GetIntMap(key string) (map[string]int, bool) {
	var val map[string]int
	var err error

	v, ok := s.GetAttribute(key)

	if ok {
		val, err = obiutils.InterfaceToIntMap(v)
		ok = err == nil
	}

	return val, ok
}

// Returning the number of times the sequence has been observed.
func (s *BioSequence) Count() int {
	count, ok := s.GetIntAttribute("count")

	if !ok {
		count = 1
	}

	return count
}

// Setting the number of times the sequence has been observed.
func (s *BioSequence) SetCount(count int) {
	annot := s.Annotations()
	annot["count"] = count
}

// Returning the taxid of the sequence.
func (s *BioSequence) Taxid() int {
	taxid, ok := s.GetIntAttribute("taxid")

	if !ok {
		taxid = 1
	}

	return taxid
}

// Setting the taxid of the sequence.
func (s *BioSequence) SetTaxid(taxid int) {
	annot := s.Annotations()
	annot["taxid"] = taxid
}

func (s *BioSequence) OBITagRefIndex() map[int]string {

	var val map[int]string

	i, ok := s.GetAttribute("obitag_ref_index")

	if !ok {
		return nil
	}

	switch i := i.(type) {
	case map[int]string:
		val = i
	case map[string]interface{}:
		val = make(map[int]string, len(i))
		for k, v := range i {
			score, err := strconv.Atoi(k)
			if err != nil {
				log.Panicln(err)
			}

			val[score], err = obiutils.InterfaceToString(v)
			if err != nil {
				log.Panicln(err)
			}
		}
	case map[string]string:
		val = make(map[int]string, len(i))
		for k, v := range i {
			score, err := strconv.Atoi(k)
			if err != nil {
				log.Panicln(err)
			}
			val[score] = v

		}
	default:
		log.Panicln("value of attribute obitag_ref_index cannot be casted to a map[int]string")
	}

	return val
}

func (s *BioSequence) SetOBITagRefIndex(idx map[int]string) {
	s.SetAttribute("obitag_ref_index", idx)
}