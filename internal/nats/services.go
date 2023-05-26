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
			n.StartScan(&scan)
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
	}); err != nil {
		n.L.Error().Err(err).Msgf("err subscribing")
		return err
	}
	return nil
}

func (n *Nats) scanRetryQueueGroup() error {
	if _, err := n.Conn.QueueSubscribe(ScanRetrySubj, ScanQueue, func(msg *nats.Msg) {
		n.L.Debug().Msgf("msg.Data received: %s", string(msg.Data))
	}); err != nil {
		n.L.Error().Err(err).Msgf("err subscribing")
		return err
	}
	return nil
}
func (n *Nats) scanFailQueueGroup() error {
	if _, err := n.Conn.QueueSubscribe(ScanFailSubj, ScanQueue, func(msg *nats.Msg) {
		n.L.Debug().Msgf("msg.Data received: %s", string(msg.Data))
	}); err != nil {
		n.L.Error().Err(err).Msgf("err subscribing")
		return err
	}
	return nil
}

func (n *Nats) scanError() {
	_ = n.Conn.Publish(ScanFailSubj, []byte("scan failed"))
}

func (n *Nats) StartScan(s *api.Scan) {
	scanner, err := services.ScannerFactory([]string{s.Host}, s.Ports)
	if err != nil {
		n.scanError()
		return
	}

	result, warnings, err := scanner.Run()
	if err != nil {
		n.scanError()
		return
	}

	if len(*warnings) > 0 {
		n.L.Warn().Msgf("run finished with warnings: %s\n", *warnings)
	}

	for _, host := range result.Hosts {
		if len(host.Ports) == 0 || len(host.Addresses) == 0 {
			continue
		}

		n.L.Info().Msgf("Host %q:\n", host.Addresses[0])

		for _, port := range host.Ports {
			n.L.Info().Msgf("\tPort %d/%s %s %s\n", port.ID, port.Protocol, port.State, port.Service.Name)
		}
	}
	r, _ := json.Marshal(result)
	_ = n.Conn.Publish(ScanCompleteSubj, r)
}
