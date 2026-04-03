#!/bin/bash
# PreToolUse hook: block send-keys commands that tell workers to autonomously
# pick up work from the backlog.
#
# Exit 2 = block the tool call and send feedback to Claude.

input=$(cat)
command=$(echo "$input" | jq -r '.tool_input.command // empty' 2>/dev/null)

# Only check amux send-keys and amux type-keys commands
if ! echo "$command" | grep -qE 'amux (send-keys|type-keys)'; then
    exit 0
fi

# Block commands that tell workers to pick up work autonomously
if echo "$command" | grep -qiE 'pick up.*(work|issue|task|ticket)|from.*(backlog|queue|linear)|new work|next (issue|task|ticket)|find.*(issue|task|work).*backlog'; then
    echo "BLOCKED: Do not tell workers to autonomously pick up work from the backlog. Only assign specific, user-approved issues. Ask the user which issue to assign." >&2
    exit 2
fi

exit 0
