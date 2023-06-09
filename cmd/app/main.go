package main

import (
	"github.com/danielmichaels/onpicket/internal/config"
	"github.com/danielmichaels/onpicket/internal/database"
	natsio "github.com/danielmichaels/onpicket/internal/nats"
	"github.com/danielmichaels/onpicket/internal/server"
	"github.com/danielmichaels/onpicket/pkg/api"
	"github.com/go-chi/httplog"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"os"
	"sync"
)

func main() {
	err := run()
	if err != nil {
		log.Error().Err(err).Msg("server failed to start. exiting...")
		os.Exit(1)
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
	dbConn, err := database.InitDatabase(cfg.Db.DbName)
	if err != nil {
		return err
	}
	natsConn, err := natsio.Connect(cfg.Nats.URI)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to NATS. exiting")
	}
	logger.Info().Strs("servers", natsConn.Servers()).Msg("connected to NATS")
	n := natsio.New(
		natsConn, logger, cfg, dbConn,
	)
	s := server.S{
		Config: cfg,
		Logger: logger,
		DB:     dbConn,
		WG:     &sync.WaitGroup{},
		Nats:   n,
	}
	var si api.ServerInterface = &server.Application{}
	app := &server.Application{
		Api: &si,
		S:   s,
	}
	err = s.Nats.InitSubscribers()
	if err != nil {
		app.Logger.Error().Err(err).Msg("NATS failed to start")
		return err
	}
	err = app.Serve()
	if err != nil {
		app.Logger.Error().Err(err).Msg("server failed to start")
		return err
	}
	app.Logger.Info().Msg("draining NATS")
	err = app.Nats.Conn.Drain()
	if err != nil {
		app.Logger.Error().Err(err).Msg("error: failed to drain NATS")
		return err
	}
	app.Logger.Info().Msg("system shutdown")
	return nil
}
