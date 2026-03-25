// Package scanner provides a barcode scanner for comic books via a local
// HTTPS server that can be accessed from a mobile browser on the same network.
//
// Usage:
//
//	cfg := scanner.Config{
//	    Port:     ":8080",
//	    CertFile: "server.pem",
//	    KeyFile:  "server-key.pem",
//	}
//
//	// One-shot: block until a single barcode is scanned
//	result, err := scanner.Start(ctx, cfg)
//	fmt.Println(result.Barcode) // full 17/18 digits
//
//	// Stream: receive each scan as it comes in
//	results, err := scanner.StartStream(ctx, cfg)
//	for r := range results {
//	    fmt.Println(r.Barcode)
//	}
package scanner

import (
	"context"
	"time"
)

// Config holds the configuration for the scanner server.
type Config struct {
	// Port to listen on, e.g. ":8080"
	Port string

	// CertFile and KeyFile are paths to the TLS certificate and key.
	// Generate them with: just gen-cert
	CertFile string
	KeyFile  string
}

// Result represents a successfully scanned barcode.
type Result struct {
	// Barcode is the full scanned code, combining SKU and Supplement:
	//   17 digits: UPC-A (12) + EAN-5 (5)  — single-issue comics
	//   18 digits: EAN-13 (13) + EAN-5 (5) — trade paperbacks (ISBN 978/979)
	//
	Barcode string

	// ScannedAt is the time the barcode was received by the server.
	ScannedAt time.Time
}

// BarcodeType indicates the format of the scanned barcode.
type BarcodeType int

const (
	BarcodeTypeUnknown        BarcodeType = iota
	BarcodeTypeComic                      // 17 digits: UPC-A + EAN-5
	BarcodeTypeTradePaperback             // 18 digits: EAN-13 (978/979) + EAN-5
)

// Type returns the BarcodeType based on the barcode length and prefix.
func (r Result) Type() BarcodeType {
	switch {
	case len(r.Barcode) == 17:
		return BarcodeTypeComic
	case len(r.Barcode) == 18 && (r.Barcode[:3] == "978" || r.Barcode[:3] == "979"):
		return BarcodeTypeTradePaperback
	default:
		return BarcodeTypeUnknown
	}
}

// SKU returns the main barcode portion without the supplement:
//   - 12 digits for single-issue comics (UPC-A)
//   - 13 digits for trade paperbacks (EAN-13 / ISBN)
func (r Result) SKU() string {
	switch r.Type() {
	case BarcodeTypeComic:
		return r.Barcode[:12]
	case BarcodeTypeTradePaperback:
		return r.Barcode[:13]
	default:
		return r.Barcode
	}
}

// Supplement returns the 5-digit EAN-5 addon that encodes issue-specific data:
//   - digits 1-3: issue number
//   - digit 4:    cover variant
//   - digit 5:    printing number
func (r Result) Supplement() string {
	if len(r.Barcode) < 5 {
		return ""
	}
	return r.Barcode[len(r.Barcode)-5:]
}

// Start launches the scanner server and blocks until exactly one barcode is
// scanned or the context is cancelled.
func Start(ctx context.Context, cfg Config) (Result, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results, err := StartStream(ctx, cfg)
	if err != nil {
		return Result{}, err
	}
	select {
	case r, ok := <-results:
		if !ok {
			return Result{}, ctx.Err()
		}
		return r, nil
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}
}

// StartStream launches the scanner server and sends each scanned barcode to
// the returned channel. The channel is closed when the context is cancelled
// or the server encounters a fatal error.
//
// The server runs until the context is cancelled.
func StartStream(ctx context.Context, cfg Config) (<-chan Result, error) {
	results := make(chan Result)
	srv, err := newServer(cfg, results)
	if err != nil {
		return nil, err
	}
	go srv.run(ctx)
	return results, nil
}
