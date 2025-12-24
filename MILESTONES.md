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

## Milestone 8: Testing & Validation 🔄 IN PROGRESS
**Goal**: Comprehensive testing with public test suites

**Spec References**:
- CSS 2.1 Test Suite (W3C)
- WPT (Web Platform Tests)

### Tasks:
- [x] Integrate WPT reftest harness
- [x] Add CSS 2.1 reference tests
- [x] Document test results
- [ ] Expand test coverage
- [ ] Fix failing tests

### Current Test Results:
- **WPT CSS Tests**: 81.8% pass rate (9/11 tests)
- **Unit Test Coverage**: 90%+ across all modules
- **Test Categories Passing**:
  - ✅ css-box (longhand properties): 100%
  - ✅ css-cascade: 100%
  - ✅ css-display: 100%
  - ✅ css-selectors: 100%
- **Test Categories Failing**:
  - ❌ css-box (shorthand properties): 0% (not implemented)

### Deliverables:
- ✅ Test coverage report
- ✅ Documentation of spec compliance
- ✅ Known limitations documented
- ✅ CI integration with WPT tests

---

## Future Work: Full Hacker News Rendering

To render the actual Hacker News homepage correctly (not just a simplified visual approximation), the following features are needed:

### Required Features:
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
  - [ ] `<table>`, `<tr>`, `<td>` elements
  - [ ] Table layout algorithm
  - [ ] HN uses tables for layout

- [ ] **Additional Selectors**
  - [ ] Child combinator (`>`)
  - [ ] Sibling combinators (`+`, `~`)
  - [ ] Pseudo-classes (`:hover`, `:visited`)

- [ ] **CSS Inheritance**
  - [ ] Inherit property values from parents
  - [ ] Computed value calculation

- [ ] **Network Support**
  - [ ] HTTP/HTTPS requests
  - [ ] Load external stylesheets
  - [ ] Load remote images

### Current Workaround:
The `test/hackernews.html` file is a simplified visual approximation that uses colored boxes to demonstrate layout capabilities without requiring full text layout or table support.

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
**Completed**: Milestones 1-7 (Foundation through Image Rendering)  
**In Progress**: Milestone 8 (Testing & Validation)  
**Last Updated**: 2025-12-24
