package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/danielmichaels/onpicket/internal/config"
	"github.com/danielmichaels/onpicket/internal/database"
	natsio "github.com/danielmichaels/onpicket/internal/nats"
	"github.com/danielmichaels/onpicket/internal/version"
	"github.com/danielmichaels/onpicket/pkg/api"
	"github.com/rs/zerolog"
)

type Envelope map[string]any

// Application is the applications main struct. It holds references to
// all the methods and handlers of this application.
type Application struct {
	Api *api.ServerInterface
	S
}
type S struct {
	Config *config.Conf
	Logger zerolog.Logger
	WG     *sync.WaitGroup
	Models *database.Queries
	Nats   *natsio.Nats
}

func (app *Application) Serve() error {
	openapi, err := api.GetSwagger()
	if err != nil {
		app.Logger.Fatal().Err(err).Msg("error loading swagger")
		os.Exit(1)
	}
	openapi.Servers = nil
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.Config.Server.Port),
		Handler:      app.routes(openapi),
		IdleTimeout:  app.Config.Server.TimeoutIdle,
		ReadTimeout:  app.Config.Server.TimeoutRead,
		WriteTimeout: app.Config.Server.TimeoutWrite,
	}

	shutdownError := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.Logger.Warn().Str("signal", s.String()).Msg("caught signal")

		// Allow processes to finish with a ten-second window
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}
		app.Logger.Warn().Str("tasks", srv.Addr).Msg("completing background tasks")
		// Call wait so that the wait group can decrement to zero.
		app.WG.Wait()
		shutdownError <- nil
	}()
	app.Logger.Info().Str("server", srv.Addr).Str("version", version.Get()).Msg("starting server")

	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	err = <-shutdownError
	if err != nil {
		app.Logger.Warn().Str("server", srv.Addr).Msg("stopped server")
		return err
	}
	return nil
}
