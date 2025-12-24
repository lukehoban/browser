# Copilot Instructions for Browser Project

## Overview

This is a simple web browser implementation in Go that parses HTML and CSS, computes styles, calculates layout, and renders to PNG images.

## Screenshot Requirements

**When making changes that affect visual output, always attach a screenshot of the rendered result.**

This project renders HTML/CSS to PNG images. Visual changes are difficult to review from code alone. To ensure quality:

1. **Before submitting changes** that affect rendering, layout, or styling:
   - Build the browser: `go build ./cmd/browser`
   - Render a test page: `./browser -output screenshot.png test/render_test.html`
   - Attach the screenshot to your PR or commit

2. **Create comparison screenshots** when modifying existing behavior:
   - Capture "before" screenshot using the main branch
   - Capture "after" screenshot with your changes
   - Include both in your PR description

3. **Use descriptive test HTML files** to demonstrate changes:
   - Create minimal HTML that showcases the specific feature
   - Save test files in the `test/` directory

## Rendering Commands

```bash
# Build the browser
go build ./cmd/browser

# Render HTML to PNG
./browser -output output.png test/styled.html

# Render with custom viewport size
./browser -output output.png -width 1024 -height 768 test/styled.html

# View layout tree without rendering
./browser test/styled.html
```

## Test Files

- `test/simple.html` - Basic HTML structure
- `test/styled.html` - HTML with CSS styling
- `test/render_test.html` - Visual rendering test with colored boxes

## Code Guidelines

- Follow existing code style and patterns
- Add spec references for CSS/HTML implementations (see existing code for examples)
- Write unit tests for new functionality
- Run `go test ./...` before submitting changes
