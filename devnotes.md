# OA2

## Initialisation of the GO directory structure

Principles are extracted from :

    - https://www.wolfe.id.au/2020/03/10/starting-a-go-project/
    - https://dev.to/johanlejdung/a-mini-guide-go-modules-and-private-repositories-4c7o
    - https://blog.otso.fr/2020-10-11-organisation-projet-go-standard

```bash
$ go mod init git.metabarcoding.org/lecasofts/go/obitools.git
go: creating new go.mod: module git.metabarcoding.org/lecasofts/go/obitools.git
```

```bash
go env -w GOPRIVATE=git.metabarcoding.org/lecasofts/go/*
```

```bash
mkdir cmd       # Inside this add a folder per program
mkdir pkg       # Inside this add a folder per library of the project
mkdir internal  # Inside this add a folder per private library of the project
```


## Some information in interfacing Go and C

    - <https://karthikkaranth.me/blog/calling-c-code-from-go/>
    
## Reading Fasta/Fastq files

    - <http://lh3lh3.users.sourceforge.net/parsefastq.shtml>
    
## Some consideration on `zlib`

### Closing a `gzFile`

https://stackoverflow.com/questions/65704567/how-to-properly-open-and-close-an-already-fopened-gzip-file-with-zlib

File descriptors are obtained from calls like open, dup, creat, pipe or fileno 
(in the file has been previously opened with fopen). 
The next call of gzclose on the returned gzFile will also close the file 
descriptor fd, just like fclose(fdopen(fd), mode) closes the file descriptor fd. 
If you want to keep fd open, use fd = dup(fd_keep); gz = gzdopen(fd, mode);. 
If you are using fileno() to get the file descriptor from a FILE *, then you 
will have to use dup() to avoid double-close()ing the file descriptor. 
Both gzclose() and fclose() will close the associated file descriptor, 
so they need to have different file descriptors.

