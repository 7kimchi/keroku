#!/usr/bin/env bash
# Installs the pinned lint and security tools with the current Go toolchain.
set -euo pipefail
binDir="$(go env GOPATH)/bin"

# staticcheck 0.8.1 ships an x/tools too old for Go 1.27 export data, so build it against a newer one.
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
(
  cd "$work"
  go mod init sctools >/dev/null 2>&1
  go get honnef.co/go/tools/cmd/staticcheck@v0.8.1 golang.org/x/tools@v0.51.0 >/dev/null 2>&1
  go build -o "$binDir/staticcheck" honnef.co/go/tools/cmd/staticcheck
)

go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0
go install golang.org/x/vuln/cmd/govulncheck@v1.8.0
go install github.com/securego/gosec/v2/cmd/gosec@v2.29.0
