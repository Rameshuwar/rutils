package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	// Blank import so all formatters register themselves before the tests run.
	_ "file-converter/internal/formatters"

	formatters "file-converter/internal/formatters"
)

func TestListFormats_ReturnsMatrix(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/formats", nil)
	rec := httptest.NewRecorder()

	HandleListFormats(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var body formatters.FormatListResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(body.Sources) == 0 {
		t.Fatal("expected non-empty sources map")
	}
	if len(body.Conversions) == 0 {
		t.Fatal("expected non-empty conversions list")
	}

	// Sanity-check a few known formats.
	for _, want := range []string{"jpg", "png", "pdf", "csv", "json"} {
		if _, ok := body.Sources[want]; !ok {
			t.Errorf("expected source %q to be present in sources map", want)
		}
	}

	// The image-to-image edges we just added must all appear.
	mustHaveEdge := func(from, to string) {
		for _, c := range body.Conversions {
			if c.From == from && c.To == to {
				return
			}
		}
		t.Errorf("expected conversion %s→%s to be present", from, to)
	}
	mustHaveEdge("jpg", "png")
	mustHaveEdge("png", "jpg")
	mustHaveEdge("webp", "jpg")
	mustHaveEdge("tiff", "png")
	mustHaveEdge("bmp", "jpg")
	mustHaveEdge("pdf", "png")
	mustHaveEdge("jpg", "pdf")
	mustHaveEdge("jpg", "txt")
}

func TestListFormats_RejectsNonGET(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/formats", nil)
	rec := httptest.NewRecorder()

	HandleListFormats(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestFormatDetail_KnownType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/formats/jpg", nil)
	rec := httptest.NewRecorder()

	HandleFormatDetail(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var detail formatters.FormatDetail
	if err := json.NewDecoder(rec.Body).Decode(&detail); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if detail.Type != "jpg" {
		t.Errorf("expected type=jpg, got %q", detail.Type)
	}
	if detail.MIME == "" || detail.Ext == "" {
		t.Errorf("expected non-empty MIME/ext, got MIME=%q Ext=%q", detail.MIME, detail.Ext)
	}
	if len(detail.CanConvertTo) == 0 {
		t.Errorf("expected jpg to have at least one reachable target")
	}

	// jpg → png must be reachable.
	foundPNG := false
	for _, to := range detail.CanConvertTo {
		if to == "png" {
			foundPNG = true
			break
		}
	}
	if !foundPNG {
		t.Errorf("expected jpg's canConvertTo to include png")
	}
}

func TestFormatDetail_UnknownType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/formats/madeup", nil)
	rec := httptest.NewRecorder()

	HandleFormatDetail(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "unknown format") {
		t.Errorf("expected error to mention unknown format, got %q", rec.Body.String())
	}
}

func TestFormatDetail_MissingType(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/formats/", nil)
	rec := httptest.NewRecorder()

	HandleFormatDetail(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
