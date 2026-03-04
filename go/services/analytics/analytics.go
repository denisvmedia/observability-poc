// Package analytics provides pure KPI comparison and alert generation functions.
package analytics

import (
	"fmt"
	"math"

	"github.com/denisvmedia/observability-poc/models"
)

// AlertCode identifies a specific quality issue.
type AlertCode string

const (
	// AlertNoAttempts fires when no sessions have been recorded for a version.
	AlertNoAttempts AlertCode = "NO_ATTEMPTS"
	// AlertLowSample fires when the session count is below the statistical minimum.
	AlertLowSample AlertCode = "LOW_SAMPLE"
	// AlertHighVSF fires when the video start failure rate exceeds the threshold.
	AlertHighVSF AlertCode = "HIGH_VSF"
	// AlertHighVPF fires when the video playback failure rate exceeds the threshold.
	AlertHighVPF AlertCode = "HIGH_VPF"
	// AlertHighCIRR fires when the CIRR exceeds the threshold.
	AlertHighCIRR AlertCode = "HIGH_CIRR"
	// AlertHighVST fires when the average video start time exceeds the threshold.
	AlertHighVST AlertCode = "HIGH_VST"
)

// Alert describes a quality issue detected for a specific version.
type Alert struct {
	Code    AlertCode `json:"code"`
	Message string    `json:"message"`
}

// KPIDimension holds the per-dimension comparison result between two versions.
type KPIDimension struct {
	Name        string  `json:"name"`
	VersionA    float64 `json:"version_a"`
	VersionB    float64 `json:"version_b"`
	Winner      string  `json:"winner"` // "A", "B", or "tie"
	LowerBetter bool    `json:"lower_better"`
}

// Recommendation is the overall comparison result between two versions.
type Recommendation struct {
	Winner     string         `json:"winner"` // version string of the winner, or "" if tie
	WinsA      int            `json:"wins_a"`
	WinsB      int            `json:"wins_b"`
	Dimensions []KPIDimension `json:"dimensions"`
	Reason     string         `json:"reason"`
}

const minSampleSize = 100

// ComputeRecommendation compares two versions across 6 KPI dimensions and
// returns a Recommendation declaring the overall winner.
func ComputeRecommendation(a, b models.VersionKPIs) Recommendation {
	type dim struct {
		name        string
		valA, valB  float64
		lowerBetter bool
	}

	dims := []dim{
		{"VSF Rate", a.VSFRate, b.VSFRate, true},
		{"VPF Rate", a.VPFRate, b.VPFRate, true},
		{"CIRR", a.CIRRRate, b.CIRRRate, true},
		{"Avg VST", a.AvgVST, b.AvgVST, true},
		{"Play Rate", a.PlayRate, b.PlayRate, false},
		{"Completion Rate", a.CompletionRate, b.CompletionRate, false},
	}

	var rec Recommendation
	for _, d := range dims {
		kd := KPIDimension{
			Name:        d.name,
			VersionA:    d.valA,
			VersionB:    d.valB,
			LowerBetter: d.lowerBetter,
		}
		kd.Winner = dimensionWinner(d.valA, d.valB, d.lowerBetter)
		switch kd.Winner {
		case "A":
			rec.WinsA++
		case "B":
			rec.WinsB++
		}
		rec.Dimensions = append(rec.Dimensions, kd)
	}

	total := len(dims)
	switch {
	case rec.WinsA > rec.WinsB:
		rec.Winner = a.Version
		rec.Reason = fmt.Sprintf("Version %s wins on %d/%d metrics", a.Version, rec.WinsA, total)
	case rec.WinsB > rec.WinsA:
		rec.Winner = b.Version
		rec.Reason = fmt.Sprintf("Version %s wins on %d/%d metrics", b.Version, rec.WinsB, total)
	default:
		rec.Winner = ""
		rec.Reason = fmt.Sprintf("Tie (%d/%d each)", rec.WinsA, total)
	}

	return rec
}

// floatEpsilon is the tolerance used when comparing aggregated float64 KPI values.
// Values this close are treated as equal (a tie) rather than risking a spurious winner
// due to floating-point rounding in ClickHouse aggregations.
const floatEpsilon = 1e-9

func dimensionWinner(valA, valB float64, lowerBetter bool) string {
	if math.Abs(valA-valB) < floatEpsilon {
		return "tie"
	}
	switch {
	case lowerBetter && valA < valB:
		return "A"
	case lowerBetter && valA > valB:
		return "B"
	case !lowerBetter && valA > valB:
		return "A"
	default:
		return "B"
	}
}

// ComputeAlerts evaluates a single version's KPIs and returns any alerts that apply.
// Alerts are returned in a defined order: NO_ATTEMPTS, LOW_SAMPLE, HIGH_VSF,
// HIGH_VPF, HIGH_CIRR, HIGH_VST.
func ComputeAlerts(kpis models.VersionKPIs) []Alert {
	var alerts []Alert

	if kpis.SessionCount == 0 {
		alerts = append(alerts, Alert{
			Code:    AlertNoAttempts,
			Message: "No sessions recorded for version " + kpis.Version,
		})
		return alerts
	}

	if kpis.SessionCount < minSampleSize {
		alerts = append(alerts, Alert{
			Code:    AlertLowSample,
			Message: fmt.Sprintf("Only %d sessions — statistical confidence is low", kpis.SessionCount),
		})
	}

	if kpis.VSFRate > 0.05 {
		alerts = append(alerts, Alert{
			Code:    AlertHighVSF,
			Message: fmt.Sprintf("VSF rate is %.2f%% (threshold: 5%%)", kpis.VSFRate*100),
		})
	}

	if kpis.VPFRate > 0.05 {
		alerts = append(alerts, Alert{
			Code:    AlertHighVPF,
			Message: fmt.Sprintf("VPF rate is %.2f%% (threshold: 5%%)", kpis.VPFRate*100),
		})
	}

	if kpis.CIRRRate > 0.10 {
		alerts = append(alerts, Alert{
			Code:    AlertHighCIRR,
			Message: fmt.Sprintf("CIRR is %.2f%% (threshold: 10%%)", kpis.CIRRRate*100),
		})
	}

	if kpis.AvgVST > 5.0 {
		alerts = append(alerts, Alert{
			Code:    AlertHighVST,
			Message: fmt.Sprintf("Avg video start time is %.2fs (threshold: 5s)", kpis.AvgVST),
		})
	}

	return alerts
}
