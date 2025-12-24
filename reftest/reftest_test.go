package reftest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindReferenceLink(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		wantRelType string
		wantFound   bool
	}{
		{
			name:        "match link with rel first",
			html:        `<link rel="match" href="test-ref.html">`,
			wantRelType: "match",
			wantFound:   true,
		},
		{
			name:        "mismatch link with rel first",
			html:        `<link rel="mismatch" href="test-ref.html">`,
			wantRelType: "mismatch",
			wantFound:   true,
		},
		{
			name:        "match link with href first",
			html:        `<link href="test-ref.html" rel="match">`,
			wantRelType: "match",
			wantFound:   true,
		},
		{
			name:        "no reference link",
			html:        `<link rel="stylesheet" href="styles.css">`,
			wantRelType: "",
			wantFound:   false,
		},
		{
			name:        "link in full HTML",
			html:        `<!DOCTYPE html><html><head><link rel="match" href="ref.html"></head><body></body></html>`,
			wantRelType: "match",
			wantFound:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a temporary test path for reference resolution
			_, relType, err := findReferenceLink(tt.html, "/tmp/test.html")

			if tt.wantFound && err != nil {
				t.Errorf("expected to find reference link, got error: %v", err)
			}
			if !tt.wantFound && err == nil {
				t.Errorf("expected no reference link, but found one")
			}
			if tt.wantFound && relType != tt.wantRelType {
				t.Errorf("expected relType %q, got %q", tt.wantRelType, relType)
			}
		})
	}
}

func TestHasReferenceLink(t *testing.T) {
	tests := []struct {
		name string
		html string
		want bool
	}{
		{
			name: "has match link",
			html: `<link rel="match" href="ref.html">`,
			want: true,
		},
		{
			name: "has mismatch link",
			html: `<link rel="mismatch" href="ref.html">`,
			want: true,
		},
		{
			name: "stylesheet link",
			html: `<link rel="stylesheet" href="styles.css">`,
			want: false,
		},
		{
			name: "empty",
			html: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasReferenceLink(tt.html)
			if got != tt.want {
				t.Errorf("hasReferenceLink() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractCSS(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "single style block",
			html: `<style>body { color: red; }</style>`,
			want: "body { color: red; }\n",
		},
		{
			name: "multiple style blocks",
			html: `<style>body { color: red; }</style><style>.foo { width: 100px; }</style>`,
			want: "body { color: red; }\n.foo { width: 100px; }\n",
		},
		{
			name: "no style block",
			html: `<div>Hello</div>`,
			want: "",
		},
		{
			name: "style in full HTML",
			html: `<!DOCTYPE html><html><head><style>div { margin: 10px; }</style></head><body><div></div></body></html>`,
			want: "div { margin: 10px; }\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCSS(tt.html)
			if got != tt.want {
				t.Errorf("extractCSS() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFloatEqual(t *testing.T) {
	tests := []struct {
		a, b      float64
		tolerance float64
		want      bool
	}{
		{1.0, 1.0, 0.1, true},
		{1.0, 1.05, 0.1, true},
		{1.0, 1.15, 0.1, false},
		{0.0, 0.0, 0.1, true},
		{-1.0, -1.05, 0.1, true},
	}

	for _, tt := range tests {
		got := floatEqual(tt.a, tt.b, tt.tolerance)
		if got != tt.want {
			t.Errorf("floatEqual(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.tolerance, got, tt.want)
		}
	}
}

func TestStatus_String(t *testing.T) {
	tests := []struct {
		status Status
		want   string
	}{
		{Pass, "PASS"},
		{Fail, "FAIL"},
		{Error, "ERROR"},
		{Skip, "SKIP"},
	}

	for _, tt := range tests {
		if got := tt.status.String(); got != tt.want {
			t.Errorf("Status.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestSummary_PassRate(t *testing.T) {
	tests := []struct {
		name    string
		summary Summary
		want    float64
	}{
		{
			name:    "empty",
			summary: Summary{Total: 0, Passed: 0},
			want:    0,
		},
		{
			name:    "all passed",
			summary: Summary{Total: 10, Passed: 10},
			want:    100,
		},
		{
			name:    "half passed",
			summary: Summary{Total: 10, Passed: 5},
			want:    50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.summary.PassRate()
			if got != tt.want {
				t.Errorf("PassRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunner_RunTest(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()

	// Create a matching test
	testFile := filepath.Join(tmpDir, "test.html")
	refFile := filepath.Join(tmpDir, "test-ref.html")

	testHTML := `<!DOCTYPE html>
<html>
<head>
<link rel="match" href="test-ref.html">
<style>
div { width: 100px; height: 100px; }
</style>
</head>
<body>
<div></div>
</body>
</html>`

	refHTML := `<!DOCTYPE html>
<html>
<head>
<style>
div { width: 100px; height: 100px; }
</style>
</head>
<body>
<div></div>
</body>
</html>`

	err := os.WriteFile(testFile, []byte(testHTML), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err = os.WriteFile(refFile, []byte(refHTML), 0644)
	if err != nil {
		t.Fatalf("failed to write reference file: %v", err)
	}

	runner := NewRunner(tmpDir, false)
	result := runner.RunTest(testFile)

	if result.Status != Pass {
		t.Errorf("expected Pass, got %v: %s", result.Status, result.Message)
	}
	if result.RelationType != "match" {
		t.Errorf("expected match relation, got %s", result.RelationType)
	}
}

func TestRunner_RunTestMismatch(t *testing.T) {
	// Create temporary test files
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.html")
	refFile := filepath.Join(tmpDir, "test-ref.html")

	// These should produce different layouts
	testHTML := `<!DOCTYPE html>
<html>
<head>
<link rel="mismatch" href="test-ref.html">
<style>
div { width: 100px; height: 100px; }
</style>
</head>
<body>
<div></div>
</body>
</html>`

	refHTML := `<!DOCTYPE html>
<html>
<head>
<style>
div { width: 200px; height: 200px; }
</style>
</head>
<body>
<div></div>
</body>
</html>`

	err := os.WriteFile(testFile, []byte(testHTML), 0644)
	if err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	err = os.WriteFile(refFile, []byte(refHTML), 0644)
	if err != nil {
		t.Fatalf("failed to write reference file: %v", err)
	}

	runner := NewRunner(tmpDir, false)
	result := runner.RunTest(testFile)

	if result.Status != Pass {
		t.Errorf("expected Pass for mismatch test, got %v: %s", result.Status, result.Message)
	}
}

func TestRunner_RunDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple test files
	tests := []struct {
		name     string
		testHTML string
		refHTML  string
	}{
		{
			name: "test1",
			testHTML: `<!DOCTYPE html>
<html>
<head>
<link rel="match" href="test1-ref.html">
<style>div { width: 50px; }</style>
</head>
<body>
<div></div>
</body>
</html>`,
			refHTML: `<!DOCTYPE html>
<html>
<head>
<style>div { width: 50px; }</style>
</head>
<body>
<div></div>
</body>
</html>`,
		},
		{
			name: "test2",
			testHTML: `<!DOCTYPE html>
<html>
<head>
<link rel="match" href="test2-ref.html">
<style>p { margin: 10px; }</style>
</head>
<body>
<p></p>
</body>
</html>`,
			refHTML: `<!DOCTYPE html>
<html>
<head>
<style>p { margin: 10px; }</style>
</head>
<body>
<p></p>
</body>
</html>`,
		},
	}

	for _, tt := range tests {
		testFile := filepath.Join(tmpDir, tt.name+".html")
		refFile := filepath.Join(tmpDir, tt.name+"-ref.html")

		err := os.WriteFile(testFile, []byte(tt.testHTML), 0644)
		if err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}
		err = os.WriteFile(refFile, []byte(tt.refHTML), 0644)
		if err != nil {
			t.Fatalf("failed to write ref file: %v", err)
		}
	}

	runner := NewRunner(tmpDir, false)
	summary := runner.RunDirectory(tmpDir)

	if summary.Total != 2 {
		t.Errorf("expected 2 tests, got %d", summary.Total)
	}
	if summary.Passed != 2 {
		t.Errorf("expected 2 passed, got %d", summary.Passed)
	}
}
