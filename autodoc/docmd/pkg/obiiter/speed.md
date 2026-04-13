# `Speed` Functionality Description

The provided Go code defines a method and helper function to add **real-time progress tracking** to biosequence iterators in the OBITools4 framework.

## Core Features

- **Non-intrusive progress bar**:  
  The `Speed()` method wraps an existing iterator and displays a visual progress indicator on stderr, using the [`progressbar`](https://github.com/schollz/progressbar) library.

- **Conditional rendering**:  
  The progress bar is only shown when:
    - `--no-progressbar` flag is *not* set (via `obidefault.ProgressBar()`),
    - stderr is connected to a terminal (`os.ModeCharDevice`),
    - stdout is *not* piped (to avoid interfering with file output).

- **Batch-aware counting**:  
  Progress is updated per batch (`batch.Len()`), not item-by-item, for efficiency and smoother UI updates (throttled to ≥100ms).

- **Paired-end support**:  
  If the input iterator is paired (`IsPaired()`), this property is preserved in the returned iterator.

- **Pipeable wrapper**:  
  `SpeedPipe()` enables integration into functional pipelines (e.g., `.Map(...).Filter(...)`) by returning a `Pipeable` function.

## Implementation Highlights

- Uses goroutines to decouple iteration and progress updates.
- Automatically closes the output iterator when input ends (`WaitAndClose()`).
- Prints a final newline to stderr upon completion.

This utility enhances user experience during long-running sequence processing (e.g., FASTQ parsing, alignment), without affecting correctness or performance in non-interactive contexts.
