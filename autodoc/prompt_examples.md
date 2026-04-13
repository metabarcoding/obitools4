# Task

Given `autodoc/cmd/obi{xxx}.md`, produce:
1. `autodoc/examples/obi{xxx}/` — a directory containing synthetic input sequence files
   that allow every example in the EXAMPLES section to be executed and validated.
2. An updated `autodoc/cmd/obi{xxx}.md` — with corrected EXAMPLES and an enriched OUTPUT
   section describing observed output annotations.

---

## TOOL CALL FORMAT — enforce before every call

A tool call is exactly:

    <function=tool_name>
    {"param": "value"}
    </function>

Rules (no exceptions):
- `<` is immediately followed by `f` — zero spaces, zero characters in between.
- Parameters are a **single JSON object** — no XML tags, no `<parameter=...>`, no `</parameter>`.
- No outer wrapper — never use `<tool_call>`, `<tool_use>`, or any other enclosing tag.
- Tool name is lowercase with double underscores.

---

## HALLUCINATION GUARD

Every sequence written in STATE 2 must be biologically valid for the command being
tested. Derive sequence content from the OPTIONS and OUTPUT sections of `$doc` — never
invent behaviour not described there.

**EXECUTION GUARD — critical:** The `## Observed output example` subsection added in
STATE 5 MUST contain verbatim bytes from `$outputs` (actual tool output read in STATE 4).
It MUST NOT be invented or approximated. If no command succeeded, omit the subsection
entirely rather than writing invented content.

---

## DOCUMENT PRESERVATION — critical

The output of STATE 5 is `$doc` with **surgical edits only**. The rules are:

- Copy the ENTIRE content of `$doc` verbatim into the new file.
- Apply ONLY the three modifications described in STATE 5 (EXAMPLES update,
  prose corrections, OUTPUT subsection addition).
- Do NOT reformat, reorder, rewrite, or restructure any heading, paragraph,
  option list, or prose from `$doc` **unless it is factually contradicted by
  actual execution results** (see Modification 2 in STATE 5).
- Do NOT add new top-level sections (no ENVIRONMENT VARIABLES, no duplicate OUTPUT, etc.).
- Do NOT change section title casing, Markdown heading levels, or list syntax.
- If in doubt, leave the section exactly as it appears in `$doc`.

---

## FASTQ FORMAT — mandatory structure

A valid FASTQ record is **exactly 4 lines** in this order:

```
@<identifier> <optional description>
<nucleotide sequence>          ← MUST be non-empty (≥ 10 characters, A/T/G/C only)
+
<quality string>               ← MUST be the exact same length as the sequence line
```

Common mistakes that are **forbidden**:
- Writing `@header\n+\nquality` with the sequence line missing.
- Writing a quality string shorter or longer than the sequence.
- Mixing `>` (FASTA) and `@` (FASTQ) headers in the same file.
- Writing `~`-separated fields (e.g. `@seq002~description~here`) — use a space.

---

## OUTPUT FORMAT GUARD

OBITools4 determines the output format from the **data content and explicit flags**,
**not from the output filename extension**. A file named `out.fasta` will contain FASTQ
if quality scores are present and no `--fasta-output` flag is given.

Rules when designing examples:
- If the example is meant to produce FASTA output from FASTQ input, the command MUST
  include `--fasta-output`.
- If the example is meant to produce FASTQ output from FASTA input, the command MUST
  include `--fastq-output`.
- Never assume an output format from the filename alone.
- Verify the actual format of each output file in STATE 3b by checking its first
  character (`>` = FASTA, `@` = FASTQ, `[` or `{` = JSON).

---

## OPTION VALIDATION GUARD

Before writing any example command in STATE 2, explicitly cross-check each option
against the OPTIONS section of `$doc`:

- Every flag used must appear in the OPTIONS section with the claimed semantics.
- Input-format flags (`--fasta`, `--fastq`, `--csv`, `--genbank`, `--embl`,
  `--ecopcr`) tell the tool how to **read** the input. They do NOT affect the
  output format.
- Output-format flags (`--fasta-output`, `--fastq-output`, `--json-output`) tell
  the tool what format to **write**. If there is no `--csv-output` (or similar) in
  the OPTIONS section, do NOT write an example claiming CSV output.
- If an option needed for a working example is absent from `$doc`, mark that example
  as SKIP rather than inventing a flag.

---

## ANNOTATION RULES — CRITICAL

When creating FASTA/FASTQ files with annotations:
- Use **only** valid annotation attribute names: `taxid`, `scientific_name`, `rank`, `definition`, `sample`, `run_id`, `instrument`
- For taxonomy data: use `taxid` (NCBI Taxonomy ID) and `scientific_name` — never invent taxids
- Examples of valid taxonomy annotations:
  - `>seq001 {"taxid":2}` — Bacteria (valid NCBI taxid)
  - `>seq002 {"taxid":2157,"scientific_name":"Archaea"}` — Archaea (valid NCBI taxid)
  - `>seq003 {"taxid":2759,"scientific_name":"Eukaryota"}` — Eukaryota (valid NCBI taxid)
- NEVER use invented taxids
- **Map attributes** (JSON maps) must have names ending with `_merged` (e.g., `taxid_merged`, `sample_merged`)

---

## CSV FILES FOR JOINS

When creating CSV files for `obijoin`:
- Do NOT include the ID column in the CSV (the join key is specified separately via `--by`)
- The CSV format is auto-detected; do NOT use `--csv` flag
- Example CSV structure for taxid join:
  ```
  taxid,scientific_name,phylum
  2,Bacteria,Proteobacteria
  2157,Archaea,Euryarchaeota
  2759,Eukaryota,Arthropoda
  ```
- Example command: `obijoin --join-with taxonomy.csv --by taxid sequences.fasta`

---

## PIPELINE

Execute the five states below in order. Do not skip states. Do not merge states.

---

### STATE 1 — Read the documentation file and fetch pipeline command docs

**Input:** nothing.
**Action:**

Step 1a — read the autodoc file:
```
<function=Read>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/cmd/obi{xxx}.md"}
</function>
```

Step 1b — scan the EXAMPLES section of the file just read for any `obi*` commands
other than `obi{xxx}` itself that appear in pipeline examples (e.g. `obigrep`, `obiuniq`,
`obiclean`). For each such command found, emit a WebFetch call to retrieve its online
documentation (in the same parallel message as Step 1a if possible, otherwise
immediately after):
```
<function=WebFetch>
{"url": "https://obitools4.metabarcoding.org/obitools/obi<other>/"}
</function>
```
If the page returns a 404 or error, store an empty string for that command.

**Output:** store content as `$doc`, and store fetched pages as `$pipeline_docs`
(a map from command name to page content).
**Stop.** Do not interpret or summarise. Proceed to STATE 2.

---

### STATE 2 — Analyse examples and design input files

**Input:** `$doc`.
**Action (no tool calls):**

1. Extract every example command from the EXAMPLES section of `$doc`.
   - Identify every distinct input filename referenced (e.g. `sequences.fasta`,
     `reads_R1.fastq`, …).
   - Identify every option used and verify each against the OPTIONS section (see
     OPTION VALIDATION GUARD above).
   - For any `obi*` command used in a pipeline (not `obi{xxx}` itself), verify its
     flags and expression syntax against `$pipeline_docs`. If `$pipeline_docs` for
     that command is empty (page not found), mark the example as SKIP rather than
     guessing the syntax.
   - **Coverage check — command-specific options:** list all command-specific options
     from the OPTIONS section (excluding those covered by standard option-sets: `--fasta`,
     `--fastq`, `--out`, `--compress`, `--max-cpu`, etc.). Verify that every such option
     appears in at least one non-skipped example. If any option is not covered, **add an
     additional example** that exercises it before proceeding.
   - **Skip any example that requires an external resource** (taxonomy database,
     remote URL, pre-existing output file from a previous step not produced here).
     Mark it as SKIP — it will be kept verbatim in the EXAMPLES section without
     a `**Expected output:**` annotation.
   - **`--paired-with` examples:** `--paired-with` requires `--out` (standard output
     cannot be used). The command produces TWO output files named `<prefix>_R1.ext`
     and `<prefix>_R2.ext` where `<prefix>` is the stem of the value given to `--out`
     and `.ext` is the format extension. For example:
     `obi{xxx} --paired-with reverse.fastq --out out_paired.fastq forward.fastq`
     produces `out_paired_R1.fastq` and `out_paired_R2.fastq`.
     Do NOT use `>` redirection for paired-with examples — use `--out` only.
     In STATE 4, read both `_R1` and `_R2` output files.

2. For each distinct input filename, design synthetic sequence content that:
   - Is **minimal** (≤ 20 sequences, each ≤ 300 bp).
   - Contains sequences that **will** produce output for the given command (positive cases).
   - Contains at least one sequence that **will not** produce output, to confirm filtering
     (negative case), when the command filters sequences.
   - Exercises every option combination present in the non-skipped examples.
   - Uses realistic-looking identifiers (`seq001`, `seq002`, …) and a short
     definition that describes what makes the sequence relevant to the test.

3. **File format rules (strictly enforced):**

   **FASTA:** one `>id description` header line, then the sequence on one or more
   lines (60 bp per line). Every sequence must be non-empty (≥ 10 bp, A/T/G/C only).

   **FASTQ:** exactly 4 lines per record — see FASTQ FORMAT section above.
   Before finalising the FASTQ content, mentally verify each record:
   - Line 1 starts with `@`, has an identifier, optionally a space and description.
   - Line 2 is the nucleotide sequence (non-empty, ≥ 10 characters).
   - Line 3 is exactly `+` (nothing else required).
   - Line 4 is the quality string with **exactly the same number of characters**
     as line 2.
   If any record fails this check, fix it before proceeding.

4. Rewrite every non-skipped example command into two forms:
   - `$cmds_doc`: the bare command as it will appear in the documentation — references
     only filenames present in `autodoc/examples/obi{xxx}/`, output redirected to a
     descriptive filename (e.g. `out_default.fasta`). **No `cd` prefix.**
   - `$cmds_run`: the same command prefixed with the `cd` so it can be executed:
     `cd /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx} &&`

**Output:** store file designs as `$files`, `$cmds_doc`, and `$cmds_run`.
**Stop.** Proceed to STATE 3.

---

### STATE 3 — Write input files, validate them, and run examples

**Input:** `$files`, `$cmds_doc`, `$cmds_run`.

**Step 3a — create input files (parallel):**
Emit one Write call per input file designed in STATE 2.

```
<function=Write>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx}/FILENAME", "content": "..."}
</function>
```

**Stop.** Wait for all writes to complete. Then proceed to Step 3b.

**Step 3b — validate input files:**
Before running any example, emit one Bash call that checks every written input file:

```
<function=Bash>
{"command": "cd /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx} && python3 -c \"\nimport sys\nfor fname in $(echo FILENAMES):\n    lines = open(fname).readlines()\n    if fname.endswith('.fastq'):\n        assert len(lines) % 4 == 0, f'{fname}: line count not multiple of 4'\n        for i in range(0, len(lines), 4):\n            hdr, seq, plus, qual = lines[i:i+4]\n            assert hdr.startswith('@'), f'{fname} record {i//4+1}: header must start with @'\n            seq = seq.rstrip(); qual = qual.rstrip()\n            assert len(seq) >= 10, f'{fname} record {i//4+1}: sequence too short ({len(seq)})'\n            assert len(seq) == len(qual), f'{fname} record {i//4+1}: seq len {len(seq)} != qual len {len(qual)}'\n    elif fname.endswith('.fasta') or fname.endswith('.fa'):\n        assert lines[0].startswith('>'), f'{fname}: first line must start with >'\nprint('All input files valid')\n\" 2>&1; echo EXIT:$?"}
</function>
```

If validation fails (EXIT non-zero or output is not `All input files valid`): fix the
offending file(s) with new Write calls, then re-run validation. Do NOT proceed to
Step 3c until validation passes.

**Step 3c — run examples (sequential, one Bash call at a time):**
Emit ONE Bash call, wait for the result, then emit the next. Do NOT batch them.

```
<function=Bash>
{"command": "cd /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx} && COMMAND 2>&1; echo EXIT:$?"}
</function>
```

After each successful run (EXIT:0), immediately verify the output file was actually
created and is non-empty with a second Bash call:

```
<function=Bash>
{"command": "ls -la /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx}/OUTPUT_FILE && head -c 100 /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx}/OUTPUT_FILE"}
</function>
```

Also verify the output format matches expectation using the first character rule
(see OUTPUT FORMAT GUARD): `>` = FASTA, `@` = FASTQ, `[`/`{` = JSON. If the format
is wrong, add the missing `--fasta-output` / `--fastq-output` / `--json-output` flag,
update `$cmds_doc` and `$cmds_run`, and re-run.

For each command, record in `$runs`:
- The `$cmds_doc` form (bare command for documentation).
- Exit code.
- The output filename(s).
- The confirmed output format (FASTA / FASTQ / JSON).
- The full stdout/stderr text.

If a command fails (EXIT non-zero): diagnose the error from stderr, fix the command,
update both `$cmds_doc` and `$cmds_run`, and re-run.
Do NOT proceed to STATE 4 until all non-skipped commands have EXIT:0 and verified
non-empty output files.

**Output:** store per-command results as `$runs`.
**Stop.** Proceed to STATE 4.

---

### STATE 4 — Read output files

**Input:** `$runs` (output file paths from STATE 3).
**Action:** emit one Read call per output file that was successfully produced (EXIT:0).

```
<function=Read>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx}/OUTPUT_FILE"}
</function>
```

Emit all reads in a single parallel message.

**Output:** store contents as `$outputs`.
**Stop.** Proceed to STATE 5.

---

### STATE 5 — Update the documentation file

**Input:** `$doc`, `$runs`, `$outputs`, `$cmds_doc`.

Produce the updated file by copying `$doc` **verbatim** and applying ONLY the
three modifications below. Re-read the DOCUMENT PRESERVATION rules at the top before
writing.

#### Modification 1 — EXAMPLES section

For each non-skipped example:
- Replace the original command with the rewritten `$cmds_doc` form.
- Keep the one-line biological use-case comment above the code block unchanged.
- The `**Expected output:**` annotation goes on its own line **after** the closing
  triple-backtick of the code block, never inside it:

  ```
  ```bash
  obi{xxx} [options] input_file > out_name.fasta
  ```

  **Expected output:** N sequences written to `out_name.fasta`.
  ```

  where N is the count of lines starting with `>` or `@` in the corresponding
  `$outputs` entry.

For skipped examples: keep them exactly as they are in `$doc` with no annotation.

#### Modification 2 — Prose corrections (DESCRIPTION, OPTIONS, NOTES, …)

After completing all runs in STATE 3, compare `$runs` and `$outputs` against the
prose in `$doc` outside the EXAMPLES section. For each **factual contradiction**
found — where the documentation claims a behaviour that actual execution disproves —
apply a minimal correction:

- Fix only the specific sentence or phrase that is wrong. Do not rewrite the
  surrounding paragraph.
- Preserve the original wording as much as possible; change only what is incorrect.
- Examples of things to correct:
  - An option described as producing output X when it actually produces output Y.
  - A default value stated incorrectly.
  - An attribute name that differs from what appears in actual output.
  - A claim about which sequences are selected/discarded that contradicts observed results.
  - An output format claimed by the documentation that differs from the actual output
    format observed (e.g. claiming CSV output when the tool produces FASTA).
- After each corrected passage, add an inline HTML comment documenting the fix:
  `<!-- corrected: <brief reason, e.g. "actual output is FASTA not CSV"> -->`
- Do NOT "improve" text that is merely incomplete or imprecise — only fix outright
  contradictions with observed behaviour.

#### Modification 3 — OUTPUT section

Find the existing `# OUTPUT` section in `$doc`. At the very end of that section
(before the next `---` or `#` heading), append a single new subsection:

```markdown
## Observed output example

```
<verbatim excerpt — first ≤ 10 sequences from the first successful $outputs entry>
```
```

Rules:
- The excerpt is copied byte-for-byte from `$outputs`. No editing, no truncation
  within a sequence record.
- Do NOT duplicate the OUTPUT section. There must be exactly one `# OUTPUT` heading
  in the resulting file.
- If no output was successfully produced, omit this subsection entirely.

#### Final write

```
<function=Write>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/cmd/obi{xxx}.md", "content": "..."}
</function>
```

**Stop. Do not emit any text after the Write call.**
