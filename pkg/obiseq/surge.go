package obiseq

import "github.com/renproject/surge"

func (sequence BioSequence) SizeHint() int {
	return surge.SizeHintString(sequence.sequence.id) +
		surge.SizeHintString(sequence.sequence.definition) +
		surge.SizeHintBytes(sequence.sequence.sequence) +
		surge.SizeHintBytes(sequence.sequence.qualities) +
		surge.SizeHintBytes(sequence.sequence.feature) +
		surge.SizeHint(sequence.sequence.annotations)
}

func (sequence BioSequence) Marshal(buf []byte, rem int) ([]byte, int, error) {
	var err error
	if buf, rem, err = surge.MarshalString(sequence.sequence.id, buf, rem); err != nil {
		return buf, rem, err
	}
	if buf, rem, err = surge.MarshalString(sequence.sequence.definition, buf, rem); err != nil {
		return buf, rem, err
	}
	if buf, rem, err = surge.MarshalBytes(sequence.sequence.sequence, buf, rem); err != nil {
		return buf, rem, err
	}
	if buf, rem, err = surge.MarshalBytes(sequence.sequence.qualities, buf, rem); err != nil {
		return buf, rem, err
	}
	if buf, rem, err = surge.MarshalBytes(sequence.sequence.feature, buf, rem); err != nil {
		return buf, rem, err
	}
	if buf, rem, err = surge.Marshal(sequence.sequence.annotations, buf, rem); err != nil {
		return buf, rem, err
	}

	return buf, rem, err
}

// Unmarshal is the opposite of Marshal, and requires
// a pointer receiver.
func (sequence *BioSequence) Unmarshal(buf []byte, rem int) ([]byte, int, error) {
	var err error
	if buf, rem, err = surge.UnmarshalString(&(sequence.sequence.id), buf, rem); err != nil {
		return buf, rem, err
	}
	if buf, rem, err = surge.UnmarshalString(&(sequence.sequence.definition), buf, rem); err != nil {
		return buf, rem, err
	}
	if buf, rem, err = surge.UnmarshalBytes(&(sequence.sequence.sequence), buf, rem); err != nil {
		return buf, rem, err
	}
	if buf, rem, err = surge.UnmarshalBytes(&(sequence.sequence.qualities), buf, rem); err != nil {
		return buf, rem, err
	}
	if buf, rem, err = surge.UnmarshalBytes(&(sequence.sequence.feature), buf, rem); err != nil {
		return buf, rem, err
	}
	if buf, rem, err = surge.Unmarshal(&(sequence.sequence.annotations), buf, rem); err != nil {
		return buf, rem, err
	}
	return buf, rem, err
}

func (sequences BioSequenceSlice) SizeHint() int {
	size := surge.SizeHintI64
	for _, s := range sequences {
		size += s.SizeHint()
	}

	return size
}

func (sequences BioSequenceSlice) Marshal(buf []byte, rem int) ([]byte, int, error) {
	var err error

	if buf, rem, err = surge.MarshalI64(int64(len(sequences)), buf, rem); err != nil {
		return buf, rem, err
	}

	for _, s := range sequences {
		if buf, rem, err = s.Marshal(buf, rem); err != nil {
			return buf, rem, err
		}
	}

	return buf, rem, err
}

func (sequences *BioSequenceSlice) Unmarshal(buf []byte, rem int) ([]byte, int, error) {
	var err error
	var length int64

	if buf, rem, err = surge.UnmarshalI64(&length, buf, rem); err != nil {
		return buf, rem, err
	}

	*sequences = make(BioSequenceSlice, length)

	for i := 0; i < int(length); i++ {
		if buf, rem, err = ((*sequences)[i]).Unmarshal(buf, rem); err != nil {
			return buf, rem, err
		}
	}

	return buf, rem, err
}
