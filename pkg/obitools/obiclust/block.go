package obiclust

// A block stores the 4-mer tables of a few centers, transposed: the counter
// of the word w for the lane j is data[w*width+j]. A vector instruction then
// handles the same word for all the centers of the block, and only the words
// that a sequence contains (a hundred for a short amplicon, out of 256) need
// to be visited to compare it with all the centers of the block.
type tableBlock struct {
	data []uint8
}

func newTableBlock(width int) tableBlock {
	return tableBlock{data: make([]uint8, 256*width)}
}

// set stores the table t as the lane `lane` of the block.
func (b *tableBlock) set(width, lane int, t *table4) {
	for w := range t.count {
		b.data[w*width+lane] = t.count[w]
	}
}

// common computes the sums of the minima of the counters of the sequence
// given by its words with every lane of the block (see blockCommon).
func (b *tableBlock) common(width int, words []uint16, out []uint8) {
	blockCommon(b.data, width, words, out)
}

// sparseWords lists the words of a table as w<<8|count, so that the visit of
// a block only goes through the words present in the sequence. nil is
// returned when the counters add up to more than 255, which would overflow the
// 8-bit sums of the blocks.
func sparseWords(t *table4) []uint16 {
	words := make([]uint16, 0, 128)
	mass := 0
	for w, c := range t.count {
		if c > 0 {
			words = append(words, uint16(w)<<8|uint16(c))
			mass += int(c)
		}
	}
	if mass > 255 {
		return nil
	}
	return words
}
