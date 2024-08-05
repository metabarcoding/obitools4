package obiformats

import (
	"bytes"
	"io"
	"slices"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiseq"
	log "github.com/sirupsen/logrus"
)

type SeqFileChunkParser func(string, io.Reader) (obiseq.BioSequenceSlice, error)

type SeqFileChunk struct {
	Source string
	Raw    *bytes.Buffer
	Order  int
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
func ReadSeqFileChunk(
	source string,
	reader io.Reader,
	buff []byte,
	splitter LastSeqRecord) ChannelSeqFileChunk {
	var err error
	var fullbuff []byte

	chunk_channel := make(ChannelSeqFileChunk)

	fileChunkSize := len(buff)

	go func() {
		size := 0
		l := 0
		i := 0

		// Initialize the buffer to the size of a chunk of data
		fullbuff = buff

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
				buff = slices.Grow(buff, fileChunkSize)
				l := len(buff)
				extbuff := buff[l:(l + fileChunkSize - 1)]
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
					chunk_channel <- SeqFileChunk{source, io, i}
					i++
				}

				if lremain > 0 {
					buff = fullbuff[0:lremain]
					lcp := copy(buff, fullbuff[pnext:])
					if lcp < lremain {
						log.Fatalf("Error copying remaining data of chunk %d : %d < %d", i, lcp, lremain)
					}
				} else {
					buff = buff[:0]
				}

			}
		}

		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Fatalf("Error reading data from file : %s", err)
		}

		// Send the last chunk to the channel
		if len(buff) > 0 {
			io := bytes.NewBuffer(slices.Clone(buff))
			chunk_channel <- SeqFileChunk{source, io, i}
		}

		// Close the readers channel when the end of the file is reached
		close(chunk_channel)
	}()

	return chunk_channel

}
