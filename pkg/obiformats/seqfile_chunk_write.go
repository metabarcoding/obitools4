package obiformats

import (
	"io"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiiter"

	log "github.com/sirupsen/logrus"
)

func WriteSeqFileChunk(
	writer io.WriteCloser,
	toBeClosed bool) ChannelSeqFileChunk {

	obiiter.RegisterAPipe()

	chunk_channel := make(ChannelSeqFileChunk)

	go func() {
		nextToPrint := 0
		toBePrinted := make(map[int]SeqFileChunk)
		for chunk := range chunk_channel {
			if chunk.Order == nextToPrint {
				_, _ = writer.Write(chunk.Raw.Bytes())
				nextToPrint++

				chunk, ok := toBePrinted[nextToPrint]
				for ok {
					_, _ = writer.Write(chunk.Raw.Bytes())
					delete(toBePrinted, nextToPrint)
					nextToPrint++
					chunk, ok = toBePrinted[nextToPrint]
				}
			} else {
				toBePrinted[chunk.Order] = chunk
			}
		}

		if toBeClosed {
			err := writer.Close()
			if err != nil {
				log.Fatalf("Cannot close the writer : %v", err)
			}
		}

		obiiter.UnregisterPipe()
		log.Debugf("The writer has been closed")
	}()

	return chunk_channel
}
