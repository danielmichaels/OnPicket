package natsio

import (
	"context"
	"encoding/json"
	"github.com/danielmichaels/onpicket/internal/database"
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
		n.L.Debug().Msgf("%q received: %s", ScanStartSubj, string(msg.Data))
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

			nmapResult := services.NmapScanIn{
				ID:               scan.Id,
				Args:             sRes.Args,
				ProfileName:      sRes.ProfileName,
				Scanner:          sRes.Scanner,
				StartStr:         sRes.StartStr,
				Version:          sRes.Version,
				XMLOutputVersion: sRes.XMLOutputVersion,
				Debugging:        sRes.Debugging,
				Stats:            sRes.Stats,
				ScanInfo:         sRes.ScanInfo,
				Start:            sRes.Start,
				Verbose:          sRes.Verbose,
				Hosts:            sRes.Hosts,
				PostScripts:      sRes.PostScripts,
				PreScripts:       sRes.PreScripts,
				Targets:          sRes.Targets,
				TaskBegin:        sRes.TaskBegin,
				TaskProgress:     sRes.TaskProgress,
				TaskEnd:          sRes.TaskEnd,
			}

			r, err := json.Marshal(nmapResult)
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
		n.L.Debug().Msgf("%q received: %s", ScanCompleteSubj, string(msg.Data))

		var s services.NmapScanIn
		err := json.Unmarshal(msg.Data, &s)
		if err != nil {
			n.L.Error().Err(err).Msgf("err: unmarshalling NATS message")
			return
		}
		_, err = n.DB.Collection(database.ScanCollection).InsertOne(context.TODO(), s)
		if err != nil {
			n.L.Error().Err(err).Msgf("err: inserting document")
			return
		}
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
		n.L.Debug().Msgf("%q received: %s", ScanRetrySubj, string(msg.Data))
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
		n.L.Debug().Msgf("%q received: %s", ScanFailSubj, string(msg.Data))
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
