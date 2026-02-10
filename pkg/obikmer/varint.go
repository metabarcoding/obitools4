package obikmer

import "io"

// EncodeVarint writes a uint64 value as a variable-length integer to w.
// Uses 7 bits per byte with the high bit as a continuation flag
// (identical to protobuf unsigned varint encoding).
// Returns the number of bytes written.
func EncodeVarint(w io.Writer, v uint64) (int, error) {
	var buf [10]byte // max 10 bytes for uint64 varint
	n := 0
	for v >= 0x80 {
		buf[n] = byte(v) | 0x80
		v >>= 7
		n++
	}
	buf[n] = byte(v)
	n++
	return w.Write(buf[:n])
}

// DecodeVarint reads a variable-length encoded uint64 from r.
// Returns the decoded value and any error encountered.
func DecodeVarint(r io.Reader) (uint64, error) {
	var val uint64
	var shift uint
	var buf [1]byte

	for {
		if _, err := io.ReadFull(r, buf[:]); err != nil {
			return 0, err
		}
		b := buf[0]
		val |= uint64(b&0x7F) << shift
		if b < 0x80 {
			return val, nil
		}
		shift += 7
		if shift >= 70 {
			return 0, io.ErrUnexpectedEOF
		}
	}
}

// VarintLen returns the number of bytes needed to encode v as a varint.
func VarintLen(v uint64) int {
	n := 1
	for v >= 0x80 {
		v >>= 7
		n++
	}
	return n
}
