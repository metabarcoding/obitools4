# OBITools release notes

## Latest changes

### Bugs

- Patch a bug in the install-script for correctly follow download redirection.
- Patch a bug in `obitagpcr` to consider the renaming of the `forward_mismatch` and `reverse_mismatch` tags
  to `forward_error` and `reverse_error`.

### Enhancement

- Comparison algorithms in `obitag` and `obirefidx` take more advantage of the data structure to limit the number of alignments
  actually computed. This increase a bit the speed of both the software. `obirefidx` is nevertheless still too slow
  compared to my expectation.

### New feature

- In every *OBITools*, writing an empty sequence (sequence of length equal to zero) through an error and
  stops the execution of the tool, except if the **--skip-empty** option is set. In that case, the empty
  sequence is ignored and not printed to the output. When output involved paired sequence the **--skip-empty** 
  option is ignored. 
- In `obiannotate` adds the **--set-identifier** option to edit the sequence identifier
- In `obitag` adds the **--save-db** option allowing at the end of the run of `obitag` to save a
  modified version of the reference database containing the computed index. This allows next
  time using this partially indexed reference library to accelerate the taxonomic annotations. 
- Adding of the function `gsub` to the expression language for substituting string pattern.

## May 2nd, 2023. Release 4.0.3

### New features

- Adding of the function `contains` to the expression language for testing if a map contains a key.
  It can be used from `obibrep` to select only sequences occurring in a given sample :

  ```{bash}
  obigrep -p 'contains(annotations.merged_sample,"15a_F730814")' wolf_new_tag.fasta
  ```   
- Adding of a new command `obipcrtag`. It tags raw Illumina reads with the identifier of their corresponding
  sample. The tags added are the same as those added by `obimultiplex`. The produced forward and reverse files
  can then be split into different files using the `obidistribute` command.

  ```{bash}
  obitagpcr -F library_R1.fastq \
            -R library_R2.fastq \
            -t sample_ngsfilter.txt \
            --out tagged_library.fastq \
            --unidentified not_assigned.fastq
  ```

  the command produced four files : `tagged_library_R1.fastq` and `tagged_library_R2.fastq` containing the assigned reads and `not_assigned_R1.fastq` and `not_assigned_R2.fastq` containing the unassignable reads.

  the tagged library files can then be split using `obidistribute`: 

  ```{bash}
  mkdir pcr_reads
  obidistribute --pattern "pcr_reads/sample_%s_R1.fastq" -c sample tagged_library_R1.fastq
  obidistribute --pattern "pcr_reads/sample_%s_R2.fastq" -c sample tagged_library_R2.fastq
  ```

- Adding of two options **--add-lca-in** and **--lca-error** to `obiannotate`. These options aim to help during
  construction of reference database using `obipcr`. On obipcr output, it is commonly run obiuniq. To merge
  identical sequences annotated with different taxids, it is now possible to use the following strategie :

  ```{bash}
  obiuniq -m taxid myrefdb.obipcr.fasta \
  | obiannotate -t taxdump --lca-error 0.05 --add-lca-in taxid \
  > myrefdb.obipcr.unique.fasta
  ```

  The `obiuniq` call merge identical sequences keeping track of the diversity of the taxonomic annotations in the
  `merged_taxid` slot, while `obiannotate` loads a NCBI taxdump and computes the lowest common ancestor of the taxids represented in `merged_taxid`. By specifying **--lca-error** 0.05, we indicate that we allow for at most 5% of the taxids disagreeing with the computed LCA. The computed LCA is stored in the slot specified as a parameter of the option **--add-lca-in**. Scientific name and actual error rate corresponding to the estimated LCA are also stored in the sequence annotation.

### Enhancement

- Rename the `forward_mismatches` and `reverse_mismatches` from instanced by `obimutiplex` into
  `forward_error` and `reverse_error` to be coherent with the tags instanced by `obipcr`

### Corrected bugs

- Correction of a bug in memory management and Slice recycling.
- Correction of the **--fragmented** option help and logging information
- Correction of a bug in `obiconsensus` leading into the deletion of a base close to the
  beginning of the consensus sequence.

## March 31th, 2023. Release 4.0.2

### Compiler change

*OBItools4* requires now GO 1.20 to compile.

### New features

- Add the possibility for looking pattern with indels. This has been added to `obimultiplex` 
  through the **--with-indels** option.
- Every obitools command has a **--pprof** option making the command publishing a profiling web
  site available at the address : [http://localhost:8080/debug/pprof/](http://localhost:8080/debug/pprof/)
- A new `obiconsensus` command has been added. It is a prototype. It aims to build a consensus sequence
  from a set of reads. The consensus is estimated for all the sequences contained in the input file.
  If several input files, or a directory name are provided the result contains a consensus per file.
  The id of the sequence is the name of the input file depleted of its directory name and of all its
  extensions.
- In `obipcr` an experimental option **--fragmented** allows for spliting very long query sequences into
  shorter fragments with an overlap between the two contiguous fragment insuring that no amplicons are
  missed despite the split. As a site effect some amplicon can be identified twice.
- In `obipcr` the -L option is now mandatory.
  

### Enhancement

- Add support for IUPAC DNA code into the DNA sequence LCS computation and an end free gap mode.
  This impact `obitag` and `obimultiplex` in the **--with-indels** mode.
- Print the synopsis of the command when an error is done by the user at typing the command
- Reduced the memory copy and allocation during the sequence creation.


### Corrected bugs

- Better management of non-existing files. The produced error message is not yet perfectly clear.
- Patch a bug leading with some programs to crash because of : "*empty batch pushed on the channel*"
- Patch a bug when directory names are used as input data name preventing the system to actually
  analyze the collected files.
- Make the **--help** or **-h** options working when mandatory options are declared
- In `obimultiplex` correct a bug leading to a wrong report of the count of reverse mismatch for 
  sequences in reverse direction.
- In `obimultiplex` correct a bug when not enough space exist between the extremities of the sequence
  and the primer matches to fit the sample identification tag 
- In `obipcr` correction of bug leading to miss some amplicons when several amplicons are present on the
  same large sequence. 
  
## March 7th, 2023. Release 4.0.1

### Corrected bugs

- Makes progress bar updating at most 10 times per second.
- Makes the command exiting on error if undefined options are used.
  
### Enhancement

- *OBITools* are automatically processing all the sequences files contained in a directory and its sub-directory   
  recursively if its name is provided as input. To process easily Genbank files, the corresponding filename
  extensions have been added. Today the following extensions are recognized as sequence files : `.fasta`, `.fastq`, 
  `.seq`, `.gb`, `.dat`, and `.ecopcr`. The corresponding gziped version are also recognized (e.g. `.fasta.gz`)

### New features

- Takes into account the `OBIMAXCPU` environmental variable to limit the number of CPU cores used
  by OBITools in bash the below command will limit to 4 cores the usage of OBITools

  ```bash
  export OBICPUMAX=4
  ```

- Adds a new option --out|-o allowing to specify the name of an outpout file.
  
  ```bash
  obiconvert -o xyz.fasta xxx.fastq
  ```

  is thus equivalent to

  ```bash
  obiconvert  xxx.fastq > xyz.fasta
  ````

  That option is actually mainly useful for dealing with paired reads sequence files.

- Some OBITools (now `obigrep` and `obiconvert`) are capable of using paired read files. 
  Options have been added for this (**--paired-with** _FILENAME_, and **--paired-mode** _forward|reverse|and|andnot|xor_). This, in combination with the **--out** option shown above, ensures that the two matched files remain consistent when processed. 

 - Adding of the function `ifelse` to the expression language for computing conditionnal values. 
 - Adding two function to the expression language related to sequence conposition : `composition` and `gcskew`.
   Both are taking a sequence as single argument.

## February 18th, 2023. Release 4.0.0

It is the first version of the *OBITools* version 4. I decided to tag then following two weeks
of intensive data analysis with them allowing to discover many small bugs present in the previous
non-official version. Obviously other bugs are certainly persent in the code, and you are welcome
to use the git ticket system to mention them. But they seems to produce now reliable results.

### Corrected bugs

- On some computers the end of the output file was lost, leading to the loose of sequences and
  to the production of incorrect file because of the last sequence record, sometime truncated in 
  its middle. This was only occurring when more than a single CPU was used. It was affecting every obitools.
- The `obiparing` software had a bug in the right aligment procedure. This led to the non alignment
  of very sort barcode during the paring of the forward and reverse reads.
- The `obipairing` tools had a non deterministic comportment when aligning a paor very low quality reads.
  This induced that the result of the same low quality read pair was not the same from run to run.

### New features

- Adding of a `--compress|-Z` option to every obitools allowing to produce `gz` compressed output. OBITools
  were already able to deal with gziped input files transparently. They can now produce their résults in the same format.
- Adding of a `--append|-A` option to the `obidistribute` tool. It allows to append the result of an 
  `obidistribute` execution to preexisting files.
- Adding of a `--directory|-d` option to the `obidistribute` tool. It allows to declare a secondary 
  classification key over the one defined by the '--category|-c` option. This extra key leads to produce
  directories in which files produced according to the primary criterion are stored.
- Adding of the functions `subspc`, `printf`, `int`, `numeric`, and `bool` to the expression language. 
