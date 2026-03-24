// Package dom provides the Document Object Model tree structure.
// It represents the parsed HTML document as a tree of nodes.
//
// Spec references:
// - DOM Level 2 Core: https://www.w3.org/TR/DOM-Level-2-Core/
package dom

import "strings"

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

// RemoveChild removes a child node from this node and returns it.
// Returns nil if the child is not found.
func (n *Node) RemoveChild(child *Node) *Node {
	for i, c := range n.Children {
		if c == child {
			n.Children = append(n.Children[:i], n.Children[i+1:]...)
			child.Parent = nil
			return child
		}
	}
	return nil
}

// InsertBefore inserts newChild before refChild. If refChild is nil,
// newChild is appended at the end.
func (n *Node) InsertBefore(newChild, refChild *Node) {
	if refChild == nil {
		n.AppendChild(newChild)
		return
	}
	for i, c := range n.Children {
		if c == refChild {
			newChild.Parent = n
			// Insert at position i
			n.Children = append(n.Children[:i], append([]*Node{newChild}, n.Children[i:]...)...)
			return
		}
	}
	// refChild not found, append
	n.AppendChild(newChild)
}

// GetElementByID searches the subtree for an element with the given ID.
// Returns nil if not found.
func (n *Node) GetElementByID(id string) *Node {
	if n.Type == ElementNode && n.ID() == id {
		return n
	}
	for _, child := range n.Children {
		if found := child.GetElementByID(id); found != nil {
			return found
		}
	}
	return nil
}

// GetElementsByTagName returns all descendant elements with the given tag name.
func (n *Node) GetElementsByTagName(tagName string) []*Node {
	var results []*Node
	tagName = strings.ToLower(tagName)
	n.getElementsByTagNameHelper(tagName, &results)
	return results
}

func (n *Node) getElementsByTagNameHelper(tagName string, results *[]*Node) {
	if n.Type == ElementNode && strings.ToLower(n.Data) == tagName {
		*results = append(*results, n)
	}
	for _, child := range n.Children {
		child.getElementsByTagNameHelper(tagName, results)
	}
}

// GetElementsByClassName returns all descendant elements with the given class name.
func (n *Node) GetElementsByClassName(className string) []*Node {
	var results []*Node
	n.getElementsByClassNameHelper(className, &results)
	return results
}

func (n *Node) getElementsByClassNameHelper(className string, results *[]*Node) {
	if n.Type == ElementNode {
		for _, c := range n.Classes() {
			if c == className {
				*results = append(*results, n)
				break
			}
		}
	}
	for _, child := range n.Children {
		child.getElementsByClassNameHelper(className, results)
	}
}

// TextContent returns the text content of the node and its descendants.
func (n *Node) TextContent() string {
	if n.Type == TextNode {
		return n.Data
	}
	var sb strings.Builder
	for _, child := range n.Children {
		sb.WriteString(child.TextContent())
	}
	return sb.String()
}

// SetTextContent replaces the node's children with a single text node.
func (n *Node) SetTextContent(text string) {
	// Clear existing children
	for _, child := range n.Children {
		child.Parent = nil
	}
	n.Children = make([]*Node, 0)
	if text != "" {
		n.AppendChild(NewText(text))
	}
}
