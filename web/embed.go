// Package web embeds the built frontend (web/dist) into the dtvlt binary.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Assets returns the built frontend with index.html at its root. Before
// `npm run build` it only holds a placeholder, and the server says so.
func Assets() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}
