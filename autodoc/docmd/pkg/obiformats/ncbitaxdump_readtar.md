## NCBI Taxonomy Archive Support in `obiformats`

This Go package provides utilities for handling **NCBI Taxonomy dumps archived as `.tar` files**.

### Core Functionalities

1. **Archive Validation (`IsNCBITarTaxDump`)**  
   - Checks whether a given `.tar` file contains all required NCBI Taxonomy dump files: `citations.dmp`, `division.dmp`, `gencode.dmp`, `names.dmp`, `delnodes.dmp`, `gc.prt`, `merged.dmp`, and `nodes.dmp`.  
   - Returns a boolean indicating if the archive is a complete NCBI tax dump.

2. **Taxonomy Loading (`LoadNCBITarTaxDump`)**  
   - Parses the `.tar` archive and extracts key files to build a `Taxonomy` object.  
   - Steps include:
     - **Nodes**: Loads taxonomic hierarchy (`nodes.dmp`) via `loadNodeTable`.
     - **Names**: Parses scientific and common names (`names.dmp`) via `loadNameTable`, with an option to load *only scientific names* (`onlysn`).
     - **Merged Taxa**: Integrates taxonomic aliases from `merged.dmp`, using `loadMergedTable`.
   - Sets the root taxon to NCBI’s default (`taxid = 1`, i.e., *root*).

3. **Integration with Other Modules**  
   - Uses `obiutils.Ropen`, `TarFileReader` for robust file handling.
   - Leverages `obitax.Taxonomy`, a structured representation of taxonomic data.

### Key Parameters
- `onlysn`: If true, only scientific names are loaded (reduces memory usage).
- `seqAsTaxa`: Reserved for future use; currently unused.

### Logging & Error Handling  
- Uses `logrus` to log loading progress and counts.
- Returns descriptive errors if required files or the root taxon are missing.

> **Note**: Designed for efficient, standards-compliant ingestion of NCBI Taxonomy data in bioinformatics pipelines.
