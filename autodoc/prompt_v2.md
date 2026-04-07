# Task

Produce `autodoc/cmd/obi{xxx}.md` — the documentation file for the `obi{xxx}` CLI command.

---

## TOOL CALL FORMAT — enforce before every call

A tool call is exactly:

    <function=tool_name>
    {"param": "value"}
    </function>

Rules (no exceptions):
- `<` is immediately followed by `f` — zero spaces, zero characters in between.
- Parameters are a **single JSON object** — no XML tags, no `<parameter=...>`, no `</parameter>`, no `<description>`.
- No outer wrapper — never use `<tool_call>`, `<tool_use>`, or any other enclosing tag.
- Tool name is lowercase with double underscores — never ALL_CAPS, never single underscore between server and tool name.

CORRECT:  `<function=mcp__jcodemunch__get_file_outline>`
WRONG:    `< function=mcp__jcodemunch__get_file_outline >` ← spaces → parse error
WRONG:    `<function=mcp_jcodemunch_get_file_outline>`     ← wrong separator
WRONG:    `<function=MCP__JCODEMUNCH__GET_FILE_OUTLINE>`   ← wrong casing

---

## HALLUCINATION GUARD — enforce before writing anything

OBITools4 is a **complete rewrite**. Training data about OBITools v1/v2/v3 is wrong for this version.

Before writing any sentence, apply this check:

> "Can I point to the exact line in $help or $docs that justifies this claim?"
> If NO → do not write it.

This applies to: option names, option flags, default values, file formats, behaviours, algorithms, output fields.
Omit rather than guess. A shorter correct page is better than a longer hallucinated one.

---

## PIPELINE

Execute the four states below in order. Do not skip states. Do not merge states.

---

### STATE 1 — Gather raw data (parallel)

**Input:** nothing.
**Action:** emit all of the following tool calls in a single message (parallel execution).

Call 1 — dependencies of the main entry point:
```
<function=mcp__treesitter__treesitter_get_dependencies>
{"language": "go", "file_path": "cmd/obitools/obi{xxx}/main.go"}
</function>
```

Call 2 — dependencies of every file in the command package (one call per file found by glob `pkg/obitools/obi{xxx}/*.go`):
```
<function=mcp__treesitter__treesitter_get_dependencies>
{"language": "go", "file_path": "pkg/obitools/obi{xxx}/FILE.go"}
</function>
```

Call 3 — symbol outline of the command package (single batch call):
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

Call 5 — list of already-documented OBITools4 commands (to inform SEE ALSO):
```
<function=WebFetch>
{"url": "https://obitools4.metabarcoding.org/obitools/"}
</function>
```

**Output:** store results as `$deps`, `$outline`, `$help`, `$web_doc`.
`$web_doc` contains the index of documented commands — use it only to determine which
`obi*` commands have existing documentation pages, so that SEE ALSO only links to pages
that actually exist. Do NOT use it as a source for option names, flags, or defaults.
**Stop.** Do not interpret, summarise, or write anything. Proceed to STATE 2.

---

### STATE 2 — Resolve documentation files

**Input:** `$deps` (import paths from STATE 1).
**Action (no tool calls):**

1. Collect every import path that starts with `git.metabarcoding.org/obitools/obitools4/obitools4/pkg/`.
2. Remove these infrastructure packages (already covered by the convert doc):
   - `pkg/obidefault`, `pkg/obiiter`, `pkg/obiformats`, `pkg/obioptions`, `pkg/obiutils`, `pkg/obiparams`
   - `pkg/obiseq` — keep only if `$outline` shows non-trivial sequence manipulation.
3. Always add: `pkg/obitools/obi{xxx}` and `pkg/obitools/obiconvert`.
4. Map each remaining package path to its doc file:
   - take all path segments from `pkg/` onward (inclusive), replace `/` with `_` → `autodoc/docmd/<joined_segments>.md`
   - examples: `git.metabarcoding.org/.../pkg/obialign` → `autodoc/docmd/pkg_obialign.md`
   - examples: `git.metabarcoding.org/.../pkg/obitools/obiuniq` → `autodoc/docmd/pkg_obitools_obiuniq.md`

**Output:** store file list as `$docfiles`.
**Stop.** Proceed to STATE 3.

---

### STATE 3 — Read documentation files (parallel)

**Input:** `$docfiles`.
**Action:** emit one `Read` call per file in `$docfiles`, all in a single parallel message.

```
<function=Read>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/DOCFILE"}
</function>
```

Do NOT read any `.go` source file.
Do NOT produce any summary or analysis.

**Output:** store all file contents as `$docs`.
**Stop.** Proceed to STATE 4.

---

### STATE 4 — Write the documentation file

**Input:** `$help`, `$docs`, `$outline`, `$web_doc`.
**Action:** fill the template below, then emit exactly one `Write` call. Then stop.

**Source discipline:** every piece of information in the template MUST come from the labelled source.
If the source does not contain the information, write `_(not available)_` — never invent.
`$web_doc` (index of documented commands) is used exclusively to filter the SEE ALSO
section — only list commands that appear as documented pages in `$web_doc`.
It is NOT a source for option names, descriptions, or behaviour.

---

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

[FILL: output format and what fields/attributes are added or changed. Source: $help + $docs.]

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
3. COVERAGE RULE: every command-specific option documented in the OPTIONS section MUST
   appear in at least one example. Add extra examples if needed to achieve full coverage.
Source for flags and options: $help only.]

---

# SEE ALSO

[FILL: related obi commands mentioned in $docs or $help, AND present in $web_doc
(i.e. commands that have an existing documentation page online).
If none qualify: omit section.]

---

# NOTES

[FILL: caveats, performance notes, known limitations. Source: $docs only.
If none: omit section.]
```

---

Emit the Write call:
```
<function=Write>
{"file_path": "/Users/coissac/Sync/travail/__MOI__/GO/obitools4/autodoc/cmd/obi{xxx}.md", "content": "..."}
</function>
```

**Stop. Do not emit any text after the Write call.**
