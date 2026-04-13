# `obistats.SampleIntWithoutReplacement` — Semantic Description

The function **`SampleIntWithoutReplacement(n, max int) []int`** implements a *random sampling without replacement* algorithm over the integer range `[0, max)`.

## Core Purpose  
Generates **`n` distinct integers**, uniformly at random and *without repetition*, from the interval `[0, max)`.

## Algorithmic Strategy  
Uses an **incremental reservoir-like mapping** (`draw map[int]int`) to maintain uniqueness:
- Iteratively draws `y = rand.Intn(max)` (i.e., uniform in `[0, max)`).
- If `y` is already present (`ok = true`), it retrieves and reuses the stored value (a *swap trick*).
- Then, `draw[y]` is set to the current upper bound (`max - 1`) and `max` decremented — effectively *removing* one value from the future draw space.
- This preserves uniformity while avoiding collisions, in **O(n)** time and memory.

## Key Properties  
- ✅ Guarantees uniqueness: no duplicates in the returned slice.  
- ⚖️ Uniform distribution over all possible `n`-element subsets of `[0, max)`.  
- 🧠 Space-efficient: uses a map (O(n)) instead of shuffling an array of size `max`.  
- 🚀 Efficient for large `max` and moderate `n`, where full-shuffle methods would be wasteful.

## Return Value  
A slice of length `n`, containing the sampled integers (order is *not* sorted or deterministic — reflects insertion order in `draw`).

## Typical Use Cases  
- Random subset selection (e.g., cross-validation folds, bootstrapping indices).  
- shuffling without full permutation.  
- Monte Carlo simulations requiring unique random IDs or positions.

## Limitations / Notes  
- Assumes `0 ≤ n ≤ max`; behavior is undefined otherwise.  
- Relies on the global `math/rand` source (not seeded here); users should call `rand.Seed()` if reproducibility is needed.
