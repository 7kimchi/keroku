// Package migrations embeds the versioned SQL schema into the binary.
package migrations

import (
	"embed"
	"io/fs"
)

//go:embed *.sql
var files embed.FS

// Files returns the embedded migration files.
func Files() fs.FS {
	return files
}
