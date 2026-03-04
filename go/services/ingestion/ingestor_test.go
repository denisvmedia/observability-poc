package ingestion_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/xuri/excelize/v2"

	"github.com/denisvmedia/observability-poc/models"
	"github.com/denisvmedia/observability-poc/registry"
	"github.com/denisvmedia/observability-poc/services/ingestion"
)

// defaultHeaders is the canonical column order used by most tests.
var defaultHeaders = []string{
	"timestamp", "uuid", "app_version", "player_version", "player_name",
	"attempts", "plays", "ended_plays", "vsf", "vpf", "cirr", "vst",
}

// goodRow returns a valid data row for use with defaultHeaders.
func goodRow(uuid, version string) []string {
	return []string{"2024-01-15", uuid, version, "player-1", "chrome", "1", "1", "1", "0.01", "0.02", "0.03", "1.5"}
}

// makeXLSX builds an in-memory XLSX reader from the given rows.
func makeXLSX(rows [][]string) io.Reader {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	for i, row := range rows {
		for j, val := range row {
			coord, _ := excelize.CoordinatesToCellName(j+1, i+1)
			_ = f.SetCellValue(sheet, coord, val)
		}
	}
	buf := &bytes.Buffer{}
	_ = f.Write(buf)
	return buf
}

// mockRegistry is a test double that records InsertBatch calls.
type mockRegistry struct {
	mu      sync.Mutex
	batches [][]models.PlaybackSession
	err     error
}

var _ registry.SessionRegistry = (*mockRegistry)(nil)

func (m *mockRegistry) InsertBatch(_ context.Context, sessions []models.PlaybackSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	cp := make([]models.PlaybackSession, len(sessions))
	copy(cp, sessions)
	m.batches = append(m.batches, cp)
	return nil
}

func (m *mockRegistry) ListVersions(_ context.Context) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil, nil
}

func (m *mockRegistry) GetKPIs(_ context.Context, _ []string) ([]models.VersionKPIs, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return nil, nil
}

func TestIngest_HappyPath(t *testing.T) {
	c := qt.New(t)

	rows := [][]string{defaultHeaders, goodRow("u1", "1.0"), goodRow("u2", "1.0"), goodRow("u3", "2.0")}
	reg := &mockRegistry{}
	result, err := ingestion.Ingest(context.Background(), makeXLSX(rows), reg)

	c.Assert(err, qt.IsNil)
	c.Assert(result.RowsInserted, qt.Equals, 3)
	c.Assert(result.RowsSkipped, qt.Equals, 0)
}

func TestIngest_CommaDecimalFloat(t *testing.T) {
	c := qt.New(t)

	row := []string{"2024-01-15", "u1", "1.0", "p", "chrome", "1", "1", "1", "0,001", "0,002", "0,003", "1,5"}
	reg := &mockRegistry{}
	result, err := ingestion.Ingest(context.Background(), makeXLSX([][]string{defaultHeaders, row}), reg)

	c.Assert(err, qt.IsNil)
	c.Assert(result.RowsInserted, qt.Equals, 1)
	c.Assert(reg.batches[0][0].VSF, qt.Equals, 0.001)
}

func TestIngest_HeaderRowNotCounted(t *testing.T) {
	c := qt.New(t)

	reg := &mockRegistry{}
	result, err := ingestion.Ingest(context.Background(), makeXLSX([][]string{defaultHeaders}), reg)

	c.Assert(err, qt.IsNil)
	c.Assert(result.RowsInserted, qt.Equals, 0)
}

func TestIngest_UnparseableFloat_Skipped(t *testing.T) {
	c := qt.New(t)

	bad := []string{"2024-01-15", "u1", "1.0", "p", "chrome", "1", "1", "1", "not-a-float", "0", "0", "0"}
	rows := [][]string{defaultHeaders, goodRow("u2", "1.0"), bad, goodRow("u3", "1.0")}
	reg := &mockRegistry{}
	result, err := ingestion.Ingest(context.Background(), makeXLSX(rows), reg)

	c.Assert(err, qt.IsNil)
	c.Assert(result.RowsInserted, qt.Equals, 2)
	c.Assert(result.RowsSkipped, qt.Equals, 1)
	c.Assert(result.Errors, qt.HasLen, 1)
}

func TestIngest_EmptyTimestamp_Skipped(t *testing.T) {
	c := qt.New(t)

	noTS := []string{"", "u1", "1.0", "p", "chrome", "1", "1", "1", "0", "0", "0", "0"}
	reg := &mockRegistry{}
	result, err := ingestion.Ingest(context.Background(), makeXLSX([][]string{defaultHeaders, noTS}), reg)

	c.Assert(err, qt.IsNil)
	c.Assert(result.RowsSkipped, qt.Equals, 1)
}

func TestIngest_ColumnDetectionByName(t *testing.T) {
	c := qt.New(t)

	// Shuffle column order: vst first, then the rest
	headers := []string{"vst", "cirr", "vpf", "vsf", "ended_plays", "plays", "attempts", "player_name", "player_version", "app_version", "uuid", "timestamp"}
	row := []string{"1.5", "0.03", "0.02", "0.01", "1", "1", "1", "chrome", "player-1", "3.0", "u1", "2024-06-01"}
	reg := &mockRegistry{}
	result, err := ingestion.Ingest(context.Background(), makeXLSX([][]string{headers, row}), reg)

	c.Assert(err, qt.IsNil)
	c.Assert(result.RowsInserted, qt.Equals, 1)
	c.Assert(reg.batches[0][0].AppVersion, qt.Equals, "3.0")
	c.Assert(reg.batches[0][0].VST, qt.Equals, 1.5)
}

func TestIngest_BatchBoundary(t *testing.T) {
	c := qt.New(t)

	const total = 1001
	rows := make([][]string, total+1)
	rows[0] = defaultHeaders
	for i := range total {
		rows[i+1] = goodRow(fmt.Sprintf("u%d", i), "1.0")
	}

	reg := &mockRegistry{}
	result, err := ingestion.Ingest(context.Background(), makeXLSX(rows), reg)

	c.Assert(err, qt.IsNil)
	c.Assert(result.RowsInserted, qt.Equals, total)
	c.Assert(reg.batches, qt.HasLen, 2) // 1000 + 1
}
