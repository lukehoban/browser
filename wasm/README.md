# Browser WebAssembly Demo

This directory contains a WebAssembly build of the browser that runs entirely in your web browser.

## Live Demo

The demo is automatically deployed to GitHub Pages and available at:
**https://lukehoban.github.io/browser/**

The deployment happens automatically via GitHub Actions whenever changes are pushed to the `main` branch.

## Files

- `index.html` - The demo web page
- `wasm_exec.js` - Go's WebAssembly JavaScript support file
- `browser.wasm` - The compiled WebAssembly binary (generated during build)

## Building and Running Locally

From the repository root:

```bash
GOOS=js GOARCH=wasm go build -o wasm/browser.wasm ./cmd/browser-wasm
cd wasm && python3 -m http.server 8080
```

Then open http://localhost:8080 in your web browser.

## Usage

1. Enter HTML (with inline CSS in `<style>` tags) in the left panel
2. Adjust viewport width and height if desired
3. Click "Render" to see the result
4. The rendered image will appear in the right panel

## Example Pages

The demo includes several example pages:
- **Simple** - Basic HTML with minimal styling
- **Colors** - Colored boxes demonstrating CSS colors
- **Layout** - Header/content/footer layout
- **Text Styles** - Various text styling options (bold, italic, underline, different sizes)

## Browser Support

Requires a modern browser with WebAssembly support:
- Chrome 57+
- Firefox 52+
- Safari 11+
- Edge 79+

## Limitations

- The browser renders HTML/CSS to a PNG image
- Only inline CSS in `<style>` tags is supported (no external stylesheets)
- **Network features (loading external resources) are not available in WASM mode** due to browser security restrictions (CORS)
- Image support is limited due to cross-origin restrictions

### Why Can't WASM Load Live Pages?

WebAssembly runs in the browser's sandbox and cannot make arbitrary HTTP requests due to CORS (Cross-Origin Resource Sharing) policies. The Go `http.Get()` function used by the CLI browser doesn't work in WASM because:

1. **CORS restrictions**: Websites must explicitly allow cross-origin requests
2. **Browser security**: WASM cannot bypass browser security policies
3. **Network APIs**: Go's standard library HTTP client isn't available in WASM

### Alternatives for Loading Live Content

If you want to render live web pages:

1. **Use the CLI browser**: The native CLI binary (`./browser https://news.ycombinator.com/`) has full network support
2. **Proxy server**: Create a server that fetches content and serves it to the WASM app (not implemented in this demo)
3. **Browser extension**: Build a browser extension that can bypass CORS restrictions (not implemented)

## GitHub Pages Setup

To enable GitHub Pages for this repository:

1. Go to repository **Settings** → **Pages**
2. Under "Build and deployment", set Source to **GitHub Actions**
3. The `.github/workflows/pages.yml` workflow will automatically build and deploy the WASM demo
4. After the workflow runs, the demo will be available at `https://[username].github.io/[repository]/`
