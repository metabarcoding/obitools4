#!/bin/bash

basePath="/Users/coissac/Sync/travail/__MOI__/GO/obitools4"
OUTPUT_FILE="${1:-/dev/stdout}"

# Get all SetAttribute calls
rg -n '\.SetAttribute\(' "$basePath/pkg" --type go | while read -r line; do
	file="${line%%:*}"
	line_num="${line%:*}"
	line_num="${line_num##*:}"
	context="${line##*: }"

	# Extract key (only literal strings)
	key=$(echo "$context" | sed -n 's/.*SetAttribute("\([^"]*\)".*/\1/p')
	[ -z "$key" ] && continue

	# Get function name using treesitter
	func=$(kilo treesitter_cursor_walk \
		--file_path "$file" \
		--row "$((line_num - 1))" \
		--column 0 \
		--max_depth 10 2>/dev/null |
		jq -r '.ancestors[] | select(.type == "function_declaration" or .type == "method_declaration") | .children[] | select(.type == "identifier" or .type == "field_identifier") | .text' 2>/dev/null)

	# Fallback to func_literal for closures
	if [ -z "$func" ]; then
		func=$(kilo treesitter_cursor_walk \
			--file_path "$file" \
			--row "$((line_num - 1))" \
			--column 0 \
			--max_depth 10 2>/dev/null |
			jq -r '.ancestors[] | select(.type == "func_literal") | "closure"' 2>/dev/null)
	fi

	echo "$(basename "$file")|$line_num|$key|${func:-unknown}|$context"
done | sort -t'|' -k1,1 -k2,2n
