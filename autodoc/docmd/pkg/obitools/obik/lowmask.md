# Low-Complexity Sequence Masking with Entropy-Based Detection

This Go package implements a low-complexity masking tool for DNA sequences, based on entropy analysis of *k*-mer frequencies across multiple window sizes. It identifies regions with reduced sequence diversity—typical of repeats or simple sequences—and supports three operational modes: **masking**, **splitting**, and **extracting** masked fragments.

## Core Functionality

- `lowMaskWorker()` constructs a reusable worker function that processes sequences using entropy-based detection.
- Entropy is computed for sliding windows of varying sizes (from 1 to `level_max`) using normalized canonical circular *k*-mers.
- Ambiguous nucleotides (non-acgt) are automatically masked and excluded from entropy calculations.

## Algorithm Highlights

- **Multi-scale analysis**: Computes sequence complexity at multiple *k*-mer lengths to capture both local and broader low-complexity patterns.
- **Sliding window entropy**: Uses a frequency table updated incrementally (via deque-based sliding minimum optimization) to efficiently compute Shannon entropy per position.
- **Thresholding**: Positions with entropy ≤ `threshold` are flagged as low-complexity.

## Output Modes

- **MaskMode**: Replaces masked positions with a user-defined character (`maskChar`).
- **SplitMode (default)**: Splits the sequence into high-complexity fragments ≥ *k*-mer size.
- **ExtractMode**: Extracts only the low-complexity fragments (e.g., for downstream filtering or analysis).

## Additional Features

- Preserves short fragments if `keepShorter` is enabled.
- Attaches metadata attributes (`mask`, `Entropies`) to each sequence for inspection or post-processing.
- Integrates with the OBITools4 pipeline via `runLowmask()` for CLI usage and batch processing.

This implementation is optimized for speed (via incremental updates, precomputed normalization tables) while maintaining biological accuracy in complexity estimation.
