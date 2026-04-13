# `obisummary` Package Overview

The `obisummary` package provides command-line interface (CLI) configuration and output formatting utilities for the `obicount` tool within the OBITools4 ecosystem.

## Core Functionality

- **Option Parsing Setup**  
  - `SummaryOptionSet()`: Registers CLI flags specific to summary reporting:
    - `-json-output`, `-yaml-output` (boolean): Select output format.
    - `-map <attr>`: Specifies one or more map attributes to include in the summary.

- **Extended Option Aggregation**  
  - `OptionSet()`: Extends `SummaryOptionSet()` by appending input-handling options from the `obiconvert` package.

- **Output Format Detection**  
  - `CLIOutFormat()`: Returns `"yaml"` or `"json"` based on active flags (YAML takes precedence only if JSON is *not* enabled).

- **Map Attribute Access**  
  - `CLIHasMapSummary()`: Returns whether any map attributes were specified.
  - `CLIMapSummary()`: Retrieves the list of requested attribute names.

## Design Notes

- Uses global variables for state (e.g., `__json_output__`, `__map_summary__`).
- Designed for integration with the [`go-getoptions`](https://github.com/DavidGamba/go-getoptions) library.
- Minimal, focused scope: solely configures CLI behavior for summary generation—no data processing logic included.
