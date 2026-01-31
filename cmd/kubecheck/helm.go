package main

import (
	"fmt"
	"os"
	"os/exec"
)

// processHelmChart renders a Helm chart and returns temporary YAML files
func processHelmChart(chartPath string) ([]string, error) {
	// Check if helm is installed
	if !isHelmInstalled() {
		return nil, fmt.Errorf("helm is not installed. Please install Helm to validate charts")
	}

	fmt.Printf("Rendering Helm chart: %s\n", chartPath)

	// Create temp directory for rendered templates
	tmpDir, err := os.MkdirTemp("", "kubecheck-helm-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Run helm template
	cmd := exec.Command("helm", "template", chartPath, "--output-dir", tmpDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("helm template failed: %s\n%s", err, output)
	}

	// Find all rendered YAML files
	var files []string
	err = walkDir(tmpDir, func(path string, info os.FileInfo) error {
		if !info.IsDir() && isYAMLFile(path) {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no YAML files found in rendered chart")
	}

	return files, nil
}

// isHelmInstalled checks if helm command is available
func isHelmInstalled() bool {
	_, err := exec.LookPath("helm")
	return err == nil
}
