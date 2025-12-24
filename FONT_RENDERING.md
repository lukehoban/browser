# Font Rendering Implementation

This document describes the font rendering capabilities of the browser, including supported CSS properties and implementation details.

## Supported CSS 2.1 Font Properties

### font-family (CSS 2.1 §15.3)

The browser supports the following font families:

- **sans-serif** (default): Proportional-width sans-serif font using Go Regular
- **monospace**: Fixed-width monospace font using Go Mono
- **serif**: Falls back to sans-serif (Go Regular)

Generic font names are also mapped:
- `Arial`, `Helvetica` → sans-serif
- `Courier`, `Courier New` → monospace
- `Times`, `Times New Roman` → serif (falls back to sans-serif)

**Example:**
```css
body {
    font-family: sans-serif;
}

code {
    font-family: monospace;
}
```

### font-size (CSS 2.1 §15.7)

The browser supports multiple font-size units and keywords:

#### Absolute Size Keywords
- `xx-small` → 9px
- `x-small` → 10px
- `small` → 13px
- `medium` → 16px (default)
- `large` → 18px
- `x-large` → 24px
- `xx-large` → 32px

#### Relative Size Keywords
- `smaller` → 0.83 × parent size
- `larger` → 1.2 × parent size

#### Length Units
- **px** (pixels): Direct pixel values (e.g., `14px`)
- **pt** (points): Converted to pixels at 96 DPI (e.g., `12pt` ≈ 16px)
- **em** (em units): Relative to parent font size (e.g., `1.5em` = 1.5 × parent)

**Examples:**
```css
h1 {
    font-size: 32px;
}

p {
    font-size: medium;  /* 16px */
}

small {
    font-size: 0.875em; /* 87.5% of parent */
}
```

### font-weight (CSS 2.1 §15.6)

The browser supports two font weights:

- **normal** (default): Regular weight using Go Regular
- **bold**: Bold weight using Go Bold

Numeric weights are mapped:
- `400` or less → normal
- `700` or more → bold

**Example:**
```css
strong {
    font-weight: bold;
}

p {
    font-weight: normal;
}
```

### color (CSS 2.1 §14.1)

Text color is supported through the `color` property:

- Named colors: `black`, `white`, `red`, `blue`, etc.
- Hex colors: `#RGB`, `#RRGGBB`

**Example:**
```css
h1 {
    color: #2c3e50;
}

.warning {
    color: red;
}
```

## Implementation Details

### Font Loading

The browser uses the `golang.org/x/image/font` package ecosystem:

- **goregular**: Go Regular TrueType font (sans-serif)
- **gobold**: Go Bold TrueType font (bold weight)
- **gomono**: Go Mono TrueType font (monospace)
- **opentype**: TrueType font parser and renderer

Fonts are parsed once and cached in a `FontManager` to improve performance.

### Text Measurement

Text dimensions are calculated using actual font metrics:

- **Width**: Measured by summing glyph advances for each character
- **Height**: Taken from font metrics (ascent + descent)

This ensures accurate layout and positioning of text elements.

### Rendering Pipeline

1. **Style Computation**: CSS font properties are extracted during style computation
2. **Layout**: Text dimensions are calculated using actual font metrics
3. **Rendering**: Text is drawn using the specified font face with proper positioning

### Spec Compliance

The implementation follows these CSS 2.1 specifications:

- **§15 Fonts**: Font properties and matching
- **§15.3 Font family**: The font-family property
- **§15.6 Font weight**: The font-weight property  
- **§15.7 Font size**: The font-size property
- **§16 Text**: Text rendering

## Test Files

The following test files demonstrate font rendering:

- `test/font_test.html`: Basic font sizes and weights
- `test/font_comprehensive.html`: Comprehensive demonstration of all features
- `test/styled.html`: Styled page with various font properties

## Known Limitations

- **Limited font families**: Only sans-serif and monospace are fully implemented
- **No font-style**: Italic and oblique styles are not supported
- **No web fonts**: Only built-in TrueType fonts are supported
- **No font fallback**: Font family fallback chains are not implemented
- **Fixed DPI**: Assumes 72 DPI for font rendering

## Future Enhancements

Potential improvements for font rendering:

1. **Font-style support**: Add italic fonts (goitalic package)
2. **Additional font families**: Add serif fonts
3. **Web fonts**: Support @font-face and external font loading
4. **Font fallback**: Implement font family fallback chains
5. **Advanced text layout**: Line-height, text-align, text-decoration
6. **Text shaping**: Support for complex scripts and ligatures

## Usage Examples

### Basic Text with Different Sizes

```html
<h1 style="font-size: 32px;">Large Heading</h1>
<p style="font-size: 14px;">Regular paragraph text.</p>
<small style="font-size: 10px;">Small fine print.</small>
```

### Bold Text

```html
<p style="font-weight: bold;">This text is bold.</p>
<strong>This is also bold (via HTML semantics).</strong>
```

### Monospace Code

```html
<code style="font-family: monospace;">function hello() { return "Hi"; }</code>
```

### Colored Text

```html
<p style="color: blue; font-size: 18px;">Blue text at 18px.</p>
```

## References

- [CSS 2.1 Fonts Specification](https://www.w3.org/TR/CSS21/fonts.html)
- [Go Fonts Blog Post](https://blog.golang.org/go-fonts)
- [golang.org/x/image/font Package](https://pkg.go.dev/golang.org/x/image/font)
