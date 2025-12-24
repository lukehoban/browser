# Browser - A Simple Web Browser in Go

[![CI](https://github.com/lukehoban/browser/actions/workflows/ci.yml/badge.svg)](https://github.com/lukehoban/browser/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lukehoban/browser)](https://goreportcard.com/report/github.com/lukehoban/browser)

A simple web browser implementation in Go, focusing on static HTML and CSS 2.1 compliance. This project aims to stay close to W3C specifications and provide a clean, well-organized codebase for educational purposes.

## Features

- HTML parsing with DOM tree construction
- CSS 2.1 parsing and style computation
- Visual formatting model (box model, block/inline layout)
- Text rendering with color support
- Background and border rendering
- PNG image output

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
✅ **Milestone 6: Rendering** - Complete (Text, backgrounds, borders)

### What Works

- **HTML Parsing**: Tokenization, tree construction, nested elements, attributes
- **CSS Parsing**: Selectors (element, class, ID, descendant), declarations, multiple rules
- **Style Computation**: Selector matching, specificity calculation, cascade
- **Layout**: Box model (content, padding, border, margin), block layout, auto width calculation
- **Rendering**: Text rendering with colors, backgrounds, borders, PNG output

### Example Usage

```bash
# Build the browser
go build ./cmd/browser

# Render HTML to PNG
./browser -output output.png test/styled.html

# View layout tree without rendering
./browser test/styled.html
```

## License

MIT