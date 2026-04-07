# `obimicrosat`: Microsatellite Detection Module for OBITools4

This Go package provides a modular, CLI-integrated framework to detect and annotate simple sequence repeats (SSRs), also known as microsatellites, in biological DNA sequences. It is designed for integration into sequence processing pipelines—especially those focused on marker discovery, PCR primer design, or genomic feature annotation.

## Core Capabilities

### 1. **Flexible Microsatellite Detection**
- Detects tandem repeats of DNA motifs (units) with user-defined constraints:
  - Unit length range (`minUnitLength` to `maxUnitLength`, typically 1–6 bp)
  - Minimum repeat count (`minUnits`)
  - Total microsatellite length threshold (`minLength`)
- Uses robust regex-based scanning via `regexp2`, followed by precise boundary refinement.

### 2. **Canonical Unit Normalization**
- Determines the *lexicographically smallest* rotation of the detected unit.
- Optionally computes its reverse complement to define orientation (`direct` or `reverse`).
- If enabled, reorients the full microsatellite region to its canonical (smallest) form.

### 3. **Flanking Sequence Validation**
- Ensures sufficient unique sequence on both sides of the repeat (`minflankLength`).
- Stores flanking regions as `microsat_left` and `microsat_right`.

### 4. **Structured Annotation Output**
Each detected microsatellite enriches the input `BioSequence` with standardized attributes:
- `microsat_unit_length`, `microsat_unit_count`
- `seq_length` (full repeat region length), `microsat` (repeat sequence)
- Positions: `microsat_from`, `microsat_to`
- Canonical unit: `microsat_unit_normalized`
- Orientation flag (`direct`/`reverse`) and flanks

### 5. **CLI Integration & Pipeline Compatibility**
- `MicroSatelliteOptionSet()` registers all detection parameters for CLI use (via `go-getoptions`).
- Supported flags:
  - `-m, --min-unit-length`: min unit size (default: `1`)
  - `-M, --max-unit-length`: max unit size (default: `6`)
  - `--min-unit-count`: min repeat count (default: `5`)
  - `-l, --min-length`: total SSR length threshold (default: `20`)
  - `-f, --min-flank-length`: required flanking length (default: `0`)
  - `-n, --not-reoriented`: disable sequence reorientation
- Helper functions (e.g., `CLIMinUnitCount()`, `CLIReoriented()`) expose runtime config.
- `MakeMicrosatWorker()` returns a reusable `SeqWorker` for parallel, iterator-based processing.
- `CLIAnnotateMicrosat()` integrates the worker into a conversion pipeline, filtering sequences without qualifying SSRs.

### 6. **Dependencies & Ecosystem Integration**
- Built on `obitools4` core types (`BioSequence`, iterators, default annotation schema).
- Uses only external dependency: `github.com/dlclark/regexp2` for advanced regex support.
- Fully compatible with existing `obiconvert.OptionSet`; extends it via `OptionSet()`.

## Use Cases
- Identification of polymorphic SSR markers for population genetics.
- Preprocessing step in PCR primer design tools (to avoid repeat-rich regions).
- Quality control: flagging low-complexity sequences in NGS data.

> **Note**: Only *public* APIs are documented. Internal helpers (e.g., `min_unit`, rotation logic) remain implementation details.
