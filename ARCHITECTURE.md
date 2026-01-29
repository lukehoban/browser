# Architecture Diagram

This document provides visual architecture diagrams for the browser implementation.

## High-Level Architecture

The browser follows a classic web rendering pipeline, transforming HTML/CSS input into rendered PNG output.

```mermaid
flowchart TB
    subgraph Input["📥 Input Sources"]
        URL["HTTP/HTTPS URL"]
        FILE["Local HTML File"]
    end

    subgraph Parsing["🔍 Parsing Layer"]
        HTML_TOKENIZER["HTML Tokenizer<br/><i>html/tokenizer.go</i>"]
        HTML_PARSER["HTML Parser<br/><i>html/parser.go</i>"]
        CSS_TOKENIZER["CSS Tokenizer<br/><i>css/tokenizer.go</i>"]
        CSS_PARSER["CSS Parser<br/><i>css/parser.go</i>"]
    end

    subgraph DataStructures["📊 Data Structures"]
        DOM["DOM Tree<br/><i>dom/node.go</i>"]
        STYLESHEET["Stylesheet<br/>(Rules + Selectors)"]
    end

    subgraph StyleComputation["🎨 Style Computation"]
        STYLE_ENGINE["Style Engine<br/><i>style/style.go</i>"]
        STYLED_TREE["Styled Tree<br/>(DOM + Computed Styles)"]
    end

    subgraph LayoutEngine["📐 Layout Engine"]
        LAYOUT["Layout Calculator<br/><i>layout/layout.go</i>"]
        LAYOUT_TREE["Layout Tree<br/>(Boxes + Dimensions)"]
    end

    subgraph RenderEngine["🖼️ Render Engine"]
        RENDERER["Renderer<br/><i>render/render.go</i>"]
        CANVAS["Canvas<br/>(Pixel Buffer)"]
    end

    subgraph Output["📤 Output"]
        PNG["PNG Image"]
        WASM["WebAssembly<br/>(Browser Canvas)"]
    end

    %% Input flow
    URL --> HTML_TOKENIZER
    FILE --> HTML_TOKENIZER
    
    %% HTML parsing
    HTML_TOKENIZER --> HTML_PARSER
    HTML_PARSER --> DOM
    
    %% CSS extraction and parsing
    DOM -->|"Extract &lt;style&gt; tags<br/>& &lt;link&gt; stylesheets"| CSS_TOKENIZER
    CSS_TOKENIZER --> CSS_PARSER
    CSS_PARSER --> STYLESHEET
    
    %% Style computation
    DOM --> STYLE_ENGINE
    STYLESHEET --> STYLE_ENGINE
    STYLE_ENGINE --> STYLED_TREE
    
    %% Layout
    STYLED_TREE --> LAYOUT
    LAYOUT --> LAYOUT_TREE
    
    %% Rendering
    LAYOUT_TREE --> RENDERER
    RENDERER --> CANVAS
    
    %% Output
    CANVAS --> PNG
    CANVAS --> WASM
```

## Component Details

### Parsing Pipeline

```mermaid
flowchart LR
    subgraph HTML["HTML Processing"]
        direction TB
        HT["Tokenizer<br/>State Machine"]
        HP["Parser<br/>Tree Builder"]
        HT --> HP
    end

    subgraph CSS["CSS Processing"]
        direction TB
        CT["Tokenizer<br/>Lexical Analysis"]
        CP["Parser<br/>Rule Parser"]
        CT --> CP
    end

    INPUT["HTML Input"] --> HTML
    HTML --> DOM_TREE["DOM Tree"]
    
    DOM_TREE -->|"Extract CSS"| CSS
    CSS --> RULES["CSS Rules"]
```

### Style Cascade & Computation

```mermaid
flowchart TB
    subgraph Inputs
        DOM["DOM Tree"]
        UA["User Agent Styles<br/><i>style/useragent.go</i>"]
        CSS["Author Stylesheets"]
        INLINE["Inline Styles<br/>(style attribute)"]
    end

    subgraph Computation["Style Computation"]
        MATCH["Selector Matching"]
        SPEC["Specificity Calculation<br/>(ID, Class, Type)"]
        CASCADE["Cascade Resolution"]
        INHERIT["Inheritance"]
    end

    subgraph Output
        STYLED["Styled Tree<br/>(Node + Computed Styles)"]
    end

    DOM --> MATCH
    CSS --> MATCH
    UA --> CASCADE
    INLINE --> CASCADE
    MATCH --> SPEC
    SPEC --> CASCADE
    CASCADE --> INHERIT
    INHERIT --> STYLED
```

### Box Model & Layout

```mermaid
flowchart TB
    subgraph BoxModel["CSS Box Model"]
        direction TB
        MARGIN["Margin"]
        BORDER["Border"]
        PADDING["Padding"]
        CONTENT["Content"]
        
        MARGIN --> BORDER
        BORDER --> PADDING
        PADDING --> CONTENT
    end

    subgraph LayoutTypes["Layout Types"]
        BLOCK["Block Layout<br/>(Vertical Stack)"]
        INLINE["Inline Layout<br/>(Horizontal Flow)"]
        TABLE["Table Layout<br/>(Grid)"]
    end

    subgraph Calculations["Dimension Calculations"]
        WIDTH["Width Calculation<br/>CSS 2.1 §10.3"]
        HEIGHT["Height Calculation<br/>CSS 2.1 §10.6"]
        POSITION["Position Calculation<br/>Normal Flow"]
    end

    BoxModel --> Calculations
    LayoutTypes --> Calculations
    Calculations --> LAYOUT_TREE["Layout Tree"]
```

### Rendering Pipeline

```mermaid
flowchart TB
    subgraph LayoutTree["Layout Tree"]
        BOX1["LayoutBox 1"]
        BOX2["LayoutBox 2"]
        BOX3["LayoutBox N..."]
    end

    subgraph RenderOps["Render Operations"]
        BG["Background Colors"]
        BORDERS["Borders"]
        TEXT["Text Rendering<br/>(Go Fonts)"]
        IMAGES["Images<br/>(PNG, JPEG, GIF, SVG)"]
    end

    subgraph Canvas["Canvas"]
        PIXELS["Pixel Buffer"]
    end

    subgraph Output["Output Formats"]
        PNG["PNG File"]
        WASM_CANVAS["WASM Canvas<br/>(Browser)"]
    end

    LayoutTree --> RenderOps
    RenderOps --> Canvas
    Canvas --> Output
```

## Package Dependencies

```mermaid
flowchart TB
    CMD["cmd/browser<br/><i>Main Entry Point</i>"]
    
    HTML["html<br/><i>HTML Parsing</i>"]
    CSS["css<br/><i>CSS Parsing</i>"]
    DOM["dom<br/><i>DOM Tree</i>"]
    STYLE["style<br/><i>Style Computation</i>"]
    LAYOUT["layout<br/><i>Layout Engine</i>"]
    RENDER["render<br/><i>Rendering</i>"]
    FONT["font<br/><i>Font Management</i>"]
    LOG["log<br/><i>Logging</i>"]
    SVG["svg<br/><i>SVG Rendering</i>"]

    CMD --> HTML
    CMD --> CSS
    CMD --> DOM
    CMD --> STYLE
    CMD --> LAYOUT
    CMD --> RENDER
    CMD --> LOG

    HTML --> DOM
    STYLE --> DOM
    STYLE --> CSS
    LAYOUT --> STYLE
    RENDER --> LAYOUT
    RENDER --> FONT
    RENDER --> DOM
    RENDER --> SVG
```

## Data Flow Summary

| Stage | Input | Output | Key Files |
|-------|-------|--------|-----------|
| **1. Fetch** | URL or File Path | HTML String | `cmd/browser/main.go` |
| **2. HTML Parse** | HTML String | DOM Tree | `html/tokenizer.go`, `html/parser.go` |
| **3. CSS Parse** | CSS String | Stylesheet | `css/tokenizer.go`, `css/parser.go` |
| **4. Style** | DOM + Stylesheet | Styled Tree | `style/style.go` |
| **5. Layout** | Styled Tree | Layout Tree | `layout/layout.go` |
| **6. Render** | Layout Tree | PNG Image | `render/render.go` |

## Specifications Implemented

- **HTML5** - Tokenization and parsing (§12)
- **CSS 2.1** - Syntax, selectors, cascade, box model, visual formatting
- **RFC 2397** - Data URL scheme
