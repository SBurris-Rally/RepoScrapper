package data

import (
    "time"
)

type PrAnalysis struct {
	RepoName string
	Status string
	Team string
	JiraProject string
	IsLive string
	PrTitle string
	CreatedAt time.Time
	UpdatedAt time.Time
	ClosedAt time.Time
	MergedAt time.Time
	TicketFound bool
	Ticket string
	IsTicketValid bool
}


type Settings struct {
	FileSystemBaseFolder string
	OutputFile string
}





type RepositoryFile struct {
	Path string
	Mode string
	Type string
	Sha string
	Size int
	URL string
	DownloadedDate time.Time
}

type ConftestScanResult []struct {
	Filename  string `json:"filename"`
	Namespace string `json:"namespace"`
	Successes int    `json:"successes"`
	Warnings  []struct {
		Msg string `json:"msg"`
	} `json:"warnings,omitempty"`
	Failures []struct {
		Msg string `json:"msg"`
	} `json:"failures,omitempty"`
}