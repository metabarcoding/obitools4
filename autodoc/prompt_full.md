# Task

Produce complete documentation for the `obi{xxx}` CLI command:
1. `autodoc/cmd/obi{xxx}.md` — the markdown documentation file
2. `autodoc/examples/obi{xxx}/` — synthetic input sequence files for testing examples
3. `obitools4-doc/content/docs/commands/<category>/obi{xxx}/_index.md` — Hugo documentation page

Execute the three phases below **in order**. Do not skip phases. Do not merge phases.

---

## TOOL CALL FORMAT — enforce before every call

A tool call is exactly:

    <function=tool_name>
    {"param": "value"}
    </function>

Rules:
- `<` is immediately followed by `f` — zero spaces, zero characters in between.
- Parameters are a **single JSON object** — no XML tags, no `<parameter=...>`, no `</parameter>`.
- No outer wrapper — never use `<tool_call>`, `<tool_use>`, or any other enclosing tag.
- Tool name lowercase with double underscores — never ALL_CAPS, never single underscore between server and tool name.

CORRECT:  `<function=mcp__treesitter__treesitter_get_dependencies>`
WRONG:    `< function=mcp__treesitter__treesitter_get_dependencies >` ← spaces
WRONG:    `<function=mcp_treesitter_treesitter_get_dependencies>`     ← wrong separator

---

## HALLUCINATION GUARD — enforce before writing anything

OBITools4 is a **complete rewrite**. Training data about OBITools v1/v2/v3 is wrong for this version.

Before writing any sentence, apply this check:

> "Can I point to the exact line in $help or $docs that justifies this claim?"
> If NO → do not write it.

This applies to: option names, option flags, default values, file formats, behaviours, algorithms, output fields.
Omit rather than guess. A shorter correct page is better than a longer hallucinated one.

---

## PHASE 1 — Generate initial documentation file

(Equivalent to prompt_v2.md)

### STATE 1 — Gather raw data (parallel)

**Input:** nothing.
**Action:** emit all of the following tool calls in a single message.

Call 1 — dependencies of the main entry point:
```
<function=mcp__treesitter__treesitter_get_dependencies>
{"language": "go", "file_path": "cmd/obitools/obi{xxx}/main.go"}
</function>
```

Call 2 — dependencies of every file in the command package (one call per `.go` file in `pkg/obitools/obi{xxx}/`):
```
<function=mcp__treesitter__treesitter_get_dependencies>
{"language": "go", "file_path": "pkg/obitools/obi{xxx}/options.go"}
</function>
<function=mcp__treesitter__treesitter_get_dependencies>
{"language": "go", "file_path": "pkg/obitools/obi{xxx}/obi{xxx}.go"}
</function>
```

Call 3 — symbol outline of the command package:
```
<function=mcp__jcodemunch__get_file_outline>
{"repo": "git.metabarcoding.org/obitools/obitools4/obitools4", "file_paths": ["pkg/obitools/obi{xxx}/options.go", "pkg/obitools/obi{xxx}/obi{xxx}.go"]}
</function>
```

Call 4 — CLI help text:
```
<function=Bash>
{"command": "cd /Users/coissac/Sync/travail/__MOI__/GO/obitools4 && obi{xxx} --help 2>&1"}
</function>
```

**Output:** store results as `$deps`, `$outline`, `$help`.
**Stop.** Do not interpret, summarise, or write anything. Proceed to STATE 2.

---

### STATE 2 — Resolve documentation files

**Input:** `$deps`, `$outline`.
**Action (no tool calls):**

1. Collect every import path that starts with `git.metabarcoding.org/obitools/obitools4/obitools4/pkg/`.
2. Remove these infrastructure packages:
   - `pkg/obidefault`, `pkg/obiiter`, `pkg/obiformats`, `pkg/obioptions`, `pkg/obiutils`, `pkg/obiparams`
   - `pkg/obiseq` — keep only if `$outline` shows non-trivial sequence manipulation (custom methods or transformations beyond simple access).
3. Always add: `pkg/obitools/obi{xxx}` and `pkg/obitools/obiconvert`.
4. Map each remaining package path to its doc file:
   - take all path segments from `pkg/` onward (inclusive), replace `/` with `_` → `autodoc/docmd/<joined_segments>.md`
   - Example: `git.metabarcoding.org/.../pkg/obitools/obicsv` → `autodoc/docmd/pkg_obitools_obicsv.md`
   - Example: `git.metabarcoding.org/.../pkg/obiseq` → `autodoc/docmd/pkg_obiseq.md`

**Output:** store file list as `$docfiles`.
**Stop.** Proceed to STATE 3.

---

### STATE 3 — Read documentation files (parallel)

**Input:** `$docfiles`.
**Action:** emit one `Read` call per file in `$docfiles`, all in a single parallel message.

Do NOT read any `.go` source file. Do NOT produce any summary or analysis.

```
<function=Read>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/DOCFILE"}
</function>
```

**Output:** store all file contents as `$docs`.
**Stop.** Proceed to STATE 4.

---

### STATE 4 — Write the documentation file

**Input:** `$help`, `$docs`, `$outline`.
**Action:** fill the template below, then emit exactly one `Write` call.

**Source discipline:** every piece of information in the template MUST come from `$help`, `$docs`, or `$outline`.
If the source does not contain the information, write `_(not available)_` — never invent.

#### Output template

```markdown
# NAME

obi{xxx} — [FILL: one-line description. Source: first line of $help]

---

# SYNOPSIS

[FILL: verbatim USAGE block from $help, inside a code block]

---

# DESCRIPTION

[FILL: 2–4 paragraphs explaining what the command does, why a biologist would use it,
and what it produces. Source: $help description section + $docs.
No jargon. No implementation details (goroutines, channels, GC, arena).
No options that belong in the OPTIONS section.]

---

# INPUT

[FILL: accepted input formats and how to provide them. Source: $help + $docs/obiconvert.]

---

# OUTPUT

[FILL: output format and what fields/attributes are added or changed. Source: $help + $docs.
If the default output is JSON, state it clearly. If YAML, state it clearly. Do not assume.]

---

# OPTIONS

[FILL: one subsection per thematic group found in $help.
For each flag:
- Flag name(s): long form + short form if it exists. Source: $help exactly.
- Default: state it. Source: $help exactly. If absent from $help: write "none".
- Meaning: explain the biological or practical purpose, not just the mechanical action.
- Do NOT include any flag not present in $help.]

---

# EXAMPLES

[FILL: at least four copy-pasteable examples.
Each example:
1. A one-line comment explaining the biological use case.
2. The command inside a code block.
3. MUST include output redirection (`> out.fasta`) or `-o` flag so the user can reproduce the output.
4. Use consistent file names across examples (same input file for different options).
5. COVERAGE RULE: every command-specific option documented in the OPTIONS section MUST
   appear in at least one example. If needed, add extra examples to achieve full coverage.
Source for flags and options: $help only.]

---

# SEE ALSO

[FILL: related obi commands mentioned in $docs or $help. If none: omit section.]

---

# NOTES

[FILL: caveats, performance notes, known limitations. Source: $docs only.
If none: omit section.]
```

Emit the Write call:
```
<function=Write>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/cmd/obi{xxx}.md", "content": "..."}
</function>
```

**Stop.** Proceed to PHASE 2.

---

## PHASE 2 — Test examples and enrich documentation

(Equivalent to prompt_examples.md)

### DOCUMENT PRESERVATION — critical

The output of STATE 5 is `$doc` with **surgical edits only**. The rules are:

- Copy the ENTIRE content of `$doc` verbatim into the new file.
- Apply ONLY the three modifications described in STATE 5 (EXAMPLES update, prose corrections, OUTPUT subsection addition).
- Do NOT reformat, reorder, rewrite, or restructure any heading, paragraph, option list, or prose from `$doc` **unless it is factually contradicted by actual execution results**.
- Do NOT add new top-level sections (no ENVIRONMENT VARIABLES, no duplicate OUTPUT, etc.).
- Do NOT change section title casing, Markdown heading levels, or list syntax.
- If in doubt, leave the section exactly as it appears in `$doc`.

---

### Prerequisites — FASTQ FORMAT

A valid FASTQ record is **exactly 4 lines** in this order:

```
@<identifier> <optional description>
<nucleotide sequence>          ← MUST be non-empty (≥ 10 characters, A/T/G/C only)
+
<quality string>               ← MUST be the exact same length as the sequence line
```

Common mistakes **forbidden**:
- Writing `@header\n+\nquality` with the sequence line missing.
- Writing a quality string shorter or longer than the sequence.
- Mixing `>` (FASTA) and `@` (FASTQ) headers in the same file.
- Writing `~`-separated fields (e.g. `@seq002~description`) — use a space.
- Writing a quality string containing characters outside printable ASCII (33–126).

---

### OUTPUT FORMAT GUARD

OBITools4 determines the output format from **data content and explicit flags**, NOT from filename extension.

- If the example is meant to produce FASTA output from FASTQ input, the command MUST include `--fasta-output`.
- If the example is meant to produce FASTQ output from FASTA input, the command MUST include `--fastq-output`.
- Never assume an output format from the filename alone.
- Verify the actual format of each output file by checking its first character: `>` = FASTA, `@` = FASTQ, `[`/`{` = JSON.
- If the format is wrong, add the missing flag, update `$cmds_doc` and `$cmds_run`, and re-run.

---

### OPTION VALIDATION GUARD

Before writing any example command in STATE 2, explicitly cross-check each option against the OPTIONS section of `$doc`:

- Every flag used must appear in the OPTIONS section with the claimed semantics.
- Input-format flags (`--fasta`, `--fastq`, `--csv`, `--genbank`, `--embl`, `--ecopcr`) tell the tool how to **read** input — they do NOT affect output format.
- Output-format flags (`--fasta-output`, `--fastq-output`, `--json-output`) control **write** format.
- If an option needed for a working example is absent from `$doc`, mark that example as SKIP rather than inventing a flag.

---

### ANNOTATION RULES — CRITICAL

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

### CSV FILES FOR JOINS

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

### STATE 1 — Read the documentation file

**Input:** nothing.
**Action:** emit a single Read call.

```
<function=Read>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/cmd/obi{xxx}.md"}
</function>
```

**Output:** store content as `$doc`.
**Stop.** Do not interpret or summarise. Proceed to STATE 2.

---

### STATE 2 — Analyse examples and design input files

**Input:** `$doc`.
**Action (no tool calls):**

1. Extract every example command from the EXAMPLES section of `$doc`.
   - Identify every distinct input filename referenced.
   - Identify every option used and verify each against the OPTIONS section (OPTION VALIDATION GUARD).
   - **Skip any example that requires an external resource** (taxonomy database, remote URL, pre-existing output file). Mark it SKIP — it will be kept verbatim in the final doc without a `**Expected output:**` annotation.
   - **`--paired-with` examples:** `--paired-with` requires `--out` (standard output cannot be used). The command produces TWO output files `<prefix>_R1.ext` and `<prefix>_R2.ext`. Do NOT use `>` redirection for these. In STATE 4, read both `_R1` and `_R2` files.

2. **Coverage check — command-specific options:**
   From the OPTIONS section of `$doc`, list all command-specific options (excluding those covered by standard option-sets: input, output, common — see Phase 3 STATE 2 item 8 for the full list).
   Verify that every such option appears in at least one non-skipped example.
   If any option is not covered, **add an additional example** that exercises it before proceeding.

3. For each distinct input filename, design synthetic sequence content that:
   - Is **minimal** (≤ 20 sequences, each ≤ 300 bp).
   - Contains sequences that **will** produce output (positive cases) AND at least one that **will not** produce output (negative case), when the command filters sequences.
   - Exercises every option combination present in the non-skipped examples.
   - Uses realistic identifiers (`seq001`, `seq002`, …).

4. **File format rules (strictly enforced):**
   - **FASTA:** `>id description` header, then sequence on one or more lines (60 bp per line), ≥ 10 bp, A/T/G/C only.
   - **FASTQ:** exactly 4 lines per record. Before finalising, mentally verify each record:
     - Line 1 starts with `@`, has an identifier, optionally a space and description.
     - Line 2 is the nucleotide sequence (≥ 10 characters, A/T/G/C only).
     - Line 3 is exactly `+`.
     - Line 4 has **exactly the same number of characters** as line 2.
     If any record fails this check, fix it before proceeding.

5. Rewrite every non-skipped example command into two forms:
   - `$cmds_doc`: the bare command as it will appear in documentation — filenames only, **no `cd` prefix**.
   - `$cmds_run`: the same command prefixed with `cd /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx} &&`

**Output:** store file designs as `$files`, `$cmds_doc`, `$cmds_run`.
**Stop.** Proceed to STATE 3.

---

### STATE 3 — Write input files, validate, and run examples

**Step 3a — create input files (parallel):**
```
<function=Write>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx}/FILENAME", "content": "..."}
</function>
```

**Step 3b — validate input files:**
```
<function=Bash>
{"command": "cd /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx} && python3 -c \"\nimport sys\nfor fname in $(echo FILENAMES):\n    lines = open(fname).readlines()\n    if fname.endswith('.fastq'):\n        assert len(lines) % 4 == 0, f'{fname}: line count not multiple of 4'\n        for i in range(0, len(lines), 4):\n            hdr, seq, plus, qual = lines[i:i+4]\n            assert hdr.startswith('@'), f'{fname} record {i//4+1}: header must start with @'\n            seq = seq.rstrip(); qual = qual.rstrip()\n            assert len(seq) >= 10, f'{fname} record {i//4+1}: sequence too short ({len(seq)})'\n            assert len(seq) == len(qual), f'{fname} record {i//4+1}: seq len {len(seq)} != qual len {len(qual)}'\n    elif fname.endswith('.fasta') or fname.endswith('.fa'):\n        assert lines[0].startswith('>'), f'{fname}: first line must start with >'\n        seq_len = sum(len(l.rstrip()) for l in lines[1:] if not l.startswith('>'))\n        assert seq_len >= 10, f'{fname}: total sequence length too short ({seq_len})'\nprint('All input files valid')\n\" 2>&1; echo EXIT:$?"}
</function>
```

If validation fails (EXIT non-zero): fix the offending file(s) and re-run. Do NOT proceed until validation passes.

**Step 3c — run examples (sequential, one Bash call at a time):**

For each non-skipped example, emit the run command, wait for the result, then immediately verify.

Run:
```
<function=Bash>
{"command": "cd /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx} && COMMAND 2>&1; echo EXIT:$?"}
</function>
```

After each EXIT:0, verify the output file exists and is non-empty:
```
<function=Bash>
{"command": "ls -la /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx}/OUTPUT_FILE && head -c 200 /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx}/OUTPUT_FILE"}
</function>
```

Check output format from the first character: `>` = FASTA, `@` = FASTQ, `[`/`{` = JSON.
If the format does not match expectation, add the missing `--fasta-output` / `--fastq-output` / `--json-output` flag, update `$cmds_doc` and `$cmds_run`, and re-run.

If a command fails (EXIT non-zero): diagnose, fix, update `$cmds_doc` and `$cmds_run`, and re-run.
Do NOT proceed to STATE 4 until all non-skipped commands have EXIT:0 and verified non-empty output files.

**Output:** store per-command results as `$runs`.
**Stop.** Proceed to STATE 4.

---

### STATE 4 — Read output files

**Input:** `$runs` (output file paths from STATE 3).
**Action:** emit one Read call per output file successfully produced (EXIT:0), all in a single parallel message.

Do NOT re-run commands — read only files already generated in STATE 3.

```
<function=Read>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx}/OUTPUT_FILE"}
</function>
```

**Output:** store contents as `$outputs`.
**Stop.** Proceed to STATE 5.

---

### STATE 5 — Update the documentation file

**Input:** `$doc`, `$runs`, `$outputs`, `$cmds_doc`.

Re-read DOCUMENT PRESERVATION at the top before writing. Apply ONLY these modifications:

#### Modification 1 — EXAMPLES section

For each non-skipped example:
- Replace the original command with the `$cmds_doc` form (no `cd` prefix).
- Keep the one-line biological use-case comment unchanged.
- Add `**Expected output:**` on its own line **after** the closing triple-backtick of the code block:
  ```
  **Expected output:** N sequences written to `out_name.fasta`.
  ```
  where N = number of lines starting with `>` or `@` in the corresponding `$outputs` entry.

For skipped examples: keep them exactly as they are in `$doc`, no annotation added.

#### Modification 2 — Prose corrections (if any factual contradiction)

Fix only specific sentences that are contradicted by actual execution results. Examples of things to correct:
- An attribute name that differs from actual output.
- An output format claimed that differs from the actual format observed.
- A default value stated incorrectly.
- A claim about which sequences are selected/discarded that contradicts observed results.

After each corrected passage, add: `<!-- corrected: <brief reason> -->`
Do NOT "improve" text that is merely incomplete — only fix outright contradictions.

#### Modification 3 — OUTPUT section

Find the `# OUTPUT` section and append at its very end (before the next `---` or `#`):

```markdown
## Observed output example

```
<verbatim excerpt — first ≤ 10 sequences from the first successful $outputs entry>
```
```

Rules:
- The excerpt is copied byte-for-byte from `$outputs`. No editing, no truncation within a record.
- Do NOT duplicate the OUTPUT section. There must be exactly one `# OUTPUT` heading.
- If no output was successfully produced, omit this subsection entirely.

```
<function=Write>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/cmd/obi{xxx}.md", "content": "..."}
</function>
```

**Stop.** Proceed to PHASE 3.

---

## PHASE 3 — Generate Hugo documentation

(Equivalent to prompt_hugo.md)

### HUGO SHORTCODE REFERENCE

Use only the shortcodes listed below. Never invent others.

| Shortcode | Syntax | Effect |
|-----------|--------|--------|
| Command link | `{{< obi obi{xxx} >}}` | Renders command name as internal link |
| Format name | `{{% fasta %}}` `{{% fastq %}}` `{{% csv %}}` `{{% json %}}` `{{% yaml %}}` | Renders format name (use in prose) |
| Suite name | `{{% obitools4 %}}` | Renders "OBITools4" as styled text |
| Embed data file | `{{< code "FILENAME" FORMAT true >}}` | Embeds file content; FORMAT = `fasta`, `fastq`, `txt`, `csv`, `json`, `yaml` |
| Standard option set | `{{< option-sets/input >}}` | Renders shared input-format options |
| Standard option set | `{{< option-sets/output >}}` | Renders shared output-format options |
| Standard option set | `{{< option-sets/common >}}` | Renders shared performance/logging options |
| Standard option set | `{{< option-sets/selection >}}` | Renders shared sequence-selection options (obigrep only) |
| Single shared option | `{{< cmd-options/paired-with >}}` | Renders `--paired-with` option |
| Custom option block | `{{< cmd-option name="NAME" short="S" param="PARAM" >}}` text `{{< /cmd-option >}}` | Renders a command-specific option; `short` and `param` are optional |
| Workflow diagram | `{{< mermaid class="workflow" >}}` … `{{< /mermaid >}}` | Renders Mermaid flowchart |

---

### SECTION STRUCTURE OF A HUGO COMMAND PAGE

```markdown
---
archetype: "command"
title: "obi{xxx}"
date: YYYY-MM-DD
command: "obi{xxx}"
category: <category>
url: "/obitools/obi{xxx}"
weight: <weight>
---

# `obi{xxx}`: <one-line description>

> [!WARNING] Preliminary AI-generated documentation
> This page was automatically generated by an AI assistant and has **not yet been
> reviewed or validated** by the {{% obitools4 %}} development team. It may contain
> inaccuracies or incomplete information. Use with caution and refer to the command's
> `--help` output for authoritative option descriptions.

## Description

<narrative prose — 2–5 paragraphs, uses {{< obi >}} and {{% format %}} shortcodes>
<workflow diagram>
<data files with {{< code >}} shortcodes and paired command+output blocks>

## Synopsis

```bash
obi{xxx} [--option1] [--option2|-s PARAM] ... [<args>]
```

## Options

#### {{< obi obi{xxx} >}} specific options

- {{< cmd-option name="NAME" short="S" param="PARAM" >}}
  Description of option.
  {{< /cmd-option >}}

#### Taxonomic options        ← include only if command uses taxonomy

- {{< cmd-options/taxonomy/taxonomy >}}

{{< option-sets/input >}}

{{< option-sets/output >}}   ← omit if command has no output (e.g. obicount)

{{< option-sets/common >}}

## Examples

...

```bash
obi{xxx} --help
```
```

**YAML front matter fields:**
- `archetype`: always `"command"`.
- `title` and `command`: the command name.
- `date`: today's date in `YYYY-MM-DD` format.
- `category`: the subdirectory name under `commands/`.
- `url`: always `/obitools/obi{xxx}`.
- `weight`: copy from the existing `_index.md` if it exists; otherwise use `50`.

---

### STATE 1 — Read source material (parallel)

**Input:** nothing.
**Action:** emit all of the following calls in a single parallel message.

```
<function=Read>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/cmd/obi{xxx}.md"}
</function>
<function=Bash>
{"command": "ls /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx}/ 2>/dev/null || echo NO_EXAMPLES"}
</function>
<function=Read>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4-doc/content/docs/commands/basics/obi{xxx}/_index.md"}
</function>
<function=Bash>
{"command": "find /Users/coissac/Sync/travail/__MOI__/GO/obitools4-doc/content/docs/commands -type d -name 'obi{xxx}'"}
</function>
```

**Output:** store as `$doc`, `$examples_list`, `$existing_hugo`, `$category_path`.
**Stop.** Do not interpret. Proceed to STATE 2.

---

### STATE 2 — Determine category and plan content

**Input:** `$doc`, `$examples_list`, `$existing_hugo`, `$category_path`.

1. **Category:**
   - If `$category_path` found a directory, extract the category name (segment between `commands/` and `obi{xxx}`).
   - If `$existing_hugo` contains `category:`, use that value.
   - Otherwise, default to `basics`.

2. **Weight:**
   - If `$existing_hugo` contains `weight:`, reuse that value.
   - Otherwise, use `50`.

3. **Example files to copy:** input files AND output files from the working examples. Skip compressed files (`.gz`).

4. **Naming convention for output files:** Use simple names like `out.json`, `out.yaml`, `out.fasta`, `out.fastq` — NOT `out_json.json` or similar.

5. **Command-output file consistency:** Each example command MUST produce the file shown below it. Verify that the flag in the command creates the displayed file.

6. **Plan replacements:**
   - `obi{xxx}` → `{{< obi obi{xxx} >}}`
   - Format names in prose → `{{% fasta %}}`, etc.
   - Input filenames → `{{< code "FILENAME" FORMAT true >}}`

7. **Workflow diagram consistency:** The Mermaid diagram MUST use the exact same files as the first working example.

8. **Options section plan — standard option-sets coverage:**
   The following options are covered by standard option-sets and must NOT be re-documented:
   - `{{< option-sets/input >}}`: `--fasta`, `--fastq`, `--embl`, `--genbank`, `--ecopcr`, `--csv`, `--input-OBI-header`, `--input-json-header`, `--u-to-t`, `--solexa`, `--skip-empty`, `--no-order`.
   - `{{< option-sets/output >}}`: `--fasta-output`, `--fastq-output`, `--json-output`, `--output-OBI-header`/`-O`, `--output-json-header`, `--out`/`-o`, `--compress`/`-Z`.
   - `{{< option-sets/common >}}`: `--max-cpu`, `--batch-size`, `--batch-size-max`, `--batch-mem`, `--no-progressbar`, `--debug`, `--silent-warning`, `--pprof`, `--pprof-goroutine`, `--pprof-mutex`, `--version`, `--help`.
   - `--paired-with` → `{{< cmd-options/paired-with >}}`.
   - Taxonomy options (`--taxonomy`, `--fail-on-taxonomy`, `--raw-taxid`, `--update-taxid`, `--with-leaves`) → grouped under `#### Taxonomic options` with `{{< cmd-options/taxonomy/taxonomy >}}` for `--taxonomy`; document the rest with `{{< cmd-option >}}` blocks.
   - All remaining command-specific options → `{{< cmd-option >}}` blocks under `#### {{< obi obi{xxx} >}} specific options`.

9. **Examples:** keep only examples whose input files exist in `$examples_list`. Add final `obi{xxx} --help`.

10. **Remove duplicates:** If the same files appear in both Description and Examples sections, keep only the first occurrence.

**Output:** store as `$plan`.
**Stop.** Proceed to STATE 3.

---

### STATE 3 — Read example input files (parallel)

Emit one Read call per file to be used in the Hugo page (both input and output files). Do not read compressed files (`.gz`).

```
<function=Read>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx}/FILENAME"}
</function>
```

**Output:** store file contents as `$input_files`.
**Stop.** Proceed to STATE 4.

---

### STATE 4 — Write Hugo files (parallel)

**Step 4a — write `_index.md`** following SECTION STRUCTURE and CONTENT RULES below.

**CRITICAL:** The Synopsis section MUST use the **full verbatim** synopsis from `$doc`, not a simplified version.

**Step 4b — copy data files (parallel with 4a):**
```
<function=Bash>
{"command": "cp /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx}/*.fasta /Users/coissac/Sync/travail/__MOI__/GO/obitools4-doc/content/docs/commands/<category>/obi{xxx}/ 2>/dev/null || true"}
</function>
<function=Bash>
{"command": "cp /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx}/*.fastq /Users/coissac/Sync/travail/__MOI__/GO/obitools4-doc/content/docs/commands/<category>/obi{xxx}/ 2>/dev/null || true"}
</function>
<function=Bash>
{"command": "cp /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx}/*.csv /Users/coissac/Sync/travail/__MOI__/GO/obitools4-doc/content/docs/commands/<category>/obi{xxx}/ 2>/dev/null || true"}
</function>
<function=Bash>
{"command": "cp /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx}/*.json /Users/coissac/Sync/travail/__MOI__/GO/obitools4-doc/content/docs/commands/<category>/obi{xxx}/ 2>/dev/null || true"}
</function>
<function=Bash>
{"command": "cp /Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/examples/obi{xxx}/*.yaml /Users/coissac/Sync/travail/__MOI__/GO/obitools4-doc/content/docs/commands/<category>/obi{xxx}/ 2>/dev/null || true"}
</function>
```

Skip compressed files (`.gz`).

**Step 4c — delete unused files from Hugo directory:**
```
<function=Bash>
{"command": "ls -la /Users/coissac/Sync/travail/__MOI__/GO/obitools4-doc/content/docs/commands/<category>/obi{xxx}/"}
</function>
```
Compare with `$examples_list` and remove any files not present in the working examples.

---

## CONTENT RULES (apply throughout Phase 3 STATE 4)

### Workflow diagram

Place in the Description section, after introductory prose and before the first `{{< code >}}` block.

```
{{< mermaid class="workflow" >}}
graph TD
  A@{ shape: doc, label: "input_file.fastq" }
  C[obi{xxx}]
  D@{ shape: doc, label: "output_file.fasta" }
  A --> C:::obitools
  C --> D
  classDef obitools fill:#99d57c
{{< /mermaid >}}
```

Rules:
- One `@{ shape: doc, label: "FILENAME" }` node per input file; use filenames from the first example.
- One output node for the output file of the same example.
- Apply `:::obitools` on the **last arrow pointing to the command node**, not on the node definition line.
- `classDef obitools fill:#99d57c` must always be the last line inside the block.
- If the command produces no file output (prints to stdout only), use a terminal node `D([stdout])`.

### Description section

Narrative prose teaching the reader how the command works, one concept at a time. These are NOT the same examples as in the Examples section — simpler, focused on a single behaviour, chosen to clarify specific options or edge cases.

- Write flowing paragraphs, not bullet lists of options.
- Explain **why** a biologist would use the command and **what** it does.
- Introduce data files with `{{< code "FILENAME" FORMAT true >}}` before the first command that uses them.
- Show example commands and their output in **paired** fenced blocks — no `**Expected output:**` label:
  ````markdown
  ```bash
  obi{xxx} [options] input_file
  ```
  ```
  <actual output lines>
  ```
  ````
- Replace tool name in prose with `{{< obi obi{xxx} >}}`.
- Replace format names in prose with `{{% fasta %}}`, `{{% fastq %}}`, etc.
- **Do NOT reuse** examples from the Examples section verbatim. Description examples are simpler and pedagogical.

### Synopsis section

Use the synopsis from `$doc` verbatim. Wrap in a `bash` fenced code block.

### Options section

- Only document options **not** covered by `{{< option-sets/... >}}` (see STATE 2 item 8).
- Group under `#### {{< obi obi{xxx} >}} specific options`, then `#### Taxonomic options` if applicable, then the three `{{< option-sets/... >}}`.

### Examples section

Practical, real-world recipes. Each example addresses a distinct use case not already shown in Description.

- **Never duplicate** an example from the Description section.
- Every example that produces sequence or annotation file output uses this pattern:
  1. Short intro paragraph (2–4 sentences) with a Markdown hyperlink to the input file, e.g. `The file [input.fasta](input.fasta) contains …`.
  2. `{{< code "input_file" FORMAT true >}}`
  3. `bash` fenced block with `-o out_name.ext` (never `>` redirection for non-paired examples).
  4. `{{< code "out_name.ext" FORMAT true >}}`
- **`--paired-with` examples:** use `--out <prefix>.fastq` (not `>`), show both output files:
  ````markdown
  ```bash
  obi{xxx} --paired-with reverse.fastq --out out_paired.fastq forward.fastq
  ```
  {{< code "out_paired_R1.fastq" fastq true >}}
  {{< code "out_paired_R2.fastq" fastq true >}}
  ````
  Both `_R1` and `_R2` files must be copied to the Hugo command directory (Step 4b).
- **CSV output:** pipe through `csvlook`, no file redirection, no `{{< code >}}`:
  ````markdown
  ```bash
  obi{xxx} [options] input_file | csvlook
  ```
  ```
  | col1 | col2 |
  | ---- | ---- |
  | val1 | val2 |
  ```
  ````
- Last example always: `` ```bash\nobi{xxx} --help\n``` `` (no output block).
- Never inline file content as raw fenced blocks — always use `{{< code >}}`.
- Output files must be copied to the Hugo command directory alongside input files (Step 4b).
