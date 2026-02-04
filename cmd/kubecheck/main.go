package main

import (
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
	configFile := flag.String("config", "", "Path to kubecheck config file (default: ./kubecheck.yaml or ~/.kubecheck/config.yaml)")
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

	// Load rule configuration
	var ruleConfig *RuleConfig
	if *configFile != "" {
		// User specified a config file
		cfg, err := LoadRuleConfig(*configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config file: %v\n", err)
			os.Exit(ExitError)
		}
		ruleConfig = cfg
		if config.Verbose {
			fmt.Printf("Using config file: %s\n", *configFile)
		}
	} else {
		// Try default locations
		configPaths := []string{
			"./kubecheck.yaml",
			"./kubecheck.yml",
			filepath.Join(os.Getenv("HOME"), ".kubecheck", "config.yaml"),
			filepath.Join(os.Getenv("HOME"), ".kubecheck", "config.yml"),
		}

		foundConfig := false
		for _, path := range configPaths {
			if _, err := os.Stat(path); err == nil {
				cfg, err := LoadRuleConfig(path)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error loading config file %s: %v\n", path, err)
					os.Exit(ExitError)
				}
				ruleConfig = cfg
				foundConfig = true
				if config.Verbose {
					fmt.Printf("Using config file: %s\n", path)
				}
				break
			}
		}

		if !foundConfig {
			// Use default built-in rules
			ruleConfig = GetDefaultConfig()
			if config.Verbose {
				fmt.Println("Using built-in default rules")
			}
		}
	}

	// Create rule engine
	ruleEngine := NewRuleEngine(ruleConfig)

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
			// Use rule engine to evaluate
			violations := ruleEngine.EvaluateResource(resource)

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
