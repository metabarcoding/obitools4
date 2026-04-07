# NAME

obimicrosat — looks for microsatellites sequences in a sequence file

---

# SYNOPSIS

```
obimicrosat [options] [<filename>...]
```

---

# DESCRIPTION

`obimicrosat` scans DNA sequences for simple sequence repeats (SSRs), also called
microsatellites — tandem repetitions of a short motif (1–6 bp by default). For each
sequence containing a qualifying repeat, the command annotates it with the location,
unit sequence, repeat count, and flanking regions, then writes it to output. Sequences
with no detected microsatellite are silently discarded.

The detection works in two passes. A first regular expression finds any tandem repeat
satisfying the unit-length and repeat-count constraints. The true minimal repeat unit
is then determined, and a second scan refines the exact boundaries. The repeat unit is
normalized to its lexicographically smallest rotation across all rotations and its
reverse complement, which allows equivalent loci to be grouped consistently across
samples.

By default, when the canonical form of a unit requires the reverse complement, the
whole sequence is reoriented so that the microsatellite is always reported on the
direct strand of the normalized unit. This behaviour can be disabled with
`--not-reoriented`.

A common use case is identifying polymorphic SSR markers for population genetics, or
flagging repeat-rich regions before designing PCR primers.

---

# INPUT

Accepts one or more sequence files on the command line. If no file is given, sequences
are read from standard input. Supported formats include FASTA, FASTQ, JSON/OBI, GenBank,
EMBL, ecoPCR output, and CSV. Compressed files (gzip) are handled transparently.
Format is detected automatically unless overridden by input flags.

---

# OUTPUT

Outputs only the sequences in which a microsatellite was found. Each retained sequence
carries the following additional attributes:

| Attribute | Content |
|---|---|
| `microsat` | Full repeat region as a string |
| `microsat_from` | 1-based start position of the repeat |
| `microsat_to` | End position of the repeat (inclusive) |
| `microsat_unit` | Repeat unit as observed in the sequence |
| `microsat_unit_normalized` | Lexicographically smallest canonical form |
| `microsat_unit_orientation` | `direct` or `reverse` |
| `microsat_unit_length` | Length of the repeat unit (bp) |
| `microsat_unit_count` | Number of complete unit repetitions |
| `seq_length` | Total length of the (possibly reoriented) sequence |
| `microsat_left` | Flanking sequence to the left of the repeat |
| `microsat_right` | Flanking sequence to the right of the repeat |

When a sequence is reoriented (reverse-complemented), `_cmp` is appended to its
identifier.

The output format follows the same rules as the rest of OBITools4: FASTQ when quality
scores are present, FASTA or JSON/OBI otherwise, configurable via output flags.

## Observed output example

```
>seq001 {"definition":"dinucleotide AC repeat 16x with 40bp non-repetitive flanks","microsat":"acacacacacacacacacacacacacacacac","microsat_from":40,"microsat_left":"agtcgaacttgcatgccttcagggcaagtctagcttacg","microsat_right":"cgatagtcatgcaagtcttgcggcatagatcgttacca","microsat_to":71,"microsat_unit":"ac","microsat_unit_count":16,"microsat_unit_length":2,"microsat_unit_normalized":"ac","microsat_unit_orientation":"direct","seq_length":109}
agtcgaacttgcatgccttcagggcaagtctagcttacgacacacacacacacacacaca
cacacacacaccgatagtcatgcaagtcttgcggcatagatcgttacca
>seq006_cmp {"definition":"GT repeat 16x with 40bp non-repetitive flanks canonical form is AC","microsat":"acacacacacacacacacacacacacacacac","microsat_from":39,"microsat_left":"tggtaacgatctatgccgcaagacttgcatgactatcg","microsat_right":"cgtaagctagacttgccctgaaggcatgcaagttcgact","microsat_to":70,"microsat_unit":"ac","microsat_unit_count":16,"microsat_unit_length":2,"microsat_unit_normalized":"ac","microsat_unit_orientation":"reverse","seq_length":109}
tggtaacgatctatgccgcaagacttgcatgactatcgacacacacacacacacacacac
acacacacaccgtaagctagacttgccctgaaggcatgcaagttcgact
```

---

# OPTIONS

## Microsatellite detection

**`--min-unit-length` / `-m`**
- Default: `1`
- Minimum length in base pairs of the repeated motif. Set to `2` to exclude
  mononucleotide repeats, `3` for di- and mononucleotide-free searches, etc.

**`--max-unit-length` / `-M`**
- Default: `6`
- Maximum length in base pairs of the repeated motif. Increasing this value detects
  longer repeat units (minisatellites) at the cost of more complex patterns.

**`--min-unit-count`**
- Default: `5`
- Minimum number of times the motif must be repeated. A value of `5` with a 2 bp unit
  requires at least 10 bp of pure repeat.

**`--min-length` / `-l`**
- Default: `20`
- Minimum total length (in bp) of the repeat region. This filter applies after the
  unit-count filter and is useful to exclude very short but technically qualifying
  repeats.

**`--min-flank-length` / `-f`**
- Default: `0`
- Minimum length of the flanking sequence on each side of the repeat. Sequences with
  flanks shorter than this threshold are discarded, which is useful when the output
  will feed a primer-design step.

**`--not-reoriented` / `-n`**
- Default: `false` (reorientation is active by default)
- When set, sequences are never reverse-complemented to match the canonical orientation
  of the repeat unit. The microsatellite is reported as found, in its original
  orientation.

## Input / output

Inherited from the standard OBITools4 conversion layer. Common flags include:

**`--input-OBI-header`** — parse OBI-style FASTA/FASTQ headers.
**`--input-json-header`** — parse JSON-encoded headers.
**`--skip-empty`** — skip sequences with no nucleotides.
**`--u-to-t`** — convert U to T (RNA → DNA).
**`--output-json-header`** — write JSON-encoded headers.
**`--output-obi-header`** — write OBI-style headers.
**`--gzip`** — compress output with gzip.
**`--workers` / `-p`** — number of parallel processing workers.

---

# EXAMPLES

```bash
# Detect default microsatellites (unit 1–6 bp, ≥5 repeats, ≥20 bp total)
obimicrosat sequences.fasta > out_default.fasta
```

**Expected output:** 6 sequences written to `out_default.fasta`.

```bash
# Restrict to di- and trinucleotide repeats only
obimicrosat -m 2 -M 3 sequences.fasta > out_dinucleotide.fasta
```

**Expected output:** 4 sequences written to `out_dinucleotide.fasta`
(mononucleotide and tetranucleotide repeats excluded).

```bash
# Require at least 30 bp flanking sequence on each side (for primer design)
obimicrosat -f 30 sequences.fasta > out_primer_ready.fasta
```

**Expected output:** 3 sequences written to `out_primer_ready.fasta`
(sequences with flanks shorter than 30 bp are discarded).

```bash
# Keep sequences in their original orientation (no reverse-complement)
obimicrosat --not-reoriented sequences.fasta > out_no_reorient.fasta
```

**Expected output:** 6 sequences written to `out_no_reorient.fasta`
(GT-repeat sequence kept as-is without `_cmp` suffix; `microsat_unit_orientation` is `reverse`).

```bash
# Require at least 8 repeat units and a minimum repeat length of 30 bp
obimicrosat --min-unit-count 8 -l 30 sequences.fasta > out_strict.fasta
```

**Expected output:** 4 sequences written to `out_strict.fasta`
(short or low-count repeats excluded).

---

# SEE ALSO

`obigrep` — filter sequences by annotation after microsatellite detection.
`obiannotate` — add or modify sequence annotations.
`obiconvert` — format conversion for sequence files.

---

# NOTES

- Only sequences with at least one qualifying microsatellite are written to output;
  all others are silently filtered out.
- The normalization algorithm considers all rotations of the unit and their reverse
  complements, selecting the lexicographically smallest string. This ensures consistent
  grouping of loci regardless of which strand was sequenced.
- When reorientation is active (the default), sequences whose canonical unit falls on
  the reverse strand are reverse-complemented and their ID receives the suffix `_cmp`.
  Coordinate attributes (`microsat_from`, `microsat_to`) always refer to the
  (possibly reoriented) output sequence.
- Repetitive low-complexity sequences may match multiple overlapping patterns; only the
  first match is reported per sequence.
- Flanking sequences must be **non-repetitive** to avoid the tool detecting a tandem
  repeat within the flank instead of the intended SSR. When designing synthetic test
  data, ensure flanking regions do not contain tandem repeat motifs of their own.
