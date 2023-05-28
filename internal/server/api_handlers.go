package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/danielmichaels/onpicket/internal/database"
	natsio "github.com/danielmichaels/onpicket/internal/nats"
	"github.com/danielmichaels/onpicket/internal/request"
	"github.com/danielmichaels/onpicket/internal/response"
	"github.com/danielmichaels/onpicket/internal/services"
	"github.com/danielmichaels/onpicket/internal/validator"
	"github.com/danielmichaels/onpicket/internal/version"
	"github.com/danielmichaels/onpicket/pkg/api"
	"go.mongodb.org/mongo-driver/bson"
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

func (app *Application) ListScans(w http.ResponseWriter, r *http.Request) {
	filter := bson.D{}
	cursor, err := app.DB.Collection(database.ScanCollection).Find(context.TODO(), filter)
	if err != nil {
		app.notFound(w, r)
		return
	}
	var data []services.NmapScanOut
	if err = cursor.All(context.TODO(), &data); err != nil {
		app.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	_ = response.JSON(w, http.StatusOK, data)
}
func (app *Application) CreateScan(w http.ResponseWriter, r *http.Request) {
	var ns api.ScanBody
	err := request.DecodeJSON(w, r, &ns)
	if err != nil {
		app.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// make this a function
	v := validator.Validator{}
	v.CheckField(ns.Host != "", "host", "host must not be empty")
	v.CheckField(ns.Ports != nil, "ports", "ports must not be empty")
	// todo: validate against:
	// -p-
	// 1-200
	// 22,33,44
	v.CheckField(validator.NotBlank(string(ns.Type)), "type", "type must not be empty")
	enumTypes := []api.NewScanType{api.PortScan, api.ServiceDiscovery, api.ServiceDiscoveryDefaultScripts}
	v.CheckField(validator.In(ns.Type, enumTypes...), "type", "type must be a valid option")

	if v.HasErrors() {
		app.apiValidationError(w, "validation failed", v.FieldErrors)
		return
	}
	scan := api.Scan{
		Id:     generateName(string(ns.Type)),
		Ports:  ns.Ports,
		Host:   ns.Host,
		Type:   string(ns.Type),
		Status: api.Scheduled,
	}

	// poc
	// publish message to NATS
	data, err := json.Marshal(scan)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	err = app.Nats.Conn.Publish(natsio.ScanStartSubj, data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	_ = response.JSON(w, http.StatusCreated, Envelope{"scan": scan.Id})
}
