# Sequence Predicate Framework in `obiseq`

This Go package provides a flexible and composable predicate system for filtering biological sequences (`BioSequence`) based on diverse criteria.

## Core Concepts

- **`SequencePredicate`**: A function type `func(*BioSequence) bool`, enabling conditional logic on sequences.
- **Predicate Composition**: Supports logical operations (`And`, `Or`, `Xor`, `Not`) and chaining.
- **Paired-end Support**: Predicates can be adapted to consider read pairs via `PredicateOnPaired` and `PairedPredicat`, with modes:  
  - `ForwardOnly`: Only the forward read is evaluated.  
  - `ReverseOnly`, `And`, `Or`, `AndNot`, `Xor`: Combine forward and reverse evaluations.

## Built-in Predicates

| Predicate | Description |
|-----------|-------------|
| `HasAttribute(name)` | Checks if a sequence has an annotation with the given name. |
| `IsAttributeMatch(name, pattern)` | Tests if a named annotation matches the provided regex (case-sensitive). |
| `IsMoreAbundantOrEqualTo(count)` / `IsLessAbundantOrEqualTo(count)` | Filters by sequence abundance (count field). |
| `IsLongerOrEqualTo(length)` / `IsShorterOrEqualTo(length)` | Filters by sequence length. |
| `OccurInAtleast(sample, n)` | Checks if the sequence appears in at least *n* samples (via description stats). |
| `IsSequenceMatch(pattern)` | Matches the raw sequence against a regex (case-insensitive). |
| `IsDefinitionMatch(pattern)` | Matches the definition/description line against a regex. |
| `IsIdMatch(pattern)` / `IsIdIn(ids...)` | Filters by sequence ID using regex or explicit set. |
| `ExpressionPredicat(expression)` | Evaluates a custom boolean expression (via OBILang) using annotations and sequence metadata. |

## Design Highlights

- **Null-safe**: `nil` predicates are handled gracefully in compositions.
- **Extensible**: Custom predicates can be defined and combined seamlessly.
- **Logging & Safety**: Invalid regex patterns or expression syntax trigger fatal errors; runtime evaluation issues emit warnings.

This framework enables powerful, declarative filtering pipelines for high-throughput sequencing data analysis.
