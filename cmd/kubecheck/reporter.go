package main

import (
	"fmt"
)

// Severity levels
const (
	SeverityOK    = "OK"
	SeverityWarn  = "WARN"
	SeverityError = "ERROR"
)

// Violation represents a single validation violation
type Violation struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Rule     string `json:"rule"`
}

// Reporter handles output formatting and violation tracking
type Reporter struct {
	verbose      bool
	totalFiles   int
	okFiles      int
	warnFiles    int
	errorFiles   int
	totalViolations int
}

// NewReporter creates a new reporter
func NewReporter(verbose bool) *Reporter {
	return &Reporter{
		verbose: verbose,
	}
}

// ReportViolations reports violations for a resource and returns the highest severity
func (r *Reporter) ReportViolations(filename string, resource K8sResource, violations []Violation) int {
	r.totalFiles++

	if len(violations) == 0 {
		r.okFiles++
		if r.verbose {
			fmt.Printf("OK: %s\n", filename)
		}
		return ExitOK
	}

	// Determine highest severity
	maxSeverity := ExitOK
	hasError := false
	hasWarn := false

	for _, v := range violations {
		r.totalViolations++
		if v.Severity == SeverityError {
			hasError = true
		} else if v.Severity == SeverityWarn {
			hasWarn = true
		}
	}

	if hasError {
		maxSeverity = ExitError
		r.errorFiles++
		fmt.Printf("\n%s: %s\n", SeverityError, filename)
	} else if hasWarn {
		maxSeverity = ExitWarn
		r.warnFiles++
		fmt.Printf("\n%s: %s\n", SeverityWarn, filename)
	}

	// Print resource info
	resourceName := getResourceName(resource)
	if resourceName != "" {
		fmt.Printf("  Resource: %s/%s\n", resource.Kind, resourceName)
	}

	// Print violations
	for _, v := range violations {
		fmt.Printf("  [%s] %s\n", v.Severity, v.Message)
	}

	return maxSeverity
}

// PrintSummary prints the final summary
func (r *Reporter) PrintSummary() {
	if r.totalFiles == 0 {
		return
	}

	fmt.Println("\n" + "─────────────────────────────────────")
	fmt.Printf("Summary: %d files checked\n", r.totalFiles)
	
	if r.okFiles > 0 {
		fmt.Printf("  ✓ %d OK\n", r.okFiles)
	}
	if r.warnFiles > 0 {
		fmt.Printf("  ⚠ %d with warnings\n", r.warnFiles)
	}
	if r.errorFiles > 0 {
		fmt.Printf("  ✗ %d with errors\n", r.errorFiles)
	}
	
	fmt.Printf("Total violations: %d\n", r.totalViolations)
	fmt.Println("─────────────────────────────────────")
}

// getResourceName extracts the name from metadata
func getResourceName(resource K8sResource) string {
	if resource.Metadata == nil {
		return ""
	}

	if name, ok := resource.Metadata["name"].(string); ok {
		return name
	}

	return ""
}
