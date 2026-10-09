#!/usr/bin/env bash
# Fails if any tracked or new text file has a character outside printable ASCII, tab and newline.
set -euo pipefail
cd "$(dirname "$0")/.."
found=$(git ls-files -z --cached --others --exclude-standard | xargs -0 env LC_ALL=C grep -lI '[^ -~	]' -- || true)
if [ -n "$found" ]; then
  echo "non keyboard characters in:"
  echo "$found"
  exit 1
fi
