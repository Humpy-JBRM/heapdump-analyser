package data

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Heapdump interface {
	GetId() string
	GetTimestamp() time.Time
	GetRoot() string
	GetReports() ([]HeapdumpReport, error)
}

type HeapdumpReport interface {
	GetName() string
	GetZipfilePath() string
}

type heapdumpImpl struct {
	id        string
	timestamp time.Time
	root      string
	reports   map[string]HeapdumpReport
}

func NewHeapdump(hprofPath string) (Heapdump, error) {
	return &heapdumpImpl{}, nil
}

func (h *heapdumpImpl) GetId() string {
	return h.id
}

func (h *heapdumpImpl) GetTimestamp() time.Time {
	return h.timestamp
}

func (h *heapdumpImpl) GetRoot() string {
	return h.root
}

func (h *heapdumpImpl) ParseReports() error {
	// Step 1: find the zip files
	var reportZips = map[string]string{
		"Leak Suspects":   "heapdump_Leak_Suspects.zip",
		"System Overview": "heapdump_System_Overview.zip",
		"Top Components":  "heapdump_Top_Components.zip",
	}
	for name, zipfile := range reportZips {
		hr, err := NewHeapdumpReport(name, filepath.Join(h.root, zipfile))
		if err != nil {
			continue
		}
		h.reports[name] = hr
	}

	return nil
}

func (h *heapdumpImpl) GetReports() ([]HeapdumpReport, error) {
	reportsList := make([]HeapdumpReport, 0)
	for _, report := range h.reports {
		reportsList = append(reportsList, report)
	}

	return reportsList, nil
}

type heapdumpReportImpl struct {
	name    string
	zipfile string
}

func NewHeapdumpReport(name string, zipfile string) (HeapdumpReport, error) {
	_, err := os.Stat(zipfile)
	if err != nil {
		return nil, fmt.Errorf("NewHeapdump[Report(%s, %s): %s", name, zipfile, err.Error())
	}
	hr := &heapdumpReportImpl{
		name:    name,
		zipfile: zipfile,
	}

	return hr, nil
}

func (hr heapdumpReportImpl) GetName() string {
	return hr.name
}

func (hr *heapdumpReportImpl) GetZipfilePath() string {
	return hr.zipfile
}
