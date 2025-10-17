package obiclean

import (
	"fmt"
	"sort"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

// Bitmasks over A,C,G,T
// A=1, C=2, G=4, T/U=8
var iupacMask = [256]uint8{
	'A': 1, 'a': 1,
	'C': 2, 'c': 2,
	'G': 4, 'g': 4,
	'T': 8, 't': 8,
	'U': 8, 'u': 8,
	'R': 1 | 4, 'r': 1 | 4, // A or G
	'Y': 2 | 8, 'y': 2 | 8, // C or T
	'S': 2 | 4, 's': 2 | 4, // G or C
	'W': 1 | 8, 'w': 1 | 8, // A or T
	'K': 4 | 8, 'k': 4 | 8, // G or T
	'M': 1 | 2, 'm': 1 | 2, // A or C
	'B': 2 | 4 | 8, 'b': 2 | 4 | 8, // C or G or T
	'D': 1 | 4 | 8, 'd': 1 | 4 | 8, // A or G or T
	'H': 1 | 2 | 8, 'h': 1 | 2 | 8, // A or C or T
	'V': 1 | 2 | 4, 'v': 1 | 2 | 4, // A or C or G
	'N': 1 | 2 | 4 | 8, 'n': 1 | 2 | 4 | 8, // any
	// Optional: treat '.', '-', '?' as unknowns
	'.': 1 | 2 | 4 | 8,
	'-': 1 | 2 | 4 | 8,
	'?': 1 | 2 | 4 | 8,
}

func iupacCompatible(a, b byte) bool {
	return (iupacMask[a] & iupacMask[b]) != 0
}

func commonPrefix(a, b *obiseq.BioSequence) int {
	i := 0
	l := min(a.Len(), b.Len())

	if l == 0 {
		return 0
	}
	as := a.Sequence()
	bs := b.Sequence()

	for i < l && iupacCompatible(as[i], bs[i]) {
		i++
	}

	// if obiutils.UnsafeString(as[:i]) != obiutils.UnsafeString(bs[:i]) {
	// 	log.Fatalf("i: %d, j: %d (%s/%s)", i, i, as[:i], bs[:i])
	// }

	return i
}

func commonSuffix(a, b *obiseq.BioSequence) int {
	i := a.Len() - 1
	j := b.Len() - 1

	if i < 0 || j < 0 {
		return 0
	}

	as := a.Sequence()
	bs := b.Sequence()

	l := 0
	for i >= 0 && j >= 0 && iupacCompatible(as[i], bs[j]) {
		i--
		j--
		l++
	}

	// if obiutils.UnsafeString(as[i+1:]) != obiutils.UnsafeString(bs[j+1:]) {
	// 	log.Fatalf("i: %d, j: %d (%s/%s)", i, j, as[i+1:], bs[j+1:])
	// }
	// log.Warnf("i: %d, j: %d (%s)", i, j, as[i+1:])

	return l
}

// oneDifference return true if s1 and s2 differ by exactly 1 operation
// (substitution, insertion or deletion)
func oneDifference(s1, s2 []byte) bool {
	l1 := len(s1)
	l2 := len(s2)

	// Case 1: same lengths → test substitution
	if l1 == l2 {
		diff := 0
		for i := 0; i < l1; i++ {
			if s1[i] != s2[i] {
				diff++
				if diff > 1 {
					return false
				}
			}
		}
		return diff == 1
	}

	// Case 2: difference of 1 character → insertion/deletion
	if l1 == l2+1 {
		for i := 0; i < l1; i++ {
			if string(s1[:i])+string(s1[i+1:]) == string(s2) {
				return true
			}
		}
		return false
	}
	if l2 == l1+1 {
		for i := 0; i < l2; i++ {
			if string(s2[:i])+string(s2[i+1:]) == string(s1) {
				return true
			}
		}
		return false
	}

	// Case 3: difference > 1 character
	return false
}

func GetChimera(sequence *obiseq.BioSequence) map[string]string {
	annotation := sequence.Annotations()
	iobistatus, ok := annotation["chimera"]
	var chimera map[string]string
	var err error

	if ok {
		switch iobistatus := iobistatus.(type) {
		case map[string]string:
			chimera = iobistatus
		case map[string]interface{}:
			chimera = make(map[string]string)
			for k, v := range iobistatus {
				chimera[k], err = obiutils.InterfaceToString(v)
				if err != nil {
					log.Panicf("chimera attribute of sequence %s must be castable to a map[string]string", sequence.Id())
				}
			}
		}
	} else {
		chimera = make(map[string]string)
		annotation["chimera"] = chimera
	}

	return chimera
}

// add the suffix/prefix comparisons to detect chimeras (handle IUPAC codes)
func AnnotateChimera(samples map[string]*[]*seqPCR) {

	w := func(sample string, seqs *[]*seqPCR) {
		ls := len(*seqs)
		pcrs := make([]*seqPCR, 0, ls)

		// select only sequences without edges (head sequences)
		for _, s := range *seqs {
			if len(s.Edges) == 0 {
				pcrs = append(pcrs, s)
			}
		}

		lp := len(pcrs)

		// sort by increasing weight (like increasing abundance)
		sort.Slice(pcrs, func(i, j int) bool {
			return pcrs[i].Weight < pcrs[j].Weight
		})

		for i, s := range pcrs {
			seqRef := s.Sequence
			L := seqRef.Len()

			maxLeft, maxRight := 0, 0
			var nameLeft, nameRight string

			// looking for potential parents
			for j := 0; j < lp; j++ {
				if j == i {
					continue
				}
				seqParent := pcrs[j].Sequence

				// Check abundance: parent must be more abundant
				if pcrs[j].Weight <= s.Weight {
					continue
				}

				// Check edit distance (skip if only one diff, supposed to never happen)
				if oneDifference(seqRef.Sequence(), seqParent.Sequence()) {
					continue
				}

				// Common prefix
				left := commonPrefix(seqRef, seqParent)
				if left > maxLeft {
					maxLeft = left
					nameLeft = seqParent.Id()
				}

				// Common suffix
				right := commonSuffix(seqRef, seqParent)
				if right > maxRight {
					maxRight = right
					nameRight = seqParent.Id()
				}
			}

			// Select parents with longuest prefix/suffix
			// Condition prefix+suffix covers the sequence and sequence not include into parent
			if maxLeft >= L-maxRight && maxLeft > 0 && maxRight < L {

				chimeraMap := GetChimera(s.Sequence)
				// overlap sequence
				overlap := seqRef.Sequence()[L-maxRight : maxLeft]
				chimeraMap[sample] = fmt.Sprintf("{%s}/{%s}@(%s)(%d)(%d)(%d)", nameLeft, nameRight, overlap, L-maxRight, maxLeft, len(overlap))
			}
		}
	}

	for sn, sqs := range samples {
		w(sn, sqs)
	}
}

// func AnnotateChimera(samples map[string]*[]*seqPCR) {

// 	w := func(sample string, seqs *[]*seqPCR) {
// 		ls := len(*seqs)
// 		cp := make([]int, ls)
// 		cs := make([]int, ls)

// 		pcrs := make([]*seqPCR, 0, ls)

// 		for _, s := range *seqs {
// 			if len(s.Edges) == 0 {
// 				pcrs = append(pcrs, s)
// 			}
// 		}

// 		lp := len(pcrs)

// 		sort.Slice(pcrs, func(i, j int) bool {
// 			return pcrs[i].Weight < pcrs[j].Weight
// 		})

// 		for i, s := range pcrs {
// 			for j := i + 1; j < lp; j++ {
// 				s2 := pcrs[j]
// 				cp[j] = commonPrefix(s.Sequence, s2.Sequence)
// 				cs[j] = commonSuffix(s.Sequence, s2.Sequence)
// 			}

// 			var cm map[string]string
// 			var err error

// 			chimera, ok := s.Sequence.GetAttribute("chimera")

// 			if !ok {
// 				cm = map[string]string{}
// 			} else {
// 				cm, err = obiutils.InterfaceToStringMap(chimera)
// 				if err != nil {
// 					log.Fatalf("type of chimera not map[string]string: %T (%v)",
// 						chimera, err)
// 				}
// 			}

// 			ls := s.Sequence.Len()

// 			for k := i + 1; k < lp; k++ {
// 				for l := i + 1; l < lp; l++ {
// 					if k != l && cp[k]+cs[l] == ls {
// 						cm[sample] = fmt.Sprintf("{%s}/{%s}@(%d)",
// 							pcrs[k].Sequence.Id(),
// 							pcrs[l].Sequence.Id(),
// 							cp[k])
// 					}
// 				}
// 			}

// 			if len(cm) > 0 {
// 				s.Sequence.SetAttribute("chimera", cm)
// 			}
// 		}

// 	}

// 	for sn, sqs := range samples {
// 		w(sn, sqs)
// 	}

// }
