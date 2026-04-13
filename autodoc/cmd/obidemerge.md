# obidemerge

## NAME

`obidemerge` — split merged sequence records back into individual, sample-annotated copies

## SYNOPSIS

```
obidemerge [options] [input_files...]
```

## DESCRIPTION

In a typical metabarcoding workflow, `obiuniq` or similar tools collapse identical sequences
from multiple samples into a single representative record. That record carries a statistics
attribute (for example `merged_sample`) that stores, for every original sample, how many
times the sequence was observed. This compact representation is convenient for clustering
and denoising, but some downstream analyses need the original, per-sample view.

`obidemerge` reverses that merging step. For each input sequence, it reads the statistics
stored under a chosen attribute (by default `sample`) and produces one output sequence per
entry in that statistics map. Each output sequence is a copy of the original, but:

- its `sample` attribute (or whichever slot you chose) is set to the name of the individual
  sample,
- its read count is set to the abundance recorded for that sample.

The original statistics attribute is removed from all output sequences.

Sequences that carry no statistics for the chosen slot are passed through unchanged.

The command reads sequences from one or more files, or from standard input when no file is
given, and writes the results to standard output or to the file specified with `--out`.

## INPUT FORMATS

`obidemerge` accepts all sequence formats supported by OBITools4:

| Format | Description |
|--------|-------------|
| FASTA | Plain nucleotide sequences with annotation in the title line |
| FASTQ | Sequences with per-base quality scores |
| EMBL | European Nucleotide Archive flat-file format |
| GenBank | NCBI GenBank flat-file format |
| ecoPCR | Output produced by the ecoPCR tool |
| CSV | Comma-separated values with sequence and metadata columns |

The format is detected automatically from the file extension or content. You can override
detection with the format flags listed under **Input format options** below.

Annotations embedded in FASTA/FASTQ title lines can follow the OBI key=value style
(`--input-OBI-header`) or JSON style (`--input-json-header`).

## OUTPUT FORMATS

By default, the output format mirrors the input:

- If the input contains quality scores, output is FASTQ.
- Otherwise, output is FASTA with OBI-style annotations.

You can force a specific format with `--fasta-output`, `--fastq-output`, or `--json-output`.

## OPTIONS

### Demerge option

`--demerge <slot>`, `-d <slot>`
: Name of the sequence attribute that holds the per-sample statistics to expand.
  Each key in that statistics map becomes a separate output sequence.
  **Default:** `sample`

### Output options

`--out <FILENAME>`, `-o <FILENAME>`
: Write output to this file instead of standard output. Use `-` for standard output.
  **Default:** `-` (standard output)

`--fasta-output`
: Write output in FASTA format, even when quality scores are available.
  **Default:** false

`--fastq-output`
: Write output in FASTQ format (requires quality scores in the input).
  **Default:** false

`--json-output`
: Write output in JSON format, one record per line.
  **Default:** false

`--output-OBI-header`, `-O`
: Write FASTA/FASTQ title lines in OBI key=value annotation style.
  **Default:** false (JSON-style headers)

`--output-json-header`
: Write FASTA/FASTQ title lines in JSON annotation style.
  **Default:** false

`--compress`, `-Z`
: Compress the output with gzip.
  **Default:** false

`--skip-empty`
: Discard sequences of length zero from the output.
  **Default:** false

### Input format options

`--fasta`
: Force reading in FASTA format.

`--fastq`
: Force reading in FASTQ format.

`--embl`
: Force reading in EMBL flat-file format.

`--genbank`
: Force reading in GenBank flat-file format.

`--ecopcr`
: Force reading in ecoPCR output format.

`--csv`
: Force reading in CSV format.

`--input-OBI-header`
: Parse FASTA/FASTQ title lines as OBI-style key=value annotations.

`--input-json-header`
: Parse FASTA/FASTQ title lines as JSON annotations.

`--solexa`
: Decode quality scores using the Solexa/Illumina 1.0 convention instead of the standard
  Phred scale. Use this only for very old sequencing data.
  **Default:** false

`--u-to-t`
: Convert uracil (U) to thymine (T) in all sequences. Useful when working with RNA-derived
  data that should be treated as DNA.
  **Default:** false

`--no-order`
: When reading from several input files, do not attempt to preserve the order of records
  across files. May improve speed when order does not matter.
  **Default:** false

### Taxonomy options

`--taxonomy <path>`, `-t <path>`
: Path to the OBITools4 taxonomy database. Required only if taxonomic identifiers need to
  be resolved or validated during output.
  **Default:** none

`--fail-on-taxonomy`
: Stop with an error if a taxonomic identifier in the data is not found in the loaded
  taxonomy database.
  **Default:** false

`--raw-taxid`
: Print taxonomic identifiers as plain numbers, without appending the taxon name and rank.
  **Default:** false

`--update-taxid`
: Automatically replace deprecated taxonomic identifiers with their current equivalents,
  as declared in the taxonomy database.
  **Default:** false

`--with-leaves`
: When a taxonomy is extracted from the sequence file itself, treat each sequence as a
  leaf node under its annotated taxonomic identifier.
  **Default:** false

### Performance options

`--max-cpu <int>`
: Maximum number of parallel processing threads. Increase for faster processing on
  multi-core machines.
  **Default:** 16 (or the value of the `OBIMAXCPU` environment variable)

`--batch-size <int>`
: Minimum number of sequences processed together as a group.
  **Default:** 1

`--batch-size-max <int>`
: Maximum number of sequences processed together as a group.
  **Default:** 2000

`--batch-mem <size>`
: Maximum memory used per processing group (e.g. `64M`, `1G`). Set to `0` to disable the
  memory limit and rely on `--batch-size-max` alone.
  **Default:** `128M`

### Display options

`--no-progressbar`
: Hide the progress bar.
  **Default:** false

`--silent-warning`
: Suppress warning messages.
  **Default:** false

`--debug`
: Enable verbose debug logging.
  **Default:** false

`--version`
: Print the OBITools4 version and exit.

`--help`, `-h`, `-?`
: Print this help message and exit.

## EXAMPLES

### Example 1 — basic demerge using the default slot

After running `obiuniq`, the file `unique.fasta` contains merged sequences whose
`merged_sample` attribute records abundance per sample. Demerge back to one
sequence per sample:
<!-- corrected: -d sample (not -d merged_sample) because HasStatsOn("sample") looks for the merged_sample attribute -->

```bash
obidemerge -d sample unique.fasta > per_sample_merged.fasta
```

**Expected output:** 7 sequences written to `per_sample_merged.fasta`.

### Example 2 — demerge with the default `sample` slot

If the statistics are already stored under the attribute named `sample` (the default),
no `-d` flag is needed:

```bash
obidemerge unique.fasta > per_sample_default.fasta
```

**Expected output:** 7 sequences written to `per_sample_default.fasta`.

### Example 3 — write compressed output to a file

```bash
obidemerge -d sample -o per_sample.fasta.gz --compress unique.fasta
```

**Expected output:** 7 sequences written (compressed) to `per_sample.fasta.gz`.

### Example 4 — pipeline use: cluster, then demerge

Obtain unique sequences, cluster them, then expand the clusters back to individual
sample records for ecological analysis:

```bash
obiuniq -m sample reads.fastq \
  | obiclean ... \
  | obidemerge -d sample -o demerged.fasta
```

### Example 5 — process multiple input files

```bash
obidemerge -d sample run1_unique.fasta run2_unique.fasta > combined_demerged.fasta
```

**Expected output:** 6 sequences written to `combined_demerged.fasta`.

## SEE ALSO

`obiuniq(1)` — collapses identical sequences and records per-sample counts (the inverse operation)
`obiclean(1)` — removes PCR/sequencing artefacts from a set of unique sequences
`obiannotate(1)` — adds or modifies sequence attributes
`obigrep(1)` — filters sequences by attributes or sequence content
`obicount(1)` — counts sequences and total reads in a file

## NOTES

**Relationship to `obiuniq`.**
`obiuniq --merge sample` stores per-sample counts under an attribute named `merged_sample`.
When you later call `obidemerge`, you must therefore pass `-d sample` to match that
attribute name. The `-d` option takes the **logical** slot name (here `sample`), not the
internal storage name (`merged_sample`).
<!-- corrected: -d sample is correct (not -d merged_sample); the tool prepends "merged_" internally when looking up the attribute -->

**Read counts after demerging.**
Each output sequence has its read count set to the value recorded in the statistics map for
that sample. If you sum the counts of all output sequences that share the same identifier,
you recover the total count of the original merged record.

**Order of output sequences.**
The order in which the per-sample copies of a single merged sequence appear in the output
is not guaranteed. If a stable order is required, pipe the output through `obisort`.

## OUTPUT

`obidemerge` writes one sequence record per sample entry found in the statistics attribute.
Each output record is a copy of the input sequence, with:

- the statistics attribute (`merged_<slot>`) removed,
- the `<slot>` attribute set to the sample name,
- the `count` attribute set to the abundance for that sample.

Sequences with no statistics for the chosen slot are passed through unchanged.

## Observed output example

```
>seq001 {"count":5,"sample":"sampleA"}
acgtacgtacgtacgtacgtacgtacgtacgtacgtacgt
>seq001 {"count":3,"sample":"sampleB"}
acgtacgtacgtacgtacgtacgtacgtacgtacgtacgt
>seq001 {"count":1,"sample":"sampleC"}
acgtacgtacgtacgtacgtacgtacgtacgtacgtacgt
>seq002 {"count":2,"sample":"sampleA"}
ttggccaattggccaattggccaattggccaattggccaa
>seq002 {"count":7,"sample":"sampleD"}
ttggccaattggccaattggccaattggccaattggccaa
>seq003 {"count":4,"sample":"sampleB"}
gctagctagctagctagctagctagctagctagctagcta
>seq004 {"count":6}
aaaaccccggggttttaaaaccccggggttttaaaacccc
```
