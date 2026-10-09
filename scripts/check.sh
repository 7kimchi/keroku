#!/usr/bin/env bash
# Runs every gate in order. Stops at the first failure.
set -euo pipefail
cd "$(dirname "$0")/.."
export PATH="$PATH:$(go env GOPATH)/bin"

step() { echo "== $1"; }

step "line count"
scripts/checkLines.sh

step "gofmt"
unformatted=$(gofmt -l $(git ls-files --cached --others --exclude-standard '*.go'))
if [ -n "$unformatted" ]; then
  echo "$unformatted"
  exit 1
fi

step "go vet"
go vet ./...
go vet -tags load ./...

step "staticcheck"
staticcheck ./...

step "golangci-lint"
golangci-lint run ./...

step "govulncheck"
govulncheck ./...

step "gosec"
gosec -quiet ./...

step "go test -race"
if [ -z "${KEROKU_TEST_DATABASE_URL:-}" ]; then
  KEROKU_TEST_DATABASE_URL=$(scripts/testDb.sh start)
  export KEROKU_TEST_DATABASE_URL
fi
go test -race -count=1 ./...

echo "== ok"
