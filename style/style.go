// Package style handles style computation and the CSS cascade.
// It matches CSS selectors to DOM elements and computes final styles.
//
// Spec references:
// - CSS 2.1 §6 Assigning property values, Cascading, and Inheritance
package style

import (
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
	A int // Inline styles (not used in this implementation)
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
func StyleTree(root *dom.Node, stylesheet *css.Stylesheet) *StyledNode {
	return styleNodeWithParent(root, stylesheet, nil)
}

// styleNode computes styles for a single node and its children.
func styleNode(node *dom.Node, stylesheet *css.Stylesheet) *StyledNode {
	return styleNodeWithParent(node, stylesheet, nil)
}

// styleNodeWithParent computes styles for a single node with parent context.
// CSS 2.1 §6.2: Inheritance - text nodes inherit from their parent element
func styleNodeWithParent(node *dom.Node, stylesheet *css.Stylesheet, parentStyles map[string]string) *StyledNode {
	styled := &StyledNode{
		Node:     node,
		Styles:   make(map[string]string),
		Children: make([]*StyledNode, 0),
	}

	// Only compute styles for element nodes
	if node.Type == dom.ElementNode {
		// Find all matching rules
		matchedRules := matchRules(node, stylesheet)

		// Apply rules in order of specificity
		for _, matched := range matchedRules {
			for _, decl := range matched.Rule.Declarations {
				// Expand shorthand properties
				// CSS 2.1 §8.3 Margin properties, §8.4 Padding properties
				expandedProps := expandShorthand(decl.Property, decl.Value)
				for prop, val := range expandedProps {
					styled.Styles[prop] = val
				}
			}
		}
	} else if node.Type == dom.TextNode && parentStyles != nil {
		// For text nodes, inherit specific properties from parent
		// CSS 2.1 §6.2: Inherited properties
		inheritedProperties := []string{
			"color", "font-family", "font-size", "font-weight", "font-style",
			"line-height", "text-align", "text-indent", "white-space",
		}
		for _, prop := range inheritedProperties {
			if val, ok := parentStyles[prop]; ok {
				styled.Styles[prop] = val
			}
		}
	}

	// Recursively style children, passing current element's styles as parent
	for _, child := range node.Children {
		styledChild := styleNodeWithParent(child, stylesheet, styled.Styles)
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
		spec.C += len(simple.Classes)
		if simple.TagName != "" {
			spec.D++
		}
	}

	return spec
}

// expandShorthand expands CSS shorthand properties to their longhand equivalents.
// CSS 2.1 §8.3 Margin properties, §8.4 Padding properties
//
// Supported shorthand properties:
//   - margin: Expands to margin-top, margin-right, margin-bottom, margin-left
//   - padding: Expands to padding-top, padding-right, padding-bottom, padding-left
//
// The value patterns follow CSS 2.1 specification:
//   - 1 value: applies to all four sides (e.g., "10px" → all sides 10px)
//   - 2 values: vertical | horizontal (e.g., "10px 20px" → top/bottom 10px, left/right 20px)
//   - 3 values: top | horizontal | bottom (e.g., "10px 20px 30px")
//   - 4 values: top | right | bottom | left (e.g., "10px 20px 30px 40px")
func expandShorthand(property, value string) map[string]string {
	result := make(map[string]string)

	// Check if this is a shorthand property
	var prefix string
	switch property {
	case "margin":
		prefix = "margin"
	case "padding":
		prefix = "padding"
	default:
		// Not a shorthand property, return as-is
		result[property] = value
		return result
	}

	// Parse the value into individual components
	// Split on whitespace to get individual values
	values := splitWhitespace(value)

	var top, right, bottom, left string

	switch len(values) {
	case 1:
		// All four sides
		top = values[0]
		right = values[0]
		bottom = values[0]
		left = values[0]
	case 2:
		// Vertical | Horizontal
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
		// Top | Right | Bottom | Left
		top = values[0]
		right = values[1]
		bottom = values[2]
		left = values[3]
	default:
		// Invalid number of values, return as-is
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
