package obiseq

import (
	"bytes"
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

type __sequence__ struct {
	id          bytes.Buffer
	definition  bytes.Buffer
	sequence    bytes.Buffer
	qualities   bytes.Buffer
	feature     bytes.Buffer
	annotations Annotation
}

type BioSequence struct {
	sequence *__sequence__
}

type BioSequenceSlice []BioSequence

var NilBioSequence = BioSequence{sequence: nil}

func (s BioSequence) IsNil() bool {
	return s.sequence == nil
}

func (s *BioSequence) Reset() {
	s.sequence.id.Reset()
	s.sequence.definition.Reset()
	s.sequence.sequence.Reset()
	s.sequence.qualities.Reset()
	s.sequence.feature.Reset()

	for k := range s.sequence.annotations {
		delete(s.sequence.annotations, k)
	}

}

func (s BioSequence) Copy() BioSequence {
	new_seq := MakeEmptyBioSequence()
	new_seq.sequence.id.Write(s.sequence.id.Bytes())
	new_seq.sequence.definition.Write(s.sequence.definition.Bytes())
	new_seq.sequence.sequence.Write(s.sequence.sequence.Bytes())
	new_seq.sequence.qualities.Write(s.sequence.qualities.Bytes())
	new_seq.sequence.feature.Write(s.sequence.feature.Bytes())

	if len(s.sequence.annotations) > 0 {
		goutils.CopyMap(new_seq.sequence.annotations,
			s.sequence.annotations)
	}

	return new_seq
}

func (s BioSequence) Id() string {
	return s.sequence.id.String()
}
func (s BioSequence) Definition() string {
	return s.sequence.definition.String()
}

func (s BioSequence) Sequence() []byte {
	return s.sequence.sequence.Bytes()
}

func (s BioSequence) String() string {
	return s.sequence.sequence.String()
}
func (s BioSequence) Length() int {
	return s.sequence.sequence.Len()
}

func (s BioSequence) HasQualities() bool {
	return s.sequence.qualities.Len() > 0
}

func (s BioSequence) Qualities() Quality {
	if s.HasQualities() {
		return s.sequence.qualities.Bytes()
	} else {
		return __make_default_qualities__(s.sequence.sequence.Len())
	}
}

func (s BioSequence) Features() string {
	return s.sequence.feature.String()
}

func (s BioSequence) Annotations() Annotation {
	return s.sequence.annotations
}

func (s BioSequence) MD5() [16]byte {
	return md5.Sum(s.sequence.sequence.Bytes())
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
	s.sequence.id.Reset()
	s.sequence.id.WriteString(id)
}

func (s BioSequence) SetDefinition(definition string) {
	s.sequence.definition.Reset()
	s.sequence.definition.WriteString(definition)
}

func (s BioSequence) SetFeatures(feature string) {
	s.sequence.feature.Reset()
	s.sequence.feature.WriteString(feature)
}

func (s BioSequence) SetSequence(sequence []byte) {
	s.sequence.sequence.Reset()
	s.sequence.sequence.Write(sequence)
}

func (s BioSequence) SetQualities(qualities Quality) {
	s.sequence.qualities.Reset()
	s.sequence.qualities.Write(qualities)
}

func (s BioSequence) WriteQualities(data []byte) (int, error) {
	return s.sequence.qualities.Write(data)
}

func (s BioSequence) WriteByteQualities(data byte) error {
	return s.sequence.qualities.WriteByte(data)
}

func (s BioSequence) Write(data []byte) (int, error) {
	return s.sequence.sequence.Write(data)
}

func (s BioSequence) WriteString(data string) (int, error) {
	return s.sequence.sequence.WriteString(data)
}

func (s BioSequence) WriteByte(data byte) error {
	return s.sequence.sequence.WriteByte(data)
}

func (s BioSequence) WriteRune(data rune) (int, error) {
	return s.sequence.sequence.WriteRune(data)
}
