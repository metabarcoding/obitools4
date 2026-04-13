# Objective

Fully document OBITools (version 4, written in Go) in English, using a 4‑phase incremental pipeline.

You **MUST** use the available MCP servers:

- `cclsp` – exact definitions, references, diagnostics  
- `jcodemunch` – code indexing, symbol extraction  
- `treesitter` – AST and CLI parsing  
- `context7` – external documentation  

All tool calls must follow the exact API described in the MCP server documentation. If a required tool is unavailable, you **MUST** log the error and stop execution.

### Tool call format (CRITICAL)

Tool calls **MUST** use this exact XML format — no spaces inside the angle brackets:

```
<function=tool_name>
{"param": "value"}
</function>
```

**FORBIDDEN** — these variants will cause parse errors and must NEVER be used:
- `< function=tool_name >` (spaces around the tag name)
- `< function = tool_name>` (spaces around `=`)
- `<function = tool_name>` (space before `=`)

The opening tag is `<function=tool_name>` with **zero spaces** inside `<` and `>`.

---

# Global rules

** You are not allowed to read twice the same file in a row. **

## Language

- All generated documentation **MUST** be in English.  
- If an existing documentation file is in French:  
  1. Translate it to English  
  2. Save the original as `.fr.md` **before** overwriting  
  3. Write the new English version  

---

## Execution mode (STRICT)

You are operating in **STRICT TOOL MODE**:

- If a file must be written, you **MUST** use the `Shell` tool.
- You **MUST NOT** read entire directory listings into memory.
- You **MUST** work with **one item at a time** using a simple text file as a task queue.

### Reading files before writing

- **Before writing to an existing documentation file**, you must first read it using the `Read` tool.
- **When documenting a single Go source file**, you only need to read that one file (plus up to 4-5 helper files if needed for context).
- Do NOT read the entire codebase - only what is necessary to document the current file.

---

### Rules

- Always write the **full** file (no partial updates).  
- Paths are relative to the project root; directories are created implicitly.  
- Content must be valid UTF‑8; use `\n` line endings.  
- Do **not** wrap content in backticks.  

---

## Progress tracking: task queue files

We use **line‑oriented task files** to avoid loading large lists into memory. Each phase has its own task file:

- `docs/todo/phase1.txt` – list of Go files (one per line) to document.  
- `docs/todo/phase1bis.txt` – same list, but after phase1 is done.  
- `docs/todo/phase2.txt` – list of packages.  
- `docs/todo/phase3.txt` – list of tools.

**How it works:**

1. At the start of a phase, if the task file does not exist, it is created by scanning the codebase once (Phase 0 or Phase X init).  
2. **Each run of the LLM processes only the first line of the task file.**  
3. After processing the item (success or permanent failure), the line is removed from the task file.  
   - On success, the line is deleted (no extra sentinel file needed).  
   - On transient failure (retry < 3), we keep the line but increment a retry counter stored in a separate file.  
   - On permanent failure (retry ≥ 3), we move the line to a `failed.txt` file and log the error.  
4. The LLM then exits (or continues if the task file is still non‑empty, but it must never load more than one line).

This way, the LLM’s context never holds more than a single task at a time.

### Retry mechanism

For each item (e.g., `internal/align/align.go`), we maintain a retry counter in:

- `docs/retry/phase1/internal/align/align.go.count`

If the file does not exist, retries = 0.  
Each time processing fails, we increment the counter (write the new number).  
If after increment the counter < 3, we keep the line in the task file.  
If counter reaches 3, we **remove the line from the task file**, add it to `docs/failed/phase1/internal/align/align.go.failed` (just a marker), and log the error.

---

## Documentation quality requirements (CRITICAL)

Documentation MUST NOT be superficial. For each documented element (file, function, struct, package):

###  You MUST explain:

- what it does
- why it exists (context, problem solved)
- how it is used
- assumptions and preconditions
- possible edge cases

### Forbidden patterns

- Vague phrases like “This function handles…”, “Utility for…”, “Helper function…”.
- Generic descriptions that could apply to any project.

### Required content per element type

- Functions:
  - Purpose
  - Parameter meaning
  - Return values
  - Notable behaviour (panic conditions, side effects, concurrency)
- Structs:
  - Role in the system
  - Meaning of key fields
- Files:
  - Role within the package
  - Interactions with other files

### Anti‑generic rule

If the description could apply to any project, it is INVALID. You MUST include domain‑specific context (bioinformatics, sequence processing, etc.) and concrete behaviour.

### Quality validation

Before marking an item as done (i.e., creating the .done sentinel), you MUST perform a self‑validation:

- Check that all required sections are present.
- Verify that no forbidden patterns remain.

If validation fails, increment the retry counter and keep the item pending.


---

# Directory structure

```
docs/
  todo/                          # task queues
    phase1.txt
    phase1bis.txt
    phase2.txt
    phase3.txt
  retry/                         # retry counters
    phase1/                      # mirrors file structure
      internal/align/align.go.count
    phase1bis/
    phase2/
    phase3/
  failed/                        # permanent failure markers
    phase1/
      internal/align/align.go.failed
    phase1bis/
    phase2/
    phase3/
  phase1/                        # actual documentation
    <relative_path>/<file>.go.md
  phase2/
    <package>.md
  phase3/
    <tool>.md
  error.log
```

---

# Phase 0: Initialization

1. Ensure required directories exist: `docs/todo`, `docs/retry`, `docs/failed`, `docs/phase1`, `docs/phase2`, `docs/phase3`.  
2. **If `docs/todo/phase1.txt` does not exist**:  
   - Use `find pkg -name "*.go" ! -name "*_test.go" ! -path "*/cmd/*"` to list all Go files (excluding tests and main.go).  
   - Write the list (one relative path per line, e.g., `internal/align/align.go`) to `docs/todo/phase1.txt`.  
3. Do the same for phase2 and phase3 later when those phases start.  
4. **No other state is stored.**

---

# Phase 1: File documentation

**Processing rule:**
- Read the **first line** of `docs/todo/phase1.txt` (using `head -n 1`).  
- If the file is empty, Phase 1 is complete → proceed to Phase 1bis initialization.  
- Otherwise, process that single file.

**Processing a file:**

1. Let `relpath` be the line content (e.g., `internal/align/align.go`).  
2. Check if a permanent failure marker exists at `docs/failed/phase1/${relpath}.failed`. If yes, remove the line from the task file and skip (line will be deleted). 
3. If the documentation file `docs/phase1/${relpath}.go.md` exists go directly to its validation (step 6).
4. Otherwise, generate documentation for that file (using MCP tools as before).  
5. Write the documentation to `docs/phase1/${relpath}.go.md`.  
6. Validate quality.  
7. If validation succeeds:  
   - Remove the line from the task file.  
   - Remove any retry counter file for this item.  
   - (No sentinel needed; the removal from todo indicates completion.)  
8. If validation fails:  
   - Increment retry counter:  
     - If `docs/retry/phase1/${relpath}.count` does not exist, set to 1.  
     - Else read it, add 1, write back.  
   - If new counter >= 3:  
     - Remove line from task file.  
     - Create `docs/failed/phase1/${relpath}.failed`.  
     - Log error.  
   - If new counter < 3:  
     - Keep the line in the task file (do nothing, it stays as first line for next run).  
9. **Exit** (or stop if this was a single run). The next invocation will read the first line again (same if retry, or next if removed).

**Important:**  
- Do **not** read more than one line.  
- Do **not** attempt to process multiple items in one run.  
- The LLM should finish after handling one item.

---

# Phase 1bis: Review and harmonization

When Phase 1 is complete (i.e., `docs/todo/phase1.txt` empty), we initialize `docs/todo/phase1bis.txt` with the same list of files (the ones that succeeded).  
But note: we need to know which files were successfully documented. Since we removed lines from `phase1.txt` on success, we need a record. The simplest is to reuse the same list but we can generate it by listing the existing `.go.md` files in `docs/phase1/` (since every successful file has a `.go.md`).  
Thus, Phase 1bis initialization:

- If `docs/todo/phase1bis.txt` does not exist, create it by listing all `.go.md` files under `docs/phase1/`, stripping the `docs/phase1/` prefix and the `.go.md` suffix, and writing the relative path (same format as phase1).  

Then processing is identical to Phase 1, but using `docs/todo/phase1bis.txt` and output is overwriting the same `.go.md` files (with improvements). Retry counters go in `docs/retry/phase1bis/`.

---

# Phase 2: Package documentation

When Phase 1bis is complete (`docs/todo/phase1bis.txt` empty), initialize `docs/todo/phase2.txt`:

- List all packages: unique directories under `pkg/` that contain at least one `.go` file and are not tools.  
- Write each package identifier (e.g., `align`, `internal/align`) as a line.

Processing: read first line, generate `docs/phase2/<package>.md`, validate, remove line on success, retry logic in `docs/retry/phase2/`.

---

# Phase 3: Tool documentation

When Phase 2 complete, initialize `docs/todo/phase3.txt`:

- List all directories under `cmd/` that contain a `main.go`. Write each tool name as a line.

Processing: read first line, generate `docs/phase3/<tool>.md`, validate, remove line on success, retry logic in `docs/retry/phase3/`.

---

# Finalization

When all task files are empty and no pending phases, generate `docs/README.md` by:

- Listing all package docs (files in `docs/phase2/`) and linking.  
- Listing all tool docs (files in `docs/phase3/`) and linking.  

Write using `Shell`.

---

# Execution flow summary

1. **Phase 0**: Create directories and initial `todo/phase1.txt` if missing. Exit.  
2. **Phase 1**:  
   - If `todo/phase1.txt` exists and non‑empty → process first line.  
   - Else → move to Phase 1bis initialization.  
3. **Phase 1bis**:  
   - If `todo/phase1bis.txt` does not exist → create from successful phase1 docs.  
   - If non‑empty → process first line.  
   - Else → move to Phase 2 initialization.  
4. **Phase 2**: similar.  
5. **Phase 3**: similar.  
6. **Finalization**: generate README.

The LLM should be invoked repeatedly (e.g., by a scheduler) until all phases are done. Each invocation processes exactly one item.

---

# Important reminders

- Always call `Shell` to write files; never output content in plain text.  
- Validate quality before removing a line from the task file.  
- Log all failures to `docs/error.log` in JSON lines format.  
- If any MCP tool fails, treat as failure and increment retry counter.  
- Never read more than one line from a task file in a single run.
