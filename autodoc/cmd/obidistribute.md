# NAME

obidistribute — divided an input set of sequences into subsets

---

# SYNOPSIS

```
obidistribute --pattern|-p <string> [--append|-A] [--batch-mem <string>]
              [--batch-size <int>] [--batch-size-max <int>]
              [--batches|-n <int>] [--classifier|-c <string>] [--compress|-Z]
              [--csv] [--debug] [--directory|-d <string>] [--ecopcr] [--embl]
              [--fasta] [--fasta-output] [--fastq] [--fastq-output]
              [--genbank] [--hash|-H <int>] [--help|-h|-?]
              [--input-OBI-header] [--input-json-header] [--json-output]
              [--max-cpu <int>] [--na-value <string>] [--no-order]
              [--no-progressbar] [--out|-o <FILENAME>]
              [--output-OBI-header|-O] [--output-json-header] [--pprof]
              [--pprof-goroutine <int>] [--pprof-mutex <int>]
              [--silent-warning] [--skip-empty] [--solexa] [--u-to-t]
              [--version] [<args>]
```

---

# DESCRIPTION

`obidistribute` splits a set of biological sequences into multiple output files according to one of three distribution strategies: annotation-based classification, round-robin batch assignment, or hash-based sharding.

The most common use case in metabarcoding is demultiplexing: sequences carry a tag annotation (e.g., `sample_id`) and `obidistribute` writes each sample's sequences into its own file. The output filename for each group is built from a user-supplied pattern containing `%s`, which is replaced by the classifier value or batch index.

When no classifier is specified, sequences can be split into a fixed number of batches (`--batches`) for parallel downstream processing, or sharded deterministically by hash (`--hash`) to ensure reproducible partitioning regardless of input order.

Output files can be organised into subdirectories (one per classifier value) using `--directory`, and existing files can be extended rather than overwritten with `--append`. Sequences lacking the classifier annotation are assigned to a file whose name uses the NA value (default: `"NA"`).

---

# INPUT

`obidistribute` reads biological sequences from one or more files supplied as positional arguments, or from standard input when no files are given. All major NGS and flat-file formats are supported and auto-detected:

- FASTA / FASTQ (plain or gzip-compressed)
- GenBank and EMBL flat files
- ecoPCR output
- CSV

Format can be forced with `--fasta`, `--fastq`, `--embl`, `--genbank`, `--ecopcr`, or `--csv`. Header annotation style can be specified with `--input-OBI-header` or `--input-json-header`.

---

# OUTPUT

Each distribution group produces a separate output file named according to the `--pattern` template. The `%s` placeholder in the pattern is replaced by the classifier value, batch index, or hash shard index, depending on the chosen distribution mode.

Output format follows the same rules as other OBITools commands: FASTQ is used when quality scores are present, FASTA otherwise. The format can be forced with `--fasta-output`, `--fastq-output`, or `--json-output`. All annotations present in the input sequences are preserved in the output files.

When `--directory` is used together with `--classifier`, output files are placed in subdirectories named after the classifier values, allowing hierarchical organisation of results.

## Observed output example

```
@seq001 {"sample_id":"sampleA"}
atcgatcgatcgatcgatcg
+
IIIIIIIIIIIIIIIIIIII
@seq002 {"sample_id":"sampleA"}
gctagctagctagctagcta
+
IIIIIIIIIIIIIIIIIIII
@seq003 {"sample_id":"sampleA"}
ttagctaatcggtaatcggt
+
IIIIIIIIIIIIIIIIIIII
@seq009 {"sample_id":"sampleA"}
atgatgatgatgatgatgat
+
IIIIIIIIIIIIIIIIIIII
```

---

# OPTIONS

## Distribution mode

- **`--pattern|-p <string>`** — _(required)_
  Default: none.
  The template used to build output filenames. The variable part is represented by `%s`. Example: `toto_%s.fastq`.

- **`--classifier|-c <string>`**
  Default: `""`.
  The name of an annotation tag on the sequences. Sequences are dispatched into separate files based on the value of this tag. The tag value must be a string, integer, or boolean.

- **`--batches|-n <int>`**
  Default: `0`.
  Splits the input into exactly *N* batches by round-robin assignment, regardless of sequence metadata.

- **`--hash|-H <int>`**
  Default: `0`.
  Splits the input into at most *N* batches using a hash of the sequence. Produces deterministic, reproducible sharding.

- **`--directory|-d <string>`**
  Default: `""`.
  Used together with `--classifier`: organises output files into subdirectories named after classifier values.

## Output file handling

- **`--append|-A`**
  Default: `false`.
  Appends sequences to output files if they already exist, instead of overwriting them.

- **`--na-value <string>`**
  Default: `"NA"`.
  Value used as the filename component when a sequence does not have the classifier tag defined.

- **`--compress|-Z`**
  Default: `false`.
  Compresses all output files using gzip.

## Input format

- **`--fasta`**
  Default: `false`.
  Read data following the FASTA format.

- **`--fastq`**
  Default: `false`.
  Read data following the FASTQ format.

- **`--embl`**
  Default: `false`.
  Read data following the EMBL flatfile format.

- **`--genbank`**
  Default: `false`.
  Read data following the GenBank flatfile format.

- **`--ecopcr`**
  Default: `false`.
  Read data following the ecoPCR output format.

- **`--csv`**
  Default: `false`.
  Read data following the CSV format.

- **`--input-OBI-header`**
  Default: `false`.
  FASTA/FASTQ title line annotations follow OBI format.

- **`--input-json-header`**
  Default: `false`.
  FASTA/FASTQ title line annotations follow JSON format.

- **`--solexa`**
  Default: `false`.
  Decodes quality string according to the Solexa specification.

- **`--u-to-t`**
  Default: `false`.
  Convert Uracil to Thymine.

- **`--skip-empty`**
  Default: `false`.
  Sequences of length equal to zero are suppressed from the output.

- **`--no-order`**
  Default: `false`.
  When several input files are provided, indicates that there is no order among them.

## Output format

- **`--fasta-output`**
  Default: `false`.
  Write sequences in FASTA format (default if no quality data available).

- **`--fastq-output`**
  Default: `false`.
  Write sequences in FASTQ format (default if quality data available).

- **`--json-output`**
  Default: `false`.
  Write sequences in JSON format.

- **`--output-OBI-header|-O`**
  Default: `false`.
  Output FASTA/FASTQ title line annotations follow OBI format.

- **`--output-json-header`**
  Default: `false`.
  Output FASTA/FASTQ title line annotations follow JSON format.

- **`--out|-o <FILENAME>`**
  Default: `"-"`.
  Filename used for saving the output.

## Performance

- **`--max-cpu <int>`**
  Default: `16`.
  Number of parallel threads computing the result.

- **`--batch-size <int>`**
  Default: `1`.
  Minimum number of sequences per batch.

- **`--batch-size-max <int>`**
  Default: `2000`.
  Maximum number of sequences per batch.

- **`--batch-mem <string>`**
  Default: `""` (128M).
  Maximum memory per batch (e.g. `128K`, `64M`, `1G`). Set to `0` to disable.

## Diagnostic & debug

- **`--debug`**
  Default: `false`.
  Enable debug mode, by setting log level to debug.

- **`--no-progressbar`**
  Default: `false`.
  Disable the progress bar printing.

- **`--silent-warning`**
  Default: `false`.
  Stop printing of warning messages.

- **`--pprof`**
  Default: `false`.
  Enable pprof server. Look at the log for details.

- **`--pprof-goroutine <int>`**
  Default: `6060`.
  Enable profiling of goroutine blocking profile.

- **`--pprof-mutex <int>`**
  Default: `10`.
  Enable profiling of mutex lock.

---

# EXAMPLES

```bash
# Demultiplex sequences by sample_id annotation into per-sample FASTQ files
obidistribute --classifier sample_id --pattern out_ex1_%s.fastq --no-progressbar --input-json-header reads.fastq
```

**Expected output:** 10 sequences written to 4 files: `out_ex1_sampleA.fastq` (4 sequences), `out_ex1_sampleB.fastq` (3 sequences), `out_ex1_sampleC.fastq` (2 sequences), `out_ex1_NA.fastq` (1 sequence).

```bash
# Demultiplex into subdirectories, one directory per sample
obidistribute --classifier sample_id --directory --pattern %s/reads.fastq reads.fastq
```

```bash
# Split a large dataset into 3 equal batches for parallel processing
obidistribute --batches 3 --pattern chunk_%s.fasta --fasta-output --no-progressbar sequences.fasta
```

**Expected output:** 10 sequences written to 3 files: `chunk_1.fasta` (4 sequences), `chunk_2.fasta` (3 sequences), `chunk_3.fasta` (3 sequences). Batch indices are 1-based.

```bash
# Hash-based sharding into 4 reproducible shards
obidistribute --hash 4 --pattern shard_%s.fastq --no-progressbar reads.fastq
```

**Expected output:** 10 sequences written to 4 files: `shard_0.fastq` through `shard_3.fastq`. Shard indices are 0-based.

```bash
# Append new sequences to existing per-sample files (incremental demultiplexing)
obidistribute --classifier sample_id --pattern samples_%s.fastq --append new_reads.fastq
```

```bash
# Demultiplex sequences, replacing the NA label for unclassified sequences
obidistribute --classifier sample_id --na-value unclassified --pattern out_ex6_%s.fastq --no-progressbar --input-json-header reads.fastq
```

**Expected output:** 10 sequences written to 4 files including `out_ex6_unclassified.fastq` (1 sequence without `sample_id` annotation).

---

# SEE ALSO

`obiconvert`, `obisplit`, `obigrep`

---

# NOTES

- Sequences that lack the annotation specified by `--classifier` are written to the file whose name is built using the `--na-value` (default: `"NA"`).
- The three distribution modes (`--classifier`, `--batches`, `--hash`) are mutually exclusive.
- When using `--directory` together with `--classifier`, subdirectories are created automatically if they do not exist.
- Batch indices produced by `--batches` are 1-based; hash shard indices produced by `--hash` are 0-based.
