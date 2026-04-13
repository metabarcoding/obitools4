# `obiutils` — Semantic Description of Core Functionality

This Go package provides generic and type-specific utilities for **ranking** and **ordering** data without modifying the original slice. It leverages Go’s `sort` package to compute index permutations that reflect sorted order.

## Key Components

- **IntOrder(data []int) []int**  
  Returns indices that would sort a slice of integers in *ascending* order. The original data remains unchanged.

- **ReverseIntOrder(data []int) []int**  
  Same as `IntOrder`, but returns indices for *descending* order.

- **Order[T sort.Interface](data T) []int**  
  Generic version accepting any type implementing `sort.Interface`. Returns stable sorted indices.

## Internal Design

- **intRanker** and **Ranker[T]**: Helper types wrapping data + index list (`r`).  
  They implement `sort.Interface` *indirectly*—sorting indices instead of mutating data.

- **Index-based sorting**:  
  By permuting a list of indices (`r = [0,1,...]`), the original data is never copied or altered—ideal for large datasets or immutable inputs.

- **Stability**: `Order` uses `sort.Stable`, preserving relative order of equal elements.

## Use Cases

- Sorting metadata (e.g., sorting labels by associated scores).  
- Preparing orderings for downstream operations (plots, ranking metrics).  
- Efficiently tracking original positions after sort.

## Constraints

- Requires `sort.Interface` for generic version (e.g., custom structs with methods).  
- Returns empty slice (`nil`) on zero-length input.
