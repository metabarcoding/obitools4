# CSVSequenceRecord Function Description

The `CSVSequenceRecord` function converts a biological sequence object (`*obiseq.BioSequence`) into a slice of strings suitable for CSV output. It dynamically constructs the record based on user-defined options (`opt Options`), enabling flexible column selection.

## Core Features

- **Sequence ID**: Includes the sequence identifier if `opt.CSVId()` is enabled.
- **Abundance Count**: Appends the sequence count (e.g., read depth) if `opt.CSVCount()` is true.
- **Taxonomic Information**: Adds both NCBI taxid and scientific name (retrieved from attributes or fallback via `opt.CSVNAValue()`).
- **Definition Line**: Includes the sequence definition/description if requested via `opt.CSVDefinition()`.
- **Custom Attributes**: Iterates over keys from `opt.CSVKeys()` and appends corresponding attribute values (or NA if missing).
- **Nucleotide Sequence**: Appends the raw sequence string when `opt.CSVSequence()` is enabled.
- **Quality Scores**: Converts Phred-quality scores to ASCII characters (using a configurable shift) if available; otherwise inserts NA.

## Design Highlights

- Uses `obiutils.InterfaceToString()` for safe type conversion of arbitrary attribute values.
- Handles missing data consistently via `opt.CSVNAValue()`.
- Supports both standard and user-defined metadata fields.
- Adapts quality encoding to common formats (e.g., Sanger/Illumina) via `obidefault.WriteQualitiesShift()`.

This function enables interoperable, configurable export of sequence data to tabular formats.
