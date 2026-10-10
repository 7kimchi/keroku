#!/usr/bin/env bash
# Commits files one at a time. Reads "path|message" lines from stdin.
set -euo pipefail
cd "$(dirname "$0")/.."
while IFS='|' read -r path message; do
  [ -n "$path" ] || continue
  if [ -f "$path" ] && [ "$path" != go.sum ] && [ "$(awk 'END { print NR }' "$path")" -gt 100 ]; then
    echo "refusing $path: over 100 lines" >&2
    exit 1
  fi
  git add -- "$path"
  if git diff --cached --quiet -- "$path"; then
    echo "skip $path: no change"
    continue
  fi
  git commit -qm "$message" -- "$path"
done
