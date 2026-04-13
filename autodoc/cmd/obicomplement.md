# NAME

obicomplement — reverse complement of sequences

---

# SYNOPSIS

```
obicomplement [--batch-mem <string>] [--batch-size <int>]
              [--batch-size-max <int>] [--compress|-Z] [--csv] [--debug]
              [--ecopcr] [--embl] [--fail-on-taxonomy] [--fasta]
              [--fasta-output] [--fastq] [--fastq-output] [--genbank]
              [--help|-h|-?] [--input-OBI-header] [--input-json-header]
              [--json-output] [--max-cpu <int>] [--no-order]
              [--no-progressbar] [--out|-o <FILENAME>]
              [--output-OBI-header|-O] [--output-json-header]
              [--paired-with <FILENAME>] [--raw-taxid] [--silent-warning]
              [--skip-empty] [--solexa] [--taxonomy|-t <string>] [--u-to-t]
              [--update-taxid] [--with-leaves] [<args>]
```

---

# DESCRIPTION

`obicomplement` computes the reverse complement of every sequence in the
input. For each input sequence, the nucleotides are first reversed, then
each base is replaced by its Watson–Crick complement (A↔T, C↔G), yielding
the strand that would pair with the original sequence read in the opposite
direction.

When quality scores are present (FASTQ data), they are reversed in the same
order as the sequence so that each quality value remains associated with its
corresponding base. Ambiguous IUPAC characters (e.g. `N`, `R`, `Y`) are
handled correctly and preserved in the output.

This operation is commonly needed when sequences have been sequenced on the
wrong strand, when a primer is designed on the reverse strand, or when
preparing sequences for strand-aware downstream analyses.

The command reads from standard input or from one or more files, processes
sequences in parallel, and writes the result to standard output or to the
file specified with `--out`.

---

# INPUT

`obicomplement` accepts biological sequence data in FASTA, FASTQ, EMBL,
GenBank, ecoPCR output, and CSV formats. When no format flag is given, the
format is inferred automatically from the file contents or extension.

Input is read from standard input when no filename argument is provided, or
from one or more files passed as positional arguments. Gzip-compressed files
are handled transparently.

Paired-end data can be provided with `--paired-with`, which specifies the
file containing the second mate. Both mates are reverse-complemented and
written to separate output files.

---

# OUTPUT

The output is a sequence file in which every sequence is the reverse
complement of the corresponding input sequence. The output format matches
the input by default (FASTA if no quality data, FASTQ if quality data are
present), and can be overridden with `--fasta-output`, `--fastq-output`, or
`--json-output`.

All annotations (attributes stored in the sequence header) are preserved
unchanged. Quality scores, when present, are reversed to stay aligned with
their bases.

## Observed output example

```
>seq001 {"definition":"basic DNA sequence"}
cgatcgatcgatcgatcgat
>seq002 {"definition":"GC-rich sequence"}
gcgcgcgcgcgcgcgcgcgc
>seq003 {"definition":"AT-rich sequence"}
atatatatatatatatatat
>seq004 {"definition":"palindromic sequence"}
aattccggaattccggaatt
>seq005 {"definition":"mixed sequence"}
agctagcatgcatagccgat
```

---

# OPTIONS

## Input format

**`--fasta`**
: Default: false. Force parsing of input as FASTA format.

**`--fastq`**
: Default: false. Force parsing of input as FASTQ format.

**`--embl`**
: Default: false. Force parsing of input as EMBL flatfile format.

**`--genbank`**
: Default: false. Force parsing of input as GenBank flatfile format.

**`--ecopcr`**
: Default: false. Force parsing of input as ecoPCR output format.

**`--csv`**
: Default: false. Force parsing of input as CSV format.

**`--solexa`**
: Default: false. Decode quality scores using the Solexa/Illumina pre-1.3
  convention instead of the standard Phred+33 encoding.

**`--input-OBI-header`**
: Default: false. Interpret FASTA/FASTQ header annotations using the OBI
  key=value format.

**`--input-json-header`**
: Default: false. Interpret FASTA/FASTQ header annotations using JSON
  format.

**`--no-order`**
: Default: false. When several input files are given, declare that no
  ordering relationship exists among them, allowing the reader to interleave
  records freely.

**`--paired-with <FILENAME>`**
: Default: none. File containing the paired (R2) reads. When set,
  `obicomplement` processes both mates and writes them to separate output
  files.

## Sequence preprocessing

**`--u-to-t`**
: Default: false. Convert Uracil (U) to Thymine (T) before computing the
  reverse complement. Useful when processing RNA sequences that must be
  treated as DNA.

**`--skip-empty`**
: Default: false. Discard sequences of length zero from the output.

## Output format

**`--fasta-output`**
: Default: false. Write output in FASTA format regardless of whether quality
  scores are present.

**`--fastq-output`**
: Default: false. Write output in FASTQ format (requires quality data).

**`--json-output`**
: Default: false. Write output in JSON format.

**`--out|-o <FILENAME>`**
: Default: `-` (standard output). File used to save the output.

**`--output-OBI-header|-O`**
: Default: false. Write FASTA/FASTQ header annotations in OBI key=value
  format.

**`--output-json-header`**
: Default: false. Write FASTA/FASTQ header annotations in JSON format.

**`--compress|-Z`**
: Default: false. Compress the output with gzip.

## Taxonomy

**`--taxonomy|-t <string>`**
: Default: none. Path to a taxonomy database. Required only when the input
  sequences carry taxid annotations that need to be validated or updated.

**`--fail-on-taxonomy`**
: Default: false. Cause `obicomplement` to exit with an error if a taxid
  referenced in the data is not a currently valid node in the loaded
  taxonomy.

**`--update-taxid`**
: Default: false. Automatically replace taxids that have been declared
  merged into a newer node by the taxonomy database.

**`--raw-taxid`**
: Default: false. Print taxids without appending the taxon name and rank.

**`--with-leaves`**
: Default: false. When the taxonomy is extracted from the sequence file,
  attach sequences as leaves of their taxid node.

## Performance and diagnostics

**`--max-cpu <int>`**
: Default: 16 (env: `OBIMAXCPU`). Number of parallel threads used to
  process sequences.

**`--batch-size <int>`**
: Default: 1 (env: `OBIBATCHSIZE`). Minimum number of sequences per
  processing batch.

**`--batch-size-max <int>`**
: Default: 2000 (env: `OBIBATCHSIZEMAX`). Maximum number of sequences per
  processing batch.

**`--batch-mem <string>`**
: Default: `128M` (env: `OBIBATCHMEM`). Maximum memory allocated per batch
  (e.g. `128K`, `64M`, `1G`). Set to `0` to disable the memory limit.

**`--no-progressbar`**
: Default: false. Disable the progress bar printed to stderr.

**`--silent-warning`**
: Default: false (env: `OBIWARNING`). Suppress warning messages.

**`--debug`**
: Default: false (env: `OBIDEBUG`). Enable debug logging.

---

# EXAMPLES

```bash
# Reverse complement all sequences in a FASTA file
obicomplement sequences.fasta > out_default.fasta
```

**Expected output:** 5 sequences written to `out_default.fasta`.

```bash
# Reverse complement a FASTQ file, preserving quality scores
obicomplement reads.fastq --fastq-output --out out_fastq.fastq
```

**Expected output:** 5 sequences written to `out_fastq.fastq`.

```bash
# Convert RNA sequences to their reverse complement DNA strand
obicomplement --u-to-t rna_sequences.fasta > out_rna_rc.fasta
```

**Expected output:** 3 sequences written to `out_rna_rc.fasta`.

```bash
# Reverse complement paired-end reads into two separate output files
obicomplement R1.fastq --paired-with R2.fastq --out out_paired.fastq
```

**Expected output:** 3 sequences written to `out_paired_R1.fastq` and 3 sequences to `out_paired_R2.fastq`.

```bash
# Reverse complement and compress output, skipping any empty sequences
obicomplement --skip-empty --compress sequences.fasta --out out_compressed.fasta.gz
```

**Expected output:** 5 sequences written to `out_compressed.fasta.gz` (gzip-compressed FASTA).

```bash
# Reverse complement with OBI-format header output
obicomplement --output-OBI-header sequences.fasta --out out_obi.fasta
```

**Expected output:** 5 sequences written to `out_obi.fasta`.

```bash
# Reverse complement with explicit JSON-format header output
obicomplement --output-json-header sequences.fasta --out out_jsonheader.fasta
```

**Expected output:** 5 sequences written to `out_jsonheader.fasta`.

```bash
# Reverse complement and write full JSON output format
obicomplement --json-output sequences.fasta --out out_json.json
```

**Expected output:** 5 sequences written to `out_json.json`.

---

# SEE ALSO

- `obiconvert` — format conversion and sequence filtering pipeline
- `obipairing` — paired-end read merging (uses reverse complement internally)
- `obigrep` — sequence filtering and selection

---

# NOTES

Quality scores (Phred-scaled) are reversed in lock-step with the sequence
so that positional quality information remains valid after the reverse
complement operation. This is essential for downstream tools that rely on
per-base quality for alignment or variant calling.

Ambiguous IUPAC characters and gap symbols (`-`) are handled gracefully:
standard ambiguous bases are complemented according to IUPAC rules, while
gap and missing-data symbols are preserved unchanged.
