# Architecture Diagram

This document provides a visual overview of the browser's architecture.

## What Does This Browser Do?

This is an **educational web browser implementation in Go** that:

1. **Parses HTML** documents into a DOM tree
2. **Parses CSS** stylesheets (inline and external)
3. **Computes styles** by matching CSS selectors to DOM elements
4. **Calculates layout** using the CSS box model
5. **Renders** the final output to a PNG image

It follows W3C specifications (HTML5 and CSS 2.1) and demonstrates fundamental web rendering concepts.

## High-Level Rendering Pipeline

```mermaid
flowchart LR
    subgraph Input
        HTML[HTML File/URL]
        ExtCSS[External CSS]
    end

    subgraph "HTML Parser"
        Tokenizer1["html/tokenizer.go"]
        Parser1["html/parser.go"]
    end

    subgraph "CSS Parser"
        Tokenizer2["css/tokenizer.go"]
        Parser2["css/parser.go"]
    end

    subgraph "DOM"
        DOMTree["dom/node.go<br/>DOM Tree"]
        URLResolver["dom/url.go<br/>URL Resolution"]
        Loader["dom/loader.go<br/>Resource Loader"]
    end

    subgraph "Style Engine"
        StyleComp["style/style.go<br/>Selector Matching<br/>Specificity<br/>Cascade"]
        UASheet["style/useragent.go<br/>Default Styles"]
    end

    subgraph "Layout Engine"
        LayoutCalc["layout/layout.go<br/>Box Model<br/>Block Layout<br/>Inline Layout<br/>Table Layout"]
    end

    subgraph "Render Engine"
        Renderer["render/render.go<br/>Canvas<br/>Drawing<br/>Text & Images"]
        FontPkg["font/<br/>Go Fonts"]
    end

    subgraph Output
        PNG[PNG Image]
    end

    HTML --> Tokenizer1 --> Parser1 --> DOMTree
    ExtCSS --> Tokenizer2
    DOMTree --> URLResolver --> StyleComp
    Parser2 --> StyleComp
    UASheet --> StyleComp
    StyleComp --> LayoutCalc
    LayoutCalc --> Renderer
    FontPkg --> Renderer
    Loader --> Renderer
    Renderer --> PNG
```

## Data Flow Diagram

```mermaid
flowchart TB
    subgraph "1. Input Stage"
        Input["HTML Input<br/>(file or URL)"]
    end

    subgraph "2. Parsing Stage"
        HTMLParse["HTML Parsing<br/>━━━━━━━━━━━━<br/>• Tokenization<br/>• Tree Construction"]
        CSSParse["CSS Parsing<br/>━━━━━━━━━━━━<br/>• Tokenization<br/>• Selectors<br/>• Declarations"]
    end

    subgraph "3. DOM Stage"
        DOMTree2["DOM Tree<br/>━━━━━━━━━━━━<br/>• Element Nodes<br/>• Text Nodes<br/>• Attributes"]
        URLRes["URL Resolution<br/>━━━━━━━━━━━━<br/>• Relative → Absolute<br/>• Image paths<br/>• Stylesheet links"]
    end

    subgraph "4. Style Stage"
        Stylesheet["Stylesheet<br/>━━━━━━━━━━━━<br/>• User-agent styles<br/>• Author styles<br/>• Inline styles"]
        StyledTree["Styled Tree<br/>━━━━━━━━━━━━<br/>• Matched rules<br/>• Computed styles<br/>• Inheritance"]
    end

    subgraph "5. Layout Stage"
        LayoutTree["Layout Tree<br/>━━━━━━━━━━━━<br/>• Box dimensions<br/>• Positions (x, y)<br/>• Margin/Padding/Border"]
    end

    subgraph "6. Render Stage"
        Canvas["Canvas<br/>━━━━━━━━━━━━<br/>• Background colors<br/>• Borders<br/>• Text rendering<br/>• Image rendering"]
    end

    subgraph "7. Output Stage"
        Output["PNG Image"]
    end

    Input --> HTMLParse
    HTMLParse --> DOMTree2
    DOMTree2 --> URLRes
    URLRes --> StyledTree
    HTMLParse -.->|"extract &lt;style&gt;"| CSSParse
    URLRes -.->|"fetch external CSS"| CSSParse
    CSSParse --> Stylesheet
    Stylesheet --> StyledTree
    StyledTree --> LayoutTree
    LayoutTree --> Canvas
    Canvas --> Output
```

## Package Dependencies

```mermaid
graph TB
    subgraph "Entry Point"
        CMD["cmd/browser/main.go"]
        WASM["cmd/browser-wasm/main.go"]
    end

    subgraph "Core Packages"
        HTML["html/<br/>HTML Parser"]
        CSS["css/<br/>CSS Parser"]
        DOM["dom/<br/>DOM Tree"]
        STYLE["style/<br/>Style Engine"]
        LAYOUT["layout/<br/>Layout Engine"]
        RENDER["render/<br/>Renderer"]
        FONT["font/<br/>Font Loader"]
        LOG["log/<br/>Logging"]
    end

    subgraph "Testing"
        REFTEST["reftest/<br/>WPT Runner"]
    end

    CMD --> HTML
    CMD --> CSS
    CMD --> DOM
    CMD --> STYLE
    CMD --> LAYOUT
    CMD --> RENDER
    CMD --> LOG

    WASM --> HTML
    WASM --> CSS
    WASM --> DOM
    WASM --> STYLE
    WASM --> LAYOUT
    WASM --> RENDER

    HTML --> DOM
    STYLE --> DOM
    STYLE --> CSS
    LAYOUT --> STYLE
    LAYOUT --> DOM
    LAYOUT --> CSS
    LAYOUT --> FONT
    RENDER --> LAYOUT
    RENDER --> STYLE
    RENDER --> DOM
    RENDER --> FONT

    REFTEST --> HTML
    REFTEST --> CSS
    REFTEST --> DOM
    REFTEST --> STYLE
    REFTEST --> LAYOUT
    REFTEST --> RENDER
```

## Key Data Structures

```mermaid
classDiagram
    class Node {
        +NodeType Type
        +string Data
        +map Attributes
        +[]Node Children
        +Node Parent
    }

    class StyledNode {
        +Node Node
        +map Styles
        +[]StyledNode Children
    }

    class LayoutBox {
        +BoxType BoxType
        +StyledNode StyledNode
        +Dimensions Dimensions
        +[]LayoutBox Children
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

    class Canvas {
        +Image Image
        +int Width
        +int Height
    }

    Node --> StyledNode : styled into
    StyledNode --> LayoutBox : laid out into
    LayoutBox --> Dimensions : has
    Dimensions --> Rect : contains
    LayoutBox --> Canvas : rendered to
```

## Processing Steps

```mermaid
sequenceDiagram
    participant Main as main.go
    participant HTML as html/parser
    participant DOM as dom/node
    participant CSS as css/parser
    participant Style as style/style
    participant Layout as layout/layout
    participant Render as render/render

    Main->>HTML: Parse(htmlContent)
    HTML->>DOM: Build DOM tree
    DOM-->>Main: *dom.Node

    Main->>DOM: ResolveURLs(doc, baseURL)
    Main->>DOM: FetchExternalStylesheets(doc)

    Main->>CSS: Parse(cssContent)
    CSS-->>Main: *css.Stylesheet

    Main->>Style: StyleTree(doc, stylesheet)
    Style->>Style: Match selectors
    Style->>Style: Calculate specificity
    Style->>Style: Apply cascade
    Style-->>Main: *style.StyledNode

    Main->>Layout: LayoutTree(styledTree, viewport)
    Layout->>Layout: Calculate box dimensions
    Layout->>Layout: Position elements
    Layout-->>Main: *layout.LayoutBox

    Main->>Render: Render(layoutTree, width, height)
    Render->>Render: Draw backgrounds
    Render->>Render: Draw borders
    Render->>Render: Draw text
    Render->>Render: Draw images
    Render-->>Main: *render.Canvas

    Main->>Render: canvas.SavePNG(outputFile)
```

## Feature Summary

| Component | Features Implemented | Not Yet Implemented |
|-----------|---------------------|---------------------|
| **HTML Parser** | Tokenization, tree construction, void elements, attributes | Character references, namespaces, script execution |
| **CSS Parser** | Selectors (element, class, ID), descendant combinators, declarations | Pseudo-classes, pseudo-elements, attribute selectors |
| **Style Engine** | Selector matching, specificity, cascade, inheritance | !important, computed values |
| **Layout Engine** | Box model, block layout, inline layout, tables | Floats, positioning, flexbox, grid |
| **Render Engine** | Backgrounds, borders, text, images, SVG | Background images (CSS), gradients, transforms |

## WebAssembly Support

The browser can also run in a web browser via WebAssembly:

```mermaid
flowchart LR
    subgraph "Browser (Go → WASM)"
        WASM_Entry["cmd/browser-wasm/"]
        CorePkgs["Core Packages<br/>(html, css, dom,<br/>style, layout, render)"]
    end

    subgraph "Web Browser"
        JS["JavaScript<br/>Interface"]
        HTMLCanvas["HTML Canvas<br/>Element"]
    end

    WASM_Entry --> CorePkgs
    CorePkgs --> JS
    JS --> HTMLCanvas
```

Live demo: https://lukehoban.github.io/browser/
