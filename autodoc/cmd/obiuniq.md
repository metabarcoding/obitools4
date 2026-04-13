# NAME

obiuniq — dereplicate sequence data sets

---

# SYNOPSIS

```
obiuniq [--batch-mem <string>] [--batch-size <int>] [--batch-size-max <int>]
        [--category-attribute|-c <CATEGORY>]... [--chunk-count <int>]
        [--compress|-Z] [--csv] [--debug] [--ecopcr] [--embl]
        [--fail-on-taxonomy] [--fasta] [--fasta-output] [--fastq]
        [--fastq-output] [--genbank] [--help|-h|-?] [--in-memory]
        [--input-OBI-header] [--input-json-header] [--json-output]
        [--max-cpu <int>] [--merge|-m <KEY>]... [--na-value <NA_NAME>]
        [--no-order] [--no-progressbar] [--no-singleton]
        [--out|-o <FILENAME>] [--output-OBI-header|-O] [--output-json-header]
        [--pprof] [--pprof-goroutine <int>] [--pprof-mutex <int>]
        [--raw-taxid] [--silent-warning] [--skip-empty] [--solexa]
        [--taxonomy|-t <string>] [--u-to-t] [--update-taxid] [--version]
        [--with-leaves] [<args>]
```

---

# DESCRIPTION

`obiuniq` groups identical sequences together and replaces them with a single
representative, recording the total number of original occurrences as an
abundance count. This process — called dereplication — is a standard step in
amplicon sequencing workflows: it dramatically reduces the number of sequence
records to process, while preserving exact counts needed for downstream
statistical analyses.

By default, two sequences are considered identical if and only if their
nucleotide strings are the same. Using `--category-attribute` (repeatable),
additional metadata fields can be included in the identity criterion. For
example, grouping by sample name keeps the same sequence as separate records
when it occurs in different samples, enabling per-sample abundance tracking.

For each group of identical sequences, `obiuniq` emits one output record
carrying the merged metadata of all members. The `--merge` option (repeatable)
instructs the command to also record, in an attribute named `merged_<KEY>`, the
distribution of `KEY` attribute values across the sequences collapsed into each
group — useful for provenance tracking and quality control. <!-- corrected: actual attribute name is merged_KEY (not KEY); tracks attribute value distributions, not a list of sequence IDs -->

Sequences that appear only once in the entire dataset (singletons) can be
removed with `--no-singleton`. Singletons often represent sequencing errors
rather than genuine biological variants, so their removal is a common
noise-reduction step.

---

# INPUT

`obiuniq` accepts biological sequence data in FASTA, FASTQ, EMBL, GenBank,
ecoPCR, or CSV format (auto-detected by default, or forced with format flags
such as `--fasta`, `--fastq`, `--embl`, etc.). Input is read from one or more
files given as positional arguments, or from standard input when no files are
provided.

When multiple input files are provided, `obiuniq` assumes they are ordered
(e.g., paired-end reads in the same read order). If no such ordering exists,
use `--no-order` to signal that files can be consumed independently.

FASTA/FASTQ header annotations are parsed heuristically by default. Use
`--input-OBI-header` for OBI-formatted headers or `--input-json-header` for
JSON-formatted headers. RNA sequences can be normalised to DNA on the fly with
`--u-to-t`.

---

# OUTPUT

`obiuniq` writes dereplicated sequences to standard output or to the file
specified by `--out`. Each output record represents one group of identical
sequences (identical under the chosen grouping criterion). The output carries
the merged metadata from all input records in the group.

The output format defaults to FASTA. Even when the input contains quality
scores (FASTQ), quality information is not preserved across merged sequences,
so the output is written in FASTA format unless `--fastq-output` is explicitly
requested. <!-- corrected: actual output is always FASTA when dereplicating; quality scores are dropped during merging -->
Output annotations follow the OBI header format when `--output-OBI-header` is
set, or JSON when `--output-json-header` is set. The output can be
gzip-compressed with `--compress`.

For each output record:
- The abundance count reflects how many input sequences were merged into the
  group.
- Attributes created by `--merge KEY` are named `merged_KEY` and map each
  observed value of the `KEY` attribute to the count of input sequences
  carrying that value within the group. <!-- corrected: attribute name is merged_KEY; value is a map not a list -->
- All other attributes are merged from the contributing records according to
  the standard OBITools4 merging rules.

## Observed output example

```
>seq008 {"count":1,"primer":"p1"}
cccccccccccccccccccc
>seq001 {"count":4,"primer":"p1"}
atcgatcgatcgatcgatcg
>seq004 {"count":2,"primer":"p1","sample":"s1"}
gctagctagctagctagcta
>seq007 {"count":1,"primer":"p1","sample":"s2"}
tttttttttttttttttttt
```

---

# OPTIONS

## Dereplication Options

**`--category-attribute|-c <CATEGORY>`** (default: `[]`)  
Adds one metadata attribute to the grouping criterion. Two sequences are
placed in the same group only when they are nucleotide-identical **and** share
the same value for every attribute listed with `-c`. This option can be
repeated to combine multiple attributes (e.g., `-c sample -c primer`).
Records that lack a listed attribute receive the value set by `--na-value`.

**`--chunk-count <int>`** (default: `100`)  
Controls how many internal partitions the dataset is split into during
processing. A higher value reduces per-partition memory usage at the cost of
more temporary files; a lower value increases per-partition memory but reduces
I/O overhead. Tune this when processing very large or very small datasets.

**`--in-memory`** (default: `false`)  
Stores intermediate data chunks in RAM rather than in temporary disk files.
Speeds up processing on datasets that fit comfortably in available memory;
omit this flag (the default) for large datasets that exceed available RAM.

**`--merge|-m <KEY>`** (default: `[]`)  
Creates an output attribute named `merged_KEY` that maps each observed value
of the `KEY` attribute to the count of input sequences carrying that value
within the group. Repeat to track multiple attributes. <!-- corrected: actual attribute name is merged_KEY (not KEY); value is a map of attribute values to counts, not a list of sequence IDs -->
Useful for tracking which sample or category contributions were collapsed into each group.

**`--na-value <NA_NAME>`** (default: `"NA"`)  
Value assigned to a category attribute when a sequence record does not carry
that attribute. All sequences lacking the attribute are grouped together under
this placeholder, rather than being treated as incomparable.

**`--no-singleton`** (default: `false`)  
Discards all output records whose abundance count is exactly one — i.e.,
sequences that occur only once across the entire input. Removing singletons
is a standard heuristic for excluding sequencing errors from further analysis.

## Input Options

**`--batch-mem <string>`** (default: `""`, env: `OBIBATCHMEM`)  
Maximum memory budget per processing batch (e.g. `128K`, `64M`, `1G`). Set
to `0` to disable the memory ceiling. Overrides `--batch-size-max` when
both are set.

**`--batch-size <int>`** (default: `10`, env: `OBIBATCHSIZE`)  
Minimum number of sequences per batch (floor).

**`--batch-size-max <int>`** (default: `2000`, env: `OBIBATCHSIZEMAX`)  
Maximum number of sequences per batch (ceiling).

**`--csv`** (default: `false`)  
Parse input as CSV format.

**`--ecopcr`** (default: `false`)  
Parse input as ecoPCR output format.

**`--embl`** (default: `false`)  
Parse input as EMBL flatfile format.

**`--fasta`** (default: `false`)  
Parse input as FASTA format.

**`--fastq`** (default: `false`)  
Parse input as FASTQ format.

**`--genbank`** (default: `false`)  
Parse input as GenBank flatfile format.

**`--input-OBI-header`** (default: `false`)  
Treat FASTA/FASTQ title line annotations as OBI-format key=value pairs.

**`--input-json-header`** (default: `false`)  
Treat FASTA/FASTQ title line annotations as JSON objects.

**`--no-order`** (default: `false`)  
When multiple input files are provided, indicates that there is no ordering
relationship among them.

**`--skip-empty`** (default: `false`)  
Suppress sequences of length zero from the output.

**`--solexa`** (default: `false`, env: `OBISOLEXA`)  
Decode quality strings according to the Solexa specification rather than the
standard Phred encoding.

**`--u-to-t`** (default: `false`)  
Convert uracil (U) to thymine (T) in all input sequences, normalising RNA to
DNA representation.

## Output Options

**`--compress|-Z`** (default: `false`)  
Compress output using gzip.

**`--fasta-output`** (default: `false`)  
Write output in FASTA format (default when no quality scores are available).

**`--fastq-output`** (default: `false`)  
Write output in FASTQ format (default when quality scores are present).

**`--json-output`** (default: `false`)  
Write output in JSON format.

**`--out|-o <FILENAME>`** (default: `"-"`)  
Write output to the specified file instead of standard output.

**`--output-OBI-header|-O`** (default: `false`)  
Write FASTA/FASTQ title line annotations in OBI format.

**`--output-json-header`** (default: `false`)  
Write FASTA/FASTQ title line annotations in JSON format.

## Taxonomy Options

**`--fail-on-taxonomy`** (default: `false`)  
Cause `obiuniq` to exit with an error if a taxid in the data is not a
currently valid taxon in the loaded taxonomy.

**`--raw-taxid`** (default: `false`)  
Print taxids in output without supplementary information (taxon name and rank).

**`--taxonomy|-t <string>`** (default: `""`)  
Path to the taxonomy database used to validate or update taxids.

**`--update-taxid`** (default: `false`)  
Automatically replace merged taxids with the most recent valid taxid.

**`--with-leaves`** (default: `false`)  
When taxonomy is extracted from a sequence file, add sequences as leaves of
their taxid annotation.

## Execution Options

**`--max-cpu <int>`** (default: `16`, env: `OBIMAXCPU`)  
Number of parallel threads used to compute the result.

**`--debug`** (default: `false`, env: `OBIDEBUG`)  
Enable debug mode by setting the log level to debug.

**`--no-progressbar`** (default: `false`)  
Disable the progress bar.

**`--silent-warning`** (default: `false`, env: `OBIWARNING`)  
Suppress warning messages.

**`--pprof`** (default: `false`)  
Enable the pprof profiling server (address logged at startup).

**`--pprof-goroutine <int>`** (default: `6060`, env: `OBIPPROFGOROUTINE`)  
Port for the goroutine blocking profile endpoint.

**`--pprof-mutex <int>`** (default: `10`, env: `OBIPPROFMUTEX`)  
Rate for the mutex contention profile.

**`--version`** (default: `false`)  
Print the version string and exit.

**`--help|-h|-?`** (default: `false`)  
Print usage information and exit.

---

# EXAMPLES

```bash
# Dereplicate a FASTQ file of amplicon reads; write unique sequences to a FASTA output file.
obiuniq reads.fastq -o out_basic.fastq
```

**Expected output:** 4 sequences written to `out_basic.fastq`.

```bash
# Dereplicate keeping sequences separate per sample (category attribute),
# and discard singletons to remove likely sequencing errors.
obiuniq -c sample --no-singleton reads.fastq -o out_no_singleton.fastq
```

**Expected output:** 2 sequences written to `out_no_singleton.fastq`.

```bash
# Dereplicate per sample, recording the sample distribution in 'merged_sample',
# and use 'UNKNOWN' for reads missing the sample attribute.
obiuniq -c sample --merge sample --na-value UNKNOWN reads.fastq -o out_merge.fastq
```

**Expected output:** 5 sequences written to `out_merge.fastq`.

```bash
# Process a dataset entirely in memory using 200 internal partitions,
# writing gzip-compressed output.
obiuniq --in-memory --chunk-count 200 --compress -o out_inmemory.fastq.gz reads.fastq
```

**Expected output:** 4 sequences written to `out_inmemory.fastq.gz`.

```bash
# Dereplicate reads from two sample files with no assumed ordering between them,
# grouping by both sample and primer attributes.
obiuniq --no-order -c sample -c primer sample1.fastq sample2.fastq -o out_multifile.fastq
```

**Expected output:** 4 sequences written to `out_multifile.fastq`.

---

# SEE ALSO

- `obigrep` — filter dereplicated sequences by abundance, length, or annotation
- `obiannotate` — add or modify annotations on dereplicated records
- `obicount` — count sequences or groups in a dataset
- `obiclean` — remove sequencing artefacts from a dereplicated dataset
- `obisummary` — summarise annotation distributions across a sequence set

---

# NOTES

For datasets that do not fit in RAM, `obiuniq` uses temporary disk-backed
chunk files by default. The number of chunks is controlled by `--chunk-count`
(default 100). Increasing this value lowers per-chunk memory requirements;
decreasing it reduces I/O at the cost of higher peak memory. Use `--in-memory`
only when the full working set fits in available RAM, as exceeding memory will
degrade performance or cause out-of-memory failures.

Singletons (sequences with abundance = 1) are a common source of noise in
amplicon sequencing, often arising from PCR or sequencing errors. The
`--no-singleton` flag is therefore recommended for most metabarcoding
workflows, unless the study design requires retaining all observed variants.

When the `--category-attribute` option is used, records that lack the
specified attribute are grouped together under the `--na-value` placeholder
(default `"NA"`). This ensures that all records participate in dereplication
without being silently dropped, but users should be aware that heterogeneous
records with different missing attributes may be unintentionally merged.
