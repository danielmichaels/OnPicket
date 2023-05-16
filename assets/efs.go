package assets

import (
	"embed"
)

// go:embed "templates" "static" "migrations"
//
//go:embed "templates" "static"
var EmbeddedFiles embed.FS
