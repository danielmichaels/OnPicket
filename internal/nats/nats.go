package natsio

import (
	"os"
	"time"

	"github.com/danielmichaels/onpicket/internal/config"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Nats struct {
	Conn *nats.Conn

	L zerolog.Logger
	C *config.Conf
}

func New(conn *nats.Conn, l zerolog.Logger, cfg *config.Conf) *Nats {
	return &Nats{
		Conn: conn,
		L:    l,
		C:    cfg,
	}
}

func Connect(name string) (*nats.Conn, error) {
	uri := config.AppConfig().Nats.URI
	opts := []nats.Option{nats.Name(name)}
	opts = setupConnOptions(opts)
	nc, err := nats.Connect(
		uri,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return nc, nil
}

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second
	pingInterval := 20 * time.Second
	maxPingOutstanding := 5
	timeout := 30 * time.Second

	opts = append(opts, nats.Timeout(timeout))
	opts = append(opts, nats.PingInterval(pingInterval))
	opts = append(opts, nats.MaxPingsOutstanding(maxPingOutstanding))
	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		log.Printf("Disconnected due to: %s, will attempt reconnects for %.0fm", err, totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Error().Msgf("Exiting: %v", nc.LastError())
		os.Exit(1)
	}))
	opts = append(opts, nats.DrainTimeout(10*time.Second))
	return opts
}
