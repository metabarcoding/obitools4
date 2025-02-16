package obikmer

import (
	"iter"
	"slices"
)

var baseError = map[byte]byte{
	'a': 'b',
	'c': 'd',
	'g': 'h',
	't': 'v',
	'r': 'y',
	'y': 'r',
	's': 'w',
	'w': 's',
	'k': 'm',
	'm': 'k',
	'd': 'c',
	'v': 't',
	'h': 'g',
	'b': 'a',
}

type BytesItem []byte

func IterateOneError(kmer []byte) iter.Seq[BytesItem] {
	lkmer := len(kmer)
	return func(yield func(BytesItem) bool) {
		for p := 0; p < lkmer; p++ {
			for p < lkmer && kmer[p] == 'n' {
				p++
			}

			if p < lkmer {
				nkmer := slices.Clone(kmer)
				nkmer[p] = baseError[kmer[p]]
				if !yield(nkmer) {
					return
				}
			}
		}
	}

}
