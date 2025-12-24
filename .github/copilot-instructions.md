# Copilot Instructions for Browser Project

## Overview

This is a simple web browser implementation in Go that parses HTML and CSS, computes styles, calculates layout, and renders to PNG images.

## Milestones Document

**When implementing new features or making significant changes, always update MILESTONES.md:**

1. Mark tasks as complete with `[x]` when they are implemented
2. Update validation status (✅/⚠️/❌) to reflect current state
3. Add new milestones if implementing features not yet tracked
4. Update "Known Limitations" when fixing or discovering issues
5. Keep "Last Updated" date current
6. Update "Current Status" section to reflect active work

The MILESTONES.md document is the source of truth for project progress and should always reflect the actual implementation state.

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
- `test/hackernews.html` - HN-inspired layout test with image

## Code Guidelines

- Follow existing code style and patterns
- **ALWAYS add spec references for CSS/HTML/DOM implementations** (see existing code for examples)
  - For CSS features: Reference CSS 2.1 specification sections (e.g., "CSS 2.1 §5.8 Attribute selectors")
  - For HTML features: Reference HTML5 specification sections (e.g., "HTML5 §2.5 URLs")
  - For DOM features: Reference DOM Level 2 Core or relevant specs
  - For non-standard or practical implementations: Document the reasoning
- Write unit tests for new functionality
- Run `go test ./...` before submitting changes
- Update MILESTONES.md when completing tasks or discovering limitations
