# Browser - A Simple Web Browser in Go

[![CI](https://github.com/lukehoban/browser/actions/workflows/ci.yml/badge.svg)](https://github.com/lukehoban/browser/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lukehoban/browser)](https://goreportcard.com/report/github.com/lukehoban/browser)

A simple web browser implementation in Go, focusing on static HTML and CSS 2.1 compliance. This project aims to stay close to W3C specifications and provide a clean, well-organized codebase for educational purposes.

## Features

- HTML parsing with DOM tree construction
- CSS 2.1 parsing and style computation
- Visual formatting model (box model, block layout)
- Text rendering with color support
- Image rendering (PNG, JPEG, GIF support)
- Background and border rendering
- PNG image output
- **Network support**: Load pages via HTTP/HTTPS
- **External CSS**: Fetch and apply stylesheets from `<link>` tags
- **Network images**: Load images from remote URLs
- **Web interface**: Test from any device with the built-in web server

## Project Structure

```
browser/
├── cmd/browser/      # Main browser application (CLI)
├── cmd/webserver/    # Web server for testing from any device
├── html/            # HTML tokenization and parsing
├── css/             # CSS parsing
├── dom/             # DOM tree structure
├── style/           # Style computation and cascade
├── layout/          # Layout engine (visual formatting model)
├── render/          # Rendering engine
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

## Quick Start

### Building

```bash
go build ./cmd/browser
```

### Running

#### Command-Line Renderer

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

#### Web Server (Test from Any Device)

```bash
# Build and start the web server
go build ./cmd/webserver
./webserver

# The server will display URLs like:
#   Local:   http://localhost:8080
#   Network: http://<YOUR_MACHINE_IP>:8080

# Options:
./webserver -port 3000              # Custom port
./webserver -host 0.0.0.0 -port 8080  # Bind to all interfaces
```

**To test from your phone:**
1. Make sure your phone is on the same WiFi network as your computer
2. Open your phone's browser and navigate to the Network URL shown when you start the server
3. Enter HTML in the text area and click "Render" to see the output
4. Try the quick example buttons for pre-built demos

### Testing

```bash
go test ./...
```

## Documentation

- **[MILESTONES.md](MILESTONES.md)** - Implementation milestones and progress tracking
- **[IMPLEMENTATION.md](IMPLEMENTATION.md)** - Detailed implementation summary and architecture
- **[TESTING.md](TESTING.md)** - Testing strategy and public test suite integration

## Current Status

✅ Milestones 1-7 Complete: Foundation, HTML Parsing, CSS Parsing, Style Computation, Layout Engine, Rendering, Image Rendering  
✅ Milestone 9 Complete: Network Support (HTTP/HTTPS, external CSS, remote images)  
🔄 Milestone 8 In Progress: Testing & Validation (81.8% WPT pass rate)

See [MILESTONES.md](MILESTONES.md) for detailed progress and known limitations.

## License

MIT