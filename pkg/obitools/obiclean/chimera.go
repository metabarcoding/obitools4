package obiclean

import (
	"fmt"
	"sort"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

func commonPrefix(a, b *obiseq.BioSequence) int {
	i := 0
	l := min(a.Len(), b.Len())

	if l == 0 {
		return 0
	}
	as := a.Sequence()
	bs := b.Sequence()

	for i < l && as[i] == bs[i] {
		i++
	}

	if obiutils.UnsafeString(as[:i]) != obiutils.UnsafeString(bs[:i]) {
		log.Fatalf("i: %d, j: %d (%s/%s)", i, i, as[:i], bs[:i])
	}

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
	for i >= 0 && j >= 0 && as[i] == bs[j] {
		i--
		j--
		l++
	}

	if obiutils.UnsafeString(as[i+1:]) != obiutils.UnsafeString(bs[j+1:]) {
		log.Fatalf("i: %d, j: %d (%s/%s)", i, j, as[i+1:], bs[j+1:])
	}
	// log.Warnf("i: %d, j: %d (%s)", i, j, as[i+1:])

	return l
}

func AnnotateChimera(samples map[string]*[]*seqPCR) {

	w := func(sample string, seqs *[]*seqPCR) {
		ls := len(*seqs)
		cp := make([]int, ls)
		cs := make([]int, ls)

		pcrs := make([]*seqPCR, 0, ls)

		for _, s := range *seqs {
			if len(s.Edges) == 0 {
				pcrs = append(pcrs, s)
			}
		}

		lp := len(pcrs)

		sort.Slice(pcrs, func(i, j int) bool {
			return pcrs[i].Weight < pcrs[j].Weight
		})

		for i, s := range pcrs {
			for j := i + 1; j < lp; j++ {
				s2 := pcrs[j]
				cp[j] = commonPrefix(s.Sequence, s2.Sequence)
				cs[j] = commonSuffix(s.Sequence, s2.Sequence)
			}

			var cm map[string]string
			var err error

			chimera, ok := s.Sequence.GetAttribute("chimera")

			if !ok {
				cm = map[string]string{}
			} else {
				cm, err = obiutils.InterfaceToStringMap(chimera)
				if err != nil {
					log.Fatalf("type of chimera not map[string]string: %T (%v)",
						chimera, err)
				}
			}

			ls := s.Sequence.Len()

			for k := i + 1; k < lp; k++ {
				for l := i + 1; l < lp; l++ {
					if k != l && cp[k]+cs[l] == ls {
						cm[sample] = fmt.Sprintf("{%s}/{%s}@(%d)",
							pcrs[k].Sequence.Id(),
							pcrs[l].Sequence.Id(),
							cp[k])
					}
				}
			}

			if len(cm) > 0 {
				s.Sequence.SetAttribute("chimera", cm)
			}
		}

	}

	for sn, sqs := range samples {
		w(sn, sqs)
	}

}
