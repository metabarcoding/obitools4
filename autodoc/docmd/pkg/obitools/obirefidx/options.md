# `obirefidx` Package Overview

The `obirefidx` package provides command-line option configuration for the `obiuniq` tool within the OBITools4 ecosystem.

- **Purpose**: Extends generic option parsing to support `obiuniq`'s specific flags.
- **Core Function**:  
  ```go
  func OptionSet(options *getoptions.GetOpt)
  ```
- **Behavior**:  
  Delegates to `obiconvert.OptionSet(false)`, inheriting all standard conversion options (e.g., input/output formats, filtering thresholds), but *without* enabling verbose mode (`false` → no extra logging).
- **Dependencies**:  
  - `getoptions`: For robust CLI argument parsing.  
  - `obiconvert`: Shared conversion utilities and option definitions.
- **Semantic Role**: Acts as a *feature gate*—ensuring only relevant `obiconvert` options are exposed to the user for deduplication tasks.
- **Use Case**: Used during CLI initialization (e.g., `obiuniq --input file.fastq`) to validate and bind user-provided flags.

In essence, `obirefidx` ensures consistent, minimal option exposure for reference-based deduplication workflows in OBITools4.
