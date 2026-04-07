# NAME

obisummary — resume main information from a sequence file

---

# SYNOPSIS

```
obisummary [--batch-mem <string>] [--batch-size <int>]
           [--batch-size-max <int>] [--csv] [--debug] [--ecopcr] [--embl]
           [--fasta] [--fastq] [--genbank] [--help|-h|-?]
           [--input-OBI-header] [--input-json-header] [--json-output]
           [--map <string>]... [--max-cpu <int>] [--no-order] [--pprof]
           [--pprof-goroutine <int>] [--pprof-mutex <int>] [--silent-warning]
           [--solexa] [--u-to-t] [--version] [--yaml-output] [<args>]
```

---

# DESCRIPTION

`obisummary` reads a set of biological sequences and computes a statistical
summary of their content and annotations. Rather than producing a new sequence
file, it outputs a single structured record describing the dataset as a whole.

The summary covers three main areas. First, global counts: the total number of
reads (sequences weighted by their `count` attribute), the number of distinct
sequence variants, and the total sequence length across all records. Second,
annotation profiling: `obisummary` inspects every annotation key present in
the dataset and classifies it as a scalar attribute (single value per
sequence), a map attribute (key-to-count mapping), or a vector attribute
(multi-value per sequence). Third, per-sample statistics: when sequences carry
sample information (via `merged_sample` or equivalent per-sample annotations),
`obisummary` reports for each sample the number of reads, the number of
variants, and the number of singletons. If `obiclean` has been run previously,
the summary also captures `obiclean_status` and related quality flags per
sample.

The output is a single JSON record by default, or YAML when `--yaml-output` is
requested. <!-- corrected: actual default output is JSON, not YAML -->
`obisummary` is typically used after processing steps such as
`obiclean` or `obiuniq` to quickly validate the state of a dataset before
downstream analysis.

---

# INPUT

`obisummary` accepts biological sequence data from one or more files supplied
as positional arguments, or from standard input when no files are given.
Supported formats include FASTA, FASTQ, GenBank flatfile, EMBL flatfile,
ecoPCR output, and CSV. By default the format is detected automatically; use
the format flags (`--fasta`, `--fastq`, `--genbank`, `--embl`, `--ecopcr`,
`--csv`) to force a specific parser.

FASTA/FASTQ annotation headers may follow the OBI format (`--input-OBI-header`)
or JSON format (`--input-json-header`). RNA sequences can be read as DNA by
converting uracil to thymine with `--u-to-t`. Quality strings encoded according
to the Solexa specification are handled with `--solexa`.

When multiple input files are provided, `obisummary` assumes they are ordered;
use `--no-order` to indicate that no ordering exists among them.

---

# OUTPUT

`obisummary` writes a single structured record to standard output. The default
format is JSON; use `--yaml-output` to obtain YAML instead.
<!-- corrected: actual default output is JSON, not YAML -->

The record contains three top-level sections:

- **`count`**: global metrics including `variants` (distinct sequences),
  `reads` (total weighted count), and `total_length` (sum of all sequence
  lengths).

- **`annotations`**: a breakdown of all annotation keys found in the dataset,
  classified as `scalar_attributes`, `map_attributes`, or `vector_attributes`,
  together with the observed keys and their occurrence counts within each
  category.

- **`samples`**: when sample information is present, `sample_count` and a
  per-sample `sample_stats` table with `reads`, `variants`, and `singletons`
  fields. If `obiclean` data is present, an `obiclean_bad` field is also
  reported per sample.

When `--map` is used, the named map attribute is included in the annotation
detail for that attribute.

## Observed output example

```
{
  "annotations": {
    "keys": {
      "scalar": {
        "count": 5
      }
    },
    "map_attributes": 0,
    "scalar_attributes": 1,
    "vector_attributes": 0
  },
  "count": {
    "reads": 21,
    "total_length": 100,
    "variants": 5
  }
}
```

---

# OPTIONS

## Summary output

**`--json-output`**
- Default: `false`
- Print the result as a JSON record (this is the default behaviour; this flag
  makes the choice explicit).
<!-- corrected: JSON is the default output format, not YAML -->

**`--yaml-output`**
- Default: `false`
- Print the result as a YAML record instead of the default JSON format.
<!-- corrected: YAML is not the default; JSON is -->

**`--map <string>`**
- Default: `[]` (none)
- Name of a map attribute to include in the summary. This option may be
  repeated to request multiple map attributes. Each named attribute will be
  detailed in the `map_attributes` section of the output.

## Input format

**`--fasta`**
- Default: `false`
- Read data following the FASTA format.

**`--fastq`**
- Default: `false`
- Read data following the FASTQ format.

**`--genbank`**
- Default: `false`
- Read data following the GenBank flatfile format.

**`--embl`**
- Default: `false`
- Read data following the EMBL flatfile format.

**`--ecopcr`**
- Default: `false`
- Read data following the ecoPCR output format.

**`--csv`**
- Default: `false`
- Read data following the CSV format.

**`--input-OBI-header`**
- Default: `false`
- FASTA/FASTQ title line annotations follow OBI format.

**`--input-json-header`**
- Default: `false`
- FASTA/FASTQ title line annotations follow JSON format.

**`--solexa`**
- Default: `false`
- Decode quality strings according to the Solexa specification.

**`--u-to-t`**
- Default: `false`
- Convert uracil (U) to thymine (T) when reading RNA sequences.

## Batch control

**`--batch-size <int>`**
- Default: `1`
- Minimum number of sequences per processing batch.

**`--batch-size-max <int>`**
- Default: `2000`
- Maximum number of sequences per processing batch.

**`--batch-mem <string>`**
- Default: `""` (128M effective)
- Maximum memory per batch (e.g. `128K`, `64M`, `1G`). Set to `0` to disable
  the memory limit.

## Processing

**`--max-cpu <int>`**
- Default: `16`
- Number of parallel threads used to compute the result.

**`--no-order`**
- Default: `false`
- When several input files are provided, indicates that there is no order
  among them.

## General

**`--debug`**
- Default: `false`
- Enable debug mode by setting the log level to debug.

**`--silent-warning`**
- Default: `false`
- Stop printing warning messages.

**`--version`**
- Default: `false`
- Print the version and exit.

**`--help` / `-h` / `-?`**
- Default: `false`
- Display help and exit.

**`--pprof`**
- Default: `false`
- Enable the pprof profiling server. Consult the log for the server address.

**`--pprof-goroutine <int>`**
- Default: `6060`
- Port for goroutine blocking profile.

**`--pprof-mutex <int>`**
- Default: `10`
- Port for mutex lock profiling.

---

# EXAMPLES

```bash
# Get a JSON summary of a FASTA file produced by obiclean
obisummary cleaned.fasta > out_default.yaml
```

**Expected output:** a JSON summary record in `out_default.yaml`.

```bash
# Get the summary as an explicit JSON record for programmatic processing
obisummary --json-output cleaned.fasta > out_json.json
```

**Expected output:** a JSON summary record in `out_json.json`.

```bash
# Get a YAML record from a FASTQ file
obisummary --yaml-output --fastq reads.fastq > out_yaml.yaml
```

**Expected output:** a YAML summary record in `out_yaml.yaml`.

```bash
# Summarise data read from standard input, forcing FASTA format
obigrep -p 'annotations.count > 1' sequences.fasta | obisummary --fasta > out_pipeline.yaml
```

**Expected output:** a JSON summary record in `out_pipeline.yaml` (3 variants, 10 reads).

---

# SEE ALSO

`obiclean`, `obiuniq`, `obicount`
