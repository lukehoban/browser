# Browser - A Simple Web Browser in Go

[![CI](https://github.com/lukehoban/browser/actions/workflows/ci.yml/badge.svg)](https://github.com/lukehoban/browser/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lukehoban/browser)](https://goreportcard.com/report/github.com/lukehoban/browser)

A simple web browser implementation in Go, focusing on static HTML and CSS 2.1 compliance. This project aims to stay close to W3C specifications and provide a clean, well-organized codebase for educational purposes.

## Features

- HTML parsing with DOM tree construction
- CSS 2.1 parsing and style computation
- Visual formatting model (box model, block/inline layout)
- Basic rendering capabilities

## Project Structure

```
browser/
├── cmd/browser/      # Main browser application
├── html/            # HTML tokenization and parsing
├── css/             # CSS parsing
├── dom/             # DOM tree structure
├── style/           # Style computation and cascade
├── layout/          # Layout engine (visual formatting model)
├── render/          # Rendering engine
└── test/            # Test files and fixtures
```

## Specifications

This browser implementation follows these W3C specifications:

- **HTML5**: Tokenization and parsing ([HTML5 §12](https://html.spec.whatwg.org/multipage/parsing.html))
- **CSS 2.1**: Syntax, selectors, cascade, box model, and visual formatting
  - [CSS 2.1 §4 Syntax](https://www.w3.org/TR/CSS21/syndata.html)
  - [CSS 2.1 §5 Selectors](https://www.w3.org/TR/CSS21/selector.html)
  - [CSS 2.1 §6 Cascade](https://www.w3.org/TR/CSS21/cascade.html)
  - [CSS 2.1 §8 Box Model](https://www.w3.org/TR/CSS21/box.html)
  - [CSS 2.1 §9 Visual Formatting Model](https://www.w3.org/TR/CSS21/visuren.html)

## Building

```bash
go build ./cmd/browser
```

## Testing

```bash
go test ./...
```

## Milestones

See [MILESTONES.md](MILESTONES.md) for detailed implementation milestones and progress tracking.

## Current Status

✅ **Milestone 1: Foundation** - Complete
✅ **Milestone 2: HTML Parsing** - Complete  
✅ **Milestone 3: CSS Parsing** - Complete
✅ **Milestone 4: Style Computation** - Complete
✅ **Milestone 5: Layout Engine** - Complete (Basic box model)

🚧 **In Development** - Milestone 6: Rendering

### What Works

- **HTML Parsing**: Tokenization, tree construction, nested elements, attributes
- **CSS Parsing**: Selectors (element, class, ID, descendant), declarations, multiple rules
- **Style Computation**: Selector matching, specificity calculation, cascade
- **Layout**: Box model (content, padding, border, margin), block layout, auto width calculation

### Example Usage

```bash
# Build the browser
go build ./cmd/browser

# Parse and display HTML with styles
./browser test/styled.html
```

## TODO: Next Major Features

The following are the next major features planned for implementation, listed in priority order:

1. **Text Rendering** - Render actual text content from HTML elements, including font support and basic typography
2. **Inline Layout** - Implement inline formatting context (CSS 2.1 §9.4.2) to properly handle inline elements like `<span>`, `<a>`, and text flow
3. **CSS Inheritance** - Implement property inheritance (CSS 2.1 §6.2) so child elements inherit applicable properties from parents
4. **Advanced Selectors** - Add support for:
   - Child combinator (`>`)
   - Adjacent sibling combinator (`+`)
   - Attribute selectors (`[attr]`, `[attr=value]`)
   - Pseudo-classes (`:hover`, `:first-child`, `:last-child`)
5. **Positioning Schemes** - Implement additional positioning (CSS 2.1 §9.3):
   - Relative positioning (`position: relative`)
   - Absolute positioning (`position: absolute`)
   - Fixed positioning (`position: fixed`)
6. **Float and Clear** - Implement floating elements (CSS 2.1 §9.5) and the clear property
7. **Shorthand Properties** - Expand shorthand CSS properties like `margin`, `padding`, `border`, `font`, `background`
8. **Additional Color Formats** - Support `rgb()`, `rgba()`, `hsl()`, `hsla()` color functions
9. **CSS Box Model Extensions** - Add support for:
   - `min-width`, `max-width`, `min-height`, `max-height`
   - `box-sizing` property
10. **Network Stack** - Add HTTP/HTTPS support to fetch remote HTML and CSS resources
11. **Image Support** - Render `<img>` elements with support for common image formats (PNG, JPEG, GIF)
12. **Table Layout** - Implement table formatting context (CSS 2.1 §17) for proper `<table>` rendering
13. **List Styling** - Support list-style properties for `<ul>`, `<ol>`, and `<li>` elements
14. **Interactive Elements** - Add support for forms and form controls (`<input>`, `<button>`, `<select>`, etc.)
15. **CSS 2.1 Test Suite Integration** - Automate running and reporting on W3C CSS 2.1 test suite compliance

## License

MIT