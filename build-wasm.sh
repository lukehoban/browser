#!/bin/bash
set -e

echo "Building browser WASM module..."
cd "$(dirname "$0")"

GOOS=js GOARCH=wasm go build -o wasm/browser.wasm ./cmd/browser-wasm

echo "Build complete! WASM module saved to wasm/browser.wasm"
echo ""
echo "To run the demo:"
echo "  1. Start a local web server in the wasm directory:"
echo "     cd wasm && python3 -m http.server 8080"
echo "  2. Open http://localhost:8080 in your browser"
