// Command wptrunner runs WPT reference tests against the browser implementation.
//
// Usage:
//
//	wptrunner [options] <directory>
//
// Options:
//
//	-v        Verbose output
//	-json     Output results as JSON
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/lukehoban/browser/reftest"
)

func main() {
	verbose := flag.Bool("v", false, "verbose output")
	jsonOutput := flag.Bool("json", false, "output results as JSON")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <directory>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	dir := flag.Arg(0)

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: directory %s does not exist\n", dir)
		os.Exit(1)
	}

	runner := reftest.NewRunner(dir, *verbose)
	summary := runner.RunDirectory(dir)

	if *jsonOutput {
		// Output as JSON
		output, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to marshal JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
	} else {
		// Print human-readable summary
		reftest.PrintSummary(summary)
	}

	// Exit with non-zero status if there are failures
	if summary.Failed > 0 || summary.Errors > 0 {
		os.Exit(1)
	}
}
