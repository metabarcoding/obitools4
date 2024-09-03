package obifp

import (
	"math"
	"reflect"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUint128_Add(t *testing.T) {
	u := Uint128{w1: 1, w0: 2}
	v := Uint128{w1: 3, w0: 4}
	w := u.Add(v)
	assert.Equal(t, Uint128{w1: 4, w0: 6}, w)

	u = Uint128{w1: 0, w0: 0}
	v = Uint128{w1: 0, w0: 0}
	w = u.Add(v)
	assert.Equal(t, Uint128{w1: 0, w0: 0}, w)

	u = Uint128{w1: 0, w0: math.MaxUint64}
	v = Uint128{w1: 0, w0: 1}
	w = u.Add(v)
	assert.Equal(t, Uint128{w1: 1, w0: 0}, w)
}

func TestUint128_Add64(t *testing.T) {
	u := Uint128{w1: 1, w0: 2}
	v := uint64(3)
	w := u.Add64(v)
	assert.Equal(t, Uint128{w1: 1, w0: 5}, w)

	u = Uint128{w1: 0, w0: 0}
	v = uint64(0)
	w = u.Add64(v)
	assert.Equal(t, Uint128{w1: 0, w0: 0}, w)

	u = Uint128{w1: 0, w0: math.MaxUint64}
	v = uint64(1)
	w = u.Add64(v)
	assert.Equal(t, Uint128{w1: 1, w0: 0}, w)
}

func TestUint128_Sub(t *testing.T) {
	u := Uint128{w1: 4, w0: 6}
	v := Uint128{w1: 3, w0: 4}
	w := u.Sub(v)
	assert.Equal(t, Uint128{w1: 1, w0: 2}, w)

	u = Uint128{w1: 0, w0: 0}
	v = Uint128{w1: 0, w0: 0}
	w = u.Sub(v)
	assert.Equal(t, Uint128{w1: 0, w0: 0}, w)

	u = Uint128{w1: 1, w0: 0}
	v = Uint128{w1: 0, w0: 1}
	w = u.Sub(v)
	assert.Equal(t, Uint128{w1: 0, w0: math.MaxUint64}, w)
}

func TestUint128_Mul64(t *testing.T) {
	u := Uint128{w1: 1, w0: 2}
	v := uint64(3)
	w := u.Mul64(v)

	if w.w1 != 3 || w.w0 != 6 {
		t.Errorf("Mul64(%v, %v) = %v, want %v", u, v, w, Uint128{w1: 3, w0: 6})
	}

	u = Uint128{w1: 0, w0: 0}
	v = uint64(0)
	w = u.Mul64(v)

	if w.w1 != 0 || w.w0 != 0 {
		t.Errorf("Mul64(%v, %v) = %v, want %v", u, v, w, Uint128{w1: 0, w0: 0})
	}

	u = Uint128{w1: 0, w0: math.MaxUint64}
	v = uint64(2)
	w = u.Mul64(v)

	if w.w1 != 1 || w.w0 != 18446744073709551614 {
		t.Errorf("Mul64(%v, %v) = %v, want %v", u, v, w, Uint128{w1: 1, w0: 18446744073709551614})
	}

}

func TestUint128_Mul(t *testing.T) {
	tests := []struct {
		name     string
		u        Uint128
		v        Uint128
		expected Uint128
	}{
		{
			name:     "simple multiplication",
			u:        Uint128{w1: 1, w0: 2},
			v:        Uint128{w1: 3, w0: 4},
			expected: Uint128{w1: 10, w0: 8},
		},
		{
			name:     "multiplication with overflow",
			u:        Uint128{w1: 0, w0: math.MaxUint64},
			v:        Uint128{w1: 0, w0: 2},
			expected: Uint128{w1: 1, w0: 18446744073709551614},
		},
		{
			name:     "multiplication with zero",
			u:        Uint128{w1: 0, w0: 0},
			v:        Uint128{w1: 0, w0: 0},
			expected: Uint128{w1: 0, w0: 0},
		},
		{
			name:     "multiplication with large numbers",
			u:        Uint128{w1: 100, w0: 200},
			v:        Uint128{w1: 300, w0: 400},
			expected: Uint128{w1: 100000, w0: 80000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.u.Mul(tt.v)
			if !reflect.DeepEqual(actual, tt.expected) {
				t.Errorf("Mul(%v, %v) = %v, want %v", tt.u, tt.v, actual, tt.expected)
			}
		})
	}
}

func TestUint128_QuoRem(t *testing.T) {
	u := Uint128{w1: 3, w0: 8}
	v := Uint128{w1: 0, w0: 4}
	q, r := u.QuoRem(v)
	assert.Equal(t, Uint128{w1: 0, w0: 2}, q)
	assert.Equal(t, Uint128{w1: 0, w0: 0}, r)
}

func TestUint128_QuoRem64(t *testing.T) {
	u := Uint128{w1: 0, w0: 6}
	v := uint64(3)
	q, r := u.QuoRem64(v)
	assert.Equal(t, Uint128{w1: 0, w0: 2}, q)
	assert.Equal(t, uint64(0), r)
}

func TestUint128_Div(t *testing.T) {
	u := Uint128{w1: 3, w0: 8}
	v := Uint128{w1: 0, w0: 4}
	q := u.Div(v)
	assert.Equal(t, Uint128{w1: 0, w0: 2}, q)
}

func TestUint128_Div64(t *testing.T) {
	u := Uint128{w1: 0, w0: 6}
	v := uint64(3)
	q := u.Div64(v)
	assert.Equal(t, Uint128{w1: 0, w0: 2}, q)
}

func TestUint128_Mod(t *testing.T) {
	u := Uint128{w1: 3, w0: 8}
	v := Uint128{w1: 0, w0: 4}
	r := u.Mod(v)
	assert.Equal(t, Uint128{w1: 0, w0: 0}, r)
}

func TestUint128_Mod64(t *testing.T) {
	u := Uint128{w1: 0, w0: 6}
	v := uint64(3)
	r := u.Mod64(v)
	assert.Equal(t, uint64(0), r)
}

func TestUint128_Cmp(t *testing.T) {
	u := Uint128{w1: 1, w0: 2}
	v := Uint128{w1: 3, w0: 4}
	assert.Equal(t, -1, u.Cmp(v))
}

func TestUint128_Cmp64(t *testing.T) {
	u := Uint128{w1: 1, w0: 2}
	v := uint64(3)
	assert.Equal(t, -1, u.Cmp64(v))
}

func TestUint128_Equals(t *testing.T) {
	u := Uint128{w1: 1, w0: 2}
	v := Uint128{w1: 1, w0: 2}
	assert.Equal(t, true, u.Equals(v))
}

func TestUint128_LessThan(t *testing.T) {
	u := Uint128{w1: 1, w0: 2}
	v := Uint128{w1: 3, w0: 4}
	assert.Equal(t, true, u.LessThan(v))
}

func TestUint128_GreaterThan(t *testing.T) {
	u := Uint128{w1: 1, w0: 2}
	v := Uint128{w1: 3, w0: 4}
	assert.Equal(t, false, u.GreaterThan(v))
}

func TestUint128_LessThanOrEqual(t *testing.T) {
	u := Uint128{w1: 1, w0: 2}
	v := Uint128{w1: 3, w0: 4}
	assert.Equal(t, true, u.LessThanOrEqual(v))
}

func TestUint128_GreaterThanOrEqual(t *testing.T) {
	u := Uint128{w1: 1, w0: 2}
	v := Uint128{w1: 3, w0: 4}
	assert.Equal(t, false, u.GreaterThanOrEqual(v))
}

func TestUint128_And(t *testing.T) {
	u := Uint128{w1: 1, w0: 2}
	v := Uint128{w1: 3, w0: 4}
	w := u.And(v)
	assert.Equal(t, Uint128{w1: 1, w0: 0}, w)
}

func TestUint128_Or(t *testing.T) {
	u := Uint128{w1: 1, w0: 2}
	v := Uint128{w1: 3, w0: 4}
	w := u.Or(v)
	assert.Equal(t, Uint128{w1: 3, w0: 6}, w)
}

func TestUint128_Xor(t *testing.T) {
	u := Uint128{w1: 1, w0: 2}
	v := Uint128{w1: 3, w0: 4}
	w := u.Xor(v)
	assert.Equal(t, Uint128{w1: 2, w0: 6}, w)
}

func TestUint128_Not(t *testing.T) {
	u := Uint128{w1: 1, w0: 2}
	w := u.Not()
	assert.Equal(t, Uint128{w1: math.MaxUint64 - 1, w0: math.MaxUint64 - 2}, w)
}

func TestUint128_AsUint64(t *testing.T) {
	u := Uint128{w1: 0, w0: 5}
	v := u.AsUint64()
	assert.Equal(t, uint64(5), v)
}
