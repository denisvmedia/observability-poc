package clickhouse

import (
	"context"
	"fmt"

	"github.com/denisvmedia/observability-poc/models"
)

// InsertBatch inserts a batch of playback sessions into ClickHouse.
func (r *Registry) InsertBatch(ctx context.Context, sessions []models.PlaybackSession) error {
	batch, err := r.conn.PrepareBatch(ctx, "INSERT INTO playback_sessions")
	if err != nil {
		return fmt.Errorf("clickhouse: prepare batch: %w", err)
	}

	for _, s := range sessions {
		if err = batch.AppendStruct(&s); err != nil {
			return fmt.Errorf("clickhouse: append row: %w", err)
		}
	}

	if err = batch.Send(); err != nil {
		return fmt.Errorf("clickhouse: send batch: %w", err)
	}

	return nil
}

// ListVersions returns distinct app_version values ordered alphabetically.
func (r *Registry) ListVersions(ctx context.Context) ([]string, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT DISTINCT app_version
		FROM playback_sessions
		ORDER BY app_version
	`)
	if err != nil {
		return nil, fmt.Errorf("clickhouse: list versions: %w", err)
	}
	defer rows.Close()

	var versions []string
	for rows.Next() {
		var v string
		if err = rows.Scan(&v); err != nil {
			return nil, fmt.Errorf("clickhouse: scan version: %w", err)
		}
		versions = append(versions, v)
	}

	return versions, rows.Err()
}

// GetKPIs returns aggregated quality metrics for the requested versions.
func (r *Registry) GetKPIs(ctx context.Context, versions []string) ([]models.VersionKPIs, error) {
	rows, err := r.conn.Query(ctx, `
		SELECT
			app_version,
			count()                                                          AS session_count,
			sumIf(vsf,  attempts = 1) / countIf(attempts = 1)               AS vsf_rate,
			sumIf(vpf,  plays = 1)    / countIf(plays = 1)                  AS vpf_rate,
			sumIf(cirr, plays = 1)    / countIf(plays = 1)                  AS cirr_rate,
			avgIf(vst,  attempts = 1)                                        AS avg_vst,
			sumIf(plays,       attempts = 1) / countIf(attempts = 1)        AS play_rate,
			sumIf(ended_plays, plays = 1)    / countIf(plays = 1)           AS completion_rate
		FROM playback_sessions
		WHERE app_version IN (?)
		GROUP BY app_version
	`, versions)
	if err != nil {
		return nil, fmt.Errorf("clickhouse: get kpis: %w", err)
	}
	defer rows.Close()

	index := make(map[string]models.VersionKPIs, len(versions))
	for rows.Next() {
		var kpi models.VersionKPIs
		if err = rows.Scan(
			&kpi.Version,
			&kpi.SessionCount,
			&kpi.VSFRate,
			&kpi.VPFRate,
			&kpi.CIRRRate,
			&kpi.AvgVST,
			&kpi.PlayRate,
			&kpi.CompletionRate,
		); err != nil {
			return nil, fmt.Errorf("clickhouse: scan kpi row: %w", err)
		}
		index[kpi.Version] = kpi
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("clickhouse: iterate kpi rows: %w", err)
	}

	result := make([]models.VersionKPIs, 0, len(versions))
	for _, v := range versions {
		if kpi, ok := index[v]; ok {
			result = append(result, kpi)
		} else {
			result = append(result, models.VersionKPIs{Version: v})
		}
	}

	return result, nil
}
