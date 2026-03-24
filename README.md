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

The browser follows a classic rendering pipeline that transforms HTML/CSS input into a rendered PNG image:

```mermaid
flowchart TB
    subgraph Input["📥 Input"]
        HTML["HTML File/URL"]
        CSS["CSS Stylesheets"]
    end

    subgraph Parsing["🔍 Parsing"]
        HTMLTokenizer["HTML Tokenizer<br/><i>html/tokenizer.go</i>"]
        HTMLParser["HTML Parser<br/><i>html/parser.go</i>"]
        CSSTokenizer["CSS Tokenizer<br/><i>css/tokenizer.go</i>"]
        CSSParser["CSS Parser<br/><i>css/parser.go</i>"]
    end

    subgraph Trees["🌳 Data Structures"]
        DOM["DOM Tree<br/><i>dom/node.go</i>"]
        Stylesheet["Stylesheet<br/>(Rules + Selectors)"]
        StyledTree["Styled Tree<br/>(DOM + Computed Styles)"]
        LayoutTree["Layout Tree<br/>(Box Model)"]
    end

    subgraph Processing["⚙️ Processing"]
        StyleEngine["Style Engine<br/><i>style/style.go</i><br/>Selector Matching<br/>Specificity & Cascade"]
        LayoutEngine["Layout Engine<br/><i>layout/layout.go</i><br/>Box Model<br/>Block/Inline Layout"]
    end

    subgraph Output["📤 Output"]
        RenderEngine["Render Engine<br/><i>render/render.go</i><br/>Canvas Drawing<br/>Text & Images"]
        PNG["PNG Image"]
    end

    HTML --> HTMLTokenizer --> HTMLParser --> DOM
    CSS --> CSSTokenizer --> CSSParser --> Stylesheet
    DOM --> StyleEngine
    Stylesheet --> StyleEngine
    StyleEngine --> StyledTree
    StyledTree --> LayoutEngine
    LayoutEngine --> LayoutTree
    LayoutTree --> RenderEngine
    RenderEngine --> PNG
```

### Pipeline Stages

| Stage | Module | Description |
|-------|--------|-------------|
| **1. HTML Parsing** | `html/` | Tokenizes HTML into tokens, builds DOM tree via tree construction algorithm |
| **2. CSS Parsing** | `css/` | Tokenizes CSS, parses selectors and declarations into stylesheet rules |
| **3. Style Computation** | `style/` | Matches selectors to DOM nodes, calculates specificity, cascades styles |
| **4. Layout** | `layout/` | Calculates box dimensions (content, padding, border, margin) and positions |
| **5. Rendering** | `render/` | Draws backgrounds, borders, text, and images to canvas; outputs PNG |

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
