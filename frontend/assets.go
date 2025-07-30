package frontend

import (
	"embed"
	"io/fs"
)

//go:embed *
var assets embed.FS

// Assets returns the embedded filesystem containing the frontend assets
func Assets() fs.FS {
	return assets
}
