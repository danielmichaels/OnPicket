package natsio

import (
	"encoding/json"
	"github.com/danielmichaels/onpicket/internal/funcs"
	"github.com/danielmichaels/onpicket/internal/services"
	"github.com/danielmichaels/onpicket/pkg/api"
	"github.com/nats-io/nats.go"
)

const (
	ScanQueue        = "scans"
	ScanStartSubj    = "scan.start"
	ScanCompleteSubj = "scan.complete"
	ScanRetrySubj    = "scan.retry"
	ScanFailSubj     = "scan.fail"
)

func (n *Nats) InitSubscribers() error {
	if err := n.scanStartQueueGroup(); err != nil {
		return err
	}
	if err := n.scanRetryQueueGroup(); err != nil {
		return err
	}
	if err := n.scanCompleteQueueGroup(); err != nil {
		return err
	}
	if err := n.scanFailQueueGroup(); err != nil {
		return err
	}
	return nil
}

func (n *Nats) scanStartQueueGroup() error {
	if _, err := n.Conn.QueueSubscribe(ScanStartSubj, ScanQueue, func(msg *nats.Msg) {
		n.L.Debug().Msgf("msg.Data received: %s", string(msg.Data))
		var scan api.Scan
		err := json.Unmarshal(msg.Data, &scan)
		if err != nil {
			n.L.Error().Err(err).Msgf("err unmarshalling published message")
			return
		}
		funcs.BackgroundFunc(func() {
			sRes, err := services.StartScan(&scan)
			if err != nil {
				n.L.Error().Err(err).Msg("scan error")
				err = n.scanError()
				if err != nil {
					n.L.Error().Err(err).Msg("scan publish ScanFail error")
					return
				}
				return
			}

			r, err := json.Marshal(sRes)
			if err != nil {
				n.L.Error().Err(err).Msg("scan unmarshal error")
				return
			}
			err = n.Conn.Publish(ScanCompleteSubj, r)
			if err != nil {
				n.L.Error().Err(err).Msg("scan publish ScanComplete error")
				return
			}
		})
	}); err != nil {
		n.L.Error().Err(err).Msgf("err subscribing")
		return err
	}
	return nil
}

func (n *Nats) scanCompleteQueueGroup() error {
	if _, err := n.Conn.QueueSubscribe(ScanCompleteSubj, ScanQueue, func(msg *nats.Msg) {
		n.L.Debug().Msgf("msg.Data received: %s", string(msg.Data))
		// update db with scan sucess
		// (optional) alert customer scan completed
	}); err != nil {
		n.L.Error().Err(err).Msgf("err subscribing")
		return err
	}
	return nil
}

func (n *Nats) scanRetryQueueGroup() error {
	if _, err := n.Conn.QueueSubscribe(ScanRetrySubj, ScanQueue, func(msg *nats.Msg) {
		n.L.Debug().Msgf("msg.Data received: %s", string(msg.Data))
		// if total retries >= retry qty; send scanFail
		// else; send to scanStart
	}); err != nil {
		n.L.Error().Err(err).Msgf("err subscribing")
		return err
	}
	return nil
}
func (n *Nats) scanFailQueueGroup() error {
	if _, err := n.Conn.QueueSubscribe(ScanFailSubj, ScanQueue, func(msg *nats.Msg) {
		n.L.Debug().Msgf("msg.Data received: %s", string(msg.Data))
		// save to database that scan failed
		// (optional) alert customer scan failed
	}); err != nil {
		n.L.Error().Err(err).Msgf("err subscribing")
		return err
	}
	return nil
}

func (n *Nats) scanError() error {
	return n.Conn.Publish(ScanFailSubj, []byte("scan failed"))
}
