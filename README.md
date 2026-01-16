# Browser - A Simple Web Browser in Go

[![CI](https://github.com/lukehoban/browser/actions/workflows/ci.yml/badge.svg)](https://github.com/lukehoban/browser/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lukehoban/browser)](https://goreportcard.com/report/github.com/lukehoban/browser)
[![WPT Tests](https://img.shields.io/badge/WPT%20Tests-92.3%25-brightgreen)](./TESTING.md)

A simple web browser implementation in Go that parses HTML and CSS, computes styles, calculates layout, and renders to PNG images. Built for educational purposes with a focus on CSS 2.1 compliance and W3C specifications.

> **[Try the Live Demo →](https://lukehoban.github.io/browser/)** — Runs entirely in your browser via WebAssembly

## Features

### Core Rendering
- **HTML parsing** with DOM tree construction (HTML5 §12.2)
- **CSS 2.1 parsing** and style computation with cascade/specificity
- **Visual formatting model** — box model, block layout, table layout
- **High-quality text rendering** with [Go fonts](https://blog.golang.org/go-fonts) (proportional sans-serif)
- **Font styling** — bold, italic, underline, variable sizes
- **Image rendering** — PNG, JPEG, GIF, and SVG support
- **Background and border rendering** with full color support

### Network & Data
- **HTTP/HTTPS support** — load pages from URLs
- **External stylesheets** — fetch and apply CSS from `<link>` tags
- **Remote images** — load images from network URLs
- **Data URLs** — RFC 2397 support (base64 and URL-encoded)

### Platforms
- **CLI** — render HTML files or URLs to PNG
- **WebAssembly** — run entirely in a web browser

## Project Structure

```
browser/
├── cmd/
│   ├── browser/      # Main CLI browser application
│   └── browser-wasm/ # WebAssembly entry point
├── html/            # HTML tokenization and parsing
├── css/             # CSS parsing
├── dom/             # DOM tree structure
├── style/           # Style computation and cascade
├── layout/          # Layout engine (visual formatting model)
├── render/          # Rendering engine
├── wasm/            # WebAssembly demo page
└── test/            # Test files and fixtures
```

## Specifications

This browser implementation follows these W3C specifications:

- **HTML5**: Tokenization and parsing ([HTML5 §12](https://html.spec.whatwg.org/multipage/parsing.html))
- **CSS 2.1**: Syntax, selectors, cascade, box model, and visual formatting
  - [CSS 2.1 §4 Syntax](https://www.w3.org/TR/CSS21/syndata.html)
  - [CSS 2.1 §5 Selectors](https://www.w3.org/TR/CSS21/selector.html)
  - [CSS 2.1 §6 Cascade](https://www.w3.org/TR/CSS21/cascade.html)
  - [CSS 2.1 §8 Box Model](https://www.w3.org/TR/CSS21/box.html)
  - [CSS 2.1 §9 Visual Formatting Model](https://www.w3.org/TR/CSS21/visuren.html)
- **RFC 2397**: The "data" URL scheme for inline resources

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/lukehoban/browser.git
cd browser

# Build
go build ./cmd/browser
```

### Usage

```bash
# Render local HTML file to PNG
./browser -output output.png test/styled.html

# Load and render a web page from URL
./browser -output hn.png https://news.ycombinator.com/

# View layout tree without rendering (text output)
./browser test/styled.html

# Custom viewport size
./browser -output output.png -width 1024 -height 768 test/hackernews.html
```

## Screenshots

### Hacker News

Rendering of the live Hacker News homepage (1024×768 viewport):

![Hacker News Rendering](https://github.com/user-attachments/assets/1182c930-c17c-4f65-8fcc-4ebed2aa3aee)

### Font Rendering

The browser uses the [Go fonts](https://blog.golang.org/go-fonts) — high-quality, proportional, sans-serif fonts with support for bold, italic, and various sizes:

![Font Comparison](https://github.com/user-attachments/assets/6c138aa6-7b3d-48e7-8d9c-c4469baad477)

### Styled HTML

Example of styled HTML with backgrounds, borders, and text formatting:

![Test Case Rendering](https://github.com/user-attachments/assets/a6f4a864-cd34-4919-a273-75e9d1d682e5)

## Testing

```bash
# Run all unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run WPT reference tests
go test ./reftest/... -v
```

See [TESTING.md](TESTING.md) for detailed test results and WPT integration.

## WebAssembly

The browser compiles to WebAssembly and runs entirely in a web browser.

**[Live Demo →](https://lukehoban.github.io/browser/)** (automatically deployed via GitHub Actions)

To build and run locally:

```bash
# Build WASM binary
GOOS=js GOARCH=wasm go build -o wasm/browser.wasm ./cmd/browser-wasm

# Serve locally
cd wasm && python3 -m http.server 8080
# Open http://localhost:8080
```

See [wasm/README.md](wasm/README.md) for more details.

## Documentation

- **[MILESTONES.md](MILESTONES.md)** - Implementation milestones and progress tracking
- **[IMPLEMENTATION.md](IMPLEMENTATION.md)** - Detailed implementation summary and architecture
- **[TESTING.md](TESTING.md)** - Testing strategy and public test suite integration

## Current Status

| Component | Status | Details |
|-----------|--------|---------|
| HTML Parsing | ✅ Complete | Tokenization, DOM tree, character entities |
| CSS Parsing | ✅ Complete | Selectors, declarations, cascade |
| Style Computation | ✅ Complete | Specificity, inheritance, inline styles |
| Layout Engine | ✅ Complete | Block, inline, and table layout |
| Rendering | ✅ Complete | Text, colors, borders, images |
| Network Support | ✅ Complete | HTTP/HTTPS, external CSS, remote images |
| WebAssembly | ✅ Complete | Interactive demo deployed |
| **WPT Tests** | **92.3%** | 36/39 tests passing |

See [MILESTONES.md](MILESTONES.md) for detailed progress and known limitations.

## Contributing

Contributions are welcome! Here's how to get started:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes with tests
4. Run the test suite (`go test ./...`)
5. Submit a pull request

Please see [IMPLEMENTATION.md](IMPLEMENTATION.md) for architecture details and coding guidelines.

## License

MIT
