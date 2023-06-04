package server

import (
	"github.com/danielmichaels/onpicket/assets"
	"github.com/danielmichaels/onpicket/pkg/api"
	oapimiddleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog"
	"github.com/go-chi/httprate"
	"net/http"
)

func (app *Application) routes(openapi *openapi3.T) http.Handler {
	router := chi.NewRouter()
	router.Use(app.recoverPanic)
	router.Use(app.securityHeaders)
	router.Use(middleware.RealIP)
	router.Use(middleware.GetHead)
	router.Use(middleware.Compress(5))
	router.Use(httplog.RequestLogger(app.Logger))
	if app.Config.Limiter.Enabled {
		router.Use(httprate.Limit(
			app.Config.Limiter.Rps,
			app.Config.Limiter.BackOff,
			httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
			httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
				app.rateLimitExceededResponse(w, r)
			})))
	}
	router.NotFound(app.notFound)
	router.MethodNotAllowed(app.methodNotAllowed)

	fileServer := http.FileServer(http.FS(assets.EmbeddedFiles))
	router.Handle("/static/*", fileServer)

	router.Get("/status", app.status)
	router.Group(func(web chi.Router) {
		web.Get("/", app.home)
		web.Get("/docs", app.docs)
	})

	router.Group(func(oapi chi.Router) {
		oapi.Use(oapimiddleware.OapiRequestValidator(openapi))
		api.HandlerFromMuxWithBaseURL(app, router, "/api")
	})
	return router
}
