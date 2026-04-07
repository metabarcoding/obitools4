# `obiformats` Package — Semantic Overview

The **`obiformats`** package provides a unified, extensible framework for parsing and writing biological sequence data in standard bioinformatics formats (FASTA/FASTQ, EMBL, GenBank, CSV, EcoPCR), while supporting streaming, batching, parallelism, and format-agnostic workflows.

## Core Objectives

1. **Format-Agnostic Input**: Automatically detect and parse diverse sequence formats via MIME-type inference.
2. **Streaming & Scalability**: Enable memory-efficient ingestion of large NGS datasets through chunked, concurrent parsing.
3. **Structured Output**: Support flexible export to FASTA/FASTQ, JSON, CSV, Newick, and taxonomy-aware formats.
4. **Interoperability**: Integrate seamlessly with OBITools4 abstractions (`obiseq.BioSequence`, `obiiter.IBioSequence`, `obitax.Taxon`).
5. **Extensibility**: Allow new readers/writers to be plugged in via functional interfaces and options.

---

## Public Functionalities (Grouped by Domain)

### 📥 **Sequence Reading & Parsing**

| Function | Format(s) Supported |
|---------|---------------------|
| `ReadSequencesFromFile`, `ReadSequencesFromStdin` | Auto-detected (FASTA/FASTQ/EMBL/GenBank/EcoPCR/CSV) |
| `ReadFasta`, `ReadFastq` | FASTA, FASTQ (with rope/buffered variants) |
| `ReadEMBL`, `ReadGenbank` | EMBL, GenBank (rope-aware for large files) |
| `ReadCSV`, `ReadEcoPCR` | Tabular/amplicon outputs (e.g., EcoPCR v1/v2) |
| `LoadCSVTaxonomy`, `LoadNCBITaxDump` | Taxonomic data (CSV, NCBI dump dir/tar) |

- **Concurrent Parsing**: Configurable worker count (`OptionsParallelWorkers`) with ordered batch output.
- **Rope-Based Parsing**: Zero-copy parsing for large files (`FastaChunkParserRope`, `EmblChunkParserRope`).
- **Header Parsing**: JSON (`ParseFastSeqJsonHeader`) and legacy OBI-style (`ParseOBIFeatures`).
- **Quality Handling**: Phred offset adjustment, optional `U→T` conversion.

### 📤 **Sequence Writing & Formatting**

| Function | Format(s) Supported |
|---------|---------------------|
| `WriteFasta`, `FormatFastq` | FASTA/FASTQ (single/batch, parallel I/O) |
| `WriteJSON` | Structured JSON with annotations (batched + ordered writes) |
| `FormatFastaBatch`, `WriteFastqToFile` | Optimized batch formatting with compression |
| `CSVTaxaIterator`, `CSVSequenceRecord` | Taxonomic/sequence CSV export (configurable columns) |
| `WriteNewick`, `Tree.Newick` | Taxonomy → Newick tree (with optional annotations) |

- **Compression Support**: Automatic gzip/bgzip via `obiutils.CompressStream`.
- **Paired-End Handling**: Split forward/reverse reads to separate files.
- **Ordered Output**: Preserves sequence order across parallel writes (`WriteFileChunk`).
- **Format-Aware Dispatching**: `WriteSequence()` auto-selects FASTQ/FASTA based on quality presence.

### 🧬 **Taxonomy & Metadata Handling**

| Function | Purpose |
|---------|--------|
| `LoadCSVTaxonomy`, `LoadNCBITarTaxDump` | Load taxonomies from CSV/NCBI dumps |
| `DetectTaxonomyFormat`, `LoadTaxonomy` | Auto-detect and load taxonomy from diverse sources |
| `CSVTaxaIterator`, `WriteNewick` | Export taxonomies to CSV or Newick |
| Taxon annotation extraction (e.g., `taxid`, path, rank) | via structured metadata fields |

- **Root Enforcement**: Ensures presence of NCBI root (`taxid=1`) during loading.
- **Alias Resolution**: Merged taxids mapped to current IDs (`AddAlias`).
- **Flexible Output Fields**: CSV/Newick support configurable metadata (scientific name, taxid, rank, path).

### ⚙️ **Configuration & Options**

- `Options` encapsulates all runtime settings via functional setters (`WithOption`, e.g., `BatchSize(1024)`, `OptionsCompressed(true)`).
- Key options include:
  - I/O: file append/truncate, compression (`OptionsCompressed`)
  - Parsing: header parser toggle, quality read flag
  - Export: CSV columns (`CSVId`, `CSVTaxid`), NA value, separator
  - Taxonomy: include path/root/rank (`OptionWithoutRootPath`, `WithTaxid`)
  - Performance: parallel workers, buffer size
- Defaults ensure safe behavior; options are composable and immutable.

### 🧵 **Streaming & Chunking Primitives**

| Type/Function | Purpose |
|---------------|---------|
| `PieceOfChunk`, `FileChunk` | Rope-based buffers for zero-copy streaming |
| `ReadFileChunk()` | Chunk file by record boundaries (not fixed size) |
| `EndOfLastFastaEntry`, `EndOfLastFastqEntry` | Find last complete record in buffer (for safe splitting) |
| `ropeScanner`, `_readline__` | Line-by-line scanner over ropes (no full materialization) |
| `WriteFileChunk()` | Ordered, thread-safe chunk reassembly |

- Designed for **large-file resilience**: avoids full file load; splits only at valid boundaries.
- Integrates with `obiiter` for push-style streaming iterators.

### 🔍 **Format Detection & Discovery**

| Function | Role |
|---------|------|
| `OBIMimeTypeGuesser`, `NGSFilterCsvDetector` | Content-based MIME detection (e.g., FASTA via `>`, EcoPCR via `#@ecopcr-v2`) |
| `DetectTaxonomyFormat` | Detects NCBI dump, CSV, FASTA/FASTQ as taxonomy sources |
| `OBIMimeNGSFilterTypeGuesser` | Distinguishes legacy vs. CSV NGS filter configs |

- Uses `github.com/gabriel-vasile/mimetype` for robust format sniffing.
- Preserves unread bytes to allow downstream parsers.

### 📋 **Specialized Parsers & Writers**

- `ReadCSVFromStdin`, `_ParseFastqFile`: Convenience wrappers for stdin/file I/O.
- `JSONRecord()`, `FormatFastaBatch()`: Optimized serialization with minimal allocations.
- `_parse_json_*` helpers: High-performance JSON parsing using `jsonparser`.
- `WriteFastaToFile`, `_UnescapeUnicodeCharactersInJSON()`: Robust output handling.

---

## Design Principles

- **Streaming First**: All parsers return `obiiter.IBioSequence` — lazy, batched iterators.
- **Functional Abstraction**: Format handling via `IBatchReader`, `FormatHeader` — decoupled from core logic.
- **Extensibility**: New formats added via `ReadSequencesFromFile()` extension points and MIME registration.
- **Fail-Safe Defaults**: Empty files → empty iterator; missing root taxon → fatal error.
- **Ordered Semantics**: Despite parallelism, batches preserve global order via atomic counters (`nextCounter`).

---

## Integration Highlights

- **Dependencies**: Uses `obiseq`, `obiiter`, `obitax`, and utilities (`obiutils`/`obidefault`) for core data models.
- **Logging**: Structured logs via `logrus` (format detection, errors, progress).
- **Error Handling**: Panics on unrecoverable issues; graceful fallbacks (e.g., `ReadEmptyFile`).
- **Performance**: Rope-based parsing, zero-copy where possible (`unsafe.String`, buffered writes).

> ✅ `obiformats` enables scalable, reproducible NGS data processing — from raw ingestion to structured export.
