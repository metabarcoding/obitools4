# Semantic Description of `obiseq` Package Functionalities

This Go package provides **sequence filtering predicates** for biological sequences, integrated with taxonomic validation and hierarchy analysis.

- `IsAValidTaxon(taxonomy, ...bool) SequencePredicate`:  
  Returns a predicate that checks whether a sequence has an associated valid taxon in the given taxonomy.  
  Optionally supports *auto-correction* of outdated/incorrect `taxid` values to match the current taxonomy node.

- `IsSubCladeOf(taxonomy, parent) SequencePredicate`:  
  Filters sequences whose taxonomic assignment is a descendant (sub-clade) of the specified `parent` taxon.

- `IsSubCladeOfSlot(taxonomy, key) SequencePredicate`:  
  Enables filtering based on a *sequence attribute* (e.g., `"taxon"` or `"classification"`) that holds a taxonomic label.  
  Validates the label against the taxonomy, then checks if the sequence’s assigned taxon falls under it.

- `HasRequiredRank(taxonomy, rank) SequencePredicate`:  
  Ensures the sequence’s taxon is assigned at or below a specified rank (e.g., `"species"`, `"genus"`).  
  Validates the requested `rank` against taxonomy’s rank list; exits on invalid input.

All predicates follow a functional, composable design pattern (`SequencePredicate = func(*BioSequence) bool`), enabling flexible pipeline construction (e.g., filtering, classification validation).
