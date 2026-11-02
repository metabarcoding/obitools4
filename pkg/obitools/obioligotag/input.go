package obioligotag

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strings"

	log "github.com/sirupsen/logrus"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obikmer"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obitools/obiconvert"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

// maxOligoSize is the longest oligonucleotide handled.
const maxOligoSize = 12

// isPlainOligoLine tells whether a line looks like an oligonucleotide: only
// nucleotide letters, which a sequence file title line or a header never is.
func isPlainOligoLine(line string) bool {
	if line == "" {
		return false
	}
	for i := 0; i < len(line); i++ {
		if obiseq.IUPACNucMask(line[i]) == 0 {
			return false
		}
	}
	return true
}

// isPlainOligoFile guesses whether filename is a text file with one
// oligonucleotide per line, by looking at its first significant line.
func isPlainOligoFile(filename string) (bool, error) {
	file, err := obiutils.Ropen(filename)
	if err != nil {
		return false, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			return isPlainOligoLine(line), nil
		}
		if err != nil {
			if err == io.EOF {
				return false, fmt.Errorf("file %s is empty", filename)
			}
			return false, err
		}
	}
}

// oligoCollector accumulates encoded oligonucleotides of a single size.
type oligoCollector struct {
	size    int
	kmers   []uint64
	skipped int
}

func (oc *oligoCollector) add(seq []byte) error {
	if len(seq) == 0 {
		return nil
	}
	if len(seq) > maxOligoSize {
		return fmt.Errorf("oligonucleotide %q is longer than the maximum size %d", seq, maxOligoSize)
	}
	if oc.size == 0 {
		oc.size = len(seq)
	}
	if len(seq) != oc.size {
		return fmt.Errorf("all the oligonucleotides must have the same length: found %d and %d", oc.size, len(seq))
	}
	for _, b := range seq {
		switch b | 32 {
		case 'a', 'c', 'g', 't', 'u':
		default:
			oc.skipped++
			return nil
		}
	}
	oc.kmers = append(oc.kmers, obikmer.EncodeKmer(seq, len(seq)))
	return nil
}

// readOligoList reads the oligonucleotides of filename, either as a sequence
// file or as a text file with one oligonucleotide per line. Oligonucleotides
// containing something else than A, C, G, T or U are skipped, and
// duplicates removed. The result is sorted.
func readOligoList(filename string) (kmers []uint64, size int, err error) {
	plain := false
	if obiconvert.CLIInputFormat() == "guessed" {
		plain, err = isPlainOligoFile(filename)
		if err != nil {
			return nil, 0, err
		}
	}

	oc := &oligoCollector{}

	if plain {
		file, err := obiutils.Ropen(filename)
		if err != nil {
			return nil, 0, err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if err := oc.add([]byte(line)); err != nil {
				return nil, 0, err
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, 0, err
		}
	} else {
		sequences, err := obiconvert.CLIReadBioSequences(filename)
		if err != nil {
			return nil, 0, err
		}
		for sequences.Next() {
			for _, s := range sequences.Get().Slice() {
				if err := oc.add(s.Sequence()); err != nil {
					return nil, 0, err
				}
			}
		}
	}

	if oc.skipped > 0 {
		log.Warnf("%d oligonucleotides containing characters other than A, C, G, T were ignored", oc.skipped)
	}

	slices.Sort(oc.kmers)
	oc.kmers = slices.Compact(oc.kmers)
	return oc.kmers, oc.size, nil
}
