package scanner

import (
	"testing"
)

// TestResultType tests the Type() method of the Result struct to ensure it correctly identifies the barcode type based on the barcode string.
func TestResultType(t *testing.T) {
	tests := []struct {
		name     string
		barcode  string
		expected BarcodeType
	}{
		{"comic book 17 digits", "759123456780" + "12345", BarcodeTypeComic},
		{"tpb 978 code", "9781234567890" + "12345", BarcodeTypeTradePaperback},
		{"tpb 979 code", "9791234567891" + "12345", BarcodeTypeTradePaperback},
		{"18 digits no isbn prefix", "1231234567890" + "12345", BarcodeTypeUnknown},
		{"too short", "123456", BarcodeTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Result{Barcode: tt.barcode}
			if got := r.Type(); got != tt.expected {
				t.Errorf("BarcodeType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestResultSKU tests the SKU() method of the Result struct to ensure it correctly extracts the SKU portion of the barcode based on the barcode type.
func TestResultSKU(t *testing.T) {
	tests := []struct {
		name     string
		barcode  string
		expected string
	}{
		{"comic book", "759123456780" + "12345", "759123456780"},
		{"tpb 978 code", "9781234567890" + "12345", "9781234567890"},
		{"tpb 979 code", "9791234567891" + "12345", "9791234567891"},
		{"unknown type", "1234567890123" + "12345", "123456789012312345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Result{Barcode: tt.barcode}
			if got := r.SKU(); got != tt.expected {
				t.Errorf("SKU() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestResultSupplement tests the Supplement() method of the Result struct to ensure it correctly extracts the supplement portion of the barcode based on the barcode type and length.
func TestResultSupplement(t *testing.T) {
	tests := []struct {
		name     string
		barcode  string
		expected string
	}{
		{"comic book", "759123456780" + "12345", "12345"},
		{"tpb 978 code", "9781234567890" + "12345", "12345"},
		{"tpb 979 code", "9791234567891" + "12345", "12345"},
		{"unknown type", "1234567890123" + "12345", ""},
		{"too short", "1234", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Result{Barcode: tt.barcode}
			if got := r.Supplement(); got != tt.expected {
				t.Errorf("Supplement() = %v, want %v", got, tt.expected)
			}
		})
	}
}
