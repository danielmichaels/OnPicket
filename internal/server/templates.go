package server

import (
	"github.com/danielmichaels/onpicket/internal/version"
	"net/http"
)

func (app *Application) newTemplateData(r *http.Request) map[string]any {
	data := map[string]any{
		"Version": version.Get(),
	}

	return data
}
