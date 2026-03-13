package obiutils

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// ParseMemSize parses a human-readable memory size string and returns the
// equivalent number of bytes. The value is a number optionally followed by a
// unit suffix (case-insensitive):
//
//	B  or (no suffix) — bytes
//	K  or KB           — kibibytes  (1 024)
//	M  or MB           — mebibytes  (1 048 576)
//	G  or GB           — gibibytes  (1 073 741 824)
//	T  or TB           — tebibytes  (1 099 511 627 776)
//
// Examples: "512", "128K", "128k", "64M", "1G", "2GB"
func ParseMemSize(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty memory size string")
	}

	// split numeric prefix from unit suffix
	i := 0
	for i < len(s) && (unicode.IsDigit(rune(s[i])) || s[i] == '.') {
		i++
	}
	numStr := s[:i]
	unit := strings.ToUpper(strings.TrimSpace(s[i:]))
	// strip trailing 'B' from two-letter units (KB→K, MB→M …)
	if len(unit) == 2 && unit[1] == 'B' {
		unit = unit[:1]
	}

	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid memory size %q: %w", s, err)
	}

	var multiplier float64
	switch unit {
	case "", "B":
		multiplier = 1
	case "K":
		multiplier = 1024
	case "M":
		multiplier = 1024 * 1024
	case "G":
		multiplier = 1024 * 1024 * 1024
	case "T":
		multiplier = 1024 * 1024 * 1024 * 1024
	default:
		return 0, fmt.Errorf("unknown memory unit %q in %q", unit, s)
	}

	return int(val * multiplier), nil
}

// FormatMemSize formats a byte count as a human-readable string with the
// largest unit that produces a value ≥ 1 (e.g. 1536 → "1.5K").
func FormatMemSize(n int) string {
	units := []struct {
		suffix string
		size   int
	}{
		{"T", 1024 * 1024 * 1024 * 1024},
		{"G", 1024 * 1024 * 1024},
		{"M", 1024 * 1024},
		{"K", 1024},
	}
	for _, u := range units {
		if n >= u.size {
			v := float64(n) / float64(u.size)
			if v == float64(int(v)) {
				return fmt.Sprintf("%d%s", int(v), u.suffix)
			}
			return fmt.Sprintf("%.1f%s", v, u.suffix)
		}
	}
	return fmt.Sprintf("%dB", n)
}
