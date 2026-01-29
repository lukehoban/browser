# Architecture

This document describes the architecture of the browser project - a web browser implementation in Go that parses HTML and CSS, computes styles, calculates layout, and renders to PNG images.

## High-Level Overview

The browser follows a classic rendering pipeline, transforming HTML/CSS input into visual output:

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              BROWSER RENDERING PIPELINE                          │
└─────────────────────────────────────────────────────────────────────────────────┘

┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│    INPUT     │    │   PARSING    │    │    STYLE     │    │    LAYOUT    │    │   RENDER     │
│              │───▶│              │───▶│              │───▶│              │───▶│              │
│  HTML + CSS  │    │  DOM + CSSOM │    │  Styled Tree │    │  Layout Tree │    │  PNG Output  │
└──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘
```

## Component Diagram

```mermaid
graph TB
    subgraph Input["📥 Input Layer"]
        URL["URL / File Path"]
        Network["Network Fetcher<br/>(HTTP/HTTPS)"]
        FileSystem["File System<br/>(Local Files)"]
    end

    subgraph Parsing["📝 Parsing Layer"]
        HTMLTokenizer["HTML Tokenizer<br/><code>html/tokenizer.go</code>"]
        HTMLParser["HTML Parser<br/><code>html/parser.go</code>"]
        CSSTokenizer["CSS Tokenizer<br/><code>css/tokenizer.go</code>"]
        CSSParser["CSS Parser<br/><code>css/parser.go</code>"]
    end

    subgraph DataStructures["📊 Core Data Structures"]
        DOM["DOM Tree<br/><code>dom/node.go</code>"]
        Stylesheet["Stylesheet<br/>(Rules + Selectors)"]
    end

    subgraph Styling["🎨 Styling Layer"]
        StyleEngine["Style Engine<br/><code>style/style.go</code>"]
        UserAgent["User Agent Styles<br/><code>style/useragent.go</code>"]
        StyledTree["Styled Tree<br/>(DOM + Computed Styles)"]
    end

    subgraph Layout["📐 Layout Layer"]
        LayoutEngine["Layout Engine<br/><code>layout/layout.go</code>"]
        LayoutTree["Layout Tree<br/>(Boxes + Dimensions)"]
    end

    subgraph Rendering["🖼️ Rendering Layer"]
        RenderEngine["Render Engine<br/><code>render/render.go</code>"]
        FontEngine["Font Engine<br/><code>font/font.go</code>"]
        SVGRasterizer["SVG Rasterizer<br/><code>svg/</code>"]
        Canvas["Canvas<br/>(Pixel Buffer)"]
    end

    subgraph Output["📤 Output"]
        PNG["PNG Image"]
    end

    %% Flow connections
    URL --> Network
    URL --> FileSystem
    Network --> HTMLTokenizer
    FileSystem --> HTMLTokenizer
    
    HTMLTokenizer --> HTMLParser
    HTMLParser --> DOM
    
    DOM --> StyleEngine
    CSSTokenizer --> CSSParser
    CSSParser --> Stylesheet
    Stylesheet --> StyleEngine
    UserAgent --> StyleEngine
    
    StyleEngine --> StyledTree
    StyledTree --> LayoutEngine
    LayoutEngine --> LayoutTree
    
    LayoutTree --> RenderEngine
    FontEngine --> RenderEngine
    SVGRasterizer --> RenderEngine
    RenderEngine --> Canvas
    Canvas --> PNG
```

## Data Flow

```mermaid
flowchart LR
    subgraph Phase1["Phase 1: Fetch"]
        A1["HTML Content"] --> A2["CSS Content<br/>(inline + external)"]
    end

    subgraph Phase2["Phase 2: Parse"]
        B1["Tokenize HTML"] --> B2["Build DOM Tree"]
        B3["Tokenize CSS"] --> B4["Build Stylesheet"]
    end

    subgraph Phase3["Phase 3: Style"]
        C1["Match Selectors"] --> C2["Calculate Specificity"]
        C2 --> C3["Apply Cascade"]
        C3 --> C4["Styled Tree"]
    end

    subgraph Phase4["Phase 4: Layout"]
        D1["Generate Boxes"] --> D2["Calculate Widths"]
        D2 --> D3["Calculate Heights"]
        D3 --> D4["Position Elements"]
    end

    subgraph Phase5["Phase 5: Render"]
        E1["Draw Backgrounds"] --> E2["Draw Borders"]
        E2 --> E3["Draw Text"]
        E3 --> E4["Draw Images"]
        E4 --> E5["Output PNG"]
    end

    Phase1 --> Phase2 --> Phase3 --> Phase4 --> Phase5
```

## Module Dependencies

```mermaid
graph BT
    cmd["cmd/browser<br/>(CLI Entry Point)"]
    
    render["render<br/>(Drawing)"]
    layout["layout<br/>(Box Model)"]
    style["style<br/>(Cascade)"]
    css["css<br/>(CSS Parsing)"]
    html["html<br/>(HTML Parsing)"]
    dom["dom<br/>(Data Structure)"]
    font["font<br/>(Typography)"]
    svg["svg<br/>(Vector Graphics)"]
    log["log<br/>(Logging)"]
    
    cmd --> render
    cmd --> layout
    cmd --> style
    cmd --> css
    cmd --> html
    cmd --> dom
    cmd --> log
    
    render --> layout
    render --> font
    render --> svg
    render --> dom
    
    layout --> style
    layout --> dom
    
    style --> css
    style --> dom
    
    html --> dom
```

## Key Data Structures

### DOM Node (`dom/node.go`)

```
┌─────────────────────────────────────────────┐
│                  Node                        │
├─────────────────────────────────────────────┤
│  Type: NodeType (Document/Element/Text)     │
│  Data: string (tag name or text content)    │
│  Attributes: map[string]string              │
│  Children: []*Node                          │
│  Parent: *Node                              │
└─────────────────────────────────────────────┘
```

### Stylesheet (`css/parser.go`)

```
┌─────────────────────────────────────────────┐
│               Stylesheet                     │
├─────────────────────────────────────────────┤
│  Rules: []Rule                              │
│    ├── Selectors: []Selector                │
│    │     ├── Tag: string                    │
│    │     ├── ID: string                     │
│    │     ├── Classes: []string              │
│    │     └── Combinators: []Combinator      │
│    └── Declarations: []Declaration          │
│          ├── Property: string               │
│          └── Value: string                  │
└─────────────────────────────────────────────┘
```

### Styled Node (`style/style.go`)

```
┌─────────────────────────────────────────────┐
│              StyledNode                      │
├─────────────────────────────────────────────┤
│  Node: *dom.Node                            │
│  Styles: map[string]string                  │
│  Children: []*StyledNode                    │
└─────────────────────────────────────────────┘
```

### Layout Box (`layout/layout.go`)

```
┌─────────────────────────────────────────────┐
│               LayoutBox                      │
├─────────────────────────────────────────────┤
│  BoxType: BoxType (Block/Inline/Anonymous)  │
│  Dimensions: Dimensions                     │
│    ├── Content: Rect (x, y, width, height)  │
│    ├── Padding: EdgeSize (t, r, b, l)       │
│    ├── Border: EdgeSize (t, r, b, l)        │
│    └── Margin: EdgeSize (t, r, b, l)        │
│  StyledNode: *StyledNode                    │
│  Children: []*LayoutBox                     │
└─────────────────────────────────────────────┘
```

## CSS Box Model

```
┌───────────────────────────────────────────────────────────────┐
│                         MARGIN                                 │
│   ┌───────────────────────────────────────────────────────┐   │
│   │                       BORDER                           │   │
│   │   ┌───────────────────────────────────────────────┐   │   │
│   │   │                   PADDING                      │   │   │
│   │   │   ┌───────────────────────────────────────┐   │   │   │
│   │   │   │                                       │   │   │   │
│   │   │   │              CONTENT                  │   │   │   │
│   │   │   │         (width × height)              │   │   │   │
│   │   │   │                                       │   │   │   │
│   │   │   └───────────────────────────────────────┘   │   │   │
│   │   │                                               │   │   │
│   │   └───────────────────────────────────────────────┘   │   │
│   │                                                       │   │
│   └───────────────────────────────────────────────────────┘   │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

## Rendering Pipeline Detail

```mermaid
sequenceDiagram
    participant CLI as CLI (main.go)
    participant HTML as HTML Parser
    participant CSS as CSS Parser
    participant DOM as DOM Tree
    participant Style as Style Engine
    participant Layout as Layout Engine
    participant Render as Render Engine

    CLI->>HTML: Parse HTML content
    HTML->>DOM: Build DOM tree
    CLI->>CSS: Parse CSS content
    CSS->>Style: Provide stylesheet
    CLI->>Style: StyleTree(DOM, Stylesheet)
    Style->>Style: Match selectors to elements
    Style->>Style: Calculate specificity
    Style->>Style: Apply cascade
    Style-->>CLI: Return StyledTree
    CLI->>Layout: LayoutTree(StyledTree, viewport)
    Layout->>Layout: Generate boxes
    Layout->>Layout: Calculate dimensions
    Layout->>Layout: Position elements
    Layout-->>CLI: Return LayoutTree
    CLI->>Render: Render(LayoutTree, width, height)
    Render->>Render: Draw backgrounds
    Render->>Render: Draw borders
    Render->>Render: Draw text
    Render->>Render: Draw images
    Render-->>CLI: Return Canvas
    CLI->>CLI: SavePNG(output)
```

## Directory Structure

```
browser/
├── cmd/
│   ├── browser/           # Main CLI application
│   │   └── main.go        # Entry point, orchestrates pipeline
│   └── browser-wasm/      # WebAssembly entry point
│       └── main.go        # WASM-specific initialization
│
├── html/                  # HTML Parsing (HTML5 §12)
│   ├── tokenizer.go       # State machine tokenizer
│   └── parser.go          # Tree construction algorithm
│
├── css/                   # CSS Parsing (CSS 2.1 §4)
│   ├── tokenizer.go       # CSS token generation
│   ├── parser.go          # Selector & declaration parsing
│   └── values.go          # CSS value parsing utilities
│
├── dom/                   # DOM Data Structure
│   ├── node.go            # Node type definitions
│   ├── url.go             # URL resolution (HTML5 §2.5)
│   └── loader.go          # External resource loading
│
├── style/                 # Style Computation (CSS 2.1 §6)
│   ├── style.go           # Selector matching, specificity, cascade
│   └── useragent.go       # Default browser styles
│
├── layout/                # Layout Engine (CSS 2.1 §8-10)
│   └── layout.go          # Box model, dimensions, positioning
│
├── render/                # Rendering Engine (CSS 2.1 §14-16)
│   └── render.go          # Canvas, drawing operations, PNG output
│
├── font/                  # Font Engine
│   └── font.go            # Go fonts integration, text measurement
│
├── svg/                   # SVG Support
│   ├── svg.go             # SVG parsing
│   └── rasterizer.go      # SVG to raster conversion
│
├── log/                   # Logging
│   └── log.go             # Configurable logging levels
│
├── reftest/               # Reference Testing
│   └── reftest.go         # WPT reftest harness
│
├── wasm/                  # WebAssembly Demo
│   ├── index.html         # Demo page
│   └── README.md          # WASM documentation
│
└── test/                  # Test Files
    ├── simple.html        # Basic HTML test
    ├── styled.html        # CSS styling test
    └── hackernews.html    # Complex layout test
```

## Supported Features

| Component | Feature | Status |
|-----------|---------|--------|
| **HTML** | Tokenization | ✅ |
| | Tree construction | ✅ |
| | Void elements | ✅ |
| | Attributes | ✅ |
| **CSS** | Selectors (element, class, ID) | ✅ |
| | Descendant combinator | ✅ |
| | Specificity | ✅ |
| | Cascade | ✅ |
| **Layout** | Block layout | ✅ |
| | Box model | ✅ |
| | Auto/fixed/% widths | ✅ |
| | Tables (basic) | ✅ |
| **Render** | Backgrounds | ✅ |
| | Borders | ✅ |
| | Text (Go fonts) | ✅ |
| | Images (PNG/JPEG/GIF/SVG) | ✅ |
| **Network** | HTTP/HTTPS | ✅ |
| | External stylesheets | ✅ |
| | Data URLs | ✅ |
| **Output** | PNG | ✅ |
| | WebAssembly | ✅ |

## W3C Specification References

- **HTML5**: [Parsing](https://html.spec.whatwg.org/multipage/parsing.html) (§12)
- **CSS 2.1**: [Syntax](https://www.w3.org/TR/CSS21/syndata.html) (§4)
- **CSS 2.1**: [Selectors](https://www.w3.org/TR/CSS21/selector.html) (§5)
- **CSS 2.1**: [Cascade](https://www.w3.org/TR/CSS21/cascade.html) (§6)
- **CSS 2.1**: [Box Model](https://www.w3.org/TR/CSS21/box.html) (§8)
- **CSS 2.1**: [Visual Formatting](https://www.w3.org/TR/CSS21/visuren.html) (§9)
- **RFC 2397**: [Data URLs](https://datatracker.ietf.org/doc/html/rfc2397)
