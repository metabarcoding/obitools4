# Quality Shift and Read/Write Control Module

This Go package (`obidefault`) provides configurable controls over quality score handling in sequence data processing (e.g., FASTQ files). It defines three global variables and corresponding accessor/mutator functions:

- `_Quality_Shift_Input`: Input quality score offset (default: `33`, i.e., Phred+33/Sanger format).
- `_Quality_Shift_Output`: Output quality score offset (default: `33`), allowing format conversion.
- `_Read_Qualities`: Boolean flag indicating whether quality scores should be parsed/processed (`true` by default).

## Public API

| Function | Purpose |
|---------|--------|
| `SetReadQualitiesShift(shift byte)` | Sets the quality score offset for *input* data (e.g., when reading FASTQ). |
| `ReadQualitiesShift() byte` | Returns the current input quality offset. |
| `SetWriteQualitiesShift(shift byte)` | Sets the quality score offset for *output* data (e.g., when writing FASTQ). |
| `WriteQualitiesShift() byte` | Returns the current output quality offset. |
| `SetReadQualities(read bool)` | Enables/disables reading/processing of quality scores. |
| `ReadQualities() bool` | Returns whether qualities are currently being read/used. |

## Semantic Use Cases

- **Format Interoperability**: Allows seamless conversion between Phred+33 (Sanger), Phred+64, or other quality encodings.
- **Performance Optimization**: Disabling `ReadQualities` skips parsing of quality strings, useful when only sequences are needed.
- **Centralized Configuration**: Global state enables consistent behavior across modules without passing parameters.

All functions are thread-unsafe by design—intended for initialization before concurrent processing begins.
