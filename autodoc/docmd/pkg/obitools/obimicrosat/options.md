# MicroSatellite Module Overview

This Go package (`obimicrosat`) provides command-line interface (CLI) configuration and utility functions for detecting microsatellite sequences in DNA data within the OBITools4 ecosystem.

## Core Functionality

- **CLI Option Setup**:  
  `MicroSatelliteOptionSet()` registers user-configurable parameters for microsatellite detection via the `go-getoptions` library.

- **Supported Options**:
  - `-m, --min-unit-length`: Minimum length (1–6 bp) of the repeating unit.
  - `-M, --max-unit-length`: Maximum length (default: 6 bp) of the repeating unit.
  - `--min-unit-count`: Minimum number of repeated units (default: 5).
  - `-l, --min-length`: Minimum total microsatellite length (default: 20 bp).
  - `-f, --min-flank-length`: Minimum length of flanking regions (default: 0).
  - `-n, --not-reoriented`: If set, disables reorientation of detected microsatellites.

- **Helper Functions**:
  - `CLIMinUnitLength()` / `CLIMaxUnitLength()`: Return min/max unit lengths.
  - `CLIMinUnitCount()` / `CLIMicroSatRegex()`: Return min unit count and a regex pattern for detection (e.g., `([acgt]{1,6})\1{4}`).
  - `CLIMinLength()` / `CLIMinFlankLength()`: Return min total length and flank size.
  - `CLIReoriented()` / `_NotReoriented`: Indicates whether reorientation is enabled.

- **Integration**:
  `OptionSet()` extends the base OBITools4 conversion options (`obiconvert.OptionSet`) with microsatellite-specific settings.

## Use Case

Designed for use in PCR simulation or marker identification pipelines, enabling flexible tuning of microsatellite detection thresholds directly from the CLI.
