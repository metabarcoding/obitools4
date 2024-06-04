package obiformats

import (
	"bytes"
	"io"
	"slices"

	log "github.com/sirupsen/logrus"
)

var _FileChunkSize = 1 << 28

type SeqFileChunk struct {
	raw   io.Reader
	order int
}

type ChannelSeqFileChunk chan SeqFileChunk

type LastSeqRecord func([]byte) int

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
func ReadSeqFileChunk(reader io.Reader,
	splitter LastSeqRecord) ChannelSeqFileChunk {
	var err error
	var fullbuff []byte
	var buff []byte

	chunk_channel := make(ChannelSeqFileChunk)

	go func() {
		size := 0
		l := 0
		i := 0

		// Initialize the buffer to the size of a chunk of data
		fullbuff = make([]byte, _FileChunkSize, _FileChunkSize*2)
		buff = fullbuff

		// Read from the reader until the buffer is full or the end of the file is reached
		l, err = io.ReadFull(reader, buff)
		buff = buff[:l]

		if err == io.ErrUnexpectedEOF {
			err = nil
		}

		// Read from the reader until the end of the last entry is found or the end of the file is reached
		for err == nil {
			// Create an extended buffer to read from if the end of the last entry is not found in the current buffer
			end := 0
			ic := 0

			// Read from the reader in 1 MB increments until the end of the last entry is found
			for end = splitter(buff); err == nil && end < 0; end = splitter(buff) {
				ic++
				buff = slices.Grow(buff, _FileChunkSize)
				l := len(buff)
				extbuff := buff[l:(l + _FileChunkSize - 1)]
				size, err = io.ReadFull(reader, extbuff)
				buff = buff[0:(l + size)]
			}

			fullbuff = buff

			if len(buff) > 0 {
				if end < 0 {
					end = len(buff)
				}

				pnext := end
				lremain := len(buff) - pnext
				buff = buff[:end]
				for len(buff) > 0 && (buff[len(buff)-1] == '\n' || buff[len(buff)-1] == '\r') {
					buff = buff[:len(buff)-1]
				}

				if len(buff) > 0 {
					io := bytes.NewBuffer(slices.Clone(buff))
					chunk_channel <- SeqFileChunk{io, i}
					i++
				}

				if lremain > 0 {
					buff = fullbuff[0:lremain]
					lcp := copy(buff, fullbuff[pnext:])
					if lcp < lremain {
						log.Fatalf("Error copying remaining data of chunck %d : %d < %d", i, lcp, lremain)
					}
				} else {
					buff = buff[:0]
				}

			}
		}

		// Close the readers channel when the end of the file is reached
		close(chunk_channel)
	}()

	return chunk_channel

}
