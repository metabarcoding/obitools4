# NAME

obiscript — executes a lua script on the input sequences

---

# SYNOPSIS

```
obiscript [--allows-indels] [--approx-pattern <PATTERN>]...
          [--attribute|-a <KEY=VALUE>]... [--batch-mem <string>]
          [--batch-size <int>] [--batch-size-max <int>] [--compress|-Z]
          [--csv] [--debug] [--definition|-D <PATTERN>]... [--ecopcr]
          [--embl] [--fail-on-taxonomy] [--fasta] [--fasta-output] [--fastq]
          [--fastq-output] [--genbank] [--has-attribute|-A <KEY>]...
          [--help|-h|-?] [--id-list <FILENAME>]
          [--identifier|-I <PATTERN>]... [--ignore-taxon|-i <TAXID>]...
          [--input-OBI-header] [--input-json-header] [--inverse-match|-v]
          [--json-output] [--max-count|-C <COUNT>] [--max-cpu <int>]
          [--max-length|-L <LENGTH>] [--min-count|-c <COUNT>]
          [--min-length|-l <LENGTH>] [--no-order] [--no-progressbar]
          [--only-forward] [--out|-o <FILENAME>] [--output-OBI-header|-O]
          [--output-json-header]
          [--paired-mode <forward|reverse|and|or|andnot|xor>]
          [--pattern-error <int>] [--pprof] [--pprof-goroutine <int>]
          [--pprof-mutex <int>] [--predicate|-p <EXPRESSION>]...
          [--raw-taxid] [--require-rank <RANK_NAME>]...
          [--restrict-to-taxon|-r <TAXID>]... [--script|-S <string>]
          [--sequence|-s <PATTERN>]... [--silent-warning] [--skip-empty]
          [--solexa] [--taxonomy|-t <string>] [--template] [--u-to-t]
          [--update-taxid] [--valid-taxid] [--version] [--with-leaves]
          [<args>]
```

---

# DESCRIPTION

`obiscript` applies a user-provided Lua script to a stream of biological sequences. For each input sequence record, the script's `worker(sequence)` function is called, giving the user full programmatic access to the sequence's identifier, data, quality scores, and metadata attributes. This makes it possible to implement custom annotation logic, computed filters, or record transformations that go beyond what fixed-function OBITools commands offer.

The Lua script may also define two optional lifecycle hooks: `begin()`, called once before any sequence is processed (useful for initialising counters or opening files), and `finish()`, called after the last sequence (useful for printing summary statistics or flushing output). A thread-safe shared table `obicontext` is available across all workers, allowing aggregation across parallel executions.

Sequences are read from files or standard input in any format supported by OBITools4 (FASTA, FASTQ, EMBL, GenBank, ecoPCR, CSV). The sequence filtering flags (such as `--min-length`, `--predicate`, etc.) select which sequences the Lua script is applied to; sequences that do not match the filter pass through to the output unchanged without the script being executed on them. All sequences — scripted or not — are written to the output. <!-- corrected: non-matching sequences are passed through unchanged, not discarded -->

To get started, use `--template` to print a minimal Lua script skeleton with stubs for all three hooks and inline documentation.

---

# INPUT

`obiscript` reads biological sequences from one or more files supplied as positional arguments, or from standard input if no files are given. All formats supported by OBITools4 are accepted: FASTA, FASTQ, EMBL flatfile, GenBank flatfile, ecoPCR output, and CSV. Format auto-detection is used by default; explicit format flags (`--fasta`, `--fastq`, `--embl`, `--genbank`, `--ecopcr`, `--csv`) override it. Header annotation style can be forced with `--input-OBI-header` or `--input-json-header`.

---

# OUTPUT

Sequences processed by the Lua script are written to standard output, or to the file given by `--out`. Any modifications made to sequence records inside `worker()` (identifier, sequence, attributes) are reflected in the output. The output format defaults to FASTA when no quality data are present and to FASTQ otherwise; use `--fasta-output`, `--fastq-output`, or `--json-output` to override. Header annotation style in FASTA/FASTQ output can be set with `--output-OBI-header` or `--output-json-header`. Output can be gzip-compressed with `--compress`.

## Observed output example

```
>sample1_seq001 {"definition":"control sequence for annotation test","sample":"sample1"}
atcgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcgatcg
>sample1_seq002 {"definition":"another control sequence from sample1","sample":"sample1"}
gctagctagctagctagctagctagctagctagctagctagctagcta
>sample2_seq003 {"definition":"second sample sequence","sample":"sample2"}
ttaattaattaattaattaattaattaattaattaattaattaattaa
>sample2_seq004 {"definition":"second sample another sequence","sample":"sample2"}
ccggccggccggccggccggccggccggccggccggccggccggccgg
>sample3_seq005 {"definition":"third sample first sequence","sample":"sample3"}
aaaattttccccggggaaaattttccccggggaaaattttccccgggg
>sample3_seq006 {"definition":"third sample second sequence","sample":"sample3"}
ttttaaaaccccggggttttaaaaccccggggttttaaaaccccgggg
```

---

# OPTIONS

## Script

### `--script|-S <string>`
- Default: `""`
- Path to the Lua script file to execute. The file must exist and be syntactically valid Lua. The script should define a `worker(sequence)` function, and optionally `begin()` and `finish()`.

### `--template`
- Default: `false`
- Print a minimal Lua script template to standard output, with stubs for `begin()`, `worker()`, and `finish()` and inline documentation, then exit. Use this to bootstrap a new script.

## Sequence filtering (selects sequences on which the script is applied; non-matching sequences pass through unchanged)

### `--predicate|-p <EXPRESSION>`
- Default: `[]`
- Boolean expression evaluated for each sequence record. Attribute keys are accessible as variable names; `sequence` refers to the record itself. Multiple `-p` options are combined with AND.

### `--sequence|-s <PATTERN>`
- Default: `[]`
- Regular expression matched against the nucleotide sequence. Case-insensitive. Multiple patterns are combined with AND.

### `--identifier|-I <PATTERN>`
- Default: `[]`
- Regular expression matched against the sequence identifier. Case-insensitive.

### `--definition|-D <PATTERN>`
- Default: `[]`
- Regular expression matched against the sequence definition line. Case-insensitive.

### `--approx-pattern <PATTERN>`
- Default: `[]`
- Pattern matched approximately against the sequence. Use `--pattern-error` to set the maximum number of errors allowed.

### `--pattern-error <int>`
- Default: `0`
- Maximum number of errors (mismatches) allowed during approximate pattern matching.

### `--allows-indels`
- Default: `false`
- Allow insertions and deletions (in addition to mismatches) during approximate pattern matching.

### `--only-forward`
- Default: `false`
- Restrict pattern matching to the forward strand only.

### `--has-attribute|-A <KEY>`
- Default: `[]`
- Apply the script only to records that have an attribute with key `<KEY>`; others pass through.

### `--attribute|-a <KEY=VALUE>`
- Default: `{}`
- Apply the script only to records where the attribute `KEY` matches the regular expression `VALUE`. Case-sensitive. Multiple `-a` options are combined with AND.

### `--id-list <FILENAME>`
- Default: `""`
- Path to a text file containing one sequence identifier per line. The script is applied only to records whose identifier appears in the file; others pass through.

### `--min-length|-l <LENGTH>`
- Default: `1`
- Apply the script only to sequences whose length is at least `LENGTH`; shorter sequences pass through unchanged.

### `--max-length|-L <LENGTH>`
- Default: `2000000000`
- Apply the script only to sequences whose length is at most `LENGTH`; longer sequences pass through unchanged.

### `--min-count|-c <COUNT>`
- Default: `1`
- Apply the script only to sequences with a count (abundance) of at least `COUNT`; others pass through unchanged.

### `--max-count|-C <COUNT>`
- Default: `2000000000`
- Apply the script only to sequences with a count (abundance) of at most `COUNT`; others pass through unchanged.

### `--inverse-match|-v`
- Default: `false`
- Invert the selection: apply the script to records that do NOT match the filter criteria; matching records pass through unchanged.

## Taxonomic filtering

### `--taxonomy|-t <string>`
- Default: `""`
- Path to the taxonomy database. Required for taxonomy-based options.

### `--restrict-to-taxon|-r <TAXID>`
- Default: `[]`
- Retain only sequences whose taxid belongs to the specified taxon.

### `--ignore-taxon|-i <TAXID>`
- Default: `[]`
- Exclude sequences whose taxid belongs to the specified taxon.

### `--require-rank <RANK_NAME>`
- Default: `[]`
- Retain only sequences whose taxon has the specified rank (e.g., `species`, `genus`).

### `--valid-taxid`
- Default: `false`
- Retain only sequences that carry a currently valid NCBI taxid.

### `--fail-on-taxonomy`
- Default: `false`
- Abort with an error if a taxid used during filtering is not currently valid.

### `--update-taxid`
- Default: `false`
- Automatically replace taxids declared as merged with their current equivalent.

### `--raw-taxid`
- Default: `false`
- Print taxids in output without supplementary information (taxon name and rank).

### `--with-leaves`
- Default: `false`
- When extracting taxonomy from a sequence file, attach sequences as leaves of their taxid annotation.

## Paired-end mode

### `--paired-mode <forward|reverse|and|or|andnot|xor>`
- Default: `"forward"`
- When paired reads are provided, determines how filter conditions are applied to both reads of a pair.

## Input format

### `--fasta`
- Default: `false`
- Force FASTA format parsing.

### `--fastq`
- Default: `false`
- Force FASTQ format parsing.

### `--embl`
- Default: `false`
- Force EMBL flatfile format parsing.

### `--genbank`
- Default: `false`
- Force GenBank flatfile format parsing.

### `--ecopcr`
- Default: `false`
- Force ecoPCR output format parsing.

### `--csv`
- Default: `false`
- Force CSV format parsing.

### `--input-OBI-header`
- Default: `false`
- Parse FASTA/FASTQ title line annotations as OBI format.

### `--input-json-header`
- Default: `false`
- Parse FASTA/FASTQ title line annotations as JSON format.

### `--solexa`
- Default: `false`
- Decode quality strings according to the Solexa specification.

### `--u-to-t`
- Default: `false`
- Convert uracil (U) to thymine (T) in sequences.

### `--skip-empty`
- Default: `false`
- Suppress sequences of length zero from the output.

### `--no-order`
- Default: `false`
- When multiple input files are provided, indicates that no ordering is assumed among them.

## Output format

### `--out|-o <FILENAME>`
- Default: `"-"` (standard output)
- File path for saving the output.

### `--fasta-output`
- Default: `false`
- Write output in FASTA format.

### `--fastq-output`
- Default: `false`
- Write output in FASTQ format.

### `--json-output`
- Default: `false`
- Write output in JSON format.

### `--output-OBI-header|-O`
- Default: `false`
- Write FASTA/FASTQ title line annotations in OBI format.

### `--output-json-header`
- Default: `false`
- Write FASTA/FASTQ title line annotations in JSON format.

### `--compress|-Z`
- Default: `false`
- Compress output using gzip.

## Performance

### `--max-cpu <int>`
- Default: `16` (env: `OBIMAXCPU`)
- Number of parallel threads used for processing.

### `--batch-size <int>`
- Default: `1` (env: `OBIBATCHSIZE`)
- Minimum number of sequences per processing batch.

### `--batch-size-max <int>`
- Default: `2000` (env: `OBIBATCHSIZEMAX`)
- Maximum number of sequences per processing batch.

### `--batch-mem <string>`
- Default: `""` → `128M` (env: `OBIBATCHMEM`)
- Maximum memory per batch (e.g. `128K`, `64M`, `1G`). Set to `0` to disable.

## Diagnostics

### `--debug`
- Default: `false` (env: `OBIDEBUG`)
- Enable debug logging.

### `--no-progressbar`
- Default: `false`
- Disable the progress bar.

### `--silent-warning`
- Default: `false` (env: `OBIWARNING`)
- Suppress warning messages.

### `--pprof`
- Default: `false`
- Enable the pprof profiling HTTP server (see log for address).

### `--pprof-goroutine <int>`
- Default: `6060` (env: `OBIPPROFGOROUTINE`)
- Port for goroutine blocking profile.

### `--pprof-mutex <int>`
- Default: `10` (env: `OBIPPROFMUTEX`)
- Rate for mutex lock profiling.

---

# EXAMPLES

```bash
# Print a starter script template and save it to my_script.lua
obiscript --template > my_script.lua
```

**Expected output:** Lua template with `begin()`, `worker()`, and `finish()` stubs written to `my_script.lua`.

```bash
# Add a custom annotation to every sequence record
# (the script sets a new attribute 'sample' from the identifier prefix)
obiscript --script annotate.lua --fasta-output sequences.fasta > annotated.fasta
```

**Expected output:** 6 sequences written to `annotated.fasta`.

```bash
# Count reads per taxon using the finish() hook, filtering to a specific taxon
obiscript --script count_taxa.lua \
          --restrict-to-taxon 6231 \
          --taxonomy /data/ncbi_tax \
          sequences.fasta > filtered_annotated.fasta
```

```bash
# Apply a script to FASTQ sequences with a length filter
obiscript --script process_pairs.lua \
          --min-length 100 \
          --out result.fastq \
          reads.fastq
```

**Expected output:** 4 sequences written to `result.fastq`.

```bash
# Run on FASTQ input, output JSON, using 4 CPU threads
obiscript --script enrich.lua \
          --json-output \
          --max-cpu 4 \
          sequences.fastq > enriched.json
```

**Expected output:** 4 sequences written to `enriched.json`.

---

# SEE ALSO

`obigrep` — filter sequences using the same selection criteria without scripting.
`obiannotate` — add or modify sequence attributes without scripting.

---

# NOTES

- The Lua `worker(sequence)` function is called in parallel across multiple goroutines. Use the thread-safe `obicontext` table (with `obicontext:lock()` / `obicontext:unlock()`) for any shared mutable state accessed across workers.
- The `begin()` and `finish()` hooks each run in a single goroutine and do not need locking for their own internal state.
- Sequence records modified inside `worker()` must be returned (or the original returned unmodified) for the record to appear in the output. Returning `nil` drops the sequence.
