package dom

import "testing"

func TestNewElement(t *testing.T) {
	elem := NewElement("div")
	if elem.Type != ElementNode {
		t.Errorf("Expected ElementNode, got %v", elem.Type)
	}
	if elem.Data != "div" {
		t.Errorf("Expected tag name 'div', got %v", elem.Data)
	}
	if elem.Attributes == nil {
		t.Error("Expected attributes map to be initialized")
	}
	if elem.Children == nil {
		t.Error("Expected children slice to be initialized")
	}
}

func TestNewText(t *testing.T) {
	text := NewText("Hello, World!")
	if text.Type != TextNode {
		t.Errorf("Expected TextNode, got %v", text.Type)
	}
	if text.Data != "Hello, World!" {
		t.Errorf("Expected text 'Hello, World!', got %v", text.Data)
	}
}

func TestAppendChild(t *testing.T) {
	parent := NewElement("div")
	child := NewElement("p")

	parent.AppendChild(child)

	if len(parent.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(parent.Children))
	}
	if parent.Children[0] != child {
		t.Error("Child not properly appended")
	}
	if child.Parent != parent {
		t.Error("Child's parent not set correctly")
	}
}

func TestAttributes(t *testing.T) {
	elem := NewElement("div")
	elem.SetAttribute("id", "main")
	elem.SetAttribute("class", "container")

	if elem.GetAttribute("id") != "main" {
		t.Errorf("Expected id 'main', got %v", elem.GetAttribute("id"))
	}
	if elem.GetAttribute("class") != "container" {
		t.Errorf("Expected class 'container', got %v", elem.GetAttribute("class"))
	}
	if elem.GetAttribute("nonexistent") != "" {
		t.Error("Expected empty string for nonexistent attribute")
	}
}

func TestID(t *testing.T) {
	elem := NewElement("div")
	elem.SetAttribute("id", "header")

	if elem.ID() != "header" {
		t.Errorf("Expected ID 'header', got %v", elem.ID())
	}
}

func TestClasses(t *testing.T) {
	tests := []struct {
		name     string
		class    string
		expected []string
	}{
		{
			name:     "single class",
			class:    "container",
			expected: []string{"container"},
		},
		{
			name:     "multiple classes",
			class:    "container main active",
			expected: []string{"container", "main", "active"},
		},
		{
			name:     "empty class",
			class:    "",
			expected: nil,
		},
		{
			name:     "class with extra spaces",
			class:    "  container  main  ",
			expected: []string{"container", "main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elem := NewElement("div")
			if tt.class != "" {
				elem.SetAttribute("class", tt.class)
			}

			classes := elem.Classes()
			if len(classes) != len(tt.expected) {
				t.Errorf("Expected %d classes, got %d", len(tt.expected), len(classes))
				return
			}

			for i, class := range classes {
				if class != tt.expected[i] {
					t.Errorf("Expected class[%d] = %v, got %v", i, tt.expected[i], class)
				}
			}
		})
	}
}

func TestNewDocument(t *testing.T) {
doc := NewDocument()
if doc.Type != DocumentNode {
t.Errorf("Expected DocumentNode type, got %v", doc.Type)
}
if doc.Data != "#document" {
t.Errorf("Expected data '#document', got %q", doc.Data)
}
if len(doc.Children) != 0 {
t.Errorf("Expected empty children, got %d", len(doc.Children))
}
}

func TestGetAttributeNilAttributes(t *testing.T) {
// Node with nil attributes map should return empty string
node := &Node{
Type:       ElementNode,
Data:       "div",
Attributes: nil,
}
if got := node.GetAttribute("class"); got != "" {
t.Errorf("Expected empty string for nil attributes, got %q", got)
}
}

func TestSetAttributeNilAttributes(t *testing.T) {
// SetAttribute on a node with nil attributes should create the map
node := &Node{
Type:       ElementNode,
Data:       "div",
Attributes: nil,
}
node.SetAttribute("id", "test")
if got := node.GetAttribute("id"); got != "test" {
t.Errorf("Expected 'test', got %q", got)
}
}

func TestAppendChildSetsParent(t *testing.T) {
parent := NewElement("div")
child1 := NewElement("p")
child2 := NewText("hello")

parent.AppendChild(child1)
parent.AppendChild(child2)

if child1.Parent != parent {
t.Error("Expected child1.Parent to be parent")
}
if child2.Parent != parent {
t.Error("Expected child2.Parent to be parent")
}
if len(parent.Children) != 2 {
t.Errorf("Expected 2 children, got %d", len(parent.Children))
}
}

func TestIDReturnsEmptyForNoID(t *testing.T) {
node := NewElement("div")
if got := node.ID(); got != "" {
t.Errorf("Expected empty ID, got %q", got)
}
}

func TestClassesMultipleSpaces(t *testing.T) {
node := NewElement("div")
node.SetAttribute("class", "  foo   bar  baz  ")
classes := node.Classes()
expected := []string{"foo", "bar", "baz"}
if len(classes) != len(expected) {
t.Fatalf("Expected %d classes, got %d: %v", len(expected), len(classes), classes)
}
for i, c := range classes {
if c != expected[i] {
t.Errorf("Class[%d] = %q, want %q", i, c, expected[i])
}
}
}
