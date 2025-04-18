# OBITools release notes

## New changes

### Bug fixes

- In `obipairing` correct the misspelling of the `obiparing_*` tags where the `i`
  was missing to `obipairing_`.  

- In `obigrep` the **-C** option that excludes sequences too abundant was not 
  functional.

- In `obitaxonomy` the **-l** option that lists all the taxonomic rank defined by
  a taxonomy was not functional

- The file type guesser was not using enough data to be able to correctly detect
  file format when sequences were too long in fastq and fasta or when lines were
  to long in CSV files. That's now corrected

- Options **--fasta** or **--fastq** usable to specify input format were ignored.
  They are now correctly considered

- The `obiannotate` command were crashing when a selection option was used but
  no editing option.
  
- The `--fail-on-taxonomy` led to an error on merged taxa even when the
  `--update-taxid` option was used.

- The `--compressed` option was not correctly named. It was renamed to `--compress`
  
### Enhancement

- Some sequences in the Genbank and EMBL databases are several gigabases long. The 
  sequence parser had to reallocate and recopy memory many times to read them, 
  resulting in a complexity of O(N^2) for reading such large sequences. 
  The new file chunk reader has a linear algorithm that speeds up the reading 
  of very long sequences.

- A new option **--csv** is added to every obitools to indicate that the input
  format is CSV

- The new version of obitools are now printing the taxids in a fancy way
  including the scientific name and the taxonomic rank (`"taxon:9606 [Homo
  sapiens]@species"`). But if you need the old fashion raw taxid, a new option
  **--raw-taxid** has been added to get obitools printing the taxids without any
  decorations (`"9606"`).


## March 1st, 2025. Release 4.4.0

A new documentation website is available at https://obitools4.metabarcoding.org.
Its development is still in progress. 

The biggest step forward in this new version is taxonomy management. The new
version is now able to handle taxonomic identifiers that are not just integer
values. This is a first step towards an easy way to handle other taxonomy
databases soon, such as the GBIF or Catalog of Life taxonomies. This version
is able to handle files containing taxonomic information created by previous
versions of OBITools, but files created by this new version may have some
problems to be analyzed by previous versions, at least for the taxonomic
information.


### Breaking changes

- In `obimultiplex`, the short version of the **--tag-list** option used to
  specify the list of tags and primers to be used for the demultiplexing has
  been changed from `-t` to `-s`.

- The command `obifind` is now renamed `obitaxonomy`.

- The **--taxdump** option used to specify the path to the taxdump containing
  the NCBI taxonomy has been renamed to **--taxonomy**.

### Bug fixes

- Correction of a bug when using paired sequence file with the **--out** option.

- Correction of a bug in `obitag` when trying to annotate very short sequences of
  4 bases or less.
  

- In `obipairing`, correct the stats `seq_a_single` and `seq_b_single` when
  on right alignment mode

- Not really a bug but the memory impact of `obiuniq` has been reduced by reducing
  the batch size and not reading the qualities from the fastq files as `obiuniq`
  is producing only fasta output without qualities.

-   In `obitag`, correct the wrong assignment of the **obitag_bestmatch**
    attribute.

-   In `obiclean`, the **--no-progress-bar** option disables all progress bars,
    not just the data.

-   Several fixes in reading FASTA and FASTQ files, including some code
    simplification and factorization.

-   Fixed a bug in all obitools that caused the same file to be processed
    multiple times, when specifying a directory name as input.


### New features

-   `obigrep` add a new **--valid-taxid** option to keep only sequence with a
    valid taxid

-   `obiclean` add a new **--min-sample-count** option with a default value of 1,
    asking to filter out sequences which are not occurring in at least the
    specified number of samples.

-   `obitoaxonomy` a new **--dump|D** option allows for dumping a sub-taxonomy.
  
-   Taxonomy dump can now be provided as a four-columns CSV file to the
    **--taxonomy** option.

-   NCBI Taxonomy dump does not need to be uncompressed and unarchived anymore. The
    path of the tar and gziped dump file can be directly specified using the
    **--taxonomy** option.

-   Most of the time obitools identify automatically sequence file format. But
    it fails sometimes. Two new option **--fasta** and **--fastq** are added to
    allow the processing of the rare fasta and fastq files not recognized.
    
-   In `obiscript`, adds new methods to the Lua sequence object:
    - `md5_string()`: returning the MD5 check sum as a hexadecimal string, 
    - `subsequence(from,to)`: allows extracting a subsequence on a 0 based
      coordinate system, upper bound excluded like in go. 
    - `reverse_complement`: returning a sequence object corresponding to the
      reverse complement of the current sequence.

### Enhancement

-   All obitools now have a **--taxonomy** option. If specified, the taxonomy is
    loaded first and taxids annotating the sequences are validated against that
    taxonomy. A warning is issued for any invalid taxid and for any taxid that
    is transferred to a new taxid. The **--update-taxid** option allows these
    old taxids to be replaced with their new equivalent in the result of the
    obitools command.

-   The scoring system used by the `obipairing` command has been changed to be
    more coherent. In the new version, the scores associated to a match and a
    mismatch involving a nucleotide with a quality score of 0 are equal. Which
    is normal as a zero quality score means a perfect indecision on the read
    nucleotide, therefore there is no reason to penalize a match differently
    from a mismatch (see 
    https://obitools4.metabarcoding.org/docs/commands/alignments/obipairing/exact-alignment/).

-   In every *OBITools* command, the progress bar is automatically deactivated
    when the standard error output is redirected.

-   Because Genbank and ENA:EMBL contain very large sequences, while OBITools4
    are optimized As Genbank and ENA:EMBL contain very large sequences, while
    OBITools4 is optimized for short sequences, `obipcr` faces some problems
    with excessive consumption of computer resources, especially memory. Several
    improvements in the tuning of the default `obipcr` parameters and some new
    features, currently only available for FASTA and FASTQ file readers, have
    been implemented to limit the memory impact of `obipcr` without changing the
    computational efficiency too much.

-   Logging system and therefore format, have been homogenized.

## August 2nd, 2024. Release 4.3.0

### Change of git repository

-   The OBITools4 git repository has been moved to the GitHub repository. 
    The new address is: https://github.com/metabarcoding/obitools4.
    Take care for using the new install script for retrieving the new version.

    ```bash
    curl -L https://metabarcoding.org/obitools4/install.sh \
      | bash
    ```

    or with options:

    ```bash
    curl -L https://metabarcoding.org/obitools4/install.sh \
      | bash -s -- --install-dir test_install --obitools-prefix k
    ```
-   The output of the obitools will evolve to produce results only in standard
    formats such as fasta and fastq. For non-sequential data, the output will be
    in CSV format, with the separator `,`, the decimal separator `.`, and a
    header line with the column names. It is more convenient to use the output
    in other programs. For example, you can use the `csvtomd` command to
    reformat the CSV output into a Markdown table. The first command to initiate
    this change is `obicount`, which now produces a 3-line CSV output.

    ```bash
    obicount data.csv | csvtomd 
    ```

-   Adds the new experimental `obicleandb` utility to clean up reference
    database files created with `obipcr`. An easy way to create a reference
    database for `obitag` is to use `obipcr` on a local copy of Genbank or EMBL.
    However, these sequence databases are known to contain many taxonomic
    errors, such as bacterial sequences annotated with the taxid of their host
    species. `obicleandb` tries to detect these errors. To do this, it first keeps
    only sequences annotated with the taxid to which a species, genus, and
    family taxid can be assigned. Then, for each sequence, it compares the
    distance of the sequence to the other sequences belonging to the same genus
    to the same number of distances between the considered sequence and a
    randomly selected set of sequences belonging to another family using a
    Mann-Whitney U test. The alternative hypothesis is that out-of-family
    distances are greater than intrageneric distances. Sequences are annotated
    with the p-value of the Mann-Whitney U test in the **obicleandb_trusted**
    slot. Later, the distribution of this p-value can be analyzed to determine a
    threshold. Empirically, a threshold of 0.05 is a good compromise and allows
    filtering out less than 1‰ of the sequences. These sequences can then be
    removed using `obigrep`.

-   Adds a new `obijoin` utility to join information contained in a sequence
    file with that contained in another sequence or CSV file. The command allows
    you to specify the names of the keys in the main sequence file and in the
    secondary data file that will be used to perform the join operation.

-   Adds a new tool `obidemerge` to demerge a `merge_xxx` slot by recreating the 
    multiple identical sequences having the slot `xxx` recreated with its initial
    value and the sequence count set to the number of occurrences referred in the
    `merge_xxx` slot. During the operation, the `merge_xxx` slot is removed.

-   Adds CSV as one of the input format for every obitools command. To encode
    sequence the CSV file must include a column named `sequence` and another
    column named `id`. An extra column named `qualities` can be added to specify 
    the quality scores of the sequence following the same ASCII encoding than the
    fastq format. All the other columns will be considered as annotations and will
    be interpreted as JSON objects encoding potentially for atomic values. If a 
    column value can not be decoded as JSON it will be considered as a string.

-   A new option **--version** has been added to every obitools command. It will
    print the version of the command.

-   In `obiscript` a `qualities` method has been added to retrieve or set the
    quality scores from a BioSequence object.\

-   In `obimultuplex` the ngsfilter file describing the samples can be no provided
    not only using the classical ngsfilter format but also using the CSV format.
    When using CSV, the first line must contain the column names. 5 columns are
    expected:

    -   `experiment` the name of the experiment
    -   `sample` the name of the sample
    -   `sample_tag` the tag used to identify the sample
    -   `forward_primer` the forward primer sequence
    -   `reverse_primer` the reverse primer sequence
   
    The order of the columns is not important, as long as they are present and
    named correctly. The `obiparing` command will print an error message if 
    some column is missing. It now includes a **--template ** option that can
    be used to create an example CSV file.

    Supplementary columns are allowed. Their names and content will be used to
    annotate the sequence corresponding to the sample, as the `key=value;` did
    in the ngsfilter format.

    The CSV format used allows for comment lines starting with `#` character.
    Special data lines starting with `@param` in the first column allow configuring the algorithm. The options **--template** provided an over
    commented example of the CSV format, including all the possible options.
    
### CPU limitation

-   By default, *OBITools4* tries to use all the computing power available on
    your computer. In some circumstances this can be problematic (e.g. if you
    are running on a computer cluster managed by your university). You can limit
    the number of CPU cores used by *OBITools4* or by using the **--max-cpu**
    option or by setting the **OBIMAXCPU** environment variable. Some strange
    behavior of *OBITools4* has been observed when users try to limit the
    maximum number of usable CPU cores to one. This seems to be caused by the Go
    language, and it is not obvious to get *OBITools4* to run correctly on a
    single core in all circumstances. Therefore, if you ask to use a single
    core, **OBITools4** will print a warning message and actually set this
    parameter to two cores. If you really want a single core, you can use the
    **--force-one-core** option. But be aware that this can lead to incorrect
    calculations.


## April 2nd, 2024. Release 4.2.0

### New features

-   A new OBITools named `obiscript` allows processing each sequence according
    to a Lua script. This is an experimental tool. The **--template** option
    allows for generating an example script on the `stdout`.

### API Changes

-   Two of the main class `obiseq.SeqWorker` and `obiseq.SeqWorker` have their
    declaration changed. Both now return two values a `obiseq.BioSequenceSlice`
    and an `error`. This allows a worker to return potentially several sequences
    as the result of the processing of a single sequence, or zero, which is
    equivalent to filter out the input sequence.

### Enhancement

-   In `obitag` if the reference database contains sequences annotated by taxid
    not referenced in the taxonomy, the corresponding sequences are discarded
    from the reference database and a warning indicating the sequence *id* and the
    wrong taxid is emitted.
-   The bug corrected in the parsing of EMBL and Genbank files as implemented in
    version 4.1.2 of OBITools4, potentially induced some reduction in the
    performance of the parsing. This should have been now fixed.
-   In the same idea, parsing of Genbank and EMBL files were reading and storing
    in memory not only the sequence but also the annotations (features table).
    Up to now none of the OBITools are using this information, but with large
    complete genomes, it is occupying a lot of memory. To reduce this impact,
    the new version of the parser doesn't any more store in memory the
    annotations by default.
-   Add a **--taxonomic-path** to `obiannotate`. The option adds a
    `taxonomic_path` tag to sequences describing the taxonomic classification of
    the sequence according to its taxid. The path is a string. Each level of the
    path is delimited by a `|` character. A level consists of three parts
    separated by a `@`. The first part is the taxid, the second the scientific
    name and the last the taxonomic rank. The first level described is always
    the root of the taxonomy. The latest corresponds to the taxid of the
    sequence. If a sequence is not annotated by a taxid, as usual the sequence
    is assumed having the taxid 1 (the root of the taxonomy).

### Bug fixes

-   Fix a bug in the parsing of the JSON header of FASTA and FASTQ files
    occurring when a string includes a curly brace.
-   Fix a bug in the function looking for the closest match in `obitag`. This
    error led to some wrong taxonomic assignment.
-   Fix a bug in the writing of the fastq files, when quality of a nucleotide
    was not in the range 0-93.

## February 16th, 2024. Release 4.1.2

### Bug fixes

-   Several bugs in the parsing of EMBL and Genbank files have been fixed. The
    bugs occurred in the case of very large sequences, such as complete genomes.
    The Genbank parser is now more robust. It breaks for more errors than the
    previous version. This allows to detect parsing errors instead of hiding
    them and producing wrong results.

## December 20th, 2023. Release 4.1.1

### New feature

-   In `obimatrix` a **--transpose** option allows transposing the produced
    matrix table in CSV format.
-   In `obitpairing` and `obipcrtag` two new options **--exact-mode** and
    **--fast-absolute** to control the heuristic used in the alignment
    procedure. **--exact-mode** allows for disconnecting the heuristic and run
    the exact algorithm at the cost of a speed. **--fast-absolute** change the
    scoring schema of the heuristic.
-   In `obiannotate` adds the possibility to annotate the first match of a
    pattern using the same algorithm as the one used in `obipcr` and
    `obimultiplex`. For that four option were added :
    -   **--pattern** : to specify the pattern. It can use IUPAC codes and
        position with no error tolerated has to be followed by a `#` character.
    -   **--pattern-name** : To specify the names of the slot used to report the
        results. Default is *pattern*
    -   **--pattern-error** : To specify the maximum number of error tolerated
        during matching process.
    -   **--allows-indels** : By default considered errors are mismatched, this
        flag allows for indels. Only the first match is reported if several
        occurrences exist. If no match is found on direct strand then pattern is
        looked for on the reverse complemented strand of the sequence.

### Enhancement

-   For efficiency purposes, now the `obiuniq` command run on disk by default.
    Consequently, the **--on-disk** option has been replaced by **--in-memory**
    to ask explicitly to use memory.
-   Adds an option **--penalty-scale** to the `obipairing` and `obipcrtag`
    command to fine tune the pairing score in the system of the alignment
    procedure by applying a scaling factor to the mismatch score and the gap
    score relatively to the match score.

### Bug fixes

-   In `obicsv`, the **--keep count** was not equivalent to **--count**.
-   In `obipairing` and `obipcrtag`, correct a bug in the alignment procedure
    leading to negative scores.
-   In `obimultiplex`, correct a bug leading to a miss-read of the ngsfilter
    file when tags where written in lower case.
-   In `obitag`, correct a bug leading to the annotation by taxid 1 (root) all
    the sequences having a 100% match with one the reference sequence.
-   Correct a bug in the EMBL reader.

## November 16th, 2023. Release 4.1.0

### New feature

-   In the OBITools language a new `gc` computes the GC fraction of a sequence.
-   First version of the `obisummary` command. It produces summary statistics of
    the sequence file provided as input. The statistics includes, the number of
    reads, of variants, the total length of the DNA sequences (equivalent to
    `obicount`), some summaries about tags used in the sequence annotations and
    their frequencies of usage.
-   First version of the `obimatrix` command. It allows producing OTU tables
    from sequence files in CSV format.
-   The `obicsv` command has now a **--auto** option, that extract automatically
    the attributes present in a file for inspecting the beginning of the
    sequence file. Only attributes that do not correspond to map are reported.
    To extract information from map attributes, see the `obimatrix` command.

### Enhancement

-   A new completely rewritten GO version of the fastq and fasta parser is now
    used instead of the original C version.
-   A new file format guesser is now implemented. This is a first step towards
    allowing new formats to be managed by OBITools.
-   New way of handling header definitions of fasta and fastq formats with JSON
    headers. The sequence definition is now printed in new files as an attribute
    of the JSON header named "definition". That's facilitates the writing of
    parsers for the sequence headers.
-   The -D (--delta) option has been added to `obipcr`. It allows extracting
    flanking sequences of the barcode.
    -   If -D is not set, the output sequence is the barcode itself without the
        priming sites.
    -   If -D is set to 0, the output sequence is the barcode with the priming
        sites.
    -   When -D is set to \### (where \### is an integer), the output sequence
        is the barcode with the priming sites,\
        and \### base pairs of flanking sequences.
-   A new output format in JSON is proposed using the **--json-output**. The
    sequence file is printed as a JSON vector, where each element is a map
    corresponding to a sequence. The map has at most four elements:
    -   *"id"* : which is the only mandatory element (string)
    -   *"sequence"* : if sequence data is present in the record (string)
    -   *"qualities"* : if quality data is associated to the record (string)
    -   *"annotations"* : annotations is associated to the record (a map of
        annotations).

### Bugs

-   In the obitools language, the `composition` function now returns a map
    indexed by lowercase string "a", "c", "g", "t" and "o" for other instead of
    being indexed by the ASCII codes of the corresponding letters.
-   Correction of the reverse-complement operation. Every reverse complement of
    the DNA sequence follow now the following rules :
    -   Nucleotide codes are complemented to their lower complementary base
    -   `.` and `-` characters are returned without change
    -   `[` is complemented to `]` and oppositely
    -   all other characters are complemented as `n`
-   Correction of a bug is the `Subsequence` method of the `BioSequence` class,
    duplicating the quality values. This made `obimultiplex` to produce fastq
    files with sequences having quality values duplicated.

### Be careful

GO 1.21.0 is out, and it includes new functionalities which are used in the
OBITools4 code. If you use the recommended method for compiling OBITools on your
computer, there is no problem, as the script always load the latest GO version.
If you rely on your personal GO install, please think to update.

## August 29th, 2023. Release 4.0.5

### Bugs

-   Patch a bug in the `obiseq.BioSequence` constructor leading to an error on
    almost every obitools. The error message indicates : `fatal error: sync:
    unlock of unlocked mutex` This bug was introduced in the release 4.0.4

## August 27th, 2023. Release 4.0.4

### Bugs

-   Patch a bug in the install-script for correctly follow download redirection.
-   Patch a bug in `obitagpcr` to consider the renaming of the
    `forward_mismatch` and `reverse_mismatch` tags to `forward_error` and
    `reverse_error`.

### Enhancement

-   Comparison algorithms in `obitag` and `obirefidx` take more advantage of the
    data structure to limit the number of alignments actually computed. This
    increase a bit the speed of both the software. `obirefidx` is nevertheless
    still too slow compared to my expectation.
-   Switch to a parallel version of the GZIP library, allowing for high speed
    compress and decompress operation on files.

### New feature

-   In every *OBITools*, writing an empty sequence (sequence of length equal to
    zero) through an error and stops the execution of the tool, except if the
    **--skip-empty** option is set. In that case, the empty sequence is ignored
    and not printed to the output. When output involved paired sequence the
    **--skip-empty** option is ignored.
-   In `obiannotate` adds the **--set-identifier** option to edit the sequence
    identifier
-   In `obitag` adds the **--save-db** option allowing at the end of the run of
    `obitag` to save a modified version of the reference database containing the
    computed index. This allows next time using this partially indexed reference
    library to accelerate the taxonomic annotations.
-   Adding of the function `gsub` to the expression language for substituting
    string pattern.

## May 2nd, 2023. Release 4.0.3

### New features

-   Adding of the function `contains` to the expression language for testing if
    a map contains a key. It can be used from `obibrep` to select only sequences
    occurring in a given sample :

    ```{bash}
    obigrep -p 'contains(annotations.merged_sample,"15a_F730814")' wolf_new_tag.fasta
    ```

-   Adding of a new command `obipcrtag`. It tags raw Illumina reads with the
    identifier of their corresponding sample. The tags added are the same as
    those added by `obimultiplex`. The produced forward and reverse files can
    then be split into different files using the `obidistribute` command.

    ```{bash}
    obitagpcr -F library_R1.fastq \
              -R library_R2.fastq \
              -t sample_ngsfilter.txt \
              --out tagged_library.fastq \
              --unidentified not_assigned.fastq
    ```

    The command produced four files : `tagged_library_R1.fastq` and
    `tagged_library_R2.fastq` containing the assigned reads and
    `not_assigned_R1.fastq` and `not_assigned_R2.fastq` containing the
    unassignable reads.

    The tagged library files can then be split using `obidistribute`:

    ```{bash}
    mkdir pcr_reads
    obidistribute --pattern "pcr_reads/sample_%s_R1.fastq" -c sample tagged_library_R1.fastq
    obidistribute --pattern "pcr_reads/sample_%s_R2.fastq" -c sample tagged_library_R2.fastq
    ```

-   Adding of two options **--add-lca-in** and **--lca-error** to `obiannotate`.
    These options aim to help during construction of reference database using
    `obipcr`. On `obipcr` output, it is commonly run `obiuniq`. To merge identical
    sequences annotated with different taxids, it is now possible to use the
    following strategies :

    ```{bash}
    obiuniq -m taxid myrefdb.obipcr.fasta \
    | obiannotate -t taxdump --lca-error 0.05 --add-lca-in taxid \
    > myrefdb.obipcr.unique.fasta
    ```

    The `obiuniq` call merge identical sequences keeping track of the diversity
    of the taxonomic annotations in the `merged_taxid` slot, while `obiannotate`
    loads a NCBI taxdump and computes the lowest common ancestor of the taxids
    represented in `merged_taxid`. By specifying **--lca-error** 0.05, we
    indicate that we allow for at most 5% of the taxids disagreeing with the
    computed LCA. The computed LCA is stored in the slot specified as a
    parameter of the option **--add-lca-in**. Scientific name and actual error
    rate corresponding to the estimated LCA are also stored in the sequence
    annotation.

### Enhancement

-   Rename the `forward_mismatches` and `reverse_mismatches` from instanced by
    `obimutiplex` into `forward_error` and `reverse_error` to be coherent with
    the tags instanced by `obipcr`

### Corrected bugs

-   Correction of a bug in memory management and Slice recycling.
-   Correction of the **--fragmented** option help and logging information
-   Correction of a bug in `obiconsensus` leading into the deletion of a base
    close to the beginning of the consensus sequence.

## March 31st, 2023. Release 4.0.2

### Compiler change

*OBItools4* requires now GO 1.20 to compile.

### New features

-   Add the possibility for looking pattern with indels. This has been added to
    `obimultiplex` through the **--with-indels** option.
-   Every obitools command has a **--pprof** option making the command
    publishing a profiling website available at the address :
    <http://localhost:8080/debug/pprof/>
-   A new `obiconsensus` command has been added. It is a prototype. It aims to
    build a consensus sequence from a set of reads. The consensus is estimated
    for all the sequences contained in the input file. If several input files,
    or a directory name are provided the result contains a consensus per file.
    The *id* of the sequence is the name of the input file depleted of its
    directory name and of all its extensions.
-   In `obipcr` an experimental option **--fragmented** allows for splitting very
    long query sequences into shorter fragments with an overlap between the two
    contiguous fragment insuring that no amplicons are missed despite the split.
    As a site effect some amplicon can be identified twice.
-   In `obipcr` the -L option is now mandatory.

### Enhancement

-   Add support for IUPAC DNA code into the DNA sequence LCS computation and an
    end free gap mode. This impact `obitag` and `obimultiplex` in the
    **--with-indels** mode.
-   Print the synopsis of the command when an error is done by the user at
    typing the command
-   Reduced the memory copy and allocation during the sequence creation.

### Corrected bugs

-   Better management of non-existing files. The produced error message is not
    yet perfectly clear.
-   Patch a bug leading with some programs to crash because of : "*empty batch
    pushed on the channel*"
-   Patch a bug when directory names are used as input data name preventing the
    system to actually analyze the collected files.
-   Make the **--help** or **-h** options working when mandatory options are
    declared
-   In `obimultiplex` correct a bug leading to a wrong report of the count of
    reverse mismatch for sequences in reverse direction.
-   In `obimultiplex` correct a bug when not enough space exist between the
    extremities of the sequence and the primer matches to fit the sample
    identification tag
-   In `obipcr` correction of bug leading to miss some amplicons when several
    amplicons are present on the same large sequence.

## March 7th, 2023. Release 4.0.1

### Corrected bugs

-   Makes progress bar updating at most 10 times per second.
-   Makes the command exiting on error if undefined options are used.

### Enhancement

-   *OBITools* are automatically processing all the sequences files contained in
    a directory and its subdirectory\
    recursively if its name is provided as input. To process easily Genbank
    files, the corresponding filename extensions have been added. Today the
    following extensions are recognized as sequence files : `.fasta`, `.fastq`,
    `.seq`, `.gb`, `.dat`, and `.ecopcr`. The corresponding gziped version are
    also recognized (e.g. `.fasta.gz`)

### New features

-   Takes into account the `OBIMAXCPU` environmental variable to limit the
    number of CPU cores used by OBITools in bash the below command will limit to
    4 cores the usage of OBITools

    ``` bash
    export OBICPUMAX=4
    ```

-   Adds a new option --out\|-o allowing to specify the name of an output file.

    ``` bash
    obiconvert -o xyz.fasta xxx.fastq
    ```

    is thus equivalent to

    ``` bash
    obiconvert  xxx.fastq > xyz.fasta
    ```

    That option is actually mainly useful for dealing with paired reads sequence
    files.

-   Some OBITools (now `obigrep` and `obiconvert`) are capable of using paired
    read files. Options have been added for this (**--paired-with** *FILENAME*,
    and **--paired-mode** *forward\|reverse\|and\|andnot\|xor*). This, in
    combination with the **--out** option shown above, ensures that the two
    matched files remain consistent when processed.

-   Adding of the function `ifelse` to the expression language for computing
    conditional values.

-   Adding two function to the expression language related to sequence
    composition : `composition` and `gcskew`. Both are taking a sequence as
    single argument.

## February 18th, 2023. Release 4.0.0

It is the first version of the *OBITools* version 4. I decided to tag then
following two weeks of intensive data analysis with them allowing to discover
many small bugs present in the previous non-official version. Obviously other
bugs are certainly present in the code, and you are welcome to use the git
ticket system to mention them. But they seem to produce now reliable results.

### Corrected bugs

-   On some computers the end of the output file was lost, leading to the loose
    of sequences and to the production of incorrect file because of the last
    sequence record, sometime truncated in its middle. This was only occurring
    when more than a single CPU was used. It was affecting every obitools.
-   The `obiparing` software had a bug in the right alignment procedure. This led
    to the non-alignment of very sort barcode during the paring of the forward
    and reverse reads.
-   The `obipairing` tools had a non-deterministic comportment when aligning a
    pair very low quality reads. This induced that the result of the same low
    quality read pair was not the same from run to run.

### New features

-   Adding of a `--compress|-Z` option to every obitools allowing to produce
    `gz` compressed output. OBITools were already able to deal with gziped input
    files transparently. They can now produce their results in the same format.
    - Adding of a `--append|-A` option to the `obidistribute` tool. It allows appending the result of an `obidistribute` execution to preexisting files. -
    Adding of a `--directory|-d` option to the `obidistribute` tool. It allows
    declaring a secondary classification key over the one defined by the
    `--category\|-c\` option. This extra key leads to produce directories in
    which files produced according to the primary criterion are stored.
-   Adding of the functions `subspc`, `printf`, `int`, `numeric`, and `bool` to
    the expression language.