# NAME

obiconvert — convertion of sequence files to various formats

---

# SYNOPSIS

```
obiconvert [--batch-mem <string>] [--batch-size <int>]
           [--batch-size-max <int>] [--compress|-Z] [--csv] [--debug]
           [--ecopcr] [--embl] [--fail-on-taxonomy] [--fasta]
           [--fasta-output] [--fastq] [--fastq-output] [--genbank]
           [--help|-h|-?] [--input-OBI-header] [--input-json-header]
           [--json-output] [--max-cpu <int>] [--no-order] [--no-progressbar]
           [--out|-o <FILENAME>] [--output-OBI-header|-O]
           [--output-json-header] [--paired-with <FILENAME>] [--pprof]
           [--pprof-goroutine <int>] [--pprof-mutex <int>] [--raw-taxid]
           [--silent-warning] [--skip-empty] [--solexa]
           [--taxonomy|-t <string>] [--u-to-t] [--update-taxid] [--version]
           [--with-leaves] [<args>]
```

---

# DESCRIPTION

obiconvert is a versatile command-line tool that converts biological sequence data between multiple standard bioinformatics formats. It enables biologists to process large datasets by reading from one format and writing to another, with support for quality scores, taxonomic annotations, and various input/output combinations. The tool is optimized for high-performance processing with configurable batching, parallel execution, and memory management.

Biologists use obiconvert to standardize sequence data for compatibility with different bioinformatics tools, extract quality information from FASTQ files into more readable formats, or convert between FASTA and FASTQ when working with DNA/RNA sequences that have associated quality data. The tool automatically detects input formats and intelligently selects output formats based on data presence (e.g., FASTQ when quality scores exist, FASTA otherwise). To force a specific output format regardless of input content, use the explicit output flags (`--fasta-output`, `--fastq-output`, `--json-output`). <!-- corrected: without --fasta-output, a FASTQ input with quality scores stays in FASTQ format even when the output filename has a .fasta extension -->

---

# INPUT

obiconvert accepts input in multiple biological sequence formats:

- **FASTA**: Standard text-based format with `>` headers and sequence data
- **FASTQ**: Binary quality-score format (default when both sequence and quality data present)
- **GenBank**: Comprehensive biological record format with annotations
- **EMBL**: EMBL flatfile format for sequence and feature information
- **ecoPCR**: Specialized output format from ecoPCR amplification tools
- **CSV**: Tabular sequence data with configurable delimiters

Input is provided as positional arguments (file paths or `-` for stdin). The tool automatically detects the input format based on file content and can handle multiple files in sequence. When paired-end sequencing is used, the `--paired-with` option specifies the mate read file.

---

# OUTPUT

obiconvert produces sequence data in several output formats depending on input content and user selection:

- **FASTA**: Text format with sequence only (use `--fasta-output` to force)
- **FASTQ**: Format including quality scores (default when quality data present; use `--fastq-output` to force)
- **JSON**: Structured output with all sequence metadata and attributes (use `--json-output`)

The tool preserves all sequence annotations (taxonomic information, custom attributes) during conversion. When converting to FASTA/FASTQ formats, title line annotations can be formatted as OBI structured data or JSON using the `--output-OBI-header`/`--output-json-header` flags. Sequences of zero length can be suppressed with `--skip-empty`.

## Observed output example

```
>seq001 {"definition":"DNA sequence with quality scores for FASTQ to FASTA conversion"}
atcgatcgatcgatcgatcgatcgatcgatcgatcgatcg
>seq002 {"definition":"Second sequence with moderate quality scores"}
gctagctagctagctagctagctagctagctagctagct
>seq003 {"definition":"Third sequence with high quality scores"}
ttaaccggttaaccggttaaccggttaaccggttaaccg
>seq004 {"definition":"Fourth sequence with variable quality scores"}
acgtacgtacgtacgtacgtacgtacgtacgtacgtacg
```

---

# OPTIONS

## Input Format Options
- **--fasta**: Read data following the fasta format. (default: false)
- **--fastq**: Read data following the fastq format. (default: false)
- **--genbank**: Read data following the Genbank flatfile format. (default: false)
- **--embl**: Read data following the EMBL flatfile format. (default: false)
- **--ecopcr**: Read data following the ecoPCR output format. (default: false)
- **--csv**: Read data following the CSV format. (default: false)

## Input Header Options
- **--input-OBI-header**: FASTA/FASTQ title line annotations follow OBI format. (default: false)
- **--input-json-header**: FASTA/FASTQ title line annotations follow json format. (default: false)

## Output Format Options
- **--fasta-output**: Write sequence in fasta format (default if no quality data available). (default: false)
- **--fastq-output**: Write sequence in fastq format (default if quality data available). (default: false)
- **--json-output**: Write sequence in json format. (default: false)

## Output Header Options
- **--output-OBI-header|-O**: output FASTA/FASTQ title line annotations follow OBI format. (default: false)
- **--output-json-header**: output FASTA/FASTQ title line annotations follow json format. (default: false)

## Processing Options
- **--skip-empty**: Sequences of length equal to zero are suppressed from the output (default: false)
- **--no-order**: When several input files are provided, indicates that there is no order among them. (default: false)
- **--u-to-t**: Convert Uracil to Thymine. (default: false)
- **--update-taxid**: Make obitools automatically updating the taxid that are declared merged to a newest one. (default: false)
- **--raw-taxid**: When set, taxids are printed in files with any supplementary information (taxon name and rank) (default: false)
- **--fail-on-taxonomy**: Make obitools failing on error if a used taxid is not a currently valid one (default: false)
- **--with-leaves**: If taxonomy is extracted from a sequence file, sequences are added as leave of their taxid annotation (default: false)

## File Options
- **--out|-o <FILENAME>**: Filename used for saving the output (default: "-")
- **--paired-with <FILENAME>**: Filename containing the paired reads (default: "")

## Performance Options
- **--batch-mem <string>**: Maximum memory per batch (e.g. 128K, 64M, 1G; default: 128M). Set to 0 to disable. (default: "", env: OBIBATCHMEM)
- **--batch-size <int>**: Minimum number of sequences per batch (floor, default 1) (default: 1, env: OBIBATCHSIZE)
- **--batch-size-max <int>**: Maximum number of sequences per batch (ceiling, default 2000) (default: 2000, env: OBIBATCHSIZEMAX)
- **--max-cpu <int>**: Number of parallele threads computing the result (default: 16, env: OBIMAXCPU)
- **--compress|-Z**: Compress all the result using gzip (default: false)

## Debug Options
- **--debug**: Enable debug mode, by setting log level to debug. (default: false, env: OBIDEBUG)
- **--silent-warning**: Stop printing of the warning message (default: false, env: OBIWARNING)
- **--no-progressbar**: Disable the progress bar printing (default: false)

## Profiling Options
- **--pprof**: Enable pprof server. Look at the log for details. (default: false)
- **--pprof-goroutine <int>**: Enable profiling of goroutine blocking profile. (default: 6060, env: OBIPPROFGOROUTINE)
- **--pprof-mutex <int>**: Enable profiling of mutex lock. (default: 10, env: OBIPPROFMUTEX)

## Utility Options
- **--taxonomy|-t <string>**: Path to the taxonomy database. (default: "")
- **--solexa**: Decodes quality string according to the Solexa specification. (default: false, env: OBISOLEXA)
- **--help|-h|-?**: Show help message (default: false)
- **--version**: Prints the version and exits. (default: false)

---

# EXAMPLES

## Convert FASTQ to FASTA
```bash
# Convert quality-score data from FASTQ to readable FASTA format
obiconvert --fastq --fasta-output input.fastq -o output.fasta
```

**Expected output:** 4 sequences written to `output.fasta`.

## Convert FASTA to JSON
```bash
# Convert sequences to structured JSON format preserving all annotations
obiconvert --fasta --json-output input.fasta -o output.json
```

**Expected output:** 3 sequences written to `output.json`.

## Process paired-end sequencing data
```bash
# Convert paired FASTQ files preserving read pairing
obiconvert --fastq --fasta-output forward.fastq --paired-with reverse.fastq -o merged_sequences.fasta
```

**Expected output:** 4 sequences written to `merged_sequences_R1.fasta` and `merged_sequences_R2.fasta`.

---

# SEE ALSO

- obiannotate: Add taxonomic and functional annotations to sequences
- obicount: Count sequences in files
- obigrep: Filter sequences based on attributes or patterns
- obisummary: Generate statistics from sequence files
- obiuniq: Remove duplicate sequences

---

# NOTES

obiconvert automatically selects the optimal output format based on input data presence, preferring FASTQ when quality scores are available and FASTA otherwise. To force a specific output format, use `--fasta-output`, `--fastq-output`, or `--json-output` explicitly. <!-- corrected: the output format is NOT determined by the output filename extension; it must be forced with explicit flags -->

Memory usage is controlled through batch processing, with configurable memory limits per batch to handle large datasets efficiently. Progress reporting can be disabled for scripting purposes using `--no-progressbar`.

When working with taxonomic data, ensure the taxonomy database is accessible and properly formatted to avoid failures during sequence annotation processing.
