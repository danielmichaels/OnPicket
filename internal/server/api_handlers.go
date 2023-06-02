package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/danielmichaels/onpicket/internal/database"
	natsio "github.com/danielmichaels/onpicket/internal/nats"
	"github.com/danielmichaels/onpicket/internal/request"
	"github.com/danielmichaels/onpicket/internal/response"
	"github.com/danielmichaels/onpicket/internal/services"
	"github.com/danielmichaels/onpicket/internal/validator"
	"github.com/danielmichaels/onpicket/internal/version"
	"github.com/danielmichaels/onpicket/pkg/api"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// generateName creates a random name for use in identifiers
func generateName(s string) string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s", s, hex.EncodeToString(b))
}

func (app *Application) Healthz(w http.ResponseWriter, _ *http.Request) {
	health := api.Healthz{
		Status:  "OK",
		Version: version.Get(),
	}
	_ = response.JSON(w, http.StatusOK, health)
}
func sortDirection(s string) bson.D {
	switch strings.ToLower(s) {
	case "asc":
		return bson.D{{Key: "_id", Value: -1}}
	case "desc":
		return bson.D{{Key: "_id", Value: 1}}
	default:
		return bson.D{{Key: "_id", Value: -1}}
	}
}
func (app *Application) ListScans(w http.ResponseWriter, r *http.Request, params api.ListScansParams) {
	var v validator.Validator

	pageNo := request.ReadInt(params.Page, "page", 1, &v)
	pageSize := request.ReadInt(params.PageSize, "page_size", 20, &v)
	sort := request.ReadString(params.Sort, "asc")
	if v.HasErrors() {
		app.apiValidationError(w, v.FieldErrors)
		return
	}

	opts := options.Find().SetLimit(pageSize).SetSkip((pageNo - 1) * pageSize).SetSort(sortDirection(sort))
	filter := bson.D{}
	cursor, err := app.DB.Collection(database.ScanCollection).Find(context.TODO(), filter, opts)
	if err != nil {
		app.notFound(w, r)
		return
	}

	total, err := app.DB.Collection(database.ScanCollection).CountDocuments(context.TODO(), bson.D{}, options.Count().SetHint("_id_"))
	if err != nil {
		app.notFound(w, r)
		return
	}

	var data []services.NmapScanOut
	if err = cursor.All(context.TODO(), &data); err != nil {
		app.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	_ = response.JSON(w, http.StatusOK, Envelope{
		"data":     data,
		"metadata": services.CalculateMetadata(total, pageNo, pageSize)})
}

func (app *Application) RetrieveScan(w http.ResponseWriter, r *http.Request, id string) {
	var v validator.Validator
	paramId := chi.URLParam(r, "id")

	v.CheckField(validator.NotBlank(id), "id", "scan_id must be provided")
	if v.HasErrors() {
		app.apiValidationError(w, v.FieldErrors)
		return
	}

	filter := bson.D{{Key: "id", Value: paramId}}
	cursor, err := app.DB.Collection(database.ScanCollection).Find(context.TODO(), filter, nil)
	if err != nil {
		app.notFound(w, r)
		return
	}

	var data []services.NmapScanOut
	if err = cursor.All(context.TODO(), &data); err != nil {
		app.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	_ = response.JSON(w, http.StatusOK, Envelope{
		"id":   paramId,
		"data": data,
	})
}

func (app *Application) CreateScan(w http.ResponseWriter, r *http.Request) {
	var v validator.Validator
	var ns api.ScanBody
	err := request.DecodeJSON(w, r, &ns)
	if err != nil {
		v.AddFieldError("json", err.Error())
		app.apiValidationError(w, v.FieldErrors)
		return
	}

	// make this a function
	v.CheckField(ns.Hosts != nil, "hosts", "host must not be empty")
	v.CheckField(ns.Ports != nil, "ports", "ports must not be empty")
	// todo: validate against:
	// -p-
	// 1-200
	// 22,33,44
	v.CheckField(validator.NotBlank(string(ns.Type)), "type", "type must not be empty")
	enumTypes := []api.NewScanType{api.PortScan, api.ServiceDiscovery, api.ServiceDiscoveryDefaultScripts}
	v.CheckField(validator.In(ns.Type, enumTypes...), "type", "type must be a valid option")

	if v.HasErrors() {
		app.apiValidationError(w, v.FieldErrors)
		return
	}
	scan := api.Scan{
		Id:          generateName(string(ns.Type)),
		Ports:       ns.Ports,
		Hosts:       ns.Hosts,
		Type:        string(ns.Type),
		Status:      api.Scheduled,
		Description: ns.Description,
	}
	if ns.Timeout != nil {
		scan.Timeout = ns.Timeout
	}

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
	_ = response.JSON(w, http.StatusCreated, scan)
}
