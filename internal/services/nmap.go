package services

import (
	"context"
	"github.com/Ullaakut/nmap/v3"
)

/**

Some ideas,

A RunScan factory

NewRunScan(targets..., ports []string, type string) {
	var ctx context.Background()
	var p string
	if type == api.ServiceDiscovery {
		NewScanner(
			ctx,
			nmap.WithTargets(targets...),
			nmap.WithPorts(ports...)
			WithServiceInfo()
		)
	}
	// something like this...
}

**/

func ScannerFactory(targets, ports []string) (*nmap.Scanner, error) {
	s, err := nmap.NewScanner(
		context.Background(),
		nmap.WithTargets(targets...),
		nmap.WithPorts(ports...),
		//nmap.WithServiceInfo(),
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}
