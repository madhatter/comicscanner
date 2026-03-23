package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// ScanPayload represents the expected JSON structure from the HTML scanner
type ScanPayload struct {
	Barcode string `json:"barcode"`
}

// serveHtmlInterface serves the static HTML scanner file to the mobile browser
func serveHtmlInterface(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

// handleScanEndpoint receives the POST request with the barcode data and validates it
func handleScanEndpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload ScanPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Clean up the barcode string just in case there are accidental whitespaces
	cleanBarcode := strings.TrimSpace(payload.Barcode)

	// Validate barcode length (12 digit UPC + 5 digit addon = 17 digits)
	if len(cleanBarcode) < 17 {
		fmt.Printf("[Scanner] Warning: Barcode too short (%s). Waiting for full 17 digits.\n", cleanBarcode)
		// Return HTTP 400 with a specific error message for the frontend
		http.Error(w, "Barcode missing the 5-digit extension.", http.StatusBadRequest)
		return
	}

	fmt.Printf("[Scanner] Success: Received complete barcode: %s\n", cleanBarcode)

	// TODO: Trigger API lookup (Metron) and database storage here
	// Example: metadata, err := ParseBarcode(cleanBarcode)

	w.WriteHeader(http.StatusOK)
}

func main() {
	mux := http.NewServeMux()

	// Route for the mobile browser to load the UI
	mux.HandleFunc("/scanner", serveHtmlInterface)

	// Route for the JS fetch() call to post data
	mux.HandleFunc("/api/scan", handleScanEndpoint)

	mux.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "style.css")
})

	port := ":8080"
	fmt.Printf("Starting local server...\n")
	fmt.Printf("1. Ensure your PC and phone are on the same Wi-Fi network.\n")
	fmt.Printf("2. Open your phone's browser and navigate to: http://<YOUR_PC_IP>%s/scanner\n", port)
	fmt.Printf("   (Replace <YOUR_PC_IP> with your actual local IP address, e.g., 192.168.178.50)\n\n")
	fmt.Printf("3. Important: Accept the self-signed certificate warning in your mobile browser.\n")

	// Start the server
	// Start the TLS server using the generated certificates
	log.Fatal(http.ListenAndServeTLS(port, "192.168.0.110.pem", "192.168.0.110-key.pem", mux))
}
