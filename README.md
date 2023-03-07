# obitools V4

*OBITools V4* is the new version of *OBITools*. They are closer to *OBITools2* than to *OBITools3*.
They are implemented in *GO* and are tens of times faster than OBITools2.

The git for *OBITools4* is available at :

> https://metabarcoding.org/obitools4

## Installing *OBITools V4*

An installation script that compiles the new *OBITools* on your Unix-like system is available online.
The easiest way to run it is to copy and paste the following command into your terminal

```{bash}
curl -L https://metabarcoding.org/obitools4/install.sh | bash
```

By default, the script installs the *OBITools* commands and other associated files into the `/usr/local` directory.
The names of the commands in the new *OBITools4* are mostly identical to those in *OBITools2*.
Therefore, installing the new *OBITools* may hide or delete the old ones. If you want both versions to be 
available on your system, the installation script offers two options:


>  -i, --install-dir       Directory where obitools are installed 
>                          (as example use `/usr/local` not `/usr/local/bin`).
> 
>  -p, --obitools-prefix   Prefix added to the obitools command names if you
>                          want to have several versions of obitools at the
>                          same time on your system (as example `-p g` will produce 
>                          `gobigrep` command instead of `obigrep`).

You can use these options by following the installation command:

```{bash}
curl -L https://metabarcoding.org/obitools4/install.sh | \
      bash -s -- --install-dir test_install --obitools-prefix k
```

In this case, the binaries will be installed in the `test_install` directory and all command names will be prefixed with the letter `k`. Thus `obigrep` will be named `kobigrep`.

## Continuing the analysis...

Before with _OBITools2_ to continue the analysis, `obitab` was used as last command to produce a tab delimited file that was loadable in R or in any spreadsheet. The generated file was huge and required to load the full dataset in memory to be produced. Hereby _OBITools4_ proposes to substitute the `obitab` usage by the [ROBIFastRead](https://git.metabarcoding.org/obitools/obitools4/robireadfasta) R module.


