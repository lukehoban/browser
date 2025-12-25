# Testing and Validation

This document describes how to test and validate the browser implementation against public test suites.

## Public Test Suites

### CSS 2.1 Test Suite

The official CSS 2.1 test suite from W3C can be used to validate CSS parsing and rendering:

**Location**: https://test.csswg.org/suites/css2.1/

**Description**: The CSS 2.1 Test Suite contains thousands of tests covering all aspects of the CSS 2.1 specification. Tests are organized by specification section.

**Usage**:
```bash
# Download the test suite
git clone https://github.com/web-platform-tests/wpt.git
cd wpt/css/CSS2

# Run individual tests
go run ./cmd/browser path/to/test.html
```

### HTML5 Test Suite

For HTML parsing, the html5lib-tests repository provides comprehensive tests:

**Location**: https://github.com/html5lib/html5lib-tests

**Description**: Contains tests for HTML tokenization and tree construction based on the HTML5 specification.

**Usage**:
```bash
# Clone the test repository
git clone https://github.com/html5lib/html5lib-tests.git

# Tests are in JSON format and need to be adapted
# Each test contains input HTML and expected DOM structure
```

## Current Test Coverage

### Unit Tests

All core modules have comprehensive unit tests:

- `dom/`: DOM tree structure and manipulation
- `html/`: HTML tokenization and parsing
- `css/`: CSS tokenization and parsing
- `style/`: Style computation, selector matching, specificity

Run all unit tests:
```bash
go test ./...
```

Run with coverage:
```bash
go test -cover ./...
```

### Integration Tests

Integration tests use complete HTML documents with embedded CSS:

- `test/simple.html`: Basic HTML structure
- `test/styled.html`: HTML with CSS styles
- `test/hackernews.html`: Simplified Hacker News homepage for testing real-world layout

## Test Results

### HTML Parsing

✅ **Passing**:
- Basic HTML tokenization
- Tag parsing (start, end, self-closing)
- Attribute parsing (quoted and unquoted)
- Nested elements
- Text nodes
- Comments
- DOCTYPE declarations
- Mixed content (text and elements)

⚠️ **Known Limitations**:
- No support for character references (e.g., `&amp;`, `&lt;`)
- Simplified error recovery
- No support for script/style CDATA sections
- No namespace support

### CSS Parsing

✅ **Passing**:
- CSS tokenization (identifiers, strings, numbers, punctuation)
- Simple selectors (element, class, ID)
- Combined selectors (e.g., `div#id.class`)
- Descendant combinators (e.g., `div p`)
- Multiple selectors (comma-separated)
- Declaration parsing
- Comments

⚠️ **Known Limitations**:
- No support for pseudo-classes (`:hover`, `:first-child`)
- No support for pseudo-elements (`::before`, `::after`)
- No attribute selectors (`[attr="value"]`)
- No child/adjacent sibling combinators (`>`, `+`, `~`)
- Limited value parsing (treated as strings)

### Style Computation

✅ **Passing**:
- Selector matching
- Specificity calculation (CSS 2.1 §6.4.3)
- Cascade (specificity-based)
- Descendant selector matching

⚠️ **Known Limitations**:
- No inheritance implementation
- No important (`!important`) support
- No shorthand property expansion
- No computed value calculation

## Future Testing Goals

1. ~~**Integrate CSS 2.1 Test Suite**~~ ✅ **Done**
   - Create test runner for official W3C tests
   - Track pass/fail rates by category
   - Document compliance level

2. **HTML5lib Tests**
   - Adapt JSON test format
   - Validate tokenization against reference
   - Validate tree construction

3. **Visual Regression Testing**
   - Generate reference renderings
   - Compare output images
   - Track visual differences

4. **Performance Benchmarks**
   - Parse large documents
   - Style computation performance
   - Memory usage profiling

## WPT Reftest Harness

The browser includes a WPT (Web Platform Tests) reference test harness for benchmarking CSS compliance.

### Running the WPT Reftest Suite

```bash
# Build and run the test runner
go build -o wptrunner ./cmd/wptrunner
./wptrunner -v test/wpt/css/

# Or run as Go tests
go test ./reftest/... -v
```

### WPT Tests in CI

WPT tests run automatically in CI as a separate job that:
- **Does not block merges** - The WPT job uses `continue-on-error: true` to allow PRs to merge even when tests fail
- **Generates reports** - Test results are uploaded as CI artifacts (wpt-report.txt and wpt-summary.md)
- **Provides visibility** - The test summary is displayed in the CI logs
- **Tracks progress** - Results are retained for 30 days to track improvements over time

To view WPT test results from CI:
1. Go to the Actions tab in GitHub
2. Click on a workflow run
3. Find the "WPT Tests" job
4. View the summary in the logs or download the "wpt-test-report" artifact

### Current WPT CSS Reftest Results

| Category | Tests | Passed | Failed | Pass Rate |
|----------|-------|--------|--------|-----------|
| css-borders | 1 | 1 | 0 | 100% |
| css-box | 9 | 9 | 0 | 100% |
| css-cascade | 3 | 3 | 0 | 100% |
| css-color | 2 | 2 | 0 | 100% |
| css-display | 2 | 2 | 0 | 100% |
| css-fonts | 4 | 4 | 0 | 100% |
| css-inheritance | 3 | 3 | 0 | 100% |
| css-selectors | 5 | 5 | 0 | 100% |
| css-text-decor | 1 | 1 | 0 | 100% |
| **Total** | **30** | **30** | **0** | **100%** 🎉 |

### Test Categories

1. **css-borders**: Border property tests
   - border-color: ✅ Passing
   - border-width with border-style: ✅ Passing

2. **css-box**: Box model tests (width, height, padding, margin)
   - Longhand properties: ✅ Passing
   - Shorthand properties: ✅ Passing
   - Combined box model: ✅ Passing
   - Percentage widths: ✅ Passing
   - Different padding values: ✅ Passing

3. **css-cascade**: Cascade and specificity tests
   - Specificity calculation: ✅ Passing
   - ID vs class priority: ✅ Passing
   - Multiple classes specificity: ✅ Passing

4. **css-color**: Color property tests
   - Hex colors (#RRGGBB): ✅ Passing
   - Named colors: ✅ Passing
   - Text color: ✅ Passing

5. **css-display**: Display property tests
   - Block display: ✅ Passing
   - Table display: ✅ Passing

6. **css-fonts**: Font property tests
   - font-size: ✅ Passing
   - font-weight (bold): ✅ Passing
   - font-style (italic): ✅ Passing
   - Combined font properties: ✅ Passing

7. **css-inheritance**: CSS inheritance tests
   - color inheritance: ✅ Passing
   - font-size inheritance: ✅ Passing
   - font-weight inheritance: ✅ Passing

8. **css-selectors**: Selector tests
   - Class selector: ✅ Passing
   - ID selector: ✅ Passing
   - Descendant combinator: ✅ Passing
   - Element.class combined: ✅ Passing
   - Element#id.class combined: ✅ Passing
   - Multiple selectors (comma-separated): ✅ Passing

9. **css-text-decor**: Text decoration tests
   - text-decoration underline: ✅ Passing

### Adding New Tests

To add new WPT-style reference tests:

1. Create a test HTML file with `<link rel="match" href="reference.html">`
2. Create a reference HTML file that produces the expected layout
3. Place both files in `test/wpt/css/<category>/`
4. Run `./wptrunner test/wpt/css/` to verify

### Recently Implemented Features

#### CSS Shorthand Property Expansion ✅
**Status**: Implemented in `style/style.go`

Shorthand properties are now automatically expanded to their longhand equivalents:
- **Margin**: `margin: 20px` → `margin-top`, `margin-right`, `margin-bottom`, `margin-left`
- **Padding**: `padding: 10px` → `padding-top`, `padding-right`, `padding-bottom`, `padding-left`

**Supported patterns** (CSS 2.1 §8.3, §8.4):
- 1 value: all sides (e.g., `margin: 10px`)
- 2 values: vertical | horizontal (e.g., `margin: 10px 20px`)
- 3 values: top | horizontal | bottom (e.g., `margin: 10px 20px 30px`)
- 4 values: top | right | bottom | left (e.g., `margin: 10px 20px 30px 40px`)

### Gaps Identified

Based on the reftest benchmark, the following features could improve test coverage in the future:

1. **CSS inheritance** - **Not currently tested**
   - Inheritable properties should cascade from parent to child
   - Affects: color, font properties, line-height, text-align, etc.
   - CSS 2.1 §6.2 Inheritance

2. **!important support** - **Not currently tested**
   - Override cascade based on !important declarations
   - CSS 2.1 §6.4.2 !important rules

3. **Computed value calculation** - **Not currently tested**
   - Convert relative values to absolute (e.g., em to px)
   - CSS 2.1 §6.1.2 Computed values

4. **Child/sibling combinators** (>, +, ~) - **Not currently tested**
   - Child combinator: `parent > child`
   - Adjacent sibling: `element + sibling`
   - General sibling: `element ~ sibling`
   - CSS 2.1 §5.5 Child selectors, §5.7 Adjacent sibling selectors

5. **Pseudo-classes and pseudo-elements** - **Not currently tested**
   - :hover, :focus, :first-child, :last-child, etc.
   - ::before, ::after
   - CSS 2.1 §5.11 Pseudo-classes, §5.12 Pseudo-elements

## Running Tests

### All Tests
```bash
go test ./...
```

### Specific Module
```bash
go test ./html
go test ./css
go test ./style
```

### With Verbose Output
```bash
go test -v ./...
```

### With Coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Contributing Tests

When adding new features:

1. Add unit tests in the appropriate `*_test.go` file
2. Add integration tests in `test/` directory
3. Update this document with test coverage
4. Document any known limitations

## References

- CSS 2.1 Test Suite: https://test.csswg.org/suites/css2.1/
- HTML5lib Tests: https://github.com/html5lib/html5lib-tests
- CSS 2.1 Specification: https://www.w3.org/TR/CSS21/
- HTML5 Specification: https://html.spec.whatwg.org/
