# Taxonomic Analysis Functions in `obiseq` Package

This module provides tools for assigning taxonomic labels to biological sequences using a reference taxonomy.

- **`TaxonomicDistribution(taxonomy)`**:  
  Returns a map from taxonomic nodes to read counts, based on `taxid` annotations in the sequence metadata. It validates taxids against the taxonomy and enforces strict handling of aliases.

- **`LCA(taxonomy, threshold)`**:  
  Computes the *Lowest Common Ancestor* (LCA) of all taxonomic assignments for a sequence, weighted by their abundances.  
  - Iteratively traverses upward from each taxon’s path in the taxonomy tree.  
  - At each level, computes the relative weight (`rmax`) of the most frequent taxon.  
  - Stops when `rmax < threshold`, returning:  
    • the LCA taxon,  
    • its confidence score (`rans`), and  
    • total read count used.

- **`AddLCAWorker(...)`**:  
  Creates a `SeqWorker` function to annotate sequences with LCA results:  
    - Sets attributes like `<slot>_taxid`, `<slot>_name`, and `<slot>_error` (rounded to 3 decimals).  
    - Automatically appends `_taxid` if missing in `slot_name`.  

All functions integrate with the OBITools4 ecosystem, supporting robust taxonomic inference for metabarcoding workflows.
