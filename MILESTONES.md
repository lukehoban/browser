# Browser Implementation Milestones

## Overview
This document tracks the milestones for implementing a simple web browser in Go, focusing on static HTML and CSS 2.1 compliance.

## Milestone 1: Foundation (Initial Setup) ✓
**Goal**: Set up project structure and basic architecture

### Tasks:
- [x] Initialize Go module
- [x] Create project directory structure
- [x] Document milestones
- [x] Add .gitignore

### Deliverables:
- Go module initialized
- Clear project structure
- Documentation framework

---

## Milestone 2: HTML Tokenization & Parsing
**Goal**: Parse static HTML into a DOM tree

**Spec References**: 
- HTML5 §12.2 Parsing HTML documents
- HTML5 §12.2.5 Tokenization

### Tasks:
- [ ] Implement HTML tokenizer
  - [ ] Data state
  - [ ] Tag open/close states
  - [ ] Character references
- [ ] Build DOM tree structure
  - [ ] Element nodes
  - [ ] Text nodes
  - [ ] Attribute nodes
- [ ] Parse common HTML elements (div, p, span, h1-h6, a, etc.)
- [ ] Add unit tests with sample HTML

### Deliverables:
- HTML tokenizer that produces tokens from HTML strings
- DOM tree builder that constructs a tree from tokens
- Test suite validating parsing of basic HTML documents

### Validation:
- Parse simple HTML documents successfully
- Handle nested elements correctly
- Preserve text content and attributes

---

## Milestone 3: CSS Parsing
**Goal**: Parse CSS 2.1 stylesheets

**Spec References**:
- CSS 2.1 §4 Syntax and basic data types
- CSS 2.1 §5 Selectors
- CSS 2.1 §6 Assigning property values

### Tasks:
- [ ] Implement CSS tokenizer
  - [ ] Identifiers, strings, numbers
  - [ ] Operators and delimiters
- [ ] Parse selectors
  - [ ] Type selectors (element)
  - [ ] Class selectors (.class)
  - [ ] ID selectors (#id)
  - [ ] Descendant combinators
- [ ] Parse declarations
  - [ ] Property names
  - [ ] Values (colors, lengths, keywords)
- [ ] Build stylesheet structure

### Deliverables:
- CSS tokenizer
- CSS parser producing stylesheet objects
- Support for basic selectors and properties

### Validation:
- Parse CSS rules correctly
- Handle multiple selectors
- Parse common properties (color, font-size, margin, padding, border)

---

## Milestone 4: Style Computation
**Goal**: Match CSS rules to DOM elements and compute styles

**Spec References**:
- CSS 2.1 §6.4 The cascade
- CSS 2.1 §6.1 Specified, computed, and actual values

### Tasks:
- [ ] Implement selector matching algorithm
- [ ] Calculate selector specificity (CSS 2.1 §6.4.3)
- [ ] Implement cascade (origin, importance, specificity, order)
- [ ] Compute inherited properties
- [ ] Resolve relative values to absolute

### Deliverables:
- Style computation engine
- Styled DOM tree with computed styles

### Validation:
- Correct selector matching
- Proper cascade order
- Inheritance working correctly

---

## Milestone 5: Layout Engine
**Goal**: Implement CSS 2.1 visual formatting model

**Spec References**:
- CSS 2.1 §8 Box model
- CSS 2.1 §9 Visual formatting model
- CSS 2.1 §10 Visual formatting model details

### Tasks:
- [ ] Implement box model (content, padding, border, margin)
- [ ] Block formatting context
- [ ] Inline formatting context
- [ ] Normal flow layout
- [ ] Width and height calculations

### Deliverables:
- Layout engine producing positioned boxes
- Support for block and inline elements

### Validation:
- Correct box dimensions
- Proper positioning of elements
- Margins, padding, borders applied correctly

---

## Milestone 6: Rendering ✓
**Goal**: Render the laid-out page

**Spec References**:
- CSS 2.1 §14 Colors and backgrounds
- CSS 2.1 §16 Text

### Tasks:
- [x] Implement display list generation
- [x] Render backgrounds and borders
- [x] Render text content
- [x] Output to image (PNG)

### Deliverables:
- Basic renderer with text support
- Visual output of simple pages
- Color support for text and backgrounds

### Validation:
- ✓ Rendered pages show text content
- ✓ Colors and borders display correctly
- ✓ Text is readable with proper color styling
- ✓ PNG output works correctly

---

## Milestone 7: Testing & Validation
**Goal**: Comprehensive testing with public test suites

**Spec References**:
- CSS 2.1 Test Suite (W3C)

### Tasks:
- [ ] Integrate CSS 2.1 test suite
- [ ] Document test results
- [ ] Fix failing tests
- [ ] Add regression tests

### Deliverables:
- Test coverage report
- Documentation of spec compliance
- Known limitations documented

### Validation:
- Pass subset of CSS 2.1 tests
- No regressions in basic functionality

---

## Future Enhancements (Post-MVP)
- JavaScript support
- CSS 3 features
- Network stack (HTTP)
- Form handling
- Advanced selectors
- Flexbox/Grid layout

---

## Current Status
**Active Milestone**: Milestone 1 (Foundation)
**Last Updated**: 2025-12-24
