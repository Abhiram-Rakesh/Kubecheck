package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ExitOK    = 0
	ExitWarn  = 1
	ExitError = 2
)

type Config struct {
	Verbose bool
}

func main() {
	// Parse command line flags
	verbose := flag.Bool("v", false, "Verbose output")
	flag.Parse()

	config := Config{
		Verbose: *verbose,
	}

	// Get input path(s)
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: kubecheck [options] <file|directory|helm-chart|->")
		fmt.Fprintln(os.Stderr, "Options:")
		flag.PrintDefaults()
		os.Exit(ExitError)
	}

	input := args[0]

	// Process input
	var files []string
	var err error

	if input == "-" {
		// Read from stdin
		files, err = processStdin()
	} else if isHelmChart(input) {
		// Helm chart
		files, err = processHelmChart(input)
	} else if isDirectory(input) {
		// Directory
		files, err = processDirectory(input)
	} else {
		// Single file
		files = []string{input}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing input: %v\n", err)
		os.Exit(ExitError)
	}

	// Validate all files
	maxSeverity := ExitOK
	reporter := NewReporter(config.Verbose)

	// Enable directory mode if processing multiple files
	if len(files) > 1 || isDirectory(input) {
		reporter.SetDirectoryMode(true)
		if isDirectory(input) {
			reporter.PrintDirectoryHeader(input)
		}
	}

	for _, file := range files {
		resources, err := parseYAMLFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", file, err)
			continue
		}

		for _, resource := range resources {
			violations, err := validateResource(resource)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error validating %s: %v\n", file, err)
				continue
			}

			severity := reporter.ReportViolations(file, resource, violations)
			if severity > maxSeverity {
				maxSeverity = severity
			}
		}
	}

	reporter.PrintSummary()
	os.Exit(maxSeverity)
}

// isHelmChart checks if the path is a Helm chart directory
func isHelmChart(path string) bool {
	chartPath := filepath.Join(path, "Chart.yaml")
	_, err := os.Stat(chartPath)
	return err == nil
}

// isDirectory checks if the path is a directory
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// validateResource calls the Haskell rule engine
func validateResource(resource K8sResource) ([]Violation, error) {
	// Prepare JSON input for Haskell rule engine
	input, err := json.Marshal(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource: %w", err)
	}

	// Call Haskell rule engine
	output, err := callRuleEngine(input)
	if err != nil {
		return nil, fmt.Errorf("rule engine error: %w", err)
	}

	// Parse violations
	var violations []Violation
	if err := json.Unmarshal(output, &violations); err != nil {
		return nil, fmt.Errorf("failed to parse violations: %w", err)
	}

	return violations, nil
}
