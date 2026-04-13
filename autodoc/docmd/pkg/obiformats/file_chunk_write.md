# `WriteFileChunk` Function — Semantic Description

The `WriteFileChunk` function in the `obiformats` package implements a **thread-safe, ordered chunk writer** for streaming data to an `io.WriteCloser`. It accepts a destination writer and a flag indicating whether the writer should be closed upon completion.

- **Input**:  
  - `writer`: An `io.WriteCloser` (e.g., file, buffer) to which data chunks are written.  
  - `toBeClosed`: Boolean flag specifying if the writer should be closed after all chunks are processed.

- **Core Behavior**:  
  - Launches a goroutine that consumes `FileChunk` items from an unbuffered channel (`chunk_channel`).  
  - Ensures **strict sequential ordering** of chunks by their `Order` field (intended for reassembly after parallel or out-of-order processing).  
  - If a chunk arrives in order (`chunk.Order == nextToPrint`), it is immediately written.  
  - Out-of-order chunks are buffered in a map (`toBePrinted`) until their predecessor arrives.

- **Buffer Management**:  
  - After writing an in-order chunk, the function checks for newly consecutive buffered chunks and writes them greedily (e.g., if order 2 arrives, it triggers writing of buffered orders 3,4,... as available).

- **Error Handling**:  
  - Logs fatal errors on write failures or writer closure issues using `log.Fatalf`.

- **Cleanup & Lifecycle**:  
  - Closes the underlying writer if requested and unregisters a pipe registration (via `obiutils`) to signal end-of-stream.  
  - Returns the input channel, enabling external producers to stream `FileChunk` structs.

- **Use Case**:  
  Designed for robust, ordered reconstruction of large binary/data streams (e.g., sequencing reads) in OBITools4 pipelines, especially where parallel chunking and reassembly occur.
