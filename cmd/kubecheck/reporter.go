package main

import (
	"fmt"
	"strings"
)

// Severity levels
const (
	SeverityOK    = "OK"
	SeverityWarn  = "WARN"
	SeverityError = "ERROR"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[90m"
	ColorBold   = "\033[1m"
)

// Box-drawing characters
const (
	BoxTopLeft     = "┌"
	BoxTopRight    = "┐"
	BoxBottomLeft  = "└"
	BoxBottomRight = "┘"
	BoxHorizontal  = "─"
	BoxVertical    = "│"
	BoxDivider     = "━"
)

// Symbols
const (
	SymbolError   = "✖"
	SymbolWarning = "⚠"
	SymbolOK      = "✔"
	SymbolPointer = "▲"
	SymbolArrow   = "➔"
	SymbolBullet  = "●"
	SymbolTree    = "└─"
)

// Violation represents a single validation violation
type Violation struct {
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Rule     string `json:"rule"`
}

// Reporter handles output formatting and violation tracking
type Reporter struct {
	verbose         bool
	totalFiles      int
	okFiles         int
	warnFiles       int
	errorFiles      int
	totalViolations int
	isDirectory     bool
}

// NewReporter creates a new reporter
func NewReporter(verbose bool) *Reporter {
	return &Reporter{
		verbose: verbose,
	}
}

// SetDirectoryMode enables directory scanning mode
func (r *Reporter) SetDirectoryMode(enabled bool) {
	r.isDirectory = enabled
}

// ReportViolations reports violations for a resource and returns the highest severity
func (r *Reporter) ReportViolations(filename string, resource K8sResource, violations []Violation) int {
	r.totalFiles++

	if len(violations) == 0 {
		r.okFiles++
		if r.verbose || !r.isDirectory {
			r.printOK(filename, resource)
		}
		return ExitOK
	}

	// Count violations by severity
	errorCount := 0
	warnCount := 0
	for _, v := range violations {
		r.totalViolations++
		if v.Severity == SeverityError {
			errorCount++
		} else if v.Severity == SeverityWarn {
			warnCount++
		}
	}

	maxSeverity := ExitOK
	if errorCount > 0 {
		maxSeverity = ExitError
		r.errorFiles++
	} else if warnCount > 0 {
		maxSeverity = ExitWarn
		r.warnFiles++
	}

	// Print violations based on mode
	if r.isDirectory {
		r.printDirectoryViolations(filename, resource, violations, errorCount, warnCount)
	} else {
		r.printFileViolations(filename, resource, violations, errorCount, warnCount)
	}

	return maxSeverity
}

// printOK prints success message
func (r *Reporter) printOK(filename string, resource K8sResource) {
	if r.isDirectory {
		// Compact format for directory mode
		fmt.Printf("  %s%s%s  %s %s PASSED%s\n",
			ColorGreen, SymbolOK, ColorReset,
			filename,
			strings.Repeat(".", max(1, 50-len(filename))),
			ColorGray)
		if r.verbose {
			resourceName := getResourceName(resource)
			if resourceName != "" {
				fmt.Printf("     %s Resource: %s/%s%s\n",
					ColorGray, resource.Kind, resourceName, ColorReset)
			}
		}
	} else {
		// Detailed format for single file
		fmt.Printf("\n  %s%s File: %s%s\n", ColorBold, SymbolBullet, filename, ColorReset)
		resourceName := getResourceName(resource)
		if resourceName != "" {
			fmt.Printf("  %s%s %s: %s %s\n",
				ColorGreen, BoxTopLeft, resource.Kind, resourceName,
				strings.Repeat(BoxHorizontal, max(1, 60-len(resource.Kind)-len(resourceName))))
			fmt.Printf("  %s  %s%s All checks passed%s\n",
				BoxVertical, ColorGreen, SymbolOK, ColorReset)
			fmt.Printf("  %s%s\n", ColorGreen, BoxBottomLeft+strings.Repeat(BoxHorizontal, 68))
		}
	}
}

// printFileViolations prints violations in detailed box format (single file mode)
func (r *Reporter) printFileViolations(filename string, resource K8sResource, violations []Violation, errorCount, warnCount int) {
	resourceName := getResourceName(resource)
	title := fmt.Sprintf(" %s: %s ", resource.Kind, resourceName)
	padding := max(1, 68-len(title))

	fmt.Printf("\n  %s%s File: %s%s\n", ColorBold, SymbolBullet, filename, ColorReset)
	fmt.Printf("  %s%s%s%s%s\n",
		ColorCyan, BoxTopLeft, BoxHorizontal, title,
		strings.Repeat(BoxHorizontal, padding)+BoxTopRight)

	// Group violations by type
	errorViolations := []Violation{}
	warnViolations := []Violation{}

	for _, v := range violations {
		if v.Severity == SeverityError {
			errorViolations = append(errorViolations, v)
		} else {
			warnViolations = append(warnViolations, v)
		}
	}

	// Print errors first
	for i, v := range errorViolations {
		if i > 0 {
			fmt.Printf("  %s%s%s\n", ColorCyan, BoxVertical, ColorReset)
		}
		r.printViolationDetail(v, BoxVertical)
	}

	// Print warnings
	for i, v := range warnViolations {
		if i > 0 || len(errorViolations) > 0 {
			fmt.Printf("  %s%s%s\n", ColorCyan, BoxVertical, ColorReset)
		}
		r.printViolationDetail(v, BoxVertical)
	}

	// Bottom border with summary
	summary := fmt.Sprintf(" [ %d errors | %d warns ] ", errorCount, warnCount)
	summaryPadding := max(1, 70-len(summary))
	fmt.Printf("  %s%s%s%s%s\n",
		ColorCyan, BoxBottomLeft,
		strings.Repeat(BoxHorizontal, summaryPadding),
		summary, BoxBottomRight+ColorReset)
}

// printDirectoryViolations prints violations in compact format (directory mode)
func (r *Reporter) printDirectoryViolations(filename string, resource K8sResource, violations []Violation, errorCount, warnCount int) {
	// Determine status symbol and color
	symbol := SymbolWarning
	color := ColorYellow
	status := fmt.Sprintf("%d WARN", warnCount)

	if errorCount > 0 {
		symbol = SymbolError
		color = ColorRed
		status = fmt.Sprintf("%d ERR", errorCount)
	}

	// Print file status line
	dots := strings.Repeat(".", max(1, 50-len(filename)))
	fmt.Printf("  %s%s%s  %s %s %s\n",
		color, symbol, ColorReset,
		filename, dots, status)

	// Print violations in compact tree format
	for i, v := range violations {
		isLast := i == len(violations)-1
		resourceName := getResourceName(resource)

		if i == 0 {
			fmt.Printf("     %s [%s] %s%s\n",
				ColorGray+SymbolTree, resourceName, v.Message, ColorReset)
		} else if isLast && v.Severity == SeverityError {
			// Show pointer for errors
			fmt.Printf("        %s> %s%s\n",
				ColorGray, v.Message, ColorReset)
		} else {
			fmt.Printf("        %s%s\n", ColorGray+v.Message, ColorReset)
		}
	}
}

// printViolationDetail prints a single violation with detailed formatting
func (r *Reporter) printViolationDetail(v Violation, border string) {
	var symbol, color, label string

	if v.Severity == SeverityError {
		symbol = SymbolError
		color = ColorRed
		label = "Security Violation"
	} else {
		symbol = SymbolWarning
		color = ColorYellow
		label = "Resource Hygiene"
	}

	fmt.Printf("  %s%s  %s%s  %s%s\n",
		ColorCyan, border, color, symbol, label, ColorReset)
	fmt.Printf("  %s%s     %s%s%s\n",
		ColorCyan, border, ColorBold, v.Message, ColorReset)

	// Add helpful pointer or suggestion
	if v.Rule == "no-latest-image" {
		fmt.Printf("  %s%s     %s%s use a specific version or digest%s\n",
			ColorCyan, border, ColorGray, SymbolPointer+"───", ColorReset)
	} else if v.Rule == "no-root-containers" {
		fmt.Printf("  %s%s     %shelp: set 'runAsNonRoot: true' to improve pod security%s\n",
			ColorCyan, border, ColorGray, ColorReset)
	}
}

// PrintSummary prints the final summary
func (r *Reporter) PrintSummary() {
	if r.totalFiles == 0 {
		return
	}

	fmt.Println()

	if r.isDirectory {
		// Directory mode summary with divider
		fmt.Printf("  %s\n\n", strings.Repeat(BoxDivider, 70))
		fmt.Printf("  Summary %s %d files checked\n", SymbolArrow, r.totalFiles)
		fmt.Printf("  Result  %s ", SymbolArrow)

		if r.okFiles > 0 {
			fmt.Printf("%s%d OK%s", ColorGreen, r.okFiles, ColorReset)
		}
		if r.warnFiles > 0 {
			if r.okFiles > 0 {
				fmt.Print("  |  ")
			}
			fmt.Printf("%s%d Warning%s", ColorYellow, r.warnFiles, ColorReset)
		}
		if r.errorFiles > 0 {
			if r.okFiles > 0 || r.warnFiles > 0 {
				fmt.Print("  |  ")
			}
			fmt.Printf("%s%d Error%s", ColorRed, r.errorFiles, ColorReset)
		}
		fmt.Println()

		// Final status
		if r.errorFiles > 0 {
			fmt.Printf("  Status  %s %sFAILED%s Exit code: 2\n",
				SymbolArrow, ColorRed+ColorBold, ColorReset)
		} else if r.warnFiles > 0 {
			fmt.Printf("  Status  %s %sPASSED WITH WARNINGS%s Exit code: 1\n",
				SymbolArrow, ColorYellow+ColorBold, ColorReset)
		} else {
			fmt.Printf("  Status  %s %sPASSED%s Exit code: 0\n",
				SymbolArrow, ColorGreen+ColorBold, ColorReset)
		}

		fmt.Printf("\n  %s\n", strings.Repeat(BoxDivider, 70))
	} else {
		// Single file mode summary
		fmt.Printf("\n  Summary %s %d file checked. %s%d violation%s found.%s\n",
			SymbolArrow, r.totalFiles,
			ColorBold, r.totalViolations, pluralize(r.totalViolations), ColorReset)
	}
}

// PrintDirectoryHeader prints the header for directory scanning
func (r *Reporter) PrintDirectoryHeader(dir string) {
	fmt.Printf("\n  🔍 Scanning directory: %s\n", dir)
	fmt.Printf("  %s\n\n", strings.Repeat(BoxDivider, 70))
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

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}
