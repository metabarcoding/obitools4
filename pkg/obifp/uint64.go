package obifp

import (
	"math"
	"math/bits"

	log "github.com/sirupsen/logrus"
)

type Uint64 struct {
	w0 uint64
}

// Zero returns a zero value of type Uint64.
//
// No parameters.
// Returns a Uint64 value of 0.
func (u Uint64) Zero() Uint64 {
	return Uint64{0}
}

// MaxValue returns the maximum possible value of type Uint64.
//
// No parameters.
// Returns the maximum value of type Uint64.
func (u Uint64) MaxValue() Uint64 {
	return Uint64{math.MaxUint64}
}

// IsZero checks if the Uint64 value is zero.
//
// No parameters.
// Returns a boolean indicating if the value is zero.
func (u Uint64) IsZero() bool {
	return u.w0 == 0
}

// Cast a Uint64 to a Uint64.
//
// Which is a no-op.
//
// No parameters.
// Returns the Uint64 value itself.
func (u Uint64) Uint64() Uint64 {
	return u
}

// Cast a Uint64 to a Uint128.
//
// No parameters.
// Returns a Uint128 value with the high field set to 0 and the low field set to the value of the Uint64.
func (u Uint64) Uint128() Uint128 {
	return Uint128{w1: 0, w0: u.w0}
}

// Cast a Uint64 to a Uint256.
//
// No parameters.
// Returns a Uint256 value with the high fields set to 0 and the low fields set to the value of the Uint64.
func (u Uint64) Uint256() Uint256 {
	return Uint256{w3: 0, w2: 0, w1: 0, w0: u.w0}
}

func (u Uint64) Set64(v uint64) Uint64 {

	return Uint64{
		w0: v,
	}
}

// LeftShift64 performs a left shift operation on the Uint64 value by n bits, with carry-in from carryIn.
//
// The carry-in value is used as the first bit of the shifted value.
//
// The function returns u << n | (carryIn & ((1 << n) - 1)).
//
// This is a shift left operation, lowest bits are set with the lowest bits of
// the carry-in value instead of 0 as they would be in classical a left shift operation.
//
// Parameters:
// - n: the number of bits to shift by.
// - carryIn: the carry-in value.
//
// Returns:
// - value: the result of the left shift operation.
// - carry: the carry-out value.
func (u Uint64) LeftShift64(n uint, carryIn uint64) (value, carry uint64) {
	switch {
	case n == 0:
		return u.w0, 0

	case n < 64:
		return u.w0<<n | (carryIn & ((1 << n) - 1)), u.w0 >> (64 - n)

	case n == 64:
		return carryIn, u.w0
	}

	log.Warnf("Uint64 overflow at LeftShift64(%v, %v)", u, n)
	return 0, 0

}

// RightShift64 performs a right shift operation on the Uint64 value by n bits, with carry-out to carry.
//
// The function returns the result of the right shift operation and the carry-out value.
//
// Parameters:
// - n: the number of bits to shift by.
//
// Returns:
// - value: the result of the right shift operation.
// - carry: the carry-out value.
func (u Uint64) RightShift64(n uint, carryIn uint64) (value, carry uint64) {
	switch {
	case n == 0:
		return u.w0, 0

	case n < 64:
		return u.w0>>n | (carryIn & ^((1 << (64 - n)) - 1)), u.w0 << (n - 64)

	case n == 64:
		return carryIn, u.w0
	}

	log.Warnf("Uint64 overflow at RightShift64(%v, %v)", u, n)
	return 0, 0
}

func (u Uint64) Add64(v Uint64, carryIn uint64) (value, carry uint64) {
	return bits.Add64(u.w0, v.w0, uint64(carryIn))
}

func (u Uint64) Sub64(v Uint64, carryIn uint64) (value, carry uint64) {
	return bits.Sub64(u.w0, v.w0, uint64(carryIn))
}

func (u Uint64) Mul64(v Uint64) (value, carry uint64) {
	return bits.Mul64(u.w0, v.w0)
}

func (u Uint64) LeftShift(n uint) Uint64 {
	sl, _ := u.LeftShift64(n, 0)
	return Uint64{w0: sl}
}

func (u Uint64) RightShift(n uint) Uint64 {
	sr, _ := u.RightShift64(n, 0)
	return Uint64{w0: sr}
}

func (u Uint64) Add(v Uint64) Uint64 {
	value, carry := u.Add64(v, 0)

	if carry != 0 {
		log.Panicf("Uint64 overflow at Add(%v, %v)", u, v)
	}

	return Uint64{w0: value}
}

func (u Uint64) Sub(v Uint64) Uint64 {
	value, carry := u.Sub64(v, 0)

	if carry != 0 {
		log.Panicf("Uint64 overflow at Sub(%v, %v)", u, v)
	}

	return Uint64{w0: value}
}

func (u Uint64) Mul(v Uint64) Uint64 {
	value, carry := u.Mul64(v)

	if carry != 0 {
		log.Panicf("Uint64 overflow at Mul(%v, %v)", u, v)
	}

	return Uint64{w0: value}
}

func (u Uint64) Cmp(v Uint64) int {
	switch {
	case u.w0 < v.w0:
		return -1
	case u.w0 > v.w0:
		return 1
	default:
		return 0
	}
}

func (u Uint64) Equals(v Uint64) bool {
	return u.Cmp(v) == 0
}

func (u Uint64) LessThan(v Uint64) bool {
	return u.Cmp(v) < 0
}

func (u Uint64) GreaterThan(v Uint64) bool {
	return u.Cmp(v) > 0
}

func (u Uint64) LessThanOrEqual(v Uint64) bool {
	return !u.GreaterThan(v)
}

func (u Uint64) GreaterThanOrEqual(v Uint64) bool {
	return !u.LessThan(v)
}

func (u Uint64) And(v Uint64) Uint64 {
	return Uint64{w0: u.w0 & v.w0}
}

func (u Uint64) Or(v Uint64) Uint64 {
	return Uint64{w0: u.w0 | v.w0}
}

func (u Uint64) Xor(v Uint64) Uint64 {
	return Uint64{w0: u.w0 ^ v.w0}
}

func (u Uint64) Not() Uint64 {
	return Uint64{w0: ^u.w0}
}

func (u Uint64) AsUint64() uint64 {
	return u.w0
}
