package reftest

import (
	"path/filepath"
	"runtime"
	"testing"
)

// TestWPTCSSReftests runs the WPT CSS reference tests as a benchmark.
// This test discovers and runs all reftest files in the test/wpt/css directory.
func TestWPTCSSReftests(t *testing.T) {
	// Get the path to the test/wpt/css directory relative to this test file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to get current file path")
	}

	baseDir := filepath.Dir(filename)
	wptDir := filepath.Join(baseDir, "..", "test", "wpt", "css")

	runner := NewRunner(wptDir, true)
	summary := runner.RunDirectory(wptDir)

	// Log summary
	t.Logf("WPT CSS Reftest Results:")
	t.Logf("  Total:   %d", summary.Total)
	t.Logf("  Passed:  %d (%.1f%%)", summary.Passed, summary.PassRate())
	t.Logf("  Failed:  %d", summary.Failed)
	t.Logf("  Errors:  %d", summary.Errors)
	t.Logf("  Skipped: %d", summary.Skipped)

	// Log failed tests for debugging
	if summary.Failed > 0 {
		t.Logf("\nFailed tests (expected, documenting gaps):")
		for _, result := range summary.Results {
			if result.Status == Fail {
				t.Logf("  - %s: %s", filepath.Base(result.TestFile), result.Message)
			}
		}
	}

	// Note: We don't fail the test on reftest failures since this is a benchmark
	// to track progress. Instead, we document expected failures.
	expectedFailures := map[string]bool{
		// All tests currently pass!
	}

	unexpectedFailures := 0
	for _, result := range summary.Results {
		if result.Status == Fail {
			testName := filepath.Base(result.TestFile)
			if !expectedFailures[testName] {
				t.Errorf("Unexpected failure: %s - %s", testName, result.Message)
				unexpectedFailures++
			}
		}
		if result.Status == Error {
			t.Errorf("Test error: %s - %s", filepath.Base(result.TestFile), result.Message)
		}
	}

	if unexpectedFailures > 0 {
		t.Errorf("%d unexpected failures detected", unexpectedFailures)
	}
}

// BenchmarkWPTCSSReftests benchmarks the WPT CSS reftest suite.
func BenchmarkWPTCSSReftests(b *testing.B) {
	// Get the path to the test/wpt/css directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		b.Fatal("failed to get current file path")
	}

	baseDir := filepath.Dir(filename)
	wptDir := filepath.Join(baseDir, "..", "test", "wpt", "css")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		runner := NewRunner(wptDir, false)
		runner.RunDirectory(wptDir)
	}
}
