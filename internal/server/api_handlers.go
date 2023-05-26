package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	natsio "github.com/danielmichaels/onpicket/internal/nats"
	"github.com/danielmichaels/onpicket/internal/request"
	"github.com/danielmichaels/onpicket/internal/response"
	"github.com/danielmichaels/onpicket/internal/validator"
	"github.com/danielmichaels/onpicket/internal/version"
	"github.com/danielmichaels/onpicket/pkg/api"
	"net/http"
)

// generateName creates a random name for use in identifiers
func generateName(s string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s", s, hex.EncodeToString(b))
}

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

	// make this a function
	v := validator.Validator{}
	v.CheckField(newScan.Host != "", "host", "host must not be empty")
	v.CheckField(newScan.Ports != nil, "ports", "ports must not be empty")
	// todo: validate against:
	// -p-
	// 1-200
	// 22,33,44
	v.CheckField(validator.NotBlank(string(newScan.Type)), "type", "type must not be empty")
	enumTypes := []api.NewScanType{api.PortScan, api.ServiceDiscovery}
	v.CheckField(validator.In(newScan.Type, enumTypes...), "type", "type must be a valid option")

	if v.HasErrors() {
		app.apiValidationError(w, "validation failed", v.FieldErrors)
		return
	}
	scan := api.Scan{
		Id:     generateName(string(newScan.Type)),
		Ports:  newScan.Ports,
		Host:   newScan.Host,
		Type:   string(newScan.Type),
		Status: api.Scheduled,
	}

	scans = append(scans, scan)

	// poc
	// publish message to NATS
	data, _ := json.Marshal(scan)
	err = app.Nats.Conn.Publish(natsio.ScanStartSubj, data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	_ = response.JSON(w, http.StatusCreated, nil)
	// _ = response.JSON(w, http.StatusCreated, scan)
}
