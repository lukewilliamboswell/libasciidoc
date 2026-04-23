#!/usr/bin/env bash
set -euo pipefail

PORT="${1:-8080}"
SITE_DIR="_site"

echo "Building ascii2doc..."
go build -o ascii2doc ./cmd/ascii2doc/

echo "Generating static site..."
rm -rf "$SITE_DIR"
./ascii2doc --static-site -o "$SITE_DIR" www/

echo "Serving at http://localhost:$PORT"
echo "Press Ctrl+C to stop."
python3 -m http.server "$PORT" -d "$SITE_DIR"
