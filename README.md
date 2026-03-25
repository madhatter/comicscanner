# comicscanner

A local HTTPS server for scanning comic book barcodes via a mobile browser.
Point your phone camera at a comic's barcode — the app decodes it into a
standardized 17 or 18 digit string and returns it to your application.

This package does one thing: scan barcodes and produce the result.
What you do with the result (API lookups, database writes, etc.) is up to you.

## Barcode format

Comics use two barcodes printed together:

| Type | Format | Length |
|------|--------|--------|
| Single-issue | UPC-A (12) + EAN-5 supplement (5) | 17 digits |
| Trade paperback | EAN-13/ISBN 978–979 (13) + EAN-5 supplement (5) | 18 digits |

The EAN-5 supplement encodes issue number (digits 1–3), cover variant (4),
and printing number (5).

## Requirements

- Go 1.21+
- [mkcert](https://github.com/FiloSottile/mkcert) for local TLS certificates
- [just](https://github.com/casey/just) (optional, for the provided task runner)

HTTPS is required because browsers only grant camera access on secure origins.

## Setup

```sh
just setup        # installs mkcert, creates local CA, generates certificate
just run          # starts the server
```

Or manually:

```sh
mkcert -install
mkcert -cert-file server.pem -key-file server-key.pem <YOUR_LOCAL_IP> localhost
go run ./cmd/
```

Open `https://<YOUR_LOCAL_IP>:8080/scanner` on your phone.

### Trusting the certificate on your phone

The scanner uses a locally-signed certificate. Your phone needs to trust the
root CA once.

**Android:**
1. Run `just ca-path` to find the `rootCA.pem` file
2. Copy it to your phone (download via USB, email, etc. — it must be saved
   locally before importing, not opened directly from cloud storage)
3. Settings → Security → Encryption & credentials → Install a certificate →
   CA certificate

**iOS:**
1. AirDrop or email the `rootCA.pem` to yourself
2. Settings → General → VPN & Device Management → install the profile
3. Settings → General → About → Certificate Trust Settings → enable the CA

## Usage as a library

```go
import "github.com/madhatter/comicscanner"

cfg := scanner.Config{
    Port:     ":8080",
    CertFile: "server.pem",
    KeyFile:  "server-key.pem",
}

// Stream: receive each scan as it comes in
results, err := scanner.StartStream(ctx, cfg)
for r := range results {
    fmt.Println(r.Barcode)      // full 17 or 18 digit string
    fmt.Println(r.SKU())        // main barcode without supplement
    fmt.Println(r.Supplement()) // 5-digit EAN-5 addon
}

// One-shot: block until a single barcode is scanned
result, err := scanner.Start(ctx, cfg)
```

## API reference

### `Config`

| Field | Description |
|-------|-------------|
| `Port` | Listen address, e.g. `":8080"` |
| `CertFile` | Path to TLS certificate |
| `KeyFile` | Path to TLS private key |

### `Result`

| Method/Field | Returns |
|--------------|---------|
| `result.Barcode` | Full scanned code (17 or 18 digits) |
| `result.Type()` | `BarcodeTypeComic`, `BarcodeTypeTradePaperback`, or `BarcodeTypeUnknown` |
| `result.SKU()` | Main barcode without supplement (12 or 13 digits) |
| `result.Supplement()` | 5-digit EAN-5 addon |
| `result.ScannedAt` | `time.Time` of when the server received the scan |