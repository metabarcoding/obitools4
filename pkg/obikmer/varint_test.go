package obikmer

import (
	"bytes"
	"testing"
)

func TestVarintRoundTrip(t *testing.T) {
	values := []uint64{
		0, 1, 127, 128, 255, 256,
		16383, 16384,
		1<<21 - 1, 1 << 21,
		1<<28 - 1, 1 << 28,
		1<<35 - 1, 1 << 35,
		1<<42 - 1, 1 << 42,
		1<<49 - 1, 1 << 49,
		1<<56 - 1, 1 << 56,
		1<<63 - 1, 1 << 63,
		^uint64(0), // max uint64
	}

	for _, v := range values {
		var buf bytes.Buffer
		n, err := EncodeVarint(&buf, v)
		if err != nil {
			t.Fatalf("EncodeVarint(%d): %v", v, err)
		}
		if n != VarintLen(v) {
			t.Fatalf("EncodeVarint(%d): wrote %d bytes, VarintLen says %d", v, n, VarintLen(v))
		}

		decoded, err := DecodeVarint(&buf)
		if err != nil {
			t.Fatalf("DecodeVarint for %d: %v", v, err)
		}
		if decoded != v {
			t.Fatalf("roundtrip failed: encoded %d, decoded %d", v, decoded)
		}
	}
}

func TestVarintLen(t *testing.T) {
	tests := []struct {
		value    uint64
		expected int
	}{
		{0, 1},
		{127, 1},
		{128, 2},
		{16383, 2},
		{16384, 3},
		{^uint64(0), 10},
	}

	for _, tc := range tests {
		got := VarintLen(tc.value)
		if got != tc.expected {
			t.Errorf("VarintLen(%d) = %d, want %d", tc.value, got, tc.expected)
		}
	}
}

func TestVarintSequence(t *testing.T) {
	var buf bytes.Buffer
	values := []uint64{0, 42, 1000000, ^uint64(0), 1}

	for _, v := range values {
		if _, err := EncodeVarint(&buf, v); err != nil {
			t.Fatalf("EncodeVarint(%d): %v", v, err)
		}
	}

	for _, expected := range values {
		got, err := DecodeVarint(&buf)
		if err != nil {
			t.Fatalf("DecodeVarint: %v", err)
		}
		if got != expected {
			t.Errorf("got %d, want %d", got, expected)
		}
	}
}
