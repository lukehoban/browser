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

1. **Integrate CSS 2.1 Test Suite**
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
