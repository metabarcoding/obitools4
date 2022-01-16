package obiseq

import (
	"crypto/md5"

	"git.metabarcoding.org/lecasofts/go/obitools/pkg/goutils"
)

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

type Annotation map[string]interface{}

type _BioSequence struct {
	id          string
	definition  string
	sequence    []byte
	qualities   []byte
	feature     []byte
	annotations Annotation
}

type BioSequence struct {
	sequence *_BioSequence
}

func MakeEmptyBioSequence() BioSequence {
	bs := _BioSequence{
		id:          "",
		definition:  "",
		sequence:    nil,
		qualities:   nil,
		feature:     nil,
		annotations: nil,
	}
	return BioSequence{&bs}
}

func MakeBioSequence(id string,
	sequence []byte,
	definition string) BioSequence {
	bs := MakeEmptyBioSequence()
	bs.SetId(id)
	bs.SetSequence(sequence)
	bs.SetDefinition(definition)
	return bs
}

func (sequence *BioSequence) Recycle() {

	pseq := sequence.sequence

	RecycleSlice(pseq.sequence)
	RecycleSlice(pseq.feature)
	RecycleSlice(pseq.feature)

	RecycleAnnotation(pseq.annotations)

	sequence.sequence = nil
}

var NilBioSequence = BioSequence{sequence: nil}

func (s BioSequence) IsNil() bool {
	return s.sequence == nil
}

func (s BioSequence) Copy() BioSequence {
	newSeq := MakeEmptyBioSequence()

	newSeq.sequence.id = s.sequence.id
	newSeq.sequence.definition = s.sequence.definition

	newSeq.sequence.sequence = GetSlice(s.sequence.sequence...)
	newSeq.sequence.qualities = GetSlice(s.sequence.qualities...)
	newSeq.sequence.feature = GetSlice(s.sequence.feature...)

	if len(s.sequence.annotations) > 0 {
		newSeq.sequence.annotations = GetAnnotation(s.sequence.annotations)
	}

	return newSeq
}

func (s BioSequence) Id() string {
	return s.sequence.id
}
func (s BioSequence) Definition() string {
	return s.sequence.definition
}

func (s BioSequence) Sequence() []byte {
	return s.sequence.sequence
}

func (s BioSequence) String() string {
	return string(s.sequence.sequence)
}
func (s BioSequence) Length() int {
	return len(s.sequence.sequence)
}

func (s BioSequence) HasQualities() bool {
	return len(s.sequence.qualities) > 0
}

func (s BioSequence) Qualities() Quality {
	if s.HasQualities() {
		return s.sequence.qualities
	} else {
		return __make_default_qualities__(len(s.sequence.sequence))
	}
}

func (s BioSequence) Features() string {
	return string(s.sequence.feature)
}

func (s BioSequence) HasAnnotation() bool {
	return len(s.sequence.annotations) > 0
}

func (s BioSequence) Annotations() Annotation {
	if s.sequence.annotations == nil {
		s.sequence.annotations = GetAnnotation()
	}
	return s.sequence.annotations
}

func (s BioSequence) MD5() [16]byte {
	return md5.Sum(s.sequence.sequence)
}

func (s BioSequence) Count() int {
	if s.sequence.annotations == nil {
		return 1
	}

	if val, ok := (s.sequence.annotations)["count"]; ok {
		val, err := goutils.InterfaceToInt(val)
		if err == nil {
			return val
		}
	}
	return 1
}

func (s BioSequence) Taxid() int {
	if s.sequence.annotations == nil {
		return 1
	}

	if val, ok := (s.sequence.annotations)["taxid"]; ok {
		val, err := goutils.InterfaceToInt(val)
		if err == nil {
			return val
		}
	}
	return 1
}

func (s BioSequence) SetId(id string) {
	s.sequence.id = id
}

func (s BioSequence) SetDefinition(definition string) {
	s.sequence.definition = definition
}

func (s BioSequence) SetFeatures(feature []byte) {
	if cap(s.sequence.feature) >= 300 {
		RecycleSlice(s.sequence.feature)
	}
	s.sequence.feature = feature
}

func (s BioSequence) SetSequence(sequence []byte) {
	if s.sequence.sequence != nil {
		RecycleSlice(s.sequence.sequence)
	}
	s.sequence.sequence = sequence
}

func (s BioSequence) SetQualities(qualities Quality) {
	if s.sequence.qualities != nil {
		RecycleSlice(s.sequence.qualities)
	}
	s.sequence.qualities = qualities
}

func (s BioSequence) WriteQualities(data []byte) (int, error) {
	s.sequence.qualities = append(s.sequence.qualities, data...)
	return len(data), nil
}

func (s BioSequence) WriteByteQualities(data byte) error {
	s.sequence.qualities = append(s.sequence.qualities, data)
	return nil
}

func (s BioSequence) Write(data []byte) (int, error) {
	s.sequence.sequence = append(s.sequence.sequence, data...)
	return len(data), nil
}

func (s BioSequence) WriteString(data string) (int, error) {
	bdata := []byte(data)
	return s.Write(bdata)
}

func (s BioSequence) WriteByte(data byte) error {
	s.sequence.sequence = append(s.sequence.sequence, data)
	return nil
}
