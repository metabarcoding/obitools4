# OBITools release notes

## February $18^th$, 2023. Release 4.0.0

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

### New functionality

- Adding of a `--compress|-Z` option to every obitools allowing to produce `gz` compressed output. OBITools
  were already able to deal with gziped input files transparently. They can now produce their résults in the same format.
- Adding of a `--append|-A` option to the `obidistribute` tool. It allows to append the result of an 
  `obidistribute` execution to preexisting files.
- Adding of a `--directory|-d` option to the `obidistribute` tool. It allows to declare a secondary 
  classification key over the one defined by the '--category|-c` option. This extra key leads to produce
  directories in which files produced according to the primary criterion are stored.
- Adding of the functions `subspc`, `printf`, `int`, `numeric`, and `bool` to the expression language. 