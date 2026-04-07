# Semantic Description of `obialign.ReadAlign`

The `ReadAlign` function performs **paired-end read alignment** with quality-aware scoring, optimized for overlapping consensus construction in NGS data processing.

## Core Functionality

- **Input**: Two biological sequences (`seqA`, `seqB`) as `BioSequence` objects, plus alignment parameters:  
  - `gap`: gap penalty (linear)  
  - `scale`: scaling factor for quality scores  
  - `delta`: extension buffer around initial overlap estimate  
  - `fastScoreRel`: use relative vs absolute k-mer matching score  

## Algorithm Overview

1. **Preprocessing & Initialization**  
   - Ensures DNA scoring matrix is initialized (`_InitDNAScoreMatrix`).  

2. **Fast Overlap Estimation via 4-mer Indexing**  
   - Builds a k-mer index of `seqA` using `obikmer.Index4mer`.  
   - Computes optimal shift via `_FastShiftFourMer` in both forward and reverse-complement orientations.  
   - Selects orientation (direct or reversed) yielding highest k-mer match count (`fastCount`) and score (`fastScore`).  

3. **Overlap Computation**  
   - Determines overlap length `over` based on shift:  
     ```text
       over = |seqA| - shift    if shift > 0  
              |seqB| + shift    if shift < 0  
              min(|seqA|,|seqB)| otherwise
     ```

4. **Dynamic Programming Alignment**  
   - If overlap is *not* identical (`fastCount + 3 < over`):  
     - Extracts subregions with `delta`-buffered boundaries.  
     - Calls either `_FillMatrixPeLeftAlign` (left-aligned case) or `_FillMatrixPERightAlign`.  
     - Backtracks via `_Backtracking` to produce alignment path.  
   - Else (near-perfect overlap):  
     - Skips DP; computes score directly from quality scores using `_NucScorePartMatchMatch`.  
     - Returns trivial path `[extra5, partLen]`.

## Output

Returns:  

| Index | Type     | Meaning |
|-------|----------|---------|
| 0️⃣    | `int`     | Final alignment score (weighted by quality) |
| 1️⃣    | `[]int`   | Alignment path (list of positions: `[startA, endA, startB, endB]` or similar) |
| 2️⃣    | `int`     | K-mer match count (`fastCount`) |
| 3️⃣    | `int`     | Overlap length (`over`) |
| 4️⃣    | `float64` | K-mer-based score (`fastScore`) |
| 5️⃣    | `bool`    | Whether alignment was performed in direct orientation (`true`) or on reverse-complement of `seqB` |

## Key Design Highlights

- **Efficient pre-filtering** using 4-mers avoids full DP for nearly identical reads.  
- **Quality-aware scoring**, leveraging Phred scores via `_NucScorePartMatchMatch`.  
- Supports **asymmetric overlaps** (left/right alignment) with boundary padding (`delta`).  
- Uses preallocated memory arenas to minimize GC pressure in high-throughput pipelines.
