// A package that defines a BioSequence struct.
//
// BioSequence are used to représente biological DNA sequences.
// The structure stores not only the sequence itself, but also some
// complementaty information. Among them:
//   - an identifier
//   - a definition
//   - the sequence quality scores
//   - the features
//   - the annotations
package obiseq

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"strconv"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
)

var _NewSeq = int32(0)
var _RecycleSeq = int32(0)
var _InMemSeq = int32(0)
var _MaxInMemSeq = int32(0)
var _BioLogRate = int(100000)

func LogBioSeqStatus() {
	log.Debugf("Created seq : %d Destroyed : %d  In Memory : %d", _NewSeq, _RecycleSeq, _InMemSeq)
}

type Quality []uint8

var __default_qualities__ = make(Quality, 0, 500)

func __make_default_qualities__(length int) Quality {
	cl := len(__default_qualities__)
	if cl < length {
		for i := cl; i <= length; i++ {
			__default_qualities__ = append(__default_qualities__, 40)
		}
	}
	return __default_qualities__[0:length]
}

// `Annotation` is a map of strings to interfaces.
// It is used to store
type Annotation map[string]interface{}

// A BioSequence is a sequence of bytes with an identifier, a definition, a sequence, qualities,
// features and annotations. It aims to represent a biological sequence
type BioSequence struct {
	id          string // The identidier of the sequence (private accessible through the method Id)
	definition  string // The documentation of the sequence (private accessible through the method Definition)
	sequence    []byte // The sequence itself, it is accessible by the methode Sequence
	qualities   []byte // The quality scores of the sequence.
	feature     []byte
	annotations Annotation
}

// MakeEmptyBioSequence() creates a new BioSequence object with no data
func MakeEmptyBioSequence() BioSequence {
	atomic.AddInt32(&_NewSeq, 1)
	atomic.AddInt32(&_InMemSeq, 1)

	return BioSequence{
		id:          "",
		definition:  "",
		sequence:    nil,
		qualities:   nil,
		feature:     nil,
		annotations: nil,
	}
}

// `NewEmptyBioSequence()` returns a pointer to a new empty BioSequence
func NewEmptyBioSequence() *BioSequence {
	s := MakeEmptyBioSequence()
	return &s
}

// `MakeBioSequence` creates a new `BioSequence` with the given `id`, `sequence`, and `definition`
func MakeBioSequence(id string,
	sequence []byte,
	definition string) BioSequence {
	bs := MakeEmptyBioSequence()
	bs.SetId(id)
	bs.SetSequence(sequence)
	bs.SetDefinition(definition)
	return bs
}

// `NewBioSequence` creates a new `BioSequence` struct and returns a pointer to it
func NewBioSequence(id string,
	sequence []byte,
	definition string) *BioSequence {
	s := MakeBioSequence(id, sequence, definition)
	return &s
}

// A method that is called when the sequence is no longer needed.
func (sequence *BioSequence) Recycle() {

	atomic.AddInt32(&_RecycleSeq, 1)
	atomic.AddInt32(&_InMemSeq, -1)

	// if int(_RecycleSeq)%int(_BioLogRate) == 0 {
	// 	LogBioSeqStatus()
	// }

	if sequence != nil {
		RecycleSlice(&sequence.sequence)
		sequence.sequence = nil
		RecycleSlice(&sequence.feature)
		sequence.feature = nil
		RecycleSlice(&sequence.qualities)
		sequence.qualities = nil

		RecycleAnnotation(&sequence.annotations)
		sequence.annotations = nil
	}
}

// Copying the BioSequence.
func (s *BioSequence) Copy() *BioSequence {
	newSeq := MakeEmptyBioSequence()

	newSeq.id = s.id
	newSeq.definition = s.definition

	newSeq.sequence = CopySlice(s.sequence)
	newSeq.qualities = CopySlice(s.qualities)
	newSeq.feature = CopySlice(s.feature)

	if len(s.annotations) > 0 {
		newSeq.annotations = GetAnnotation(s.annotations)
	}

	return &newSeq
}

// A method that returns the id of the sequence.
func (s *BioSequence) Id() string {
	return s.id
}

// A method that returns the definition of the sequence.
func (s *BioSequence) Definition() string {
	return s.definition
}

// A method that returns the sequence as a byte slice.
func (s *BioSequence) Sequence() []byte {
	return s.sequence
}

// A method that returns the sequence as a string.
func (s *BioSequence) String() string {
	return string(s.sequence)
}

// Returning the length of the sequence.
func (s *BioSequence) Len() int {
	return len(s.sequence)
}

// Checking if the BioSequence has quality scores.
func (s *BioSequence) HasQualities() bool {
	return len(s.qualities) > 0
}

// Returning the qualities of the sequence.
func (s *BioSequence) Qualities() Quality {
	if s.HasQualities() {
		return s.qualities
	} else {
		return __make_default_qualities__(len(s.sequence))
	}
}

func (s *BioSequence) Features() string {
	return string(s.feature)
}

// Checking if the BioSequence has annotations.
func (s *BioSequence) HasAnnotation() bool {
	return len(s.annotations) > 0
}

// Returning the annotations of the BioSequence.
func (s *BioSequence) Annotations() Annotation {

	if s.annotations == nil {
		s.annotations = GetAnnotation()
	}

	return s.annotations
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
		val, err = goutils.InterfaceToInt(v)
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
func (s *BioSequence) GetBool(key string) (bool, bool) {
	var val bool
	var err error

	v, ok := s.GetAttribute(key)

	if ok {
		val, err = goutils.InterfaceToBool(v)
		ok = err == nil
	}

	return val, ok
}

func (s *BioSequence) GetIntMap(key string) (map[string]int, bool) {
	var val map[string]int
	var err error

	v, ok := s.GetAttribute(key)

	if ok {
		val, err = goutils.InterfaceToIntMap(v)
		ok = err == nil
	}

	return val, ok
}

// Returning the MD5 hash of the sequence.
func (s *BioSequence) MD5() [16]byte {
	return md5.Sum(s.sequence)
}

// Returning the number of times the sequence has been observed.
func (s *BioSequence) Count() int {
	count, ok := s.GetIntAttribute("count")

	if !ok {
		count = 1
	}

	return count
}

// Returning the taxid of the sequence.
func (s *BioSequence) Taxid() int {
	taxid, ok := s.GetIntAttribute("taxid")

	if !ok {
		taxid = 1
	}

	return taxid
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

			val[score], err = goutils.InterfaceToString(v)
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

func (s *BioSequence) SetTaxid(taxid int) {
	annot := s.Annotations()
	annot["taxid"] = taxid
}

// Setting the id of the BioSequence.
func (s *BioSequence) SetId(id string) {
	s.id = id
}

// Setting the definition of the sequence.
func (s *BioSequence) SetDefinition(definition string) {
	s.definition = definition
}

// Setting the features of the BioSequence.
func (s *BioSequence) SetFeatures(feature []byte) {
	if cap(s.feature) >= 300 {
		RecycleSlice(&s.feature)
	}
	s.feature = feature
}

// Setting the sequence of the BioSequence.
func (s *BioSequence) SetSequence(sequence []byte) {
	if s.sequence != nil {
		RecycleSlice(&s.sequence)
	}
	s.sequence = bytes.ToLower(sequence)
}

// Setting the qualities of the BioSequence.
func (s *BioSequence) SetQualities(qualities Quality) {
	if s.qualities != nil {
		RecycleSlice(&s.qualities)
	}
	s.qualities = qualities
}

// A method that appends a byte slice to the qualities of the BioSequence.
func (s *BioSequence) WriteQualities(data []byte) (int, error) {
	s.qualities = append(s.qualities, data...)
	return len(data), nil
}

// Appending a byte to the qualities of the BioSequence.
func (s *BioSequence) WriteByteQualities(data byte) error {
	s.qualities = append(s.qualities, data)
	return nil
}

// Clearing the sequence.
func (s *BioSequence) ClearQualities() {
	s.qualities = s.qualities[0:0]
}

// A method that appends a byte slice to the sequence.
func (s *BioSequence) Write(data []byte) (int, error) {
	s.sequence = append(s.sequence, data...)
	return len(data), nil
}

// A method that appends a string to the sequence.
func (s *BioSequence) WriteString(data string) (int, error) {
	bdata := []byte(data)
	return s.Write(bdata)
}

// A method that appends a byte to the sequence.
func (s *BioSequence) WriteByte(data byte) error {
	s.sequence = append(s.sequence, data)
	return nil
}

// Clearing the sequence.
func (s *BioSequence) Clear() {
	s.sequence = s.sequence[0:0]
}
