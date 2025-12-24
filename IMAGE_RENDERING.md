# Image Rendering Implementation

## Overview
This implementation adds support for rendering `<img>` elements in the browser, as required for the HackerNews test case.

## Key Features

### 1. Image Loading
- **File Format Support**: PNG, JPEG, and GIF formats via Go's standard image decoders
- **URL Resolution**: Relative URLs are resolved in the DOM layer per HTML5 §2.5 URLs
- **Caching**: Images are cached after first load to avoid redundant file I/O
- **Error Handling**: Gracefully handles missing or invalid images

### 2. Image Rendering
- **Scaling**: Simple nearest-neighbor scaling to fit images into their CSS-defined dimensions
- **Alpha Blending**: Proper alpha blending for transparent images
- **Bounds Checking**: Safe pixel access with validation to prevent panics
- **Integration**: Seamlessly integrates with existing layout and rendering pipeline

### 3. Architecture
The implementation follows the existing browser architecture and HTML5 specifications:

```
HTML Parser → DOM Tree → URL Resolution (dom.ResolveURLs) → Style → Layout → Rendering
                            ↓
                    Resolves relative URLs
                    to absolute file paths
                            ↓
                      Render loads images
                      from absolute paths
```

URL resolution follows HTML5 §2.5 URLs, which states that relative URLs in documents
should be resolved against the document's base URL. In this implementation, the base
URL is the directory of the HTML file being rendered.

## Code Changes

### `dom/url.go` (NEW)
- **ResolveURLs()**: Resolves relative URLs in the DOM tree against a base directory
- Follows HTML5 §2.5 URLs specification for URL resolution
- Currently handles file system paths; designed for future URL support

### `render/render.go`
- **Canvas**: Added `ImageCache` field for caching loaded images
- **LoadImage()**: Loads images from absolute file paths with caching
- **DrawImage()**: Draws images with scaling and alpha blending
- **renderImage()**: Renders img elements in the layout tree

### `cmd/browser/main.go`
- Calls `dom.ResolveURLs()` after parsing to resolve image paths in the DOM
- URL resolution happens in the DOM layer, not the render layer

### Tests
- **dom/url_test.go**: Tests for URL resolution in the DOM layer
- **render/image_test.go**: Comprehensive unit tests for:
  - Image drawing with scaling
  - Alpha blending
  - Image loading and caching
  - End-to-end rendering

## Usage Example

```html
<!DOCTYPE html>
<html>
<head>
    <style>
        img {
            width: 50px;
            height: 50px;
        }
    </style>
</head>
<body>
    <img src="logo.png" alt="Logo">
</body>
</html>
```

```bash
./browser -output result.png test.html
```

## Test Cases

1. **test/hackernews.html**: Updated to include Y Combinator logo (y18.png)
2. **test/image_test.html**: Simple demonstration with multiple images
3. **test/y18.png**: 18×18 orange square test image

## Limitations

- No support for background-image CSS property (only `<img>` elements)
- No support for image srcset or responsive images
- Scaling uses simple nearest-neighbor algorithm (not bicubic/bilinear)
- No support for external URLs (only local file paths)

## Browser Spec Compliance

- **HTML5 §2.5**: URLs - URL resolution is performed in the DOM layer against the document's base URL
- **HTML5 §4.8.2**: The img element (basic support)
- **HTML5 §12.1.2**: Void elements (img is correctly handled as void element)

## Performance Considerations

- Images are cached after first load
- Scaling is done once during rendering
- No dynamic re-rendering on image load (images must exist before rendering)

## Future Enhancements

- Support for background-image CSS property
- Better scaling algorithms (bilinear, bicubic)
- Network image loading (HTTP/HTTPS)
- Lazy loading and progressive rendering
- Image srcset and responsive images
