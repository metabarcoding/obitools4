package obiprimers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiprimer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitax"
)

// Three sequences below two families (5 and 6) of the same root, annotated like
// obiannotate --taxonomic-path does.
const embeddedFasta = `>s1 {"taxid":"taxon:10 [Alpha beta]@species","taxonomic_path":["taxon:1 [root]@no rank","taxon:5 [Famone]@family","taxon:10 [Alpha beta]@species"]}
acgtacgtacgtacgtacgtacgtacgtacgtacgtacgt
>s2 {"taxid":"taxon:11 [Alpha gamma]@species","taxonomic_path":["taxon:1 [root]@no rank","taxon:5 [Famone]@family","taxon:11 [Alpha gamma]@species"]}
acgtacgtacgtacgtacgtacgtacgtacgtacgtacgt
>s3 {"taxid":"taxon:20 [Delta eps]@species","taxonomic_path":["taxon:1 [root]@no rank","taxon:6 [Famtwo]@family","taxon:20 [Delta eps]@species"]}
acgtacgtacgtacgtacgtacgtacgtacgtacgtacgt
`

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "seqs.fasta")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func drain(it obiiter.IBioSequence) (n int) {
	for it.Next() {
		batch := it.Get()
		n += len(batch.Slice())
	}
	return
}

func TestEmbeddedTaxonomyFromFiles(t *testing.T) {
	path := writeTemp(t, embeddedFasta)

	sequences, taxonomy, err := readSequencesAndTaxonomy([]string{path}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if taxonomy == nil || taxonomy.Len() != 6 { // root, 2 families, 3 species
		t.Fatalf("expected an embedded taxonomy of 6 taxa, got %v", taxonomy)
	}

	// the sequences are read in full on the second pass
	restrict, err := obiprimer.BelongingTo(taxonomy, []string{"5"})
	if err != nil {
		t.Fatalf("a taxid of the embedded taxonomy must be usable with -r: %v", err)
	}
	corpus, err := obiprimer.LoadCorpus(sequences, obiprimer.Selection{Restrict: restrict})
	if err != nil {
		t.Fatal(err)
	}
	if c := corpus.Counts(); c.Examples != 2 || c.Ignored != 1 {
		t.Errorf("restricting to family 5: %+v", c)
	}

	taxa, err := corpus.AssignTaxa(taxonomy, "species")
	if err != nil {
		t.Fatal(err)
	}
	if taxa.ExampleTaxa != 2 {
		t.Errorf("%d example species, want 2", taxa.ExampleTaxa)
	}
}

func TestEmbeddedTaxonomyFromMemory(t *testing.T) { // the standard input case
	var slice obiseq.BioSequenceSlice
	for _, s := range []struct{ id, taxid string }{{"a", "10"}, {"b", "20"}} {
		seq := obiseq.NewBioSequence(s.id, []byte("acgtacgtacgt"), "")
		seq.SetAttribute("taxonomic_path", []string{
			"taxon:1 [root]@no rank", "taxon:" + s.taxid + " [Sp" + s.taxid + "]@species"})
		slice = append(slice, seq)
	}

	sequences, taxonomy, err := embeddedTaxonomyInMemory(obiiter.IBatchOver("stdin", slice, 10))
	if err != nil {
		t.Fatal(err)
	}
	if taxonomy.Len() != 3 {
		t.Errorf("%d taxa, want 3", taxonomy.Len())
	}
	if n := drain(sequences); n != 2 {
		t.Errorf("the sequences must be given back intact: %d of 2", n)
	}
}

func TestMissingEmbeddedTaxonomyGivesAdvice(t *testing.T) {
	path := writeTemp(t, ">s1 {\"taxid\":\"taxon:10 [Alpha beta]@species\"}\nacgtacgtacgtacgt\n")
	_, _, err := readSequencesAndTaxonomy([]string{path}, nil)
	if err == nil {
		t.Fatal("sequences without a taxonomic path and no -t must be refused")
	}
	for _, want := range []string{"-t", "obiannotate --taxonomic-path"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error should mention %q: %v", want, err)
		}
	}

	var slice obiseq.BioSequenceSlice
	slice = append(slice, obiseq.NewBioSequence("a", []byte("acgt"), ""))
	if _, _, err := embeddedTaxonomyInMemory(obiiter.IBatchOver("stdin", slice, 10)); err == nil ||
		!strings.Contains(err.Error(), "obiannotate --taxonomic-path") {
		t.Errorf("the in-memory case must give the same advice: %v", err)
	}
}

func TestGivenTaxonomyHasPrecedence(t *testing.T) {
	path := writeTemp(t, embeddedFasta)
	_, embedded, err := readSequencesAndTaxonomy([]string{path}, nil)
	if err != nil {
		t.Fatal(err)
	}

	// With a taxonomy given (-t), the embedded paths are ignored, even if present.
	other := obitax.NewTaxonomy("other", "other", obitax.DefaultTaxidAlphabet)
	sequences, used, err := readSequencesAndTaxonomy([]string{path}, other)
	if err != nil {
		t.Fatal(err)
	}
	if used != other || used == embedded {
		t.Error("the given taxonomy must be returned as is")
	}
	if n := drain(sequences); n != 3 {
		t.Errorf("%d sequences read, want 3", n)
	}
}

func TestPrimerPairsNeedsATaxonomy(t *testing.T) {
	if _, err := CLIPrimerPairs(obiiter.NilIBioSequence, nil); err == nil {
		t.Error("a nil taxonomy must be refused")
	}
}

// hasPath reports whether any sequence still carries its taxonomic path, and
// how many still have their taxid (which the search needs).
func scanSequences(it obiiter.IBioSequence) (withPath, withTaxid, total int) {
	for it.Next() {
		batch := it.Get()
		for _, s := range batch.Slice() {
			total++
			if s.HasAttribute(obiseq.TaxonomicPathAttribute) {
				withPath++
			}
			if s.HasAttribute("taxid") {
				withTaxid++
			}
		}
	}
	return
}

// Once the taxonomy is built the paths only take memory: they are dropped,
// whatever the way the taxonomy was obtained, and nothing else is.
func TestTaxonomicPathIsDroppedAfterExtraction(t *testing.T) {
	path := writeTemp(t, embeddedFasta)

	sequences, _, err := readSequencesAndTaxonomy([]string{path}, nil) // two passes over a file
	if err != nil {
		t.Fatal(err)
	}
	if withPath, withTaxid, total := scanSequences(sequences); total != 3 || withPath != 0 || withTaxid != 3 {
		t.Errorf("file: %d sequences, %d with a path left, %d with a taxid", total, withPath, withTaxid)
	}

	other := obitax.NewTaxonomy("other", "other", obitax.DefaultTaxidAlphabet) // -t given
	sequences, _, err = readSequencesAndTaxonomy([]string{path}, other)
	if err != nil {
		t.Fatal(err)
	}
	if withPath, _, total := scanSequences(sequences); total != 3 || withPath != 0 {
		t.Errorf("with -t: %d sequences, %d with a path left", total, withPath)
	}

	var slice obiseq.BioSequenceSlice // standard input
	for _, id := range []string{"a", "b"} {
		seq := obiseq.NewBioSequence(id, []byte("acgtacgt"), "")
		seq.SetAttribute("taxid", "taxon:10 [X]@species")
		seq.SetAttribute(obiseq.TaxonomicPathAttribute, []string{"taxon:1 [root]@no rank", "taxon:10 [X]@species"})
		slice = append(slice, seq)
	}
	sequences, _, err = embeddedTaxonomyInMemory(obiiter.IBatchOver("stdin", slice, 10))
	if err != nil {
		t.Fatal(err)
	}
	if withPath, withTaxid, total := scanSequences(sequences); total != 2 || withPath != 0 || withTaxid != 2 {
		t.Errorf("stdin: %d sequences, %d with a path left, %d with a taxid", total, withPath, withTaxid)
	}
}
