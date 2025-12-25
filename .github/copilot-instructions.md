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

1. **Keep the README Hacker News screenshot current.** When you make meaningful rendering/layout changes, regenerate `hackernews_screenshot.png` (e.g., `./browser -output hackernews_screenshot.png https://news.ycombinator.com/`) and update the README reference.

2. **Before submitting changes** that affect rendering, layout, or styling:
   - Build the browser: `go build ./cmd/browser`
   - Render a test page: `./browser -output screenshot.png test/render_test.html`
   - Attach the screenshot to your PR or commit

3. **Create comparison screenshots** when modifying existing behavior:
   - Capture "before" screenshot using the main branch
   - Capture "after" screenshot with your changes
   - Include both in your PR description
