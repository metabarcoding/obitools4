# Microsatellite Detection Module (`obimicrosat`)

This Go package provides tools for identifying and annotating microsatellite (simple sequence repeat, SSR) regions within biological sequences.

## Core Functionality

- **`MakeMicrosatWorker(...)`**  
  Returns a `SeqWorker` that scans DNA sequences for microsatellite patterns matching user-defined constraints:
  - Minimum/maximum unit length (`minUnitLength`, `maxUnitLength`)
  - Minimum number of repeats (`minUnits`)
  - Overall minimum microsatellite length (`minLength`)
  - Minimum required flanking sequences on each side (`minflankLength`)
  - Optional reverse-complement reorientation flag (`reoriented`)

## Detection Algorithm

1. **Initial Pattern Matching**  
   Uses a regex of the form `([acgt]{m,n})\1{k,}` to find candidate repeats (where *m*, *n* = unit bounds; *k+1* ≥ `minUnits`).

2. **Unit Length Refinement**  
   Computes the minimal repeating unit via string rotation symmetry detection (`min_unit`).

3. **Strict Re-Scan**  
   Builds a refined regex using the exact unit length to ensure precise boundary detection.

4. **Flank Validation**  
   Ensures sufficient left/right flanking sequences (length ≥ `minflankLength`).

5. **Normalization & Orientation**  
   - Computes the lexicographically smallest rotation (and its reverse complement) to define a canonical unit.
   - Records orientation (`direct`/`reverse`) and, if `reoriented=true`, converts the sequence to its reverse complement.

## Output Annotations

Each detected microsatellite adds metadata attributes:
- `microsat_unit_length`, `microsat_unit_count`
- `seq_length`, `microsat` (full repeat region)
- Start/end positions (`microsat_from`, `microsat_to`)
- Canonical unit: `microsat_unit_normalized`
- Orientation flag and flanks (`microsat_left`, `microsat_right`)

## CLI Integration

- **`CLIAnnotateMicrosat(...)`**  
  Wraps the worker in a pipeline stage, applying it to an iterator of sequences.
- Uses CLI-configurable parameters (e.g., `CLIMinUnitLength()`) and supports parallel processing.
- Filters out sequences with no qualifying microsatellite matches.

## Dependencies

Leverages `obitools4` core types (`BioSequence`, iterators, default attributes) and the `regexp2` library for robust regex matching.
