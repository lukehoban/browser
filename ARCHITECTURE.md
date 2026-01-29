# Browser Architecture

This document provides a comprehensive architecture diagram and overview of the browser's rendering pipeline.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                                    BROWSER                                          │
│                                                                                     │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │    INPUT     │───►│   PARSING    │───►│   STYLING    │───►│   LAYOUT     │      │
│  │              │    │              │    │              │    │              │      │
│  │ - HTML file  │    │ - Tokenize   │    │ - Cascade    │    │ - Box Model  │      │
│  │ - URL (HTTP) │    │ - Parse DOM  │    │ - Specificity│    │ - Dimensions │      │
│  │              │    │ - Parse CSS  │    │ - Inheritance│    │ - Positioning│      │
│  └──────────────┘    └──────────────┘    └──────────────┘    └──────────────┘      │
│                                                                       │             │
│                                                                       ▼             │
│                      ┌──────────────────────────────────────────────────┐           │
│                      │                   RENDERING                       │           │
│                      │                                                   │           │
│                      │  - Background colors    - Text drawing            │           │
│                      │  - Border rendering     - Image loading           │           │
│                      │  - Box painting         - PNG output              │           │
│                      └──────────────────────────────────────────────────┘           │
│                                                │                                     │
│                                                ▼                                     │
│                                        ┌──────────────┐                             │
│                                        │   OUTPUT     │                             │
│                                        │              │                             │
│                                        │  PNG Image   │                             │
│                                        └──────────────┘                             │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## Detailed Data Flow

```
                              ┌─────────────────┐
                              │  HTML Source    │
                              │  (file or URL)  │
                              └────────┬────────┘
                                       │
                                       ▼
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                               HTML PROCESSING                                         │
│  ┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐              │
│  │   Tokenizer     │─────►│    Parser       │─────►│    DOM Tree     │              │
│  │   (html/)       │      │    (html/)      │      │    (dom/)       │              │
│  │                 │      │                 │      │                 │              │
│  │ HTML5 §12.2.5   │      │ Tree            │      │ Element nodes   │              │
│  │ State machine   │      │ construction    │      │ Text nodes      │              │
│  │ Tokens output   │      │ algorithm       │      │ Attributes      │              │
│  └─────────────────┘      └─────────────────┘      └────────┬────────┘              │
└──────────────────────────────────────────────────────────────┼──────────────────────┘
                                                               │
                              ┌─────────────────┐              │
                              │  CSS Source     │              │
                              │  (<style> tags, │              │
                              │   <link> refs)  │              │
                              └────────┬────────┘              │
                                       │                       │
                                       ▼                       │
┌──────────────────────────────────────────────────────────────┼──────────────────────┐
│                               CSS PROCESSING                  │                      │
│  ┌─────────────────┐      ┌─────────────────┐                │                      │
│  │   Tokenizer     │─────►│    Parser       │───┐            │                      │
│  │   (css/)        │      │    (css/)       │   │            │                      │
│  │                 │      │                 │   │            │                      │
│  │ CSS 2.1 §4      │      │ Selectors       │   │            │                      │
│  │ Identifiers,    │      │ Declarations    │   │            │                      │
│  │ strings, etc.   │      │ Rule sets       │   │            │                      │
│  └─────────────────┘      └─────────────────┘   │            │                      │
│                                                  │            │                      │
│                           ┌─────────────────┐   │            │                      │
│                           │   Stylesheet    │◄──┘            │                      │
│                           │                 │                │                      │
│                           │ List of rules   │                │                      │
│                           │ with selectors  │                │                      │
│                           └────────┬────────┘                │                      │
└────────────────────────────────────┼─────────────────────────┼──────────────────────┘
                                     │                         │
                                     ▼                         ▼
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                              STYLE COMPUTATION                                        │
│                                  (style/)                                            │
│                                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────────────┐    │
│  │                          Selector Matching                                   │    │
│  │                                                                             │    │
│  │   For each DOM node:                                                        │    │
│  │   1. Match against all CSS rules                                            │    │
│  │   2. Calculate specificity (CSS 2.1 §6.4.3)                                 │    │
│  │   3. Apply cascade (sort by specificity, then source order)                 │    │
│  │   4. Expand shorthand properties (margin, padding, border)                  │    │
│  │   5. Inherit font properties from parent                                    │    │
│  │   6. Parse inline styles (style="...")                                      │    │
│  └─────────────────────────────────────────────────────────────────────────────┘    │
│                                       │                                              │
│                                       ▼                                              │
│                           ┌─────────────────┐                                        │
│                           │   Styled Tree   │                                        │
│                           │                 │                                        │
│                           │ StyledNode with │                                        │
│                           │ computed styles │                                        │
│                           └────────┬────────┘                                        │
└────────────────────────────────────┼─────────────────────────────────────────────────┘
                                     │
                                     ▼
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                              LAYOUT ENGINE                                            │
│                                 (layout/)                                            │
│                                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────────────┐    │
│  │                           Box Model Layout                                   │    │
│  │                                                                             │    │
│  │   ┌─────────────────────────────────────────────────────────────────────┐   │    │
│  │   │                            Margin                                    │   │    │
│  │   │   ┌─────────────────────────────────────────────────────────────┐   │   │    │
│  │   │   │                         Border                               │   │   │    │
│  │   │   │   ┌─────────────────────────────────────────────────────┐   │   │   │    │
│  │   │   │   │                      Padding                         │   │   │   │    │
│  │   │   │   │   ┌─────────────────────────────────────────────┐   │   │   │   │    │
│  │   │   │   │   │              Content Box                     │   │   │   │   │    │
│  │   │   │   │   │                                             │   │   │   │   │    │
│  │   │   │   │   │  x, y, width, height                        │   │   │   │   │    │
│  │   │   │   │   │                                             │   │   │   │   │    │
│  │   │   │   │   └─────────────────────────────────────────────┘   │   │   │   │    │
│  │   │   │   └─────────────────────────────────────────────────────┘   │   │   │    │
│  │   │   └─────────────────────────────────────────────────────────────┘   │   │    │
│  │   └─────────────────────────────────────────────────────────────────────┘   │    │
│  └─────────────────────────────────────────────────────────────────────────────┘    │
│                                                                                      │
│  Layout Algorithm (CSS 2.1 §9 Visual Formatting Model):                             │
│  1. Calculate width (auto fills container, or explicit value)                        │
│  2. Calculate horizontal margins                                                     │
│  3. Position children vertically (block formatting context)                          │
│  4. Calculate height (auto = sum of children, or explicit)                          │
│                                       │                                              │
│                                       ▼                                              │
│                           ┌─────────────────┐                                        │
│                           │   Layout Tree   │                                        │
│                           │                 │                                        │
│                           │ LayoutBox with  │                                        │
│                           │ dimensions      │                                        │
│                           └────────┬────────┘                                        │
└────────────────────────────────────┼─────────────────────────────────────────────────┘
                                     │
                                     ▼
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                              RENDERING ENGINE                                         │
│                                 (render/)                                            │
│                                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────────────┐    │
│  │                          Render Operations                                   │    │
│  │                                                                             │    │
│  │   1. Create pixel canvas (width × height)                                   │    │
│  │   2. Fill background (default white)                                        │    │
│  │   3. Traverse layout tree depth-first:                                      │    │
│  │      a. Draw background color (background-color)                            │    │
│  │      b. Draw borders (border-width, border-color)                           │    │
│  │      c. Draw text content (color, font-family, font-size)                   │    │
│  │      d. Draw images (<img> elements, background-image)                      │    │
│  │   4. Encode to PNG format                                                   │    │
│  └─────────────────────────────────────────────────────────────────────────────┘    │
│                                       │                                              │
│                                       ▼                                              │
│                           ┌─────────────────┐                                        │
│                           │   PNG Image     │                                        │
│                           │                 │                                        │
│                           │ Final rendered  │                                        │
│                           │ output          │                                        │
│                           └─────────────────┘                                        │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

## Package Structure

```
browser/
├── cmd/
│   ├── browser/           # Main CLI application
│   │   └── main.go        # Entry point, orchestrates pipeline
│   └── browser-wasm/      # WebAssembly entry point
│       └── main.go        # WASM bindings for web execution
│
├── dom/                   # Document Object Model
│   ├── node.go            # Node type, Element/Text/Document nodes
│   ├── url.go             # URL resolution (HTML5 §2.5)
│   └── loader.go          # External resource fetching
│
├── html/                  # HTML Processing
│   ├── tokenizer.go       # HTML5 §12.2.5 tokenization
│   └── parser.go          # Tree construction algorithm
│
├── css/                   # CSS Processing
│   ├── tokenizer.go       # CSS 2.1 §4 tokenization
│   ├── parser.go          # Selector & declaration parsing
│   └── values.go          # CSS value types and parsing
│
├── style/                 # Style Computation
│   ├── style.go           # Selector matching, cascade
│   └── useragent.go       # Default browser stylesheet
│
├── layout/                # Layout Engine
│   └── layout.go          # Box model, positioning
│
├── render/                # Rendering
│   └── render.go          # Canvas, painting, PNG output
│
├── font/                  # Font handling
│   └── font.go            # Go fonts integration
│
├── svg/                   # SVG Support
│   └── svg.go             # SVG image rendering
│
├── log/                   # Logging utilities
│   └── log.go             # Leveled logging
│
├── wasm/                  # WebAssembly demo
│   ├── index.html         # Demo page
│   └── README.md          # WASM documentation
│
├── reftest/               # Reference testing
│   └── reftest.go         # WPT reftest harness
│
└── test/                  # Test fixtures
    ├── simple.html
    ├── styled.html
    └── hackernews.html
```

## Core Data Structures

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              DOM Node (dom/node.go)                                  │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  type Node struct {                                                                 │
│      Type       NodeType              // ElementNode, TextNode, DocumentNode        │
│      Data       string                // Tag name or text content                   │
│      Attributes map[string]string     // Element attributes                         │
│      Children   []*Node               // Child nodes                                │
│      Parent     *Node                 // Parent reference                           │
│  }                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        │ style.StyleTree()
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                            Styled Node (style/style.go)                              │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  type StyledNode struct {                                                           │
│      Node     *dom.Node               // Reference to DOM node                      │
│      Styles   map[string]string       // Computed CSS properties                    │
│      Children []*StyledNode           // Styled children                            │
│  }                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        │ layout.LayoutTree()
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                           Layout Box (layout/layout.go)                              │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  type LayoutBox struct {                                                            │
│      BoxType     BoxType              // BlockBox, InlineBox, AnonymousBox          │
│      Dimensions  Dimensions           // Position and size info                     │
│      StyledNode  *StyledNode          // Reference to styled node                   │
│      Children    []*LayoutBox         // Child layout boxes                         │
│  }                                                                                  │
│                                                                                     │
│  type Dimensions struct {                                                           │
│      Content  Rect      // x, y, width, height                                      │
│      Padding  EdgeSize  // top, right, bottom, left                                 │
│      Border   EdgeSize  // top, right, bottom, left                                 │
│      Margin   EdgeSize  // top, right, bottom, left                                 │
│  }                                                                                  │
└─────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        │ render.Render()
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              Canvas (render/render.go)                               │
├─────────────────────────────────────────────────────────────────────────────────────┤
│  type Canvas struct {                                                               │
│      Pixels  []color.Color           // Pixel buffer                                │
│      Width   int                     // Canvas width                                │
│      Height  int                     // Canvas height                               │
│  }                                                                                  │
│                                                                                     │
│  Methods:                                                                           │
│  - DrawRect(rect, color)             // Fill rectangle                              │
│  - DrawBorder(dims, width, color)    // Draw border                                 │
│  - DrawText(text, x, y, font, color) // Render text                                 │
│  - DrawImage(img, x, y, w, h)        // Draw image                                  │
│  - SavePNG(filename)                 // Export to PNG                               │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

## Processing Pipeline in Detail

### 1. Input Handling (cmd/browser/main.go)

```
Input → isURL() check → fetch/read content → establish base URL
                                │
              ┌─────────────────┴─────────────────┐
              │                                   │
       HTTP/HTTPS URL                        Local File
              │                                   │
        fetchURL()                          os.ReadFile()
              │                                   │
              └─────────────────┬─────────────────┘
                                │
                                ▼
                         HTML content string
```

### 2. HTML Parsing (html/)

```
HTML string → Tokenizer → Token stream → Parser → DOM Tree

Tokenizer States (HTML5 §12.2.5):
- Data state (text content)
- Tag open state (<)
- Tag name state (div, p, etc.)
- Attribute name state
- Attribute value state (quoted/unquoted)
- Self-closing state (/>)

Parser Operations:
- Create element nodes
- Handle void elements (img, br, hr)
- Build parent-child relationships
- Handle implicit tag closing
```

### 3. CSS Parsing (css/)

```
CSS string → Tokenizer → Token stream → Parser → Stylesheet

Stylesheet = List of Rules
Rule = {
    Selectors: [Selector]
    Declarations: [Declaration]
}

Selector Types:
- Element: div, p, h1
- Class: .classname
- ID: #idname
- Descendant: div p
- Multiple: h1, h2, h3

Declaration = property: value;
```

### 4. Style Computation (style/)

```
DOM Tree + Stylesheet → Style Matching → Styled Tree

For each element:
1. Collect matching rules
   - User-agent stylesheet (defaults)
   - Author stylesheets
   - Inline styles (style="...")

2. Calculate specificity (a,b,c,d):
   a = inline style (0 or 1)
   b = count of ID selectors
   c = count of class selectors
   d = count of element selectors

3. Sort by specificity, then source order

4. Apply cascade (later rules override earlier)

5. Expand shorthands (margin → margin-top, etc.)

6. Inherit font properties from parent
```

### 5. Layout Calculation (layout/)

```
Styled Tree + Viewport → Layout Algorithm → Layout Tree

Width Calculation (CSS 2.1 §10.3.3):
- auto: fills containing block
- percentage: relative to container
- pixels: exact value

Height Calculation (CSS 2.1 §10.6.3):
- auto: sum of children heights
- percentage: relative to container (if height defined)
- pixels: exact value

Position Calculation:
- x = parent.x + margin.left + border.left + padding.left
- y = previous_sibling.bottom + margin collapsing
```

### 6. Rendering (render/)

```
Layout Tree + Viewport → Paint Operations → Canvas → PNG

Paint Order (CSS 2.1 Appendix E):
1. Background colors
2. Background images
3. Borders
4. Block children (recursive)
5. Inline content (text, images)

Text Rendering:
- Load Go fonts (proportional sans-serif)
- Calculate text position
- Draw glyphs with color

Image Rendering:
- Load and decode image (PNG, JPEG, GIF, SVG)
- Scale to specified dimensions
- Blend with background (alpha)
```

## Network Support

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                              Network Resource Loading                                 │
│                                                                                      │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐        │
│  │ Main HTML   │────►│ Parse DOM   │────►│ Find <link> │────►│ Fetch CSS   │        │
│  │ (URL/file)  │     │             │     │ elements    │     │ stylesheets │        │
│  └─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘        │
│                                                                     │                │
│                                                                     ▼                │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐        │
│  │ Resolve     │◄────│ Find <img>  │◄────│ Style/      │◄────│ Parse CSS   │        │
│  │ image URLs  │     │ elements    │     │ Layout      │     │             │        │
│  └──────┬──────┘     └─────────────┘     └─────────────┘     └─────────────┘        │
│         │                                                                            │
│         ▼                                                                            │
│  ┌─────────────┐                                                                     │
│  │ Fetch &     │     Supported URL schemes:                                          │
│  │ cache imgs  │     - http:// and https:// (network)                               │
│  └─────────────┘     - file:// (local filesystem)                                   │
│                      - data: (RFC 2397 inline data)                                 │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

## WebAssembly Architecture

```
┌──────────────────────────────────────────────────────────────────────────────────────┐
│                                   Browser (Web)                                       │
│  ┌──────────────────────────────────────────────────────────────────────────────┐   │
│  │                              JavaScript Host                                   │   │
│  │                                                                               │   │
│  │   - Load browser.wasm                                                         │   │
│  │   - Provide fetch() for network requests                                      │   │
│  │   - Receive PNG data                                                          │   │
│  │   - Display in <canvas> or <img>                                              │   │
│  └──────────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                             │
│                                        │ js.Global()                                 │
│                                        ▼                                             │
│  ┌──────────────────────────────────────────────────────────────────────────────┐   │
│  │                              Go WASM Module                                    │   │
│  │                            (browser-wasm/)                                    │   │
│  │                                                                               │   │
│  │   - Export render(htmlURL, width, height) function                           │   │
│  │   - Full browser pipeline in WASM                                            │   │
│  │   - Return PNG as base64 or blob                                             │   │
│  └──────────────────────────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────────────────────────┘
```

## Specification Compliance

| Component | W3C Specification | Coverage |
|-----------|-------------------|----------|
| HTML Tokenizer | HTML5 §12.2.5 | Partial (common states) |
| HTML Parser | HTML5 §12.2.6 | Simplified tree construction |
| CSS Tokenizer | CSS 2.1 §4 | Core tokens |
| CSS Parser | CSS 2.1 §4.1.7, §4.1.8 | Rule sets, declarations |
| Selectors | CSS 2.1 §5 | Element, class, ID, descendant |
| Cascade | CSS 2.1 §6 | Specificity-based |
| Box Model | CSS 2.1 §8 | Full box dimensions |
| Layout | CSS 2.1 §9, §10 | Block layout, normal flow |
| Colors | CSS 2.1 §14 | Named colors, hex colors |
| URLs | HTML5 §2.5, RFC 2397 | Relative resolution, data URLs |

## Design Principles

1. **Specification-driven**: Code cites relevant W3C spec sections
2. **Modular**: Clear separation between parsing, styling, layout, rendering
3. **Testable**: High unit test coverage with WPT integration
4. **Educational**: Code structure mirrors spec organization
5. **Incremental**: Built in layers from bottom up
