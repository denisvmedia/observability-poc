//go:build with_frontend

package frontend

import (
	"embed"
	"io/fs"
)

//go:embed dist
var dist embed.FS

// GetDist returns the embedded frontend dist filesystem.
func GetDist() fs.ReadFileFS {
	return dist
}

