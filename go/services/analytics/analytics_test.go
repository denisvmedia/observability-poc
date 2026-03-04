package analytics_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/denisvmedia/observability-poc/models"
	"github.com/denisvmedia/observability-poc/services/analytics"
)

// good returns a VersionKPIs with all healthy values for a given version string.
func good(version string) models.VersionKPIs {
	return models.VersionKPIs{
		Version:        version,
		SessionCount:   500,
		VSFRate:        0.01,
		VPFRate:        0.01,
		CIRRRate:       0.05,
		AvgVST:         1.0,
		PlayRate:       0.90,
		CompletionRate: 0.85,
	}
}

// --- ComputeRecommendation tests ---

func TestComputeRecommendation_AWinsAll(t *testing.T) {
	c := qt.New(t)
	a := good("1.0")
	b := good("2.0")
	// Make A strictly better on all "lower is better" and all "higher is better"
	b.VSFRate = 0.05
	b.VPFRate = 0.05
	b.CIRRRate = 0.09
	b.AvgVST = 3.0
	b.PlayRate = 0.50
	b.CompletionRate = 0.50

	rec := analytics.ComputeRecommendation(a, b)
	c.Assert(rec.Winner, qt.Equals, "1.0")
	c.Assert(rec.WinsA, qt.Equals, 6)
	c.Assert(rec.WinsB, qt.Equals, 0)
	c.Assert(rec.Dimensions, qt.HasLen, 6)
}

func TestComputeRecommendation_BWinsAll(t *testing.T) {
	c := qt.New(t)
	a := good("1.0")
	b := good("2.0")
	a.VSFRate = 0.05
	a.VPFRate = 0.05
	a.CIRRRate = 0.09
	a.AvgVST = 3.0
	a.PlayRate = 0.50
	a.CompletionRate = 0.50

	rec := analytics.ComputeRecommendation(a, b)
	c.Assert(rec.Winner, qt.Equals, "2.0")
	c.Assert(rec.WinsA, qt.Equals, 0)
	c.Assert(rec.WinsB, qt.Equals, 6)
}

func TestComputeRecommendation_Tie(t *testing.T) {
	c := qt.New(t)
	// Identical values → all ties → WinsA = WinsB = 0
	a := good("1.0")
	b := good("2.0")
	b.VSFRate = a.VSFRate
	b.VPFRate = a.VPFRate
	b.CIRRRate = a.CIRRRate
	b.AvgVST = a.AvgVST
	b.PlayRate = a.PlayRate
	b.CompletionRate = a.CompletionRate

	rec := analytics.ComputeRecommendation(a, b)
	c.Assert(rec.Winner, qt.Equals, "")
	c.Assert(rec.WinsA, qt.Equals, 0)
	c.Assert(rec.WinsB, qt.Equals, 0)
}

func TestComputeRecommendation_4_2_Split(t *testing.T) {
	c := qt.New(t)
	a := good("1.0")
	b := good("2.0")
	// A wins 4: VSF, VPF, CIRR, VST (lower is better, A has lower values)
	b.VSFRate = 0.03
	b.VPFRate = 0.03
	b.CIRRRate = 0.08
	b.AvgVST = 2.5
	// B wins 2: PlayRate, CompletionRate (higher is better, B has higher values)
	b.PlayRate = 0.99
	b.CompletionRate = 0.99

	rec := analytics.ComputeRecommendation(a, b)
	c.Assert(rec.WinsA, qt.Equals, 4)
	c.Assert(rec.WinsB, qt.Equals, 2)
	c.Assert(rec.Winner, qt.Equals, "1.0")
}

func TestComputeRecommendation_PlayRateHigherIsBetter(t *testing.T) {
	c := qt.New(t)
	a := good("1.0")
	b := good("2.0")
	// Only difference: B has a higher play rate
	b.PlayRate = 0.99

	rec := analytics.ComputeRecommendation(a, b)
	// Find the play-rate dimension
	var playDim analytics.KPIDimension
	for _, d := range rec.Dimensions {
		if d.Name == "Play Rate" {
			playDim = d
			break
		}
	}
	c.Assert(playDim.Winner, qt.Equals, "B")
}

// --- ComputeAlerts tests ---

func TestComputeAlerts_NoAttempts(t *testing.T) {
	c := qt.New(t)
	kpis := models.VersionKPIs{Version: "1.0", SessionCount: 0}
	alerts := analytics.ComputeAlerts(kpis)
	c.Assert(alerts, qt.HasLen, 1)
	c.Assert(alerts[0].Code, qt.Equals, analytics.AlertNoAttempts)
}

func TestComputeAlerts_LowSample(t *testing.T) {
	c := qt.New(t)
	kpis := models.VersionKPIs{Version: "1.0", SessionCount: 50}
	alerts := analytics.ComputeAlerts(kpis)
	c.Assert(alerts, qt.HasLen, 1)
	c.Assert(alerts[0].Code, qt.Equals, analytics.AlertLowSample)
}

func TestComputeAlerts_SampleAtThreshold_NoAlert(t *testing.T) {
	c := qt.New(t)
	kpis := models.VersionKPIs{Version: "1.0", SessionCount: 100}
	alerts := analytics.ComputeAlerts(kpis)
	c.Assert(alerts, qt.HasLen, 0)
}

func TestComputeAlerts_HighVSF_Fires(t *testing.T) {
	c := qt.New(t)
	kpis := models.VersionKPIs{Version: "1.0", SessionCount: 500, VSFRate: 0.06}
	alerts := analytics.ComputeAlerts(kpis)
	c.Assert(alerts, qt.HasLen, 1)
	c.Assert(alerts[0].Code, qt.Equals, analytics.AlertHighVSF)
}

func TestComputeAlerts_VSFAtThreshold_NoAlert(t *testing.T) {
	c := qt.New(t)
	kpis := models.VersionKPIs{Version: "1.0", SessionCount: 500, VSFRate: 0.05}
	alerts := analytics.ComputeAlerts(kpis)
	c.Assert(alerts, qt.HasLen, 0)
}

func TestComputeAlerts_HighVPF_Fires(t *testing.T) {
	c := qt.New(t)
	kpis := models.VersionKPIs{Version: "1.0", SessionCount: 500, VPFRate: 0.051}
	alerts := analytics.ComputeAlerts(kpis)
	c.Assert(alerts, qt.HasLen, 1)
	c.Assert(alerts[0].Code, qt.Equals, analytics.AlertHighVPF)
}

func TestComputeAlerts_HighCIRR_Fires(t *testing.T) {
	c := qt.New(t)
	kpis := models.VersionKPIs{Version: "1.0", SessionCount: 500, CIRRRate: 0.101}
	alerts := analytics.ComputeAlerts(kpis)
	c.Assert(alerts, qt.HasLen, 1)
	c.Assert(alerts[0].Code, qt.Equals, analytics.AlertHighCIRR)
}

func TestComputeAlerts_HighVST_Fires(t *testing.T) {
	c := qt.New(t)
	kpis := models.VersionKPIs{Version: "1.0", SessionCount: 500, AvgVST: 5.001}
	alerts := analytics.ComputeAlerts(kpis)
	c.Assert(alerts, qt.HasLen, 1)
	c.Assert(alerts[0].Code, qt.Equals, analytics.AlertHighVST)
}

func TestComputeAlerts_VSTAtThreshold_NoAlert(t *testing.T) {
	c := qt.New(t)
	kpis := models.VersionKPIs{Version: "1.0", SessionCount: 500, AvgVST: 5.0}
	alerts := analytics.ComputeAlerts(kpis)
	c.Assert(alerts, qt.HasLen, 0)
}

func TestComputeAlerts_MultipleConditions_OrderPreserved(t *testing.T) {
	c := qt.New(t)
	kpis := models.VersionKPIs{
		Version:      "1.0",
		SessionCount: 50,    // LOW_SAMPLE
		VSFRate:      0.06,  // HIGH_VSF
		VPFRate:      0.051, // HIGH_VPF
		CIRRRate:     0.101, // HIGH_CIRR
		AvgVST:       5.001, // HIGH_VST
	}
	alerts := analytics.ComputeAlerts(kpis)
	c.Assert(alerts, qt.HasLen, 5)
	c.Assert(alerts[0].Code, qt.Equals, analytics.AlertLowSample)
	c.Assert(alerts[1].Code, qt.Equals, analytics.AlertHighVSF)
	c.Assert(alerts[2].Code, qt.Equals, analytics.AlertHighVPF)
	c.Assert(alerts[3].Code, qt.Equals, analytics.AlertHighCIRR)
	c.Assert(alerts[4].Code, qt.Equals, analytics.AlertHighVST)
}
