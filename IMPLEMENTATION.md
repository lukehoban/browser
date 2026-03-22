# Implementation Summary

This document provides a technical summary of the browser implementation, detailing the architecture and design decisions.

## Architecture Overview

The browser follows a classic rendering pipeline:

```
HTML Input → Tokenization → DOM Tree ──→ URL Resolution ──→ CSS Extraction ──→ CSS Parsing
                                │                                   ↑
                                │                          External stylesheets
                                ↓                          (<link rel="stylesheet">)
                          Style Computation (cascade, inheritance, specificity)
                                ↓
                          Layout Engine (box model, visual formatting model)
                                ↓
                          Render Engine → PNG Output
```

Supporting modules: `font` (TrueType Go fonts), `svg` (parse and rasterize), `log` (structured logging).

### Key Components

1. **HTML Parser** (`html/`): Tokenizes and parses HTML into a DOM tree
2. **CSS Parser** (`css/`): Tokenizes and parses CSS stylesheets, including shorthand expansion
3. **DOM** (`dom/`): DOM tree structure, URL resolution, and resource loading
4. **Style Engine** (`style/`): Matches selectors, computes cascade, applies inheritance
5. **Layout Engine** (`layout/`): Calculates box model and positions
6. **Render Engine** (`render/`): Draws to canvas and outputs PNG
7. **Font** (`font/`): Embedded Go TrueType fonts (regular, bold, italic, bold-italic)
8. **SVG** (`svg/`): SVG parsing and rasterization
9. **Log** (`log/`): Structured logging with configurable levels

## Implementation Details

### 1. HTML Parser
**Files**: `html/tokenizer.go`, `html/parser.go`, `dom/node.go`  
**Specification**: HTML5 §12.2 Parsing HTML documents

**Key Features**:
- State machine-based tokenizer following HTML5 §12.2.5
- Tree construction with proper element nesting
- Support for start tags, end tags, self-closing tags, and void elements
- Attribute parsing with quoted and unquoted values
- Text node and comment handling
- HTML character entity references: 70+ named entities (`&amp;`, `&lt;`, `&nbsp;`, `&copy;`, etc.), decimal (`&#60;`), and hexadecimal (`&#x3C;`) numeric references

**Design Decisions**:
- Simplified error recovery for educational clarity
- No namespace support (SVG/MathML in HTML context)
- Focus on common HTML structure rather than edge cases

**Test Coverage**: 90.1%

### 2. CSS Parser
**Files**: `css/tokenizer.go`, `css/parser.go`, `css/values.go`  
**Specification**: CSS 2.1 §4 Syntax and basic data types

**Key Features**:
- Tokenization of identifiers, strings, numbers, hash values, and punctuation
- Simple selectors: element, class (`.class`), ID (`#id`)
- Combined selectors: `div#id.class1.class2`
- Descendant combinators: `div p`, `body div.content`
- Declaration parsing: property names and values
- Multiple selectors: `h1, h2, h3 { color: blue; }`
- Pseudo-classes: `:link`, `:visited`, `:hover` (counted in specificity)
- **Shorthand property expansion**: `margin`, `padding`, `border`, `font`, `background`
- `!important` detection (detected and warned; not applied to cascade priority)

**Design Decisions**:
- Shorthands are expanded at parse time into their longhand properties
- Child/sibling combinators (`>`, `+`, `~`) detected but not yet implemented
- Attribute selectors (`[attr=value]`) detected but not yet implemented

**Test Coverage**: 92.4%

### 3. Style Computation
**Files**: `style/style.go`, `style/useragent.go`  
**Specification**: CSS 2.1 §6 Assigning property values, Cascading, and Inheritance

**Key Features**:
- Selector matching algorithm walks DOM tree
- Specificity calculation per CSS 2.1 §6.4.3:
  - Count ID selectors (a)
  - Count class selectors (b)
  - Count type selectors (c)
  - Specificity = (a, b, c)
- Cascade implementation sorts rules by specificity
- Style properties stored in map per element
- **CSS inheritance** (CSS 2.1 §6.2): inherited properties (`color`, `font-size`, `font-weight`, `font-style`, `line-height`, `text-align`, etc.) propagate from parent to child
- **Inline styles** (`style="..."` attribute) applied with highest specificity
- **User agent stylesheet** (`style/useragent.go`): default styles for all common HTML elements
- **HTML presentational attributes** (`align`, `valign`, `width`, `height`, `bgcolor`) honored
- **`<center>` element** and `align` attribute support

**Design Decisions**:
- Cascade uses specificity and origin (author styles override UA styles)
- Inline styles override all author stylesheet rules
- `!important` is detected but does not currently override cascade order

**Test Coverage**: 91.5%

### 4. Layout Engine
**Files**: `layout/layout.go`  
**Specification**: CSS 2.1 §8 Box model, §9 Visual formatting model, §10 Details

**Key Features**:
- Box model implementation (CSS 2.1 §8.1):
  - Content box
  - Padding (top, right, bottom, left)
  - Border (width on all sides)
  - Margin (top, right, bottom, left)
- Block-level layout (CSS 2.1 §9.2)
- Inline layout (CSS 2.1 §9.2.2): inline elements flowed within anonymous block boxes
- Width calculation (CSS 2.1 §10.3.3):
  - Auto width: fills containing block
  - Fixed width: respects specified value
  - Percentage width: relative to containing block
- Height calculation (CSS 2.1 §10.6.3):
  - Auto height: sum of children's heights
  - Fixed height: respects specified value
- Position calculation in normal flow
- **Table layout** (CSS 2.1 §17): `display: table`, `table-row`, `table-cell`; colspan support; content-based column sizing
- **Baseline alignment** for inline elements

**Data Structures**:
```go
type LayoutBox struct {
    BoxType     BoxType        // Block, Inline, Anonymous, Table, TableRow, TableCell
    Dimensions  Dimensions     // Position and size
    StyledNode  *StyledNode    // Reference to styled DOM node
    Children    []*LayoutBox   // Child boxes
}

type Dimensions struct {
    Content Rect      // Content box
    Padding EdgeSize  // Padding on all sides
    Border  EdgeSize  // Border on all sides
    Margin  EdgeSize  // Margin on all sides
}
```

**Design Decisions**:
- Normal flow only (floats and positioned layout not yet implemented)
- Table layout uses a two-pass approach: measure column widths, then lay out cells

**Test Coverage**: 90.0%

### 5. Rendering Engine
**Files**: `render/render.go`  
**Specification**: CSS 2.1 §14 Colors and backgrounds, §16 Text

**Key Features**:
- Canvas-based rendering with pixel buffer
- Background color rendering (CSS 2.1 §14.2)
- Background image rendering (CSS 2.1 §14.2.1): `background-image: url(...)` with `background-repeat`
- Border rendering (CSS 2.1 §8.5):
  - Solid borders with specified width and color
  - Rendered as rectangles around content+padding
- **Text rendering** using embedded Go fonts (TrueType, proportional):
  - Regular, bold, italic, bold-italic variants
  - Variable font sizes via `font-size` (px, em, pt)
  - `text-decoration: underline`
  - Text color via `color` property
  - Baseline alignment for inline elements
- Color parsing:
  - Named colors (150+ standard CSS color names)
  - Hex colors (`#RGB`, `#RRGGBB`)
  - `rgb()` and `rgba()` functions
- **Image rendering**: PNG, JPEG, GIF with scaling and alpha blending
- **SVG rendering**: SVG images rasterized via `svg/` package
- **Data URLs**: RFC 2397 inline image data (`data:image/png;base64,...`)
- PNG output via `image/png` package

**Design Decisions**:
- Go fonts (from `golang.org/x/image/font/gofont`) embedded in binary — no system fonts required
- Simple raster graphics; subpixel rendering not implemented
- Image cache prevents redundant decoding during a single render pass

**Test Coverage**: 90%+

### 6. Image & Network Loading
**Files**: `render/render.go`, `dom/url.go`, `dom/loader.go`, `svg/`  
**Specification**: HTML5 §2.5 URLs, §4.8.2 The img element, RFC 2397

**Key Features**:
- URL resolution following HTML5 §2.5:
  - Relative URLs resolved against document base (file directory or HTTP URL)
  - Absolute paths used as-is
  - `dom.ResolveURLs()` called after parsing
- Image loading and caching:
  - Loads PNG, JPEG, GIF via Go's `image` package
  - SVG via the `svg/` package
  - RFC 2397 data URLs (`data:image/png;base64,...`) decoded inline
  - HTTP/HTTPS remote images fetched via `net/http`
  - Cache prevents redundant I/O
- Image rendering:
  - Scales to CSS-defined width/height
  - Simple nearest-neighbor scaling
  - Alpha blending for transparency
  - Safe pixel access with bounds checking

**Architecture**:
```
HTML Parse → DOM Tree → URL Resolution (dom.ResolveURLs)
                              ↓
                     Converts relative paths to absolute URLs
                     Resolves data: URLs and HTTP/HTTPS URLs
                              ↓
Style → Layout → Render (loads from resolved URLs, caches decoded images)
```

**Design Decisions**:
- URL resolution in DOM layer (separation of concerns)
- Data URL decoding in `dom/loader.go` (RFC 2397 §2)
- Simple scaling algorithm (sufficient for typical use)

**Test Coverage**: Unit tests for URL resolution, data URL decoding, and image rendering

## Specification Compliance

### HTML5 Compliance
**Implemented**:
- §12.2.5 Tokenization (partial — common states)
- §12.2.6 Tree construction (simplified algorithm)
- §12.1.2 Void elements (img, br, hr, etc.)
- §2.5 URLs (relative URL resolution and HTTP/HTTPS fetching)
- §4.8.2 The img element (PNG, JPEG, GIF, SVG)
- Character references: named entities (70+), decimal (`&#60;`), hexadecimal (`&#x3C;`)

**Not Implemented**:
- Advanced error recovery
- Script/style CDATA sections
- Namespaces (SVG/MathML within HTML context)

### CSS 2.1 Compliance
**Implemented**:
- §4.1 Syntax (tokenization, identifiers, strings, numbers)
- §4.1.7 Rule sets, §4.1.8 Declarations
- §4.3.2 Lengths (px, em, pt, %)
- §5.2 Selector syntax (element, class, ID, combined)
- §5.5 Descendant selectors
- §5.11.2 `:link` / `:visited` pseudo-classes (link matching)
- §6.2 Inheritance (color, font-size, font-weight, font-style, line-height, text-align, etc.)
- §6.4.3 Specificity calculation
- §8.1 Box dimensions (content, padding, border, margin)
- §8.5 Border properties
- §9.2 Block and inline formatting contexts
- §9.4 Normal flow layout
- §10.3.3 Block-level non-replaced elements (width)
- §10.6.3 Block-level non-replaced elements (height)
- §14.2 Background color and background-image
- §16 Text rendering (size, weight, style, decoration, alignment)
- §17 Table layout (display: table, table-row, table-cell; colspan)
- Shorthand properties: `margin`, `padding`, `border`, `font`, `background`

**Not Implemented / Partial**:
- Pseudo-classes `:hover`, `:focus`, `:active`, `:first-child`, etc.
- Pseudo-elements (`::before`, `::after`)
- Attribute selectors (`[attr="value"]`)
- Child/sibling combinators (`>`, `+`, `~`)
- `!important` declarations (detected, not applied to cascade priority)
- Computed value calculation (values used as specified)
- Positioning schemes (absolute, relative, fixed)
- Floats
- Table rowspan

## Testing Strategy

### Unit Tests
Each module has comprehensive unit tests:
- `dom/node_test.go` - DOM tree operations
- `html/tokenizer_test.go` - HTML tokenization
- `html/parser_test.go` - HTML parsing
- `css/tokenizer_test.go` - CSS tokenization
- `css/parser_test.go` - CSS parsing
- `style/style_test.go` - Selector matching and specificity
- `layout/layout_test.go` - Box model calculations

**Coverage**: 90%+ across all modules

### Integration Tests
Test HTML files in `test/` directory:
- `simple.html` - Basic HTML structure
- `styled.html` - HTML with embedded CSS
- `hackernews.html` - Complex layout with multiple elements and image

### Reference Tests (WPT)
Web Platform Tests integration via `reftest/`:
- CSS box model tests
- CSS cascade tests
- CSS selector tests
- **Current pass rate**: 81.8% (WPT CSS test suite)
- See [TESTING.md](TESTING.md) for detailed per-test results

See [TESTING.md](TESTING.md) for detailed test results and methodology.

## Code Organization

```
browser/
├── cmd/
│   ├── browser/       # Main CLI application
│   ├── browser-wasm/  # WebAssembly entry point
│   └── wptrunner/     # WPT reftest runner
├── dom/               # DOM tree data structure
│   ├── node.go        # Node type, element/text nodes
│   ├── url.go         # URL resolution (HTML5 §2.5)
│   └── loader.go      # Network loader, data URL parser (RFC 2397)
├── html/              # HTML parsing
│   ├── tokenizer.go   # HTML tokenization (character entity support)
│   └── parser.go      # Tree construction
├── css/               # CSS parsing
│   ├── tokenizer.go   # CSS tokenization
│   ├── parser.go      # Selector and declaration parsing (shorthand expansion)
│   └── values.go      # Value parsing helpers
├── style/             # Style computation
│   ├── style.go       # Selector matching, specificity, cascade, inheritance
│   └── useragent.go   # Default user-agent stylesheet
├── layout/            # Layout engine
│   └── layout.go      # Box model, dimensions, block/inline/table layout
├── render/            # Rendering
│   └── render.go      # Canvas, drawing, PNG output
├── font/              # Font handling
│   └── font.go        # Embedded Go TrueType fonts (regular, bold, italic)
├── svg/               # SVG support
│   ├── svg.go         # SVG parsing
│   └── rasterizer.go  # SVG to raster conversion
├── log/               # Logging
│   └── logger.go      # Structured logging with configurable levels
├── reftest/           # Reference test framework
│   ├── reftest.go     # WPT reftest harness
│   └── wpt_test.go    # WPT integration tests
├── wasm/              # WebAssembly demo
│   └── index.html     # Interactive web demo page
└── test/              # Test HTML files and W3C WPT fixtures
    ├── styled.html
    ├── table_test.html
    ├── data_url_test.html
    └── wpt/           # W3C Web Platform Tests (CSS)
```

## Design Principles

1. **Specification-driven**: All implementations cite relevant W3C spec sections
2. **Simplicity over completeness**: Focus on core concepts, omit edge cases
3. **Educational clarity**: Code structure mirrors spec organization
4. **Modular design**: Clear separation between parsing, styling, layout, rendering
5. **Testable**: High unit test coverage, integration tests, reference tests
6. **Incremental**: Built in layers from bottom up

## Performance Considerations

- **Parsing**: Single-pass tokenization and tree construction
- **Styling**: O(n*m) where n=DOM nodes, m=CSS rules (acceptable for small documents)
- **Layout**: Single-pass tree traversal
- **Rendering**: Direct pixel buffer manipulation, no retained mode
- **Image caching**: Images loaded once and cached for entire render

## Conclusion

This browser demonstrates fundamental web rendering concepts with clean, specification-driven code. While simplified compared to production browsers, it successfully:

- ✅ Parses real HTML and CSS following W3C specifications
- ✅ Decodes HTML character entity references (70+ named entities, numeric)
- ✅ Implements CSS cascade, inheritance, and shorthand expansion
- ✅ Calculates accurate layout per CSS box model (block, inline, table)
- ✅ Renders visual output with high-quality Go fonts, colors, borders, and images
- ✅ Loads pages and resources over HTTP/HTTPS
- ✅ Supports RFC 2397 data URLs for inline resources
- ✅ Rasterizes SVG images
- ✅ Compiles to WebAssembly for in-browser use
- ✅ Achieves 81.8% W3C Web Platform Test pass rate
- ✅ Maintains high test coverage (90%+ per module)
- ✅ Provides educational value through clear, spec-driven code organization

For current status and progress tracking, see [MILESTONES.md](MILESTONES.md).
