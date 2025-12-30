// Package style handles style computation and the CSS cascade.
// It matches CSS selectors to DOM elements and computes final styles.
//
// Spec references:
// - CSS 2.1 §6 Assigning property values, Cascading, and Inheritance: https://www.w3.org/TR/CSS21/cascade.html
// - CSS 2.1 §6.4.3 Calculating a selector's specificity: https://www.w3.org/TR/CSS21/cascade.html#specificity
// - CSS 2.1 §6.4.4 Precedence of non-CSS presentational hints: https://www.w3.org/TR/CSS21/cascade.html#preshint
//
// Implemented features:
// - Selector matching: element, class, ID, descendant combinators
// - Specificity calculation per CSS 2.1 §6.4.3
// - Cascade by specificity and source order
// - Inline style attribute support (highest specificity)
// - User-agent stylesheet (lowest specificity)
// - Property inheritance for font properties (CSS 2.1 §6.2)
// - Shorthand property expansion (margin, padding, border)
//
// Not yet implemented (noted with log warnings where encountered):
// - !important declarations (CSS 2.1 §6.4.2)
// - Child combinator > (CSS 2.1 §5.6)
// - Sibling combinators +, ~ (CSS 2.1 §5.7, CSS3 Selectors)
// - Attribute selectors [attr=value] (CSS 2.1 §5.8)
// - Pseudo-classes :hover, :focus, etc. (CSS 2.1 §5.11)
// - Pseudo-elements ::before, ::after (CSS 2.1 §5.12)
// - Full computed value calculation (CSS 2.1 §6.1.2)
// - Inheritance of all inheritable properties (currently subset)
package style

import (
	"strconv"
	"strings"

	"github.com/lukehoban/browser/css"
	"github.com/lukehoban/browser/dom"
)

// StyledNode represents a DOM node with computed styles.
type StyledNode struct {
	Node     *dom.Node
	Styles   map[string]string
	Children []*StyledNode
}

// MatchedRule represents a CSS rule that matched a node, with its specificity.
type MatchedRule struct {
	Rule        *css.Rule
	Specificity Specificity
}

// Specificity represents the specificity of a CSS selector.
// CSS 2.1 §6.4.3 Calculating a selector's specificity
type Specificity struct {
	A int // Inline styles (A=1 for inline styles)
	B int // ID selectors
	C int // Class selectors, attribute selectors
	D int // Element selectors
}

// Compare returns true if s1 is more specific than s2.
// CSS 2.1 §6.4.3
func (s Specificity) Compare(other Specificity) int {
	if s.A != other.A {
		return s.A - other.A
	}
	if s.B != other.B {
		return s.B - other.B
	}
	if s.C != other.C {
		return s.C - other.C
	}
	return s.D - other.D
}

// StyleTree computes styles for a DOM tree using a stylesheet.
// CSS 2.1 §6 Assigning property values
// CSS 2.1 §6.4.4: User agent -> Author stylesheet cascade
func StyleTree(root *dom.Node, authorStylesheet *css.Stylesheet) *StyledNode {
	// CSS 2.1 §6.4.1: Cascading order - User agent styles come first
	// Merge user-agent stylesheet with author stylesheet
	mergedStylesheet := &css.Stylesheet{
		Rules: make([]*css.Rule, 0),
	}
	
	// Add user-agent styles first (lower specificity in cascade)
	userAgentStylesheet := DefaultUserAgentStylesheet()
	mergedStylesheet.Rules = append(mergedStylesheet.Rules, userAgentStylesheet.Rules...)
	
	// Add author styles second (higher specificity in cascade)
	if authorStylesheet != nil {
		mergedStylesheet.Rules = append(mergedStylesheet.Rules, authorStylesheet.Rules...)
	}
	
	return styleNode(root, mergedStylesheet, make(map[string]string))
}

// styleNode computes styles for a single node and its children.
// CSS 2.1 §6.2: Font properties are inherited from parent to child
func styleNode(node *dom.Node, stylesheet *css.Stylesheet, parentStyles map[string]string) *StyledNode {
	styled := &StyledNode{
		Node:     node,
		Styles:   make(map[string]string),
		Children: make([]*StyledNode, 0),
	}

	// CSS 2.1 §6.2: Inherited properties are passed from parent to child.
	// Per CSS 2.1 property definitions, the following are inherited by default:
	// - color (§14.1), font-* (§15), line-height (§10.8.1)
	// - text-indent, text-align, text-transform (§16.1-16.5)
	// - letter-spacing, word-spacing (§16.4)
	// - white-space (§16.6), list-style-* (§12.5)
	// Note: text-decoration is NOT inherited per CSS 2.1 §16.3.1
	// We implement a subset relevant to current rendering capabilities.
	inheritedProps := []string{
		"color",
		"font-size",
		"font-family",
		"font-weight",
		"font-style",
		"line-height",
		"text-align",
		"text-transform",
		"letter-spacing",
		"word-spacing",
		"white-space",
	}
	
	for _, prop := range inheritedProps {
		if val, ok := parentStyles[prop]; ok {
			styled.Styles[prop] = val
		}
	}

	// Only compute styles for element nodes
	if node.Type == dom.ElementNode {
		// HTML presentational attributes: Convert to CSS styles before applying CSS rules
		// These have lower specificity than CSS rules, so apply them first
		// HTML5 §2.4.4: Presentational hints
		applyPresentationalHints(node, styled.Styles)
		
		// Find all matching rules
		matchedRules := matchRules(node, stylesheet)

		// Apply rules in order of specificity
		for _, matched := range matchedRules {
			for _, decl := range matched.Rule.Declarations {
				applyDeclaration(decl, styled.Styles)
			}
		}
		
		// cellpadding and cellspacing attribute handling
		// HTML5 §14.3.9: These attributes override user-agent defaults
		// Applied after CSS rules but before inline styles (similar to author CSS with higher specificity)
		
		// cellspacing attribute (used on <table>)
		if node.Data == "table" {
			if cellspacing := node.GetAttribute("cellspacing"); cellspacing != "" {
				// Convert to CSS border-spacing (applies to horizontal and vertical spacing)
				styled.Styles["border-spacing"] = cellspacing + "px"
			}
		}
		
		// cellpadding attribute inheritance from table to cells
		if (node.Data == "td" || node.Data == "th") && node.Parent != nil {
			// Walk up to find the containing table
			parent := node.Parent
			for parent != nil {
				if parent.Data == "table" {
					if cellpadding := parent.GetAttribute("cellpadding"); cellpadding != "" {
						// Apply as padding to all sides of the cell
						paddingValue := cellpadding + "px"
						styled.Styles["padding-top"] = paddingValue
						styled.Styles["padding-right"] = paddingValue
						styled.Styles["padding-bottom"] = paddingValue
						styled.Styles["padding-left"] = paddingValue
					}
					break
				}
				parent = parent.Parent
			}
		}
		
		// Apply inline styles last - they have highest specificity
		// CSS 2.1 §6.4.3: Inline styles have specificity A=1, higher than any selector
		if styleAttr := node.GetAttribute("style"); styleAttr != "" {
			inlineDecls := css.ParseInlineStyle(styleAttr)
			for _, decl := range inlineDecls {
				applyDeclaration(decl, styled.Styles)
			}
		}
	}

	// Recursively style children
	for _, child := range node.Children {
		styledChild := styleNode(child, stylesheet, styled.Styles)
		styled.Children = append(styled.Children, styledChild)
	}

	return styled
}

// matchRules finds all CSS rules that match a node.
// Returns rules sorted by specificity (lowest to highest).
// CSS 2.1 §6.4.3
func matchRules(node *dom.Node, stylesheet *css.Stylesheet) []MatchedRule {
	matched := make([]MatchedRule, 0)

	for _, rule := range stylesheet.Rules {
		for _, selector := range rule.Selectors {
			if matchesSelector(node, selector) {
				specificity := calculateSpecificity(selector)
				matched = append(matched, MatchedRule{
					Rule:        rule,
					Specificity: specificity,
				})
				break // Only count each rule once
			}
		}
	}

	// Sort by specificity (simple bubble sort for small lists)
	for i := 0; i < len(matched); i++ {
		for j := i + 1; j < len(matched); j++ {
			if matched[i].Specificity.Compare(matched[j].Specificity) > 0 {
				matched[i], matched[j] = matched[j], matched[i]
			}
		}
	}

	return matched
}

// matchesSelector checks if a node matches a CSS selector.
// This handles simple selectors and descendant combinators.
// CSS 2.1 §5 Selectors
func matchesSelector(node *dom.Node, selector *css.Selector) bool {
	// For descendant selectors, we need to match from right to left
	if len(selector.Simple) == 0 {
		return false
	}

	// Start with the rightmost (most specific) selector
	if !matchesSimpleSelector(node, selector.Simple[len(selector.Simple)-1]) {
		return false
	}

	// If there's only one simple selector, we're done
	if len(selector.Simple) == 1 {
		return true
	}

	// Check for descendant combinators
	return matchesDescendant(node, selector.Simple[:len(selector.Simple)-1])
}

// matchesDescendant checks if a node has ancestors matching the given selectors.
// CSS 2.1 §5.5 Descendant selectors
func matchesDescendant(node *dom.Node, selectors []*css.SimpleSelector) bool {
	if len(selectors) == 0 {
		return true
	}

	// Walk up the tree looking for ancestors that match
	current := node.Parent
	for current != nil {
		if matchesSimpleSelector(current, selectors[len(selectors)-1]) {
			// Found a match, check remaining selectors
			if len(selectors) == 1 {
				return true
			}
			if matchesDescendant(current, selectors[:len(selectors)-1]) {
				return true
			}
		}
		current = current.Parent
	}

	return false
}

// matchesSimpleSelector checks if a node matches a simple selector.
// CSS 2.1 §5.2 Selector syntax
func matchesSimpleSelector(node *dom.Node, selector *css.SimpleSelector) bool {
	// Check element type
	if selector.TagName != "" && selector.TagName != node.Data {
		return false
	}

	// Check ID
	if selector.ID != "" && selector.ID != node.ID() {
		return false
	}

	// Check classes
	if len(selector.Classes) > 0 {
		nodeClasses := node.Classes()
		for _, selectorClass := range selector.Classes {
			found := false
			for _, nodeClass := range nodeClasses {
				if nodeClass == selectorClass {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// Check pseudo-classes for link state
	// CSS 2.1 §5.11.2: :link and :visited pseudo-classes
	// Since we don't track visited state, we treat all links as unvisited (:link)
	// Therefore, selectors with :visited should not match any elements
	if len(selector.PseudoClasses) > 0 {
		for _, pseudoClass := range selector.PseudoClasses {
			if pseudoClass == "visited" {
				// :visited pseudo-class - don't match (treat all links as unvisited)
				return false
			}
			// For other pseudo-classes like :link, :hover, :focus, etc.
			// we ignore them for matching purposes (already counted in specificity)
			// :link matches all <a> elements with href attribute
			if pseudoClass == "link" {
				// Only match if this is an <a> element with href
				if node.Data != "a" || node.GetAttribute("href") == "" {
					return false
				}
			}
			// Other pseudo-classes like :hover, :active, :focus are ignored
			// (they're counted in specificity but don't affect matching)
		}
	}

	return true
}

// calculateSpecificity calculates the specificity of a selector.
// CSS 2.1 §6.4.3 Calculating a selector's specificity
func calculateSpecificity(selector *css.Selector) Specificity {
	spec := Specificity{}

	for _, simple := range selector.Simple {
		if simple.ID != "" {
			spec.B++
		}
		// CSS 2.1 §6.4.3: Class selectors and pseudo-classes count in specificity C
		spec.C += len(simple.Classes)
		spec.C += len(simple.PseudoClasses)
		// CSS 2.1 §6.4.3: Element selectors and pseudo-elements count in specificity D
		if simple.TagName != "" {
			spec.D++
		}
		spec.D += len(simple.PseudoElements)
	}

	return spec
}

// expandShorthand expands CSS shorthand properties to their longhand equivalents.
// CSS 2.1 §8.3, §8.4: Margin and padding shorthand expansion
// CSS 2.1 §8.5: Border shorthand expansion
// Supports 1-4 value patterns per CSS 2.1 specification
func expandShorthand(property, value string) map[string]string {
	result := make(map[string]string)

	// Handle border shorthand properties
	// CSS 2.1 §8.5.4: The border shorthand property
	if property == "border" {
		return expandBorderShorthand(value)
	}

	// Handle border-top, border-right, border-bottom, border-left shorthands
	// CSS 2.1 §8.5.4: Border side shorthands
	switch property {
	case "border-top", "border-right", "border-bottom", "border-left":
		return expandBorderSideShorthand(property, value)
	}

	var prefix string
	switch property {
	case "margin":
		prefix = "margin"
	case "padding":
		prefix = "padding"
	default:
		result[property] = value
		return result
	}

	values := splitWhitespace(value)
	var top, right, bottom, left string

	switch len(values) {
	case 1:
		top = values[0]
		right = values[0]
		bottom = values[0]
		left = values[0]
	case 2:
		top = values[0]
		right = values[1]
		bottom = values[0]
		left = values[1]
	case 3:
		// Top | Horizontal | Bottom
		top = values[0]
		right = values[1]
		bottom = values[2]
		left = values[1]
	case 4:
		top = values[0]
		right = values[1]
		bottom = values[2]
		left = values[3]
	default:
		result[property] = value
		return result
	}

	// Create longhand properties
	result[prefix+"-top"] = top
	result[prefix+"-right"] = right
	result[prefix+"-bottom"] = bottom
	result[prefix+"-left"] = left

	return result
}

// expandBorderShorthand expands the border shorthand property.
// CSS 2.1 §8.5.4: border shorthand sets width, style, and color for all four sides.
// Format: border: [width] [style] [color]
func expandBorderShorthand(value string) map[string]string {
	result := make(map[string]string)
	width, style, color := parseBorderValue(value)

	// Set all four sides
	sides := []string{"top", "right", "bottom", "left"}
	for _, side := range sides {
		if width != "" {
			result["border-"+side+"-width"] = width
		}
		if style != "" {
			result["border-style"] = style // border-style applies to all sides
		}
		if color != "" {
			result["border-color"] = color // border-color applies to all sides
		}
	}

	return result
}

// expandBorderSideShorthand expands a border side shorthand property (e.g., border-top).
// CSS 2.1 §8.5.4: border-top, border-right, etc. set width, style, and color for one side.
// Format: border-side: [width] [style] [color]
func expandBorderSideShorthand(property, value string) map[string]string {
	result := make(map[string]string)
	width, style, color := parseBorderValue(value)

	// Extract side from property name (e.g., "top" from "border-top")
	side := property[7:] // Skip "border-"

	if width != "" {
		result["border-"+side+"-width"] = width
	}
	if style != "" {
		result["border-style"] = style
	}
	if color != "" {
		result["border-color"] = color
	}

	return result
}

// parseBorderValue parses a border shorthand value into width, style, and color.
// CSS 2.1 §8.5.4: The order of values doesn't matter; they are identified by type.
// Returns (width, style, color) - any may be empty if not specified.
func parseBorderValue(value string) (width, style, color string) {
	parts := splitWhitespace(value)

	// CSS 2.1 §8.5.3: border-style valid keywords
	styleKeywords := map[string]bool{
		"none": true, "hidden": true, "dotted": true, "dashed": true,
		"solid": true, "double": true, "groove": true, "ridge": true,
		"inset": true, "outset": true,
	}

	for _, part := range parts {
		partLower := strings.ToLower(part)

		// Check if it's a style keyword
		if styleKeywords[partLower] {
			style = partLower
			continue
		}

		// Check if it's a width value (ends with px, em, or is a number, or is a width keyword)
		if strings.HasSuffix(partLower, "px") || strings.HasSuffix(partLower, "em") ||
			partLower == "thin" || partLower == "medium" || partLower == "thick" {
			width = part
			continue
		}

		// Check if it's a plain number (assume px)
		if _, err := strconv.ParseFloat(part, 64); err == nil {
			width = part
			continue
		}

		// Otherwise, assume it's a color
		color = part
	}

	return width, style, color
}

// splitWhitespace splits a string on whitespace characters.
func splitWhitespace(s string) []string {
	var result []string
	var current string

	for _, ch := range s {
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}

// applyDeclaration applies a CSS declaration to a styles map, expanding shorthand properties.
// CSS 2.1 §8.3, §8.4: Shorthand properties are expanded to their longhand equivalents.
func applyDeclaration(decl *css.Declaration, styles map[string]string) {
	expandedProps := expandShorthand(decl.Property, decl.Value)
	for prop, val := range expandedProps {
		styles[prop] = val
	}
}

// applyPresentationalHints converts HTML presentational attributes to CSS styles.
// HTML5 §2.4.4: Presentational hints
// These attributes have lower specificity than CSS rules.
// Also applies default User-Agent styles for common HTML elements.
func applyPresentationalHints(node *dom.Node, styles map[string]string) {
	// HTML5 §10.3.1: Default styles for phrasing content elements
	// Apply default font styling for text-related elements
	switch node.Data {
	case "strong", "b":
		// HTML5 §10.3.1: strong and b elements are bold by default
		if styles["font-weight"] == "" {
			styles["font-weight"] = "bold"
		}
	case "em", "i":
		// HTML5 §10.3.1: em and i elements are italic by default
		if styles["font-style"] == "" {
			styles["font-style"] = "italic"
		}
	case "u":
		// HTML5 §10.3.1: u element is underlined by default
		if styles["text-decoration"] == "" {
			styles["text-decoration"] = "underline"
		}
	}

	// <font color="..."> attribute
	if node.Data == "font" {
		if color := node.GetAttribute("color"); color != "" {
			styles["color"] = color
		}
	}
	
	// bgcolor attribute (used on <table>, <tr>, <td>, <th>, <body>)
	if bgcolor := node.GetAttribute("bgcolor"); bgcolor != "" {
		styles["background-color"] = bgcolor
	}
	
	// width attribute (used on many elements including <table>, <td>, <img>)
	// HTML5 §14.3.9: Maps to CSS width property
	if width := node.GetAttribute("width"); width != "" {
		// Check if it's a percentage or pixel value
		if strings.Contains(width, "%") {
			styles["width"] = width
		} else {
			// Plain number means pixels
			styles["width"] = width + "px"
		}
	}
	
	// height attribute (used on many elements including <table>, <td>, <img>)
	// HTML5 §14.3.9: Maps to CSS height property
	if height := node.GetAttribute("height"); height != "" {
		// Check if it's a percentage or pixel value
		if strings.Contains(height, "%") {
			styles["height"] = height
		} else {
			// Plain number means pixels
			styles["height"] = height + "px"
		}
	}
	
	// Note: cellspacing and cellpadding are handled after CSS rules are applied
	// to ensure they override user-agent stylesheet defaults
	// align and valign are also handled in the layout phase
}

// ResolveCSSURLs resolves relative URLs in CSS properties against a base URL.
// This handles background-image and other CSS properties that contain URLs.
// Per HTML5 §2.5.1, URLs should be resolved against the document's base URL.
func ResolveCSSURLs(root *StyledNode, baseURL string) {
	if root == nil {
		return
	}
	
	// Resolve URLs in background and background-image properties
	for _, prop := range []string{"background", "background-image"} {
		if value, ok := root.Styles[prop]; ok && strings.Contains(value, "url(") {
			root.Styles[prop] = resolveURLsInValue(value, baseURL)
		}
	}
	
	// Recursively process children
	for _, child := range root.Children {
		ResolveCSSURLs(child, baseURL)
	}
}

// resolveURLsInValue resolves URLs within a CSS property value.
// Handles both url(...) and url("...") and url('...') formats.
func resolveURLsInValue(value, baseURL string) string {
	// Find all url(...) occurrences and resolve them
	result := value
	start := 0
	for {
		idx := strings.Index(result[start:], "url(")
		if idx == -1 {
			break
		}
		idx += start
		
		// Find the end of url(...)
		end := idx + 4 // len("url(")
		parenCount := 1
		inQuote := false
		quoteChar := byte(0)
		
		for end < len(result) && parenCount > 0 {
			ch := result[end]
			if !inQuote {
				if ch == '"' || ch == '\'' {
					inQuote = true
					quoteChar = ch
				} else if ch == '(' {
					parenCount++
				} else if ch == ')' {
					parenCount--
				}
			} else {
				if ch == quoteChar {
					inQuote = false
				}
			}
			end++
		}
		
		// Extract the URL
		urlPart := result[idx+4 : end-1] // Skip "url(" and ")"
		urlPart = strings.TrimSpace(urlPart)
		urlPart = strings.Trim(urlPart, "\"'")
		
		// Resolve the URL using dom package logic
		resolvedURL := dom.ResolveURLString(baseURL, urlPart)
		
		// Replace in the result
		newURLValue := "url(" + resolvedURL + ")"
		result = result[:idx] + newURLValue + result[end:]
		start = idx + len(newURLValue)
	}
	
	return result
}
