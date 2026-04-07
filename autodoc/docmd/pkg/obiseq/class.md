# BioSequence Classifier Module Overview  

This Go package (`obiseq`) provides a flexible and thread-safe framework for classifying biological sequences using different strategies. Each classifier implements four core methods:  
- `Code(sequence) int`: assigns an integer class to a sequence.  
- `Value(k) string`: retrieves the original value (or representation) for class index *k*.  
- `Reset()`: clears internal state.  
- `Clone() *BioSequenceClassifier`: creates a fresh copy of the classifier.

## Supported Classifier Types  

1. **`AnnotationClassifier(key, na)`**  
   Classifies sequences based on a single annotation field. Missing annotations default to `na`. Internally maps string values → integer codes via a thread-safe dictionary.

2. **`DualAnnotationClassifier(key1, key2, na)`**  
   Uses *two* annotation fields. Combines them (as JSON array) to form unique class identifiers, enabling multi-dimensional classification.

3. **`PredicateClassifier(predicate)`**  
   Binary classifier: returns `1` if the provided predicate function evaluates to true, else `0`. Useful for rule-based grouping (e.g., length > 200).

4. **`HashClassifier(size)`**  
   Assigns sequences to one of `size` buckets via CRC32 hash of the raw sequence. Deterministic and memory-efficient, but may cause collisions.

5. **`SequenceClassifier()`**  
   Unique class per *exact* sequence string (case-sensitive). Uses a lock-protected map to deduplicate and index sequences.

6. **`RotateClassifier(size)`**  
   Cyclic assignment: sequence *i* → class `i mod size`. No memoization; state resets only manually.

7. **`CompositeClassifier(...)`**  
   Combines multiple classifiers: concatenates their integer outputs (e.g., `"3:17:0"`) to form a composite class key. Enables layered or hierarchical classification.

All classifiers are immutable after creation (state is internal and synchronized), supporting concurrent use in pipelines.
