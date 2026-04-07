# GenBank Parser Module (`obiformats`)

This Go package provides high-performance parsing of **GenBank flat files**, optimized for large-scale genomic data processing. It supports both rope-based (memory-efficient) and buffered I/O parsing strategies.

## Core Functionalities

- **State-machine parser**: Processes GenBank records through well-defined states (`inHeader`, `inEntry`, `inFeature`, etc.), ensuring robust handling of structured sections (LOCUS, DEFINITION, SOURCE, FEATURES, ORIGIN/CONTIG).
- **Rope-aware parsing** (`GenbankChunkParserRope`): Directly parses from a `PieceOfChunk` rope structure, avoiding large contiguous memory allocations—critical for chromosomal-scale sequences.
- **Sequence extraction**: Efficient byte-by-byte scanning of the `ORIGIN` section, compacting bases and optionally converting uracil (`u`) to thymine (`t`).
- **Metadata extraction**: Captures sequence ID, declared length (from LOCUS), scientific name (`SOURCE`), and taxonomic ID (`/db_xref="taxon:..."`).
- **Optional feature table support**: When enabled, stores raw FEATURES section content for downstream annotation processing.
- **Parallel streaming I/O**:
  - `ReadGenbank()` and `ReadGenbankFromFile()` return an iterator (`obiiter.IBioSequence`) over parsed sequences.
  - Supports concurrent parsing via configurable worker count, with chunked file reading and batch output.

## Key Design Decisions

- **Zero-copy where possible**: Rope parser avoids `Pack()` to prevent expensive reallocation.
- **Strict state validation**: Logs fatal errors on unexpected line sequences (e.g., `DEFINITION` outside entry state).
- **Fallback parsing**: Falls back to buffered I/O (`GenbankChunkParser`) when rope data is unavailable.
- **U-to-T conversion**: Optional base modification for RNA→DNA normalization (e.g., in transcriptome data).
- **Error resilience**: Warns on empty IDs but continues processing; rejects overly long lines (>100 chars) in buffered mode.

## Output

Returns a batched iterator of `BioSequence` objects, each containing:
- Identifier (`id`)
- Compact nucleotide sequence
- Definition line (as description)
- Source file origin
- Optional feature table bytes
- Annotations: `scientific_name`, `taxid`

Ideal for pipelines requiring scalable, low-memory GenBank ingestion (e.g., metagenomic databases).
