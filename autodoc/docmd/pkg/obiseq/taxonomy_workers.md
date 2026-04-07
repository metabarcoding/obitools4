# Taxonomic Annotation Workers in `obiseq`

This Go package provides functional workers for annotating biological sequences with taxonomic information using a hierarchical taxonomy (e.g., from NCBI or UNITE). Each worker is implemented as a `SeqWorker`—a function that processes one sequence and returns an updated slice of sequences.

- **`MakeSetTaxonAtRankWorker(taxonomy, rank)`**:  
  Assigns a taxonomic label at *a specific rank* (e.g., `"genus"`, `"family"`). Validates that the requested `rank` exists in the taxonomy before proceeding.

- **`MakeSetSpeciesWorker(taxonomy)`**:  
  Annotates each sequence with its inferred species name using the provided taxonomy.

- **`MakeSetGenusWorker(taxonomy)`**:  
  Adds genus-level taxonomic assignment to sequences.

- **`MakeSetFamilyWorker(taxonomy)`**:  
  Adds family-level taxonomic assignment.

- **`MakeSetPathWorker(taxonomy)`**:  
  Populates the full taxonomic path (e.g., `"Eukaryota;Metazoa;Chordata;..."`) for each sequence.

All workers rely on methods of `BioSequence` (e.g., `.SetSpecies()`, `.SetPath()`), which internally use the `obitax.Taxonomy` object to resolve taxonomic IDs or names. Errors are logged via `logrus`; invalid ranks cause a fatal exit.

These utilities support modular, pipeline-friendly taxonomic annotation—ideal for high-throughput metabarcoding workflows.
