# OBIFORMATS Package: Semantic Description

The `obiformats` package provides robust, format-agnostic sequence reading capabilities for biological data in the OBITools4 ecosystem.

It supports automatic detection and parsing of common bioinformatics file formats via MIME-type inference:
- **FASTA** (`text/fasta`): identified by lines starting with `>`.
- **FASTQ** (`text/fastq`): detected via leading `@` characters.
- **ecoPCR2**: recognized by the header line `#@ecopcr-v2`.
- **EMBL** (`text/embl`): detected by lines starting with `ID   `.
- **GenBank** (`text/genbank`): identified by either `LOCUS       ` or legacy `"Genetic Sequence Data Bank"` headers.
- **CSV** (`text/csv`): generic tabular support.

Core functionality is exposed through:
- `OBIMimeTypeGuesser()`: inspects the first ~1 MiB of an input stream to infer MIME type using `github.com/gabriel-vasile/mimetype`, while preserving unread data for downstream processing.
- `ReadSequencesFromFile()`: reads sequences from a file path, infers format via MIME detection, and dispatches to dedicated parsers (e.g., `ReadFasta`, `ReadFastq`).
- `ReadSequencesFromStdin()`: convenience wrapper to read from stdin, treating `"-"` as filename and auto-closing the stream.

Internally leverages:
- `obiutils.Ropen()` for unified file opening (including stdin handling).
- Path extension stripping and source tagging via `OptionsSource()`.
- Logging (`logrus`) for format diagnostics.
- Iterator interface (`obiiter.IBioSequence`) to abstract sequential access over sequences.

The package ensures extensibility: new formats can be added by extending the `switch` dispatch in `ReadSequencesFromFile()` and registering corresponding MIME types.

Error handling covers empty files, invalid streams, and unsupported formats via explicit logging or fatal exits.
