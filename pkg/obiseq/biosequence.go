package obiseq

import (
	"crypto/md5"
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

type Annotation map[string]interface{}

type BioSequence struct {
	id          string
	definition  string
	sequence    []byte
	qualities   []byte
	feature     []byte
	annotations Annotation
}

func MakeEmptyBioSequence() BioSequence {
	atomic.AddInt32(&_NewSeq, 1)
	atomic.AddInt32(&_InMemSeq, 1)

	//if atomic.CompareAndSwapInt32()()

	// if int(_NewSeq)%int(_BioLogRate) == 0 {
	// 	LogBioSeqStatus()
	// }

	return BioSequence{
		id:          "",
		definition:  "",
		sequence:    nil,
		qualities:   nil,
		feature:     nil,
		annotations: nil,
	}
}

func NewEmptyBioSequence() *BioSequence {
	s := MakeEmptyBioSequence()
	return &s
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

func NewBioSequence(id string,
	sequence []byte,
	definition string) *BioSequence {
	s := MakeBioSequence(id, sequence, definition)
	return &s
}

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

func (s *BioSequence) Copy() *BioSequence {
	newSeq := MakeEmptyBioSequence()

	newSeq.id = s.id
	newSeq.definition = s.definition

	newSeq.sequence = GetSlice(s.sequence...)
	newSeq.qualities = GetSlice(s.qualities...)
	newSeq.feature = GetSlice(s.feature...)

	if len(s.annotations) > 0 {
		newSeq.annotations = GetAnnotation(s.annotations)
	}

	return &newSeq
}

func (s *BioSequence) Id() string {
	return s.id
}
func (s *BioSequence) Definition() string {
	return s.definition
}

func (s *BioSequence) Sequence() []byte {
	return s.sequence
}

func (s *BioSequence) String() string {
	return string(s.sequence)
}
func (s *BioSequence) Length() int {
	return len(s.sequence)
}

func (s *BioSequence) HasQualities() bool {
	return len(s.qualities) > 0
}

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

func (s *BioSequence) HasAnnotation() bool {
	return len(s.annotations) > 0
}

func (s *BioSequence) Annotations() Annotation {

	if s.annotations == nil {
		s.annotations = GetAnnotation()
	}

	return s.annotations
}

func (s *BioSequence) MD5() [16]byte {
	return md5.Sum(s.sequence)
}

func (s *BioSequence) Count() int {
	if s.annotations == nil {
		return 1
	}

	if val, ok := (s.annotations)["count"]; ok {
		val, err := goutils.InterfaceToInt(val)
		if err == nil {
			return val
		}
	}
	return 1
}

func (s *BioSequence) Taxid() int {
	if s.annotations == nil {
		return 1
	}

	if val, ok := (s.annotations)["taxid"]; ok {
		val, err := goutils.InterfaceToInt(val)
		if err == nil {
			return val
		}
	}
	return 1
}

func (s *BioSequence) SetId(id string) {
	s.id = id
}

func (s *BioSequence) SetDefinition(definition string) {
	s.definition = definition
}

func (s *BioSequence) SetFeatures(feature []byte) {
	if cap(s.feature) >= 300 {
		RecycleSlice(&s.feature)
	}
	s.feature = feature
}

func (s *BioSequence) SetSequence(sequence []byte) {
	if s.sequence != nil {
		RecycleSlice(&s.sequence)
	}
	s.sequence = sequence
}

func (s *BioSequence) SetQualities(qualities Quality) {
	if s.qualities != nil {
		RecycleSlice(&s.qualities)
	}
	s.qualities = qualities
}

func (s *BioSequence) WriteQualities(data []byte) (int, error) {
	s.qualities = append(s.qualities, data...)
	return len(data), nil
}

func (s *BioSequence) WriteByteQualities(data byte) error {
	s.qualities = append(s.qualities, data)
	return nil
}

func (s *BioSequence) Write(data []byte) (int, error) {
	s.sequence = append(s.sequence, data...)
	return len(data), nil
}

func (s *BioSequence) WriteString(data string) (int, error) {
	bdata := []byte(data)
	return s.Write(bdata)
}

func (s *BioSequence) WriteByte(data byte) error {
	s.sequence = append(s.sequence, data)
	return nil
}

