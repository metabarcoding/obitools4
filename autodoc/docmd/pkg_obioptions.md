# OBIOptions Package: Semantic Documentation

The `obioptions` package centralizes command-line interface (CLI) infrastructure for OBITools4, enabling consistent parsing of shared arguments and runtime configuration across tools. It standardizes logging, profiling controls, taxonomy integration, version reporting, and batch processing options—ensuring modularity, maintainability, and reproducibility.

## Core CLI Infrastructure

### Global Option Registration & Processing  
- `RegisterGlobalOptions(parser)` injects shared flags into any argument parser, including:
  - Version (`--version`) and debug mode (`--debug`)
  - Resource control: `--max-cpu`, thread limits, memory/batch tuning (`--batch-size`, `-size-max`, `--batch-mem`)
  - Quality encoding toggle: `--solexa`
  - Warning suppression (`--silent-warning`)  
- `ProcessParsedOptions(parser)` handles post-parsing logic:
  - Exits early on help/version requests
  - Loads taxonomy database via `obiformats.LoadTaxonomy()`
  - Sets log level (`logrus.SetLevel`)
  - Enables performance profiling via `pprof`:
    - Generic heap/goroutine dumps (`/debug/pprof`)
    - Mutex contention profiling (via `--pprof-mutex` + `runtime.SetMutexProfileFraction()`)
    - Goroutine blocking profiling (via `--pprof-goroutine` + `runtime.SetBlockProfileRate()`)

### Parser Generation  
- `GenerateOptionParser(program, documentation)` returns:
  - A reusable parser with bundled short options (`-abc`) and strict unknown-option rejection
  - Built-in `--help` support (via `go-getoptions`)
- Designed for reuse across commands with minimal boilerplate.

## Taxonomy Handling

### Option Set Registration  
- `LoadTaxonomyOptionSet(parser)` adds taxonomy-specific flags:
  - Required/optional DB path: `--taxonomy`, `-t`
  - Alternative names lookup (`--alternative-names`)
  - Validation strictness: `--fail-on-taxonomy`
  - Auto-update taxonomic IDs (`--update-taxid`)
  - Raw output mode: `--raw-taxid`
  - Inclusion of leaf sequences (`--with-leaves`)  
- Taxonomy loading is thread-safe (mutex-guarded) and lazy-loaded.

### Runtime Accessors  
- `CLIIsDebugMode()` → returns current debug state
- `SeqAsTaxa()` → indicates if sequence IDs should be treated as taxa (e.g., for `--raw-taxid`)
- `SetDebugOn()`, `SetDebugOff()` → programmatic toggling of debug mode

## Subcommand-Aware Parsing  

### `GenerateSubcommandParser(program, documentation, setup)`  
- Builds a hierarchical CLI:
  - Registers global options inherited by all subcommands
  - Invokes `setup(parser)` to define per-subcommand flags and commands  
- Automatically adds a built-in `help` subcommand for command-level documentation
- Returns:
  - Root parser (`*GetOpt`) and an `ArgumentParser` function with signature:  
    ```go
    func([]string) (*GetOpt, []string)
    ```
  - Parses CLI args (skipping binary name), handles errors via `ProcessParsedOptions`, and returns parsed state + positional arguments

## Versioning & Diagnostics  

### `VersionString()`  
- Returns the current OBITools version as `"Release X.Y.Z"` (e.g., `Release 4.4.29`)
- Version is auto-populated from a build-time-generated `version.txt` (via Makefile)
  - Patch level increments per commit → precise tracking of development iterations
- Pure function: no side effects, safe for logging/diagnostics/compatibility checks  
- Supports CI validation and runtime introspection (e.g., error reports, feature gates)

## Design Principles  

- **Environment Variables**: Configurable via `OBIMAXCPU`, `OBIWARNING`, etc.
- **Error Handling**: Parse errors → print help + exit gracefully
- **Standard Tooling Integration**:
  - `logrus` for structured logging  
  - Go’s native `pprof` (HTTP servers, mutex/block profiles)
- **Zero External Dependencies** for versioning module
