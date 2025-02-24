package obialign

// import (
// 	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
// 	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
// )

const wsize = 16
const dwsize = wsize * 2

// Out values are always the smallest
// Among in values, they rank according to their score
// For equal score the shortest path is the best
func encodeValues(score, length int, out bool) uint64 {
	const mask = uint64(1<<wsize) - 1
	us := uint64(score)
	fo := (us << wsize) | (uint64((^length)-1) & mask)
	if !out {
		fo |= (uint64(1) << dwsize)
	}
	return fo
}

// func _isout(value uint64) bool {
// 	const outmask = uint64(1) << dwsize
// 	return (value & outmask) == 0
// }

// func _lpath(value uint64) int {
// 	const mask = uint64(1<<wsize) - 1
// 	return int(((value + 1) ^ mask) & mask)
// }

func decodeValues(value uint64) (int, int, bool) {
	const mask = uint64(1<<wsize) - 1
	const outmask = uint64(1) << dwsize
	score := int((value >> wsize) & mask)
	length := int(((value + 1) ^ mask) & mask)
	out := (value & outmask) == 0
	return score, length, out
}

func _incpath(value uint64) uint64 {
	return value - 1
}

func _incscore(value uint64) uint64 {
	const incr = uint64(1) << wsize
	return value + incr
}

func _setout(value uint64) uint64 {
	const outmask = ^(uint64(1) << dwsize)
	return value & outmask
}

var _empty = encodeValues(0, 0, false)
var _out = encodeValues(0, 30000, true)
var _notavail = encodeValues(0, 30000, false)
