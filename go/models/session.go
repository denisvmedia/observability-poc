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
	Version        string  `json:"version"`
	SessionCount   uint64  `json:"session_count"`
	VSFRate        float64 `json:"vsf_rate"`
	VPFRate        float64 `json:"vpf_rate"`
	CIRRRate       float64 `json:"cirr_rate"`
	AvgVST         float64 `json:"avg_vst"`
	PlayRate       float64 `json:"play_rate"`
	CompletionRate float64 `json:"completion_rate"`
}

// Alert describes a quality issue detected for a specific version.
type Alert struct {
	Version string
	Code    string
	Message string
}
