// Code generated from ecoPrimers libthermo/nnparams.c (nparam_InitParams); DO NOT EDIT.

package obithermo

// nnEntry is one explicit nearest-neighbor parameter: the 5'->3' dinucleotide of
// the top strand, the 3'->5' dinucleotide facing it, enthalpy (cal/mol) and
// entropy (cal/mol/K).
type nnEntry struct {
	top, bottom string
	dH, dS      float64
}

var nnEntries = [...]nnEntry{
	{"AA", "TT", -7600.0, -21.3},  // AA/TT
	{"AC", "TG", -8400.0, -22.4},  // AC/TG
	{"AG", "TC", -7800.0, -21.0},  // AG/TC
	{"AT", "TA", -7200.0, -20.4},  // AT/TA
	{"CA", "GT", -8500.0, -22.7},  // CA/GT
	{"CC", "GG", -8000.0, -19.9},  // CC/GG
	{"CG", "GC", -10600.0, -27.2}, // CG/GC
	{"CT", "GA", -7800.0, -21.0},  // CT/GA
	{"GA", "CT", -8200.0, -22.2},  // GA/CT
	{"GC", "CG", -9800.0, -24.4},  // GC/CG
	{"GG", "CC", -8000.0, -19.9},  // GG/CC
	{"GT", "CA", -8400.0, -22.4},  // GT/CA
	{"TA", "AT", -7200.0, -21.3},  // TA/AT
	{"TC", "AG", -8200.0, -22.2},  // TC/AG
	{"TG", "AC", -8500.0, -22.7},  // TG/AC
	{"TT", "AA", -7600.0, -21.3},  // TT/AA
	{"AA", "CT", 7600.0, 20.2},    // AA/CT
	{"AA", "TC", 2300.0, 4.6},     // AA/TC
	{"AC", "CG", -700.0, -3.8},    // AC/CG
	{"AC", "TA", 5300.0, 14.6},    // AC/TA
	{"AG", "CC", 600.0, -0.6},     // AG/CC
	{"AT", "CA", 5300.0, 14.6},    // AT/CA
	{"CA", "AT", 3400.0, 8.0},     // CA/AT
	{"CA", "GC", 1900.0, 3.7},     // CA/GC
	{"CC", "AG", 5200.0, 14.2},    // CC/AG
	{"CC", "GA", 600.0, -0.6},     // CC/GA
	{"CG", "AC", 1900.0, 3.7},     // CG/AC
	{"CT", "AA", 2300.0, 4.6},     // CT/AA
	{"GA", "CC", 5200.0, 14.2},    // GA/CC
	{"GC", "CA", -700.0, -3.8},    // GC/CA
	{"TA", "AC", 3400.0, 8.0},     // TA/AC
	{"TC", "AA", 7600.0, 20.2},    // TC/AA
	{"AC", "TT", 700.0, 0.2},      // AC/TT
	{"AT", "TC", -1200.0, -6.2},   // AT/TC
	{"CA", "TT", 1000.0, 0.7},     // CA/TT
	{"CC", "GT", -800.0, -4.5},    // CC/GT
	{"CC", "TG", 5200.0, 13.5},    // CC/TG
	{"CG", "TC", -1500.0, -6.1},   // CG/TC
	{"CT", "GC", -1500.0, -6.1},   // CT/GC
	{"CT", "TA", -1200.0, -6.2},   // CT/TA
	{"GC", "CT", 2300.0, 5.4},     // GC/CT
	{"GT", "CC", 5200.0, 13.5},    // GT/CC
	{"TA", "CT", 1200.0, 0.7},     // TA/CT
	{"TC", "CG", 2300.0, 5.4},     // TC/CG
	{"TC", "AT", 1200.0, 0.7},     // TC/AT
	{"TG", "CC", -800.0, -4.5},    // TG/CC
	{"TT", "CA", 700.0, 0.2},      // TT/CA
	{"TT", "AC", 1000.0, 0.7},     // TT/AC
	{"AA", "GT", 3000.0, 7.4},     // AA/GT
	{"AA", "TG", -600.0, -2.3},    // AA/TG
	{"AC", "GG", 500.0, 3.2},      // AC/GG
	{"AG", "GC", -4000.0, -13.2},  // AG/GC
	{"AG", "TA", -700.0, -2.3},    // AG/TA
	{"AT", "GA", -700.0, -2.3},    // AT/GA
	{"CA", "GG", -700.0, -2.3},    // CA/GG
	{"CG", "GA", -4000.0, -13.2},  // CG/GA
	{"GA", "AT", 700.0, 0.7},      // GA/AT
	{"GA", "CG", -600.0, -1.0},    // GA/CG
	{"GC", "AG", -600.0, -1.0},    // GC/AG
	{"GG", "AC", -700.0, -2.3},    // GG/AC
	{"GG", "CA", 500.0, 3.2},      // GG/CA
	{"GT", "AA", -600.0, -2.3},    // GT/AA
	{"TA", "AG", 700.0, 0.7},      // TA/AG
	{"TG", "AA", 3000.0, 7.4},     // TG/AA
	{"AG", "TT", 1000.0, 0.9},     // AG/TT
	{"AT", "TG", -2500.0, -8.3},   // AT/TG
	{"CG", "GT", -4100.0, -11.7},  // CG/GT
	{"CT", "GG", -2800.0, -8.0},   // CT/GG
	{"GA", "TT", -1300.0, -5.3},   // GA/TT
	{"GC", "TG", -4400.0, -12.3},  // GC/TG
	{"GG", "CT", 3300.0, 10.4},    // GG/CT
	{"GG", "TC", -2800.0, -8.0},   // GG/TC
	{"GT", "CG", -4400.0, -12.3},  // GT/CG
	{"GT", "TA", -2500.0, -8.3},   // GT/TA
	{"TA", "GT", -100.0, -1.7},    // TA/GT
	{"TC", "GG", 3300.0, 10.4},    // TC/GG
	{"TG", "AT", -100.0, -1.7},    // TG/AT
	{"TG", "GC", -4100.0, -11.7},  // TG/GC
	{"TT", "AG", -1300.0, -5.3},   // TT/AG
	{"TT", "GA", 1000.0, 0.9},     // TT/GA
	{"AA", "AT", 4700.0, 12.9},    // AA/AT
	{"AA", "TA", 1200.0, 1.7},     // AA/TA
	{"AC", "AG", -2900.0, -9.8},   // AC/AG
	{"AG", "AC", -900.0, -4.2},    // AG/AC
	{"AT", "AA", 1200.0, 1.7},     // AT/AA
	{"CA", "GA", -900.0, -4.2},    // CA/GA
	{"GA", "CA", -2900.0, -9.8},   // GA/CA
	{"TA", "AA", 4700.0, 12.9},    // TA/AA
	{"AC", "TC", 0.0, -4.4},       // AC/TC
	{"CA", "CT", 6100.0, 16.4},    // CA/CT
	{"CC", "CG", 3600.0, 8.9},     // CC/CG
	{"CC", "GC", -1500.0, -7.2},   // CC/GC
	{"CG", "CC", -1500.0, -7.2},   // CG/CC
	{"CT", "CA", 0.0, -4.4},       // CT/CA
	{"GC", "CC", 3600.0, 8.9},     // GC/CC
	{"TC", "AC", 6100.0, 16.4},    // TC/AC
	{"AG", "TG", -3100.0, -9.5},   // AG/TG
	{"CG", "GG", -4900.0, -15.3},  // CG/GG
	{"GA", "GT", 1600.0, 3.6},     // GA/GT
	{"GC", "GG", -6000.0, -15.8},  // GC/GG
	{"GG", "CG", -6000.0, -15.8},  // GG/CG
	{"GG", "GC", -4900.0, -15.3},  // GG/GC
	{"GT", "GA", -3100.0, -9.5},   // GT/GA
	{"TG", "AG", 1600.0, 3.6},     // TG/AG
	{"AT", "TT", -2700.0, -10.8},  // AT/TT
	{"CT", "GT", -5000.0, -15.8},  // CT/GT
	{"GT", "CT", -2200.0, -8.4},   // GT/CT
	{"TA", "TT", 200.0, -1.5},     // TA/TT
	{"TC", "TG", -2200.0, -8.4},   // TC/TG
	{"TG", "TC", -5000.0, -15.8},  // TG/TC
	{"TT", "AT", 200.0, -1.5},     // TT/AT
	{"TT", "TA", -2700.0, -10.8},  // TT/TA
}
