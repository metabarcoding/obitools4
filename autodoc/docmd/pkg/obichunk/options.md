# Semantic Description of `obichunk` Package

The `obichunk` package provides a flexible and configurable options management system for data processing pipelines, particularly in the context of biological sequence analysis (e.g., metabarcoding). It defines a typed `Options` struct and associated builder-style configuration functions.

## Core Concepts

- **Immutable Configuration Builder**: Options are constructed via `MakeOptions([]WithOption)`, applying a list of functional setters (`WithOption`) to an internal `__options__` struct.
- **Encapsulation**: The concrete options are hidden behind a pointer (`pointer *__options__`) to ensure safe sharing and mutation control.

## Supported Functionalities

- **Categorization**: `OptionSubCategory(keys...)` appends category labels (e.g., sample or marker names) to an internal list; `PopCategories()` retrieves and removes the first category.
- **Missing Value Handling**: `OptionNAValue(na string)` customizes placeholder for missing data (default: `"NA"`).
- **Statistical Tracking**: `OptionStatOn(keys...)` registers statistical descriptions (via `obiseq.StatsOnDescription`) for per-field metrics collection.
- **Batch Processing Control**:
  - `OptionBatchCount(number)` sets the number of batches.
  - `OptionsBatchSize(size)` defines how many items per batch (default from `obidefault`).
- **Parallelization**: `OptionsParallelWorkers(nworkers)` configures concurrency level (default from environment).
- **Disk vs Memory Sorting**: `OptionSortOnDisk()` enables disk-backed sorting; `OptionSortOnMemory()` disables it (default).
- **Singleton Filtering**: `OptionsNoSingleton()` excludes singleton sequences; `OptionsWithSingleton()` allows them (default).

## Design Highlights

- Functional options pattern for extensibility and readability.
- Default values derived from `obidefault` where applicable (e.g., batch size, workers).
- Designed for integration with `obiseq` and `obidefault`, supporting scalable, reproducible NGS data workflows.
