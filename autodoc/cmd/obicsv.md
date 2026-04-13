# NAME

obicsv — converts sequence files to CSV format

---

# SYNOPSIS

```
obicsv [--auto] [--batch-mem <string>] [--batch-size <int>]
       [--batch-size-max <int>] [--compress|-Z] [--count] [--csv] [--debug]
       [--definition|-d] [--ecopcr] [--embl] [--fail-on-taxonomy] [--fasta]
       [--fastq] [--genbank] [--help|-h|-?] [--ids|-i] [--input-OBI-header]
       [--input-json-header] [--keep|-k <KEY>]... [--max-cpu <int>]
       [--na-value <NAVALUE>] [--no-order] [--no-progressbar] [--obipairing]
       [--out|-o <FILENAME>] [--pprof] [--pprof-goroutine <int>]
       [--pprof-mutex <int>] [--quality|-q] [--raw-taxid] [--sequence|-s]
       [--silent-warning] [--solexa] [--taxon] [--taxonomy|-t <string>]
       [--u-to-t] [--update-taxid] [--version] [--with-leaves] [<args>]
```

---

# DESCRIPTION

obicsv converts biological sequence data into CSV format for easy inspection, spreadsheet analysis, or integration with other tools. A biologist might use it to export sequences from OBITools for quality control, taxonomic inspection, or downstream analysis in R or Python.

Columns must be explicitly selected: use `--ids` for the identifier, `--sequence` for the nucleotide sequence, `--quality` for quality scores, `--taxon` for taxonomic information, `--auto` to auto-detect annotation attributes, or `--keep` for specific named attributes. Multiple flags can be combined freely.

The command uses parallel workers to process large datasets efficiently and can write output to stdout or directly to a file.

---

# INPUT

obicsv accepts input from files or stdin. The input format is automatically detected based on the file extension, but can be explicitly specified using format flags.

Supported input formats:
- FASTA (`--fasta`)
- FASTQ (`--fastq`)
- GenBank (`--genbank`)
- EMBL (`--embl`)
- ecoPCR output (`--ecopcr`)
- CSV (`--csv`)

Input sources:
- Local files (specified as arguments)
- stdin (when no input file is provided)
- Remote URLs (`http://`, `https://`, `ftp://`)
- Directories (automatically scanned for valid files)

Header formats:
- OBI format (`--input-OBI-header`)
- JSON format (`--input-json-header`)
- Auto-detection (default)

Taxonomy database can be provided with `--taxonomy|-t`.

---

# OUTPUT

The output is a CSV file with one row per sequence. The columns included depend on the flags used:

| Column | Flag | Description |
|--------|------|-------------|
| id | `--ids\|-i` | Sequence identifier |
| sequence | `--sequence\|-s` | DNA/RNA sequence |
| qualities | `--quality\|-q` | Quality scores (ASCII-encoded) |
| definition | `--definition\|-d` | Sequence description/annotation |
| count | `--count` | Number of reads represented by this sequence |
| taxid | `--taxon` | NCBI taxonomy identifier |
| scientific_name | `--taxon` | Taxonomic scientific name |
| custom attributes | `--keep\|-k` | Any attribute stored in sequence annotations |

If `--auto` is used, columns are automatically determined based on the attributes present in the first batch of sequences.

Missing values are written as the NA value (default: "NA").

## Observed output example

```csv
id,sequence
seq001,atgcatgcatgcatgcatgcatgcatgcatgcatgcatgcatgcatgcatgcatgcatgc
seq002,ggggaaaattttccccggggaaaattttccccggggaaaattttccccggggaaaatttt
seq003,cccccccccccccccccccccccccccccccccccccccccccccccccccccccccc
```

---

# OPTIONS

## Output Columns

These flags control which columns appear in the CSV output.

- **`--ids|-i`**
  - Default: `false`
  - Meaning: Include the sequence identifier column. Useful for tracking or linking sequences.

- **`--sequence|-s`**
  - Default: `false`
  - Meaning: Include the nucleotide or amino acid sequence. This is the main biological data.

- **`--quality|-q`**
  - Default: `false`
  - Meaning: Include quality scores for each position. Essential for quality control and filtering.

- **`--definition|-d`**
  - Default: `false`
  - Meaning: Include the sequence description or definition from the source file.

- **`--count`**
  - Default: `false`
  - Meaning: Include the count attribute, representing how many original reads were collapsed into this sequence (e.g., from clustering or demultiplexing).

- **`--taxon`**
  - Default: `false`
  - Meaning: Include taxonomic information. Outputs both the NCBI taxid and the scientific name. Requires a taxonomy database (see `--taxonomy`).

- **`--obipairing`**
  - Default: `false`
  - Meaning: Include attributes that were added by the `obipairing` command (pairing scores, mismatches, etc.).

- **`--auto`**
  - Default: `false`
  - Meaning: Automatically detect which columns to output by examining the first batch of sequences. Outputs all annotation attributes found in the headers. Can be combined with `--ids`, `--sequence`, etc. to add those columns on top of the auto-detected ones.

- **`--keep|-k <KEY>`**
  - Default: `none`
  - Meaning: Keep only the specified attribute(s). Can be used multiple times to keep several columns. Useful for extracting specific annotations.

- **`--na-value <NAVALUE>`**
  - Default: `"NA"`
  - Meaning: String to use for missing or unavailable values in the CSV. Customize for compatibility with other tools (e.g., empty string, "NA", "null").

## Input/Output Files

- **`--out|-o <FILENAME>`**
  - Default: `"-"` (stdout)
  - Meaning: Write output to the specified file instead of stdout.

- **`--compress|-Z`**
  - Default: `false`
  - Meaning: Compress the output using gzip.

## Input Format

- **`--fasta`**, **`--fastq`**, **`--genbank`**, **`--embl`**, **`--ecopcr`**, **`--csv`**
  - Default: auto-detection
  - Meaning: Explicitly specify the input format.

- **`--input-OBI-header`**, **`--input-json-header`**
  - Default: auto-detection
  - Meaning: Specify the header format in FASTA/FASTQ files (OBI or JSON annotations).

- **`--u-to-t`**
  - Default: `false`
  - Meaning: Convert Uracil to Thymine. Useful for RNA sequences.

- **`--solexa`**
  - Default: `false`
  - Meaning: Decode quality strings according to the Solexa specification instead of Phred.

## Taxonomy

- **`--taxonomy|-t <string>`**
  - Default: `""`
  - Meaning: Path to the taxonomy database directory. Required for `--taxon` output.

- **`--fail-on-taxonomy`**
  - Default: `false`
  - Meaning: Make OBITools fail if a used taxid is not currently valid.

- **`--update-taxid`**
  - Default: `false`
  - Meaning: Automatically update taxids that have been merged to their newest valid taxid.

- **`--raw-taxid`**
  - Default: `false`
  - Meaning: Print only taxids without supplementary information (name and rank).

- **`--with-leaves`**
  - Default: `false`
  - Meaning: Add sequences as leaves of their taxid annotation when taxonomy is extracted from a sequence file.

## Performance

- **`--max-cpu <int>`**
  - Default: `16`
  - Meaning: Number of parallel threads for processing.

- **`--batch-size <int>`**
  - Default: `1`
  - Meaning: Minimum number of sequences per batch.

- **`--batch-size-max <int>`**
  - Default: `2000`
  - Meaning: Maximum number of sequences per batch.

- **`--batch-mem <string>`**
  - Default: `"128M"`
  - Meaning: Maximum memory per batch (e.g., 128K, 64M, 1G).

- **`--no-order`**
  - Default: `false`
  - Meaning: When multiple input files are provided, indicates there is no order among them.

- **`--no-progressbar`**
  - Default: `false`
  - Meaning: Disable the progress bar.

## Other Options

- **`--debug`**
  - Default: `false`
  - Meaning: Enable debug mode by setting log level to debug.

- **`--pprof`**
  - Default: `false`
  - Meaning: Enable pprof server.

- **`--pprof-goroutine <int>`**
  - Default: `6060`
  - Meaning: Enable profiling of goroutine blocking.

- **`--pprof-mutex <int>`**
  - Default: `10`
  - Meaning: Enable profiling of mutex lock.

- **`--silent-warning`**
  - Default: `false`
  - Meaning: Suppress warning messages.

- **`--version`**
  - Default: `false`
  - Meaning: Print version information and exit.

- **`--help|-h|-?`**
  - Default: `false`
  - Meaning: Print help information.

---

# EXAMPLES

**Export sequences with identifiers to CSV**

Extracts sequence IDs and sequences from a FASTQ file.
```bash
obicsv --ids --sequence sequences.fastq -o output1.csv
```

**Expected output:** 3 sequences written to `output1.csv`.

**Export sequences with quality scores**

Useful for quality control and filtering in downstream tools.
```bash
obicsv --ids --sequence --quality sequences.fastq -o output2.csv
```

**Expected output:** 3 sequences written to `output2.csv`.

**Export with taxonomic information**

Includes taxid and scientific name for taxonomic analysis.
```bash
obicsv --ids --sequence --taxon --taxonomy /path/to/taxonomy sequences.fasta -o output.csv
```

**Auto-detect annotation columns from sequence headers**

Automatically discovers all annotation attributes present in the sequence headers and outputs them as CSV columns. Combined with `--ids` to also include the sequence identifier.
```bash
obicsv --auto --ids sequences.fasta -o output4.csv
```

**Expected output:** 3 rows in `output4.csv` with columns `id`, `sample`, `taxid` (attributes found in sequence headers).

**Extract specific attributes**

Keeps only the specified attributes as columns. Attributes not present in a sequence are written as the NA value.
```bash
obicsv --keep sample --keep taxid sequences.fasta -o output5.csv
```

**Expected output:** 3 rows in `output5.csv` with columns `taxid`, `sample`.

**Export with compression**

Writes gzip-compressed CSV output for large datasets.
```bash
obicsv --ids --sequence -Z sequences.fasta -o output6.csv.gz
```

**Expected output:** 3 sequences written to `output6.csv.gz`.

---

# SEE ALSO

- `obiconvert` — input/output handling framework
- `obipairing` — pairing information (used with `--obipairing`)
- Other export commands: `obifasta`, `obifastq`, `obijson`

---

# NOTES

- Without any column selection flag (`--ids`, `--sequence`, `--quality`, `--taxon`, `--auto`, `--keep`), the output contains no columns and no data.
- The `--taxon` option requires a valid taxonomy database specified with `--taxonomy`.
- Output is written to stdout by default; use `--out` to write to a file.
- Missing attributes are written as the NA value (customizable with `--na-value`).
- Input sequences are processed using streaming iterators to minimize memory footprint, even for large files.