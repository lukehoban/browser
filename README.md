# Browser - A Simple Web Browser in Go :)

[![CI](https://github.com/lukehoban/browser/actions/workflows/ci.yml/badge.svg)](https://github.com/lukehoban/browser/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lukehoban/browser)](https://goreportcard.com/report/github.com/lukehoban/browser)

A simple web browser implementation in Go, focusing on static HTML and CSS 2.1 compliance. This project aims to stay close to W3C specifications and provide a clean, well-organized codebase for educational purposes.

## Features

- HTML parsing with DOM tree construction
- CSS 2.1 parsing and style computation
- Visual formatting model (box model, block layout)
- **High-quality text rendering** with Go fonts (proportional sans-serif)
- Font styling support (bold, italic, underline, size)
- Image rendering (PNG, JPEG, GIF, SVG support)
- **Data URLs**: Support for RFC 2397 data URLs (base64 and URL-encoded)
- Background and border rendering
- PNG image output
- **Network support**: Load pages via HTTP/HTTPS
- **External CSS**: Fetch and apply stylesheets from `<link>` tags
- **Network images**: Load images from remote URLs
- **WebAssembly**: Run the browser entirely in a web client

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

## Architecture

### Rendering Pipeline

The browser processes a web page through a multi-stage pipeline, transforming HTML and CSS source into a rendered PNG image:

```mermaid
flowchart TD
    A["📄 HTML Source\n(file, HTTP/HTTPS URL, or data URL)"] --> B["🔤 HTML Tokenizer\n(html/tokenizer.go)\nState machine producing token stream"]
    B --> C["🌳 HTML Parser\n(html/parser.go)\nStack-based tree construction"]
    C --> D["📦 DOM Tree\n(dom/node.go)\nElement, Text, and Document nodes"]
    D --> E["🔗 URL Resolution\n(dom/url.go)\nResolve relative paths to absolute URLs"]
    E --> F["📦 DOM Tree\nwith resolved URLs"]

    F --> G["🎨 CSS Extraction\nInline <style> tags +\nexternal <link> stylesheets"]
    G --> H["🎨 CSS Tokenizer\n(css/tokenizer.go)\nIdent, String, Hash, Number tokens"]
    H --> I["📋 CSS Parser\n(css/parser.go)\nSelectors + Declarations → Stylesheet"]

    F --> J["🧮 Style Computation\n(style/style.go)"]
    I --> J
    UA["📜 User-Agent Defaults\n(style/useragent.go)"] --> J

    J --> K["✨ Styled Tree\nDOM nodes with computed CSS properties\n(cascade, specificity, inheritance)"]

    K --> L["📐 Layout Engine\n(layout/layout.go)\nBox model, block/inline/table layout"]
    L --> M["📐 Layout Tree\nBoxes with computed dimensions\n(content, padding, border, margin)"]

    M --> N["🖌️ Render Engine\n(render/render.go)\nCanvas painting: backgrounds, borders,\ntext, images, SVG"]
    N --> O["🖼️ PNG Output\nFinal rendered image"]

    N -. "font metrics" .-> P["🔡 Font Engine\n(font/font.go)\nGo TrueType fonts"]
    N -. "SVG raster" .-> Q["📐 SVG Engine\n(svg/svg.go)\nParse & rasterize SVG"]
```

### Module Dependencies

```mermaid
graph TD
    CLI["cmd/browser\nCLI entry point"] --> HTML
    WASM["cmd/browser-wasm\nWebAssembly entry point"] --> HTML

    HTML["html\nTokenizer & Parser"] --> DOM
    DOM["dom\nNode tree, URL resolution,\nresource loading"]

    CSS["css\nTokenizer & Parser"]

    DOM --> STYLE
    CSS --> STYLE
    STYLE["style\nCascade, specificity,\ninheritance, user-agent defaults"]

    STYLE --> LAYOUT
    LAYOUT["layout\nBox model, block/inline/table\ndimension calculation"]

    LAYOUT --> RENDER
    RENDER["render\nCanvas, text, images,\nbackgrounds, borders"]

    RENDER --> FONT["font\nTrueType loading &\ntext measurement"]
    RENDER --> SVG["svg\nSVG parsing &\nrasterization"]
    RENDER --> LOG["log\nStructured logging"]
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

### Building

```bash
go build ./cmd/browser
```

### Running

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

### Font Rendering

The browser uses the [Go fonts](https://blog.golang.org/go-fonts) - high-quality, proportional, sans-serif fonts designed for the Go project. These fonts are embedded in the binary and provide excellent readability with support for bold, italic, and various sizes.

![Font Comparison](./font_comparison_screenshot.png)

### Test Case Rendering

Example of styled HTML with borders, colors, and text formatting:

![Test Case Rendering](./test_case_screenshot.png)

### Hacker News

Latest Hacker News render (1024x768):

![Hacker News Rendering](./hackernews_screenshot.png)

### Testing

```bash
go test ./...
```

## WebAssembly

The browser can be compiled to WebAssembly and run entirely in a web browser. A live demo is available at **https://lukehoban.github.io/browser/** and is automatically deployed via GitHub Actions.

To build locally:
```bash
GOOS=js GOARCH=wasm go build -o wasm/browser.wasm ./cmd/browser-wasm
cd wasm && python3 -m http.server 8080
```

See [wasm/README.md](wasm/README.md) for more details.

## Documentation

- **[MILESTONES.md](MILESTONES.md)** - Implementation milestones and progress tracking
- **[IMPLEMENTATION.md](IMPLEMENTATION.md)** - Detailed implementation summary and architecture
- **[TESTING.md](TESTING.md)** - Testing strategy and public test suite integration

## Current Status

✅ Milestones 1-7 Complete: Foundation, HTML Parsing, CSS Parsing, Style Computation, Layout Engine, Rendering, Image Rendering  
✅ Milestone 9 Complete: Network Support (HTTP/HTTPS, external CSS, remote images)  
✅ Milestone 9.5 Complete: Data URL Support (RFC 2397, base64 & URL-encoded inline resources)  
🔄 Milestone 8 In Progress: Testing & Validation (81.8% WPT pass rate)

See [MILESTONES.md](MILESTONES.md) for detailed progress and known limitations.

## License

MIT
