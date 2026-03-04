package models

import "time"

// PlaybackSession represents a single playback event row stored in ClickHouse.
type PlaybackSession struct {
	Timestamp     time.Time
	UUID          string
	AppVersion    string
	PlayerVersion string
	PlayerName    string
	Attempts      uint8
	Plays         uint8
	EndedPlays    uint8
	VSF           float64
	VPF           float64
	CIRR          float64
	VST           float64
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
