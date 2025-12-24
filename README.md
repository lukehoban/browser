# Browser - A Simple Web Browser in Go

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

🚧 **In Development** - Currently implementing Milestone 1: Foundation

## License

MIT