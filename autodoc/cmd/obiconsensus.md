# obiconsensus(1) — OBITools4 Manual

## NAME

`obiconsensus` — denoise Oxford Nanopore Technology (ONT) reads by building consensus sequences

## SYNOPSIS

```
obiconsensus [OPTIONS] [FILE...]
```

## DESCRIPTION

`obiconsensus` is designed to correct sequencing errors in long reads produced by Oxford Nanopore Technology (ONT) sequencers. Because ONT reads have a relatively high error rate compared to short-read technologies, sequences originating from the same biological molecule can differ slightly from one another. `obiconsensus` groups these related reads and builds a single, more reliable consensus sequence for each group.

The tool works by constructing a *difference graph*: each unique read is represented as a node, and two nodes are connected if their sequences differ by at most a small number of positions (controlled by `--distance`). Within each sample, clusters of closely related reads are identified, and a consensus is assembled from the cluster members using a *de Bruijn graph* approach. The result is a set of high-quality representative sequences, one per cluster.

Two denoising strategies are available:

- **Standard mode** (default): identifies hub nodes (likely true sequences) in the difference graph and builds a consensus from each hub and its immediate neighbours.
- **Clustering mode** (`--cluster`): groups reads around local abundance maxima and builds a consensus from each neighbourhood.

Sequences are read from one or more files, or from standard input when no file is given. Results are written to standard output or to a file specified with `--out`.

The tool processes data on a per-sample basis. Sample identity is taken from a sequence annotation attribute (default: `sample`). Each sample's reads are denoised independently.

## INPUT FORMATS

`obiconsensus` recognises the following input formats automatically. A specific format can be forced with the corresponding flag:

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

## DENOISING OPTIONS

`--distance INT`, `-d INT`
: Maximum number of differences allowed between two reads for them to be considered related and placed in the same cluster. Default: 1. A value of 1 means reads differing by a single nucleotide substitution are grouped together.

`--cluster`, `-C`
: Switch to clustering mode. Instead of identifying hub sequences, reads are grouped around local abundance maxima. This mode may produce fewer but more representative consensus sequences.

`--kmer-size SIZE`
: Size of the short words (k-mers) used when building the de Bruijn graph for consensus assembly. The default value of `-1` means the size is estimated automatically from the data. Manual adjustment is rarely needed.

`--no-singleton`
: Discard any read (or cluster) that occurs only once across the dataset. Singleton sequences are often the result of sequencing errors and carry little biological signal.

`--low-coverage FLOAT`
: Discard any sample whose sequence coverage falls below this threshold. Default: 0 (no filtering). Useful for removing poorly sequenced samples.

`--sample ATTRIBUTE`, `-s ATTRIBUTE`
: Name of the sequence annotation attribute that identifies the sample of origin. Default: `sample`. Each unique value of this attribute is treated as an independent sample during denoising.

## OUTPUT ANNOTATION OPTIONS

`--unique`, `-U`
: After denoising, dereplicate the output sequences (equivalent to running `obiuniq`). Identical consensus sequences across samples are merged into a single record carrying abundance information.

`--save-graph DIRECTORY`
: Save the difference graphs built during denoising to the specified directory. Each graph is written in GraphML format, one file per sample. Useful for inspecting the clustering structure.

`--save-ratio FILE`
: Save a table of abundance ratios on graph edges to the specified CSV file. Each row describes the relative abundance of a read compared to its neighbours. Useful for quality control and parameter tuning.

## PERFORMANCE OPTIONS

`--max-cpu INT`
: Number of parallel threads to use for computation. Default: all available processors (up to 16). Reducing this value limits memory and CPU usage.

`--batch-size INT`
: Minimum number of sequences processed together in a single batch. Default: 1.

`--batch-size-max INT`
: Maximum number of sequences in a single batch. Default: 2000.

`--batch-mem STRING`
: Maximum memory allocated per batch (e.g., `128M`, `1G`). Default: `128M`. Set to `0` to disable the memory limit.

`--no-progressbar`
: Disable the progress bar.

`--no-order`
: When reading from multiple files, indicate that there is no meaningful order among them. This can improve performance for large multi-file inputs.

## OTHER OPTIONS

`--u-to-t`
: Convert uracil (U) to thymine (T) in all input sequences. Use this option when working with RNA data stored in a DNA context.

`--skip-empty`
: Remove sequences of length zero from the output.

`--solexa`
: Interpret quality scores using the Solexa encoding rather than the standard Phred encoding.

`--silent-warning`
: Suppress warning messages.

`--debug`
: Enable detailed logging for troubleshooting.

`--version`
: Print the version number and exit.

`--help`, `-h`
: Display a brief help message and exit.

## OUTPUT ATTRIBUTES

Each output consensus sequence carries several annotation attributes describing how it was built:

| Attribute | Description |
|-----------|-------------|
| `consensus` | Boolean flag: `true` if the sequence is a true consensus, `false` if it was kept unchanged (e.g., isolated singleton) |
| `merged_sample` | Map of sample names to read counts contributing to this consensus |
| `count` | Total number of reads merged into this consensus across all samples |
| `kmer_size` | Size of the k-mers used to build the de Bruijn graph for this consensus |
| `seq_length` | Length of the consensus sequence |

## EXAMPLES

**Basic denoising of a FASTQ file:**

```sh
obiconsensus reads.fastq > denoised.fastq
```

**Increase the allowed distance between reads to 2:**

```sh
obiconsensus --distance 2 reads.fastq > denoised.fastq
```

**Use clustering mode and remove singletons:**

```sh
obiconsensus --cluster --no-singleton reads.fastq > denoised.fastq
```

**Denoise, then dereplicate the output:**

```sh
obiconsensus --unique reads.fastq > denoised_uniq.fastq
```

**Save denoising graphs for inspection:**

```sh
obiconsensus --save-graph ./graphs reads.fastq > denoised.fastq
```

**Specify the sample annotation attribute:**

```sh
obiconsensus --sample library reads.fastq > denoised.fastq
```

## SEE ALSO

`obiuniq`(1), `obiclean`(1), `obigrep`(1), `obiconvert`(1)

## NOTES

`obiconsensus` was designed primarily for Oxford Nanopore Technology amplicon data, where individual reads of the same molecule may carry different sequencing errors. For short-read Illumina data, `obiclean` may be more appropriate.

The automatic k-mer size selection (`--kmer-size -1`) works well in most cases. If the consensus assembly fails for a group (e.g., due to circular structures in the de Bruijn graph), the k-mer size is progressively increased until the assembly succeeds or a fallback strategy is used.
