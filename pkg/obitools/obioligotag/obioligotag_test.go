package obioligotag

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
)

// defaults puts every option back to its default value.
func defaults() {
	SetOligoSize(0)
	SetOligoList("")
	SetFamilySize(0)
	SetDistance(1)
	SetDistanceType("hamming")
	SetAlphabet("ACGT")
	SetGCMax(-1)
	SetAccepted("")
	SetRejected("")
	SetHomopolymerMax(-1)
	SetHomopolymerMin(0)
	SetTimeout(60)
}

// reset starts a test from the default options and restores them at its end.
func reset(t *testing.T) {
	t.Helper()
	defaults()
	t.Cleanup(defaults)
}

func hamming(a, b string) int {
	d := 0
	for i := range a {
		if a[i] != b[i] {
			d++
		}
	}
	return d
}

func TestOligoFamilyHamming(t *testing.T) {
	reset(t)
	SetOligoSize(5)
	SetFamilySize(24)
	SetDistance(3)
	SetGCMax(3)
	SetHomopolymerMax(2)

	family, err := CLIOligoFamily()
	if err != nil {
		t.Fatal(err)
	}
	if len(family) < 24 {
		t.Fatalf("expected at least 24 oligonucleotides, got %d", len(family))
	}
	for i, a := range family {
		if strings.Count(a, "g")+strings.Count(a, "c") > 3 {
			t.Errorf("%s has too many G/C", a)
		}
		for _, h := range []string{"aaa", "ccc", "ggg", "ttt"} {
			if strings.Contains(a, h) {
				t.Errorf("%s has a homopolymer longer than 2", a)
			}
		}
		for _, b := range family[i+1:] {
			if hamming(a, b) < 3 {
				t.Errorf("%s and %s are too close", a, b)
			}
		}
	}
}

func TestOligoFamilyLCSWithoutG(t *testing.T) {
	reset(t)
	SetOligoSize(6)
	SetFamilySize(20)
	SetDistance(3)
	SetDistanceType("lcs")
	SetAlphabet("ACT")

	family, err := CLIOligoFamily()
	if err != nil {
		t.Fatal(err)
	}
	if len(family) < 20 {
		t.Fatalf("expected at least 20 oligonucleotides, got %d", len(family))
	}
	for i, a := range family {
		if strings.Contains(a, "g") {
			t.Errorf("%s contains a G", a)
		}
		for _, b := range family[i+1:] {
			ka, kb := obikmer.EncodeKmer([]byte(a), 6), obikmer.EncodeKmer([]byte(b), 6)
			if obikmer.KmerLCS(ka, kb, 6).Distance() < 3 {
				t.Errorf("%s and %s are too close", a, b)
			}
		}
	}
}

func TestOligoListPlainTextAndFasta(t *testing.T) {
	dir := t.TempDir()
	plain := filepath.Join(dir, "oligos.txt")
	fasta := filepath.Join(dir, "oligos.fasta")
	os.WriteFile(plain, []byte("# tags\nacgtac\nacgtat\nttttgg\nggcaac\ncaccat\n\nacgtac\nacgNac\n"), 0o644)
	os.WriteFile(fasta, []byte(">a\nACGTAC\n>b\nACGTAT\n>c\nTTTTGG\n>d\nGGCAAC\n>e\nCACCAT\n"), 0o644)

	for _, file := range []string{plain, fasta} {
		reset(t)
		SetOligoList(file)
		SetFamilySize(3)
		SetDistance(3)

		kmers, size, err := candidateOligos()
		if err != nil {
			t.Fatalf("%s: %v", file, err)
		}
		if size != 6 || len(kmers) != 5 { // duplicates removed, 'acgNac' ignored in the text file
			t.Fatalf("%s: got %d oligonucleotides of size %d", file, len(kmers), size)
		}
		family, err := CLIOligoFamily()
		if err != nil || len(family) < 3 {
			t.Fatalf("%s: family %v, error %v", file, family, err)
		}
	}
}

func TestOligoListFilteredByConstraints(t *testing.T) {
	reset(t)
	file := filepath.Join(t.TempDir(), "oligos.txt")
	os.WriteFile(file, []byte("aaaaaa\nacgtac\nacgtat\n"), 0o644)
	SetOligoList(file)
	SetFamilySize(1)
	SetHomopolymerMax(2)

	kmers, _, err := candidateOligos()
	if err != nil || len(kmers) != 2 {
		t.Fatalf("expected the homopolymer to be filtered out, got %d (%v)", len(kmers), err)
	}
}

func TestOligoTagErrors(t *testing.T) {
	cases := map[string]func(){
		"no size nor list":     func() { SetFamilySize(2) },
		"size and list":        func() { SetFamilySize(2); SetOligoSize(5); SetOligoList("x") },
		"no family size":       func() { SetOligoSize(5) },
		"size too large":       func() { SetFamilySize(2); SetOligoSize(13) },
		"bad alphabet":         func() { SetFamilySize(2); SetOligoSize(5); SetAlphabet("ACX") },
		"bad pattern length":   func() { SetFamilySize(2); SetOligoSize(5); SetAccepted("nnn") },
		"bad distance type":    func() { SetFamilySize(2); SetOligoSize(5); SetDistanceType("levenshtein") },
		"beyond Hamming bound": func() { SetFamilySize(100); SetOligoSize(5); SetDistance(3) },
		"empty candidate set":  func() { SetFamilySize(2); SetOligoSize(4); SetAlphabet("A"); SetHomopolymerMax(2) },
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			reset(t)
			setup()
			if family, err := CLIOligoFamily(); err == nil {
				t.Errorf("expected an error, got %v", family)
			}
		})
	}
}

func TestOligoIsPlainLine(t *testing.T) {
	for line, want := range map[string]bool{
		"acgt": true, "ACGTRYN": true, ">seq1": false, "@read": false, "a,b": false, "": false,
	} {
		if got := isPlainOligoLine(line); got != want {
			t.Errorf("isPlainOligoLine(%q) = %v, want %v", line, got, want)
		}
	}
}
