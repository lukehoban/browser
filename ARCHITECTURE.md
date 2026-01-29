# Browser Architecture

This document describes the architecture of the Go-based simplified web browser.

## Overview

This browser is an educational implementation focused on CSS 2.1 compliance. It parses HTML and CSS, computes styles following CSS cascade rules, calculates box model layouts, and renders visual output to PNG files.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              INPUT                                          │
│                    (HTML file or HTTP/HTTPS URL)                            │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         RESOURCE LOADING                                    │
│                         (dom/loader.go)                                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ HTTP/HTTPS  │  │ Local File  │  │  Data URL   │  │   URL Resolution    │ │
│  │   Fetch     │  │    Read     │  │  (RFC 2397) │  │   (dom/url.go)      │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           HTML PARSING                                      │
│                           (html/)                                           │
│  ┌──────────────────────────────┐    ┌──────────────────────────────────┐   │
│  │     Tokenizer                │    │       Parser                     │   │
│  │     (tokenizer.go)           │───▶│       (parser.go)                │   │
│  │     HTML5 state machine      │    │       Tree construction          │   │
│  └──────────────────────────────┘    └──────────────────────────────────┘   │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            DOM TREE                                         │
│                           (dom/node.go)                                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │DocumentNode │  │ ElementNode │  │  TextNode   │  │    Attributes       │ │
│  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└───────────────────────────┬─────────────────────────────────────────────────┘
                            │
           ┌────────────────┼────────────────┐
           │                │                │
           ▼                ▼                ▼
┌──────────────────┐ ┌─────────────┐ ┌───────────────────┐
│  <style> tags    │ │ <link>      │ │  Inline styles    │
│  extraction      │ │ stylesheets │ │  (style attr)     │
└────────┬─────────┘ └──────┬──────┘ └─────────┬─────────┘
         │                  │                  │
         └──────────────────┼──────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CSS PARSING                                       │
│                            (css/)                                           │
│  ┌──────────────────────────────┐    ┌──────────────────────────────────┐   │
│  │     Tokenizer                │    │       Parser                     │   │
│  │     (tokenizer.go)           │───▶│       (parser.go)                │   │
│  │     CSS tokenization         │    │       Rules, Selectors, Decls    │   │
│  └──────────────────────────────┘    └──────────────────────────────────┘   │
│                                                                             │
│  Selector Support:  element │ .class │ #id │ descendant │ pseudo-class     │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        STYLE COMPUTATION                                    │
│                         (style/style.go)                                    │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────────┐  │
│  │ Selector        │  │  Specificity    │  │   CSS Cascade               │  │
│  │ Matching        │  │  Calculation    │  │   (CSS 2.1 §6.4.3)          │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────────────────┘  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────────┐  │
│  │ Property        │  │  User-Agent     │  │   Pseudo-element Support    │  │
│  │ Inheritance     │  │  Stylesheet     │  │   (::before, ::after, etc)  │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────────────────┘  │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         STYLED DOM TREE                                     │
│                          (StyledNode)                                       │
│                    DOM Node + Computed Styles Map                           │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        LAYOUT COMPUTATION                                   │
│                        (layout/layout.go)                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────────┐  │
│  │   Box Model     │  │  Block Layout   │  │   Table Layout              │  │
│  │   (CSS 2.1)     │  │  (Normal Flow)  │  │   (Column Auto-sizing)      │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────────────────┘  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────────┐  │
│  │  Width/Height   │  │   Positioning   │  │   Text Measurement          │  │
│  │  Calculation    │  │   (x, y coords) │  │   ◄─── font/font.go         │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────────────────┘  │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          LAYOUT TREE                                        │
│                          (LayoutBox)                                        │
│              Position (x, y) + Dimensions + Box Hierarchy                   │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           RENDERING                                         │
│                        (render/render.go)                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────────┐  │
│  │  Background     │  │  Border         │  │   Text Rendering            │  │
│  │  Colors         │  │  Rendering      │  │   (TrueType fonts)          │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────────────────┘  │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────────────────┐  │
│  │  Image Loading  │  │  SVG Rendering  │  │   Image Caching             │  │
│  │  (PNG/JPEG/GIF) │  │  (svg/)         │  │                             │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────────────────┘  │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            OUTPUT                                           │
│  ┌──────────────────────────────┐    ┌──────────────────────────────────┐   │
│  │   Canvas (Pixel Buffer)      │───▶│   PNG Encoding & File Output    │   │
│  └──────────────────────────────┘    └──────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Module Dependencies

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            cmd/browser                                      │
│                         (CLI Entry Point)                                   │
└───────────────────────────────────┬─────────────────────────────────────────┘
                                    │
        ┌──────────────┬────────────┼───────────┬──────────────┐
        ▼              ▼            ▼           ▼              ▼
   ┌─────────┐   ┌──────────┐  ┌─────────┐ ┌─────────┐   ┌──────────┐
   │   dom   │   │   html   │  │   css   │ │  style  │   │  layout  │
   └────┬────┘   └────┬─────┘  └────┬────┘ └────┬────┘   └────┬─────┘
        │             │             │           │              │
        │             └─────────────┼───────────┼──────────────┤
        │                           │           │              │
        ▼                           ▼           ▼              ▼
   ┌─────────┐                 ┌─────────┐ ┌─────────┐   ┌──────────┐
   │  loader │                 │  rules  │ │ cascade │   │   font   │
   │  url    │                 │selector │ │ inherit │   │          │
   └─────────┘                 └─────────┘ └─────────┘   └────┬─────┘
                                                              │
                                                              ▼
                                                        ┌──────────┐
                                                        │  render  │
                                                        └────┬─────┘
                                                             │
                                              ┌──────────────┼──────────────┐
                                              ▼              ▼              ▼
                                         ┌─────────┐   ┌──────────┐   ┌─────────┐
                                         │  canvas │   │   svg    │   │  image  │
                                         └─────────┘   └──────────┘   └─────────┘
```

## Directory Structure

```
browser/
├── cmd/
│   ├── browser/           # CLI application
│   │   └── main.go        # Main entry point
│   └── browser-wasm/      # WebAssembly build
│       └── main.go        # WASM entry with JS bindings
├── dom/
│   ├── node.go            # DOM tree: Document, Element, Text nodes
│   ├── loader.go          # Resource loading (HTTP, file, data URL)
│   └── url.go             # URL resolution
├── html/
│   ├── tokenizer.go       # HTML5-compliant tokenizer
│   └── parser.go          # Tree construction algorithm
├── css/
│   ├── tokenizer.go       # CSS tokenization
│   ├── parser.go          # Stylesheet parsing
│   └── values.go          # CSS value handling
├── style/
│   └── style.go           # Selector matching, cascade, inheritance
├── layout/
│   └── layout.go          # Box model, positioning, dimensions
├── render/
│   └── render.go          # Rasterization, text/image drawing
├── font/
│   └── font.go            # Font loading, text measurement
├── svg/
│   ├── svg.go             # SVG parsing
│   └── rasterizer.go      # SVG path rendering
└── log/
    └── logger.go          # Internal logging
```

## Data Flow Summary

| Stage | Input | Output | Module |
|-------|-------|--------|--------|
| 1. Load | URL/File path | HTML string | `dom/loader` |
| 2. Parse HTML | HTML string | DOM Tree | `html/` |
| 3. Parse CSS | Style tags, linked CSS | Stylesheet rules | `css/` |
| 4. Compute Styles | DOM + Rules | Styled Tree | `style/` |
| 5. Layout | Styled Tree | Layout Tree | `layout/` |
| 6. Render | Layout Tree | Pixel Buffer | `render/` |
| 7. Output | Pixel Buffer | PNG file | `render/` |

## Build Targets

### Native CLI
```bash
go build ./cmd/browser
./browser -output output.png -width 800 -height 600 input.html
```

### WebAssembly
```bash
GOOS=js GOARCH=wasm go build -o browser.wasm ./cmd/browser-wasm
```

## Standards Compliance

- **HTML5**: Tokenization & parsing (WHATWG §12.2)
- **CSS 2.1**: Cascade, specificity, box model, visual formatting
- **RFC 2397**: Data URL scheme support
