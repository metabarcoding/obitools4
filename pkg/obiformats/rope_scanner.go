package obiformats

import "bytes"

// ropeScanner reads lines from a PieceOfChunk rope without heap allocation.
// The carry buffer (stack) handles lines that span two rope nodes.
type ropeScanner struct {
	current *PieceOfChunk
	pos     int
	carry   [256]byte // 256 gives ample margin for typical flat-file lines
	carryN  int
}

func newRopeScanner(rope *PieceOfChunk) *ropeScanner {
	return &ropeScanner{current: rope}
}

// ReadLine returns the next line without the trailing \n (or \r\n).
// Returns nil at end of rope. The returned slice aliases carry[] or the node
// data and is valid only until the next ReadLine call.
func (s *ropeScanner) ReadLine() []byte {
	for {
		if s.current == nil {
			if s.carryN > 0 {
				n := s.carryN
				s.carryN = 0
				return s.carry[:n]
			}
			return nil
		}

		data := s.current.data[s.pos:]
		idx := bytes.IndexByte(data, '\n')

		if idx >= 0 {
			var line []byte
			if s.carryN == 0 {
				line = data[:idx]
			} else {
				n := copy(s.carry[s.carryN:], data[:idx])
				s.carryN += n
				line = s.carry[:s.carryN]
				s.carryN = 0
			}
			s.pos += idx + 1
			if s.pos >= len(s.current.data) {
				s.current = s.current.Next()
				s.pos = 0
			}
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			return line
		}

		// No \n in this node: accumulate into carry and advance
		n := copy(s.carry[s.carryN:], data)
		s.carryN += n
		s.current = s.current.Next()
		s.pos = 0
	}
}

// skipToNewline advances the scanner past the next '\n'.
func (s *ropeScanner) skipToNewline() {
	for s.current != nil {
		data := s.current.data[s.pos:]
		idx := bytes.IndexByte(data, '\n')
		if idx >= 0 {
			s.pos += idx + 1
			if s.pos >= len(s.current.data) {
				s.current = s.current.Next()
				s.pos = 0
			}
			return
		}
		s.current = s.current.Next()
		s.pos = 0
	}
}
