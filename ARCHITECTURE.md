# Browser Architecture

This document provides a visual overview of the browser's architecture and rendering pipeline.

## High-Level Architecture

```mermaid
flowchart TB
    subgraph Input["Input Sources"]
        URL["HTTP/HTTPS URL"]
        FILE["Local HTML File"]
    end

    subgraph Parsing["Parsing Layer"]
        HTML["HTML Parser<br/><code>html/</code>"]
        CSS["CSS Parser<br/><code>css/</code>"]
    end

    subgraph DataStructures["Data Structures"]
        DOM["DOM Tree<br/><code>dom/</code>"]
        CSSOM["Stylesheet<br/>(Rules + Selectors)"]
    end

    subgraph Processing["Processing Layer"]
        STYLE["Style Engine<br/><code>style/</code>"]
        LAYOUT["Layout Engine<br/><code>layout/</code>"]
    end

    subgraph Trees["Intermediate Trees"]
        STYLEDTREE["Styled Tree<br/>(DOM + Computed Styles)"]
        LAYOUTTREE["Layout Tree<br/>(Box Model + Positions)"]
    end

    subgraph Output["Output Layer"]
        RENDER["Render Engine<br/><code>render/</code>"]
        PNG["PNG Image"]
        WASM["WebAssembly Canvas"]
    end

    URL --> HTML
    FILE --> HTML
    HTML --> DOM
    HTML -->|"&lt;style&gt; tags<br/>&lt;link&gt; stylesheets"| CSS
    CSS --> CSSOM
    DOM --> STYLE
    CSSOM --> STYLE
    STYLE --> STYLEDTREE
    STYLEDTREE --> LAYOUT
    LAYOUT --> LAYOUTTREE
    LAYOUTTREE --> RENDER
    RENDER --> PNG
    RENDER --> WASM
```

## Detailed Component Architecture

```mermaid
flowchart LR
    subgraph HTMLParsing["HTML Parsing Pipeline"]
        direction TB
        HTOK["Tokenizer<br/><code>html/tokenizer.go</code>"]
        HPARSE["Parser<br/><code>html/parser.go</code>"]
        HTOK -->|"Tokens"| HPARSE
    end

    subgraph CSSParsing["CSS Parsing Pipeline"]
        direction TB
        CTOK["Tokenizer<br/><code>css/tokenizer.go</code>"]
        CPARSE["Parser<br/><code>css/parser.go</code>"]
        CVAL["Values<br/><code>css/values.go</code>"]
        CTOK -->|"Tokens"| CPARSE
        CPARSE --> CVAL
    end

    subgraph DOMLayer["DOM Layer"]
        direction TB
        NODE["Node Structure<br/><code>dom/node.go</code>"]
        URLRES["URL Resolution<br/><code>dom/url.go</code>"]
        LOADER["Resource Loader<br/><code>dom/loader.go</code>"]
        NODE --> URLRES
        URLRES --> LOADER
    end

    subgraph StyleLayer["Style Computation"]
        direction TB
        MATCH["Selector Matching"]
        SPEC["Specificity Calculation"]
        CASCADE["Cascade Resolution"]
        UA["User Agent Styles<br/><code>style/useragent.go</code>"]
        MATCH --> SPEC
        SPEC --> CASCADE
        UA --> CASCADE
    end

    subgraph LayoutLayer["Layout Computation"]
        direction TB
        BOXMODEL["Box Model"]
        BLOCKL["Block Layout"]
        INLINEL["Inline Layout"]
        TABLEL["Table Layout"]
        BOXMODEL --> BLOCKL
        BOXMODEL --> INLINEL
        BOXMODEL --> TABLEL
    end

    subgraph RenderLayer["Rendering"]
        direction TB
        CANVAS["Canvas"]
        BGCOLOR["Background Colors"]
        BORDERS["Borders"]
        TEXT["Text Rendering"]
        IMAGES["Image Rendering"]
        CANVAS --> BGCOLOR
        CANVAS --> BORDERS
        CANVAS --> TEXT
        CANVAS --> IMAGES
    end

    HTMLParsing --> DOMLayer
    CSSParsing --> StyleLayer
    DOMLayer --> StyleLayer
    StyleLayer --> LayoutLayer
    LayoutLayer --> RenderLayer
```

## Data Flow Diagram

```mermaid
sequenceDiagram
    participant User
    participant Main as main.go
    participant HTML as html/parser.go
    participant DOM as dom/node.go
    participant CSS as css/parser.go
    participant Style as style/style.go
    participant Layout as layout/layout.go
    participant Render as render/render.go

    User->>Main: browser input.html -output out.png
    Main->>HTML: Parse(htmlContent)
    HTML->>DOM: Build DOM Tree
    DOM-->>Main: *dom.Node (document)
    
    Main->>DOM: ResolveURLs(doc, baseURL)
    Main->>DOM: FetchExternalStylesheets(doc)
    
    Main->>CSS: Parse(cssContent)
    CSS-->>Main: *Stylesheet (rules)
    
    Main->>Style: StyleTree(doc, stylesheet)
    Style->>Style: Match selectors to elements
    Style->>Style: Calculate specificity
    Style->>Style: Apply cascade
    Style-->>Main: *StyledNode (styled tree)
    
    Main->>Layout: LayoutTree(styledTree, viewport)
    Layout->>Layout: Calculate box dimensions
    Layout->>Layout: Position boxes
    Layout-->>Main: *LayoutBox (layout tree)
    
    Main->>Render: Render(layoutTree, width, height)
    Render->>Render: Draw backgrounds
    Render->>Render: Draw borders
    Render->>Render: Draw text
    Render->>Render: Draw images
    Render-->>Main: *Canvas
    
    Main->>Render: canvas.SavePNG(outputFile)
    Render-->>User: PNG file saved
```

## Key Data Structures

```mermaid
classDiagram
    class Node {
        +NodeType Type
        +string Data
        +[]Attribute Attributes
        +[]*Node Children
        +*Node Parent
    }
    
    class StyledNode {
        +*Node Node
        +map[string]string Styles
        +[]*StyledNode Children
    }
    
    class LayoutBox {
        +BoxType BoxType
        +Dimensions Dimensions
        +*StyledNode StyledNode
        +[]*LayoutBox Children
    }
    
    class Dimensions {
        +Rect Content
        +EdgeSize Padding
        +EdgeSize Border
        +EdgeSize Margin
    }
    
    class Rect {
        +float64 X
        +float64 Y
        +float64 Width
        +float64 Height
    }
    
    class Stylesheet {
        +[]Rule Rules
    }
    
    class Rule {
        +[]Selector Selectors
        +[]Declaration Declarations
    }
    
    class Canvas {
        +*image.RGBA Image
        +int Width
        +int Height
    }
    
    Node --> StyledNode : styled by
    StyledNode --> LayoutBox : laid out as
    LayoutBox --> Dimensions : has
    Dimensions --> Rect : contains
    Stylesheet --> Rule : contains
    LayoutBox --> Canvas : rendered to
```

## Module Dependencies

```mermaid
flowchart BT
    subgraph Core["Core Packages"]
        DOM["dom"]
        LOG["log"]
        FONT["font"]
    end

    subgraph Parsing["Parsing Packages"]
        HTML["html"]
        CSS["css"]
    end

    subgraph Processing["Processing Packages"]
        STYLE["style"]
        LAYOUT["layout"]
    end

    subgraph Output["Output Packages"]
        RENDER["render"]
        SVG["svg"]
        REFTEST["reftest"]
    end

    subgraph Apps["Applications"]
        BROWSER["cmd/browser"]
        WASM["cmd/browser-wasm"]
    end

    HTML --> DOM
    CSS --> DOM
    STYLE --> DOM
    STYLE --> CSS
    LAYOUT --> STYLE
    RENDER --> LAYOUT
    RENDER --> DOM
    RENDER --> FONT
    RENDER --> SVG
    REFTEST --> RENDER
    REFTEST --> HTML
    REFTEST --> CSS
    REFTEST --> STYLE
    REFTEST --> LAYOUT
    
    BROWSER --> HTML
    BROWSER --> CSS
    BROWSER --> DOM
    BROWSER --> STYLE
    BROWSER --> LAYOUT
    BROWSER --> RENDER
    BROWSER --> LOG
    
    WASM --> HTML
    WASM --> CSS
    WASM --> DOM
    WASM --> STYLE
    WASM --> LAYOUT
    WASM --> RENDER
```

## Rendering Pipeline Summary

| Stage | Input | Output | Key Files |
|-------|-------|--------|-----------|
| **1. Fetch** | URL or file path | HTML string | `cmd/browser/main.go` |
| **2. Parse HTML** | HTML string | DOM tree | `html/tokenizer.go`, `html/parser.go` |
| **3. Parse CSS** | CSS string | Stylesheet | `css/tokenizer.go`, `css/parser.go` |
| **4. Style** | DOM + Stylesheet | Styled tree | `style/style.go` |
| **5. Layout** | Styled tree + viewport | Layout tree | `layout/layout.go` |
| **6. Render** | Layout tree | PNG image | `render/render.go` |

## WebAssembly Architecture

```mermaid
flowchart LR
    subgraph Browser["Web Browser"]
        JS["JavaScript"]
        WASMRT["WASM Runtime"]
        CANVAS2D["HTML5 Canvas"]
    end

    subgraph WASMModule["Go WASM Module"]
        ENTRY["browser-wasm/main.go"]
        PIPELINE["Rendering Pipeline<br/>(same as CLI)"]
    end

    JS -->|"renderURL(url)"| WASMRT
    WASMRT --> ENTRY
    ENTRY --> PIPELINE
    PIPELINE -->|"ImageData"| WASMRT
    WASMRT -->|"putImageData()"| CANVAS2D
```

## What This Browser Does

This is a **simple, educational web browser** written in Go that:

1. **Parses HTML** into a DOM tree following HTML5 tokenization specifications
2. **Parses CSS** stylesheets (inline `<style>` tags and external `<link>` stylesheets)
3. **Computes styles** by matching CSS selectors to DOM elements with proper specificity
4. **Calculates layout** using the CSS 2.1 box model and visual formatting model
5. **Renders** the final result to a PNG image

### Supported Features

- ✅ Block and inline layout
- ✅ CSS box model (margin, padding, border)
- ✅ Colors (named, hex)
- ✅ Fonts (Go fonts with bold, italic, size)
- ✅ Images (PNG, JPEG, GIF, SVG)
- ✅ Network fetching (HTTP/HTTPS)
- ✅ Data URLs (RFC 2397)
- ✅ WebAssembly compilation

### Not Supported

- ❌ JavaScript execution
- ❌ CSS floats and positioning
- ❌ CSS animations/transitions
- ❌ Web APIs (fetch, localStorage, etc.)
- ❌ User interaction (clicking, typing)

This browser is designed for **educational purposes** to demonstrate how web browsers work internally, following W3C specifications closely.
