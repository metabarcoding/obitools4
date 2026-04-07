# `MinMultiset[T]` — A Lazy-Delete Min-Multiset Implementation

A generic, type-safe multiset data structure in Go that maintains elements with multiplicity and provides efficient access to the current minimum. Built on top of a min-heap (`container/heap`) with **lazy deletion** to support efficient removals without rebuilding the heap.

## Core Features

- ✅ **Generic over comparable types** (`T`) with custom ordering via `less` comparator  
- ✅ **Multiset semantics**: supports multiple occurrences of the same value  
- ✅ **O(log n) insertion** (`Add`) and **amortized O(1)** minimum access  
- ✅ **Lazy deletion**: `RemoveOne` marks items for removal; physical cleanup occurs on next `Min()` call  
- ✅ **Size tracking**: logical size (`Len()`) excludes deleted items, even if still in heap  
- ✅ **Memory-efficient cleanup**: `shrink()` and `cleanTop()` prevent tombstone accumulation  

## API Summary

| Method | Description |
|--------|-------------|
| `NewMinMultiset(less)` | Constructor; initializes heap, maps (`count`, `pending`), and sets ordering |
| `Add(v)` | Inserts one occurrence of `v`; increments logical size & count map |
| `RemoveOne(v)` | Removes *one* occurrence if present; returns success flag (`false` otherwise) |
| `Min()` | Returns current minimum (or zero value + `ok=false`) after cleaning stale top entries |
| `Len()` | Returns logical size (excludes pending deletions) |

## Internal Mechanism

- **`count[T]int`**: tracks how many times each value is *logically* present  
- **`pending[T]int`**: tracks how many times each value is *marked for removal*  
- **Heap invariant maintained only up to logical size** — stale entries are pruned lazily during `Min()` or after deletions  
- **No manual cleanup needed** — the structure self-balances incrementally  

## Use Cases

Priority queues with deletable arbitrary elements (e.g., Dijkstra’s algorithm where distances are updated), sliding-window minima, event scheduling with cancellation.

> ⚠️ Note: `less` must define a *strict total order* (transitive, antisymmetric, connected) for correctness.
