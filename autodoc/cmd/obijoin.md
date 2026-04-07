# NAME

obijoin — merge annotations contained in a file to another file

---

# SYNOPSIS

```
obijoin --join-with|-j <string> [--batch-mem <string>] [--batch-size <int>]
        [--batch-size-max <int>] [--by|-b <string>]... [--compress|-Z]
        [--csv] [--debug] [--ecopcr] [--embl] [--fail-on-taxonomy] [--fasta]
        [--fasta-output] [--fastq] [--fastq-output] [--genbank]
        [--help|-h|-?] [--input-OBI-header] [--input-json-header]
        [--json-output] [--max-cpu <int>] [--no-order] [--no-progressbar]
        [--out|-o <FILENAME>] [--output-OBI-header|-O] [--output-json-header]
        [--pprof] [--pprof-goroutine <int>] [--pprof-mutex <int>]
        [--raw-taxid] [--silent-warning] [--skip-empty] [--solexa]
        [--taxonomy|-t <string>] [--u-to-t] [--update-id|-i]
        [--update-quality|-q] [--update-sequence|-s] [--update-taxid]
        [--version] [--with-leaves] [<args>]
```

---

# DESCRIPTION

`obijoin` merges annotations from a secondary file into a primary sequence dataset. For each sequence in the primary input, it looks up matching records in the secondary file based on one or more shared attribute keys, then copies all annotations from matched partner records onto the primary sequence.

The join is a **left outer join**: every sequence in the primary dataset is preserved in the output, whether or not a match is found in the secondary file. Unmatched sequences simply receive no additional annotations. Key matching is exact string equality.

A common use case is enriching amplicon or read sequences with external sample metadata. The secondary file (the *annotation source*) can be a FASTA/FASTQ sequence file, a CSV table, an EMBL or GenBank flat file, or any other format accepted by OBITools4. This makes it straightforward to prepare a simple spreadsheet with sample identifiers and metadata columns, save it as CSV, and merge it directly into a sequence dataset — the CSV format is auto-detected, no format conversion or extra flag is needed. <!-- corrected: secondary CSV is auto-detected; --csv flag is not needed for the secondary file -->

In addition to transferring annotations, `obijoin` can optionally replace the sequence identifier, nucleotide sequence, or quality scores of each primary sequence with values from its matched partner, controlled by the `--update-id`, `--update-sequence`, and `--update-quality` flags.

---

# INPUT

`obijoin` accepts a primary sequence dataset on standard input or as one or more file arguments. The supported formats are automatically detected and include FASTA, FASTQ, EMBL, GenBank, ecoPCR output, CSV, and JSON. Format-specific flags (`--fasta`, `--fastq`, `--embl`, `--genbank`, `--ecopcr`, `--csv`) can force a specific parser when auto-detection is ambiguous.

The secondary file, supplied via `--join-with`, is loaded entirely into memory before processing begins, and supports the same set of formats including CSV — the format is auto-detected automatically. <!-- corrected: removed incorrect claim that --csv is needed for secondary file -->

When multiple primary input files are provided and their ordering across files is irrelevant, `--no-order` allows the reader to return batches in whichever order they complete, improving throughput.

---

# OUTPUT

The output is a sequence file in FASTA or FASTQ format (determined automatically by the presence of quality data), written to standard output or to the file specified by `--out`. Alternative output formats can be requested with `--fasta-output`, `--fastq-output`, or `--json-output`. The output can be gzip-compressed with `--compress`.

Each output sequence carries all annotations from the primary dataset, enriched with every annotation attribute copied from the matched partner record. If a field name exists in both, the partner value overwrites the primary value. When `--update-id`, `--update-sequence`, or `--update-quality` are set, the corresponding sequence-level fields are also replaced with the partner's values.

## Observed output example

```
>seq001 {"barcode":"ATGC","experiment":"amplicon_run1","location":"Paris","sample":"S1"}
atgcatgcatgcatgcatgc
>seq002 {"barcode":"GCTA","experiment":"amplicon_run2","location":"Lyon","sample":"S2"}
gctagctagctagctagcta
>seq003 {"barcode":"TTTT","sample":"S3"}
tttttttttttttttttttt
>seq004 {"barcode":"ATGC","experiment":"amplicon_run1","location":"Paris","sample":"S1"}
aaaaatttttcccccggggg
>seq005 {"barcode":"GCTA","experiment":"amplicon_run2","location":"Lyon","sample":"S2"}
gggggaaaaatttttccccc
>seq006 {"barcode":"AAAA","sample":"S4"}
ccccccgggggtttttaaaaa
```

---

# OPTIONS

## Required

`--join-with|-j <string>`
: Path to the secondary file whose records are joined onto the primary sequences. This parameter is mandatory. The file can be in any format accepted by OBITools4 (FASTA, FASTQ, CSV, EMBL, GenBank, ecoPCR); the format is auto-detected. Default: none.

## Join control

`--by|-b <string>`
: Declares a join key as an attribute name or a `primary_attr=secondary_attr` mapping. Repeat the flag to join on multiple keys simultaneously; all keys must match for a record pair to be considered a hit (intersection semantics). When omitted, the join defaults to matching by sequence identifier (`id`). Default: `[]`.

`--update-id|-i`
: Replace the identifier of each primary sequence with the identifier from its matched partner record. Default: `false`.

`--update-sequence|-s`
: Replace the nucleotide or amino acid sequence of each primary sequence with the sequence from its matched partner. Default: `false`.

`--update-quality|-q`
: Replace the per-base quality scores of each primary sequence with the quality scores from its matched partner. Relevant only when both datasets carry quality information (FASTQ). Default: `false`.

## Input format

`--csv`
: Read the primary input data in OBITools CSV format (e.g., sequences exported by `obicsv`). This flag applies to the primary input only; secondary files supplied via `--join-with` are always auto-detected. Default: `false`. <!-- corrected: --csv affects primary input only, not the secondary annotation file -->

`--ecopcr`
: Read data following the ecoPCR output format. Default: `false`.

`--embl`
: Read data following the EMBL flatfile format. Default: `false`.

`--fasta`
: Read data following the FASTA format. Default: `false`.

`--fastq`
: Read data following the FASTQ format. Default: `false`.

`--genbank`
: Read data following the GenBank flatfile format. Default: `false`.

`--input-OBI-header`
: Treat FASTA/FASTQ title line annotations as OBI format. Default: `false`.

`--input-json-header`
: Treat FASTA/FASTQ title line annotations as JSON format. Default: `false`.

`--solexa`
: Decode the quality string according to the Solexa specification. Default: `false`.

`--u-to-t`
: Convert uracil (U) to thymine (T) in input sequences. Default: `false`.

`--skip-empty`
: Suppress sequences of length zero from the output. Default: `false`.

`--no-order`
: When several input files are provided, indicates that there is no order among them. Default: `false`.

## Output format

`--out|-o <FILENAME>`
: Filename used for saving the output. Default: `-` (standard output).

`--fasta-output`
: Write sequences in FASTA format (default when no quality data are available). Default: `false`.

`--fastq-output`
: Write sequences in FASTQ format (default when quality data are available). Default: `false`.

`--json-output`
: Write sequences in JSON format. Default: `false`.

`--output-OBI-header|-O`
: Output FASTA/FASTQ title line annotations in OBI format. Default: `false`.

`--output-json-header`
: Output FASTA/FASTQ title line annotations in JSON format. Default: `false`.

`--compress|-Z`
: Compress the output using gzip. Default: `false`.

## Taxonomy

`--taxonomy|-t <string>`
: Path to the taxonomy database. Default: `""`.

`--fail-on-taxonomy`
: Cause `obijoin` to fail with an error if a taxid encountered is not currently valid. Default: `false`.

`--raw-taxid`
: Print taxids in files without supplementary information (taxon name and rank). Default: `false`.

`--update-taxid`
: Automatically update taxids that are declared as merged to a newer one. Default: `false`.

`--with-leaves`
: When taxonomy is extracted from a sequence file, add sequences as leaves of their taxid annotation. Default: `false`.

## Performance

`--max-cpu <int>`
: Number of parallel threads used to compute the result. Default: `16`.

`--batch-size <int>`
: Minimum number of sequences per processing batch. Default: `1`.

`--batch-size-max <int>`
: Maximum number of sequences per processing batch. Default: `2000`.

`--batch-mem <string>`
: Maximum memory per batch (e.g. `128K`, `64M`, `1G`). Set to `0` to disable. Default: `128M`.

## Diagnostics

`--no-progressbar`
: Disable the progress bar. Default: `false`.

`--silent-warning`
: Stop printing warning messages. Default: `false`.

`--debug`
: Enable debug mode by setting the log level to debug. Default: `false`.

---

# EXAMPLES

```bash
# Annotate amplicon sequences with sample metadata from a CSV table,
# matching on the sample attribute. CSV format is auto-detected.
obijoin --join-with metadata.csv --by sample input.fasta > out_basic.fasta
```

**Expected output:** 6 sequences written to `out_basic.fasta`.

```bash
# Join using a cross-attribute key: primary sequences have a 'sample' attribute,
# while the annotation CSV uses 'well' for the same identifier.
obijoin --join-with well_metadata.csv --by sample=well input.fasta > out_crosskey.fasta
```

**Expected output:** 6 sequences written to `out_crosskey.fasta`.

```bash
# Join on two keys simultaneously: match only when both sample and barcode agree,
# then update sequence identifiers with those from the reference file.
obijoin --join-with references.fasta \
        --by sample --by barcode \
        --update-id \
        input.fasta > out_multikey.fasta
```

**Expected output:** 6 sequences written to `out_multikey.fasta`.

```bash
# Replace sequences and quality scores of reads with values from a corrected FASTQ file,
# joining by sequence ID (default when no --by is specified).
obijoin --join-with corrected.fastq \
        --update-sequence --update-quality \
        input.fastq > out_updated.fastq
```

**Expected output:** 3 sequences written to `out_updated.fastq`.

```bash
# Use an OBITools CSV file as primary input (--csv flag), join with a metadata CSV,
# then write compressed FASTA output without showing the progress bar.
obijoin --join-with metadata.csv --by sample \
        --csv --fasta-output --compress \
        --no-progressbar \
        --out out_compressed.fasta.gz \
        primary.csv
```

**Expected output:** 3 sequences written to `out_compressed.fasta.gz`.

---

# NOTES

- The secondary file supplied via `--join-with` is loaded entirely into memory before the join begins. For very large secondary files this may require significant RAM.
- Key matching is based on exact string equality; no regular expression or fuzzy matching is applied.
- The join is a left outer join: primary sequences without a matching partner in the secondary file are still emitted, unchanged, in the output.
- When the annotation source is a plain CSV spreadsheet (columns = attributes, rows = records), the format is auto-detected — no `--csv` flag is needed. The `--csv` flag applies exclusively to the primary input and is intended for sequences stored in OBITools CSV format.
