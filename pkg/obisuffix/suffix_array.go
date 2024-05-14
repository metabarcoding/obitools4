package obisuffix

import (
	"bytes"
	"fmt"
	"sort"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

type Suffix struct {
	Idx uint32
	Pos uint32
}

type SuffixArray struct {
	Sequences *obiseq.BioSequenceSlice
	Suffixes  []Suffix
	Common    []int
}

func SuffixLess(suffixarray SuffixArray) func(i, j int) bool {
	less := func(i, j int) bool {
		si := suffixarray.Suffixes[i]
		bi := (*suffixarray.Sequences)[int(si.Idx)].Sequence()[si.Pos:]
		sj := suffixarray.Suffixes[j]
		bj := (*suffixarray.Sequences)[int(sj.Idx)].Sequence()[sj.Pos:]

		l := obiutils.Min(len(bi), len(bj))
		p := 0
		for p < l && bi[p] == bj[p] {
			p++
		}

		// if p < l {
		// 	log.Debugln(si, sj, p, l, rune(bi[p]), rune(bj[p]))
		// } else {
		// 	log.Debugln(si, sj, p, l, rune('-'), rune('-'))
		// }

		if p == l {
			switch {
			case len(bi) != len(bj):
				return len(bi) < len(bj)
			case si.Idx != sj.Idx:
				return si.Idx < sj.Idx
			default:
				return si.Pos < sj.Pos
			}
		}

		return bi[p] < bj[p]
	}

	return less
}

func BuildSuffixArray(data *obiseq.BioSequenceSlice) SuffixArray {
	totalLength := 0
	for _, s := range *data {
		totalLength += s.Len()
	}

	sa := SuffixArray{
		Sequences: data,
		Suffixes:  make([]Suffix, 0, totalLength),
	}

	for i, s := range *data {
		sl := uint32(s.Len())
		for p := uint32(0); p < sl; p++ {
			sa.Suffixes = append(sa.Suffixes, Suffix{uint32(i), p})
		}
	}

	sort.SliceStable(sa.Suffixes, SuffixLess(sa))
	return sa
}

func (suffixarray *SuffixArray) CommonSuffix() []int {
	if len(suffixarray.Common) == len(suffixarray.Suffixes) {
		return suffixarray.Common
	}

	lrep := len(suffixarray.Suffixes)
	rep := make([]int, lrep)

	sp := suffixarray.Suffixes[0]
	bp := (*suffixarray.Sequences)[int(sp.Idx)].Sequence()[sp.Pos:]
	for i := 1; i < lrep; i++ {
		si := suffixarray.Suffixes[i]
		bi := (*suffixarray.Sequences)[int(si.Idx)].Sequence()[si.Pos:]

		l := obiutils.Min(len(bi), len(bp))
		p := 0
		for p < l && bi[p] == bp[p] {
			p++
		}
		rep[i-1] = p
		sp = si
		bp = bi
	}

	rep[lrep-1] = 0

	suffixarray.Common = rep

	return suffixarray.Common
}

func (suffixarray *SuffixArray) String() string {
	sb := bytes.Buffer{}

	common := suffixarray.CommonSuffix()

	sb.WriteString("Common\tSeqIdx\tPosition\tSuffix\n")

	for i := range suffixarray.Suffixes {
		idx := suffixarray.Suffixes[i].Idx
		pos := suffixarray.Suffixes[i].Pos
		sb.WriteString(fmt.Sprintf("%6d\t%6d\t%8d\t%s\n",
			common[i],
			idx,
			pos,
			string((*suffixarray.Sequences)[idx].Sequence()[pos:]),
		))
	}

	return sb.String()
}

// func LongestInternalRepeat(suffixarray SuffixArray) []int {
// 	common := suffixarray.CommonSuffix()
// 	rep := make([]int, len(*suffixarray.Sequences))

// }
