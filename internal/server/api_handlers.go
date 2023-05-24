package server

import (
	"github.com/danielmichaels/onpicket/internal/request"
	"github.com/danielmichaels/onpicket/internal/response"
	"github.com/danielmichaels/onpicket/internal/validator"
	"github.com/danielmichaels/onpicket/internal/version"
	"github.com/danielmichaels/onpicket/pkg/api"
	"net/http"
)

func (app *Application) Healthz(w http.ResponseWriter, r *http.Request) {
	health := api.Healthz{
		Status:  "OK",
		Version: version.Get(),
	}
	_ = response.JSON(w, http.StatusOK, health)
}

var scans = []api.Scan{
	{Host: "192.168.1.1"},
	{Host: "192.168.1.1"},
}

func (app *Application) ListScans(w http.ResponseWriter, r *http.Request) {
	_ = response.JSON(w, http.StatusOK, scans)
}
func (app *Application) CreateScan(w http.ResponseWriter, r *http.Request) {
	var newScan api.ScanBody
	err := request.DecodeJSON(w, r, &newScan)
	if err != nil {
		app.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	v := validator.Validator{}
	v.CheckField(newScan.Host != "", "host", "host must not be empty")
	v.CheckField(newScan.Ports != nil, "ports", "ports must not be empty")

	if v.HasErrors() {
		app.apiValidationError(w, "validation failed", v.FieldErrors)
		return
	}
	scan := api.Scan{
		Host:  newScan.Host,
		Ports: newScan.Ports,
	}

	scans = append(scans, scan)

	_ = response.JSON(w, http.StatusCreated, scan)
}
