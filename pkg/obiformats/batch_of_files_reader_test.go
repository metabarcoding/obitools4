package obiformats

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// With a single reader, the sequences of several files must come out in the
// order of the files, whatever the order in which the batches are parsed.
func TestReadSequencesBatchFromFilesKeepsOrder(t *testing.T) {
	dir := t.TempDir()
	var names []string
	for _, prefix := range []string{"a", "b"} {
		var sb strings.Builder
		for i := 0; i < 25000; i++ {
			fmt.Fprintf(&sb, ">%s%07d {\"k\":%d}\n%s\n", prefix, i, i, strings.Repeat("acgt", 40))
		}
		name := filepath.Join(dir, prefix+".fasta")
		if err := os.WriteFile(name, []byte(sb.String()), 0o600); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}

	it := ReadSequencesBatchFromFiles(names, nil, 1)
	_, db := it.Load()
	if len(db) != 50000 {
		t.Fatalf("expected 50000 sequences, got %d", len(db))
	}
	for i := 1; i < len(db); i++ {
		if db[i-1].Id() >= db[i].Id() {
			t.Fatalf("out of order at %d: %s then %s", i, db[i-1].Id(), db[i].Id())
		}
	}
}
