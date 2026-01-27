#!/bin/bash
# Extract LSP configurations from nvim-lspconfig and generate lsps.json
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
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

echo "Extracting LSP configurations..." >&2
lua "$SCRIPT_DIR/extract-lsp-configs.lua" "$LSP_DIR" > "$OUTPUT_FILE"

count=$(grep -c '"command"' "$OUTPUT_FILE" || echo 0)
echo "Generated $OUTPUT_FILE with $count servers" >&2
