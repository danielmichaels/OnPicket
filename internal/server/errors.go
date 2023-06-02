package server

import (
	"fmt"
	"github.com/danielmichaels/onpicket/internal/response"
	"github.com/danielmichaels/onpicket/pkg/api"
	"net/http"
	"strings"
)

var (
	validationFailed    = "validation failed"
	notFound            = "the requested resource could not be found"
	internalServerError = "the server encountered a problem and could not process your request"
	methodNotAllow      = "the %s method is not supported for this resource"
	rateLimitExceeded   = "rate limit exceeded"
	//badRequest          = "bad request"
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
	app.errorMessage(w, r, http.StatusInternalServerError, internalServerError, nil)
}

func (app *Application) notFound(w http.ResponseWriter, r *http.Request) {
	app.errorMessage(w, r, http.StatusNotFound, notFound, nil)
}

func (app *Application) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf(methodNotAllow, r.Method)
	app.errorMessage(w, r, http.StatusMethodNotAllowed, message, nil)
}
func (app *Application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	app.errorMessage(w, r, http.StatusTooManyRequests, rateLimitExceeded, nil)
}

//func (app *Application) badRequest(w http.ResponseWriter, r *http.Request, err error) {
//	app.errorMessage(w, r, http.StatusBadRequest, badRequest, nil)
//}

func (app *Application) apiValidationError(w http.ResponseWriter, errorInfo map[string]interface{}) {
	app.Error(w, http.StatusUnprocessableEntity, validationFailed, Envelope{"fields": errorInfo})
}

//func (app *Application) invalidAuthenticationToken(w http.ResponseWriter, r *http.Request) {
//	headers := make(http.Header)
//	headers.Set("WWW-Authenticate", "Bearer")
//
//	app.errorMessage(w, r, http.StatusUnauthorized, "Invalid authentication token", headers)
//}

//func (app *Application) authenticationRequired(w http.ResponseWriter, r *http.Request) {
//	app.errorMessage(w, r, http.StatusUnauthorized, "You must be authenticated to access this resource", nil)
//}
//
//func (app *Application) basicAuthenticationRequired(w http.ResponseWriter, r *http.Request) {
//	headers := make(http.Header)
//	headers.Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
//
//	message := "You must be authenticated to access this resource"
//	app.errorMessage(w, r, http.StatusUnauthorized, message, headers)
//}
