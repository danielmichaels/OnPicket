package services

import (
	"context"
	"github.com/Ullaakut/nmap/v3"
	"github.com/danielmichaels/onpicket/pkg/api"
	"github.com/rs/zerolog/log"
)

func ScannerFactory(targets, ports []string, scanType string) (*nmap.Scanner, error) {
	// todo: this needs to be capable of accepting any number of options
	// e.g. verbosity, NSE scripts by name and so on.
	switch scanType {
	case string(api.PortScan):
		return nmap.NewScanner(
			context.Background(),
			nmap.WithTargets(targets...),
			nmap.WithPorts(ports...),
		)
	case string(api.ServiceDiscovery):
		return nmap.NewScanner(
			context.Background(),
			nmap.WithTargets(targets...),
			nmap.WithPorts(ports...),
			nmap.WithServiceInfo(),
		)
	case string(api.ServiceDiscoveryDefaultScripts):
		return nmap.NewScanner(
			context.Background(),
			nmap.WithTargets(targets...),
			nmap.WithPorts(ports...),
			nmap.WithServiceInfo(),
			nmap.WithDefaultScript(),
		)
	default:
	}
	return nmap.NewScanner(
		context.Background(),
		nmap.WithTargets(targets...),
		nmap.WithPorts(ports...),
	)
}

func StartScan(s *api.Scan) (*nmap.Run, error) {
	scanner, err := ScannerFactory([]string{s.Host}, s.Ports, s.Type)
	if err != nil {
		return nil, err
	}

	result, warnings, err := scanner.Run()
	if err != nil {
		return nil, err
	}

	if len(*warnings) > 0 {
		log.Warn().Msgf("run finished with warnings: %s\n", *warnings)
	}

	//for _, host := range result.Hosts {
	//	if len(host.Ports) == 0 || len(host.Addresses) == 0 {
	//		continue
	//	}
	//
	//	log.Info().Msgf("Host %q:\n", host.Addresses[0])
	//
	//	for _, port := range host.Ports {
	//		log.Info().Msgf("\tPort %d/%s %s %s\n", port.ID, port.Protocol, port.State, port.Service.Name)
	//	}
	//}
	return result, nil
}
