# Browser Architecture

A web browser rendering engine written in Go that converts HTML/CSS to PNG images.

## High-Level Architecture

```mermaid
flowchart TB
    subgraph Input
        HTML[HTML File/URL]
        CSS_EXT[External CSS]
    end

    subgraph "Parsing Layer"
        HTML_TOK[HTML Tokenizer]
        HTML_PARSE[HTML Parser]
        CSS_TOK[CSS Tokenizer]
        CSS_PARSE[CSS Parser]
    end

    subgraph "Data Structures"
        DOM[(DOM Tree)]
        SS[(Stylesheet)]
        ST[(Styled Tree)]
        LT[(Layout Tree)]
    end

    subgraph "Processing Layer"
        STYLE[Style Engine]
        LAYOUT[Layout Engine]
    end

    subgraph "Rendering Layer"
        RENDER[Render Engine]
        CANVAS[Canvas]
    end

    subgraph "Support Modules"
        FONT[Font Module]
        SVG[SVG Module]
        LOADER[DOM Loader]
    end

    subgraph Output
        PNG[PNG Image]
    end

    HTML --> HTML_TOK --> HTML_PARSE --> DOM
    HTML_PARSE --> LOADER
    LOADER --> CSS_EXT
    CSS_EXT --> CSS_TOK
    DOM --> CSS_TOK --> CSS_PARSE --> SS

    DOM --> STYLE
    SS --> STYLE
    STYLE --> ST

    ST --> LAYOUT
    FONT -.->|text measurement| LAYOUT
    LAYOUT --> LT

    LT --> RENDER
    FONT -.->|text drawing| RENDER
    SVG -.->|image rasterization| RENDER
    LOADER -.->|image loading| RENDER
    RENDER --> CANVAS --> PNG
```

## Rendering Pipeline

```mermaid
flowchart LR
    subgraph "1. Parse"
        A[HTML Input] --> B[Tokenize]
        B --> C[Build DOM]
    end

    subgraph "2. Style"
        C --> D[Extract CSS]
        D --> E[Parse CSS]
        E --> F[Match Selectors]
        F --> G[Compute Cascade]
    end

    subgraph "3. Layout"
        G --> H[Generate Boxes]
        H --> I[Calculate Dimensions]
        I --> J[Position Elements]
    end

    subgraph "4. Render"
        J --> K[Draw Backgrounds]
        K --> L[Draw Borders]
        L --> M[Draw Text/Images]
        M --> N[PNG Output]
    end
```

## Module Dependencies

```mermaid
graph TD
    CMD[cmd/browser] --> DOM
    CMD --> HTML
    CMD --> CSS
    CMD --> STYLE
    CMD --> LAYOUT
    CMD --> RENDER

    HTML[html/] --> DOM[dom/]
    CSS[css/] --> DOM
    STYLE[style/] --> DOM
    STYLE --> CSS
    LAYOUT[layout/] --> STYLE
    LAYOUT --> FONT[font/]
    RENDER[render/] --> LAYOUT
    RENDER --> FONT
    RENDER --> SVG[svg/]
    RENDER --> DOM

    DOM --> LOG[log/]
    HTML --> LOG
    CSS --> LOG
    STYLE --> LOG
    LAYOUT --> LOG
    RENDER --> LOG
```

## Data Structure Transformations

```mermaid
flowchart TB
    subgraph "DOM Node"
        DN_TYPE[Type: Element/Text]
        DN_DATA[Data: tag name]
        DN_ATTR[Attributes: map]
        DN_CHILD[Children: nodes]
    end

    subgraph "Styled Node"
        SN_NODE[Node: *dom.Node]
        SN_STYLES[Styles: computed CSS]
        SN_CHILD[Children: styled nodes]
    end

    subgraph "Layout Box"
        LB_TYPE[BoxType: block/inline]
        LB_STYLED[StyledNode: *style.StyledNode]
        LB_DIM[Dimensions: position + size]
        LB_CHILD[Children: layout boxes]
    end

    subgraph "Canvas"
        CV_SIZE[Width x Height]
        CV_PIX[Pixels: RGBA array]
        CV_CACHE[ImageCache: loaded images]
    end

    DN_TYPE --> SN_NODE
    SN_NODE --> LB_STYLED
    LB_DIM --> CV_PIX
```

## CSS Cascade & Specificity

```mermaid
flowchart TB
    subgraph "Style Resolution"
        SEL[Selector Matching]
        SPEC[Specificity Calculation]
        CASCADE[Cascade Ordering]
        INHERIT[Inheritance]
        COMPUTE[Computed Value]
    end

    RULE1[CSS Rule 1] --> SEL
    RULE2[CSS Rule 2] --> SEL
    RULE3[CSS Rule 3] --> SEL
    UA[User Agent Styles] --> SEL

    SEL --> SPEC
    SPEC --> CASCADE
    CASCADE --> INHERIT
    INHERIT --> COMPUTE

    COMPUTE --> STYLED[Styled Node]
```

## Entry Points

```mermaid
flowchart TB
    subgraph "CLI Application"
        MAIN[cmd/browser/main.go]
        FLAGS[Parse Flags]
        ORCH[Orchestrate Pipeline]
    end

    subgraph "WASM Application"
        WASM[cmd/browser-wasm/main.go]
        RENDER_HTML[renderHTML]
        GET_LAYOUT[getLayoutTree]
        GET_RENDER[getRenderTree]
    end

    MAIN --> FLAGS --> ORCH
    WASM --> RENDER_HTML
    WASM --> GET_LAYOUT
    WASM --> GET_RENDER

    ORCH --> PNG[PNG File]
    RENDER_HTML --> BASE64[Base64 PNG]
```

## Key Modules Summary

| Module | Responsibility |
|--------|----------------|
| `html/` | HTML tokenization and DOM tree construction |
| `css/` | CSS tokenization and stylesheet parsing |
| `dom/` | DOM node structure, URL resolution, resource loading |
| `style/` | Selector matching, specificity, cascade, inheritance |
| `layout/` | CSS box model, dimension calculation, positioning |
| `render/` | Drawing to canvas (backgrounds, borders, text, images) |
| `font/` | TrueType font loading, text measurement |
| `svg/` | SVG parsing and rasterization |
| `log/` | Debug logging with configurable levels |
