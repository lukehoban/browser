// Package dom provides the Document Object Model tree structure.
// It represents the parsed HTML document as a tree of nodes.
//
// Spec references:
// - DOM Level 2 Core: https://www.w3.org/TR/DOM-Level-2-Core/
package dom

// NodeType represents the type of a DOM node.
type NodeType int

const (
	// ElementNode represents an HTML element (e.g., <div>, <p>)
	ElementNode NodeType = iota
	// TextNode represents text content within an element
	TextNode
	// DocumentNode represents the root document node
	DocumentNode
)

// Node represents a node in the DOM tree.
type Node struct {
	Type       NodeType
	Data       string            // Tag name for elements, text content for text nodes
	Attributes map[string]string // Attributes for element nodes
	Children   []*Node           // Child nodes
	Parent     *Node             // Parent node (nil for root)
}

// NewElement creates a new element node with the given tag name.
func NewElement(tagName string) *Node {
	return &Node{
		Type:       ElementNode,
		Data:       tagName,
		Attributes: make(map[string]string),
		Children:   make([]*Node, 0),
	}
}

// NewText creates a new text node with the given content.
func NewText(text string) *Node {
	return &Node{
		Type:     TextNode,
		Data:     text,
		Children: make([]*Node, 0),
	}
}

// NewDocument creates a new document root node.
func NewDocument() *Node {
	return &Node{
		Type:     DocumentNode,
		Data:     "#document",
		Children: make([]*Node, 0),
	}
}

// AppendChild adds a child node to this node.
func (n *Node) AppendChild(child *Node) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

// GetAttribute returns the value of an attribute, or empty string if not found.
func (n *Node) GetAttribute(name string) string {
	if n.Attributes == nil {
		return ""
	}
	return n.Attributes[name]
}

// SetAttribute sets an attribute on this node.
func (n *Node) SetAttribute(name, value string) {
	if n.Attributes == nil {
		n.Attributes = make(map[string]string)
	}
	n.Attributes[name] = value
}

// ID returns the element's ID attribute.
func (n *Node) ID() string {
	return n.GetAttribute("id")
}

// Classes returns the element's class names as a slice.
func (n *Node) Classes() []string {
	class := n.GetAttribute("class")
	if class == "" {
		return nil
	}
	// Simple space-separated class parsing
	classes := []string{}
	start := 0
	for i := 0; i <= len(class); i++ {
		if i == len(class) || class[i] == ' ' {
			if i > start {
				classes = append(classes, class[start:i])
			}
			start = i + 1
		}
	}
	return classes
}
