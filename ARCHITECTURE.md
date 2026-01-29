# Browser Architecture

This document describes the architecture of the Go-based web browser that renders HTML/CSS to PNG images.

## High-Level Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              INPUT SOURCES                                  │
│                    Local File  │  HTTP/HTTPS URL  │  Data URL               │
└─────────────────────────────────────────┬───────────────────────────────────┘
                                          │
                                          ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           PARSING LAYER                                     │
│  ┌─────────────────────────────┐    ┌─────────────────────────────────────┐ │
│  │      HTML Parser            │    │         CSS Parser                  │ │
│  │  ┌───────────────────────┐  │    │  ┌───────────────────────────────┐  │ │
│  │  │ Tokenizer             │  │    │  │ Tokenizer                     │  │ │
│  │  │ (State Machine)       │  │    │  │ (CSS 2.1 §4)                  │  │ │
│  │  └───────────┬───────────┘  │    │  └───────────────┬───────────────┘  │ │
│  │              ▼              │    │                  ▼                  │ │
│  │  ┌───────────────────────┐  │    │  ┌───────────────────────────────┐  │ │
│  │  │ Tree Constructor      │  │    │  │ Declaration Parser            │  │ │
│  │  │ (HTML5 §12.2)         │  │    │  │ (Selectors + Properties)      │  │ │
│  │  └───────────────────────┘  │    │  └───────────────────────────────┘  │ │
│  └──────────────┬──────────────┘    └──────────────────┬──────────────────┘ │
│                 │                                      │                    │
│                 ▼                                      ▼                    │
│           DOM Tree                               Stylesheet                 │
└─────────────────┬──────────────────────────────────────┬────────────────────┘
                  │                                      │
                  └──────────────────┬───────────────────┘
                                     ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           STYLE COMPUTATION                                 │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Selector Matching  →  Specificity Calc  →  Cascade Resolution      │    │
│  │       (CSS 2.1 §5)         (CSS 2.1 §6.4.3)        (CSS 2.1 §6)     │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                     │                                       │
│                                     ▼                                       │
│                              Styled Tree                                    │
│                    (DOM nodes with computed styles)                         │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           LAYOUT ENGINE                                     │
│  ┌─────────────────────────────────────────────────────────────────────┐    │
│  │  Box Model Calculation   │   Block Layout   │   Table Layout        │    │
│  │     (CSS 2.1 §8)         │  (CSS 2.1 §9)    │   (CSS 2.1 §17)       │    │
│  └─────────────────────────────────────────────────────────────────────┘    │
│                                     │                                       │
│                                     ▼                                       │
│                              Layout Tree                                    │
│                    (Boxes with dimensions & positions)                      │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           RENDER ENGINE                                     │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐  ┌──────────────┐  │
│  │  Backgrounds  │  │    Borders    │  │     Text      │  │    Images    │  │
│  │  (Colors/Img) │  │   (Solid)     │  │  (TrueType)   │  │ (PNG/JPG/SVG)│  │
│  └───────────────┘  └───────────────┘  └───────────────┘  └──────────────┘  │
│                                     │                                       │
│                                     ▼                                       │
│                                  Canvas                                     │
│                             (Pixel Buffer)                                  │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              OUTPUT                                         │
│                    PNG File  │  Base64 (WASM)                               │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Package Structure

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              cmd/                                           │
│  ┌─────────────────────────┐    ┌─────────────────────────────────────────┐ │
│  │    cmd/browser/         │    │         cmd/browser-wasm/               │ │
│  │    CLI Entry Point      │    │         WASM Entry Point                │ │
│  └─────────────────────────┘    └─────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    ▼                 ▼                 ▼
┌──────────────────────┐ ┌──────────────────┐ ┌──────────────────────────────┐
│       html/          │ │      css/        │ │           dom/               │
│  ├── tokenizer.go    │ │ ├── tokenizer.go │ │  ├── node.go (DOM structure) │
│  └── parser.go       │ │ ├── parser.go    │ │  ├── url.go (URL resolution) │
│                      │ │ └── value.go     │ │  └── loader.go (Fetching)    │
│  HTML5 Parsing       │ │                  │ │                              │
│                      │ │  CSS 2.1 Parsing │ │     DOM & Resource Loading   │
└──────────────────────┘ └──────────────────┘ └──────────────────────────────┘
           │                      │                         │
           └──────────────────────┼─────────────────────────┘
                                  ▼
                    ┌──────────────────────────┐
                    │        style/            │
                    │  ├── style.go            │
                    │  │   (Selector matching, │
                    │  │    Cascade, Inherit)  │
                    │  └── useragent.go        │
                    │      (Default styles)    │
                    └────────────┬─────────────┘
                                 │
                                 ▼
                    ┌──────────────────────────┐
                    │        layout/           │
                    │  ├── layout.go           │
                    │  │   (Box model,         │
                    │  │    Block layout)      │
                    │  └── table.go            │
                    │      (Table layout)      │
                    └────────────┬─────────────┘
                                 │
                                 ▼
┌──────────────────────┐ ┌──────────────────┐ ┌──────────────────────────────┐
│       render/        │ │      svg/        │ │           font/              │
│  ├── render.go       │ │ ├── svg.go       │ │  └── font.go                 │
│  └── image.go        │ │ └── rasterizer.go│ │                              │
│                      │ │                  │ │     TrueType Font Loading    │
│  Canvas & Drawing    │ │  SVG Rendering   │ │     (Go Fonts Embedded)      │
└──────────────────────┘ └──────────────────┘ └──────────────────────────────┘
```

## Data Flow Diagram

```
                    ┌───────────────────────────────────────┐
                    │              User Input               │
                    │   (URL, File Path, or HTML String)    │
                    └───────────────────┬───────────────────┘
                                        │
                                        ▼
                    ┌───────────────────────────────────────┐
                    │           dom/loader.go               │
                    │    Load content (file/HTTP/data)      │
                    └───────────────────┬───────────────────┘
                                        │
                         ┌──────────────┴──────────────┐
                         ▼                             ▼
            ┌─────────────────────────┐   ┌─────────────────────────┐
            │    html/parser.go       │   │  Extract <style> tags   │
            │    Parse HTML → DOM     │   │  Fetch <link> stylesheets│
            └────────────┬────────────┘   └────────────┬────────────┘
                         │                             │
                         │                             ▼
                         │                ┌─────────────────────────┐
                         │                │    css/parser.go        │
                         │                │  Parse CSS → Stylesheet │
                         │                └────────────┬────────────┘
                         │                             │
                         └──────────────┬──────────────┘
                                        │
                                        ▼
                    ┌───────────────────────────────────────┐
                    │           style/style.go              │
                    │  Match selectors, compute styles      │
                    │                                       │
                    │  For each DOM node:                   │
                    │   1. Find matching CSS rules          │
                    │   2. Sort by specificity              │
                    │   3. Apply cascade                    │
                    │   4. Handle inheritance               │
                    └───────────────────┬───────────────────┘
                                        │
                                        ▼
                    ┌───────────────────────────────────────┐
                    │          layout/layout.go             │
                    │  Calculate box dimensions & positions │
                    │                                       │
                    │  For each styled node:                │
                    │   1. Create layout box                │
                    │   2. Calculate width/height           │
                    │   3. Apply padding/border/margin      │
                    │   4. Position in document flow        │
                    └───────────────────┬───────────────────┘
                                        │
                                        ▼
                    ┌───────────────────────────────────────┐
                    │          render/render.go             │
                    │  Paint layout boxes to canvas         │
                    │                                       │
                    │  For each layout box:                 │
                    │   1. Draw background (color/image)    │
                    │   2. Draw borders                     │
                    │   3. Draw text content                │
                    │   4. Draw images                      │
                    └───────────────────┬───────────────────┘
                                        │
                                        ▼
                    ┌───────────────────────────────────────┐
                    │              Output                   │
                    │        PNG file or base64 data        │
                    └───────────────────────────────────────┘
```

## Key Data Structures

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           DOM Node (dom/node.go)                            │
├─────────────────────────────────────────────────────────────────────────────┤
│  type Node struct {                                                         │
│      Type       NodeType           // Element, Text, Document               │
│      Data       string             // Tag name or text content              │
│      Attributes map[string]string  // HTML attributes                       │
│      Children   []*Node            // Child nodes                           │
│      Parent     *Node              // Parent node                           │
│  }                                                                          │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      │ + Computed Styles
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                      Styled Node (style/style.go)                           │
├─────────────────────────────────────────────────────────────────────────────┤
│  type StyledNode struct {                                                   │
│      Node     *dom.Node             // Reference to DOM node                │
│      Styles   map[string]string     // Computed CSS property values         │
│      Children []*StyledNode         // Styled child nodes                   │
│  }                                                                          │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      │ + Layout Dimensions
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                      Layout Box (layout/layout.go)                          │
├─────────────────────────────────────────────────────────────────────────────┤
│  type LayoutBox struct {                                                    │
│      BoxType    BoxType             // Block, Inline, Anonymous             │
│      StyledNode *style.StyledNode   // Reference to styled node             │
│      Dimensions Dimensions          // Position and size                    │
│      Children   []*LayoutBox        // Child layout boxes                   │
│  }                                                                          │
│                                                                             │
│  type Dimensions struct {                                                   │
│      Content Rect       // x, y, width, height of content area              │
│      Padding EdgeSizes  // top, right, bottom, left padding                 │
│      Border  EdgeSizes  // top, right, bottom, left border                  │
│      Margin  EdgeSizes  // top, right, bottom, left margin                  │
│  }                                                                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

## CSS Cascade & Specificity

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Specificity Calculation                              │
│                          (CSS 2.1 §6.4.3)                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   Priority (highest to lowest):                                             │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐   │
│   │  1. Inline styles (style="...")              Specificity: 1,0,0,0   │   │
│   ├─────────────────────────────────────────────────────────────────────┤   │
│   │  2. ID selectors (#id)                       Specificity: 0,1,0,0   │   │
│   ├─────────────────────────────────────────────────────────────────────┤   │
│   │  3. Class selectors (.class)                 Specificity: 0,0,1,0   │   │
│   ├─────────────────────────────────────────────────────────────────────┤   │
│   │  4. Type selectors (div, p)                  Specificity: 0,0,0,1   │   │
│   ├─────────────────────────────────────────────────────────────────────┤   │
│   │  5. User-agent stylesheet (defaults)         Specificity: 0,0,0,0   │   │
│   └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│   Cascade Order:                                                            │
│   User-Agent → Author Stylesheets → Inline Styles                           │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Supported Features

| Category | Supported | Not Supported |
|----------|-----------|---------------|
| **HTML** | Standard elements (div, p, span, h1-h6, table, img, etc.) | Scripts, forms, interactive elements |
| **CSS Selectors** | Type, class, ID, descendant | Attribute, child (>), sibling (+, ~), pseudo-classes |
| **Box Model** | Content, padding, border, margin | Box-sizing |
| **Layout** | Block, inline, table | Float, position (absolute/relative/fixed), flexbox, grid |
| **Colors** | Named colors, hex (#RGB, #RRGGBB) | RGB(), HSL(), alpha |
| **Text** | Bold, italic, underline, sizes, alignment | Web fonts, letter-spacing, word-spacing |
| **Images** | PNG, JPEG, GIF, SVG, data URLs | Video, audio, canvas |
| **Network** | HTTP, HTTPS, data URLs | JavaScript fetch, CORS |

## Entry Points

### CLI (`cmd/browser/main.go`)
```bash
# Render URL to PNG
./browser -output out.png -width 800 -height 600 https://example.com

# Render local file
./browser -output out.png test/styled.html

# Debug layout
./browser -show-layout test/styled.html
```

### WebAssembly (`cmd/browser-wasm/main.go`)
```javascript
// JavaScript API
const pngBase64 = renderHTML(htmlContent, width, height);
```

## Dependencies

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          External Dependencies                              │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  Go Standard Library:                                                       │
│  ├── net/http         HTTP/HTTPS fetching                                   │
│  ├── image/png        PNG encoding/decoding                                 │
│  ├── image/jpeg       JPEG decoding                                         │
│  ├── image/gif        GIF decoding                                          │
│  ├── encoding/base64  Data URL handling                                     │
│  └── syscall/js       WebAssembly JavaScript interop                        │
│                                                                             │
│  Third-Party:                                                               │
│  ├── golang.org/x/image (v0.34.0)  TrueType font rendering                  │
│  └── golang.org/x/text (v0.32.0)   Unicode text processing                  │
│                                                                             │
│  Embedded:                                                                  │
│  └── Go Fonts (TrueType)           Text rendering fonts                     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```
