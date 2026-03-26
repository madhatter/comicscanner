package scanner

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer creates a new server instance and a channel to capture results for testing.
func newTestServer() (*server, chan Result) {
	ch := make(chan Result, 1)
	s := &server{
		results: ch,
	}
	return s, ch
}

// TestHandleScan_InvalidMethod tests that the server returns a 405 Method Not Allowed status for unsupported HTTP methods.
func TestHandleScan_InvalidMethod(t *testing.T) {
	srv, _ := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/scan", nil)
	w := httptest.NewRecorder()

	srv.handleScan(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

// TestHandleScan_ValidRequest tests that the server correctly processes a valid scan request and sends the expected result to the channel.
func TestHandleScan_ValidRequest(t *testing.T) {
	srv, ch := newTestServer()

	body := strings.NewReader(`{"barcode":"75912345678912345"}`)

	req := httptest.NewRequest(http.MethodPost, "/api/scan", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleScan(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	select {
	case res := <-ch:
		if res.Barcode != "759123456789"+"12345" {
			t.Errorf("Expected barcode %s, got %s", "759123456789"+"12345", res.Barcode)
		}
	default:
		t.Error("Expected a result to be sent to the channel")
	}
}

// TestHandleScan_InvalidJSON tests that the server returns a 400 Bad Request status when the request body contains invalid JSON.
func TestHandleScan_InvalidJSON(t *testing.T) {
	srv, _ := newTestServer()

	body := strings.NewReader(`{"barcode":12345}`) // Invalid JSON: barcode should be a string

	req := httptest.NewRequest(http.MethodPost, "/api/scan", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleScan(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestHandleScan_InvalidBarcode tests that the server returns a 400 Bad Request status when the barcode in the request body is invalid (e.g., too short).
func TestHandleScan_InvalidBarcode(t *testing.T) {
	srv, _ := newTestServer()

	body := strings.NewReader(`{"barcode":"12345"}`) // Invalid barcode: too short

	req := httptest.NewRequest(http.MethodPost, "/api/scan", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.handleScan(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestBarcodeTypeName tests the barcodeTypeName function to ensure it returns the correct string representation for each BarcodeType.
func TestBarcodeTypeName(t *testing.T) {
	tests := []struct {
		name     string
		bt       BarcodeType
		expected string
	}{
		{"unknown", BarcodeTypeUnknown, "unknown"},
		{"comic book", BarcodeTypeComic, "comic"},
		{"trade paperback", BarcodeTypeTradePaperback, "trade-paperback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := barcodeTypeName(tt.bt); got != tt.expected {
				t.Errorf("barcodeTypeName() = %v, want %v", got, tt.expected)
			}
		})
	}
}
