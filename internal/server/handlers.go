package server

import (
	"github.com/danielmichaels/onpicket/internal/request"
	"github.com/danielmichaels/onpicket/internal/response"
	"github.com/danielmichaels/onpicket/internal/validator"
	"github.com/danielmichaels/onpicket/internal/version"
	"github.com/danielmichaels/onpicket/pkg/api"
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
func (app *Application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	err := response.Page(w, http.StatusOK, data, "pages/home.tmpl")
	if err != nil {
		app.serverError(w, r, err)
	}
}

type ApiStore struct{}
type Envelope map[string]any

func NewApiStore() *ApiStore {
	return &ApiStore{}
}

var si api.ServerInterface = NewApiStore()
//func (a *ApiStore) Error(w http.ResponseWriter, code int, message string, errorInfo map[string]interface{}) {
func (app *Application) Error(w http.ResponseWriter, code int, message string, errorInfo map[string]interface{}) {

	apiErr := api.Error{
		Code:   int32(code),
		Status: message,
	}
	if len(errorInfo) != 0 {
		apiErr.Body = &errorInfo
	}
	w.WriteHeader(code)
	_ = response.JSON(w, code, apiErr)
}

func (app *Application) apiValidationError(w http.ResponseWriter, errors string, errorInfo map[string]interface{}) {
	ei := Envelope{"detail": errorInfo}
	app.Error(w, http.StatusUnprocessableEntity, errors, ei)
}

func (app *Application) Healthz(w http.ResponseWriter, r *http.Request) {
	health := api.Healthz{
		Status:  "OK",
		Version: version.Get(),
	}
	app.Error(w, 404, "an error", nil)
	return
	_ = response.JSON(w, http.StatusOK, health)
}

var users = []api.User{
	{Timezone: "AU", Username: "Admin User"},
	{Timezone: "EU", Username: "Standard User"},
}

func (a *ApiStore) ListUsers(w http.ResponseWriter, r *http.Request) {
	_ = response.JSON(w, http.StatusOK, users)
}
func (app *Application) CreateUser(w http.ResponseWriter, r *http.Request) {
	var newUser api.UserBody
	err := request.DecodeJSON(w, r, &newUser)
	if err != nil {
		app.Error(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	v := validator.Validator{}
	v.CheckField(newUser.Password2 != "", "password2", "password2 not found")
	v.Check(newUser.Password2 != "", "password2 not found")
	if v.HasErrors() {
		a.apiValidationError(w, "validation failed", v.FieldErrors)
		//app.
		//return
	//}
	user := api.User{
		Username: newUser.Username,
		Timezone: newUser.Timezone,
	}

	users = append(users, user)

	_ = response.JSON(w, http.StatusCreated, user)
}
