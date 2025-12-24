// Package reftest provides a test harness for running WPT (Web Platform Tests)
// reference tests against this browser implementation.
//
// Reference tests (reftests) compare the rendered output of a test page against
// a reference page. If they render identically, the test passes.
//
// See: https://web-platform-tests.org/writing-tests/reftests.html
package reftest

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	cssparser "github.com/lukehoban/browser/css"
	"github.com/lukehoban/browser/html"
	"github.com/lukehoban/browser/layout"
	"github.com/lukehoban/browser/style"
)

// Result represents the outcome of a single reftest.
type Result struct {
	TestFile      string
	ReferenceFile string
	RelationType  string // "match" or "mismatch"
	Status        Status
	Message       string
}

// Status represents the status of a test.
type Status int

const (
	// Pass indicates the test passed.
	Pass Status = iota
	// Fail indicates the test failed.
	Fail
	// Error indicates an error occurred running the test.
	Error
	// Skip indicates the test was skipped (e.g., unsupported feature).
	Skip
)

func (s Status) String() string {
	switch s {
	case Pass:
		return "PASS"
	case Fail:
		return "FAIL"
	case Error:
		return "ERROR"
	case Skip:
		return "SKIP"
	default:
		return "UNKNOWN"
	}
}

// Summary provides aggregate statistics for a test run.
type Summary struct {
	Total   int
	Passed  int
	Failed  int
	Errors  int
	Skipped int
	Results []Result
}

// PassRate returns the percentage of tests that passed.
func (s *Summary) PassRate() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.Passed) / float64(s.Total) * 100
}

// Runner executes reference tests.
type Runner struct {
	baseDir string
	verbose bool
}

// NewRunner creates a new reftest runner.
func NewRunner(baseDir string, verbose bool) *Runner {
	return &Runner{
		baseDir: baseDir,
		verbose: verbose,
	}
}

// RunTest runs a single reftest.
func (r *Runner) RunTest(testPath string) Result {
	result := Result{
		TestFile: testPath,
	}

	// Read the test file
	testContent, err := os.ReadFile(testPath)
	if err != nil {
		result.Status = Error
		result.Message = fmt.Sprintf("failed to read test file: %v", err)
		return result
	}

	// Find the reference link
	refPath, relType, err := findReferenceLink(string(testContent), testPath)
	if err != nil {
		result.Status = Skip
		result.Message = fmt.Sprintf("no reference link found: %v", err)
		return result
	}

	result.ReferenceFile = refPath
	result.RelationType = relType

	// Read the reference file
	refContent, err := os.ReadFile(refPath)
	if err != nil {
		result.Status = Error
		result.Message = fmt.Sprintf("failed to read reference file: %v", err)
		return result
	}

	// Render both pages and compare
	match, err := r.compareLayouts(string(testContent), string(refContent))
	if err != nil {
		result.Status = Error
		result.Message = fmt.Sprintf("layout comparison failed: %v", err)
		return result
	}

	// Determine pass/fail based on relation type
	if relType == "match" {
		if match {
			result.Status = Pass
			result.Message = "layouts match as expected"
		} else {
			result.Status = Fail
			result.Message = "layouts do not match"
		}
	} else { // mismatch
		if !match {
			result.Status = Pass
			result.Message = "layouts differ as expected"
		} else {
			result.Status = Fail
			result.Message = "layouts unexpectedly match"
		}
	}

	return result
}

// RunDirectory runs all reftests in a directory.
func (r *Runner) RunDirectory(dir string) Summary {
	summary := Summary{
		Results: make([]Result, 0),
	}

	// Find all HTML test files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip reference files (files containing "-ref")
		if strings.Contains(filepath.Base(path), "-ref") {
			return nil
		}

		// Skip non-HTML files
		if !strings.HasSuffix(path, ".html") && !strings.HasSuffix(path, ".htm") {
			return nil
		}

		// Check if this is a test file (contains a reference link)
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		if hasReferenceLink(string(content)) {
			result := r.RunTest(path)
			summary.Results = append(summary.Results, result)
			summary.Total++

			switch result.Status {
			case Pass:
				summary.Passed++
			case Fail:
				summary.Failed++
			case Error:
				summary.Errors++
			case Skip:
				summary.Skipped++
			}

			if r.verbose {
				fmt.Printf("[%s] %s\n", result.Status, path)
				if result.Message != "" {
					fmt.Printf("        %s\n", result.Message)
				}
			}
		}

		return nil
	})

	if err != nil && r.verbose {
		fmt.Printf("Error walking directory: %v\n", err)
	}

	return summary
}

// compareLayouts renders both HTML documents and compares their layouts.
func (r *Runner) compareLayouts(testHTML, refHTML string) (bool, error) {
	// Parse and layout test document
	testLayout, err := renderDocument(testHTML)
	if err != nil {
		return false, fmt.Errorf("failed to render test: %w", err)
	}

	// Parse and layout reference document
	refLayout, err := renderDocument(refHTML)
	if err != nil {
		return false, fmt.Errorf("failed to render reference: %w", err)
	}

	// Compare layouts
	return compareLayoutTrees(testLayout, refLayout), nil
}

// renderDocument parses HTML, extracts styles, and performs layout.
func renderDocument(htmlContent string) (*layout.LayoutBox, error) {
	// Parse HTML
	doc := html.Parse(htmlContent)

	// Extract CSS from <style> elements
	cssContent := extractCSS(htmlContent)

	// Parse CSS
	stylesheet := cssparser.Parse(cssContent)

	// Compute styles
	styledTree := style.StyleTree(doc, stylesheet)

	// Layout
	containingBlock := layout.Dimensions{}
	layoutTree := layout.LayoutTree(styledTree, containingBlock)

	return layoutTree, nil
}

// extractCSS extracts CSS content from <style> elements in HTML.
func extractCSS(htmlContent string) string {
	// Simple regex to extract style content
	// This is a simplified approach - a full implementation would use the DOM
	re := regexp.MustCompile(`(?is)<style[^>]*>(.*?)</style>`)
	matches := re.FindAllStringSubmatch(htmlContent, -1)

	var css strings.Builder
	for _, match := range matches {
		if len(match) > 1 {
			css.WriteString(match[1])
			css.WriteString("\n")
		}
	}

	return css.String()
}

// findReferenceLink finds the <link rel="match|mismatch" href="..."> in the HTML.
func findReferenceLink(htmlContent, testPath string) (string, string, error) {
	// Look for <link rel="match" href="..."> or <link rel="mismatch" href="...">
	re := regexp.MustCompile(`(?i)<link[^>]+rel\s*=\s*["'](match|mismatch)["'][^>]+href\s*=\s*["']([^"']+)["']`)
	matches := re.FindStringSubmatch(htmlContent)

	if len(matches) < 3 {
		// Try alternative order: href before rel
		re = regexp.MustCompile(`(?i)<link[^>]+href\s*=\s*["']([^"']+)["'][^>]+rel\s*=\s*["'](match|mismatch)["']`)
		matches = re.FindStringSubmatch(htmlContent)
		if len(matches) < 3 {
			return "", "", fmt.Errorf("no reference link found")
		}
		// Swap order since href comes first in this pattern
		matches = []string{matches[0], matches[2], matches[1]}
	}

	relType := strings.ToLower(matches[1])
	refHref := matches[2]

	// Resolve relative path
	testDir := filepath.Dir(testPath)
	refPath := filepath.Join(testDir, refHref)

	return refPath, relType, nil
}

// hasReferenceLink checks if HTML content contains a reference link.
func hasReferenceLink(htmlContent string) bool {
	re := regexp.MustCompile(`(?i)<link[^>]+rel\s*=\s*["'](match|mismatch)["']`)
	return re.MatchString(htmlContent)
}

// compareLayoutTrees compares two layout trees for visual equality.
// This focuses on the body element's content, ignoring metadata elements.
func compareLayoutTrees(a, b *layout.LayoutBox) bool {
	// Find the body elements in both trees
	bodyA := findBodyLayout(a)
	bodyB := findBodyLayout(b)

	if bodyA == nil && bodyB == nil {
		// No body elements, compare whole trees
		return compareLayoutBoxes(a, b)
	}

	if bodyA == nil || bodyB == nil {
		return false
	}

	// Compare dimensions of the body elements
	return compareDimensions(bodyA.Dimensions, bodyB.Dimensions)
}

// findBodyLayout finds the layout box for the <body> element.
func findBodyLayout(box *layout.LayoutBox) *layout.LayoutBox {
	if box == nil {
		return nil
	}

	if box.StyledNode != nil && box.StyledNode.Node != nil {
		if box.StyledNode.Node.Data == "body" {
			return box
		}
	}

	for _, child := range box.Children {
		if result := findBodyLayout(child); result != nil {
			return result
		}
	}

	return nil
}

// compareLayoutBoxes compares two layout boxes for structural equality.
func compareLayoutBoxes(a, b *layout.LayoutBox) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare box types
	if a.BoxType != b.BoxType {
		return false
	}

	// Compare dimensions with tolerance for floating point
	if !compareDimensions(a.Dimensions, b.Dimensions) {
		return false
	}

	// Compare number of children
	if len(a.Children) != len(b.Children) {
		return false
	}

	// Recursively compare children
	for i := range a.Children {
		if !compareLayoutBoxes(a.Children[i], b.Children[i]) {
			return false
		}
	}

	return true
}

// compareDimensions compares two dimension structures with tolerance.
func compareDimensions(a, b layout.Dimensions) bool {
	const tolerance = 0.1

	if !floatEqual(a.Content.X, b.Content.X, tolerance) ||
		!floatEqual(a.Content.Y, b.Content.Y, tolerance) ||
		!floatEqual(a.Content.Width, b.Content.Width, tolerance) ||
		!floatEqual(a.Content.Height, b.Content.Height, tolerance) {
		return false
	}

	if !compareEdges(a.Padding, b.Padding, tolerance) ||
		!compareEdges(a.Border, b.Border, tolerance) ||
		!compareEdges(a.Margin, b.Margin, tolerance) {
		return false
	}

	return true
}

// compareEdges compares two EdgeSizes with tolerance.
func compareEdges(a, b layout.EdgeSizes, tolerance float64) bool {
	return floatEqual(a.Top, b.Top, tolerance) &&
		floatEqual(a.Right, b.Right, tolerance) &&
		floatEqual(a.Bottom, b.Bottom, tolerance) &&
		floatEqual(a.Left, b.Left, tolerance)
}

// floatEqual compares two floats with tolerance.
func floatEqual(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}

// PrintSummary prints a human-readable summary of test results.
func PrintSummary(summary Summary) {
	fmt.Println("\n========================================")
	fmt.Println("WPT Reftest Summary")
	fmt.Println("========================================")
	fmt.Printf("Total:   %d\n", summary.Total)
	fmt.Printf("Passed:  %d (%.1f%%)\n", summary.Passed, summary.PassRate())
	fmt.Printf("Failed:  %d\n", summary.Failed)
	fmt.Printf("Errors:  %d\n", summary.Errors)
	fmt.Printf("Skipped: %d\n", summary.Skipped)
	fmt.Println("========================================")

	// Print failed tests
	if summary.Failed > 0 {
		fmt.Println("\nFailed Tests:")
		for _, r := range summary.Results {
			if r.Status == Fail {
				fmt.Printf("  - %s: %s\n", r.TestFile, r.Message)
			}
		}
	}

	// Print errors
	if summary.Errors > 0 {
		fmt.Println("\nTests with Errors:")
		for _, r := range summary.Results {
			if r.Status == Error {
				fmt.Printf("  - %s: %s\n", r.TestFile, r.Message)
			}
		}
	}
}
