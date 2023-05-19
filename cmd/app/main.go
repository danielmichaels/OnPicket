package main

import (
	"github.com/danielmichaels/onpicket/internal/config"
	"github.com/danielmichaels/onpicket/internal/database"
	"github.com/danielmichaels/onpicket/internal/server"
	"github.com/danielmichaels/onpicket/pkg/api"
	"github.com/go-chi/httplog"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

func main() {
	err := run()
	if err != nil {
		log.Fatalln("server failed to start:", err)
	}
}

func run() error {
	cfg := config.AppConfig()
	logger := httplog.NewLogger("onpicket", httplog.Options{
		JSON:     cfg.AppConf.LogJson,
		Concise:  cfg.AppConf.LogConcise,
		LogLevel: cfg.AppConf.LogLevel,
	})
	if cfg.AppConf.LogCaller {
		logger = logger.With().Caller().Logger()
	}
	db, err := database.OpenDB(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to open database. exiting")
	}
	var si api.ServerInterface = server.ApiStore{}
	app := &server.Application{
		Config: cfg,
		Logger: logger,
		Models: database.New(db),
		Api:    si,
		//Api: server.NewApiStore(),
	}
	err = app.Serve()
	if err != nil {
		app.Logger.Error().Err(err).Msg("server failed to start")
	}
	return nil
}
