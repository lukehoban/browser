# Copilot Instructions for Browser Project

## Overview

This is a web browser implementation in Go that renders HTML/CSS to PNG images. The rendering pipeline is:

**HTML → DOM tree → CSS parsing → Style computation → Layout → PNG rendering**

### Package structure

| Package | Responsibility |
|---------|---------------|
| `html/` | HTML tokenizer and parser → DOM tree |
| `css/` | CSS tokenizer, parser, and value resolution |
| `dom/` | DOM node types, URL resolution, external stylesheet fetching |
| `style/` | CSS cascade, specificity, selector matching (uses RuleIndex for O(1) lookup), inheritance |
| `layout/` | CSS 2.1 box model — block formatting, inline flow, table layout |
| `render/` | Rasterization to PNG — text, images, backgrounds, borders |
| `font/` | Font loading and caching (Go fonts via golang.org/x/image) |
| `svg/` | Basic SVG rendering |
| `log/` | Leveled logger |
| `cmd/browser/` | CLI entry point |
| `cmd/browser-wasm/` | WebAssembly entry point for browser demo |
| `reftest/` | WPT-style reference test harness |

## Build, Test, and Run

```bash
# Build
go build -v ./cmd/browser

# Run on a URL or local file
./browser -output out.png https://news.ycombinator.com/
./browser -output out.png test/render_test.html

# Useful flags: -width, -height, -log-level (debug|info|warn|error), -verbose, -show-layout

# Run all unit tests
go test ./...

# Run with race detector and coverage (matches CI)
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Run a single package's tests
go test ./css/
go test ./layout/

# Run WPT reference tests (visual regression tests)
go test -v ./reftest

# Build WASM demo
GOOS=js GOARCH=wasm go build -o wasm/browser.wasm ./cmd/browser-wasm
```

## Reference Tests (reftests)

The `reftest/` directory contains a WPT-style visual regression test harness. Each test has an HTML file and a `<link rel="match">` (or `<link rel="mismatch">`) tag pointing to a reference HTML file. The harness renders both to PNG and compares pixel output.

- Run with: `go test -v ./reftest`
- Test HTML files live in `test/` subdirectories (e.g., `test/wpt/css/`)
- To add a new test: create an HTML file with `<link rel="match" href="ref-file.html">` and place it in the appropriate `test/` subdirectory
- Current coverage: ~39 WPT CSS tests, ~95% pass rate

## Milestones Document

When implementing new features or making significant changes, consider updating MILESTONES.md to reflect the current state — mark completed tasks, update validation status, and note new limitations discovered.

## Screenshot Requirements

This project renders HTML/CSS to PNG images. Visual changes are difficult to review from code alone.

**To regenerate the Hacker News screenshot:**
```bash
go build ./cmd/browser && ./browser -output hackernews_screenshot.png https://news.ycombinator.com/
```

**Before submitting changes that affect rendering, layout, or styling:**
```bash
go build ./cmd/browser && ./browser -output screenshot.png test/render_test.html
```

Attach the screenshot to your PR or commit. For behavioral changes, capture before/after screenshots for comparison.
