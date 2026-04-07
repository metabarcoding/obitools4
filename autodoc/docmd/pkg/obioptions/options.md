# OBIOptions Package: Global Command-Line Interface Utilities

The `obioptions` package provides shared command-line argument parsing and runtime configuration for OBITools4 commands. It centralizes common options, logging setup, profiling controls, and taxonomy handling.

## Core Functionalities

- **Global Option Registration**: `RegisterGlobalOptions()` defines shared flags such as:
  - `--version`, `--debug`
  - CPU/thread control (`--max-cpu`)
  - Batch processing parameters: `--batch-size`, `-size-max`, `--batch-mem`
  - Quality encoding (`--solexa`)
  - Warning suppression (`--silent-warning`)

- **Option Processing**: `ProcessParsedOptions()` handles post-parsing logic:
  - Prints help/version and exits on request
  - Loads default taxonomy via `obiformats.LoadTaxonomy()`
  - Configures log level (debug/info)
  - Starts `pprof` HTTP servers for performance profiling:
    - Generic (`/debug/pprof`)
    - Mutex contention (`--pprof-mutex`, `runtime.SetMutexProfileFraction()`)
    - Goroutine blocking (`--pprof-goroutine`, `runtime.SetBlockProfileRate()`)

- **Parser Generator**: `GenerateOptionParser()` builds a reusable argument parser with:
  - Bundled short options (`-abc`)
  - Strict unknown-option handling
  - Automatic `--help` support

## Taxonomy Integration

- `LoadTaxonomyOptionSet()` registers taxonomy-specific flags:
  - Required/optional path to DB (`--taxonomy`, `-t`)
  - Alternative names search (`--alternative-names`)
  - Taxonomic validation: `--fail-on-taxonomy`, automatic updates via `--update-taxid`
  - Raw taxID output (`--raw-taxid`)
  - Leaf sequences inclusion via `--with-leaves`

## Runtime Accessors

- `CLIIsDebugMode()`, `SeqAsTaxa()` → read current state
- `SetDebugOn/Off()` → programmatic debug toggling

## Design Principles

- Environment variable support (`OBIMAXCPU`, `OBIWARNING`, etc.)
- Thread-safe taxonomy loading with mutex
- Graceful error handling (parse errors → help + exit)
- Integration with `logrus` and Go’s standard profiling tools
