package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// K8sResource represents a Kubernetes resource
type K8sResource struct {
	APIVersion string                 `json:"apiVersion" yaml:"apiVersion"`
	Kind       string                 `json:"kind" yaml:"kind"`
	Metadata   map[string]interface{} `json:"metadata" yaml:"metadata"`
	Spec       map[string]interface{} `json:"spec" yaml:"spec"`
	Data       map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`
}

// parseYAMLFile parses a YAML file and returns Kubernetes resources
func parseYAMLFile(filename string) ([]K8sResource, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return parseYAML(data)
}

// parseYAML parses YAML data and returns Kubernetes resources
// Handles multi-document YAML (--- separated)
func parseYAML(data []byte) ([]K8sResource, error) {
	var resources []K8sResource

	// Split by document separator
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	for {
		var resource K8sResource
		err := decoder.Decode(&resource)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to decode YAML: %w", err)
		}

		// Skip empty documents
		if resource.Kind == "" {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// processStdin reads YAML from stdin
func processStdin() ([]string, error) {
	tmpFile, err := os.CreateTemp("", "kubecheck-*.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	// Read from stdin
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Fprintln(tmpFile, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}

	return []string{tmpFile.Name()}, nil
}

// processDirectory recursively finds YAML files in a directory
func processDirectory(dir string) ([]string, error) {
	var files []string

	err := walkDir(dir, func(path string, info os.FileInfo) error {
		if info.IsDir() {
			return nil
		}

		// Check if it's a YAML file
		if isYAMLFile(path) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// walkDir walks a directory tree
func walkDir(root string, fn func(string, os.FileInfo) error) error {
	info, err := os.Stat(root)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return fn(root, info)
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		path := root + string(os.PathSeparator) + entry.Name()
		entryInfo, err := entry.Info()
		if err != nil {
			continue
		}

		if entry.IsDir() {
			if err := walkDir(path, fn); err != nil {
				return err
			}
		} else {
			if err := fn(path, entryInfo); err != nil {
				return err
			}
		}
	}

	return nil
}

// isYAMLFile checks if a file has a YAML extension
func isYAMLFile(filename string) bool {
	ext := strings.ToLower(filename[len(filename)-min(5, len(filename)):])
	return strings.HasSuffix(ext, ".yaml") || strings.HasSuffix(ext, ".yml")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
