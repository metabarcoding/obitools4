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
	"crypto/md5"
	"slices"
	"sync"
	"sync/atomic"
	"unsafe"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obioptions"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

var _NewSeq = int32(0)
var _RecycleSeq = int32(0)
var _InMemSeq = int32(0)

// var _MaxInMemSeq = int32(0)
// var _BioLogRate = int(100000)

func LogBioSeqStatus() {
	log.Debugf("Created seq : %d Destroyed : %d  In Memory : %d", _NewSeq, _RecycleSeq, _InMemSeq)
}

type Quality []uint8

var __default_qualities__ = make(Quality, 0, 500)

// __make_default_qualities__ generates a default quality slice of the given length.
//
// It takes an integer parameter 'length' which specifies the desired length of the quality slice.
// It returns a Quality slice which is a subset of the '__default_qualities__' slice.
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
	source      string // The filename without directory name and extension from where the sequence was read.
	sequence    []byte // The sequence itself, it is accessible by the methode Sequence
	qualities   []byte // The quality scores of the sequence.
	feature     []byte
	paired      *BioSequence // A pointer to the paired sequence
	annotations Annotation
	annot_lock  sync.Mutex
}

// NewEmptyBioSequence creates a new BioSequence object with an empty sequence.
//
// The preallocate parameter specifies the number of bytes to preallocate for the sequence. If preallocate is greater than 0, the sequence will be preallocated with the specified number of bytes. If preallocate is 0, the sequence will not be preallocated.
// The function returns a pointer to the newly created BioSequence object.
func NewEmptyBioSequence(preallocate int) *BioSequence {
	atomic.AddInt32(&_NewSeq, 1)
	atomic.AddInt32(&_InMemSeq, 1)

	seq := []byte(nil)
	if preallocate > 0 {
		seq = GetSlice(preallocate)
	}

	return &BioSequence{
		id: "",
		//definition:  "",
		source:      "",
		sequence:    seq,
		qualities:   nil,
		feature:     nil,
		paired:      nil,
		annotations: nil,
		annot_lock:  sync.Mutex{},
	}
}

// NewBioSequence creates a new BioSequence object with the given ID, sequence, and definition.
//
// Parameters:
// - id: the ID of the BioSequence.
// - sequence: the sequence data of the BioSequence.
// - definition: the definition of the BioSequence.
//
// Returns:
// - *BioSequence: the newly created BioSequence object.
func NewBioSequence(id string,
	sequence []byte,
	definition string) *BioSequence {
	bs := NewEmptyBioSequence(0)
	bs.SetId(id)
	bs.SetSequence(sequence)
	bs.SetDefinition(definition)
	return bs
}

// NewBioSequenceWithQualities creates a new BioSequence object with the given id, sequence, definition, and qualities.
//
// Parameters:
// - id: the id of the BioSequence.
// - sequence: the sequence data of the BioSequence.
// - definition: the definition of the BioSequence.
// - qualities: the qualities data of the BioSequence.
//
// Returns:
// - *BioSequence: a pointer to the newly created BioSequence object.
func NewBioSequenceWithQualities(id string,
	sequence []byte,
	definition string,
	qualities []byte) *BioSequence {
	bs := NewEmptyBioSequence(0)
	bs.SetId(id)
	bs.SetSequence(sequence)
	bs.SetDefinition(definition)
	bs.SetQualities(qualities)
	return bs
}

// Recycle recycles the BioSequence object.
//
// It decreases the count of in-memory sequences and increases the count of recycled sequences.
// It also recycles the various slices and annotations of the BioSequence object.
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

// Copy returns a new BioSequence that is a copy of the original BioSequence.
//
// It copies the id of the original BioSequence to the new BioSequence.
// It also creates new slices and copies the values from the original BioSequence's sequence, qualities, and feature fields to the new BioSequence.
// If the original BioSequence has annotations, it locks the annot_lock and copies the annotations to the new BioSequence.
//
// The function returns the new BioSequence.
func (s *BioSequence) Copy() *BioSequence {
	newSeq := NewEmptyBioSequence(0)

	newSeq.id = s.id

	newSeq.sequence = CopySlice(s.sequence)
	newSeq.qualities = CopySlice(s.qualities)
	newSeq.feature = CopySlice(s.feature)

	if len(s.annotations) > 0 {
		s.annot_lock.Lock()
		defer s.annot_lock.Unlock()
		newSeq.annotations = GetAnnotation(s.annotations)
	}

	return newSeq
}

// Id returns the ID of the BioSequence.
//
// No parameters.
// Returns a string.
func (s *BioSequence) Id() string {
	return s.id
}

// Definition returns the definition of the BioSequence.
//
// No parameters.
// Returns a string.
func (s *BioSequence) Definition() string {
	definition := ""
	var err error
	def, ok := s.GetAttribute("definition")
	if ok {
		definition, err = obiutils.InterfaceToString(def)
		if err != nil {
			definition = ""
		}
	}
	return definition
}

func (s *BioSequence) HasDefinition() bool {
	return s.HasAttribute("definition")
}

// HasSequence checks if the BioSequence has a sequence.
//
// No parameters.
// Returns a boolean.
func (s *BioSequence) HasSequence() bool {
	return s.sequence != nil && len(s.sequence) > 0
}

// Sequence returns the sequence of the BioSequence.
//
// Returns:
// - []byte: The sequence of the BioSequence.
func (s *BioSequence) Sequence() []byte {
	return s.sequence
}

// String returns the string representation of the Sequence.
//
// No parameters.
// Returns a string.
func (s *BioSequence) String() string {
	return string(s.sequence)
}

// Len returns the length of the BioSequence.
//
// It does not take any parameters.
// It returns an integer representing the length of the sequence.
func (s *BioSequence) Len() int {
	if s == nil {
		return 0
	}
	return len(s.sequence)
}

// HasQualities checks if the BioSequence has sequence qualitiy scores.
//
// This function does not have any parameters.
// It returns a boolean value indicating whether the BioSequence has qualities.
func (s *BioSequence) HasQualities() bool {
	return s.qualities != nil && len(s.qualities) > 0
}

// Qualities returns the sequence quality scores of the BioSequence.
//
// It checks if the BioSequence has qualities. If it does, it returns the qualities
// stored in the BioSequence struct. Otherwise, it creates and returns default
// qualities based on the length of the sequence.
//
// Returns:
//   - Quality: The quality of the BioSequence.
func (s *BioSequence) Qualities() Quality {
	if s.HasQualities() {
		return s.qualities
	}
	return __make_default_qualities__(len(s.sequence))
}

// QualitiesString returns the string representation of the qualities of the BioSequence.
//
// Returns a string representing the qualities of the BioSequence after applying the shift.
func (s *BioSequence) QualitiesString() string {
	quality_shift := obioptions.OutputQualityShift()

	qual := s.Qualities()
	qual_ascii := make([]byte, len(qual))

	for i := 0; i < len(qual); i++ {
		quality := qual[i]
		if quality > 93 {
			quality = 93
		}
		qual_ascii[i] = quality + quality_shift
	}

	qual_sting := unsafe.String(unsafe.SliceData(qual_ascii), len(qual))
	return qual_sting
}

// Features returns the feature string of the BioSequence.
//
//	The feature string contains the EMBL/GenBank not parsed feature table
//
// as extracted from the flat file.
//
// No parameters.
// Returns a string.
func (s *BioSequence) Features() string {
	return string(s.feature)
}

// HasAnnotation checks if the BioSequence has any annotations.
//
// It does not take any parameters.
// It returns a boolean value indicating whether the BioSequence has any annotations.
func (s *BioSequence) HasAnnotation() bool {
	return len(s.annotations) > 0
}

// Annotations returns the Annotation object associated with the BioSequence.
//
// This function does not take any parameters.
// It returns an Annotation object.
func (s *BioSequence) Annotations() Annotation {
	if s.annotations == nil {
		s.annotations = GetAnnotation()
	}
	return s.annotations
}

// AnnotationsLock locks the annotation of the BioSequence.
//
// This function acquires a lock on the annotation of the BioSequence,
// preventing concurrent access to it.
func (s *BioSequence) AnnotationsLock() {
	s.annot_lock.Lock()
}

// AnnotationsUnlock unlocks the annotations mutex in the BioSequence struct.
//
// No parameters.
// No return types.
func (s *BioSequence) AnnotationsUnlock() {
	s.annot_lock.Unlock()
}

// HasSource checks if the BioSequence has a source.
//
// The source is the filename without directory name and extension from where the sequence was read.
//
// No parameters.
// Returns a boolean value indicating whether the BioSequence has a source or not.
func (s *BioSequence) HasSource() bool {
	return len(s.source) > 0
}

// Source returns the source of the BioSequence.
//
// The source is the filename without directory name and extension from where the sequence was read.
//
// This function does not take any parameters.
// It returns a string.
func (s *BioSequence) Source() string {
	return s.source
}

// MD5 calculates the MD5 hash of the BioSequence.
//
// No parameters.
// Returns [16]byte, the MD5 hash of the BioSequence.
func (s *BioSequence) MD5() [16]byte {
	return md5.Sum(s.sequence)
}

// SetId sets the id of the BioSequence.
//
// Parameters:
// - id: the new id for the BioSequence.
//
// No return value.
func (s *BioSequence) SetId(id string) {
	s.id = id
}

// SetDefinition sets the definition of the BioSequence.
//
// It takes a string parameter 'definition' and assigns it to the 'definition' field of the BioSequence struct.
func (s *BioSequence) SetDefinition(definition string) {
	if definition != "" {
		s.SetAttribute("definition", definition)
	} else {
		s.RemoveAttribute("definition")
	}
}

func (s *BioSequence) RemoveAttribute(key string) {
	if s.HasAnnotation() {
		if s.HasAttribute(key) {
			defer s.AnnotationsUnlock()
			s.AnnotationsLock()
			delete(s.annotations, key)
		}
	}
}

// SetSource sets the source of the BioSequence.
//
// Parameter:
// - source: a string representing the filename without directory name and extension from where the sequence was read.
func (s *BioSequence) SetSource(source string) {
	s.source = source
}

// SetFeatures sets the feature of the BioSequence.
//
// Parameters:
// - feature: a byte slice representing the feature to be set.
//
// No return value.
func (s *BioSequence) SetFeatures(feature []byte) {
	if cap(s.feature) >= 300 {
		RecycleSlice(&s.feature)
	}
	s.feature = feature
}

// SetSequence sets the sequence of the BioSequence.
//
// Parameters:
// - sequence: a byte slice representing the sequence to be set.
func (s *BioSequence) SetSequence(sequence []byte) {
	if s.sequence != nil {
		RecycleSlice(&s.sequence)
	}
	s.sequence = obiutils.InPlaceToLower(CopySlice(sequence))
}

// Setting the qualities of the BioSequence.
func (s *BioSequence) SetQualities(qualities Quality) {
	if s.qualities != nil {
		RecycleSlice(&s.qualities)
	}
	s.qualities = CopySlice(qualities)
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

// WriteByte appends a byte to the BioSequence's sequence.
//
// data: the byte to append to the sequence.
// error: an error if the append operation fails.
func (s *BioSequence) WriteByte(data byte) error {
	s.sequence = append(s.sequence, data)
	return nil
}

// Clear clears the BioSequence by resetting the sequence to an empty slice.
//
// No parameters.
// No return values.
func (s *BioSequence) Clear() {
	s.sequence = s.sequence[0:0]
}

// Composition calculates the composition of the BioSequence.
//
// It counts the occurrences of each nucleotide (a, c, g, t) in the sequence
// and returns a map with the counts.
//
// No parameters.
// Returns a map of byte to int, with the counts of each nucleotide.
func (s *BioSequence) Composition() map[byte]int {
	counts := map[byte]int{
		'a': 0,
		'c': 0,
		'g': 0,
		't': 0,
		'o': 0,
	}

	for _, char := range s.sequence {
		switch char | byte(32) {
		case 'a', 'c', 'g', 't':
			counts[char]++
		default:
			counts['o']++
		}
	}

	return counts
}

func (s *BioSequence) Grow(length int) {
	if s.sequence == nil {
		s.sequence = GetSlice(length)
	} else {
		s.sequence = slices.Grow(s.sequence, length)
	}

	if s.qualities != nil {
		s.qualities = slices.Grow(s.qualities, length)
	}
}
