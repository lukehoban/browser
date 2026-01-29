# Browser Architecture

A web browser implementation in Go that renders HTML/CSS to PNG images, following CSS 2.1 specifications.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              INPUT SOURCES                                   │
│                    File Path / HTTP URL / Data URL                          │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         1. HTML PARSING (html/)                             │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────────────────────┐  │
│  │  Tokenizer  │ ───▶ │   Parser    │ ───▶ │    DOM Node Tree            │  │
│  │             │      │ (HTML5 spec)│      │ {Type, Data, Attrs, Kids}   │  │
│  └─────────────┘      └─────────────┘      └─────────────────────────────┘  │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │ dom.Node
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         2. CSS PARSING (css/)                               │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  Extract <style> tags & fetch <link> stylesheets                       │ │
│  └───────────────────────────────┬────────────────────────────────────────┘ │
│                                  ▼                                          │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────────────────────┐  │
│  │  Tokenizer  │ ───▶ │   Parser    │ ───▶ │    Stylesheet               │  │
│  │             │      │ (CSS 2.1)   │      │ [Rules → Selectors → Decls] │  │
│  └─────────────┘      └─────────────┘      └─────────────────────────────┘  │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │ css.Stylesheet
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                      3. STYLE COMPUTATION (style/)                          │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  Selector Matching (CSS 2.1 §6.4.1)                                    │ │
│  │       ↓                                                                 │ │
│  │  Specificity Calculation (a,b,c,d tuple)                               │ │
│  │       ↓                                                                 │ │
│  │  Cascade: user-agent → author CSS → inline styles                      │ │
│  │       ↓                                                                 │ │
│  │  Property Inheritance                                                   │ │
│  └───────────────────────────────┬────────────────────────────────────────┘ │
│                                  ▼                                          │
│              StyledNode tree with map[string]string Styles                  │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │ style.StyledNode
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        4. LAYOUT ENGINE (layout/)                           │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  Build LayoutBox tree (Block, Inline, Table, etc.)                     │ │
│  │       ↓                                                                 │ │
│  │  Calculate Box Model (content, padding, border, margin)                │ │
│  │       ↓                             ┌──────────────────┐                │ │
│  │  Width/Height Calculation ◀────────│    font/         │                │ │
│  │       ↓                             │ (text measuring) │                │ │
│  │  Positioning (normal flow)          └──────────────────┘                │ │
│  └───────────────────────────────┬────────────────────────────────────────┘ │
│                                  ▼                                          │
│              LayoutBox tree with Dimensions                                 │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │ layout.LayoutBox
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                       5. RENDERING ENGINE (render/)                         │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  Create Canvas (pixel buffer)                                          │ │
│  │       ↓                                                                 │ │
│  │  Draw Backgrounds & Borders                                            │ │
│  │       ↓                             ┌──────────────────┐                │ │
│  │  Render Text ◀─────────────────────│    font/         │                │ │
│  │       ↓                             │ (TrueType fonts) │                │ │
│  │  Load & Draw Images ◀──────────────│    dom/loader    │                │ │
│  │       ↓                             └──────────────────┘                │ │
│  │  SVG Rasterization ◀───────────────│    svg/          │                │ │
│  └───────────────────────────────┬────────────────────────────────────────┘ │
│                                  ▼                                          │
│              Canvas with []color.RGBA pixels                                │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            6. PNG OUTPUT                                    │
│                  Encode to PNG file or Base64 (WASM)                        │
└─────────────────────────────────────────────────────────────────────────────┘
```

## Module Overview

| Module | Purpose | Key Files |
|--------|---------|-----------|
| **html/** | HTML tokenization & parsing | `tokenizer.go`, `parser.go` |
| **css/** | CSS tokenization & parsing | `tokenizer.go`, `parser.go` |
| **dom/** | DOM tree structure & resource loading | `dom.go`, `loader.go`, `url.go` |
| **style/** | Style computation & cascade | `style.go`, `cascade.go` |
| **layout/** | Box model & positioning | `layout.go`, `dimensions.go` |
| **render/** | Canvas-based rasterization | `render.go` |
| **font/** | TrueType font management | `font.go` |
| **svg/** | SVG parsing & rasterization | `svg.go` |
| **cmd/browser/** | CLI application | `main.go` |
| **cmd/browser-wasm/** | WebAssembly entry point | `main.go` |

## Data Flow

```
┌──────────────┐    ┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│   HTML Text  │    │  dom.Node    │    │ StyledNode   │    │  LayoutBox   │
│              │ ─▶ │    Tree      │ ─▶ │    Tree      │ ─▶ │    Tree      │
│  + CSS Text  │    │              │    │  + Styles    │    │ + Dimensions │
└──────────────┘    └──────────────┘    └──────────────┘    └──────┬───────┘
                                                                   │
                                                                   ▼
                                                           ┌──────────────┐
                                                           │   Canvas     │
                                                           │   (pixels)   │
                                                           └──────┬───────┘
                                                                   │
                                                                   ▼
                                                           ┌──────────────┐
                                                           │  PNG Image   │
                                                           └──────────────┘
```

## CSS Cascade Priority (highest to lowest)

```
1. Inline style="" attribute        (specificity: 1,0,0,0)
2. Author CSS rules                 (specificity: varies)
3. User-agent default stylesheet    (specificity: varies)
```

## Resource Loading (dom/loader.go)

```
┌─────────────────┐
│  URL Reference  │
│ (relative/abs)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ URL Resolution  │
│ dom.ResolveURL  │
└────────┬────────┘
         │
    ┌────┴────┬─────────────┐
    ▼         ▼             ▼
┌───────┐ ┌───────┐   ┌──────────┐
│ File  │ │ HTTP/ │   │ data://  │
│ Path  │ │ HTTPS │   │   URL    │
└───────┘ └───────┘   └──────────┘
```

## External Dependencies

- `golang.org/x/image` - Font rendering (OpenType) and Go fonts
- `golang.org/x/text` - Text processing (indirect)
- Go stdlib: `net/http`, `image/*`, `encoding/base64`

## Entry Points

### CLI (cmd/browser/)
```bash
go run ./cmd/browser -url https://example.com -output page.png
go run ./cmd/browser -file index.html -output page.png
```

### WebAssembly (cmd/browser-wasm/)
Exposes `renderHTML(html, css, width, height)` to JavaScript, returns base64 PNG.
