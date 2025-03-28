package obiformats

import (
	"io"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

	log "github.com/sirupsen/logrus"
)

func WriteFileChunk(
	writer io.WriteCloser,
	toBeClosed bool) ChannelFileChunk {

	obiutils.RegisterAPipe()
	chunk_channel := make(ChannelFileChunk)

	go func() {
		nextToPrint := 0
		toBePrinted := make(map[int]FileChunk)
		for chunk := range chunk_channel {
			if chunk.Order == nextToPrint {
				log.Debugf("Writing chunk: %d of length %d bytes",
					chunk.Order,
					len(chunk.Raw.Bytes()))

				n, err := writer.Write(chunk.Raw.Bytes())

				if err != nil {
					log.Fatalf("Cannot write chunk %d only %d bytes written on %d sended : %v",
						chunk.Order, n, len(chunk.Raw.Bytes()), err)
				}
				nextToPrint++

				chunk, ok := toBePrinted[nextToPrint]
				for ok {
					log.Debug("Writing buffered chunk : ", chunk.Order)
					_, _ = writer.Write(chunk.Raw.Bytes())
					delete(toBePrinted, nextToPrint)
					nextToPrint++
					chunk, ok = toBePrinted[nextToPrint]
				}
			} else {
				toBePrinted[chunk.Order] = chunk
			}
		}

		log.Debugf("FIle have to be closed : %v", toBeClosed)
		if toBeClosed {
			err := writer.Close()
			if err != nil {
				log.Fatalf("Cannot close the writer : %v", err)
			}
		}

		obiutils.UnregisterPipe()
		log.Debugf("The writer has been closed")
	}()

	return chunk_channel
}
