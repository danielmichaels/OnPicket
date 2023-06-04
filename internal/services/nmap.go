package services

import (
	"context"
	"github.com/Ullaakut/nmap/v3"
	"github.com/danielmichaels/onpicket/pkg/api"
	"github.com/rs/zerolog/log"
	"math"
)

type NmapMetadata struct {
	CurrentPage  int64 `json:"current_page,omitempty"`
	PageSize     int64 `json:"page_size,omitempty"`
	FirstPage    int64 `json:"first_page,omitempty"`
	LastPage     int64 `json:"last_page,omitempty"`
	TotalRecords int64 `json:"total_records,omitempty"`
}

func CalculateMetadata(totalRecords int64, page, pageSize int64) NmapMetadata {
	if totalRecords == 0 {
		return NmapMetadata{}
	}
	return NmapMetadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int64(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}

// NmapRun represents the result of a nmap.Run result but with only the information
// a user really needs.
type NmapRun struct {
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

// NmapScan overrides nmap.Run and returns only the important information to the user.
type NmapScan struct {
	ID          string          `json:"id,omitempty" bson:"id"`
	Status      string          `json:"status,omitempty" bson:"status"`
	Summary     string          `json:"summary,omitempty" bson:"summary"`
	ScanType    api.NewScanType `json:"scan_type,omitempty" bson:"scan_type"`
	Description string          `json:"description,omitempty" bson:"description"`
	HostsArray  []string        `json:"hosts_array,omitempty" bson:"hosts_array"`
	Ports       []string        `json:"ports" bson:"ports"`
	Scan        NmapRun         `json:"data,omitempty" bson:"data"`
}

// StartScan is the entrypoint to creating Scan. A cancellable timeout and api.Scan
// must be passed.
func StartScan(ctx context.Context, s *api.Scan) (*nmap.Run, error) {
	scanner, err := ScannerFactory(ctx, s.HostsArray, s.Ports, s.ScanType)
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

// ScannerFactory is the factory which creates nmap.Scanner's.
func ScannerFactory(ctx context.Context, targets, ports []string, scanType string) (*nmap.Scanner, error) {
	switch scanType {
	case string(api.PortScan):
		return nmap.NewScanner(
			ctx,
			nmap.WithTargets(targets...),
			nmap.WithPorts(ports...),
		)
	case string(api.ServiceDiscovery):
		return nmap.NewScanner(
			ctx,
			nmap.WithTargets(targets...),
			nmap.WithPorts(ports...),
			nmap.WithServiceInfo(),
		)
	case string(api.ServiceDiscoveryDefaultScripts):
		return nmap.NewScanner(
			ctx,
			nmap.WithTargets(targets...),
			nmap.WithPorts(ports...),
			nmap.WithServiceInfo(),
			nmap.WithDefaultScript(),
		)
	default:
	}
	return nmap.NewScanner(
		ctx,
		nmap.WithTargets(targets...),
		nmap.WithPorts(ports...),
	)
}
