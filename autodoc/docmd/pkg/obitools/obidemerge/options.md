## `obidemerge` Package Overview

The `obidemerge` package provides command-line interface (CLI) support for **demerging** biological sequence data—typically used to reverse the merging of paired-end reads that were previously combined (e.g., during PCR or amplicon processing).  

- **Core Functionality**:  
  - Defines a CLI option `--demerge` (short alias `-d`) to specify *which data slot* should be demerged.  
  - The default value is `"sample"`, indicating the primary sample slot as target for demerging.  

- **Integration**:  
  - Extends `obiconvert.OptionSet`, inheriting standard conversion options (e.g., input/output formats, filtering).  
  - Uses `go-getoptions` for robust CLI argument parsing.  

- **Key APIs**:  
  - `DemergeOptionSet(options)`: Registers the `-d/--demerge` flag.  
  - `CLIDemergeSlot()`: Returns the currently selected slot name (e.g., `"sample"`), enabling downstream logic to extract and split merged records accordingly.  

- **Use Case**:  
  - Enables reprocessing of merged reads (e.g., for error correction or split-read analysis) by selecting the appropriate data stream to demerge.  
