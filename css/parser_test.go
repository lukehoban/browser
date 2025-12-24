package css

import "testing"

func TestParseSimpleRule(t *testing.T) {
	input := "div { color: red; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	rule := stylesheet.Rules[0]
	if len(rule.Selectors) != 1 {
		t.Fatalf("Expected 1 selector, got %d", len(rule.Selectors))
	}

	selector := rule.Selectors[0]
	if len(selector.Simple) != 1 {
		t.Fatalf("Expected 1 simple selector, got %d", len(selector.Simple))
	}

	simple := selector.Simple[0]
	if simple.TagName != "div" {
		t.Errorf("Expected tag 'div', got %v", simple.TagName)
	}

	if len(rule.Declarations) != 1 {
		t.Fatalf("Expected 1 declaration, got %d", len(rule.Declarations))
	}

	decl := rule.Declarations[0]
	if decl.Property != "color" {
		t.Errorf("Expected property 'color', got %v", decl.Property)
	}
	if decl.Value != "red" {
		t.Errorf("Expected value 'red', got %v", decl.Value)
	}
}

func TestParseIDSelector(t *testing.T) {
	input := "#header { font-size: 20px; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	simple := stylesheet.Rules[0].Selectors[0].Simple[0]
	if simple.ID != "header" {
		t.Errorf("Expected ID 'header', got %v", simple.ID)
	}
}

func TestParseClassSelector(t *testing.T) {
	input := ".container { width: 100px; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	simple := stylesheet.Rules[0].Selectors[0].Simple[0]
	if len(simple.Classes) != 1 {
		t.Fatalf("Expected 1 class, got %d", len(simple.Classes))
	}
	if simple.Classes[0] != "container" {
		t.Errorf("Expected class 'container', got %v", simple.Classes[0])
	}
}

func TestParseCombinedSelector(t *testing.T) {
	input := "div#main.container { margin: 10px; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	simple := stylesheet.Rules[0].Selectors[0].Simple[0]
	if simple.TagName != "div" {
		t.Errorf("Expected tag 'div', got %v", simple.TagName)
	}
	if simple.ID != "main" {
		t.Errorf("Expected ID 'main', got %v", simple.ID)
	}
	if len(simple.Classes) != 1 || simple.Classes[0] != "container" {
		t.Errorf("Expected class 'container', got %v", simple.Classes)
	}
}

func TestParseMultipleClasses(t *testing.T) {
	input := ".container.active { display: block; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	simple := stylesheet.Rules[0].Selectors[0].Simple[0]
	if len(simple.Classes) != 2 {
		t.Fatalf("Expected 2 classes, got %d", len(simple.Classes))
	}
	if simple.Classes[0] != "container" {
		t.Errorf("Expected first class 'container', got %v", simple.Classes[0])
	}
	if simple.Classes[1] != "active" {
		t.Errorf("Expected second class 'active', got %v", simple.Classes[1])
	}
}

func TestParseDescendantSelector(t *testing.T) {
	input := "div p { color: blue; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	selector := stylesheet.Rules[0].Selectors[0]
	if len(selector.Simple) != 2 {
		t.Fatalf("Expected 2 simple selectors (descendant), got %d", len(selector.Simple))
	}

	if selector.Simple[0].TagName != "div" {
		t.Errorf("Expected first selector 'div', got %v", selector.Simple[0].TagName)
	}
	if selector.Simple[1].TagName != "p" {
		t.Errorf("Expected second selector 'p', got %v", selector.Simple[1].TagName)
	}
}

func TestParseMultipleSelectors(t *testing.T) {
	input := "h1, h2, h3 { font-weight: bold; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	rule := stylesheet.Rules[0]
	if len(rule.Selectors) != 3 {
		t.Fatalf("Expected 3 selectors, got %d", len(rule.Selectors))
	}

	tags := []string{"h1", "h2", "h3"}
	for i, tag := range tags {
		if rule.Selectors[i].Simple[0].TagName != tag {
			t.Errorf("Expected selector %d to be '%s', got %v", i, tag, rule.Selectors[i].Simple[0].TagName)
		}
	}
}

func TestParseMultipleDeclarations(t *testing.T) {
	input := "div { color: red; background: blue; margin: 10px; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	rule := stylesheet.Rules[0]
	if len(rule.Declarations) != 3 {
		t.Fatalf("Expected 3 declarations, got %d", len(rule.Declarations))
	}

	expected := map[string]string{
		"color":      "red",
		"background": "blue",
		"margin":     "10px",
	}

	for _, decl := range rule.Declarations {
		expectedValue, ok := expected[decl.Property]
		if !ok {
			t.Errorf("Unexpected property: %v", decl.Property)
			continue
		}
		if decl.Value != expectedValue {
			t.Errorf("Property %v: expected value %v, got %v", decl.Property, expectedValue, decl.Value)
		}
	}
}

func TestParseMultipleRules(t *testing.T) {
	input := `
		div { color: red; }
		p { font-size: 14px; }
		.container { width: 100%; }
	`
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 3 {
		t.Fatalf("Expected 3 rules, got %d", len(stylesheet.Rules))
	}
}

func TestParseComplexValue(t *testing.T) {
	input := "div { border: 1px solid black; }"
	stylesheet := Parse(input)

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(stylesheet.Rules))
	}

	decl := stylesheet.Rules[0].Declarations[0]
	if decl.Property != "border" {
		t.Errorf("Expected property 'border', got %v", decl.Property)
	}
	if decl.Value != "1px solid black" {
		t.Errorf("Expected value '1px solid black', got %v", decl.Value)
	}
}

// TestParseAttributeSelector tests that attribute selectors are skipped gracefully.
// CSS 2.1 §5.8 Attribute selectors
func TestParseAttributeSelector(t *testing.T) {
	input := `
input[type='submit'] { font-family: Verdana; }
.class { color: red; }
`
	stylesheet := Parse(input)

	// Should parse successfully and have at least the .class rule
	if len(stylesheet.Rules) < 1 {
		t.Errorf("Expected at least 1 rule, got %d", len(stylesheet.Rules))
	}

	// The .class rule should be parsed correctly
	foundClassRule := false
	for _, rule := range stylesheet.Rules {
		if len(rule.Selectors) > 0 && len(rule.Selectors[0].Simple) > 0 {
			simple := rule.Selectors[0].Simple[0]
			if len(simple.Classes) > 0 && simple.Classes[0] == "class" {
				foundClassRule = true
				break
			}
		}
	}

	if !foundClassRule {
		t.Error("Expected .class rule to be parsed")
	}
}

// TestParseAtRule tests that @-rules are skipped gracefully.
// CSS 2.1 §4.1.5 At-rules
func TestParseAtRule(t *testing.T) {
	input := `
body { color: black; }
@media screen and (max-width: 600px) {
body { color: blue; }
}
.test { color: red; }
`
	stylesheet := Parse(input)

	// Should parse successfully and have the body and .test rules
	if len(stylesheet.Rules) < 2 {
		t.Errorf("Expected at least 2 rules, got %d", len(stylesheet.Rules))
	}

	// Check that we have body and .test rules
	foundBody := false
	foundTest := false
	for _, rule := range stylesheet.Rules {
		if len(rule.Selectors) > 0 && len(rule.Selectors[0].Simple) > 0 {
			simple := rule.Selectors[0].Simple[0]
			if simple.TagName == "body" {
				foundBody = true
			}
			if len(simple.Classes) > 0 && simple.Classes[0] == "test" {
				foundTest = true
			}
		}
	}

	if !foundBody {
		t.Error("Expected body rule to be parsed")
	}
	if !foundTest {
		t.Error("Expected .test rule to be parsed")
	}
}
