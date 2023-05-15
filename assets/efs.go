package assets

import (
	"embed"
)

//go:embed "templates" "static" "migrations"
var EmbeddedFiles embed.FS
