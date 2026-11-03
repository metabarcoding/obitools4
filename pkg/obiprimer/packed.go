package obiprimer

import (
	"encoding/binary"
	"slices"
)

// packedSeq holds the bases of a sequence on 2 bits each, a quarter of the
// bytes of the sequence itself. The bytes that are not a, c, g or t (N, the
// IUPAC codes, u...) are kept apart in runs, so that unpacking gives back the
// sequence byte for byte.
type packedSeq struct {
	n    int
	bits []byte   // 4 bases per byte, the first in the low bits
	exc  []excRun // sorted by position
}

// excRun is a run of length identical bytes b that are not a, c, g or t,
// starting at pos.
type excRun struct {
	pos, length uint32
	b           byte
}

// packCode maps a, c, g, t to 0..3 and every other byte to 4.
var packCode = func() (t [256]byte) {
	for i := range t {
		t[i] = 4
	}
	for i, c := range "acgt" {
		t[c] = byte(i)
	}
	return
}()

// unpackWord gives, for a byte of 4 packed bases, the 4 bytes they stand for.
var unpackWord = func() (t [256]uint32) {
	for b := range t {
		for k := range 4 {
			t[b] |= uint32("acgt"[(b>>(2*k))&3]) << (8 * k)
		}
	}
	return
}()

// packBases packs raw (lower case).
func packBases(raw []byte) packedSeq {
	p := packedSeq{n: len(raw), bits: make([]byte, (len(raw)+3)/4)}
	for i, b := range raw {
		code := packCode[b]
		if code > 3 {
			if k := len(p.exc) - 1; k >= 0 && p.exc[k].b == b && int(p.exc[k].pos)+int(p.exc[k].length) == i {
				p.exc[k].length++
			} else {
				p.exc = append(p.exc, excRun{uint32(i), 1, b})
			}
			code = 0
		}
		p.bits[i>>2] |= code << (2 * (i & 3))
	}
	p.exc = slices.Clone(p.exc) // the exact size, nil when there is none
	return p
}

// unpack writes the sequence into dst, which it grows if need be, and returns it.
func (p packedSeq) unpack(dst []byte) []byte {
	dst = slices.Grow(dst[:0], p.n)[:p.n]
	for i := range p.n / 4 {
		binary.LittleEndian.PutUint32(dst[4*i:], unpackWord[p.bits[i]])
	}
	for t := p.n &^ 3; t < p.n; t++ {
		dst[t] = "acgt"[(p.bits[t>>2]>>(2*(t&3)))&3]
	}
	for _, e := range p.exc {
		for k := range e.length {
			dst[e.pos+k] = e.b
		}
	}
	return dst
}
