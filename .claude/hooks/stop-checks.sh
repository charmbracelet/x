#!/bin/bash

status=0

run_check() {
    local script="$1"
    local output

    if ! output=$("$script" 2>&1); then
        if [[ -n "$output" ]]; then
            printf '%s\n' "$output" >&2
        fi
        status=1
        return
    fi

    if [[ -n "$output" ]]; then
        printf '%s\n' "$output" >&2
    fi
}

run_check ".claude/hooks/tdd-check.sh"

exit "$status"
