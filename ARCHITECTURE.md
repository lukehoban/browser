# Architecture Diagram

This document provides a visual overview of the browser's architecture and rendering pipeline.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              INPUT SOURCES                                   │
│                                                                             │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                    │
│   │  Local File │    │   HTTP(S)   │    │    WASM     │                    │
│   │    .html    │    │     URL     │    │  (Browser)  │                    │
│   └──────┬──────┘    └──────┬──────┘    └──────┬──────┘                    │
│          │                  │                  │                            │
│          └──────────────────┴──────────────────┘                            │
│                             │                                               │
│                             ▼                                               │
│                    ┌─────────────────┐                                      │
│                    │   HTML Content  │                                      │
│                    └────────┬────────┘                                      │
└─────────────────────────────┼───────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         RENDERING PIPELINE                                   │
│                                                                             │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                         1. PARSING PHASE                               │ │
│  │                                                                        │ │
│  │   ┌─────────────────┐         ┌─────────────────┐                     │ │
│  │   │   HTML Parser   │         │   CSS Parser    │                     │ │
│  │   │   (html/)       │         │   (css/)        │                     │ │
│  │   │                 │         │                 │                     │ │
│  │   │  • Tokenizer    │         │  • Tokenizer    │                     │ │
│  │   │  • Parser       │         │  • Parser       │                     │ │
│  │   │  • Tree Builder │         │  • Values       │                     │ │
│  │   └────────┬────────┘         └────────┬────────┘                     │ │
│  │            │                           │                               │ │
│  │            ▼                           ▼                               │ │
│  │   ┌─────────────────┐         ┌─────────────────┐                     │ │
│  │   │    DOM Tree     │         │   Stylesheet    │                     │ │
│  │   │    (dom/)       │         │   (css.Rules)   │                     │ │
│  │   └─────────────────┘         └─────────────────┘                     │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                              │                                               │
│                              ▼                                               │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                     2. STYLE COMPUTATION                               │ │
│  │                        (style/)                                        │ │
│  │                                                                        │ │
│  │   DOM Tree + Stylesheet ──────▶ ┌─────────────────────────────┐       │ │
│  │                                 │      Style Engine           │       │ │
│  │   • Selector Matching           │  ┌───────────────────────┐  │       │ │
│  │   • Specificity Calculation     │  │ User-Agent Stylesheet │  │       │ │
│  │   • Cascade Resolution          │  │   (useragent.go)      │  │       │ │
│  │   • Inline Style Parsing        │  └───────────────────────┘  │       │ │
│  │   • Shorthand Expansion         └──────────────┬──────────────┘       │ │
│  │                                                │                       │ │
│  │                                                ▼                       │ │
│  │                                 ┌─────────────────────────────┐       │ │
│  │                                 │      Styled Tree            │       │ │
│  │                                 │   (DOM + Computed Styles)   │       │ │
│  │                                 └─────────────────────────────┘       │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                              │                                               │
│                              ▼                                               │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                       3. LAYOUT PHASE                                  │ │
│  │                        (layout/)                                       │ │
│  │                                                                        │ │
│  │   Styled Tree ──────────────────▶ ┌─────────────────────────────┐     │ │
│  │                                   │      Layout Engine          │     │ │
│  │   • Box Model Calculation         │                             │     │ │
│  │   • Block Formatting Context      │  ┌─────────────────────┐   │     │ │
│  │   • Inline Formatting Context     │  │  Box Model:         │   │     │ │
│  │   • Table Layout Algorithm        │  │  • Content          │   │     │ │
│  │   • Width/Height Resolution       │  │  • Padding          │   │     │ │
│  │   • Position Calculation          │  │  • Border           │   │     │ │
│  │                                   │  │  • Margin           │   │     │ │
│  │                                   │  └─────────────────────┘   │     │ │
│  │                                   └──────────────┬──────────────┘     │ │
│  │                                                  │                     │ │
│  │                                                  ▼                     │ │
│  │                                   ┌─────────────────────────────┐     │ │
│  │                                   │       Layout Tree           │     │ │
│  │                                   │   (Positioned Boxes)        │     │ │
│  │                                   └─────────────────────────────┘     │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
│                              │                                               │
│                              ▼                                               │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │                      4. RENDERING PHASE                                │ │
│  │                        (render/)                                       │ │
│  │                                                                        │ │
│  │   Layout Tree ──────────────────▶ ┌─────────────────────────────┐     │ │
│  │                                   │      Render Engine          │     │ │
│  │   • Background Rendering          │                             │     │ │
│  │   • Border Rendering              │  ┌─────────────────────┐   │     │ │
│  │   • Text Rendering (Go Fonts)     │  │  Canvas (Pixel      │   │     │ │
│  │   • Image Rendering               │  │  Buffer)            │   │     │ │
│  │   • SVG Rasterization             │  └─────────────────────┘   │     │ │
│  │   • Alpha Blending                │                             │     │ │
│  │                                   └──────────────┬──────────────┘     │ │
│  │                                                  │                     │ │
│  └──────────────────────────────────────────────────┼─────────────────────┘ │
└─────────────────────────────────────────────────────┼───────────────────────┘
                                                      │
                                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              OUTPUT                                          │
│                                                                             │
│   ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐        │
│   │   PNG Image     │    │  Layout Tree    │    │   Base64 PNG    │        │
│   │   (File)        │    │  (Terminal)     │    │   (WASM)        │        │
│   └─────────────────┘    └─────────────────┘    └─────────────────┘        │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Package Dependencies

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│                        ┌─────────────────────┐                              │
│                        │   cmd/browser/      │                              │
│                        │   (CLI Entry Point) │                              │
│                        └──────────┬──────────┘                              │
│                                   │                                          │
│                    ┌──────────────┼──────────────┐                          │
│                    │              │              │                          │
│                    ▼              ▼              ▼                          │
│          ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │
│          │   html/     │  │    css/     │  │   render/   │                 │
│          │  (Parser)   │  │  (Parser)   │  │  (Renderer) │                 │
│          └──────┬──────┘  └──────┬──────┘  └──────┬──────┘                 │
│                 │                │                │                         │
│                 │                │                │                         │
│                 ▼                ▼                ▼                         │
│          ┌─────────────┐  ┌─────────────┐  ┌─────────────┐                 │
│          │    dom/     │  │   style/    │  │   layout/   │                 │
│          │ (DOM Tree)  │  │  (Cascade)  │  │  (Box Model)│                 │
│          └──────┬──────┘  └──────┬──────┘  └──────┬──────┘                 │
│                 │                │                │                         │
│                 └────────────────┴────────────────┘                         │
│                                  │                                          │
│                                  ▼                                          │
│                        ┌─────────────────────┐                              │
│                        │        log/         │                              │
│                        │  (Logging Utility)  │                              │
│                        └─────────────────────┘                              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Data Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   HTML      │     │    DOM      │     │   Styled    │     │   Layout    │
│   String    │────▶│    Tree     │────▶│    Tree     │────▶│    Tree     │
│             │     │             │     │             │     │             │
│  "<html>    │     │  Document   │     │  StyledNode │     │  LayoutBox  │
│   <body>    │     │    └─HTML   │     │    + Styles │     │    + Dims   │
│   ..."      │     │       └─Body│     │    + Node   │     │    + Pos    │
└─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
       │                  │                   │                    │
       │                  │                   │                    │
  html.Parse()      style.StyleTree()   layout.LayoutTree()   render.Render()
                                                                   │
                                                                   ▼
                                                          ┌─────────────┐
                                                          │   Canvas    │
                                                          │   (PNG)     │
                                                          └─────────────┘
```

## Key Data Structures

### DOM Layer (`dom/`)
```go
type Node struct {
    Type       NodeType          // ElementNode, TextNode, DocumentNode
    Data       string            // Tag name or text content
    Attributes map[string]string // HTML attributes
    Children   []*Node           // Child nodes
    Parent     *Node             // Parent reference
}
```

### Style Layer (`style/`)
```go
type StyledNode struct {
    Node     *dom.Node           // Reference to DOM node
    Styles   map[string]string   // Computed CSS properties
    Children []*StyledNode       // Styled children
}
```

### Layout Layer (`layout/`)
```go
type LayoutBox struct {
    BoxType    BoxType      // Block, Inline, Table, etc.
    Dimensions Dimensions   // Position and size
    StyledNode *StyledNode  // Reference to styled node
    Children   []*LayoutBox // Child boxes
}

type Dimensions struct {
    Content Rect       // Content area (x, y, width, height)
    Padding EdgeSizes  // Padding on all 4 sides
    Border  EdgeSizes  // Border on all 4 sides
    Margin  EdgeSizes  // Margin on all 4 sides
}
```

### Render Layer (`render/`)
```go
type Canvas struct {
    Width      int                   // Canvas width
    Height     int                   // Canvas height
    Pixels     []color.RGBA          // Pixel buffer
    ImageCache map[string]image.Image // Cache for loaded images
}
```

## Specification Compliance

| Component | Specification |
|-----------|--------------|
| HTML Tokenizer | HTML5 §12.2.5 |
| HTML Parser | HTML5 §12.2.6 |
| CSS Tokenizer | CSS 2.1 §4.1 |
| CSS Parser | CSS 2.1 §4.1.7, §4.1.8 |
| Selectors | CSS 2.1 §5 |
| Cascade | CSS 2.1 §6.4 |
| Box Model | CSS 2.1 §8 |
| Visual Formatting | CSS 2.1 §9, §10 |
| Colors | CSS 2.1 §14 |
| Fonts | CSS 2.1 §15 |
| Text | CSS 2.1 §16 |
| Tables | CSS 2.1 §17 |
| Data URLs | RFC 2397 |
