package obiformats

import (
	"bytes"
	"io"
	"slices"
	"strings"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
	log "github.com/sirupsen/logrus"
)

type SeqFileChunkParser func(string, io.Reader) (obiseq.BioSequenceSlice, error)

type FileChunk struct {
	Source string
	Raw    *bytes.Buffer
	Order  int
}

type PieceOfChunk struct {
	head *PieceOfChunk
	next *PieceOfChunk
	data []byte
}

func NewPieceOfChunk(size int) *PieceOfChunk {
	data := make([]byte, size)
	p := &PieceOfChunk{
		next: nil,
		data: data,
	}
	p.head = p
	return p
}

func (piece *PieceOfChunk) NewPieceOfChunk(size int) *PieceOfChunk {
	if piece == nil {
		return NewPieceOfChunk(size)
	}

	if piece.next != nil {
		log.Panic("Try to create a new piece of chunk when next already exist")
	}

	n := NewPieceOfChunk(size)
	n.head = piece.head
	piece.next = n

	return n
}

func (piece *PieceOfChunk) Next() *PieceOfChunk {
	return piece.next
}

func (piece *PieceOfChunk) Head() *PieceOfChunk {
	if piece == nil {
		return nil
	}
	return piece.head
}

func (piece *PieceOfChunk) Len() int {
	if piece == nil {
		return 0
	}

	if piece.next == nil {
		return len(piece.data)
	}
	return len(piece.data) + piece.next.Len()
}

func (piece *PieceOfChunk) Pack() *PieceOfChunk {
	if piece == nil {
		return nil
	}
	size := piece.next.Len()
	piece.data = slices.Grow(piece.data, size)

	for p := piece.next; p != nil; {
		piece.data = append(piece.data, p.data...)
		p.data = nil
		n := p.next
		p.next = nil
		p = n
	}

	piece.next = nil

	return piece
}

func (piece *PieceOfChunk) IsLast() bool {
	return piece.next == nil
}

func (piece *PieceOfChunk) FileChunk(source string, order int) FileChunk {
	piece.Pack()
	return FileChunk{
		Source: source,
		Raw:    bytes.NewBuffer(piece.data),
		Order:  order,
	}
}

type ChannelFileChunk chan FileChunk

type LastSeqRecord func([]byte) int

func ispossible(data []byte, probe string) bool {
	s := obiutils.UnsafeString(data)
	return strings.Index(s, probe) != -1
}

// _ReadFlatFileChunk reads a chunk of data from the given 'reader' and sends it to the
// 'readers' channel as a _FileChunk struct. The function reads from the reader until
// the end of the last entry is found, then sends the chunk to the channel. If the end
// of the last entry is not found in the current chunk, the function reads from the reader
// in 1 MB increments until the end of the last entry is found. The function repeats this
// process until the end of the file is reached.
//
// Arguments:
// reader io.Reader - an io.Reader to read data from
// readers chan _FileChunk - a channel to send the data as a _FileChunk struct
//
// Returns:
// None
func ReadFileChunk(
	source string,
	reader io.Reader,
	fileChunkSize int,
	splitter LastSeqRecord,
	probe string) ChannelFileChunk {

	chunk_channel := make(ChannelFileChunk)

	go func() {
		var err error
		size := 0
		l := 0
		i := 0

		pieces := NewPieceOfChunk(fileChunkSize)
		// Initialize the buffer to the size of a chunk of data

		// Read from the reader until the buffer is full or the end of the file is reached
		l, err = io.ReadFull(reader, pieces.data)
		pieces.data = pieces.data[:l]

		if err == io.ErrUnexpectedEOF {
			err = nil
		}

		end := splitter(pieces.data)

		// Read from the reader until the end of the last entry is found or the end of the file is reached
		for err == nil {
			// Create an extended buffer to read from if the end of the last entry is not found in the current buffer

			// Read from the reader in 1 MB increments until the end of the last entry is found
			for err == nil && end < 0 {
				pieces = pieces.NewPieceOfChunk(fileChunkSize)
				size, err = io.ReadFull(reader, pieces.data)
				pieces.data = pieces.data[:size]

				if ispossible(pieces.data, probe) {
					pieces = pieces.Head().Pack()
					end = splitter(pieces.data)
				} else {
					end = -1
				}
				// obilog.Warnf("Splitter not found, attempting %d to read in %d B increments : len(buff) = %d/%d", ic, fileChunkSize, len(extbuff), len(buff))
			}

			pieces = pieces.Head().Pack()
			lbuff := pieces.Len()

			if lbuff > 0 {
				if end < 0 {
					end = pieces.Len()
				}

				lremain := lbuff - end

				var nextpieces *PieceOfChunk

				if lremain > 0 {
					nextpieces = NewPieceOfChunk(lremain)
					lcp := copy(nextpieces.data, pieces.data[end:])
					if lcp < lremain {
						log.Fatalf("Error copying remaining data of chunk %d : %d < %d", i, lcp, lremain)
					}
				} else {
					nextpieces = nil
				}

				pieces.data = pieces.data[:end]

				for len(pieces.data) > 0 && (pieces.data[len(pieces.data)-1] == '\n' || pieces.data[len(pieces.data)-1] == '\r') {
					pieces.data = pieces.data[:len(pieces.data)-1]
				}

				if len(pieces.data) > 0 {
					// obilog.Warnf("chuck %d :Read %d bytes from file %s", i, io.Len(), source)
					chunk_channel <- pieces.FileChunk(source, i)
					i++
				}

				pieces = nextpieces
				end = -1
			}
		}

		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Fatalf("Error reading data from file : %s", err)
		}

		pieces.Head().Pack()

		// Send the last chunk to the channel
		if pieces.Len() > 0 {
			chunk_channel <- pieces.FileChunk(source, i)
		}

		// Close the readers channel when the end of the file is reached
		close(chunk_channel)
	}()

	return chunk_channel

}
