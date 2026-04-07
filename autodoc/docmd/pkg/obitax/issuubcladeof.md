# Semantic Description of `obitax` Taxonomic Functions

The `obitax` package provides two core methods for hierarchical taxon relationship analysis:

- **`IsSubCladeOf(parent *Taxon) bool`**  
  Determines whether the current taxon is a **descendant** (i.e., subclade) of a given parent taxon.  
  - Ensures both taxa belong to the *same taxonomy*—fails with a fatal log if not.  
  - Traverses upward via `taxon.IPath()` (iterative ancestor path) to check if any node matches the parent’s ID.  
  - Returns `true` iff a match is found, indicating lineage descent.

- **`IsBelongingSubclades(clades *TaxonSet) bool`**  
  Checks whether the current taxon—or any of its **ancestors**—belongs to a specified set of clades (`TaxonSet`).  
  - Starts by testing direct membership via `clades.Contains(taxon.Node.id)`.  
  - Walks upward through the hierarchy (`taxon = taxon.Parent()`) until either:  
    - A match is found, or  
    - The root is reached.  
  - Final check at the root ensures completeness (e.g., if only root belongs).  

Both functions support **robust phylogenetic queries**, enabling classification validation, filtering by clade membership, and hierarchical consistency checks in taxonomic trees.
