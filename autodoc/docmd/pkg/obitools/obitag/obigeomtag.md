# Geometric Taxonomic Assignment Module (`obitag`)

This Go package implements a **geometric approach for taxonomic assignment** of biological sequences using landmark-based coordinate mapping and distance minimization.

## Core Functionality

- **Landmark Extraction** (`ExtractLandmarkSeqs`):  
  Retrieves sequences marked with non-default `LandmarkID`s and returns them in ID-indexed order.

- **Taxon Set Extraction** (`ExtractTaxonSet`):  
  Maps each reference sequence to its corresponding taxonomic node using the provided taxonomy; panics on missing taxa.

- **Landmark Coordinate Mapping** (`MapOnLandmarkSequences`):  
  Computes a *coordinate vector* for any query sequence by measuring LCS-based distances to each landmark.

- **Geometric Nearest Neighbor Search** (`FindGeomClosest`):  
  Finds reference sequences with minimal Euclidean distance in landmark space; among ties, selects the one with highest sequence identity (via LCS).

- **Taxonomic Assignment** (`GeomIdentify`):  
  Assigns taxonomy to a query sequence if best identity >50%: uses LCA of matching references’ taxa, weighted by geometric distance. Otherwise assigns root taxon.

- **Worker & CLI Integration** (`GeomIdentifySeqWorker`, `CLIGeomAssignTaxonomy`):  
  Wraps assignment logic into reusable sequence workers and integrates with iterator-based pipelines.

## Key Design Principles

- **Landmark-centric geometry**: Taxonomic inference relies on spatial proximity in landmark-derived feature space.
- **Robustness to alignment ambiguity**: Uses LCS (Longest Common Subsequence) scores instead of full alignments.
- **Parallelization support**: Leverages `obiiter` for scalable batch processing.

## Output Attributes

Each assigned sequence gains metadata:
- `"scientific_name"`, `"obitag_rank"`
- `"obitat_bestid"` (identity), `"obitag_min_dist"`, `"obitag_match_count"`
- `"obitat_coord"` (landmark coordinates), `"obitation_similarity_method": "geometric"`
