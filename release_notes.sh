#!/bin/bash

# Generate GitHub-compatible release notes for an OBITools4 version.
#
# Usage:
#   ./release_notes.sh                 # latest version
#   ./release_notes.sh -v 4.4.15       # specific version
#   ./release_notes.sh -l              # list available versions
#   ./release_notes.sh -r              # raw commit list (no LLM)
#   ./release_notes.sh -c -v 4.4.16   # show LLM context for a version

GITHUB_REPO="metabarcoding/obitools4"
GITHUB_API="https://api.github.com/repos/${GITHUB_REPO}"
VERSION=""
LIST_VERSIONS=false
RAW_MODE=false
CONTEXT_MODE=false
LLM_MODEL="ollama:qwen3-coder-next:latest"

# ── Helpers ──────────────────────────────────────────────────────────────

die() { echo "Error: $*" >&2; exit 1; }

display_help() {
  cat <<EOF
Usage: $(basename "$0") [OPTIONS]

Generate GitHub-compatible Markdown release notes for an OBITools4 version.

Options:
  -v, --version VERSION   Target version (e.g., 4.4.15). Default: latest.
  -l, --list              List all available versions and exit.
  -r, --raw               Output raw commit list without LLM summarization.
  -c, --context           Show the exact context (commits + prompt) sent to the LLM.
  -m, --model MODEL       LLM model for orla (default: $LLM_MODEL).
  -h, --help              Display this help message.

Examples:
  $(basename "$0")                  # release notes for the latest version
  $(basename "$0") -v 4.4.15       # release notes for a specific version
  $(basename "$0") -l              # list versions
  $(basename "$0") -r -v 4.4.15    # raw commit log for a version
  $(basename "$0") -c -v 4.4.16    # show LLM context for a version
EOF
}

# Fetch all Release tags from GitHub API (sorted newest first)
fetch_versions() {
  curl -sf "${GITHUB_API}/releases" \
    | grep '"tag_name":' \
    | sed -E 's/.*"tag_name": "Release_([0-9.]+)".*/\1/' \
    | sort -V -r
}

# ── Parse arguments ──────────────────────────────────────────────────────

while [ "$#" -gt 0 ]; do
  case "$1" in
    -v|--version)  VERSION="$2"; shift 2 ;;
    -l|--list)     LIST_VERSIONS=true; shift ;;
    -r|--raw)      RAW_MODE=true; shift ;;
    -c|--context)  CONTEXT_MODE=true; shift ;;
    -m|--model)    LLM_MODEL="$2"; shift 2 ;;
    -h|--help)     display_help; exit 0 ;;
    *)             die "Unsupported option: $1" ;;
  esac
done

# ── List mode ────────────────────────────────────────────────────────────

if [ "$LIST_VERSIONS" = true ]; then
  echo "Available OBITools4 versions:" >&2
  echo "==============================" >&2
  fetch_versions
  exit 0
fi

# ── Resolve versions ─────────────────────────────────────────────────────

all_versions=$(fetch_versions)
[ -z "$all_versions" ] && die "Could not fetch versions from GitHub"

if [ -z "$VERSION" ]; then
  VERSION=$(echo "$all_versions" | head -1)
  echo "Using latest version: $VERSION" >&2
fi

tag_name="Release_${VERSION}"

# Verify the requested version exists
if ! echo "$all_versions" | grep -qx "$VERSION"; then
  die "Version $VERSION not found. Use -l to list available versions."
fi

# Find the previous version (the one right after in the sorted-descending list)
previous_version=$(echo "$all_versions" | grep -A1 -x "$VERSION" | tail -1)

if [ "$previous_version" = "$VERSION" ] || [ -z "$previous_version" ]; then
  previous_tag=""
  echo "No previous version found -- will include all commits for $tag_name" >&2
else
  previous_tag="Release_${previous_version}"
  echo "Generating notes: $previous_tag -> $tag_name" >&2
fi

# ── Fetch commit messages between tags via GitHub compare API ────────────

if [ -n "$previous_tag" ]; then
  commits_json=$(curl -sf "${GITHUB_API}/compare/${previous_tag}...${tag_name}")
  if [ -z "$commits_json" ]; then
    die "Could not fetch commit comparison from GitHub"
  fi
  commit_list=$(echo "$commits_json" \
    | jq -r '.commits[] | (.sha[:8] + " " + (.commit.message | split("\n")[0]))' 2>/dev/null)
else
  # First release: get commits up to this tag
  commits_json=$(curl -sf "${GITHUB_API}/commits?sha=${tag_name}&per_page=50")
  if [ -z "$commits_json" ]; then
    die "Could not fetch commits from GitHub"
  fi
  commit_list=$(echo "$commits_json" \
    | jq -r '.[] | (.sha[:8] + " " + (.commit.message | split("\n")[0]))' 2>/dev/null)
fi

if [ -z "$commit_list" ]; then
  die "No commits found between $previous_tag and $tag_name"
fi

# ── LLM prompt (shared by context mode and summarization) ────────────────

LLM_PROMPT="Summarize the following commits into a GitHub release note for version ${VERSION}. \
Ignore commits related to version bumps, .gitignore changes, or any internal housekeeping \
that is irrelevant to end users. Describe each user-facing change precisely without exposing \
code. Eliminate redundancy. Output strictly valid JSON with no surrounding text, using this \
exact schema: {\"title\": \"<short release title>\", \"body\": \"<detailed markdown release notes>\"}"

# ── Raw mode: just output the commit list ────────────────────────────────

if [ "$RAW_MODE" = true ]; then
  echo "# Release ${VERSION}"
  echo ""
  echo "## Commits"
  echo ""
  echo "$commit_list" | while IFS= read -r line; do
    echo "- ${line}"
  done
  exit 0
fi

# ── Context mode: show what would be sent to the LLM ────────────────────

if [ "$CONTEXT_MODE" = true ]; then
  echo "=== LLM Model ==="
  echo "$LLM_MODEL"
  echo ""
  echo "=== Prompt ==="
  echo "$LLM_PROMPT"
  echo ""
  echo "=== Stdin (commit list) ==="
  echo "$commit_list"
  exit 0
fi

# ── LLM summarization ───────────────────────────────────────────────────

if ! command -v orla >/dev/null 2>&1; then
  die "orla is required for LLM summarization. Use -r for raw output."
fi

if ! command -v jq >/dev/null 2>&1; then
  die "jq is required for JSON parsing. Use -r for raw output."
fi

echo "Summarizing with LLM ($LLM_MODEL)..." >&2

raw_output=$(echo "$commit_list" | \
  ORLA_MAX_TOOL_CALLS=50 orla agent -m "$LLM_MODEL" \
  "$LLM_PROMPT" \
  2>/dev/null) || true

if [ -z "$raw_output" ]; then
  echo "Warning: LLM returned empty output, falling back to raw mode" >&2
  exec "$0" -r -v "$VERSION"
fi

# Sanitize: extract JSON object, strip control characters
sanitized=$(echo "$raw_output" | sed -n '/^{/,/^}/p' | tr -d '\000-\011\013-\014\016-\037')

release_title=$(echo "$sanitized" | jq -r '.title // empty' 2>/dev/null)
release_body=$(echo "$sanitized" | jq -r '.body // empty' 2>/dev/null)

if [ -n "$release_title" ] && [ -n "$release_body" ]; then
  echo "# ${release_title}"
  echo ""
  echo "$release_body"
else
  echo "Warning: JSON parsing failed, falling back to raw mode" >&2
  exec "$0" -r -v "$VERSION"
fi
