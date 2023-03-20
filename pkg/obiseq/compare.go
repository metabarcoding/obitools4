package obiseq

import (
	"bytes"
)

type Compare func(a, b *BioSequence) int

func CompareSequence(a, b *BioSequence) int {
	return bytes.Compare(a.sequence, b.sequence)
}

func CompareQuality(a, b *BioSequence) int {
	return bytes.Compare(a.qualities, b.qualities)
}

// func CompareAttributeBuillder(key string) Compare {
// 	f := func(a, b *BioSequence) int {
// 		ak, oka := a.GetAttribute(key)
// 		bk, okb := b.GetAttribute(key)

// 		switch {
// 		case !oka && !okb:
// 			return 0
// 		case !oka:
// 			return -1
// 		case !okb:
// 			return +1
// 		}

// 		//av,oka := ak.(constraints.Ordered)
// 		//bv,okb := bk.(constraints.Ordered)

// 		switch {
// 		case !oka && !okb:
// 			return 0
// 		case !oka:
// 			return -1
// 		case !okb:
// 			return +1
// 		}

// 	}

// 	return f
// }
