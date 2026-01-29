# Browser Architecture

This document describes the architecture of the Go Browser implementation.

## What Does This Repository Do?

This repository implements a **simple web browser in Go** that can:

1. **Parse HTML** documents into a DOM tree
2. **Parse CSS** stylesheets into rule sets
3. **Compute styles** by matching CSS selectors to DOM elements
4. **Calculate layout** using the CSS box model
5. **Render** the result to PNG images

The browser supports both local HTML files and remote URLs (HTTP/HTTPS), and can run natively as a CLI tool or in web browsers via WebAssembly.

## Architecture Diagram

```mermaid
flowchart TB
    subgraph Input["📥 Input Layer"]
        URL["URL (HTTP/HTTPS)"]
        File["Local HTML File"]
        WASM["WebAssembly Input"]
    end

    subgraph Loader["🌐 Network/Loader Layer"]
        direction TB
        Fetch["HTTP Fetcher<br/>(net/http)"]
        FileRead["File Reader"]
        DataURL["Data URL Parser<br/>(RFC 2397)"]
        ExtCSS["External CSS Loader<br/>(&lt;link rel='stylesheet'&gt;)"]
        ImgLoader["Image Loader<br/>(PNG, JPEG, GIF, SVG)"]
    end

    subgraph Parsing["📝 Parsing Layer"]
        direction TB
        HTMLTokenizer["HTML Tokenizer<br/>(html/tokenizer.go)"]
        HTMLParser["HTML Parser<br/>(html/parser.go)"]
        CSSTokenizer["CSS Tokenizer<br/>(css/tokenizer.go)"]
        CSSParser["CSS Parser<br/>(css/parser.go)"]
    end

    subgraph DOM["🌳 DOM Layer"]
        DOMTree["DOM Tree<br/>(dom/node.go)"]
        URLResolver["URL Resolver<br/>(dom/url.go)"]
    end

    subgraph Style["🎨 Style Layer"]
        direction TB
        Matcher["Selector Matcher"]
        Specificity["Specificity Calculator<br/>(CSS 2.1 §6.4.3)"]
        Cascade["Cascade Engine"]
        UAStyles["User-Agent Stylesheet<br/>(style/useragent.go)"]
        InlineStyles["Inline Style Parser<br/>(style attribute)"]
    end

    subgraph Layout["📐 Layout Layer"]
        direction TB
        BoxModel["Box Model Calculator<br/>(CSS 2.1 §8)"]
        BlockLayout["Block Layout<br/>(CSS 2.1 §9.2)"]
        InlineLayout["Inline Layout<br/>(CSS 2.1 §9.2.2)"]
        TableLayout["Table Layout<br/>(CSS 2.1 §17)"]
    end

    subgraph Render["🖼️ Render Layer"]
        direction TB
        Canvas["Canvas<br/>(render/render.go)"]
        TextRender["Text Renderer<br/>(Go Fonts)"]
        BGRender["Background Renderer"]
        BorderRender["Border Renderer"]
        ImgRender["Image Renderer"]
        SVGRaster["SVG Rasterizer<br/>(svg/rasterizer.go)"]
    end

    subgraph Output["📤 Output"]
        PNG["PNG Image"]
        LayoutTree["Layout Tree<br/>(Debug Output)"]
    end

    %% Input connections
    URL --> Fetch
    File --> FileRead
    WASM --> HTMLTokenizer

    %% Loader connections
    Fetch --> HTMLTokenizer
    FileRead --> HTMLTokenizer
    Fetch --> ExtCSS
    Fetch --> ImgLoader
    DataURL --> ImgLoader

    %% Parsing connections
    HTMLTokenizer --> HTMLParser
    HTMLParser --> DOMTree
    CSSTokenizer --> CSSParser
    ExtCSS --> CSSTokenizer

    %% DOM connections
    DOMTree --> URLResolver
    URLResolver --> Matcher

    %% Style connections
    CSSParser --> Matcher
    Matcher --> Specificity
    Specificity --> Cascade
    UAStyles --> Cascade
    InlineStyles --> Cascade
    Cascade --> BoxModel

    %% Layout connections
    BoxModel --> BlockLayout
    BoxModel --> InlineLayout
    BoxModel --> TableLayout
    BlockLayout --> Canvas
    InlineLayout --> Canvas
    TableLayout --> Canvas

    %% Render connections
    Canvas --> TextRender
    Canvas --> BGRender
    Canvas --> BorderRender
    Canvas --> ImgRender
    ImgRender --> SVGRaster
    ImgLoader --> ImgRender

    %% Output connections
    Canvas --> PNG
    BoxModel --> LayoutTree

    %% Styling
    classDef inputClass fill:#e1f5fe,stroke:#01579b
    classDef loaderClass fill:#fff3e0,stroke:#e65100
    classDef parseClass fill:#f3e5f5,stroke:#7b1fa2
    classDef domClass fill:#e8f5e9,stroke:#2e7d32
    classDef styleClass fill:#fce4ec,stroke:#c2185b
    classDef layoutClass fill:#fff8e1,stroke:#f9a825
    classDef renderClass fill:#e0f2f1,stroke:#00695c
    classDef outputClass fill:#eceff1,stroke:#546e7a

    class URL,File,WASM inputClass
    class Fetch,FileRead,DataURL,ExtCSS,ImgLoader loaderClass
    class HTMLTokenizer,HTMLParser,CSSTokenizer,CSSParser parseClass
    class DOMTree,URLResolver domClass
    class Matcher,Specificity,Cascade,UAStyles,InlineStyles styleClass
    class BoxModel,BlockLayout,InlineLayout,TableLayout layoutClass
    class Canvas,TextRender,BGRender,BorderRender,ImgRender,SVGRaster renderClass
    class PNG,LayoutTree outputClass
```

## Rendering Pipeline

The browser follows a classic web rendering pipeline:

```mermaid
flowchart LR
    A["HTML Input"] --> B["Tokenization"]
    B --> C["DOM Tree"]
    C --> D["Style Computation"]
    D --> E["Layout"]
    E --> F["Rendering"]
    F --> G["PNG Output"]
    
    CSS["CSS Input"] --> D
```

### Pipeline Stages

| Stage | Package | Key Files | Description |
|-------|---------|-----------|-------------|
| **1. Tokenization** | `html/` | `tokenizer.go` | Converts HTML text into tokens (start tags, end tags, text, etc.) |
| **2. Parsing** | `html/` | `parser.go` | Builds DOM tree from tokens |
| **3. CSS Parsing** | `css/` | `tokenizer.go`, `parser.go` | Parses CSS rules and selectors |
| **4. Style Computation** | `style/` | `style.go` | Matches selectors to elements, computes cascade |
| **5. Layout** | `layout/` | `layout.go` | Calculates box positions and dimensions |
| **6. Rendering** | `render/` | `render.go` | Draws to canvas, outputs PNG |

## Package Structure

```mermaid
graph TB
    subgraph cmd["cmd/"]
        browser["browser/main.go<br/>CLI Entry Point"]
        wasm["browser-wasm/main.go<br/>WASM Entry Point"]
        wptrunner["wptrunner/<br/>Test Runner"]
    end

    subgraph core["Core Packages"]
        dom["dom/<br/>DOM Tree Structure"]
        html["html/<br/>HTML Parsing"]
        css["css/<br/>CSS Parsing"]
        style["style/<br/>Style Computation"]
        layout["layout/<br/>Layout Engine"]
        render["render/<br/>Rendering"]
    end

    subgraph support["Support Packages"]
        svg["svg/<br/>SVG Parser & Rasterizer"]
        font["font/<br/>Font Embedding"]
        log["log/<br/>Logging"]
        reftest["reftest/<br/>Reference Tests"]
    end

    browser --> dom
    browser --> html
    browser --> css
    browser --> style
    browser --> layout
    browser --> render
    
    wasm --> html
    wasm --> css
    wasm --> style
    wasm --> layout
    wasm --> render

    html --> dom
    style --> dom
    style --> css
    layout --> style
    render --> layout
    render --> svg
    render --> font
```

## Data Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI as CLI/WASM
    participant HTML as HTML Parser
    participant CSS as CSS Parser
    participant Style as Style Engine
    participant Layout as Layout Engine
    participant Render as Renderer

    User->>CLI: Input (URL/File)
    CLI->>HTML: Raw HTML
    HTML->>HTML: Tokenize
    HTML->>CLI: DOM Tree
    
    CLI->>CSS: CSS Content
    CSS->>CSS: Tokenize & Parse
    CSS->>CLI: Stylesheet
    
    CLI->>Style: DOM + Stylesheet
    Style->>Style: Match selectors
    Style->>Style: Calculate specificity
    Style->>Style: Apply cascade
    Style->>CLI: Styled Tree
    
    CLI->>Layout: Styled Tree + Viewport
    Layout->>Layout: Calculate boxes
    Layout->>Layout: Position elements
    Layout->>CLI: Layout Tree
    
    CLI->>Render: Layout Tree
    Render->>Render: Draw backgrounds
    Render->>Render: Draw borders
    Render->>Render: Draw text
    Render->>Render: Draw images
    Render->>User: PNG Image
```

## Key Data Structures

### DOM Node (`dom/node.go`)
```go
type Node struct {
    Type       NodeType           // Document, Element, Text
    Data       string             // Tag name or text content
    Attributes map[string]string  // Element attributes
    Children   []*Node            // Child nodes
    Parent     *Node              // Parent reference
}
```

### CSS Rule (`css/parser.go`)
```go
type Rule struct {
    Selectors    []*Selector    // List of selectors
    Declarations []*Declaration // Property-value pairs
}
```

### Styled Node (`style/style.go`)
```go
type StyledNode struct {
    Node     *dom.Node              // DOM node reference
    Styles   map[string]string      // Computed styles
    Children []*StyledNode          // Styled children
}
```

### Layout Box (`layout/layout.go`)
```go
type LayoutBox struct {
    BoxType    BoxType      // Block, Inline, Table, etc.
    Dimensions Dimensions   // Position and size
    StyledNode *StyledNode  // Styled node reference
    Children   []*LayoutBox // Child boxes
}
```

## Specification Compliance

The browser implements portions of these W3C specifications:

| Specification | Coverage | Key Features |
|--------------|----------|--------------|
| **HTML5** | Partial | Tokenization, DOM tree construction, void elements |
| **CSS 2.1 §4** | Good | Syntax, tokenization, identifiers |
| **CSS 2.1 §5** | Good | Selectors (type, class, ID, descendant) |
| **CSS 2.1 §6** | Good | Cascade, specificity calculation |
| **CSS 2.1 §8** | Good | Box model (margin, padding, border) |
| **CSS 2.1 §9** | Partial | Block layout, basic inline layout |
| **CSS 2.1 §14** | Good | Colors and backgrounds |
| **CSS 2.1 §17** | Partial | Table layout (basic) |
| **RFC 2397** | Good | Data URLs (base64, URL-encoded) |

## Deployment Targets

```mermaid
graph TB
    subgraph Source["Source Code"]
        Go["Go Source<br/>(*.go)"]
    end

    subgraph Native["Native Build"]
        CLI["CLI Binary<br/>(go build)"]
    end

    subgraph Web["WebAssembly Build"]
        WASM["browser.wasm<br/>(GOOS=js GOARCH=wasm)"]
        Demo["wasm/index.html<br/>Interactive Demo"]
    end

    subgraph Deploy["Deployment"]
        Local["Local CLI Usage"]
        GHPages["GitHub Pages<br/>lukehoban.github.io/browser"]
    end

    Go --> CLI
    Go --> WASM
    WASM --> Demo
    CLI --> Local
    Demo --> GHPages
```

## See Also

- [MILESTONES.md](MILESTONES.md) - Implementation progress and feature tracking
- [IMPLEMENTATION.md](IMPLEMENTATION.md) - Detailed implementation notes
- [TESTING.md](TESTING.md) - Testing strategy and results
- [README.md](README.md) - Quick start guide
