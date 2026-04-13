# `obiseq` Package: Sequence Concatenation via `.Join()`

The `BioSequence.Join()` method enables semantic concatenation of two biological sequences (e.g., DNA, RNA, or protein strings).  

- **Signature**:  
  ```go
  func (sequence *BioSequence) Join(seq2 *BioSequence, inplace bool) *BioSequence
  ```

- **Purpose**:  
  Combines the current sequence (`sequence`) with a second one (`seq2`), returning a new or modified `BioSequence`.

- **Parameters**:  
  - `seq2`: The sequence to append. Must be a valid `*BioSequence`.  
  - `inplace`: Boolean flag: if `true`, modifies the receiver in-place; otherwise, operates on a copy.

- **Semantics**:  
  - If `inplace == false`, the method first creates a deep copy of the original sequence to avoid side effects.  
  - It then appends `seq2.Sequence()` (the underlying string/byte representation) to the target sequence using an internal `.Write()` method.  
  - The final concatenated result is returned as a `*BioSequence`.

- **Behavioral Guarantees**:  
  - *Pure operation*: When `inplace = false`, the original sequences remain unaltered.  
  - *Chaining-friendly*: Returns a pointer, enabling method chaining (e.g., `seq.Join(a, false).Join(b, true)`).

- **Use Cases**:  
  - Building multi-domain proteins or gene fusions.  
  - Merging fragments from sequencing reads.  
  - Constructing synthetic constructs in silico.

- **Assumptions**:  
  - `BioSequence.Sequence()` returns a valid string/byte slice.  
  - `.Write(...)` handles appending correctly (e.g., no validation of biological compatibility — e.g., frame shifts are not checked).  

This method supports flexible, functional-style sequence manipulation while preserving memory safety via optional in-place mutation.
