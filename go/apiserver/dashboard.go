package apiserver

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/denisvmedia/observability-poc/models"
	"github.com/denisvmedia/observability-poc/registry"
	"github.com/denisvmedia/observability-poc/services/analytics"
)

type dashboardVersions struct {
	A models.VersionKPIs `json:"a"`
	B models.VersionKPIs `json:"b"`
}

type dashboardAlerts struct {
	A []analytics.Alert `json:"a"`
	B []analytics.Alert `json:"b"`
}

type dashboardResponse struct {
	Versions       dashboardVersions        `json:"versions"`
	Recommendation analytics.Recommendation `json:"recommendation"`
	Alerts         dashboardAlerts          `json:"alerts"`
}

func dashboardHandler(reg registry.SessionRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v1 := r.URL.Query().Get("v1")
		v2 := r.URL.Query().Get("v2")

		if v1 == "" || v2 == "" {
			http.Error(w, `{"error":"v1 and v2 query parameters are required"}`, http.StatusBadRequest)
			return
		}

		if v1 == v2 {
			http.Error(w, `{"error":"v1 and v2 must be different versions"}`, http.StatusBadRequest)
			return
		}

		kpis, err := reg.GetKPIs(r.Context(), []string{v1, v2})
		if err != nil {
			slog.Error("failed to fetch KPIs", "error", err)
			http.Error(w, `{"error":"failed to fetch KPIs"}`, http.StatusInternalServerError)
			return
		}

		kpiA, kpiB := findKPIs(kpis, v1, v2)

		alertsA := analytics.ComputeAlerts(kpiA)
		alertsB := analytics.ComputeAlerts(kpiB)

		if alertsA == nil {
			alertsA = make([]analytics.Alert, 0)
		}

		if alertsB == nil {
			alertsB = make([]analytics.Alert, 0)
		}

		resp := dashboardResponse{
			Versions:       dashboardVersions{A: kpiA, B: kpiB},
			Recommendation: analytics.ComputeRecommendation(kpiA, kpiB),
			Alerts:         dashboardAlerts{A: alertsA, B: alertsB},
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// findKPIs looks up KPI entries by version string from the slice returned by GetKPIs.
// If a version is missing from the slice, a zero-value KPIs struct is returned.
func findKPIs(kpis []models.VersionKPIs, v1, v2 string) (a, b models.VersionKPIs) {
	a.Version = v1
	b.Version = v2

	for _, k := range kpis {
		switch k.Version {
		case v1:
			a = k
		case v2:
			b = k
		}
	}

	return a, b
}
