## Semantic Description of `obiutils.Abs` Functionality

The provided Go test suite (`TestAbs`) validates the semantic behavior of a utility function `Abs` from the package [`obiutils`](https://git.metabarcoding.org/obitools/obitools4), part of the OBITools 4 ecosystem — a toolkit for DNA metabarcoding data analysis.

- **Function Purpose**:  
  `obiutils.Abs` computes the *absolute value* of an integer, returning its non-negative magnitude regardless of sign.

- **Test Coverage**:  
  The test verifies correctness across two categories:
    - *Non-negative inputs* (`0`, `1`, `5`, `10`) → outputs unchanged.
    - *Negative inputs* (`-1`, `-5`, `-10`) → outputs their positive counterparts.

- **Semantic Semantics**:  
  The function adheres to the mathematical definition: `Abs(x) = x` if `x ≥ 0`, else `-x`.  
  It ensures robustness for edge cases (e.g., zero) and typical integer ranges used in bioinformatic pipelines.

- **Integration Context**:  
  As part of `obitools4`, such low-level utilities likely support numerical operations in sequence alignment scoring, quality filtering, or coordinate transformations — where signed differences must be normalized.

- **Test Quality**:  
  Uses table-driven testing (Go idiom), promoting maintainability and clarity. No external dependencies are required — confirming the function is pure, deterministic, and self-contained.

In summary: `Abs` provides a foundational arithmetic primitive with guaranteed correctness for integer inputs, enabling reliable downstream computation in OBITools’ data processing workflows.
