package models

import "time"

// PlaybackSession represents a single playback event row stored in ClickHouse.
type PlaybackSession struct {
	Timestamp     time.Time `ch:"timestamp"`
	UUID          string    `ch:"uuid"`
	AppVersion    string    `ch:"app_version"`
	PlayerVersion string    `ch:"player_version"`
	PlayerName    string    `ch:"player_name"`
	Attempts      uint8     `ch:"attempts"`
	Plays         uint8     `ch:"plays"`
	EndedPlays    uint8     `ch:"ended_plays"`
	VSF           float64   `ch:"vsf"`
	VPF           float64   `ch:"vpf"`
	CIRR          float64   `ch:"cirr"`
	VST           float64   `ch:"vst"`
}

// VersionKPIs holds aggregated quality metrics for a single app version.
type VersionKPIs struct {
	Version        string
	SessionCount   int64
	VSFRate        float64
	VPFRate        float64
	CIRRRate       float64
	AvgVST         float64
	PlayRate       float64
	CompletionRate float64
}

// Alert describes a quality issue detected for a specific version.
type Alert struct {
	Version string
	Code    string
	Message string
}
