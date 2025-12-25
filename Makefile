.PHONY: build build-wasm serve-wasm test clean

# Build the CLI browser
build:
	go build -o browser ./cmd/browser

# Build the WebAssembly version
build-wasm:
	GOOS=js GOARCH=wasm go build -o wasm/browser.wasm ./cmd/browser-wasm

# Build and serve the WASM demo
serve-wasm: build-wasm
	@echo "Starting web server at http://localhost:8080"
	@echo "Press Ctrl+C to stop"
	cd wasm && python3 -m http.server 8080

# Run tests
test:
	go test ./css ./dom ./html ./layout ./render ./style ./cmd/browser

# Run all tests including WASM compilation check
test-all: test
	@echo "Testing WASM compilation..."
	GOOS=js GOARCH=wasm go build -o /tmp/browser-test.wasm ./cmd/browser-wasm
	@echo "All tests passed!"

# Clean build artifacts
clean:
	rm -f browser wasm/browser.wasm
