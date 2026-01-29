# Browser Architecture

This document describes the architecture of the Go-based web browser implementation.

## Overview

This is a lightweight web browser that renders HTML and CSS to PNG images. It implements the classic web rendering pipeline following W3C specifications (HTML5, CSS 2.1).

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              INPUT LAYER                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐                    │
│   │  HTML File  │    │  HTTP URL   │    │ Data URL    │                    │
│   └──────┬──────┘    └──────┬──────┘    └──────┬──────┘                    │
│          │                  │                  │                            │
│          └──────────────────┼──────────────────┘                            │
│                             ▼                                               │
│                    ┌────────────────┐                                       │
│                    │  dom/loader.go │  Resource Loader                      │
│                    │  (HTTP, File,  │  (RFC 2397 Data URLs)                 │
│                    │   Data URLs)   │                                       │
│                    └────────┬───────┘                                       │
│                             │                                               │
└─────────────────────────────┼───────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            PARSING LAYER                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────────────────────────┐   ┌─────────────────────────────────┐│
│   │         html/ Module            │   │         css/ Module             ││
│   │                                 │   │                                 ││
│   │  ┌───────────────────────────┐  │   │  ┌───────────────────────────┐  ││
│   │  │     tokenizer.go          │  │   │  │     tokenizer.go          │  ││
│   │  │  (HTML5 §12.2.5)          │  │   │  │  (CSS 2.1 §4)             │  ││
│   │  │  State machine tokenizer  │  │   │  │  CSS token stream         │  ││
│   │  └───────────┬───────────────┘  │   │  └───────────┬───────────────┘  ││
│   │              │                  │   │              │                  ││
│   │              ▼                  │   │              ▼                  ││
│   │  ┌───────────────────────────┐  │   │  ┌───────────────────────────┐  ││
│   │  │       parser.go           │  │   │  │       parser.go           │  ││
│   │  │  (HTML5 §12.2.6)          │  │   │  │  Selectors & Declarations │  ││
│   │  │  Tree construction        │  │   │  │  Rules parsing            │  ││
│   │  └───────────┬───────────────┘  │   │  └───────────┬───────────────┘  ││
│   │              │                  │   │              │                  ││
│   └──────────────┼──────────────────┘   └──────────────┼──────────────────┘│
│                  │                                     │                   │
│                  ▼                                     ▼                   │
│         ┌────────────────┐                    ┌────────────────┐           │
│         │    DOM Tree    │                    │   Stylesheet   │           │
│         │   (dom/node)   │                    │  (css/parser)  │           │
│         └────────┬───────┘                    └────────┬───────┘           │
│                  │                                     │                   │
└──────────────────┼─────────────────────────────────────┼───────────────────┘
                   │                                     │
                   └──────────────┬──────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         STYLE COMPUTATION LAYER                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│                        ┌─────────────────────────────────┐                  │
│                        │        style/ Module            │                  │
│                        │                                 │                  │
│                        │  ┌───────────────────────────┐  │                  │
│                        │  │       style.go            │  │                  │
│                        │  │                           │  │                  │
│                        │  │  • Selector Matching      │  │                  │
│                        │  │  • Specificity (a,b,c,d)  │  │                  │
│                        │  │    - a: inline styles     │  │                  │
│                        │  │    - b: ID selectors      │  │                  │
│                        │  │    - c: class selectors   │  │                  │
│                        │  │    - d: element selectors │  │                  │
│                        │  │  • Cascade Resolution     │  │                  │
│                        │  │  • Property Inheritance   │  │                  │
│                        │  │  • Shorthand Expansion    │  │                  │
│                        │  │    (margin, padding, etc) │  │                  │
│                        │  └───────────┬───────────────┘  │                  │
│                        │              │                  │                  │
│                        │  ┌───────────────────────────┐  │                  │
│                        │  │    useragent.go           │  │                  │
│                        │  │  Default browser styles   │  │                  │
│                        │  └───────────────────────────┘  │                  │
│                        └──────────────┬──────────────────┘                  │
│                                       │                                     │
│                                       ▼                                     │
│                              ┌────────────────┐                             │
│                              │  Styled Tree   │                             │
│                              │ (StyledNode)   │                             │
│                              │ DOM + computed │                             │
│                              │    styles      │                             │
│                              └────────┬───────┘                             │
│                                       │                                     │
└───────────────────────────────────────┼─────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            LAYOUT LAYER                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │                      layout/ Module                                  │   │
│   │                                                                      │   │
│   │  ┌──────────────────────────────────────────────────────────────┐   │   │
│   │  │                     layout.go                                 │   │   │
│   │  │                                                               │   │   │
│   │  │   ┌─────────────────────────────────────────────────────┐    │   │   │
│   │  │   │              CSS Box Model (CSS 2.1 §8)             │    │   │   │
│   │  │   │   ┌─────────────────────────────────────────────┐   │    │   │   │
│   │  │   │   │                  Margin                     │   │    │   │   │
│   │  │   │   │   ┌─────────────────────────────────────┐   │   │    │   │   │
│   │  │   │   │   │              Border                 │   │   │    │   │   │
│   │  │   │   │   │   ┌─────────────────────────────┐   │   │   │    │   │   │
│   │  │   │   │   │   │          Padding            │   │   │   │    │   │   │
│   │  │   │   │   │   │   ┌─────────────────────┐   │   │   │   │    │   │   │
│   │  │   │   │   │   │   │      Content        │   │   │   │   │    │   │   │
│   │  │   │   │   │   │   │   (width × height)  │   │   │   │   │    │   │   │
│   │  │   │   │   │   │   └─────────────────────┘   │   │   │   │    │   │   │
│   │  │   │   │   │   └─────────────────────────────┘   │   │   │    │   │   │
│   │  │   │   │   └─────────────────────────────────────┘   │   │    │   │   │
│   │  │   │   └─────────────────────────────────────────────┘   │    │   │   │
│   │  │   └─────────────────────────────────────────────────────┘    │   │   │
│   │  │                                                               │   │   │
│   │  │  • Block Layout Algorithm (CSS 2.1 §9)                       │   │   │
│   │  │  • Inline Formatting Context                                  │   │   │
│   │  │  • Table Layout                                               │   │   │
│   │  │  • Text Measurement (via font/ module)                       │   │   │
│   │  │  • Image Loading & Sizing                                     │   │   │
│   │  └──────────────────────────────────────────────────────────────┘   │   │
│   │                                                                      │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                       │                                     │
│                                       ▼                                     │
│                              ┌────────────────┐                             │
│                              │  Layout Tree   │                             │
│                              │  (LayoutBox)   │                             │
│                              │ DOM + styles + │                             │
│                              │  dimensions +  │                             │
│                              │   positions    │                             │
│                              └────────┬───────┘                             │
│                                       │                                     │
└───────────────────────────────────────┼─────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           RENDERING LAYER                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌────────────────────┐  ┌────────────────────┐  ┌────────────────────┐   │
│   │   render/ Module   │  │   font/ Module     │  │    svg/ Module     │   │
│   │                    │  │                    │  │                    │   │
│   │  ┌──────────────┐  │  │  ┌──────────────┐  │  │  ┌──────────────┐  │   │
│   │  │ render.go    │  │  │  │  font.go     │  │  │  │   svg.go     │  │   │
│   │  │              │  │  │  │              │  │  │  │              │  │   │
│   │  │ • Canvas     │◄─┼──┼──│ • Font Load  │  │  │  │ • Parse SVG  │  │   │
│   │  │   creation   │  │  │  │ • Text       │  │  │  │ • Extract    │  │   │
│   │  │ • Background │  │  │  │   measure    │  │  │  │   paths      │  │   │
│   │  │   drawing    │  │  │  │ • Go fonts   │  │  │  └──────┬───────┘  │   │
│   │  │ • Border     │  │  │  │   (regular,  │  │  │         │          │   │
│   │  │   drawing    │  │  │  │    bold,     │  │  │  ┌──────▼───────┐  │   │
│   │  │ • Text       │  │  │  │    italic)   │  │  │  │rasterizer.go │  │   │
│   │  │   rendering  │  │  │  └──────────────┘  │  │  │              │  │   │
│   │  │ • Image      │◄─┼──┼────────────────────┼──┼──│ • Scanline   │  │   │
│   │  │   rendering  │  │  │                    │  │  │   rasterize  │  │   │
│   │  │ • SVG        │  │  │                    │  │  └──────────────┘  │   │
│   │  │   rasterize  │  │  │                    │  │                    │   │
│   │  └──────────────┘  │  │                    │  │                    │   │
│   └────────────────────┘  └────────────────────┘  └────────────────────┘   │
│                                       │                                     │
│                                       ▼                                     │
│                              ┌────────────────┐                             │
│                              │     Canvas     │                             │
│                              │  Pixel Buffer  │                             │
│                              │ ([]color.RGBA) │                             │
│                              └────────┬───────┘                             │
│                                       │                                     │
└───────────────────────────────────────┼─────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            OUTPUT LAYER                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│          ┌─────────────────────┐       ┌─────────────────────┐              │
│          │   CLI Application   │       │  WASM Application   │              │
│          │ cmd/browser/main.go │       │cmd/browser-wasm/    │              │
│          │                     │       │     main.go         │              │
│          │   PNG File Output   │       │  Base64 for JS      │              │
│          └─────────────────────┘       └─────────────────────┘              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Module Dependencies

```
                    ┌─────────────┐
                    │    main     │
                    │  (cmd/*)    │
                    └──────┬──────┘
                           │
           ┌───────────────┼───────────────┐
           │               │               │
           ▼               ▼               ▼
    ┌──────────┐    ┌──────────┐    ┌──────────┐
    │  render  │    │  layout  │    │  style   │
    └────┬─────┘    └────┬─────┘    └────┬─────┘
         │               │               │
    ┌────┴────┐     ┌────┴────┐     ┌────┴────┐
    │         │     │         │     │         │
    ▼         ▼     ▼         ▼     ▼         ▼
┌──────┐  ┌─────┐ ┌─────┐  ┌─────┐ ┌─────┐ ┌─────┐
│ svg  │  │font │ │font │  │ dom │ │ css │ │ dom │
└──────┘  └─────┘ └─────┘  └─────┘ └─────┘ └─────┘
                                       │
                                       ▼
                                   ┌──────┐
                                   │ html │
                                   └──────┘
```

## Data Flow Pipeline

```
HTML Input ──► Tokenizer ──► Parser ──► DOM Tree
                                            │
                                            ├──► Extract <style> content
                                            │
                                            └──► Fetch <link> stylesheets
                                                        │
CSS Input ◄─────────────────────────────────────────────┘
    │
    ▼
Tokenizer ──► Parser ──► Stylesheet (Rules)
                              │
                              ▼
              ┌───────────────────────────────┐
              │     Style Computation         │
              │  DOM Tree + Stylesheet        │
              │           ↓                   │
              │  Selector Matching            │
              │           ↓                   │
              │  Specificity Calculation      │
              │           ↓                   │
              │  Cascade Resolution           │
              │           ↓                   │
              │  Styled Tree                  │
              └───────────────┬───────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │         Layout                │
              │  Styled Tree + Viewport       │
              │           ↓                   │
              │  Box Type Determination       │
              │           ↓                   │
              │  Width/Height Calculation     │
              │           ↓                   │
              │  Position Calculation         │
              │           ↓                   │
              │  Layout Tree                  │
              └───────────────┬───────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │        Rendering              │
              │  Layout Tree + Canvas         │
              │           ↓                   │
              │  Background Drawing           │
              │           ↓                   │
              │  Border Drawing               │
              │           ↓                   │
              │  Text Rendering               │
              │           ↓                   │
              │  Image Rendering              │
              │           ↓                   │
              │  Pixel Buffer                 │
              └───────────────┬───────────────┘
                              │
                              ▼
                         PNG Output
```

## Key Data Structures

### DOM Node (dom/node.go)
```go
type Node struct {
    Type       NodeType    // Element, Text, Document
    TagName    string      // For elements: "div", "p", etc.
    Attributes map[string]string
    Children   []*Node
    Text       string      // For text nodes
}
```

### CSS Rule (css/parser.go)
```go
type Rule struct {
    Selectors    []Selector
    Declarations []Declaration
}

type Selector struct {
    Tag     string   // Element name
    ID      string   // #id
    Classes []string // .class
}

type Declaration struct {
    Property string
    Value    string
}
```

### Styled Node (style/style.go)
```go
type StyledNode struct {
    Node     *dom.Node
    Styles   map[string]string  // Computed CSS properties
    Children []*StyledNode
}
```

### Layout Box (layout/layout.go)
```go
type LayoutBox struct {
    BoxType    BoxType      // Block, Inline, Anonymous
    StyledNode *StyledNode
    Dimensions Dimensions
    Children   []*LayoutBox
}

type Dimensions struct {
    Content Rect      // x, y, width, height
    Padding EdgeSizes // top, right, bottom, left
    Border  EdgeSizes
    Margin  EdgeSizes
}
```

## Specification Compliance

| Module | Specification |
|--------|---------------|
| HTML Tokenizer | HTML5 §12.2.5 |
| HTML Parser | HTML5 §12.2.6 |
| URL Resolution | HTML5 §2.5 |
| CSS Tokenizer | CSS 2.1 §4 |
| CSS Selectors | CSS 2.1 §5 |
| Cascade | CSS 2.1 §6 |
| Box Model | CSS 2.1 §8 |
| Visual Formatting | CSS 2.1 §9, §10 |
| Colors | CSS 2.1 §14 |
| Fonts | CSS 2.1 §15 |
| Data URLs | RFC 2397 |

## External Dependencies

- `golang.org/x/image` - TrueType font rendering and image processing
- Go standard library for everything else
