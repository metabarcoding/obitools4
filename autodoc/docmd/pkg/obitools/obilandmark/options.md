# `obilandmark` Package Overview

The `obilandmark` package provides command-line interface (CLI) options and utilities for selecting a specified number of landmark sequences in the OBITools4 framework.

## Core Functionality

- **`LandmarkOptionSet(options)`**:  
  Registers the `--center` (alias `-n`) integer option, defaulting to **200**, allowing users to specify how many landmark sequences should be selected.

- **`OptionSet(options)`**:  
  Aggregates option sets from related modules:
  - Input/output handling via `obiconvert.InputOptionSet` and `.OutputOptionSet`
  - Taxonomy loading support via `obioptions.LoadTaxonomyOptionSet` (disabled for required/strict usage)
  - Landmark-specific option registration via `LandmarkOptionSet`

- **`CLINCenter()`**:  
  Returns the user-specified (or default) number of landmark sequences (`_nCenter`) as an integer.

## Semantic Role

This package enables configuration-driven control over landmark selection—a key step in representational or clustering tasks within metabarcoding workflows—by exposing a clean, modular CLI interface aligned with OBITools4’s design principles.
