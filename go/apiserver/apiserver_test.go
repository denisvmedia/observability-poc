package apiserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/xuri/excelize/v2"

	"github.com/denisvmedia/observability-poc/apiserver"
	"github.com/denisvmedia/observability-poc/models"
	"github.com/denisvmedia/observability-poc/registry"
)

// --- mock registry ---

type mockRegistry struct {
	mu       sync.Mutex
	sessions []models.PlaybackSession
	err      error
}

var _ registry.SessionRegistry = (*mockRegistry)(nil)

func (m *mockRegistry) InsertBatch(_ context.Context, sessions []models.PlaybackSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.sessions = append(m.sessions, sessions...)
	return nil
}

func (m *mockRegistry) ListVersions(_ context.Context) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return nil, m.err
	}
	seen := make(map[string]struct{})
	var out []string
	for _, s := range m.sessions {
		if _, ok := seen[s.AppVersion]; !ok {
			seen[s.AppVersion] = struct{}{}
			out = append(out, s.AppVersion)
		}
	}
	return out, nil
}

func (m *mockRegistry) GetKPIs(_ context.Context, versions []string) ([]models.VersionKPIs, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return nil, m.err
	}
	counts := make(map[string]int64)
	for _, s := range m.sessions {
		counts[s.AppVersion]++
	}
	out := make([]models.VersionKPIs, 0, len(versions))
	for _, v := range versions {
		out = append(out, models.VersionKPIs{
			Version:        v,
			SessionCount:   counts[v],
			PlayRate:       0.9,
			CompletionRate: 0.8,
		})
	}
	return out, nil
}

// --- xlsx helper ---

func makeXLSX(rows [][]string) []byte {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	for i, row := range rows {
		for j, val := range row {
			coord, _ := excelize.CoordinatesToCellName(j+1, i+1)
			_ = f.SetCellValue(sheet, coord, val)
		}
	}
	var buf bytes.Buffer
	_ = f.Write(&buf)
	return buf.Bytes()
}

var defaultHeaders = []string{
	"timestamp", "uuid", "app_version", "player_version", "player_name",
	"attempts", "plays", "ended_plays", "vsf", "vpf", "cirr", "vst",
}

func goodRow(uuid, version string) []string {
	return []string{"2024-01-15", uuid, version, "player-1", "chrome", "1", "1", "1", "0.01", "0.02", "0.03", "1.5"}
}

func makeMultipartBody(xlsxData []byte) (*bytes.Buffer, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.xlsx")
	_, _ = part.Write(xlsxData)
	_ = writer.Close()
	return body, writer.FormDataContentType()
}

// --- tests ---

func TestHealthz(t *testing.T) {
	c := qt.New(t)
	reg := &mockRegistry{}
	handler := apiserver.New(reg)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	c.Assert(rr.Code, qt.Equals, http.StatusOK)
}

func TestUpload_ValidXLSX(t *testing.T) {
	c := qt.New(t)
	reg := &mockRegistry{}
	handler := apiserver.New(reg)

	rows := [][]string{defaultHeaders, goodRow("u1", "1.0"), goodRow("u2", "2.0")}
	body, ct := makeMultipartBody(makeXLSX(rows))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", body)
	req.Header.Set("Content-Type", ct)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	c.Assert(rr.Code, qt.Equals, http.StatusOK)
	var resp map[string]any
	c.Assert(json.Unmarshal(rr.Body.Bytes(), &resp), qt.IsNil)
	c.Assert(resp["rows_inserted"], qt.Equals, float64(2))
}

func TestUpload_MissingFile(t *testing.T) {
	c := qt.New(t)
	reg := &mockRegistry{}
	handler := apiserver.New(reg)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	c.Assert(rr.Code, qt.Equals, http.StatusBadRequest)
}

func TestVersions_EmptyDB(t *testing.T) {
	c := qt.New(t)
	reg := &mockRegistry{}
	handler := apiserver.New(reg)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/versions", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	c.Assert(rr.Code, qt.Equals, http.StatusOK)
	var versions []string
	c.Assert(json.Unmarshal(rr.Body.Bytes(), &versions), qt.IsNil)
	c.Assert(versions, qt.HasLen, 0)
}

func TestVersions_AfterInsert(t *testing.T) {
	c := qt.New(t)
	reg := &mockRegistry{}
	_ = reg.InsertBatch(context.Background(), []models.PlaybackSession{
		{Timestamp: time.Now(), UUID: "u1", AppVersion: "1.0"},
		{Timestamp: time.Now(), UUID: "u2", AppVersion: "2.0"},
	})
	handler := apiserver.New(reg)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/versions", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	c.Assert(rr.Code, qt.Equals, http.StatusOK)
	var versions []string
	c.Assert(json.Unmarshal(rr.Body.Bytes(), &versions), qt.IsNil)
	c.Assert(versions, qt.HasLen, 2)
}

func TestDashboard_ValidParams(t *testing.T) {
	c := qt.New(t)
	reg := &mockRegistry{}
	_ = reg.InsertBatch(context.Background(), []models.PlaybackSession{
		{Timestamp: time.Now(), UUID: "u1", AppVersion: "1.0"},
		{Timestamp: time.Now(), UUID: "u2", AppVersion: "2.0"},
	})
	handler := apiserver.New(reg)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard?v1=1.0&v2=2.0", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	c.Assert(rr.Code, qt.Equals, http.StatusOK)
	var resp map[string]any
	c.Assert(json.Unmarshal(rr.Body.Bytes(), &resp), qt.IsNil)
	_, hasRec := resp["recommendation"]
	c.Assert(hasRec, qt.IsTrue)
}

func TestDashboard_MissingParams(t *testing.T) {
	c := qt.New(t)
	handler := apiserver.New(&mockRegistry{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	c.Assert(rr.Code, qt.Equals, http.StatusBadRequest)
}

func TestDashboard_SameVersion(t *testing.T) {
	c := qt.New(t)
	handler := apiserver.New(&mockRegistry{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard?v1=1.0&v2=1.0", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	c.Assert(rr.Code, qt.Equals, http.StatusBadRequest)
}
