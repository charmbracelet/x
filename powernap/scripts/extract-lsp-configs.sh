#!/bin/bash
# Extract LSP configurations from nvim-lspconfig and generate lsps.json
set -euo pipefail

REPO_URL="https://github.com/neovim/nvim-lspconfig.git"
TEMP_DIR=$(mktemp -d)
OUTPUT_FILE="${1:-pkg/config/lsps.json}"

cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

echo "Cloning nvim-lspconfig..." >&2
git clone --depth 1 --quiet "$REPO_URL" "$TEMP_DIR"

LSP_DIR="$TEMP_DIR/lsp"
if [[ ! -d "$LSP_DIR" ]]; then
    echo "Error: lsp/ directory not found" >&2
    exit 1
fi

# Extract a simple string array from Lua content
# Usage: extract_array "content" "field_name"
# Returns JSON array or empty string if complex/invalid
extract_array() {
    local content="$1"
    local field="$2"
    
    # Find field = { and extract balanced braces
    local start_pattern="${field}[[:space:]]*=[[:space:]]*\{"
    if ! echo "$content" | grep -qE "$start_pattern"; then
        return
    fi
    
    # Use awk to extract balanced braces
    local arr
    arr=$(echo "$content" | awk -v field="$field" '
    {
        # Find field = {
        pattern = field "[[:space:]]*=[[:space:]]*\\{"
        if (match($0, pattern)) {
            str = substr($0, RSTART + RLENGTH - 1)  # start from {
            depth = 0
            result = ""
            for (i = 1; i <= length(str); i++) {
                c = substr(str, i, 1)
                if (c == "{") depth++
                if (c == "}") depth--
                result = result c
                if (depth == 0) break
            }
            print result
            exit
        }
    }') || true
    
    [[ -z "$arr" ]] && return
    
    # Remove outer braces
    arr="${arr#\{}"
    arr="${arr%\}}"
    arr=$(echo "$arr" | tr -d '\n' | sed 's/[[:space:]]\+/ /g')
    
    # Skip if contains identifiers outside quotes (function calls, variables)
    local test_arr
    test_arr=$(echo "$arr" | sed "s/\"[^\"]*\"//g" | sed "s/'[^']*'//g")
    if echo "$test_arr" | grep -qE '[a-zA-Z_]'; then
        return
    fi
    
    # Convert to JSON array
    arr=$(echo "$arr" | sed "s/'/\"/g")  # single to double quotes
    
    # Build array
    local result="["
    local first=true
    local IFS=','
    for item in $arr; do
        item=$(echo "$item" | sed 's/^[[:space:]]*//' | sed 's/[[:space:]]*$//')
        if [[ "$item" =~ ^\" ]]; then
            if [[ "$first" == true ]]; then
                first=false
            else
                result="$result, "
            fi
            result="$result$item"
        fi
    done
    result="$result]"
    
    [[ "$result" == "[]" ]] && return
    echo "$result"
}

# Parse a Lua file and extract config as JSON
parse_lua() {
    local file="$1"
    
    # Extract return block content - join lines and normalize whitespace
    # Only strip comments that start at beginning of line (not inside strings like '--stdio')
    local content
    content=$(awk '
    BEGIN { in_return = 0; depth = 0 }
    /^[[:space:]]*--/ { next }
    /return[[:space:]]*\{/ {
        in_return = 1
        idx = index($0, "{")
        if (idx > 0) $0 = substr($0, idx)
    }
    in_return {
        for (i = 1; i <= length($0); i++) {
            c = substr($0, i, 1)
            if (c == "{") depth++
            if (c == "}") depth--
        }
        printf "%s ", $0
        if (depth == 0) exit
    }
    ' "$file" | sed 's/[[:space:]]\+/ /g')
    
    [[ -z "$content" ]] && return
    
    local cmd filetypes root_markers single_file
    cmd=$(extract_array "$content" "cmd")
    filetypes=$(extract_array "$content" "filetypes")
    root_markers=$(extract_array "$content" "root_markers")
    
    # Extract single_file_support
    single_file=$(echo "$content" | grep -oE 'single_file_support[[:space:]]*=[[:space:]]*(true|false)' | sed 's/.*=//' | tr -d ' ') || true
    
    # Skip if no valid command
    [[ -z "$cmd" || "$cmd" == "[]" ]] && return
    
    # Split command array into command (first element) and args (rest)
    # cmd is like: ["gopls"] or ["typescript-language-server", "--stdio"]
    local command_str args_str
    # Remove brackets and split
    local cmd_inner="${cmd#\[}"
    cmd_inner="${cmd_inner%\]}"
    
    # Extract first element as command
    command_str=$(echo "$cmd_inner" | sed 's/,.*//' | tr -d ' ')
    
    # Extract rest as args array
    if [[ "$cmd_inner" == *","* ]]; then
        args_str="[$(echo "$cmd_inner" | sed 's/^[^,]*,//' | sed 's/^[[:space:]]*//' | sed 's/[[:space:]]*$//')]"
    else
        args_str=""
    fi
    
    # Build JSON
    local json="{"
    json="${json}\"command\": $command_str"
    [[ -n "$args_str" && "$args_str" != "[]" ]] && json="${json}, \"args\": $args_str"
    [[ -n "$filetypes" && "$filetypes" != "[]" ]] && json="${json}, \"filetypes\": $filetypes"
    [[ -n "$root_markers" && "$root_markers" != "[]" ]] && json="${json}, \"root_markers\": $root_markers"
    [[ -n "$single_file" ]] && json="${json}, \"single_file_support\": $single_file"
    json="${json}}"
    
    echo "$json"
}

echo "{" > "$OUTPUT_FILE"
first=true

for file in "$LSP_DIR"/*.lua; do
    [[ -f "$file" ]] || continue
    
    name=$(basename "$file" .lua)
    config=$(parse_lua "$file") || true
    
    # Skip if empty
    [[ -z "$config" || "$config" == "{}" ]] && continue
    
    if [[ "$first" == true ]]; then
        first=false
    else
        echo "," >> "$OUTPUT_FILE"
    fi
    
    printf '  "%s": %s' "$name" "$config" >> "$OUTPUT_FILE"
done

echo "" >> "$OUTPUT_FILE"
echo "}" >> "$OUTPUT_FILE"

count=$(grep -c '"command"' "$OUTPUT_FILE" || echo 0)
echo "Generated $OUTPUT_FILE with $count servers" >&2
