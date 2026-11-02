package obiclust

import (
	"bytes"
	"sort"
	"strconv"
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

// record is a sequence of the dataset.
type record struct {
	seq       *obiseq.BioSequence
	pos       int    // position in the dataset
	clean     []byte // the nucleotides of the sequence, in lower case
	count     int
	key       float64 // value of the sorting key
	ambiguous bool    // contains IUPAC letters other than a, c, g and t
}

// group is a run of identical sequences: recs[start] is its head, the
// following ones are the duplicates. count is the cumulated count, key the
// value of the sorting key of the head.
type group struct {
	start, end int
	count      int
	key        float64
}

// nucleotide classes of a byte
const (
	notNucleotide = iota
	plainNucleotide
	ambiguousNucleotide
)

var nucleotideClass = func() (t [256]uint8) {
	for _, c := range "acgt" {
		t[c] = plainNucleotide
	}
	for _, c := range "umrwsykvhdbn" {
		t[c] = ambiguousNucleotide
	}
	return t
}()

// classify returns the nucleotides of a sequence and tells if it contains
// IUPAC letters other than a, c, g and t. nil is returned if the sequence is
// empty, or contains anything else than IUPAC letters (and also IUPAC letters
// other than a, c, g and t unless ambiguous ones are accepted).
func classify(s []byte, acceptAmbiguous bool) (clean []byte, ambiguous bool) {
	if len(s) == 0 {
		return nil, false
	}
	for _, c := range s {
		switch nucleotideClass[c] {
		case notNucleotide:
			return nil, false
		case ambiguousNucleotide:
			if !acceptAmbiguous {
				return nil, false
			}
			ambiguous = true
		}
	}
	return s, ambiguous
}

// atof reads the leading number of a string as C atof does.
func atof(v interface{}) float64 {
	if s, ok := v.(string); ok {
		s = strings.TrimSpace(s)
		for n := len(s); n > 0; n-- {
			if f, err := strconv.ParseFloat(s[:n], 64); err == nil {
				return f
			}
		}
		return 0
	}
	f, err := obiutils.InterfaceToFloat64(v)
	if err != nil {
		return 0
	}
	return f
}

// sortKeyOf returns the value of the sorting key of a sequence.
func sortKeyOf(s *obiseq.BioSequence, key string) float64 {
	if key == "count" {
		return float64(s.Count())
	}
	v, ok := s.GetAttribute(key)
	if !ok {
		log.Fatalf("'%s' is not an attribute of the sequence %s", key, s.Id())
	}
	return atof(v)
}

// prepare reads the sequences, groups the identical ones and orders them.
//
// The order is fully determined, whatever the sorting algorithm:
//
//  1. the records of a group of identical sequences are ordered on the
//     sorting key (decreasing, increasing if ascending), then on their
//     position in the dataset; the first one is the head of the group, whose
//     identifier names the cluster if it is a center;
//  2. the groups are ordered on the sorting key of their head (the count of
//     the whole group if the key is "count"), then on the position of the
//     head in the dataset. With the key "None" they are in the order of the
//     first occurrence of their sequence.
//
// The sequences containing anything else than a, c, g and t are discarded,
// except if acceptAmbiguous is set and they only contain IUPAC letters: they
// are then kept, but cannot be the center of a cluster.
func prepare(db obiseq.BioSequenceSlice, key string, ascending, acceptAmbiguous bool) ([]record, []group, int) {
	recs := make([]record, 0, len(db))
	discarded := 0

	for pos, s := range db {
		clean, ambiguous := classify(s.Sequence(), acceptAmbiguous)
		if clean == nil {
			discarded++
			continue
		}
		r := record{seq: s, pos: pos, clean: clean, count: s.Count(), ambiguous: ambiguous}
		if key != "None" {
			r.key = sortKeyOf(s, key)
		}
		recs = append(recs, r)
	}

	// before reports whether a is ranked before b by the sorting key
	before := func(a, b float64) bool {
		if ascending {
			return a < b
		}
		return a > b
	}

	sort.Slice(recs, func(i, j int) bool {
		a, b := &recs[i], &recs[j]
		if c := bytes.Compare(a.clean, b.clean); c != 0 {
			return c < 0
		}
		if a.key != b.key && key != "None" {
			return before(a.key, b.key)
		}
		return a.pos < b.pos
	})

	groups := make([]group, 0, len(recs))
	for i := 0; i < len(recs); {
		j, count := i+1, recs[i].count
		for j < len(recs) && bytes.Equal(recs[j].clean, recs[i].clean) {
			count += recs[j].count
			j++
		}
		g := group{start: i, end: j, count: count, key: recs[i].key}
		if key == "count" {
			g.key = float64(count)
		}
		groups = append(groups, g)
		i = j
	}

	sort.Slice(groups, func(i, j int) bool {
		a, b := &groups[i], &groups[j]
		if key != "None" && a.key != b.key {
			return before(a.key, b.key)
		}
		return recs[a.start].pos < recs[b.start].pos
	})

	return recs, groups, discarded
}
