# Browser Implementation Milestones

## Overview
This document tracks the milestones for implementing a simple web browser in Go, focusing on static HTML and CSS 2.1 compliance.

**Important**: Keep this document up to date as features are added or modified. When implementing new features, mark the corresponding tasks as complete and update the validation status.

---

## Milestone 1: Foundation (Initial Setup) ✅ COMPLETE
**Goal**: Set up project structure and basic architecture

### Tasks:
- [x] Initialize Go module
- [x] Create project directory structure
- [x] Document milestones
- [x] Add .gitignore

### Deliverables:
- ✅ Go module initialized
- ✅ Clear project structure
- ✅ Documentation framework

---

## Milestone 2: HTML Tokenization & Parsing ✅ COMPLETE
**Goal**: Parse static HTML into a DOM tree

**Spec References**: 
- HTML5 §12.2 Parsing HTML documents
- HTML5 §12.2.5 Tokenization

### Tasks:
- [x] Implement HTML tokenizer
  - [x] Data state
  - [x] Tag open/close states
  - [x] Basic character handling
- [x] Build DOM tree structure
  - [x] Element nodes
  - [x] Text nodes
  - [x] Attribute nodes
- [x] Parse common HTML elements (div, p, span, h1-h6, a, img, etc.)
- [x] Add unit tests with sample HTML

### Deliverables:
- ✅ HTML tokenizer that produces tokens from HTML strings
- ✅ DOM tree builder that constructs a tree from tokens
- ✅ Test suite validating parsing of basic HTML documents

### Validation:
- ✅ Parse simple HTML documents successfully
- ✅ Handle nested elements correctly
- ✅ Preserve text content and attributes

### Known Limitations:
- ⚠️ No character reference support (`&amp;`, `&lt;`, etc.)
- ⚠️ Simplified error recovery
- ⚠️ No script/style CDATA sections
- ⚠️ No namespace support

---

## Milestone 3: CSS Parsing ✅ COMPLETE
**Goal**: Parse CSS 2.1 stylesheets

**Spec References**:
- CSS 2.1 §4 Syntax and basic data types
- CSS 2.1 §5 Selectors
- CSS 2.1 §6 Assigning property values

### Tasks:
- [x] Implement CSS tokenizer
  - [x] Identifiers, strings, numbers
  - [x] Operators and delimiters
- [x] Parse selectors
  - [x] Type selectors (element)
  - [x] Class selectors (.class)
  - [x] ID selectors (#id)
  - [x] Descendant combinators
- [x] Parse declarations
  - [x] Property names
  - [x] Values (colors, lengths, keywords)
- [x] Build stylesheet structure

### Deliverables:
- ✅ CSS tokenizer
- ✅ CSS parser producing stylesheet objects
- ✅ Support for basic selectors and properties

### Validation:
- ✅ Parse CSS rules correctly
- ✅ Handle multiple selectors
- ✅ Parse common properties (color, font-size, margin, padding, border)

### Known Limitations:
- ⚠️ No pseudo-classes (`:hover`, `:first-child`)
- ⚠️ No pseudo-elements (`::before`, `::after`)
- ⚠️ No attribute selectors (`[attr="value"]`)
- ⚠️ No child/adjacent sibling combinators (`>`, `+`, `~`)
- ⚠️ No shorthand property expansion (e.g., `margin: 10px` → individual sides)

---

## Milestone 4: Style Computation ✅ COMPLETE
**Goal**: Match CSS rules to DOM elements and compute styles

**Spec References**:
- CSS 2.1 §6.4 The cascade
- CSS 2.1 §6.1 Specified, computed, and actual values
- CSS 2.1 §6.4.3 Specificity

### Tasks:
- [x] Implement selector matching algorithm
- [x] Calculate selector specificity (CSS 2.1 §6.4.3)
- [x] Implement cascade by specificity
- [x] Basic style property application

### Deliverables:
- ✅ Style computation engine
- ✅ Styled DOM tree with computed styles

### Validation:
- ✅ Correct selector matching
- ✅ Proper cascade order by specificity
- ✅ Descendant selectors work correctly

### Known Limitations:
- ⚠️ No inheritance implementation
- ⚠️ No `!important` support
- ⚠️ No computed value calculation (values used as-is)

---

## Milestone 5: Layout Engine ✅ COMPLETE
**Goal**: Implement CSS 2.1 visual formatting model

**Spec References**:
- CSS 2.1 §8 Box model
- CSS 2.1 §9 Visual formatting model
- CSS 2.1 §10 Visual formatting model details

### Tasks:
- [x] Implement box model (content, padding, border, margin)
- [x] Block formatting context
- [x] Normal flow layout
- [x] Width and height calculations (auto, px, %)

### Deliverables:
- ✅ Layout engine producing positioned boxes
- ✅ Support for block-level elements
- ✅ Box model with content, padding, border, margin

### Validation:
- ✅ Correct box dimensions
- ✅ Proper positioning of elements
- ✅ Margins, padding, borders applied correctly

### Known Limitations:
- ⚠️ Limited inline layout support
- ⚠️ No positioning schemes (absolute, relative, fixed)
- ⚠️ No float support
- ⚠️ No flexbox or grid layout

---

## Milestone 6: Rendering ✅ COMPLETE
**Goal**: Render the laid-out page

**Spec References**:
- CSS 2.1 §14 Colors and backgrounds
- CSS 2.1 §16 Text

### Tasks:
- [x] Implement display list generation
- [x] Render backgrounds and borders
- [x] Render text content
- [x] Output to PNG image format

### Deliverables:
- ✅ Basic renderer with text support
- ✅ Visual output of simple pages
- ✅ Color support for text and backgrounds
- ✅ Border rendering

### Validation:
- ✅ Rendered pages show text content
- ✅ Colors and borders display correctly
- ✅ Text is readable with proper color styling
- ✅ PNG output works correctly

### Known Limitations:
- ⚠️ Basic font rendering only (no font selection)
- ⚠️ Limited text layout (no text-align, line-height control)
- ⚠️ No background-image support (CSS property)

---

## Milestone 7: Image Rendering ✅ COMPLETE
**Goal**: Support `<img>` elements with common image formats

**Spec References**:
- HTML5 §2.5 URLs (URL resolution)
- HTML5 §4.8.2 The img element
- HTML5 §12.1.2 Void elements

### Tasks:
- [x] Implement URL resolution for relative paths
- [x] Load images from file system
- [x] Support PNG, JPEG, and GIF formats
- [x] Image caching to avoid redundant I/O
- [x] Scale images to CSS-defined dimensions
- [x] Alpha blending for transparent images

### Deliverables:
- ✅ `<img>` element rendering
- ✅ PNG, JPEG, GIF format support
- ✅ Image caching mechanism
- ✅ Relative URL resolution

### Validation:
- ✅ Images render at correct size
- ✅ Multiple image formats supported
- ✅ Transparent images blend correctly

### Known Limitations:
- ⚠️ Simple nearest-neighbor scaling (not bicubic)
- ⚠️ No network URL support (local files only)
- ⚠️ No srcset or responsive images
- ⚠️ No lazy loading

---

## Milestone 7.5: Basic Table Layout ✅ COMPLETE
**Goal**: Implement basic table layout support for `<table>`, `<tr>`, and `<td>` elements

**Spec References**:
- CSS 2.1 §17 Tables
- CSS 2.1 §17.5 Visual layout of table contents
- CSS 2.1 §17.5.2 Table width algorithms

### Tasks:
- [x] Add table box types (TableBox, TableRowBox, TableCellBox)
- [x] Implement display property detection for table elements
- [x] Implement table layout algorithm
- [x] Distribute column widths based on content (auto layout)
- [x] Position cells horizontally in rows
- [x] Calculate row heights based on cell content
- [x] Support padding and borders on table cells
- [x] Support colspan attribute
- [x] Add unit tests for table layout

### Deliverables:
- ✅ Table layout box types
- ✅ Auto table layout algorithm with content-based column sizing
- ✅ Colspan support for cells spanning multiple columns
- ✅ Test files demonstrating table rendering
- ✅ Unit tests for table layout

### Validation:
- ✅ Tables render with cells in correct positions
- ✅ Cells arranged horizontally in rows
- ✅ Multiple rows stack vertically
- ✅ Cell borders and padding work correctly
- ✅ Colspan attribute correctly spans columns
- ✅ Column widths sized based on content (narrow columns stay narrow)
- ✅ Hacker News table layout renders with proper proportions

### Known Limitations:
- ⚠️ No support for rowspan
- ⚠️ No table headers (`<thead>`, `<tbody>`, `<tfoot>`)
- ⚠️ No table captions
- ⚠️ No border-collapse support
- ⚠️ Simple content-width estimation (doesn't account for line wrapping)

---

## Milestone 8: Testing & Validation 🔄 IN PROGRESS
**Goal**: Comprehensive testing with public test suites

**Spec References**:
- CSS 2.1 Test Suite (W3C)
- WPT (Web Platform Tests)

### Tasks:
- [x] Integrate WPT reftest harness
- [x] Add CSS 2.1 reference tests
- [x] Document test results
- [x] Verify reftest status and document requirements for failing tests
- [ ] Implement CSS shorthand property expansion
- [ ] Expand test coverage
- [ ] Fix failing tests

### Current Test Results:
- **WPT CSS Tests**: 81.8% pass rate (9/11 tests)
- **Unit Test Coverage**: 90%+ across all modules
- **Test Categories Passing**:
  - ✅ css-box (longhand properties): 100% (3/3 tests)
  - ✅ css-cascade: 100% (2/2 tests)
  - ✅ css-display: 100% (1/1 test)
  - ✅ css-selectors: 100% (3/3 tests)
- **Test Categories Failing**:
  - ❌ css-box (shorthand properties): 0% (2/2 tests) - **shorthand property expansion not implemented**

### Failing Tests and Required Features:

#### 1. css-box/margin-shorthand-001.html
**Status**: FAIL - layouts do not match  
**Requirement**: Expand `margin: 20px` to individual longhand properties
- CSS 2.1 §8.3 Margin properties: margin shorthand
- Must expand to: `margin-top: 20px`, `margin-right: 20px`, `margin-bottom: 20px`, `margin-left: 20px`
- Implementation needed in `css/parser.go` or `style/style.go`

#### 2. css-box/padding-shorthand-001.html
**Status**: FAIL - layouts do not match  
**Requirement**: Expand `padding: 10px` to individual longhand properties
- CSS 2.1 §8.4 Padding properties: padding shorthand
- Must expand to: `padding-top: 10px`, `padding-right: 10px`, `padding-bottom: 10px`, `padding-left: 10px`
- Implementation needed in `css/parser.go` or `style/style.go`

### Implementation Strategy for Shorthand Properties:

To pass the failing tests, implement CSS shorthand property expansion:

1. **Add shorthand expansion function** in `css/parser.go` or `style/style.go`
   - Detect shorthand properties during parsing or style computation
   - Expand based on CSS 2.1 specification:
     - 1 value: all sides (e.g., `margin: 10px` → all 10px)
     - 2 values: vertical | horizontal (e.g., `margin: 10px 20px` → top/bottom 10px, left/right 20px)
     - 3 values: top | horizontal | bottom (e.g., `margin: 10px 20px 30px`)
     - 4 values: top | right | bottom | left (e.g., `margin: 10px 20px 30px 40px`)

2. **Properties to implement**:
   - `margin` → `margin-top`, `margin-right`, `margin-bottom`, `margin-left`
   - `padding` → `padding-top`, `padding-right`, `padding-bottom`, `padding-left`
   - (Future) `border` → individual border properties
   - (Future) `border-width`, `border-style`, `border-color`

3. **Where to implement**:
   - **Option A**: In `css/parser.go` - expand during parsing before creating Declaration objects
   - **Option B**: In `style/style.go` - expand during style computation when applying declarations
   - **Recommended**: Option B (style computation) for cleaner separation of concerns

### Deliverables:
- ✅ Test coverage report
- ✅ Documentation of spec compliance
- ✅ Known limitations documented
- ✅ CI integration with WPT tests
- ✅ Detailed requirements for failing tests documented

---

## Milestone 9: Network Support ✅ COMPLETE
**Goal**: Load and render web pages from HTTP/HTTPS URLs

**Spec References**:
- HTTP/HTTPS: Standard Go net/http implementation
- HTML5 §2.5 URLs: Relative URL resolution

### Tasks:
- [x] HTTP/HTTPS URL fetching
- [x] Detect URLs vs local file paths
- [x] Fetch HTML content from network
- [x] Fetch external stylesheets via `<link rel="stylesheet">`
- [x] Load images from network URLs
- [x] CSS parser robustness improvements
  - [x] Handle attribute selectors gracefully (CSS 2.1 §5.8)
  - [x] Handle @-rules gracefully (CSS 2.1 §4.1.5)

### Deliverables:
- ✅ Browser can load pages from URLs
- ✅ External CSS files are fetched and applied
- ✅ Network images are loaded and rendered
- ✅ CSS parser doesn't crash on modern CSS features

### Validation:
- ✅ Successfully loads https://news.ycombinator.com/
- ✅ Renders to PNG without crashing
- ✅ External CSS (news.css) is fetched and parsed
- ✅ Handles attribute selectors and @media queries gracefully

### Known Limitations:
- ⚠️ No HTTP caching (fetches on every request)
- ⚠️ No connection pooling or timeouts
- ⚠️ Attribute selectors are skipped (not applied)
- ⚠️ @-rules are skipped (media queries, imports, etc.)
- ⚠️ No support for relative URL resolution in CSS (e.g., background images)

---

## Future Work: Full Hacker News Rendering

The browser can now load Hacker News from the network and render content with proper table layout. Column widths are automatically sized based on content, with narrow columns (rank, vote links) staying narrow and the title column taking up the remaining space. Colspan is supported for subtext rows.

### Recent Improvements:
- [x] **Colspan Support** ✅ COMPLETE
  - Table cells can span multiple columns using colspan attribute
  - Column count calculated correctly across all rows
- [x] **Auto Table Layout** ✅ COMPLETE
  - Content-based column width calculation
  - Narrow columns (rank, votelinks) sized appropriately (~50px)
  - Wide columns (title) get remaining space
  - Maximum column width capping to prevent overflow

### Required Features for Full Fidelity:
- [ ] **Text Layout Improvements**
  - [ ] Inline text layout (wrap text within line boxes)
  - [ ] Font size support (not just default font)
  - [ ] Text-align property (left, center, right)
  - [ ] Line-height property
  - [ ] Proper inline box model

- [ ] **Link Rendering**
  - [ ] `<a>` element styling
  - [ ] Text decoration (underline)
  - [ ] Color for links

- [ ] **Table Support**
  - [x] `<table>`, `<tr>`, `<td>`, `<th>` elements
  - [x] Basic table layout algorithm (auto layout)
  - [x] Colspan attribute support
  - [ ] Auto table layout algorithm (full implementation with min/max widths)
  - [ ] Table spanning (rowspan)
  - [ ] Table captions and headers
  - [ ] Border-collapse property

- [ ] **Additional Selectors**
  - [ ] Child combinator (`>`)
  - [ ] Sibling combinators (`+`, `~`)
  - [ ] Pseudo-classes (`:hover`, `:visited`)

- [ ] **CSS Inheritance**
  - [ ] Inherit property values from parents
  - [ ] Computed value calculation

- [x] **Network Support** ✅ COMPLETE
  - [x] HTTP/HTTPS requests
  - [x] Load external stylesheets
  - [x] Load remote images

### Current Status:
The browser successfully loads and renders Hacker News from the network with improved table layout. Tables now use content-based column sizing, so narrow columns like rank numbers and vote arrows stay narrow, while title columns expand to fill available space. Colspan support allows subtext rows to properly span across multiple columns.

---

## Future Enhancements (Post-MVP)
- JavaScript support
- CSS 3 features (flexbox, grid, transitions, animations)
- Form handling
- Media queries (responsive design)
- Advanced typography (web fonts, font-weight, etc.)
- Accessibility features

---

## Current Status
**Completed**: Milestones 1-7.5 (Foundation through Basic Table Layout), Milestone 9 (Network Support)  
**In Progress**: Milestone 8 (Testing & Validation)  
**Last Updated**: 2025-12-24
