package obiseq

import (
	"fmt"
	"strconv"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obilog"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

// AttributeKeys returns the keys of the attributes in the BioSequence.
// It optionally skips keys associated with container values based on the skip_container parameter.
//
// Parameters:
//   - skip_container: A boolean indicating whether to skip keys associated with a container value.
//
// Returns:
//   - A set of strings containing the keys of the BioSequence attributes.
func (s *BioSequence) AttributeKeys(skip_container, skip_definition bool) obiutils.Set[string] {
	keys := obiutils.MakeSet[string]()

	for k, v := range s.Annotations() {
		if !((skip_container && obiutils.IsAContainer(v)) ||
			(skip_definition && k == "definition")) {
			keys.Add(k)
		}
	}

	return keys
}

// Keys returns the keys of the BioSequence, including standard keys and attribute keys.
//
// It returns a set of strings containing the keys of the BioSequence.
// The keys include "id", "sequence", "qualities", and the attribute keys of the BioSequence.
//
// Parameters:
//   - skip_container: A boolean indicating whether to skip keys associated with container values.
//
// Returns:
//   - A set of strings containing the keys of the BioSequence.
func (s *BioSequence) Keys(skip_container, skip_definition bool) obiutils.Set[string] {
	keys := s.AttributeKeys(skip_container, skip_definition)
	keys.Add("id")

	if s.HasSequence() {
		keys.Add("sequence")
	}
	if s.HasQualities() {
		keys.Add("qualities")
	}

	return keys
}

// HasAttribute checks if the BioSequence has the specified attribute.
//
// Parameters:
//   - key: A string representing the attribute key to check.
//
// Returns:
//   - A boolean indicating whether the BioSequence has the attribute.
func (s *BioSequence) HasAttribute(key string) bool {
	if key == "id" {
		return true
	}

	if key == "sequence" && s.sequence != nil {
		return true
	}

	if key == "qualities" && s.qualities != nil {
		return true
	}
	ok := s.annotations != nil

	if ok {
		s.AnnotationsLock()
		defer s.AnnotationsUnlock()
		_, ok = s.annotations[key]
	}

	return ok
}

// GetAttribute returns the value associated with the given key in the BioSequence's annotations map and a boolean indicating whether the key exists.
//
// Parameters:
// - key: The key to look up in the annotations map.
//
// Returns:
// - val: The value associated with the given key.
// - ok: A boolean indicating whether the key exists in the annotations map.
func (s *BioSequence) GetAttribute(key string) (interface{}, bool) {

	if key == "id" {
		return s.id, true
	}

	if key == "sequence" {
		if s.HasSequence() {
			return s.String(), true
		}
		return nil, false
	}

	if key == "qualities" {
		if s.HasQualities() {
			return s.QualitiesString(), true
		}
		return nil, false
	}

	var val interface{}
	ok := s.annotations != nil

	if ok {
		s.AnnotationsLock()
		defer s.AnnotationsUnlock()
		val, ok = s.annotations[key]
	}

	return val, ok
}

// SetAttribute sets the value of a given key in the BioSequence annotations.
//
// Parameters:
// - key: the key to set the value for.
// - value: the value to set for the given key.
func (s *BioSequence) SetAttribute(key string, value interface{}) {

	if key == "id" {
		s.SetId(value.(string))
		return
	}

	if key == "sequence" {
		data, err := obiutils.InterfaceToString(value)
		if err != nil {
			obilog.Warnf("%s: cannot convert value %v to sequence", s.Id(), value)
			return
		}
		s.SetSequence([]byte(data))
		return
	}

	if key == "qualities" {
		s.SetQualities(value.([]byte))
		return
	}

	annot := s.Annotations()

	s.AnnotationsLock()
	defer s.AnnotationsUnlock()
	annot[key] = value
}

// GetIntAttribute returns an integer attribute value based on the provided key.
//
// It takes a key as a parameter and returns the corresponding integer value along
// with a boolean value indicating whether the key exists in the BioSequence, and if it can be converted to an integer.
//
// If the stored values is convertible to an integer, but was not stored as an integer, then the value will be stored as an integer.
//
// The returned boolean value will be true if the key exists, and false otherwise.
func (s *BioSequence) GetIntAttribute(key string) (int, bool) {
	var val int
	var err error

	v, ok := s.GetAttribute(key)

	if ok {
		val, ok = v.(int)
		if !ok {
			val, err = obiutils.InterfaceToInt(v)
			ok = err == nil
			if ok {
				s.SetAttribute(key, val)
			}
		}
	}

	return val, ok
}

func (s *BioSequence) GetFloatAttribute(key string) (float64, bool) {
	var val float64
	var err error

	v, ok := s.GetAttribute(key)

	if ok {
		val, ok = v.(float64)
		if !ok {
			val, err = obiutils.InterfaceToFloat64(v)
			ok = err == nil
			if ok {
				s.SetAttribute(key, val)
			}
		}
	}

	return val, ok
}

// DeleteAttribute deletes the attribute with the given key from the BioSequence.
//
// Parameters:
// - key: the key of the attribute to be deleted.
//
// No return value.
func (s *BioSequence) DeleteAttribute(key string) {
	if s.annotations != nil {
		s.AnnotationsLock()
		defer s.AnnotationsUnlock()
		delete(s.annotations, key)
	}
}

// RenameAttribute renames an attribute in the BioSequence.
//
// It takes two string parameters:
// - newName: the new name for the attribute.
// - oldName: the old name of the attribute to be renamed.
// It does not return anything.
func (s *BioSequence) RenameAttribute(newName, oldName string) {
	val, ok := s.GetAttribute(oldName)

	if ok {
		s.SetAttribute(newName, val)
		s.DeleteAttribute(oldName)
	}
}

// GetNumericAttribute returns the numeric value of the specified attribute key
// in the BioSequence object.
//
// Parameters:
//   - key: the attribute key to retrieve the numeric value for.
//
// Returns:
//   - float64: the numeric value of the attribute key.
//   - bool: indicates whether the attribute key exists and can be converted to a float64.
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

// GetStringAttribute retrieves the string value of a specific attribute from the BioSequence.
//
// Parameters:
// - key: the key of the attribute to retrieve.
//
// Returns:
// - string: the value of the attribute as a string.
// - bool: a boolean indicating whether the attribute was found or not.
func (s *BioSequence) GetStringAttribute(key string) (string, bool) {
	var val string
	v, ok := s.GetAttribute(key)

	if ok {
		val = fmt.Sprint(v)
	}

	return val, ok
}

// GetBoolAttribute returns the boolean attribute value associated with the given key in the BioSequence object.
//
// Parameters:
// - key: The key to retrieve the boolean attribute value.
//
// Return:
// - val: The boolean attribute value associated with the given key and can be converted to a boolean.
// - ok: A boolean value indicating whether the attribute value was successfully retrieved.
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

// GetIntMap returns a map[string]int and a boolean value indicating whether the key exists in the BioSequence.
//
// Parameters:
// - key: The key to retrieve the value from the BioSequence.
//
// Returns:
// - val: A map[string]int representing the value associated with the key and can be converted to a map[string]int.
// - ok: A boolean value indicating whether the key exists in the BioSequence.
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

func (s *BioSequence) GetStringMap(key string) (map[string]string, bool) {
	var val map[string]string

	var err error

	v, ok := s.GetAttribute(key)

	if ok {
		val, err = obiutils.InterfaceToStringMap(v)
		ok = err == nil
	}

	return val, ok
}

// GetIntSlice returns the integer slice value associated with the given key in the BioSequence object.
//
// Parameters:
// - key: The key used to retrieve the integer slice value.
//
// Returns:
// - []int: The integer slice value associated with the given key.
// - bool: A boolean indicating whether the key exists in the BioSequence object.
func (s *BioSequence) GetIntSlice(key string) ([]int, bool) {
	var val []int
	var err error

	v, ok := s.GetAttribute(key)

	if ok {
		val, ok = v.([]int)
		if !ok {
			val, err = obiutils.InterfaceToIntSlice(v)
			ok = err == nil
			if ok {
				s.SetAttribute(key, val)
			}
		}
	}

	return val, ok
}

// Count returns the value of the "count" attribute of the BioSequence.
//
// The count of a sequence is the number of times it has been observed in the dataset.
// It is represented in the sequence header as the "count" attribute.
// If the attribute is not found, the function returns 1 as the default count.
//
// It returns an integer representing the count value.
func (s *BioSequence) Count() int {
	count, ok := s.GetIntAttribute("count")

	if !ok {
		count = 1
	}

	return count
}

// SetCount sets the count of the BioSequence.
//
// The count of a sequence is the number of times it has been observed in the dataset.
// The value of the "count" attribute is set to the new count, event if the new count is 1.
// If the count is less than 1, the count is set to 1.
//
// count - the new count to set.
func (s *BioSequence) SetCount(count int) {
	if count < 1 {
		count = 1
	}
	s.SetAttribute("count", count)
}

func (s *BioSequence) OBITagRefIndex(slot ...string) map[int]string {
	key := "obitag_ref_index"

	if len(slot) > 0 {
		key = slot[0]
	}

	var val map[int]string

	i, ok := s.GetAttribute(key)

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

func (s *BioSequence) SetOBITagGeomRefIndex(idx map[int]string) {
	s.SetAttribute("obitag_geomref_index", idx)
}

func (s *BioSequence) OBITagGeomRefIndex() map[int]string {
	var val map[int]string

	i, ok := s.GetAttribute("obitag_geomref_index")

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
		log.Panicln("value of attribute obitag_geomref_index cannot be casted to a map[int]string")
	}

	return val
}

// GetCoordinate returns the coordinate of the BioSequence.
//
// Returns the coordinate of the BioSequence in the space of its reference database landmark sequences.
// if no coordinate is found, it returns nil.
//
// This function does not take any parameters.
//
// It returns a slice of integers ([]int).
func (s *BioSequence) GetCoordinate() []int {
	coord, ok := s.GetIntSlice("landmark_coord")
	if !ok {
		return nil
	}

	return coord
}

// SetCoordinate sets the coordinate of the BioSequence.
//
// coord: An array of integers representing the coordinate.
// This function does not return anything.
func (s *BioSequence) SetCoordinate(coord []int) {
	s.SetAttribute("landmark_coord", coord)
}

// SetLandmarkID sets the landmark ID of the BioSequence.
//
// Trying to set a negative landmark ID leads to a no operation.
//
// Parameters:
// id: The ID of the landmark.
func (s *BioSequence) SetLandmarkID(id int) {
	if id < 0 {
		return
	}
	s.SetAttribute("landmark_id", id)
}

// GetLandmarkID returns the landmark ID associated with the BioSequence.
//
// It retrieves the "landmark_id" attribute from the BioSequence's attributes map.
// If the attribute is not found, the function returns -1 as the default landmark ID.
// The landmark ID is an integer representing the number of the axis in the landmark space.
//
// It does not take any parameters.
// It returns an integer representing the landmark ID.
func (s *BioSequence) GetLandmarkID() int {
	val, ok := s.GetIntAttribute("landmark_id")

	if !ok {
		return -1
	}

	return val
}

// IsALandmark checks if the BioSequence is a landmark.
//
// A sequence is a landmark if its landmark ID is set (attribute "landmark_id").
//
// It returns a boolean indicating whether the BioSequence is a landmark or not.
func (s *BioSequence) IsALandmark() bool {
	return s.GetLandmarkID() != -1
}
