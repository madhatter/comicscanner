# Comic Scanner - Development Tasks
# Requires: just, mkcert, go

# Fixed certificate names so main.go can always reference the same files
CERT_FILE := "server.pem"
KEY_FILE  := "server-key.pem"

# Default: show available tasks
default:
    @just --list

# ── Setup ─────────────────────────────────────────────────────────────────────

# Install mkcert and set up local CA (run once per machine)
setup-ca:
    #!/usr/bin/env bash
    set -euo pipefail
    if ! command -v mkcert &> /dev/null; then
        echo "Installing mkcert..."
        if command -v brew &> /dev/null; then
            brew install mkcert
        elif command -v apt &> /dev/null; then
            sudo apt install -y mkcert
        elif command -v pacman &> /dev/null; then
            sudo pacman -S --noconfirm mkcert
        else
            echo "ERROR: Could not find brew or apt. Please install mkcert manually:"
            echo "  https://github.com/FiloSottile/mkcert#installation"
            exit 1
        fi
    else
        echo "mkcert already installed: $(mkcert --version)"
    fi
    mkcert -install
    echo ""
    echo "Local CA installed. Root certificate location:"
    just ca-path

# Generate certificate for the current local IP + localhost
gen-cert:
    #!/usr/bin/env bash
    set -euo pipefail
    # Detect local IP (works on macOS and Linux)
    LOCAL_IP=$(ip route get 1 2>/dev/null | awk '{print $7; exit}' || \
               ipconfig getifaddr en0 2>/dev/null || \
               ipconfig getifaddr en1 2>/dev/null || \
               echo "127.0.0.1")
    echo "Detected local IP: $LOCAL_IP"
    mkcert \
        -cert-file {{CERT_FILE}} \
        -key-file  {{KEY_FILE}} \
        "$LOCAL_IP" localhost 127.0.0.1
    echo ""
    echo "Certificate created:"
    echo "  Cert: {{CERT_FILE}}"
    echo "  Key:  {{KEY_FILE}}"

# Show where the root CA file is (this is what needs to be imported on mobile)
ca-path:
    #!/usr/bin/env bash
    CA_ROOT=$(mkcert -CAROOT)
    echo ""
    echo "┌─────────────────────────────────────────────────────────┐"
    echo "│  Root CA for mobile import                              │"
    echo "├─────────────────────────────────────────────────────────┤"
    echo "│  File: $CA_ROOT/rootCA.pem"
    echo "│                                                         │"
    echo "│  Send this file to your phone and install it:          │"
    echo "│  • iOS:     Settings → General → VPN & Device Mgmt     │"
    echo "│  • Android: Settings → Security → Install Certificate  │"
    echo "└─────────────────────────────────────────────────────────┘"
    echo ""

# Full first-time setup: install mkcert + CA + generate cert
setup: setup-ca gen-cert ca-path
    @echo "Setup complete. Run 'just run' to start the server."

# ── Development ───────────────────────────────────────────────────────────────

# Build the server binary
build:
    go build -o comic-scanner .

# Run the server (generates cert first if missing)
run:
    #!/usr/bin/env bash
    if [ ! -f {{CERT_FILE}} ] || [ ! -f {{KEY_FILE}} ]; then
        echo "No certificate found — running gen-cert first..."
        just gen-cert
    fi
    go run .

# ── Cleanup ───────────────────────────────────────────────────────────────────

# Remove generated certificates
clean-certs:
    rm -f {{CERT_FILE}} {{KEY_FILE}}
    @echo "Certificates removed."

# Remove build artifact
clean-build:
    rm -f comic-scanner

# Remove everything generated
clean: clean-certs clean-build
