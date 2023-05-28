package services

import (
	"context"
	"github.com/Ullaakut/nmap/v3"
	"github.com/danielmichaels/onpicket/pkg/api"
	"github.com/rs/zerolog/log"
)

// NmapScanOut overrides nmap.Run and returns only the important information to the user.
type NmapScanOut struct {
	ID          string         `json:"id,omitempty" bson:"id"`
	Args        string         `xml:"args,attr" json:"args"`
	Scanner     string         `xml:"scanner,attr" json:"scanner"`
	StartStr    string         `xml:"startstr,attr" json:"start_str"`
	Version     string         `xml:"version,attr" json:"version"`
	Debugging   nmap.Debugging `xml:"debugging" json:"debugging"`
	Stats       nmap.Stats     `xml:"runstats" json:"run_stats"`
	ScanInfo    nmap.ScanInfo  `xml:"scaninfo" json:"scan_info"`
	Start       nmap.Timestamp `xml:"start,attr" json:"start"`
	Verbose     nmap.Verbose   `xml:"verbose" json:"verbose"`
	Hosts       []nmap.Host    `xml:"host" json:"hosts"`
	PostScripts []nmap.Script  `xml:"postscript>script" json:"post_scripts"`
	PreScripts  []nmap.Script  `xml:"prescript>script" json:"pre_scripts"`
	Targets     []nmap.Target  `xml:"target" json:"targets"`
	TaskBegin   []nmap.Task    `xml:"taskbegin" json:"task_begin"`
	TaskEnd     []nmap.Task    `xml:"taskend" json:"task_end"`
}

// NmapScanIn overrides nmap.Run and captures all the data which comes back from NMAP.
// This data is saved to the data layer.
type NmapScanIn struct {
	ID               string              `json:"id"`
	Args             string              `xml:"args,attr" json:"args"`
	ProfileName      string              `xml:"profile_name,attr" json:"profile_name"`
	Scanner          string              `xml:"scanner,attr" json:"scanner"`
	StartStr         string              `xml:"startstr,attr" json:"start_str"`
	Version          string              `xml:"version,attr" json:"version"`
	XMLOutputVersion string              `xml:"xmloutputversion,attr" json:"xml_output_version"`
	Debugging        nmap.Debugging      `xml:"debugging" json:"debugging"`
	Stats            nmap.Stats          `xml:"runstats" json:"run_stats"`
	ScanInfo         nmap.ScanInfo       `xml:"scaninfo" json:"scan_info"`
	Start            nmap.Timestamp      `xml:"start,attr" json:"start"`
	Verbose          nmap.Verbose        `xml:"verbose" json:"verbose"`
	Hosts            []nmap.Host         `xml:"host" json:"hosts"`
	PostScripts      []nmap.Script       `xml:"postscript>script" json:"post_scripts"`
	PreScripts       []nmap.Script       `xml:"prescript>script" json:"pre_scripts"`
	Targets          []nmap.Target       `xml:"target" json:"targets"`
	TaskBegin        []nmap.Task         `xml:"taskbegin" json:"task_begin"`
	TaskProgress     []nmap.TaskProgress `xml:"taskprogress" json:"task_progress"`
	TaskEnd          []nmap.Task         `xml:"taskend" json:"task_end"`
}

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
	return result, nil
}
