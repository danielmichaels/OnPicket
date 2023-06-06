package server

import (
	"encoding/json"
	"github.com/danielmichaels/onpicket/internal/response"
	"github.com/danielmichaels/onpicket/internal/version"
	"github.com/danielmichaels/onpicket/pkg/api"
	"github.com/getkin/kin-openapi/openapi3"
	"net/http"
)

func (app *Application) status(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"Status":  "OK",
		"Version": version.Get(),
	}

	err := response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}

//	func (app *Application) home(w http.ResponseWriter, r *http.Request) {
//		data := app.newTemplateData(r)
//
//		err := response.Page(w, http.StatusOK, data, "pages/home.tmpl")
//		if err != nil {
//			app.Logger.Error().Err(err).Str("template", "pages/home.tmpl").Msg("template render err")
//			app.serverError(w, r, err)
//		}
//	}
func (app *Application) docs(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	openapi, err := api.GetSwagger()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	loader := openapi3.NewLoader()
	spec, err := json.Marshal(openapi)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	doc, err := loader.LoadFromData(spec)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	raw, err := doc.MarshalJSON()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	data["SpecString"] = string(raw)
	err = response.ApiDocsPage(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
		app.Logger.Error().Err(err).Send()
	}
}
