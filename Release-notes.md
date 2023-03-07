# OBITools release notes

## On going changes

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
