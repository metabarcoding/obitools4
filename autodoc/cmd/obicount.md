# NAME

obicount — counts the sequences present in a file of sequences

---

# SYNOPSIS

```
obicount [--batch-mem <string>] [--batch-size <int>] [--batch-size-max <int>]
         [--csv] [--debug] [--ecopcr] [--embl] [--fasta] [--fastq]
         [--genbank] [--help|-h|-?] [--input-OBI-header]
         [--input-json-header] [--max-cpu <int>] [--no-order] [--pprof]
         [--pprof-goroutine <int>] [--pprof-mutex <int>] [--reads|-r]
         [--silent-warning] [--solexa] [--symbols|-s] [--u-to-t]
         [--variants|-v] [--version] [<args>]
```

---

# DESCRIPTION

obicount is a command-line tool designed to count biological sequences from various input formats. It helps biologists quickly obtain quantitative metrics about sequence collections, which is essential for quality control, data assessment, and pipeline monitoring. The tool can count reads (total sequences), variants (unique sequence strings), or symbols (sum of character lengths), providing flexibility to focus on specific aspects of sequence data depending on the analysis needs.

---

# INPUT

obicount accepts input from files or stdin, supporting multiple biological sequence formats:
- FASTA (.fasta[.gz])
- FASTQ (.fastq[.fq][.gz]) 
- GenBank/EMBL (.gb|.gbff|.dat[.gz])
- ecoPCR format (.ecopcr[.gz])
- CSV format (--csv flag)

Input can be provided as multiple filenames or read from stdin. The tool automatically detects file formats and parses sequences accordingly.

---

# OUTPUT

obicount outputs one or more of the following metrics, depending on the flags used:

- **Read counts**: Total number of sequences in the input
- **Variant counts**: Number of unique sequence strings (distinct sequences)
- **Symbol counts**: Sum of all character lengths across all sequences

When no specific counting flags are provided (-r, -v, -s), all three metrics are reported by default. Output is printed to stdout in CSV format with headers: `entities,n` for the type of entity counted, followed by the count value.

---

# OPTIONS

## General Options
- --help|-h|-?          
  Show help message and exit.
  
- --max-cpu <int>       
  Number of parallel threads computing the result (default: 16, env: OBIMAXCPU).
  
- --debug              
  Enable debug mode, by setting log level to debug. (default: false, env: OBIDEBUG)

- --silent-warning     
  Stop printing of the warning message (default: false, env: OBIWARNING)

## Input Format Options  
- --fasta             
  Read data following the fasta format. (default: false)
  
- --fastq            
  Read data following the fastq format. (default: false)
  
- --genbank           
  Read data following the Genbank flatfile format. (default: false)
  
- --embl              
  Read data following the EMBL flatfile format. (default: false)
  
- --ecopcr            
  Read data following the ecoPCR output format. (default: false)
  
- --csv               
  Read data following the CSV format. (default: false)

## Input Header Options
- --input-OBI-header   
  FASTA/FASTQ title line annotations follow OBI format. (default: false)
  
- --input-json-header  
  FASTA/FASTQ title line annotations follow json format. (default: false)

## Counting Mode Options
- --reads|-r          
  Prints read counts. (default: false)
  
- --variants|-v       
  Prints variant counts. (default: false)
  
- --symbols|-s        
  Prints symbol counts. (default: false)

## Processing Options
- --u-to-t            
  Convert Uracil to Thymine. (default: false, env: OBISOLEXA)
  
- --solexa            
  Decodes quality string according to the Solexa specification. (default: false, env: OBISOLEXA)
  
- --no-order          
  When several input files are provided, indicates that there is no order among them. (default: false)

## Performance Options
- --batch-mem <string>      
  Maximum memory per batch (e.g. 128K, 64M, 1G; default: 128M). Set to 0 to disable. (default: "", env: OBIBATCHMEM)
  
- --batch-size <int>        
  Minimum number of sequences per batch (floor, default 1) (default: 1, env: OBIBATCHSIZE)
  
- --batch-size-max <int>    
  Maximum number of sequences per batch (ceiling, default 2000) (default: 2000, env: OBIBATCHSIZEMAX)
  
- --max-cpu <int>          
  Number of parallele threads computing the result (default: 16, env: OBIMAXCPU)

## Profiling Options
- --pprof                  
  Enable pprof server. Look at the log for details. (default: false)
  
- --pprof-goroutine <int> 
  Enable profiling of goroutine blocking profile. (default: 6060, env: OBIPPROFGOROUTINE)
  
- --pprof-mutex <int>      
  Enable profiling of mutex lock. (default: 10, env: OBIPPROFMUTEX)
  
- --version                
  Prints the version and exits. (default: false)

---

# EXAMPLES

# Count total number of sequences in a FASTA file
# Useful for quick assessment of dataset size
obicount input.fasta
**Expected output:** 4 sequences, out_default.txt

# Count only the number of unique sequence variants  
# Helpful for identifying genetic diversity in population data
obicount --variants input.fasta
**Expected output:** 4 sequences, out_variants.txt

# Count sum of all sequence symbol lengths (nucleotides/amino acids)
# Useful for estimating total data volume or computing average read length
obicount --symbols input.fasta
**Expected output:** 4 sequences, out_symbols.txt

# Count reads from FASTQ format with quality scores
# Essential for assessing read throughput in sequencing data
obicount --fastq --reads input.fastq
**Expected output:** 4 sequences, out_fastq_reads.txt

---

# OUTPUT

## Observed output example

```
time="2026-04-02T19:33:11+02:00" level=info msg="Number of workers set 16"
time="2026-04-02T19:33:11+02:00" level=info msg="Found 1 files to process"
time="2026-04-02T19:33:11+02:00" level=info msg="input.fasta mime type: text/fasta"
entities,n
variants,5
reads,5
symbols,435
```

---

# SEE ALSO

- obiconvert - Convert between biological sequence file formats
- obiuniq - Remove duplicate sequences from files

---

# NOTES

_(not available)_