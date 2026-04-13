# K-Way Merge for Sorted k-mer Streams

This Go package implements a **k-way merge** over multiple sorted streams of *k*-mer values (`uint64`). It leverages a **min-heap** to efficiently produce the globally sorted sequence while aggregating duplicate counts across input streams.

## Core Components

- **`mergeItem`**: Stores a value and its source reader index for heap operations.
- **`mergeHeap`** & `heap.Interface`: Implements a min-heap for efficient retrieval of smallest values.
- **`KWayMerge`**: Main struct managing the heap and input readers.

## Key Functionality

- **Initialization (`NewKWayMerge`)**:
  - Takes a slice of `*KdiReader`, each expected to yield sorted values.
  - Preloads the heap with one value from each reader.

- **Streaming Output (`Next`)**:
  - Returns the next smallest *k*-mer, its frequency across readers (i.e., how many input streams contained it), and a success flag.
  - Handles duplicates: pops *all* items equal to the current minimum before advancing readers.

- **Cleanup (`Close`)**:
  - Closes all underlying `KdiReader`s and returns the first encountered error.

## Use Case

Ideal for merging sorted *k*-mer databases (e.g., from multiple files or processes), enabling:
- Efficient deduplication with multiplicity tracking.
- Scalable union/intersection operations on large *k*-mer sets.

## Complexity

| Operation | Time       |
|-----------|------------|
| `Next()`  | *O(log k)* (heap ops per unique value) |
| Init      | *O(k)*     |

Where `k` = number of input readers.
