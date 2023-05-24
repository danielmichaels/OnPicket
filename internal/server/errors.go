package server

import (
	"fmt"
	"github.com/danielmichaels/onpicket/internal/response"
	"github.com/danielmichaels/onpicket/pkg/api"
	"net/http"
	"strings"
)

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

func (app *Application) errorMessage(w http.ResponseWriter, r *http.Request, status int, message string, headers http.Header) {
	message = strings.ToUpper(message[:1]) + message[1:]

	err := response.JSONWithHeaders(w, status, map[string]string{"error": message}, headers)
	if err != nil {
		app.Logger.Error().Err(err).Msg(message)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (app *Application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	app.Logger.Error().Err(err).Msg("server error")

	message := "The server encountered a problem and could not process your request"
	app.errorMessage(w, r, http.StatusInternalServerError, message, nil)
}

func (app *Application) notFound(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found"
	app.errorMessage(w, r, http.StatusNotFound, message, nil)
}

func (app *Application) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	app.errorMessage(w, r, http.StatusMethodNotAllowed, message, nil)
}

func (app *Application) badRequest(w http.ResponseWriter, r *http.Request, err error) {
	app.errorMessage(w, r, http.StatusBadRequest, err.Error(), nil)
}

func (app *Application) apiValidationError(w http.ResponseWriter, errors string, errorInfo map[string]interface{}) {
	app.Error(w, http.StatusUnprocessableEntity, errors, Envelope{"fields": errorInfo})
}

func (app *Application) invalidAuthenticationToken(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)
	headers.Set("WWW-Authenticate", "Bearer")

	app.errorMessage(w, r, http.StatusUnauthorized, "Invalid authentication token", headers)
}

func (app *Application) authenticationRequired(w http.ResponseWriter, r *http.Request) {
	app.errorMessage(w, r, http.StatusUnauthorized, "You must be authenticated to access this resource", nil)
}

func (app *Application) basicAuthenticationRequired(w http.ResponseWriter, r *http.Request) {
	headers := make(http.Header)
	headers.Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)

	message := "You must be authenticated to access this resource"
	app.errorMessage(w, r, http.StatusUnauthorized, message, headers)
}
