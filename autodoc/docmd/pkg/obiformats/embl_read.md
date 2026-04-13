# EMBL Format Parser for OBITools4

This Go package (`obiformats`) provides robust, streaming parsers for the **EMBL nucleotide sequence format**, supporting both standard and rope-based (memory-efficient) parsing. Key features:

- **Entry Boundary Detection**: `EndOfLastFlatFileEntry()` identifies the end of EMBL entries using the signature terminator pattern `//` (with optional CR/LF), enabling chunked file processing.
- **Two Parsing Modes**:
  - `EmblChunkParser()`: Line-scanning parser for buffered I/O (`io.Reader`).
  - `EmblChunkParserRope()`: Direct rope-based parser for zero-copy processing of large files.
- **Configurable Options**:
  - `withFeatureTable`: Includes EMBL feature table (`FH`/`FT`) lines.
  - `UtoT`: Converts RNA uracil (`u/U`) to DNA thymine (`t/T`).
- **Metadata Extraction**: Captures `ID`, `OS` (scientific name), `DE` (description), and taxonomic ID (`/db_xref="taxon:..."`) into sequence annotations.
- **Sequence Handling**: Parses multi-line EMBL sequences (10-bases-per-group, with position numbers), skipping digits and whitespace.
- **Parallel Processing**: `ReadEMBL()`/`ReadEMBLFromFile()` support concurrent parsing via worker goroutines, streaming results as `BioSequenceBatch` objects.
- **Integration**: Outputs are compatible with OBITools4’s iterator framework (`obiiter.IBioSequence`) and sequence type `obiseq.BioSequence`.

Designed for scalability, the module handles large EMBL files efficiently—ideal for metagenomic or biodiversity data pipelines.
