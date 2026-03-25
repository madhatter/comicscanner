package scanner

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"
)

//go:embed static/index.html static/style.css
var staticFiles embed.FS

// server is the internal HTTP server. Not exported — callers use Start/StartStream.
type server struct {
	cfg     Config
	results chan<- Result
	http    *http.Server
}

func newServer(cfg Config, results chan<- Result) (*server, error) {
	s := &server{
		cfg:     cfg,
		results: results,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/scanner", s.handleUI)
	mux.HandleFunc("/style.css", s.handleCSS)
	mux.HandleFunc("/api/scan", s.handleScan)

	s.http = &http.Server{
		Addr:    cfg.Port,
		Handler: mux,
	}

	return s, nil
}

// run starts the TLS server and blocks until the context is cancelled.
func (s *server) run(ctx context.Context) {
	// Shut down gracefully when context is cancelled
	go func() {
		<-ctx.Done()
		log.Println("[scanner] Shutting down...")
		_ = s.http.Shutdown(context.Background())
		close(s.results)
	}()

	log.Printf("[scanner] Listening on https://localhost%s/scanner\n", s.cfg.Port)

	if err := s.http.ListenAndServeTLS(s.cfg.CertFile, s.cfg.KeyFile); err != nil && err != http.ErrServerClosed {
		log.Printf("[scanner] Server error: %v\n", err)
	}
}

// handleUI serves the scanner HTML interface.
func (s *server) handleUI(w http.ResponseWriter, r *http.Request) {
	data, err := fs.ReadFile(staticFiles, "static/index.html")
	if err != nil {
		http.Error(w, "Could not load scanner UI", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

// handleCSS serves the stylesheet.
func (s *server) handleCSS(w http.ResponseWriter, r *http.Request) {
	data, err := fs.ReadFile(staticFiles, "static/style.css")
	if err != nil {
		http.Error(w, "Could not load stylesheet", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write(data)
}

// scanPayload is the JSON body sent by the browser.
type scanPayload struct {
	Barcode string `json:"barcode"`
}

// handleScan receives the scanned barcode from the browser and forwards it
// to the results channel.
func (s *server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload scanPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	barcode := strings.TrimSpace(payload.Barcode)

	// Validate length: 17 digits (comic) or 18 digits (trade paperback)
	if len(barcode) < 17 {
		http.Error(w, "Barcode too short — missing EAN-5 supplement", http.StatusBadRequest)
		return
	}

	result := Result{
		Barcode:   barcode,
		ScannedAt: time.Now(),
	}

	// Log to stdout so the host application can see incoming scans
	fmt.Printf("[scanner] Scanned: %s  type=%s  sku=%s  supplement=%s\n",
		result.Barcode,
		barcodeTypeName(result.Type()),
		result.SKU(),
		result.Supplement(),
	)

	// Non-blocking send — if nobody is reading the channel yet, we don't hang
	select {
	case s.results <- result:
	default:
		log.Printf("[scanner] Warning: result channel full, dropping scan %s\n", barcode)
	}

	w.WriteHeader(http.StatusOK)
}

func barcodeTypeName(t BarcodeType) string {
	switch t {
	case BarcodeTypeComic:
		return "comic"
	case BarcodeTypeTradePaperback:
		return "trade-paperback"
	default:
		return "unknown"
	}
}
