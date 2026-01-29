# Browser Architecture

A web browser implementation in Go that renders static HTML/CSS to PNG output, following W3C specifications.

## High-Level Pipeline

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           BROWSER RENDERING PIPELINE                         │
└─────────────────────────────────────────────────────────────────────────────┘

     ┌──────────────┐          ┌──────────────┐
     │   URL/File   │          │  <link> CSS  │
     │    Input     │          │  Stylesheets │
     └──────┬───────┘          └──────┬───────┘
            │                         │
            ▼                         ▼
     ┌──────────────┐          ┌──────────────┐
     │     HTML     │          │     CSS      │
     │    Parser    │          │    Parser    │
     │   (html/)    │          │    (css/)    │
     └──────┬───────┘          └──────┬───────┘
            │                         │
            ▼                         ▼
     ┌──────────────┐          ┌──────────────┐
     │   DOM Tree   │          │  Stylesheet  │
     │    (dom/)    │          │    Rules     │
     └──────┬───────┘          └──────┬───────┘
            │                         │
            └───────────┬─────────────┘
                        │
                        ▼
              ┌──────────────────┐
              │   Style Engine   │
              │    (style/)      │
              │                  │
              │ • Selector match │
              │ • Specificity    │
              │ • Cascade        │
              │ • Inheritance    │
              └────────┬─────────┘
                       │
                       ▼
              ┌──────────────────┐
              │   Styled Tree    │
              │ (StyledNode +    │
              │  computed CSS)   │
              └────────┬─────────┘
                       │
                       ▼
              ┌──────────────────┐
              │  Layout Engine   │
              │    (layout/)     │
              │                  │
              │ • Box model      │
              │ • Block/Inline   │
              │ • Table layout   │
              │ • Positioning    │
              └────────┬─────────┘
                       │
                       ▼
              ┌──────────────────┐
              │   Layout Tree    │
              │ (LayoutBox +     │
              │  dimensions)     │
              └────────┬─────────┘
                       │
                       ▼
              ┌──────────────────┐
              │  Render Engine   │
              │    (render/)     │
              │                  │
              │ • Backgrounds    │
              │ • Borders        │
              │ • Text/Fonts     │
              │ • Images/SVG     │
              └────────┬─────────┘
                       │
                       ▼
              ┌──────────────────┐
              │    PNG Output    │
              └──────────────────┘
```

## Module Structure

```
browser/
├── cmd/browser/          # CLI entry point
│   └── main.go           # Orchestrates the pipeline
│
├── html/                 # HTML Parsing
│   ├── tokenizer.go      # State machine tokenizer (HTML5 spec)
│   └── parser.go         # Tree construction algorithm
│
├── css/                  # CSS Parsing
│   ├── tokenizer.go      # CSS token extraction
│   ├── parser.go         # Rule and selector parsing
│   └── values.go         # Color, font-size parsing
│
├── dom/                  # Document Object Model
│   ├── node.go           # Node types (Element, Text, Document)
│   ├── url.go            # URL resolution (RFC 3986, data URLs)
│   └── loader.go         # External resource fetching
│
├── style/                # Style Computation
│   └── style.go          # Cascade, specificity, inheritance
│
├── layout/               # Layout Calculation
│   └── layout.go         # Box model, formatting contexts
│
├── render/               # Rasterization
│   └── render.go         # Canvas drawing, text, images
│
├── svg/                  # SVG Support
│   ├── parser.go         # SVG element parsing
│   └── render.go         # SVG rasterization
│
├── font/                 # Font Management
│   └── font.go           # Embedded Go fonts, style handling
│
└── log/                  # Logging
    └── log.go            # Leveled debug logging
```

## Data Structures

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              DATA FLOW                                       │
└─────────────────────────────────────────────────────────────────────────────┘

1. DOM Tree (dom.Node)
   ┌─────────────────────────────────────┐
   │ *dom.Node                           │
   │ ├── Type: Element | Text | Document │
   │ ├── Data: tag name / text content   │
   │ ├── Attr: map[string]string         │
   │ ├── Parent: *Node                   │
   │ └── Children: []*Node               │
   └─────────────────────────────────────┘
                    │
                    ▼
2. Stylesheet (css.Stylesheet)
   ┌─────────────────────────────────────┐
   │ *css.Stylesheet                     │
   │ └── Rules: []Rule                   │
   │     ├── Selectors: []Selector       │
   │     │   └── Parts (tag/class/id)    │
   │     └── Declarations: []Declaration │
   │         ├── Property: string        │
   │         └── Value: string           │
   └─────────────────────────────────────┘
                    │
                    ▼
3. Styled Tree (style.StyledNode)
   ┌─────────────────────────────────────┐
   │ *style.StyledNode                   │
   │ ├── Node: *dom.Node                 │
   │ ├── Styles: map[string]string       │
   │ │   (computed CSS properties)       │
   │ └── Children: []*StyledNode         │
   └─────────────────────────────────────┘
                    │
                    ▼
4. Layout Tree (layout.LayoutBox)
   ┌─────────────────────────────────────┐
   │ *layout.LayoutBox                   │
   │ ├── BoxType: Block | Inline | Anon  │
   │ ├── Dimensions:                     │
   │ │   ├── Content: Rect (x,y,w,h)     │
   │ │   ├── Padding: EdgeSizes          │
   │ │   ├── Border: EdgeSizes           │
   │ │   └── Margin: EdgeSizes           │
   │ ├── StyledNode: *StyledNode         │
   │ └── Children: []*LayoutBox          │
   └─────────────────────────────────────┘
                    │
                    ▼
5. Canvas (render.Canvas)
   ┌─────────────────────────────────────┐
   │ *render.Canvas                      │
   │ ├── Pixels: *image.RGBA             │
   │ └── ImageCache: map[string]image    │
   └─────────────────────────────────────┘
```

## Style Cascade & Specificity

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         CSS SPECIFICITY ORDER                                │
└─────────────────────────────────────────────────────────────────────────────┘

Priority (lowest to highest):
┌────────────────────────────────────────────────────────────────────────────┐
│  1. User-Agent Styles (browser defaults)                                    │
│     └── Embedded in style/style.go                                         │
├────────────────────────────────────────────────────────────────────────────┤
│  2. Author Styles - by specificity (a, b, c, d):                           │
│     ├── (0,0,0,1) - Type selector:    div { }                              │
│     ├── (0,0,1,0) - Class selector:   .foo { }                             │
│     ├── (0,1,0,0) - ID selector:      #bar { }                             │
│     └── (1,0,0,0) - Inline styles:    style="..."                          │
├────────────────────────────────────────────────────────────────────────────┤
│  3. !important declarations (override all above)                            │
└────────────────────────────────────────────────────────────────────────────┘
```

## External Resource Loading

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         RESOURCE LOADING FLOW                                │
└─────────────────────────────────────────────────────────────────────────────┘

                    ┌───────────────────┐
                    │    HTML Source    │
                    └─────────┬─────────┘
                              │
          ┌───────────────────┼───────────────────┐
          │                   │                   │
          ▼                   ▼                   ▼
   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
   │ <link href> │    │ <style>     │    │ <img src>   │
   │ CSS files   │    │ inline CSS  │    │ images      │
   └──────┬──────┘    └──────┬──────┘    └──────┬──────┘
          │                  │                   │
          ▼                  │                   ▼
   ┌─────────────┐           │           ┌─────────────┐
   │ HTTP Fetch  │           │           │ HTTP Fetch  │
   │ (net/http)  │           │           │ or Data URL │
   └──────┬──────┘           │           └──────┬──────┘
          │                  │                   │
          ▼                  ▼                   ▼
   ┌──────────────────────────┐          ┌─────────────┐
   │      CSS Parser          │          │ Image Decode│
   │   (merged stylesheet)    │          │ PNG/JPG/GIF │
   └──────────────────────────┘          │ SVG/DataURL │
                                         └─────────────┘

URL Types Supported:
├── http:// / https:// - Network fetch via net/http
├── file:// - Local file read
└── data: - RFC 2397 data URLs (inline base64/text)
```

## Build Targets

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           BUILD TARGETS                                      │
└─────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────┐
│            CLI Binary                │
│     go build ./cmd/browser           │
│                                      │
│  Usage:                              │
│  ./browser -url https://example.com  │
│  ./browser -file page.html           │
│  ./browser -o output.png             │
└──────────────────────────────────────┘

┌──────────────────────────────────────┐
│         WebAssembly Module           │
│  GOOS=js GOARCH=wasm go build        │
│       -o browser.wasm ./wasm         │
│                                      │
│  Runs entirely in web browser        │
│  via wasm_exec.js                    │
└──────────────────────────────────────┘
```

## Dependencies

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          EXTERNAL DEPENDENCIES                               │
└─────────────────────────────────────────────────────────────────────────────┘

go.mod:
  github.com/lukehoban/browser
  └── golang.org/x/image v0.34.0    # Image decoding & font rendering
      └── golang.org/x/text v0.32.0  # Unicode/text processing (indirect)

Standard Library:
  ├── image, image/png, image/jpeg, image/gif  # Image handling
  ├── net/http, net/url                        # Network & URL parsing
  ├── fmt, strings, strconv, bytes             # Text processing
  └── os, flag                                 # CLI & file I/O
```

## CSS 2.1 Feature Support

| Feature | Status | Notes |
|---------|--------|-------|
| Box Model | ✅ | content, padding, border, margin |
| Block Layout | ✅ | Block formatting context |
| Inline Layout | ✅ | Inline formatting context |
| Table Layout | ✅ | Auto-sizing algorithm |
| Selectors | ✅ | element, class, ID, descendant |
| Cascade | ✅ | Specificity & source order |
| Inheritance | ✅ | font-*, color, text-* |
| Colors | ✅ | Named colors, hex, rgb() |
| Fonts | ✅ | size, weight, style |
| Backgrounds | ✅ | color, image |
| Borders | ✅ | width, style, color |
| Text | ✅ | align, decoration |
| Pseudo-classes | ✅ | :first-child, :visited, etc. |
| Pseudo-elements | ✅ | ::before, ::after |
