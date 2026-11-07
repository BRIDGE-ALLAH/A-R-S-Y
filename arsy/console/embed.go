// Package console embeds the separately built A-R-S-Y dashboard.
package console

import (
	"embed"
	"io/fs"
)

//go:embed dist
var embedded embed.FS

// FS returns the dashboard distribution as a filesystem.
func FS() fs.FS {
	files, err := fs.Sub(embedded, "dist")
	if err != nil {
		panic(err) // dist is a compile-time embedded directory.
	}
	return files
}
