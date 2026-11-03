package obithermo

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

func readTSV(t *testing.T, path string) [][]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var rows [][]string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" {
			rows = append(rows, strings.Split(line, "\t"))
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return rows
}

func mustFloat(t *testing.T, s string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// checkAgainstC compares one Go result (Celsius) with the raw Kelvin value
// returned by the C library. The C code returns 0 when the model fails.
//
// The tolerance (1e-4 K) covers the single-precision literals of the C source
// (-21.3f, 1.987f, -5.9f ...): emulating them in float32 brings the gap down
// to 1e-12 K, so the Go port uses the same model with exact decimal constants.
func checkAgainstC(t *testing.T, what string, got float64, cKelvin float64) {
	t.Helper()
	if cKelvin <= 0 {
		if !math.IsNaN(got) {
			t.Errorf("%s: C model failed, Go returned %g", what, got)
		}
		return
	}
	if math.IsNaN(got) {
		t.Errorf("%s: Go model failed, C returned %g K", what, cKelvin)
		return
	}
	if d := math.Abs(got + 273.15 - cKelvin); d > 1e-4 {
		t.Errorf("%s: Go %.9f C %.9f K (diff %g)", what, got+273.15, cKelvin, d)
	}
}

// testdata/c_reference.tsv was produced by calling nparam_CalcSelfTM and
// nparam_CalcTwoTM of ecoPrimers' libthermo on random primers and sites with
// 0..n mismatches: primer, site, self Tm, two Tm (default conditions), and
// two Tm with 1e-6 M / 2e-7 M strands and 0.2 M salt.
func TestAgainstEcoPrimersC(t *testing.T) {
	def := DefaultModel()
	alt, err := NewModel(1e-6, 2e-7, 0.2, SantaLucia)
	if err != nil {
		t.Fatal(err)
	}

	rows := readTSV(t, "testdata/c_reference.tsv")
	if len(rows) < 100 {
		t.Fatalf("reference file too short: %d", len(rows))
	}
	for _, r := range rows {
		primer, site := []byte(r[0]), []byte(r[1])

		got, err := def.SelfTm(primer)
		if err != nil {
			t.Fatal(err)
		}
		checkAgainstC(t, "self "+r[0], got, mustFloat(t, r[2]))

		got, err = def.Tm(primer, site)
		if err != nil {
			t.Fatal(err)
		}
		checkAgainstC(t, "two "+r[0]+"/"+r[1], got, mustFloat(t, r[3]))

		got, err = alt.Tm(primer, site)
		if err != nil {
			t.Fatal(err)
		}
		checkAgainstC(t, "alt "+r[0]+"/"+r[1], got, mustFloat(t, r[4]))
	}
}

// ecoPrimers prints Tm as (Kelvin - 273.0) with one decimal; obithermo uses the
// exact 273.15 offset, so the printed value is our result + 0.15.
func TestAgainstTutorialOutput(t *testing.T) {
	m := DefaultModel()
	rows := readTSV(t, "testdata/ecoprimers_tm.tsv")
	if len(rows) == 0 {
		t.Fatal("empty fixture")
	}
	for _, r := range rows {
		got, err := m.SelfTm([]byte(r[0]))
		if err != nil {
			t.Fatal(err)
		}
		want := mustFloat(t, r[1])
		if d := math.Abs(got + 0.15 - want); d > 0.0501 {
			t.Errorf("%s: printed %.1f, got %.3f (+0.15 = %.3f)", r[0], want, got, got+0.15)
		}
	}
}

func TestSaltMethodsAgreeAtOneMolar(t *testing.T) {
	sl, _ := NewModel(DefaultPrimerConc, DefaultTargetConc, 1.0, SantaLucia)
	ow, _ := NewModel(DefaultPrimerConc, DefaultTargetConc, 1.0, Owczarzy)
	p := []byte("ACACCGCCCGTCACTCTC")
	a, _ := sl.SelfTm(p)
	b, _ := ow.SelfTm(p)
	if math.Abs(a-b) > 1e-9 {
		t.Errorf("at 1 M both corrections vanish: %g vs %g", a, b)
	}
}

func TestSaltRaisesTm(t *testing.T) {
	p := []byte("ACACCGCCCGTCACTCTC")
	for _, method := range []SaltMethod{SantaLucia, Owczarzy} {
		low, _ := NewModel(DefaultPrimerConc, DefaultTargetConc, 0.02, method)
		high, _ := NewModel(DefaultPrimerConc, DefaultTargetConc, 0.5, method)
		a, _ := low.SelfTm(p)
		b, _ := high.SelfTm(p)
		if !(b > a) {
			t.Errorf("method %d: Tm should grow with salt, got %g then %g", method, a, b)
		}
	}
}

func TestMismatchLowersTm(t *testing.T) {
	m := DefaultModel()
	p := []byte("ACACCGCCCGTCACTCTC")
	perfect, _ := m.Tm(p, p)
	mm, _ := m.Tm(p, []byte("ACACCGCCCGTGACTCTC"))
	if !(mm < perfect) {
		t.Errorf("a mismatch must lower Tm: perfect %g, mismatch %g", perfect, mm)
	}
}

func TestInputErrors(t *testing.T) {
	m := DefaultModel()
	if _, err := m.SelfTm([]byte("A")); err == nil {
		t.Error("expected error for a 1-base primer")
	}
	if _, err := m.SelfTm([]byte("ACGNT")); err == nil {
		t.Error("expected error for a non-ACGT base")
	}
	if _, err := m.Tm([]byte("ACGT"), []byte("ACG")); err == nil {
		t.Error("expected error for different lengths")
	}
	if _, err := NewModel(0, 0, 0.05, SantaLucia); err == nil {
		t.Error("expected error for zero strand concentration")
	}
	if _, err := NewModel(1e-6, 0, -1, SantaLucia); err == nil {
		t.Error("expected error for negative salt")
	}
	if _, err := NewModel(1e-6, 0, 0.05, SaltMethod(9)); err == nil {
		t.Error("expected error for unknown salt method")
	}
}

func TestLowerCaseAccepted(t *testing.T) {
	m := DefaultModel()
	a, _ := m.SelfTm([]byte("ACACCGCCCGTCACTCTC"))
	b, _ := m.SelfTm([]byte("acaccgcccgtcactctc"))
	if a != b {
		t.Errorf("case must not matter: %g vs %g", a, b)
	}
}

func TestGCCount(t *testing.T) {
	if n := GCCount([]byte("ACgtCCaT")); n != 4 {
		t.Errorf("GCCount = %d, want 4", n)
	}
}

func BenchmarkTm(b *testing.B) {
	m := DefaultModel()
	p := []byte("ACACCGCCCGTCACTCTC")
	s := []byte("ACACCGCCCGTGACTCTC")
	b.ReportAllocs()
	for b.Loop() {
		if _, err := m.Tm(p, s); err != nil {
			b.Fatal(err)
		}
	}
}
