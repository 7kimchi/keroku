#!/usr/bin/env bash
# Fails if any tracked or new file is over 100 lines. go.sum, CLAUDE.md and generated files are exempt.
set -euo pipefail
cd "$(dirname "$0")/.."
maxLines=100
failed=0
while IFS= read -r -d '' file; do
  case "$file" in
    go.sum|CLAUDE.md) continue ;;
  esac
  [ -f "$file" ] || continue
  if head -n 5 "$file" | grep -q '^// Code generated .* DO NOT EDIT\.$'; then
    continue
  fi
  lines=$(awk 'END { print NR }' "$file")
  if [ "$lines" -gt "$maxLines" ]; then
    echo "$file: $lines lines (max $maxLines)"
    failed=1
  fi
done < <(git ls-files -z --cached --others --exclude-standard)
exit "$failed"
