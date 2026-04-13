# `obik` Command-Line Tool: Semantic Feature Overview

The `obik` tool is a command-line utility for managing and analyzing **k-mer indices**—disk-based data structures storing k-mer frequencies from biological sequences. Below is a semantic summary of its subcommands:

1. **`index`**  
   Builds or extends a k-mer index from input sequence files (FASTA/FASTQ), supporting metadata tagging and output mode configuration.

2. **`ls`**  
   Lists all sets (i.e., grouped k-mer collections) stored in an index, with customizable output formatting and set selection filters.

3. **`summary`**  
   Displays detailed statistics (e.g., total k-mers, unique counts) per set; optionally computes a pairwise **Jaccard distance matrix** for similarity assessment.

4. **`cp`, `mv`, `rm`**  
   Manage sets within or across indices: copy (`cp`) and move (`mv`) preserve or relocate data; remove (`rm`) deletes sets. All support force-overwrite and set-selection flags.

5. **`spectrum`**  
   Outputs the k-mer frequency spectrum (histogram of how many times each k-mer occurs) as CSV, per selected sets.

6. **`super`**  
   Extracts *super k-mers*—longer contiguous sequences built from overlapping reads—from input files, using optimized overlap logic.

7. **`lowmask`**  
   Masks low-complexity regions in sequences (e.g., homopolymers, repeats) using entropy-based detection.

8. **`match`**  
   Annotates input sequences with positions where k-mers match those in a stored index, enabling read tagging or reference mapping.

9. **`filter`**  
   Removes low-complexity k-mers from an index using entropy thresholds, optionally applied to selected sets.

All commands integrate with shared option groups (e.g., input/output handling, set selection), ensuring consistent usage and composability.
