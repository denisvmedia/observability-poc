// Package ingestion provides XLSX parsing and batch insertion into the session registry.
package ingestion

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/denisvmedia/observability-poc/models"
	"github.com/denisvmedia/observability-poc/registry"
)

const (
	batchSize = 1000
	maxErrors = 50

	colTimestamp     = "timestamp"
	colUUID          = "uuid"
	colAppVersion    = "app_version"
	colPlayerVersion = "player_version"
	colPlayerName    = "player_name"
	colAttempts      = "attempts"
	colPlays         = "plays"
	colEndedPlays    = "ended_plays"
	colVSF           = "vsf"
	colVPF           = "vpf"
	colCIRR          = "cirr"
	colVST           = "vst"
)

// IngestResult holds summary statistics for an ingestion run.
type IngestResult struct {
	RowsInserted int
	RowsSkipped  int
	Errors       []string // per-row parse errors, capped at maxErrors
}

// Ingest parses an XLSX file from r and inserts all valid rows into reg.
func Ingest(ctx context.Context, r io.Reader, reg registry.SessionRegistry) (IngestResult, error) {
	f, err := excelize.OpenReader(r)
	if err != nil {
		return IngestResult{}, fmt.Errorf("ingestion: open xlsx: %w", err)
	}
	defer f.Close()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return IngestResult{}, fmt.Errorf("ingestion: xlsx has no sheets")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return IngestResult{}, fmt.Errorf("ingestion: read rows: %w", err)
	}
	if len(rows) == 0 {
		return IngestResult{}, nil
	}

	colIdx := buildColumnIndex(rows[0])

	var result IngestResult
	buf := make([]models.PlaybackSession, 0, batchSize)

	for _, row := range rows[1:] {
		if isBlankRow(row) {
			continue
		}

		session, parseErr := parseRow(row, colIdx)
		if parseErr != nil {
			result.RowsSkipped++
			if len(result.Errors) < maxErrors {
				result.Errors = append(result.Errors, parseErr.Error())
			}
			continue
		}

		buf = append(buf, session)
		if len(buf) >= batchSize {
			if err = flushBatch(ctx, reg, buf, &result); err != nil {
				return result, err
			}
			buf = buf[:0]
		}
	}

	if len(buf) > 0 {
		if err = flushBatch(ctx, reg, buf, &result); err != nil {
			return result, err
		}
	}

	return result, nil
}

func buildColumnIndex(header []string) map[string]int {
	idx := make(map[string]int, len(header))
	for i, h := range header {
		key := strings.ToLower(strings.TrimSpace(h))
		key = strings.ReplaceAll(key, " ", "_")
		idx[key] = i
	}
	return idx
}

func isBlankRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

func flushBatch(ctx context.Context, reg registry.SessionRegistry, buf []models.PlaybackSession, result *IngestResult) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := reg.InsertBatch(ctx, buf); err != nil {
		return fmt.Errorf("ingestion: insert batch: %w", err)
	}
	result.RowsInserted += len(buf)
	return nil
}

func cellValue(row []string, colIdx map[string]int, name string) string {
	i, ok := colIdx[name]
	if !ok || i >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[i])
}

func parseRow(row []string, colIdx map[string]int) (models.PlaybackSession, error) {
	cell := func(name string) string { return cellValue(row, colIdx, name) }

	ts, err := parseTimestamp(cell(colTimestamp))
	if err != nil {
		return models.PlaybackSession{}, fmt.Errorf("timestamp: %w", err)
	}

	attempts, err := parseUint8(cell(colAttempts))
	if err != nil {
		return models.PlaybackSession{}, fmt.Errorf("attempts: %w", err)
	}

	plays, err := parseUint8(cell(colPlays))
	if err != nil {
		return models.PlaybackSession{}, fmt.Errorf("plays: %w", err)
	}

	endedPlays, err := parseUint8(cell(colEndedPlays))
	if err != nil {
		return models.PlaybackSession{}, fmt.Errorf("ended_plays: %w", err)
	}

	vsf, err := parseFloat64(cell(colVSF))
	if err != nil {
		return models.PlaybackSession{}, fmt.Errorf("vsf: %w", err)
	}

	vpf, err := parseFloat64(cell(colVPF))
	if err != nil {
		return models.PlaybackSession{}, fmt.Errorf("vpf: %w", err)
	}

	cirr, err := parseFloat64(cell(colCIRR))
	if err != nil {
		return models.PlaybackSession{}, fmt.Errorf("cirr: %w", err)
	}

	vst, err := parseFloat64(cell(colVST))
	if err != nil {
		return models.PlaybackSession{}, fmt.Errorf("vst: %w", err)
	}

	return models.PlaybackSession{
		Timestamp:     ts,
		UUID:          cell(colUUID),
		AppVersion:    cell(colAppVersion),
		PlayerVersion: cell(colPlayerVersion),
		PlayerName:    cell(colPlayerName),
		Attempts:      attempts,
		Plays:         plays,
		EndedPlays:    endedPlays,
		VSF:           vsf,
		VPF:           vpf,
		CIRR:          cirr,
		VST:           vst,
	}, nil
}

func parseTimestamp(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}
	for _, layout := range []string{"2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised timestamp format: %q", s)
}

func parseFloat64(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", ".")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float %q: %w", s, err)
	}
	return v, nil
}

func parseUint8(s string) (uint8, error) {
	v, err := strconv.ParseUint(s, 10, 8)
	if err != nil {
		return 0, fmt.Errorf("invalid uint8 %q: %w", s, err)
	}
	return uint8(v), nil
}
