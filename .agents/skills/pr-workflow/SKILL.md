---
name: pr-workflow
description: Use when creating, updating, reviewing, or merging a PR in this repo. Covers rebases, review passes, and post-merge verification.
---

# PR Workflow

Use this skill when the task involves `git push`, `gh pr create`, `gh pr merge`, PR review, or wrapping up a change.

## Rules

- Rebase onto `origin/main` before the first push: `git fetch origin main && git rebase origin/main`.
- This repo uses squash merges. Use `gh pr merge --squash`.
- Prefer `gh pr create --body-file ...` for multiline PR descriptions.
- If a change is ready for review, open the PR proactively.

## Workflow

1. Confirm the relevant tests ran and note any gaps.
2. Rebase onto `origin/main` before the first push.
3. Create or update the PR.
4. After push, check CI status: `gh pr checks --watch`.
5. If CI fails, fix and push again (up to 3 attempts).
6. Run a review pass on the diff.
7. Run a simplification pass.
8. Before merging, re-check mergeability: `gh pr view --json mergeable`.
9. After merge, verify: `git checkout main && git pull --ff-only`.

## PR Description Template

```
## Motivation
Why this change?

## Summary
- Key change 1
- Key change 2

## Testing
cd vt && go test -run TestRelevant -v
cd vt && go test ./...

## Review focus
What should reviewers look at?

Closes LAB-NNN
```
