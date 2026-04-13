# PCR Simulation CLI Options

This Go package (`obipcr`) provides a set of command-line interface (CLI) options for configuring *in silico* PCR simulations. It extends a base option parser (`getoptions.GetOpt`) with parameters specific to amplification modeling.

## Core Functionality

- **Primer Definition**: Requires user-provided forward and reverse primers (`--forward`, `--reverse`), supporting ambiguous nucleotide patterns via the `obitools4/pkg/obiapat` module.
- **Mismatch Tolerance**: Allows a configurable number of mismatches per primer (`--allowed-mismatches`, alias `-e`).
- **Amplicon Filtering**: Enforces length constraints on the amplified region (excluding primers) via `--min-length`/`-l` and `--max-length`/`L`.
- **Topology Handling**: Supports both linear (`default`) and circular sequences via `--circular`/`-c`.
- **Fragmentation Strategy**: For long input sequences, enables overlap-based fragmentation (`--fragmented`) to accelerate processing.
- **Extension Control**: Optionally appends flanking sequence fragments (`--delta`, alias `-D`) to amplicon ends.
- **Strict Flanking**: With `--only-complete-flanking`, only outputs amplicons where both primer-binding sites are fully present.

## Integration

- `PCROptionSet()` registers all PCR-specific flags.
- `OptionSet()` wraps this with standard conversion options (`obiconvert.OptionSet`).
- Getter functions (e.g., `CLIForwardPrimer()`, `CLIMinLength()`) safely expose parsed values, including pattern compilation and error handling.

## Design Notes

- All primer-related options are validated at parse time; missing required fields trigger fatal errors.
- Mismatch-tolerant primer matching is delegated to `obiapat.MakeApatPattern`.
