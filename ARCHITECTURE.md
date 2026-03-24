# Architecture Documentation

This document provides detailed architecture diagrams for the browser implementation.

## High-Level Architecture

```mermaid
graph TB
    subgraph Input
        HTML[HTML File/URL]
        CSS[CSS Stylesheets]
        Images[Images<br/>PNG/JPEG/GIF/SVG]
    end
    
    subgraph "HTML Parser (html/)"
        Tokenizer[HTML Tokenizer<br/>tokenizer.go]
        Parser[HTML Parser<br/>parser.go]
    end
    
    subgraph "DOM Layer (dom/)"
        DOM[DOM Tree<br/>node.go]
        URLResolver[URL Resolver<br/>url.go]
        Loader[Resource Loader<br/>loader.go]
    end
    
    subgraph "CSS Parser (css/)"
        CSSTokenizer[CSS Tokenizer<br/>tokenizer.go]
        CSSParser[CSS Parser<br/>parser.go]
        Values[Value Parser<br/>values.go]
    end
    
    subgraph "Style Engine (style/)"
        Matcher[Selector Matcher]
        Cascade[Cascade & Specificity]
        StyledTree[Styled DOM Tree<br/>style.go]
        UA[User Agent Styles<br/>useragent.go]
    end
    
    subgraph "Layout Engine (layout/)"
        BoxBuilder[Box Tree Builder]
        BoxModel[Box Model Calculator]
        Dimensions[Dimensions Calculator<br/>layout.go]
    end
    
    subgraph "Rendering Engine (render/)"
        Canvas[Canvas & Pixel Buffer]
        TextRender[Text Renderer]
        ImageRender[Image Renderer]
        Painter[Painter<br/>render.go]
    end
    
    subgraph "Font System (font/)"
        GoFont[Go Fonts<br/>font.go]
    end
    
    subgraph "SVG Support (svg/)"
        SVGParser[SVG Parser<br/>svg.go]
        Rasterizer[SVG Rasterizer<br/>rasterizer.go]
    end
    
    subgraph Output
        PNG[PNG Image Output]
        WASM[WebAssembly Demo]
    end
    
    HTML --> Tokenizer
    Tokenizer --> Parser
    Parser --> DOM
    
    CSS --> CSSTokenizer
    CSSTokenizer --> CSSParser
    CSSParser --> Values
    
    DOM --> URLResolver
    URLResolver --> Loader
    Loader --> Images
    
    DOM --> Matcher
    Values --> Matcher
    UA --> Matcher
    Matcher --> Cascade
    Cascade --> StyledTree
    
    StyledTree --> BoxBuilder
    BoxBuilder --> BoxModel
    BoxModel --> Dimensions
    
    Dimensions --> Painter
    GoFont --> TextRender
    TextRender --> Painter
    Images --> ImageRender
    SVGParser --> Rasterizer
    Rasterizer --> ImageRender
    ImageRender --> Painter
    Painter --> Canvas
    
    Canvas --> PNG
    Canvas --> WASM
    
    style HTML fill:#e1f5ff
    style CSS fill:#e1f5ff
    style Images fill:#e1f5ff
    style PNG fill:#c8e6c9
    style WASM fill:#c8e6c9
```

## Rendering Pipeline

```mermaid
flowchart LR
    subgraph "1. Parse"
        A[HTML Input] --> B[Tokenizer]
        B --> C[Parser]
        C --> D[DOM Tree]
    end
    
    subgraph "2. Style"
        D --> E[CSS Parser]
        E --> F[Selector Matching]
        F --> G[Cascade]
        G --> H[Styled Tree]
    end
    
    subgraph "3. Layout"
        H --> I[Box Tree]
        I --> J[Calculate Dimensions]
        J --> K[Position Elements]
    end
    
    subgraph "4. Render"
        K --> L[Draw Backgrounds]
        L --> M[Draw Borders]
        M --> N[Draw Text]
        N --> O[Draw Images]
        O --> P[PNG Output]
    end
    
    style A fill:#e1f5ff
    style P fill:#c8e6c9
```

## Data Structures Flow

```mermaid
graph TD
    subgraph "DOM Layer"
        Node["Node<br/>- Type: Element/Text<br/>- Tag Name<br/>- Attributes<br/>- Children"]
    end
    
    subgraph "Style Layer"
        StyledNode["StyledNode<br/>- Node Reference<br/>- Property Map<br/>- Computed Styles"]
        Rule["CSS Rule<br/>- Selectors<br/>- Declarations<br/>- Specificity"]
    end
    
    subgraph "Layout Layer"
        LayoutBox["LayoutBox<br/>- Box Type<br/>- Dimensions<br/>- Children"]
        Dims["Dimensions<br/>- Content Box<br/>- Padding<br/>- Border<br/>- Margin"]
    end
    
    subgraph "Render Layer"
        Canvas["Canvas<br/>- Width/Height<br/>- Pixel Buffer<br/>- Color Array"]
    end
    
    Node --> StyledNode
    Rule --> StyledNode
    StyledNode --> LayoutBox
    LayoutBox --> Dims
    Dims --> Canvas
    
    style Node fill:#fff3cd
    style StyledNode fill:#d4edda
    style LayoutBox fill:#d1ecf1
    style Canvas fill:#f8d7da
```

## CSS Box Model Implementation

```mermaid
graph TB
    subgraph "Box Model (CSS 2.1 §8)"
        Margin["Margin<br/>(Transparent)"]
        Border["Border<br/>(Colored Rectangle)"]
        Padding["Padding<br/>(Background Extends)"]
        Content["Content<br/>(Text & Images)"]
    end
    
    Margin --> Border
    Border --> Padding
    Padding --> Content
    
    Properties["CSS Properties:<br/>• margin-top/right/bottom/left<br/>• border-width, border-color<br/>• padding-top/right/bottom/left<br/>• width, height<br/>• background-color"]
    
    Properties -.-> Margin
    Properties -.-> Border
    Properties -.-> Padding
    Properties -.-> Content
    
    Calc["Layout Calculations:<br/>1. Calculate width<br/>2. Calculate horizontal margins<br/>3. Calculate height<br/>4. Position in containing block"]
    
    Content --> Calc
    
    style Margin fill:#fff3cd
    style Border fill:#f8d7da
    style Padding fill:#d4edda
    style Content fill:#d1ecf1
```

## Module Dependencies

```mermaid
graph TD
    Main["cmd/browser<br/>main.go"]
    HTML["html/<br/>Tokenizer & Parser"]
    CSS["css/<br/>Tokenizer & Parser"]
    DOM["dom/<br/>Node & Tree"]
    Style["style/<br/>Style Computation"]
    Layout["layout/<br/>Layout Engine"]
    Render["render/<br/>Rendering Engine"]
    Font["font/<br/>Go Fonts"]
    SVG["svg/<br/>SVG Support"]
    Log["log/<br/>Logging"]
    
    Main --> HTML
    Main --> CSS
    Main --> Style
    Main --> Layout
    Main --> Render
    
    HTML --> DOM
    CSS --> DOM
    Style --> DOM
    Style --> CSS
    Layout --> Style
    Render --> Layout
    Render --> Font
    Render --> SVG
    Render --> DOM
    
    HTML -.-> Log
    CSS -.-> Log
    Style -.-> Log
    Layout -.-> Log
    Render -.-> Log
    
    style Main fill:#e1f5ff
    style Log fill:#f0f0f0
```

## Network & Resource Loading

```mermaid
sequenceDiagram
    participant User
    participant Main as cmd/browser
    participant Loader as dom.Loader
    participant Net as Network/FileSystem
    participant Parser as HTML/CSS Parser
    participant Render as Render Engine
    
    User->>Main: URL/File Path
    Main->>Loader: LoadDocument(url)
    Loader->>Net: HTTP GET / File Read
    Net-->>Loader: HTML Content
    Loader->>Parser: Parse HTML
    Parser-->>Loader: DOM Tree
    
    Parser->>Parser: Find <link> tags
    loop For each <link>
        Parser->>Loader: LoadStylesheet(href)
        Loader->>Net: HTTP GET / File Read
        Net-->>Loader: CSS Content
        Loader->>Parser: Parse CSS
        Parser-->>Loader: CSS Rules
    end
    
    Parser->>Parser: Find <img> tags
    loop For each <img>
        Parser->>Loader: LoadImage(src)
        Loader->>Net: HTTP GET / File Read
        Net-->>Loader: Image Data
    end
    
    Loader-->>Main: Complete DOM + Resources
    Main->>Render: Render(dom, styles, images)
    Render-->>User: PNG Output
```

## WebAssembly Architecture

```mermaid
graph TB
    subgraph "Browser Environment"
        HTML_Page[HTML Demo Page<br/>wasm/index.html]
        JS[JavaScript Bridge<br/>wasm_exec.js]
        WASM[Go WASM Binary<br/>browser.wasm]
    end
    
    subgraph "Go Code"
        Main_WASM[cmd/browser-wasm<br/>main.go]
        Core[Core Browser Engine<br/>Shared with CLI]
    end
    
    subgraph "User Interaction"
        Input[User Input:<br/>• HTML content<br/>• CSS styles<br/>• Viewport size]
        Output[Canvas Output:<br/>• Rendered PNG<br/>• Base64 Data URL]
    end
    
    HTML_Page --> JS
    JS --> WASM
    WASM --> Main_WASM
    Main_WASM --> Core
    
    Input --> HTML_Page
    Core --> Output
    Output --> HTML_Page
    
    style WASM fill:#e1f5ff
    style Core fill:#d4edda
```

## Selector Matching Algorithm

```mermaid
flowchart TD
    Start([Start: Match Selector to Element]) --> HasID{Has ID<br/>Selector?}
    
    HasID -->|Yes| CheckID{Element ID<br/>matches?}
    CheckID -->|No| NoMatch([No Match])
    CheckID -->|Yes| HasTag
    
    HasID -->|No| HasTag{Has Tag<br/>Selector?}
    
    HasTag -->|Yes| CheckTag{Tag name<br/>matches?}
    CheckTag -->|No| NoMatch
    CheckTag -->|Yes| HasClasses
    
    HasTag -->|No| HasClasses{Has Class<br/>Selectors?}
    
    HasClasses -->|Yes| CheckClasses{All classes<br/>present?}
    CheckClasses -->|No| NoMatch
    CheckClasses -->|Yes| HasDescendant
    
    HasClasses -->|No| HasDescendant{Has Descendant<br/>Combinator?}
    
    HasDescendant -->|Yes| CheckAncestor{Ancestor<br/>matches?}
    CheckAncestor -->|No| NoMatch
    CheckAncestor -->|Yes| CalcSpec
    
    HasDescendant -->|No| CalcSpec[Calculate Specificity:<br/>IDs × 100<br/>Classes × 10<br/>Tags × 1]
    
    CalcSpec --> Match([Match!<br/>Apply Rule])
    
    style Start fill:#e1f5ff
    style Match fill:#c8e6c9
    style NoMatch fill:#f8d7da
```

## Color & Font Rendering

```mermaid
graph TB
    subgraph "Color System"
        ColorInput["Color Value:<br/>• Named: 'red', 'blue'<br/>• Hex: #RGB, #RRGGBB<br/>• RGB: rgb(r,g,b)"]
        ColorParser[Parse Color<br/>css/values.go]
        RGBA[RGBA Struct<br/>R, G, B, A bytes]
    end
    
    subgraph "Font System"
        FontStyle["Font Properties:<br/>• font-weight: bold<br/>• font-style: italic<br/>• font-size: px<br/>• text-decoration"]
        FontFace[Go Fonts<br/>font/font.go]
        FontDrawer[Font Drawer<br/>Fixed/Regular/Bold/Italic]
    end
    
    subgraph "Text Rendering"
        TextNode[Text Content]
        TextLayout[Calculate Text Position<br/>& Baseline]
        TextDraw[Draw Glyphs to Canvas]
    end
    
    ColorInput --> ColorParser
    ColorParser --> RGBA
    
    FontStyle --> FontFace
    FontFace --> FontDrawer
    
    TextNode --> TextLayout
    TextLayout --> TextDraw
    FontDrawer --> TextDraw
    RGBA --> TextDraw
    
    TextDraw --> Canvas[Canvas Pixel Buffer]
    
    style Canvas fill:#c8e6c9
```

## Image Support Architecture

```mermaid
graph TD
    subgraph "Image Sources"
        File[Local File Path]
        URL[HTTP/HTTPS URL]
        DataURL[Data URL<br/>RFC 2397]
    end
    
    subgraph "Image Formats"
        PNG[PNG Decoder]
        JPEG[JPEG Decoder]
        GIF[GIF Decoder]
        SVG[SVG Parser &<br/>Rasterizer]
    end
    
    subgraph "Processing"
        Decode[Decode Image]
        Cache[Image Cache]
        Scale[Scale to CSS Size]
        Alpha[Alpha Blending]
    end
    
    subgraph "Output"
        Composite[Composite on Canvas]
    end
    
    File --> Decode
    URL --> Decode
    DataURL --> Decode
    
    Decode --> PNG
    Decode --> JPEG
    Decode --> GIF
    Decode --> SVG
    
    PNG --> Cache
    JPEG --> Cache
    GIF --> Cache
    SVG --> Cache
    
    Cache --> Scale
    Scale --> Alpha
    Alpha --> Composite
    
    style Composite fill:#c8e6c9
```

## Testing Architecture

```mermaid
graph TB
    subgraph "Unit Tests"
        HTMLTest[html/*_test.go<br/>Tokenizer & Parser]
        CSSTest[css/*_test.go<br/>Tokenizer & Parser]
        DOMTest[dom/*_test.go<br/>Node & URL Resolution]
        StyleTest[style/*_test.go<br/>Selector Matching]
        LayoutTest[layout/*_test.go<br/>Box Model]
        RenderTest[render/*_test.go<br/>Rendering]
    end
    
    subgraph "Integration Tests"
        TestFiles[test/*.html<br/>Real HTML Pages]
        Screenshots[Visual Screenshots<br/>*.png]
    end
    
    subgraph "Reference Tests"
        WPT[Web Platform Tests<br/>test/wpt/]
        RefTest[reftest.go<br/>WPT Runner]
    end
    
    subgraph "CI/CD"
        GHA[GitHub Actions<br/>.github/workflows/ci.yml]
        Coverage[Code Coverage<br/>90%+ target]
    end
    
    HTMLTest --> GHA
    CSSTest --> GHA
    DOMTest --> GHA
    StyleTest --> GHA
    LayoutTest --> GHA
    RenderTest --> GHA
    
    TestFiles --> GHA
    Screenshots --> GHA
    
    WPT --> RefTest
    RefTest --> GHA
    
    GHA --> Coverage
    
    style Coverage fill:#c8e6c9
```

## Key Design Patterns

### 1. Pipeline Pattern
The browser uses a sequential pipeline where each stage transforms data and passes it to the next stage:
- **HTML** → **DOM** → **Styled Tree** → **Layout Tree** → **Canvas** → **PNG**

### 2. Visitor Pattern
The style matching algorithm walks the DOM tree and applies CSS rules to each node.

### 3. Composite Pattern
Both DOM and Layout trees use composite pattern where nodes can contain children.

### 4. Strategy Pattern
Different rendering strategies for different element types (block vs inline, text vs image).

### 5. Cache Pattern
Images and fonts are cached to avoid redundant loading and parsing.

## Performance Characteristics

| Component | Time Complexity | Space Complexity | Notes |
|-----------|----------------|------------------|-------|
| HTML Parsing | O(n) | O(n) | n = input size |
| CSS Parsing | O(m) | O(m) | m = stylesheet size |
| Style Matching | O(n × r) | O(n) | n = DOM nodes, r = CSS rules |
| Layout Calculation | O(n) | O(n) | Single-pass tree traversal |
| Rendering | O(p) | O(w × h) | p = pixels, w×h = canvas size |

## Specification Compliance Matrix

| Specification | Status | Coverage |
|--------------|--------|----------|
| HTML5 §12.2 Tokenization | ✅ Partial | Common states implemented |
| HTML5 §12.2 Tree Construction | ✅ Simplified | Basic algorithm implemented |
| CSS 2.1 §4 Syntax | ✅ Complete | Full tokenization |
| CSS 2.1 §5 Selectors | ✅ Partial | Element, class, ID, descendant |
| CSS 2.1 §6 Cascade | ✅ Partial | Specificity only |
| CSS 2.1 §8 Box Model | ✅ Complete | Full implementation |
| CSS 2.1 §9 Visual Formatting | ✅ Partial | Block layout only |
| CSS 2.1 §10 Width/Height | ✅ Complete | Auto and fixed values |
| CSS 2.1 §14 Colors/Backgrounds | ✅ Complete | Colors and backgrounds |
| RFC 2397 Data URLs | ✅ Complete | Base64 and percent-encoded |

## Future Architecture Considerations

### Potential Extensions
1. **Inline Layout**: Full inline formatting context with text wrapping
2. **Floats**: CSS float positioning
3. **Flexbox**: CSS Flexible Box Layout Module
4. **Grid**: CSS Grid Layout Module
5. **JavaScript Engine**: JS execution and DOM manipulation
6. **Incremental Rendering**: Progressive display during loading
7. **GPU Acceleration**: Hardware-accelerated rendering

### Scalability Improvements
1. **Parallel Parsing**: Multi-threaded HTML/CSS parsing
2. **Layout Optimization**: Incremental layout invalidation
3. **Render Layers**: Compositor with layer trees
4. **Smart Caching**: Persistent cache across renders

## References

- **HTML5 Specification**: https://html.spec.whatwg.org/
- **CSS 2.1 Specification**: https://www.w3.org/TR/CSS21/
- **RFC 2397 (Data URLs)**: https://datatracker.ietf.org/doc/html/rfc2397
- **Web Platform Tests**: https://github.com/web-platform-tests/wpt
- **Go Fonts**: https://go.googlesource.com/image/+/refs/heads/master/font/gofont/
