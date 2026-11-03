// Package obithermo computes DNA duplex melting temperatures with the
// nearest-neighbor model used by ecoPrimers (SantaLucia 1998 parameters, with
// the SantaLucia or Owczarzy 2008 salt corrections).
//
// Only the situations ecoPrimers actually evaluates are supported: a primer
// against its perfect complement (SelfTm), and a primer against a genomic site
// of the same length that may carry mismatches (Tm). Gaps, bulges and dangling
// ends are not modeled.
package obithermo

import (
	"fmt"
	"math"
)

// SaltMethod selects the monovalent salt correction.
type SaltMethod int

const (
	// SantaLucia applies a per-step entropy correction 0.368*ln[Na+].
	SantaLucia SaltMethod = 1
	// Owczarzy applies the GC-dependent correction of Owczarzy et al. 2008.
	Owczarzy SaltMethod = 2
)

// Default experimental conditions, those of ecoPrimers.
const (
	DefaultPrimerConc = 8e-7 // total primer strand concentration (M)
	DefaultTargetConc = 0.0  // target strand concentration (M)
	DefaultSalt       = 0.05 // monovalent cation concentration (M)
)

const (
	gasConstant       = 1.987   // cal/mol/K
	initEntropy       = -5.9    // duplex initiation, cal/mol/K
	kelvinOffset      = 273.15  // ecoPrimers subtracts 273.0 (see package notes)
	owczarzyGCSlope   = 4.29e-5 // (4.29*fGC - 3.95)e-5 * ln[Na]
	owczarzyOffset    = 3.95e-5
	owczarzyQuadratic = 9.40e-6 // 9.4e-6 * ln[Na]^2
)

// Model holds the thermodynamic tables and the experimental conditions.
// A Model is immutable after NewModel and safe for concurrent use.
type Model struct {
	method SaltMethod
	rlogc  float64 // R*ln(Ct factor), included in the initial entropy
	logNa  float64 // ln[Na+]

	// Indexed by (top0<<6 | top1<<4 | bot0<<2 | bot1) with A,C,G,T = 0..3.
	// dS already includes the SantaLucia salt term when that method is selected.
	dH [256]float64
	dS [256]float64
}

// NewModel builds a Model for the given strand concentrations (M), monovalent
// salt concentration (M) and salt correction method.
func NewModel(primerConc, targetConc, salt float64, method SaltMethod) (*Model, error) {
	if method != SantaLucia && method != Owczarzy {
		return nil, fmt.Errorf("obithermo: unknown salt method %d", method)
	}
	if salt <= 0 {
		return nil, fmt.Errorf("obithermo: salt concentration must be positive, got %g", salt)
	}

	// Effective strand concentration, as in nparam_InitParams.
	var ctFactor float64
	switch {
	case primerConc == targetConc:
		ctFactor = primerConc / 2
	case primerConc > targetConc:
		ctFactor = primerConc - targetConc/2
	default:
		ctFactor = targetConc - primerConc/2
	}
	if ctFactor <= 0 {
		return nil, fmt.Errorf("obithermo: strand concentrations must be positive (primer %g, target %g)",
			primerConc, targetConc)
	}

	m := &Model{
		method: method,
		rlogc:  gasConstant * math.Log(ctFactor),
		logNa:  math.Log(salt),
	}
	m.fillTables()
	return m, nil
}

// DefaultModel returns a Model with the ecoPrimers default conditions and the
// SantaLucia salt correction.
func DefaultModel() *Model {
	m, err := NewModel(DefaultPrimerConc, DefaultTargetConc, DefaultSalt, SantaLucia)
	if err != nil {
		panic(err) // the defaults are valid
	}
	return m
}

var baseCode = func() (t [256]int8) {
	for i := range t {
		t[i] = -1
	}
	for i, c := range "ACGT" {
		t[c] = int8(i)
		t[c+'a'-'A'] = int8(i)
	}
	return
}()

func stepIndex(t0, t1, b0, b1 int) int { return t0<<6 | t1<<4 | b0<<2 | b1 }

func (m *Model) fillTables() {
	// Double mismatches (no Watson-Crick pair at either end of the step) are
	// treated as interior-loop expansions. Complement of code c is 3-c.
	const loopH, loopS = 0.0, -0.97
	for x := range 4 {
		for y := range 4 {
			for a := range 4 {
				for b := range 4 {
					if x+a != 3 && y+b != 3 {
						i := stepIndex(x, y, a, b)
						m.dH[i], m.dS[i] = loopH, loopS
					}
				}
			}
		}
	}

	// Explicit parameters override the defaults (Watson-Crick and single mismatches).
	for _, e := range nnEntries {
		i := stepIndex(int(baseCode[e.top[0]]), int(baseCode[e.top[1]]),
			int(baseCode[e.bottom[0]]), int(baseCode[e.bottom[1]]))
		m.dH[i], m.dS[i] = e.dH, e.dS
	}

	// SantaLucia salt correction: 0.5*kfac per side of each step, i.e. kfac.
	if m.method == SantaLucia {
		kfac := 0.368 * m.logNa
		for i := range m.dS {
			m.dS[i] += kfac
		}
	}
}

// GCCount returns the number of G and C (upper or lower case) in seq.
func GCCount(seq []byte) int {
	n := 0
	for _, c := range seq {
		switch c {
		case 'G', 'C', 'g', 'c':
			n++
		}
	}
	return n
}

// SelfTm returns the melting temperature (degrees Celsius) of primer against
// its perfect complement. It returns NaN when the model predicts no stable duplex.
func (m *Model) SelfTm(primer []byte) (float64, error) {
	return m.duplex(primer, primer, true)
}

// Tm returns the melting temperature (degrees Celsius) of primer against the
// site, a stretch of the target read in the same orientation as the primer
// (primer and site have the same length; mismatches are primer[i] != site[i]).
// It returns NaN when the model predicts no stable duplex.
func (m *Model) Tm(primer, site []byte) (float64, error) {
	if len(site) != len(primer) {
		return 0, fmt.Errorf("obithermo: primer (%d) and site (%d) lengths differ", len(primer), len(site))
	}
	return m.duplex(primer, site, false)
}

func (m *Model) duplex(primer, site []byte, perfect bool) (float64, error) {
	n := len(primer)
	if n < 2 {
		return 0, fmt.Errorf("obithermo: primer too short (%d), need at least 2 bases", n)
	}

	dH := 0.0
	dS := initEntropy + m.rlogc

	p0 := int(baseCode[primer[0]])
	s0 := p0
	if !perfect {
		s0 = int(baseCode[site[0]])
	}
	if p0 < 0 || s0 < 0 {
		return 0, fmt.Errorf("obithermo: non-ACGT base at position 0")
	}

	for i := 1; i < n; i++ {
		p1 := int(baseCode[primer[i]])
		s1 := p1
		if !perfect {
			s1 = int(baseCode[site[i]])
		}
		if p1 < 0 || s1 < 0 {
			return 0, fmt.Errorf("obithermo: non-ACGT base at position %d", i)
		}
		// The partner strand carries the complement of the site.
		idx := stepIndex(p0, p1, 3-s0, 3-s1)
		dH += m.dH[idx]
		dS += m.dS[idx]
		p0, s0 = p1, s1
	}

	if m.method == Owczarzy {
		fgc := float64(GCCount(primer)) / float64(n)
		dS += dH * ((owczarzyGCSlope*fgc-owczarzyOffset)*m.logNa + owczarzyQuadratic*m.logNa*m.logNa)
	}

	// One-state transition, as nparam_CalcTM: valid only for dS < 0 and Tm >= 0.
	if dS >= 0 {
		return math.NaN(), nil
	}
	tm := dH / dS
	if tm < 0 {
		return math.NaN(), nil
	}
	return tm - kelvinOffset, nil
}
