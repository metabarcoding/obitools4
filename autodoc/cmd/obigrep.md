# obigrep(1) — OBITools4 Manual

## NAME

`obigrep` — select a subset of sequence records on various criteria

## SYNOPSIS

```
obigrep [OPTIONS] [FILE...]
```

## DESCRIPTION

`obigrep` filters a set of biological sequence records (in FASTA or FASTQ format) and writes only those matching all specified criteria to the output. Its name is modelled on the Unix `grep` command, but instead of filtering lines in a text file, it filters sequence records.

Filtering criteria can be combined freely: only sequence records satisfying **all** specified conditions are retained. The selection can be inverted with `--inverse-match` to keep the records that would otherwise be discarded.

Sequences are read from one or more files, or from standard input if no file is given. Results are written to standard output or to a file specified with `--out`. Records that do not pass the filters can optionally be saved to a separate file with `--save-discarded`.

## INPUT FORMATS

`obigrep` recognises the following input formats automatically. A specific format can be forced with the corresponding flag:

| Flag | Format |
|------|--------|
| `--fasta` | FASTA |
| `--fastq` | FASTQ |
| `--embl` | EMBL flat file |
| `--genbank` | GenBank flat file |
| `--ecopcr` | ecoPCR output |
| `--csv` | CSV tabular format |

Header annotation styles can be selected with `--input-OBI-header` (OBITools format) or `--input-json-header` (JSON format).

## OUTPUT FORMATS

By default, the output format matches the input format (FASTQ when quality scores are present, FASTA otherwise). The format can be forced:

- `--fasta-output` — write FASTA
- `--fastq-output` — write FASTQ
- `--json-output` — write JSON
- `--output-OBI-header` / `-O` — annotate FASTA/FASTQ title lines in OBITools format
- `--output-json-header` — annotate FASTA/FASTQ title lines in JSON format
- `--compress` / `-Z` — compress output with gzip

Use `--out FILE` / `-o FILE` to write results to a file instead of standard output.

## FILTERING OPTIONS

### By sequence length

- `--min-length LENGTH` / `-l LENGTH`
  Keep only sequences at least *LENGTH* bases long.

- `--max-length LENGTH` / `-L LENGTH`
  Keep only sequences at most *LENGTH* bases long.

### By read abundance

Sequence records can carry a `count` attribute recording how many times the sequence was observed. The following options filter on that count:

- `--min-count COUNT` / `-c COUNT`
  Keep only sequences observed at least *COUNT* times (default: 1).

- `--max-count COUNT` / `-C COUNT`
  Keep only sequences observed at most *COUNT* times.

### By sequence pattern

- `--sequence PATTERN` / `-s PATTERN`
  Keep records whose nucleotide sequence matches the regular expression *PATTERN* (case-insensitive). This option can be repeated; all patterns must match.

- `--approx-pattern PATTERN`
  Keep records whose sequence contains an approximate match to *PATTERN*. The number of allowed differences is controlled by `--pattern-error`. This option can be repeated.

- `--pattern-error N`
  Maximum number of mismatches (or indels, if `--allows-indels` is set) tolerated when using `--approx-pattern` (default: 0, i.e. exact match).

- `--allows-indels`
  Allow insertions and deletions (in addition to substitutions) when performing approximate pattern matching.

- `--only-forward`
  Search patterns on the forward strand only. By default both strands are searched.

### By identifier or definition

- `--identifier PATTERN` / `-I PATTERN`
  Keep records whose identifier matches the regular expression *PATTERN* (case-insensitive). Can be repeated.

- `--id-list FILENAME`
  Keep only records whose identifier appears in *FILENAME*, a plain-text file with one identifier per line.

- `--definition PATTERN` / `-D PATTERN`
  Keep records whose definition line matches the regular expression *PATTERN* (case-insensitive). Can be repeated.

### By attribute (metadata)

Sequence records can carry arbitrary key/value annotations:

- `--has-attribute KEY` / `-A KEY`
  Keep records that possess an attribute named *KEY*, regardless of its value. Can be repeated.

- `--attribute KEY=PATTERN` / `-a KEY=PATTERN`
  Keep records for which the value of attribute *KEY* matches the regular expression *PATTERN* (case-sensitive). Can be repeated; all constraints must be satisfied.

### By custom boolean expression

- `--predicate EXPRESSION` / `-p EXPRESSION`
  Keep records for which the boolean expression *EXPRESSION* evaluates to true. Attributes are accessed via the `annotations` map (e.g. `annotations["count"]`). The special variable `sequence` refers to the sequence object; its length can be obtained with `len(sequence)`. Can be repeated; all expressions must be true.

  Example: `-p 'annotations["count"] >= 10 && len(sequence) < 200'`

### By taxonomy

Taxonomy-based filtering requires a taxonomy database to be provided with `--taxonomy`.

- `--taxonomy PATH` / `-t PATH`
  Path to the taxonomy database.

- `--restrict-to-taxon TAXID` / `-r TAXID`
  Keep only records whose taxon belongs to the lineage of *TAXID* (i.e. is *TAXID* itself or a descendant). Can be repeated; sequences must satisfy at least one of the provided taxids.

- `--ignore-taxon TAXID` / `-i TAXID`
  Discard records whose taxon belongs to the lineage of *TAXID*. Can be repeated.

- `--valid-taxid`
  Keep only records that carry a valid, recognised taxonomic identifier.

- `--require-rank RANK_NAME`
  Keep only records whose taxon has a defined ancestor at the given rank (e.g. *species*, *genus*, *family*). Can be repeated.

- `--update-taxid`
  Automatically update merged taxids to their current valid equivalent.

- `--fail-on-taxonomy`
  Exit with an error if a taxid referenced in the data is not valid.

- `--with-leaves`
  When the taxonomy is extracted from a sequence file, attach each sequence as a leaf node under its annotated taxid.

- `--raw-taxid`
  Print taxids in output files without supplementary information (taxon name and rank).

### Inversion

- `--inverse-match` / `-v`
  Invert the selection: output the records that would otherwise be discarded.

## PAIRED-END OPTIONS

When paired-end sequencing data are provided (forward and reverse reads stored in two files), `obigrep` can apply filters taking both reads into account.

- `--paired-with FILENAME`
  File containing the reverse (paired) reads.

- `--paired-mode MODE`
  How to combine the filter result from the forward and reverse reads. *MODE* is one of:

  | Mode | Meaning |
  |------|---------|
  | `forward` | Keep the pair if the **forward** read passes (default) |
  | `reverse` | Keep the pair if the **reverse** read passes |
  | `and` | Keep the pair if **both** reads pass |
  | `or` | Keep the pair if **at least one** read passes |
  | `andnot` | Keep the pair if the **forward** passes and the **reverse** does not |
  | `xor` | Keep the pair if **exactly one** read passes |

## OUTPUT CONTROL

- `--save-discarded FILENAME`
  Write sequence records that do **not** pass the filters to *FILENAME*.

- `--out FILENAME` / `-o FILENAME`
  Write the selected records to *FILENAME* (default: standard output).

- `--skip-empty`
  Suppress sequences of length zero from the output.

## PERFORMANCE OPTIONS

- `--max-cpu N`
  Number of parallel processing threads (default: number of available CPUs).

- `--batch-size N`
  Minimum number of sequences per processing batch (default: 1).

- `--batch-size-max N`
  Maximum number of sequences per processing batch (default: 2000).

- `--batch-mem SIZE`
  Maximum memory per batch (e.g. `128M`, `1G`). Overrides `--batch-size-max` when memory is the limiting factor. Can also be set via the environment variable `OBIBATCHMEM`.

- `--no-order`
  When multiple input files are provided, indicates that no ordering is assumed between them, which can improve throughput.

- `--no-progressbar`
  Disable the progress bar.

## MISCELLANEOUS OPTIONS

- `--u-to-t`
  Convert uracil (U) to thymine (T) in all sequences (useful for RNA data).

- `--solexa`
  Decode quality scores according to the legacy Solexa specification instead of the standard Phred encoding.

- `--silent-warning`
  Suppress warning messages.

- `--debug`
  Enable verbose debug logging.

- `--version`
  Print version information and exit.

- `--help` / `-h` / `-?`
  Display the help message and exit.

## EXAMPLES

Keep all sequences longer than 100 bases:

```bash
obigrep --min-length 100 input.fasta > out_min_length.fasta
```

**Expected output:** 6 sequences written to `out_min_length.fasta`.

Select sequences observed at least 10 times:

```bash
obigrep --min-count 10 input.fasta > out_min_count.fasta
```

**Expected output:** 4 sequences written to `out_min_count.fasta`.

Keep sequences whose identifier starts with `BOLD`:

```bash
obigrep --identifier '^BOLD' input.fasta > out_bold.fasta
```

**Expected output:** 2 sequences written to `out_bold.fasta`.

Select only sequences carrying the IUPAC primer motif `GGGCWATGTTTCATAAYGGG` with up to 2 mismatches:

```bash
obigrep --approx-pattern GGGCWATGTTTCATAAYGGG --pattern-error 2 input.fasta > out_primer.fasta
```

**Expected output:** 2 sequences written to `out_primer.fasta`.

Retain sequences belonging to the genus *Homo* (taxid 9605) in an NCBI taxonomy:

```bash
obigrep --taxonomy /data/ncbi_tax --restrict-to-taxon 9605 input.fasta
```

Keep sequences that have a `sample` attribute equal to `lake1` and save the rest to a separate file:

```bash
obigrep --attribute sample='^lake1$' --save-discarded discarded.fasta \
        input.fasta > lake1.fasta
```

**Expected output:** 5 sequences written to `lake1.fasta`, 5 sequences written to `discarded.fasta`.

Invert a length filter (discard sequences shorter than 50 bases):

```bash
obigrep --min-length 50 --inverse-match input.fasta > out_short.fasta
```

**Expected output:** 1 sequence written to `out_short.fasta`.

Apply a custom predicate (sequences with count ≥ 5):

```bash
obigrep -p 'annotations["count"] >= 5' input.fasta > out_predicate.fasta
```

**Expected output:** 6 sequences written to `out_predicate.fasta`.

## OUTPUT

### Attribute table

Attributes present on sequence records are preserved unchanged in the output. No new attributes are added by `obigrep` itself — only filtering occurs.

| Attribute | Type | Description |
|-----------|------|-------------|
| `count` | integer | Number of times the sequence was observed (read from input) |
| `sample` | string | Sample identifier (read from input) |

Any other annotations present in the input are carried through to the output unmodified.

### Observed output example

```
>seq001 {"count":15,"sample":"lake1"}
acgtacgtacgtacgtacgtgggcaatgtttcataatgggacgtacgtacgtacgtacgt
acgtacgtacgtacgtacgtacgtacgtacgtacgtacgtacgtacgtacgtacgtacgt
acgtacgtacgtacgtacgtacgtacgtacgt
>seq002 {"count":3,"sample":"lake1"}
tgcatgcatgcatgcatgcatgcatgcatgcatgcatgcatgcatgcatgcatgcatgca
tgcatgcatgcatgcatgcatgcatgcatgcatgcatgcatgcatgcatgcatgcatgca
>seq004 {"count":2,"sample":"lake1"}
aaacccgggtttagctagctagctagctagctagctagctagctagctagctagctagct
agctagctagctagctagctagctagctagctagctagctagctagctagctagctagct
atacgtatcgatcg
>BOLD_005 {"count":8,"sample":"pond1"}
cgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcgat
cgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcg
>seq008 {"count":7,"sample":"river2"}
ttacgatcgatcgatcgatcgggcaatgtttcataaggggacgatcgatcgatcgatcga
tcgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcgat
```

## SEE ALSO

`obiannotate`(1), `obiuniq`(1), `obiconvert`(1), `obitag`(1), `obisplit`(1)

## OBITools4

`obigrep` is part of the **OBITools4** suite for analysing DNA metabarcoding and environmental DNA data.
