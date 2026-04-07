# Semantic Description of `GeomIndexSesquence` Function

The function computes a **geometric taxonomic index** for a given query sequence based on spatial proximity and shared taxonomy.

- **Input**:  
  - `seqidx`: index of the reference sequence to analyze.  
  - `references`: list of bio-sequences with geographic coordinates.  
  - `taxa` & `taxo`: taxonomic hierarchy and slice of taxa.

- **Core Logic**:  
  - Retrieves the geographic coordinate (lat/long) of the query sequence. Fails if missing.
  - Computes **Euclidean squared distances** between this coordinate and all others in parallel using goroutines.
  - Sorts sequences by distance via `obiutils.Order`, preserving original indices.

- **Taxonomic Aggregation**:  
  - Starts from the query sequence’s taxon (`lca`).
  - Iterates over increasing distances, updating `lca` to the **Lowest Common Ancestor (LCA)** between current taxon and each neighbor’s.
  - Records, for each distance value encountered, the **current LCA string** (e.g., `"Genus@genus"`).
  - Stops early if the root of the taxonomy is reached.

- **Output**:  
  A map from *distance* (int) → *taxonomic label* (string), encoding how taxonomic resolution degrades with increasing spatial distance.

- **Use Case**:  
  Enables rapid inference of taxonomic uncertainty or confidence bands in ecological or metabarcoding analyses, based on nearest neighbors’ taxonomy and spatial proximity.
