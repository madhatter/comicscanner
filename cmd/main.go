package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/skip2/go-qrcode"

	scanner "github.com/madhatter/comicscanner"
)

const logo = `
┌──────────────────────────────────────────────────┐
│                                                  │
│   █ ▌█▌▌▌ █▌ ▌█  COMIC SCANNER  █▌ ▌█▌ █▌█ ▌█▌█  │
│           17-Digit Barcode · EAN / UPC           │
│                                                  │
└──────────────────────────────────────────────────┘
`

func main() {
	cfg := scanner.Config{
		Port:     ":8080",
		CertFile: "server.pem",
		KeyFile:  "server-key.pem",
	}

	fmt.Print(logo)
	fmt.Println("Press Ctrl+C to stop.")
	fmt.Println()

	// Print QR code and URL
	ip := localIP()
	url := fmt.Sprintf("https://%s%s/scanner", ip, cfg.Port)

	qr, err := qrcode.New(url, qrcode.Medium)
	if err == nil {
		fmt.Println(qr.ToSmallString(false)) // ASCII im Terminal
	}
	fmt.Printf("Scan the QR code or open: %s\n", url)
	fmt.Println("Run 'just ca-path' to find the root CA file for mobile import.")
	fmt.Println()

	// Cancel context on Ctrl+C or SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	results, err := scanner.StartStream(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to start scanner: %v", err)
	}

	for result := range results {
		fmt.Printf("─────────────────────────────────\n")
		fmt.Printf("Barcode:    %s\n", result.Barcode)
		fmt.Printf("Type:       %s\n", typeName(result.Type()))
		fmt.Printf("SKU:        %s\n", result.SKU())
		fmt.Printf("Supplement: %s\n", result.Supplement())
		fmt.Printf("Scanned at: %s\n", result.ScannedAt.Format("15:04:05"))
		fmt.Printf("─────────────────────────────────\n")
	}

	fmt.Println("Scanner stopped.")
}

// localIP returns the preferred outbound local IP address.
func localIP() string {
	// Dial a remote address (doesn't actually send anything) to determine
	// which local interface would be used — that's our LAN IP.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "localhost"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

// typeName converts a BarcodeType to a human-readable string.
func typeName(t scanner.BarcodeType) string {
	switch t {
	case scanner.BarcodeTypeComic:
		return "Single-issue comic"
	case scanner.BarcodeTypeTradePaperback:
		return "Trade paperback"
	default:
		return "Unknown"
	}
}
