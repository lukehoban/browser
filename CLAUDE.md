# CLAUDE.md — Browser Project Guide

This file provides context and guidance for AI assistants working in this repository.

## Project Overview

A web browser implementation in Go, focusing on static HTML and CSS 2.1 compliance. The browser parses HTML and CSS, computes styles, calculates layout, and renders to PNG images. It can also be compiled to WebAssembly to run in a real web browser.

- **Module**: `github.com/lukehoban/browser`
- **Go version**: 1.24 (go.mod), CI targets Go 1.23+
- **Dependencies**: `golang.org/x/image`, `golang.org/x/text`
- **License**: MIT

## Repository Structure

```
browser/
├── cmd/
│   ├── browser/         # Main CLI application (main.go, main_test.go)
│   ├── browser-wasm/    # WebAssembly entry point (js/wasm build tag)
│   └── wptrunner/       # WPT reference test runner CLI
├── html/                # HTML tokenization and parsing
│   ├── tokenizer.go     # State-machine HTML5 tokenizer
│   ├── tokenizer_test.go
│   ├── parser.go        # DOM tree construction
│   └── parser_test.go
├── css/                 # CSS tokenization and parsing
│   ├── tokenizer.go     # CSS tokenizer (identifiers, strings, numbers, etc.)
│   ├── tokenizer_test.go
│   ├── parser.go        # Selector and declaration parsing
│   ├── parser_test.go
│   ├── values.go        # CSS value parsing utilities
│   └── values_test.go
├── dom/                 # DOM tree data structures
│   ├── node.go          # Node types: ElementNode, TextNode, DocumentNode
│   ├── node_test.go
│   ├── url.go           # URL resolution (HTML5 §2.5)
│   ├── url_test.go
│   ├── loader.go        # External stylesheet fetching
│   └── loader_test.go
├── style/               # Style computation and CSS cascade
│   ├── style.go         # Selector matching, specificity, cascade
│   ├── style_test.go
│   └── useragent.go     # User-agent stylesheet defaults
├── layout/              # Layout engine (CSS box model)
│   ├── layout.go        # Block layout, dimensions, box model
│   ├── layout_test.go
│   └── image_test.go
├── render/              # Rendering engine
│   ├── render.go        # Canvas drawing, PNG output, image rendering
│   └── render_test.go
├── font/                # Font handling
│   ├── font.go          # Go fonts (proportional sans-serif)
│   └── rasterizer.go    # Font rasterization
├── svg/                 # SVG rendering
│   ├── svg.go
│   └── svg_test.go
├── log/                 # Internal logging package (no external deps)
│   ├── logger.go        # Leveled logging (Debug, Info, Warn, Error)
│   └── logger_test.go
├── reftest/             # WPT reference test harness
│   ├── reftest.go
│   ├── reftest_test.go
│   └── wpt_test.go
├── wasm/                # WebAssembly demo page files
│   ├── index.html
│   ├── wasm_exec.js
│   └── README.md
└── test/                # Test HTML files and WPT fixtures
    ├── *.html           # Integration test HTML files
    └── wpt/css/         # WPT CSS reference tests (13 categories)
```

## Rendering Pipeline

The browser follows a classic pipeline:

```
HTML Input → Tokenization → DOM Tree → Style Computation → Layout → Rendering → PNG Output
                                ↓
                          CSS Parsing
                    (<style> tags + <link> stylesheets)
```

1. `html.Parse()` — tokenizes and builds a DOM tree
2. `dom.ResolveURLs()` — resolves relative URLs against the document base
3. `dom.FetchExternalStylesheets()` — fetches `<link rel="stylesheet">` resources
4. `css.Parse()` — tokenizes and parses CSS into a stylesheet
5. `style.StyleTree()` — matches selectors, computes specificity, applies cascade
6. `style.ResolveCSSURLs()` — resolves CSS resource URLs (e.g. background-image)
7. `layout.LayoutTree()` — calculates box model, widths, heights, positions
8. `render.Render()` — draws boxes, text, images to a pixel canvas
9. `canvas.SavePNG()` — writes PNG output

## Build Commands

```bash
# Build the CLI browser binary
go build ./cmd/browser

# Build all packages
go build -v ./...

# Build WebAssembly target
GOOS=js GOARCH=wasm go build -o wasm/browser.wasm ./cmd/browser-wasm

# Build the WPT test runner
go build -o wptrunner ./cmd/wptrunner
```

## Running the Browser

```bash
# Render a local HTML file to PNG
./browser -output output.png test/styled.html

# Render a web page from URL
./browser -output hn.png https://news.ycombinator.com/

# Show layout tree (text output, no PNG)
./browser test/styled.html

# Show rendered (styled) node tree
./browser -show-render test/styled.html

# Show layout tree explicitly
./browser -show-layout test/styled.html

# Custom viewport size
./browser -output output.png -width 1024 -height 768 test/hackernews.html

# Enable debug logging
./browser -log-level debug -output output.png test/styled.html
```

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-output` | (none) | Output PNG file path; omit to print layout tree |
| `-width` | `800` | Viewport width in pixels |
| `-height` | `600` | Viewport height in pixels |
| `-log-level` | `warn` | Log level: `debug`, `info`, `warn`, `error` |
| `-verbose` | `false` | Shorthand for `-log-level=info` |
| `-show-layout` | `false` | Print layout tree and exit |
| `-show-render` | `false` | Print styled node tree and exit |

## Testing

### Run All Tests

```bash
go test ./...
```

### Common Test Invocations

```bash
# Run all tests with verbose output and race detector
go test -v -race ./...

# Run with coverage report
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out   # open in browser

# Run tests for a specific package
go test ./html
go test ./css
go test ./style
go test ./layout
go test ./render

# Run WPT reference tests
go test ./reftest/... -run TestWPTCSSReftests -v

# Run WPT test runner directly
./wptrunner -v test/wpt/css/
```

### Test Structure

- **Unit tests**: Each package has `*_test.go` files co-located with source
- **Integration tests**: HTML files in `test/` (simple.html, styled.html, hackernews.html, etc.)
- **WPT reference tests**: `test/wpt/css/` — 13 categories, 39 tests, 94.9% pass rate

### Adding Tests

When adding new features:
1. Add unit tests in the relevant `*_test.go` file
2. Add integration test HTML files to `test/` if testing rendering
3. Add WPT-style reftests to `test/wpt/css/<category>/` if testing CSS compliance
   - Create `test.html` with `<link rel="match" href="reference.html">`
   - Create `reference.html` with the expected layout

### Current WPT Pass Rate

| Category | Pass Rate |
|----------|-----------|
| css-borders | 100% |
| css-box | 100% |
| css-cascade | 100% |
| css-color | 100% |
| css-display | 100% |
| css-float | 100% (gracefully ignored) |
| css-fonts | 100% |
| css-inheritance | 100% |
| css-position | 100% (gracefully ignored) |
| css-selectors | 100% |
| css-text-decor | 100% |
| css-selectors-advanced | 60% (sibling combinators not implemented) |
| **Total** | **94.9% (37/39)** |

Known failing: adjacent sibling combinator (`+`) and general sibling combinator (`~`).

## CI/CD

GitHub Actions runs on pull requests to `main`:

| Job | Description | Blocks merge? |
|-----|-------------|---------------|
| `test` | `go test -v -race ./...` + Codecov upload | Yes |
| `lint` | `golangci-lint` with 5-minute timeout | Yes |
| `build` | `go build ./...` | Yes |
| `wpt` | WPT reftest suite; generates report artifact | No (`continue-on-error: true`) |

GitHub Pages automatically deploys the WebAssembly demo from the `wasm/` directory.

## Code Conventions

### Package Documentation

Every package begins with a doc comment that:
- States what the package does
- Lists all W3C/HTML5 spec references with links
- Lists implemented features
- Lists not-yet-implemented features (with log warnings at runtime)

Example (`style/style.go`):
```go
// Package style handles style computation and the CSS cascade.
// ...
// Spec references:
// - CSS 2.1 §6 Assigning property values...: https://www.w3.org/TR/CSS21/cascade.html
//
// Not yet implemented (noted with log warnings where encountered):
// - !important declarations (CSS 2.1 §6.4.2)
```

### Specification References

All implementations cite the relevant W3C spec section in comments, for example:
```go
// CSS 2.1 §6.4.3: Calculating a selector's specificity
// Specificity = (a, b, c) where:
//   a = number of ID selectors
//   b = number of class/attribute/pseudo-class selectors
//   c = number of type/pseudo-element selectors
```

### Unimplemented Features

When the parser or renderer encounters a feature that is not yet implemented:
- Use the `log` package to emit a `Warn` level message
- Do NOT panic or return an error for graceful degradation
- Document the gap in the package-level comment

```go
log.Warnf("CSS: child combinator '>' not yet implemented, treating as descendant")
```

### Logging

Use the internal `log` package (not `fmt` or `log` from stdlib for debug output):

```go
import "github.com/lukehoban/browser/log"

log.Debugf("Tokenizer: entering state %s", stateName)
log.Infof("Fetching stylesheet: %s", url)
log.Warnf("Unimplemented CSS property: %s", propName)
log.Errorf("Failed to decode image: %v", err)
```

Default log level is `WarnLevel`. Use `-log-level debug` or `-verbose` at runtime to see more output.

### Error Handling

- The browser is designed for graceful degradation — unsupported features are skipped, not fatal
- Fatal errors (missing input file, unwritable output) call `os.Exit(1)` with a message to stderr
- Internal packages return errors via Go's standard `error` type where appropriate
- Network failures in `dom.FetchExternalStylesheets()` are logged as warnings and skipped

### Go Style

- Standard `gofmt` formatting (enforced by golangci-lint in CI)
- Package-level variables are initialized with named functions for clarity
- Prefer explicit constants over magic numbers (see `maxDisplayedStyles`, `importantStyles`)
- Pre-allocate slices with known capacity using `make([]T, 0, n)`

## Key Data Structures

```go
// dom/node.go — DOM tree node
type Node struct {
    Type       NodeType              // ElementNode, TextNode, DocumentNode
    Data       string                // Tag name or text content
    Attributes map[string]string     // Element attributes
    Children   []*Node               // Child nodes
    Parent     *Node                 // Parent node
}

// style/style.go — Node with computed styles
type StyledNode struct {
    Node     *dom.Node
    Styles   map[string]string      // property name → value (e.g. "color" → "red")
    Children []*StyledNode
}

// layout/layout.go — Box in the layout tree
type LayoutBox struct {
    BoxType     BoxType              // Block, Inline, Anonymous, Table, TableRow, TableCell
    Dimensions  Dimensions           // Position, size, padding, border, margin
    StyledNode  *StyledNode          // Reference to styled DOM node
    Children    []*LayoutBox
}

type Dimensions struct {
    Content Rect      // x, y, width, height of content box
    Padding EdgeSize  // top, right, bottom, left
    Border  EdgeSize
    Margin  EdgeSize
}
```

## Known Limitations

The following features are not yet implemented (see `MILESTONES.md` for full tracking):

**HTML**:
- No script/style CDATA sections
- No namespace support (SVG parsed separately)
- Simplified error recovery

**CSS Selectors**:
- Adjacent sibling combinator (`+`) — WPT tests failing
- General sibling combinator (`~`) — WPT tests failing

**CSS Properties/Features**:
- `!important` declarations (gracefully ignored)
- Floats (gracefully ignored, uses normal flow)
- Absolute/relative/fixed positioning (gracefully ignored)
- Full inline formatting context (inline elements treated as blocks)
- Flexbox, Grid (not implemented)
- CSS animations/transitions

**Rendering**:
- No subpixel rendering
- No text wrapping within inline boxes
- Simple nearest-neighbor image scaling

## WebAssembly Demo

The browser compiles to WebAssembly for running in a real browser:

```bash
# Build WASM binary
GOOS=js GOARCH=wasm go build -o wasm/browser.wasm ./cmd/browser-wasm

# Serve locally
cd wasm && python3 -m http.server 8080
```

Live demo: https://lukehoban.github.io/browser/

The WASM entry point (`cmd/browser-wasm/main.go`) exposes `renderHTML` to JavaScript via `syscall/js`. It uses `//go:build js && wasm` build tags and is excluded from normal `go build ./...`.

## Documentation Files

| File | Purpose |
|------|---------|
| `README.md` | Project overview, quick start, screenshots |
| `IMPLEMENTATION.md` | Architecture details, design decisions, spec compliance |
| `MILESTONES.md` | Feature tracking, task status, progress |
| `TESTING.md` | Testing strategy, WPT results, how to add tests |
| `CLAUDE.md` | This file — AI assistant guide |
