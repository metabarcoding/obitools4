package obifp

import (
	"math"
	"math/bits"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obilog"
	log "github.com/sirupsen/logrus"
)

type Uint256 struct {
	w3 uint64
	w2 uint64
	w1 uint64
	w0 uint64
}

// Zero returns a zero value of type Uint256.
//
// No parameters.
// Returns a Uint256 value of 0.
func (u Uint256) Zero() Uint256 {
	return Uint256{}
}

// MaxValue returns the maximum possible value of type Uint256.
//
// No parameters.
// Returns the maximum value of type Uint256.
func (u Uint256) MaxValue() Uint256 {
	return Uint256{
		w3: math.MaxUint64,
		w2: math.MaxUint64,
		w1: math.MaxUint64,
		w0: math.MaxUint64,
	}
}

// IsZero checks if the Uint256 value is zero.
//
// No parameters.
// Returns a boolean indicating if the value is zero.
func (u Uint256) IsZero() bool {
	return u == Uint256{}
}

// Cast a Uint256 to a Uint64.
//
// A Warning will be logged if an overflow occurs.
//
// No parameters.
// Returns a Uint64 value.
func (u Uint256) Uint64() Uint64 {
	if u.w3 != 0 || u.w2 != 0 || u.w1 != 0 {
		obilog.Warnf("Uint256 overflow at Uint64(%v)", u)
	}
	return Uint64{w0: u.w0}
}

// Cast a Uint256 to a Uint128.
//
// A Warning will be logged if an overflow occurs.
//
// No parameters.
// Returns a Uint128 value.
func (u Uint256) Uint128() Uint128 {
	if u.w3 != 0 || u.w2 != 0 {
		obilog.Warnf("Uint256 overflow at Uint128(%v)", u)
	}
	return Uint128{u.w1, u.w0}
}

// Cast a Uint128 to a Uint256.
//
// A Warning will be logged if an overflow occurs.
//
// No parameters.
// Returns a Uint256 value.
func (u Uint256) Uint256() Uint256 {
	return u
}

func (u Uint256) Set64(v uint64) Uint256 {

	return Uint256{
		w3: 0,
		w2: 0,
		w1: 0,
		w0: v,
	}
}

func (u Uint256) LeftShift(n uint) Uint256 {
	w0, carry := Uint64{w0: u.w0}.LeftShift64(n, 0)
	w1, carry := Uint64{w0: u.w1}.LeftShift64(n, carry)
	w2, carry := Uint64{w0: u.w2}.LeftShift64(n, carry)
	w3, _ := Uint64{w0: u.w3}.LeftShift64(n, carry)
	return Uint256{w3, w2, w1, w0}
}

func (u Uint256) RightShift(n uint) Uint256 {
	w3, carry := Uint64{w0: u.w3}.RightShift64(n, 0)
	w2, carry := Uint64{w0: u.w2}.RightShift64(n, carry)
	w1, carry := Uint64{w0: u.w1}.RightShift64(n, carry)
	w0, _ := Uint64{w0: u.w0}.RightShift64(n, carry)
	return Uint256{w3, w2, w1, w0}
}

func (u Uint256) Cmp(v Uint256) int {
	switch {
	case u.w3 > v.w3:
		return 1
	case u.w3 < v.w3:
		return -1
	case u.w2 > v.w2:
		return 1
	case u.w2 < v.w2:
		return -1
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

// Add performs addition of two Uint256 values and returns the result.
//
// Parameters:
//   - v: the Uint256 value to add to the receiver.
//
// Returns:
//   - Uint256: the result of the addition.
func (u Uint256) Add(v Uint256) Uint256 {
	w0, carry := bits.Add64(u.w0, v.w0, 0)
	w1, carry := bits.Add64(u.w1, v.w1, carry)
	w2, carry := bits.Add64(u.w2, v.w2, carry)
	w3, carry := bits.Add64(u.w3, v.w3, carry)
	if carry != 0 {
		log.Panicf("Uint256 overflow at Add(%v, %v)", u, v)
	}
	return Uint256{w3, w2, w1, w0}
}

// Sub performs subtraction of two Uint256 values and returns the result.
//
// Parameters:
//   - v: the Uint256 value to subtract from the receiver.
//
// Returns:
//   - Uint256: the result of the subtraction.
func (u Uint256) Sub(v Uint256) Uint256 {
	w0, borrow := bits.Sub64(u.w0, v.w0, 0)
	w1, borrow := bits.Sub64(u.w1, v.w1, borrow)
	w2, borrow := bits.Sub64(u.w2, v.w2, borrow)
	w3, borrow := bits.Sub64(u.w3, v.w3, borrow)
	if borrow != 0 {
		log.Panicf("Uint256 overflow at Sub(%v, %v)", u, v)
	}
	return Uint256{w3, w2, w1, w0}
}

// Mul performs multiplication of two Uint256 values and returns the result.
//
// Parameters:
//   - v: the Uint256 value to multiply with the receiver.
//
// Returns:
//   - Uint256: the result of the multiplication.
func (u Uint256) Mul(v Uint256) Uint256 {
	var w0, w1, w2, w3, carry uint64

	w0Low, w0High := bits.Mul64(u.w0, v.w0)
	w1Low1, w1High1 := bits.Mul64(u.w0, v.w1)
	w1Low2, w1High2 := bits.Mul64(u.w1, v.w0)
	w2Low1, w2High1 := bits.Mul64(u.w0, v.w2)
	w2Low2, w2High2 := bits.Mul64(u.w1, v.w1)
	w2Low3, w2High3 := bits.Mul64(u.w2, v.w0)
	w3Low1, w3High1 := bits.Mul64(u.w0, v.w3)
	w3Low2, w3High2 := bits.Mul64(u.w1, v.w2)
	w3Low3, w3High3 := bits.Mul64(u.w2, v.w1)
	w3Low4, w3High4 := bits.Mul64(u.w3, v.w0)

	w0 = w0Low

	w1, carry = bits.Add64(w1Low1, w1Low2, 0)
	w1, _ = bits.Add64(w1, w0High, carry)

	w2, carry = bits.Add64(w2Low1, w2Low2, 0)
	w2, carry = bits.Add64(w2, w2Low3, carry)
	w2, carry = bits.Add64(w2, w1High1, carry)
	w2, _ = bits.Add64(w2, w1High2, carry)

	w3, carry = bits.Add64(w3Low1, w3Low2, 0)
	w3, carry = bits.Add64(w3, w3Low3, carry)
	w3, carry = bits.Add64(w3, w3Low4, carry)
	w3, carry = bits.Add64(w3, w2High1, carry)
	w3, carry = bits.Add64(w3, w2High2, carry)
	w3, carry = bits.Add64(w3, w2High3, carry)

	if w3High1 != 0 || w3High2 != 0 || w3High3 != 0 || w3High4 != 0 || carry != 0 {
		log.Panicf("Uint256 overflow at Mul(%v, %v)", u, v)
	}

	return Uint256{w3, w2, w1, w0}
}

// Div performs division of two Uint256 values and returns the result.
//
// Parameters:
//   - v: the Uint256 value to divide with the receiver.
//
// Returns:
//   - Uint256: the result of the division.
func (u Uint256) Div(v Uint256) Uint256 {
	if v.IsZero() {
		log.Panicf("division by zero")
	}

	if u.IsZero() || u.LessThan(v) {
		return Uint256{}
	}

	if v.Equals(Uint256{0, 0, 0, 1}) {
		return u // Division by 1
	}

	var q, r Uint256
	r = u

	for r.GreaterThanOrEqual(v) {
		var t Uint256 = v
		var m Uint256 = Uint256{0, 0, 0, 1}
		for t.LeftShift(1).LessThanOrEqual(r) {
			t = t.LeftShift(1)
			m = m.LeftShift(1)
		}
		r = r.Sub(t)
		q = q.Add(m)
	}

	return q
}

func (u Uint256) Equals(v Uint256) bool {
	return u.Cmp(v) == 0
}

func (u Uint256) LessThan(v Uint256) bool {
	return u.Cmp(v) < 0
}

func (u Uint256) GreaterThan(v Uint256) bool {
	return u.Cmp(v) > 0
}

func (u Uint256) LessThanOrEqual(v Uint256) bool {
	return !u.GreaterThan(v)
}

func (u Uint256) GreaterThanOrEqual(v Uint256) bool {
	return !u.LessThan(v)
}

func (u Uint256) And(v Uint256) Uint256 {
	return Uint256{
		w3: u.w3 & v.w3,
		w2: u.w2 & v.w2,
		w1: u.w1 & v.w1,
		w0: u.w0 & v.w0,
	}
}

func (u Uint256) Or(v Uint256) Uint256 {
	return Uint256{
		w3: u.w3 | v.w3,
		w2: u.w2 | v.w2,
		w1: u.w1 | v.w1,
		w0: u.w0 | v.w0,
	}
}

func (u Uint256) Xor(v Uint256) Uint256 {
	return Uint256{
		w3: u.w3 ^ v.w3,
		w2: u.w2 ^ v.w2,
		w1: u.w1 ^ v.w1,
		w0: u.w0 ^ v.w0,
	}
}

func (u Uint256) Not() Uint256 {
	return Uint256{
		w3: ^u.w3,
		w2: ^u.w2,
		w1: ^u.w1,
		w0: ^u.w0,
	}
}

func (u Uint256) AsUint64() uint64 {
	return u.w0
}
