package natsio

import (
	"context"
	"encoding/json"
	"github.com/danielmichaels/onpicket/internal/database"
	"github.com/danielmichaels/onpicket/internal/funcs"
	"github.com/danielmichaels/onpicket/internal/services"
	"github.com/danielmichaels/onpicket/pkg/api"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var defaultScanTimeoutInSecs = 300

const (
	ScanQueue        = "scans"
	ScanStartSubj    = "scan.start"
	ScanCompleteSubj = "scan.complete"
	ScanRetrySubj    = "scan.retry"
	ScanFailSubj     = "scan.fail"
)

func (n *Nats) InitSubscribers() error {
	if err := n.startScanQueueGroup(); err != nil {
		return err
	}
	if err := n.retryScanQueueGroup(); err != nil {
		return err
	}
	if err := n.completeScanQueueGroup(); err != nil {
		return err
	}
	if err := n.failScanQueueGroup(); err != nil {
		return err
	}
	return nil
}

type ScanError struct {
	Message string   `json:"message,omitempty"`
	Scan    api.Scan `json:"scan"`
}

func (n *Nats) startScanQueueGroup() error {
	if _, err := n.Conn.QueueSubscribe(ScanStartSubj, ScanQueue, func(msg *nats.Msg) {
		n.L.Debug().Msgf("%q received: %s", ScanStartSubj, string(msg.Data))
		var scan api.Scan
		err := json.Unmarshal(msg.Data, &scan)
		if err != nil {
			n.L.Error().Err(err).Msgf("err unmarshalling published message")
			return
		}
		funcs.BackgroundFunc(func() {
			filter := bson.D{{Key: "id", Value: scan.Id}}
			opts := options.Update().SetUpsert(true)
			fields := bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "status", Value: string(api.InProgress)},
					{Key: "id", Value: scan.Id},
					{Key: "scan_type", Value: scan.Type},
					{Key: "description", Value: scan.Description},
					{Key: "scan_hosts", Value: scan.Hosts},
				}},
			}
			_, err = n.DB.Collection(database.ScanCollection).UpdateOne(
				context.TODO(),
				filter,
				fields,
				opts,
			)
			if err != nil {
				n.L.Error().Err(err).Msg("scan status update error")
				return
			}

			timeout := defaultScanTimeoutInSecs
			if scan.Timeout != nil {
				timeout = *scan.Timeout
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
			defer cancel()

			sRes, err := services.StartScan(ctx, &scan)
			select {
			case <-ctx.Done():
				err = n.scanError(&scan, ctx.Err())
				if err != nil {
					return
				}
				return
			default:
				if err != nil {
					err = n.scanError(&scan, err)
					if err != nil {
						return
					}
					return
				}
			}

			nmapResult := services.NmapScanIn{
				ID:               scan.Id,
				Status:           string(api.Complete),
				ScanType:         scan.Type,
				Description:      scan.Description,
				ScanHosts:        scan.Hosts,
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

func (n *Nats) completeScanQueueGroup() error {
	if _, err := n.Conn.QueueSubscribe(ScanCompleteSubj, ScanQueue, func(msg *nats.Msg) {
		n.L.Debug().Msgf("%q received: %s", ScanCompleteSubj, string(msg.Data))

		var s services.NmapScanIn
		err := json.Unmarshal(msg.Data, &s)
		if err != nil {
			n.L.Error().Err(err).Msgf("err: unmarshalling NATS message")
			return
		}
		filter := bson.D{{Key: "id", Value: s.ID}}
		opts := options.Update().SetUpsert(true)
		fields := bson.D{
			{Key: "$set", Value: s},
		}
		_, err = n.DB.Collection(database.ScanCollection).UpdateOne(
			context.TODO(),
			filter,
			fields,
			opts,
		)
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

func (n *Nats) retryScanQueueGroup() error {
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
func (n *Nats) failScanQueueGroup() error {
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

func (n *Nats) scanError(data *api.Scan, e error) error {
	sc := ScanError{
		Message: e.Error(),
		Scan:    *data,
	}
	b, err := json.Marshal(sc)
	if err != nil {
		return err
	}
	return n.Conn.Publish(ScanFailSubj, b)
}
