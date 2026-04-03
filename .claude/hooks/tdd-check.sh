#!/bin/bash
# Stop hook: warn if implementation .go files changed without any test files.
# Exits 1 (warn) when tests are missing, 0 otherwise.
#
# On first trigger for a given diff, the warning fires and the diff hash is
# saved. On subsequent stops with the same diff, the hook exits 0 silently.

ACK_FILE=".claude/.tdd-ack"

# Get modified .go files (staged + unstaged)
changed=$(git diff --name-only HEAD 2>/dev/null; git diff --name-only 2>/dev/null)
go_files=$(echo "$changed" | grep '\.go$' | sort -u)

# Separate implementation files from test files
impl_files=$(echo "$go_files" | grep -v '_test\.go$')
test_files=$(echo "$go_files" | grep '_test\.go$')

# TDD is satisfied when: no impl files changed, or at least one test file changed
if [ -z "$impl_files" ] || [ -n "$test_files" ]; then
    rm -f "$ACK_FILE"
    exit 0
fi

# Hash the current impl-only diff (staged + unstaged) to detect changes since last ack.
diff_content=$(echo "$impl_files" | xargs git diff HEAD -- 2>/dev/null; echo "$impl_files" | xargs git diff --cached -- 2>/dev/null)
current_hash=$(echo "$diff_content" | md5 -q 2>/dev/null || echo "$diff_content" | md5sum 2>/dev/null | cut -d' ' -f1)
if [ -f "$ACK_FILE" ] && [ "$(cat "$ACK_FILE")" = "$current_hash" ]; then
    exit 0
fi

# Save the hash so the next stop won't re-trigger for the same diff.
echo "$current_hash" > "$ACK_FILE"

# Implementation changed but no tests — warn (exit 1, not exit 2).
echo "TDD check: implementation files changed but no test files were modified:" >&2
echo "$impl_files" | sed 's/^/  /' >&2
echo "" >&2
echo "Write a failing test first, then implement. If this is a pure refactor" >&2
echo "with no behavior change, explain why no test is needed." >&2
exit 1
