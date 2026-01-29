# Browser Architecture

This document provides a visual overview of the browser's architecture and rendering pipeline.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              Browser Rendering Engine                                │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                      │
│   ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐       │
│   │  Input  │────▶│  Parse  │────▶│  Style  │────▶│ Layout  │────▶│ Render  │       │
│   │ (HTML)  │     │         │     │         │     │         │     │         │       │
│   └─────────┘     └─────────┘     └─────────┘     └─────────┘     └─────────┘       │
│                        │                                               │            │
│                        │                                               ▼            │
│                   ┌────┴────┐                                    ┌─────────┐        │
│                   │   CSS   │                                    │  Output │        │
│                   │ Parser  │                                    │  (PNG)  │        │
│                   └─────────┘                                    └─────────┘        │
│                                                                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## Rendering Pipeline

The browser follows a classic rendering pipeline, processing input through distinct stages:

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                RENDERING PIPELINE                                    │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                      │
│    ╔════════════════════════════════════════════════════════════════════════════╗   │
│    ║                          1. INPUT PROCESSING                                ║   │
│    ╠════════════════════════════════════════════════════════════════════════════╣   │
│    ║                                                                             ║   │
│    ║   ┌──────────────────────┐    ┌───────────────────────┐                    ║   │
│    ║   │  Local File (*.html) │    │ Remote URL (HTTP/S)   │                    ║   │
│    ║   └──────────┬───────────┘    └───────────┬───────────┘                    ║   │
│    ║              │                             │                                ║   │
│    ║              └─────────────┬───────────────┘                                ║   │
│    ║                            ▼                                                ║   │
│    ║                   ┌────────────────┐                                        ║   │
│    ║                   │  HTML Content  │                                        ║   │
│    ║                   └────────┬───────┘                                        ║   │
│    ║                            │                                                ║   │
│    ╚════════════════════════════╪════════════════════════════════════════════════╝   │
│                                 │                                                    │
│    ╔════════════════════════════╪════════════════════════════════════════════════╗   │
│    ║                            ▼         2. PARSING                             ║   │
│    ╠════════════════════════════════════════════════════════════════════════════╣   │
│    ║                                                                             ║   │
│    ║    ┌───────────────────────────────────────────────────────────────────┐   ║   │
│    ║    │                      HTML Parser (html/)                           │   ║   │
│    ║    ├───────────────────────────────────────────────────────────────────┤   ║   │
│    ║    │  ┌─────────────┐          ┌─────────────┐          ┌───────────┐  │   ║   │
│    ║    │  │  Tokenizer  │ ──────▶  │   Parser    │ ──────▶  │ DOM Tree  │  │   ║   │
│    ║    │  │ (tokenizer  │          │  (parser    │          │  (dom/    │  │   ║   │
│    ║    │  │    .go)     │          │    .go)     │          │  node.go) │  │   ║   │
│    ║    │  └─────────────┘          └─────────────┘          └─────┬─────┘  │   ║   │
│    ║    └──────────────────────────────────────────────────────────┼────────┘   ║   │
│    ║                                                                │            ║   │
│    ║    ┌───────────────────────────────────────────────────────────┼────────┐   ║   │
│    ║    │                      CSS Parser (css/)                    │        │   ║   │
│    ║    ├───────────────────────────────────────────────────────────┼────────┤   ║   │
│    ║    │  ┌─────────────┐          ┌─────────────┐                 │        │   ║   │
│    ║    │  │  Tokenizer  │ ──────▶  │   Parser    │ ──────▶ ┌──────┴──────┐ │   ║   │
│    ║    │  │ (tokenizer  │          │  (parser    │         │ Stylesheet  │ │   ║   │
│    ║    │  │    .go)     │          │    .go)     │         │   (Rules)   │ │   ║   │
│    ║    │  └─────────────┘          └─────────────┘         └──────┬──────┘ │   ║   │
│    ║    └──────────────────────────────────────────────────────────┼────────┘   ║   │
│    ║                                                                │            ║   │
│    ║    CSS Sources:                                                │            ║   │
│    ║    • <style> tags (inline)                                     │            ║   │
│    ║    • <link rel="stylesheet"> (external)                        │            ║   │
│    ║    • Inline style attributes                                   │            ║   │
│    ║                                                                │            ║   │
│    ╚════════════════════════════════════════════════════════════════╪════════════╝   │
│                                                                     │                │
│    ╔════════════════════════════════════════════════════════════════╪════════════╗   │
│    ║                                                                ▼            ║   │
│    ║                          3. STYLE COMPUTATION (style/)                      ║   │
│    ╠════════════════════════════════════════════════════════════════════════════╣   │
│    ║                                                                             ║   │
│    ║    ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                  ║   │
│    ║    │  Selector   │     │ Specificity │     │   Cascade   │                  ║   │
│    ║    │  Matching   │────▶│ Calculation │────▶│   & Merge   │                  ║   │
│    ║    └─────────────┘     └─────────────┘     └──────┬──────┘                  ║   │
│    ║                                                    │                        ║   │
│    ║                                                    ▼                        ║   │
│    ║                                           ┌──────────────┐                  ║   │
│    ║                                           │  Styled Tree │                  ║   │
│    ║                                           │ (DOM + Styles)│                  ║   │
│    ║                                           └───────┬──────┘                  ║   │
│    ║                                                   │                         ║   │
│    ╚═══════════════════════════════════════════════════╪═════════════════════════╝   │
│                                                        │                             │
│    ╔═══════════════════════════════════════════════════╪═════════════════════════╗   │
│    ║                                                   ▼                         ║   │
│    ║                          4. LAYOUT ENGINE (layout/)                         ║   │
│    ╠════════════════════════════════════════════════════════════════════════════╣   │
│    ║                                                                             ║   │
│    ║    ┌─────────────────────────────────────────────────────────────────┐     ║   │
│    ║    │                        Box Model (CSS 2.1 §8)                    │     ║   │
│    ║    │  ┌─────────────────────────────────────────────────────────┐    │     ║   │
│    ║    │  │  Margin                                                   │    │     ║   │
│    ║    │  │  ┌─────────────────────────────────────────────────────┐ │    │     ║   │
│    ║    │  │  │  Border                                              │ │    │     ║   │
│    ║    │  │  │  ┌─────────────────────────────────────────────────┐│ │    │     ║   │
│    ║    │  │  │  │  Padding                                        ││ │    │     ║   │
│    ║    │  │  │  │  ┌─────────────────────────────────────────────┐││ │    │     ║   │
│    ║    │  │  │  │  │              Content Box                    │││ │    │     ║   │
│    ║    │  │  │  │  └─────────────────────────────────────────────┘││ │    │     ║   │
│    ║    │  │  │  └─────────────────────────────────────────────────┘│ │    │     ║   │
│    ║    │  │  └─────────────────────────────────────────────────────┘ │    │     ║   │
│    ║    │  └─────────────────────────────────────────────────────────┘    │     ║   │
│    ║    └─────────────────────────────────────────────────────────────────┘     ║   │
│    ║                                                                             ║   │
│    ║    Layout Types:                           ┌───────────────┐               ║   │
│    ║    • Block Layout                          │  Layout Tree  │               ║   │
│    ║    • Inline Layout                    ────▶│   (Boxes +    │               ║   │
│    ║    • Table Layout                          │  Dimensions)  │               ║   │
│    ║    • Anonymous Boxes                       └───────┬───────┘               ║   │
│    ║                                                    │                        ║   │
│    ╚════════════════════════════════════════════════════╪════════════════════════╝   │
│                                                         │                            │
│    ╔════════════════════════════════════════════════════╪════════════════════════╗   │
│    ║                                                    ▼                        ║   │
│    ║                          5. RENDERING (render/)                             ║   │
│    ╠════════════════════════════════════════════════════════════════════════════╣   │
│    ║                                                                             ║   │
│    ║    ┌─────────────┐     ┌─────────────┐     ┌─────────────┐                  ║   │
│    ║    │ Background  │     │   Border    │     │    Text     │                  ║   │
│    ║    │  Painting   │────▶│  Painting   │────▶│  Rendering  │                  ║   │
│    ║    └─────────────┘     └─────────────┘     └─────────────┘                  ║   │
│    ║                                                    │                        ║   │
│    ║                                            ┌───────┴───────┐                ║   │
│    ║                                            │    Image      │                ║   │
│    ║                                            │   Rendering   │                ║   │
│    ║                                            └───────┬───────┘                ║   │
│    ║                                                    │                        ║   │
│    ║                                                    ▼                        ║   │
│    ║                                            ┌──────────────┐                 ║   │
│    ║                                            │  PNG Output  │                 ║   │
│    ║                                            │    Canvas    │                 ║   │
│    ║                                            └──────────────┘                 ║   │
│    ║                                                                             ║   │
│    ╚════════════════════════════════════════════════════════════════════════════╝   │
│                                                                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## Module Dependencies

```
                              ┌──────────────────────────────┐
                              │     cmd/browser/main.go      │
                              │     (Entry Point / CLI)      │
                              └───────────────┬──────────────┘
                                              │
                  ┌───────────────────────────┼───────────────────────────┐
                  │                           │                           │
                  ▼                           ▼                           ▼
         ┌────────────────┐          ┌────────────────┐          ┌────────────────┐
         │    html/       │          │     css/       │          │    render/     │
         │    Parser      │          │    Parser      │          │    Engine      │
         └───────┬────────┘          └────────┬───────┘          └────────┬───────┘
                 │                            │                           │
                 ▼                            │                           │
         ┌────────────────┐                   │                           │
         │     dom/       │◀──────────────────┘                           │
         │  DOM Tree &    │                                               │
         │  URL Utils     │                                               │
         └───────┬────────┘                                               │
                 │                                                        │
                 ▼                                                        │
         ┌────────────────┐                                               │
         │    style/      │                                               │
         │ Style Engine   │◀──────────────────────────────────────────────┤
         └───────┬────────┘                                               │
                 │                                                        │
                 ▼                                                        │
         ┌────────────────┐                                               │
         │    layout/     │                                               │
         │ Layout Engine  │◀──────────────────────────────────────────────┘
         └────────────────┘
```

## Data Flow

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                               DATA TRANSFORMATIONS                                   │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                      │
│  ┌─────────────┐                                                                    │
│  │ HTML String │                                                                    │
│  └──────┬──────┘                                                                    │
│         │ html.Parse()                                                              │
│         ▼                                                                           │
│  ┌─────────────┐      ┌─────────────┐                                               │
│  │  *dom.Node  │      │ CSS String  │                                               │
│  │  (DOM Tree) │      │             │                                               │
│  └──────┬──────┘      └──────┬──────┘                                               │
│         │                    │ css.Parse()                                          │
│         │                    ▼                                                      │
│         │             ┌─────────────┐                                               │
│         │             │ *css.Style  │                                               │
│         │             │ sheet       │                                               │
│         │             └──────┬──────┘                                               │
│         │                    │                                                      │
│         └────────┬───────────┘                                                      │
│                  │ style.StyleTree()                                                │
│                  ▼                                                                  │
│           ┌─────────────┐                                                           │
│           │*style.Styled│                                                           │
│           │    Node     │                                                           │
│           │(DOM+Styles) │                                                           │
│           └──────┬──────┘                                                           │
│                  │ layout.LayoutTree()                                              │
│                  ▼                                                                  │
│           ┌─────────────┐                                                           │
│           │*layout.     │                                                           │
│           │ LayoutBox   │                                                           │
│           │(Boxes+Dims) │                                                           │
│           └──────┬──────┘                                                           │
│                  │ render.Render()                                                  │
│                  ▼                                                                  │
│           ┌─────────────┐                                                           │
│           │*render.     │                                                           │
│           │  Canvas     │                                                           │
│           │ (Pixels)    │                                                           │
│           └──────┬──────┘                                                           │
│                  │ canvas.SavePNG()                                                 │
│                  ▼                                                                  │
│           ┌─────────────┐                                                           │
│           │  PNG File   │                                                           │
│           └─────────────┘                                                           │
│                                                                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## Key Data Structures

### DOM Node (`dom/node.go`)
```
Node
├── Type: NodeType (Document, Element, Text)
├── Data: string (tag name or text content)
├── Attributes: map[string]string
├── Children: []*Node
└── Parent: *Node
```

### CSS Stylesheet (`css/parser.go`)
```
Stylesheet
└── Rules: []*Rule
    ├── Selectors: []*Selector
    │   └── Simple: []*SimpleSelector    (list of selectors for descendant combinator)
    │       ├── TagName: string
    │       ├── ID: string
    │       ├── Classes: []string
    │       ├── PseudoClasses: []string
    │       └── PseudoElements: []string
    └── Declarations: []*Declaration
        ├── Property: string
        └── Value: string
```

### Styled Node (`style/style.go`)
```
StyledNode
├── Node: *dom.Node
├── Styles: map[string]string (computed styles)
└── Children: []*StyledNode
```

### Layout Box (`layout/layout.go`)
```
LayoutBox
├── BoxType: BoxType (Block, Inline, Anonymous, Table, etc.)
├── Dimensions: Dimensions
│   ├── Content: Rect (x, y, width, height)
│   ├── Padding: EdgeSizes (top, right, bottom, left)
│   ├── Border: EdgeSizes
│   └── Margin: EdgeSizes
├── StyledNode: *StyledNode
└── Children: []*LayoutBox
```

## WebAssembly Target

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                             WEBASSEMBLY ARCHITECTURE                                 │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                      │
│    ┌───────────────────────────────────────────────────────────────────────────┐    │
│    │                         Web Browser (Host)                                 │    │
│    │  ┌─────────────────────────────────────────────────────────────────────┐  │    │
│    │  │                     JavaScript Runtime                               │  │    │
│    │  │                                                                      │  │    │
│    │  │  ┌──────────────┐         ┌──────────────────────────────────────┐  │  │    │
│    │  │  │  HTML Input  │ ──────▶ │        browser.wasm                   │  │  │    │
│    │  │  │  (textarea)  │         │   (Go compiled to WebAssembly)       │  │  │    │
│    │  │  └──────────────┘         │                                      │  │  │    │
│    │  │                           │  ┌────────────────────────────────┐  │  │  │    │
│    │  │                           │  │ Full Rendering Pipeline:       │  │  │  │    │
│    │  │                           │  │ • HTML Parser                  │  │  │  │    │
│    │  │                           │  │ • CSS Parser                   │  │  │  │    │
│    │  │                           │  │ • Style Computation            │  │  │  │    │
│    │  │                           │  │ • Layout Engine                │  │  │  │    │
│    │  │                           │  │ • Render Engine                │  │  │  │    │
│    │  │                           │  └────────────────────────────────┘  │  │  │    │
│    │  │  ┌──────────────┐         │                │                     │  │  │    │
│    │  │  │  <canvas>    │ ◀────── │  Base64 PNG    │                     │  │  │    │
│    │  │  │  (output)    │         └────────────────┼─────────────────────┘  │  │    │
│    │  │  └──────────────┘                          │                        │  │    │
│    │  │                                            │                        │  │    │
│    │  └────────────────────────────────────────────┼────────────────────────┘  │    │
│    │                                               │                           │    │
│    └───────────────────────────────────────────────┼───────────────────────────┘    │
│                                                    │                                │
│    Entry Point: cmd/browser-wasm/main.go           │                                │
│    Demo Page: wasm/index.html                      │                                │
│                                                                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## Specifications Followed

| Component | Specification |
|-----------|--------------|
| HTML Parsing | [HTML5 §12.2](https://html.spec.whatwg.org/multipage/parsing.html) |
| CSS Syntax | [CSS 2.1 §4](https://www.w3.org/TR/CSS21/syndata.html) |
| CSS Selectors | [CSS 2.1 §5](https://www.w3.org/TR/CSS21/selector.html) |
| CSS Cascade | [CSS 2.1 §6](https://www.w3.org/TR/CSS21/cascade.html) |
| Box Model | [CSS 2.1 §8](https://www.w3.org/TR/CSS21/box.html) |
| Visual Formatting | [CSS 2.1 §9](https://www.w3.org/TR/CSS21/visuren.html) |
| URL Resolution | [HTML5 §2.5](https://html.spec.whatwg.org/multipage/urls-and-fetching.html) |
| Data URLs | [RFC 2397](https://datatracker.ietf.org/doc/html/rfc2397) |

## Directory Structure

```
browser/
├── cmd/
│   ├── browser/           # CLI application entry point
│   │   └── main.go        # Main function, orchestrates pipeline
│   └── browser-wasm/      # WebAssembly entry point
│       └── main.go        # WASM-specific initialization
├── dom/                   # Document Object Model
│   ├── node.go            # Node structure and tree operations
│   ├── url.go             # URL resolution utilities
│   └── loader.go          # External resource loading
├── html/                  # HTML parsing
│   ├── tokenizer.go       # HTML tokenization (state machine)
│   └── parser.go          # Tree construction algorithm
├── css/                   # CSS parsing
│   ├── tokenizer.go       # CSS tokenization
│   ├── parser.go          # Selector/declaration parsing
│   └── values.go          # CSS value parsing utilities
├── style/                 # Style computation
│   ├── style.go           # Selector matching, cascade
│   └── useragent.go       # Default browser styles
├── layout/                # Layout engine
│   └── layout.go          # Box model, positioning
├── render/                # Rendering engine
│   └── render.go          # Canvas, painting, PNG output
├── font/                  # Font handling
│   └── font.go            # Text measurement and rendering
├── svg/                   # SVG support
│   └── svg.go             # SVG parsing and rendering
├── log/                   # Logging utilities
│   └── log.go             # Leveled logging
├── reftest/               # Reference tests
│   └── reftest.go         # WPT test harness
├── wasm/                  # WebAssembly demo
│   ├── index.html         # Demo page
│   └── browser.wasm       # Compiled WASM (generated)
└── test/                  # Test fixtures
    ├── simple.html
    ├── styled.html
    └── (various test files)
```
