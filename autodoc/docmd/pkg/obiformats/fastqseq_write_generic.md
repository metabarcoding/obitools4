## Semantic Description of `obiformats` Package

The `obiformats` package provides core formatting utilities for biological sequence data in standard FASTX formats (FASTA and FASTQ). It defines two functional types:  
- `BioSequenceFormater`: Converts a single biological sequence (`*obiseq.BioSequence`) into its string representation.  
- `BioSequenceBatchFormater`: Converts a batch of sequences (`obiiter.BioSequenceBatch`) into raw bytes, suitable for file or stream output.

Two main constructor functions enable flexible formatting:  
- `BuildFastxSeqFormater(format, header)` returns a sequence-level formatter based on the requested format (`"fasta"` or `"fastq"`), applying optional header metadata via `FormatHeader`.  
- `BuildFastxFormater(format, header)` builds a batch formatter by composing the sequence-level function over all sequences in an iterator-driven batch, concatenating results with newline separators.

The package supports extensibility and type safety through function composition while integrating logging (via `logrus`) for critical errors—e.g., unsupported formats trigger a fatal log. It abstracts away low-level I/O, focusing purely on *semantic formatting logic*, making it ideal for pipeline integration in NGS data processing tools.
