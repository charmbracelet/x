#!/bin/bash
# PostToolUse hook: after PR creation/push, remind about review workflow.
# Reads tool input JSON from stdin. Exit 2 sends feedback back to Claude.

input=$(cat)
command=$(echo "$input" | jq -r '.tool_input.command // empty' 2>/dev/null)

# After gh pr create: remind to run review workflow
if [[ "$command" == gh\ pr\ create* ]]; then
    pr_num=$(gh pr view --json number --jq .number 2>/dev/null)
    if [[ -n "$pr_num" ]]; then
        echo "PR created. Run a review pass and a simplification pass now before considering this done." >&2
        # Check for merge conflicts
        sleep 2
        mergeable=$(gh pr view "$pr_num" --json mergeable --jq .mergeable 2>/dev/null)
        if [[ "$mergeable" == "CONFLICTING" ]]; then
            echo "WARNING: PR #$pr_num has merge conflicts. Rebase onto main and resolve before proceeding." >&2
        fi
        exit 2
    fi
fi

# After git push: remind to run review workflow
if [[ "$command" == git\ push* ]]; then
    pr_num=$(gh pr view --json number --jq .number 2>/dev/null)
    if [[ -n "$pr_num" ]]; then
        echo "Pushed to PR #$pr_num. Run a review pass and a simplification pass now." >&2
        sleep 2
        mergeable=$(gh pr view "$pr_num" --json mergeable --jq .mergeable 2>/dev/null)
        if [[ "$mergeable" == "CONFLICTING" ]]; then
            echo "WARNING: PR #$pr_num has merge conflicts. Rebase onto main and resolve before proceeding." >&2
        fi
        exit 2
    fi
fi

exit 0
