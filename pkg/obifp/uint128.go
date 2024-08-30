package obifp

import (
	"math"
	"math/bits"

	log "github.com/sirupsen/logrus"
)

type Uint128 struct {
	w1 uint64
	w0 uint64
}

// Zero returns a zero-valued uint128.
//
// No parameters.
// Returns a Uint128 value.
func (u Uint128) Zero() Uint128 {
	return Uint128{w1: 0, w0: 0}
}

// MaxValue returns the maximum possible value for a Uint128.
//
// It returns a Uint128 value with the highest possible values for high and low fields.
func (u Uint128) MaxValue() Uint128 {
	return Uint128{
		w1: math.MaxUint64,
		w0: math.MaxUint64,
	}

}

// IsZero checks if the Uint128 value is zero.
//
// It returns a boolean indicating whether the Uint128 value is zero.
func (u Uint128) IsZero() bool {
	return u.w0 == 0 && u.w1 == 0
}

// Cast a Uint128 to a Uint64.
//
// A Warning will be logged if an overflow occurs.
//
// No parameters.
// Returns a Uint64 value.
func (u Uint128) Uint64() Uint64 {
	if u.w1 != 0 {
		log.Warnf("Uint128 overflow at Uint64(%v)", u)
	}
	return Uint64{w0: u.w0}
}

// Uint128 cast a Uint128 to a Uint128.
//
// Which is a no-op.
//
// No parameters.
// Returns a Uint128 value.
func (u Uint128) Uint128() Uint128 {
	return u
}

// Cast a Uint128 to a Uint256.
//
// A Warning will be logged if an overflow occurs.
//
// No parameters.
// Returns a Uint256 value.
func (u Uint128) Uint256() Uint256 {
	return Uint256{0, 0, u.w1, u.w0}
}

func (u Uint128) Set64(v uint64) Uint128 {

	return Uint128{
		w1: 0,
		w0: v,
	}
}

// LeftShift performs a left shift operation on the Uint128 value by the specified number of bits.
//
// Parameters:
//   - n: the number of bits to shift the Uint128 value to the left.
//
// Returns:
//   - Uint128: the result of the left shift operation.
func (u Uint128) LeftShift(n uint) Uint128 {
	lo, carry := Uint64{w0: u.w0}.LeftShift64(n, 0)
	hi, _ := Uint64{w0: u.w1}.LeftShift64(n, carry)
	return Uint128{w1: hi, w0: lo}
}

// RightShift performs a right shift operation on the Uint128 value by the specified number of bits.
//
// Parameters:
//   - n: the number of bits to shift the Uint128 value to the right.
//
// Returns:
//   - Uint128: the result of the right shift operation.
func (u Uint128) RightShift(n uint) Uint128 {
	hi, carry := Uint64{w0: u.w1}.RightShift64(n, 0)
	lo, _ := Uint64{w0: u.w0}.RightShift64(n, carry)
	return Uint128{w1: hi, w0: lo}
}

// Add performs addition of two Uint128 values and returns the result.
//
// Parameters:
//   - v: the Uint128 value to add to the receiver.
//
// Returns:
//   - Uint128: the result of the addition.
func (u Uint128) Add(v Uint128) Uint128 {
	lo, carry := bits.Add64(u.w0, v.w0, 0)
	hi, carry := bits.Add64(u.w1, v.w1, carry)
	if carry != 0 {
		log.Panicf("Uint128 overflow at Add(%v, %v)", u, v)
	}
	return Uint128{w1: hi, w0: lo}
}

func (u Uint128) Add64(v uint64) Uint128 {
	lo, carry := bits.Add64(u.w0, v, 0)
	hi, carry := bits.Add64(u.w1, 0, carry)
	if carry != 0 {
		log.Panicf("Uint128 overflow at Add64(%v, %v)", u, v)
	}
	return Uint128{w1: hi, w0: lo}
}

func (u Uint128) Sub(v Uint128) Uint128 {
	lo, borrow := bits.Sub64(u.w0, v.w0, 0)
	hi, borrow := bits.Sub64(u.w1, v.w1, borrow)
	if borrow != 0 {
		log.Panicf("Uint128 underflow at Sub(%v, %v)", u, v)
	}
	return Uint128{w1: hi, w0: lo}
}

func (u Uint128) Mul(v Uint128) Uint128 {
	hi, lo := bits.Mul64(u.w0, v.w0)
	p0, p1 := bits.Mul64(u.w1, v.w0)
	p2, p3 := bits.Mul64(u.w0, v.w1)
	hi, c0 := bits.Add64(hi, p1, 0)
	hi, c1 := bits.Add64(hi, p3, c0)
	if p0 != 0 || p2 != 0 || c1 != 0 {
		log.Panicf("Uint128 overflow at Mul(%v, %v)", u, v)
	}
	return Uint128{w1: hi, w0: lo}
}

func (u Uint128) Mul64(v uint64) Uint128 {
	hi, lo := bits.Mul64(u.w0, v)
	p0, p1 := bits.Mul64(u.w1, v)
	hi, c0 := bits.Add64(hi, p1, 0)
	if p0 != 0 || c0 != 0 {
		log.Panicf("Uint128 overflow at Mul64(%v, %v)", u, v)
	}
	return Uint128{w1: hi, w0: lo}
}

func (u Uint128) QuoRem(v Uint128) (q, r Uint128) {
	if v.w1 == 0 {
		var r64 uint64
		q, r64 = u.QuoRem64(v.w0)
		r = Uint128{w1: 0, w0: r64}
	} else {
		// generate a "trial quotient," guaranteed to be within 1 of the actual
		// quotient, then adjust.
		n := uint(bits.LeadingZeros64(v.w1))
		v1 := v.LeftShift(n)
		u1 := u.RightShift(1)
		tq, _ := bits.Div64(u1.w1, u1.w0, v1.w1)
		tq >>= 63 - n
		if tq != 0 {
			tq--
		}
		q = Uint128{w1: 0, w0: tq}
		// calculate remainder using trial quotient, then adjust if remainder is
		// greater than divisor
		r = u.Sub(v.Mul64(tq))
		if r.Cmp(v) >= 0 {
			q = q.Add64(1)
			r = r.Sub(v)
		}
	}
	return
}

// QuoRem64 returns q = u/v and r = u%v.
func (u Uint128) QuoRem64(v uint64) (q Uint128, r uint64) {
	if u.w1 < v {
		q.w0, r = bits.Div64(u.w1, u.w0, v)
	} else {
		q.w1, r = bits.Div64(0, u.w1, v)
		q.w0, r = bits.Div64(r, u.w0, v)
	}
	return
}

func (u Uint128) Div(v Uint128) Uint128 {
	q, _ := u.QuoRem(v)
	return q
}

func (u Uint128) Div64(v uint64) Uint128 {
	q, _ := u.QuoRem64(v)
	return q
}

func (u Uint128) Mod(v Uint128) Uint128 {
	_, r := u.QuoRem(v)
	return r
}

func (u Uint128) Mod64(v uint64) uint64 {
	_, r := u.QuoRem64(v)
	return r
}

func (u Uint128) Cmp(v Uint128) int {
	switch {
	case u.w1 > v.w1:
		return 1
	case u.w1 < v.w1:
		return -1
	case u.w0 > v.w0:
		return 1
	case u.w0 < v.w0:
		return -1
	default:
		return 0
	}
}

func (u Uint128) Cmp64(v uint64) int {
	switch {
	case u.w1 > 0:
		return 1
	case u.w0 > v:
		return 1
	case u.w0 < v:
		return -1
	default:
		return 0
	}
}

func (u Uint128) Equals(v Uint128) bool {
	return u.Cmp(v) == 0
}

func (u Uint128) LessThan(v Uint128) bool {
	return u.Cmp(v) < 0
}

func (u Uint128) GreaterThan(v Uint128) bool {
	return u.Cmp(v) > 0
}

func (u Uint128) LessThanOrEqual(v Uint128) bool {
	return !u.GreaterThan(v)
}

func (u Uint128) GreaterThanOrEqual(v Uint128) bool {
	return !u.LessThan(v)
}

func (u Uint128) And(v Uint128) Uint128 {
	return Uint128{w1: u.w1 & v.w1, w0: u.w0 & v.w0}
}

func (u Uint128) Or(v Uint128) Uint128 {
	return Uint128{w1: u.w1 | v.w1, w0: u.w0 | v.w0}
}

func (u Uint128) Xor(v Uint128) Uint128 {
	return Uint128{w1: u.w1 ^ v.w1, w0: u.w0 ^ v.w0}
}

func (u Uint128) Not() Uint128 {
	return Uint128{w1: ^u.w1, w0: ^u.w0}
}

func (u Uint128) AsUint64() uint64 {
	return u.w0
}
