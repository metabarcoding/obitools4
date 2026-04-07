# Semantic Description of `obingslibrary` Package

The `obingslibrary` package defines core data structures and methods for managing **PCR-based NGS library designs**, particularly in metabarcoding workflows.

- `PrimerPair` and `TagPair`: Represent forward/reverse primer or tag sequences.
- `PCR`: Encapsulates a single PCR amplification experiment with sample metadata and annotations (via `obiseq.Annotation`).
- `NGSLibrary`: Central struct storing primer definitions (`Primers`) and associated marker specifications (`Markers`), where each `Marker` defines how primers (and attached tags) are processed.

Key functionality includes:
- **Dynamic marker creation**: `GetMarker()` lazily initializes a new `Marker` for any primer pair if not already present.
- **Compilation**: Two compilation modes (`Compile`, `Compile2`) prepare internal search structures (e.g., error-tolerant index) using user-defined parameters like max errors and indel allowance.
- **Tag configuration**: Methods to set spacer length, delimiter character (e.g., `N` or `X`), and indel tolerance for tags—globally (`SetTagSpacer`, etc.) or per-primer.
- **Matching strategy**: Configure alignment behavior (e.g., `"strict"` vs. `"fuzzy"`) via `SetMatching*`.
- **Unicity & validation**: `CheckPrimerUnicity()` ensures no primer is reused across multiple markers and prevents self-complementary pairs.
- **Error handling**: Supports configurable mismatch/indel budgets per primer direction.

This library enables flexible, reproducible specification of molecular identifiers (tags) and amplification primers—essential for accurate demultiplexing and sequence assignment in high-throughput sequencing pipelines.
