package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// RuleConfig represents the configuration file structure
type RuleConfig struct {
	Rules []Rule `yaml:"rules"`
}

// Rule represents a single validation rule
type Rule struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Severity    string   `yaml:"severity"` // ERROR or WARN
	Type        string   `yaml:"type"`     // image, resources, security, etc.
	Conditions  []string `yaml:"conditions"`
	Message     string   `yaml:"message"`
	Help        string   `yaml:"help,omitempty"`
}

// LoadRuleConfig loads rules from a YAML file
func LoadRuleConfig(filepath string) (*RuleConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config RuleConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GetDefaultConfig returns the default rule configuration
func GetDefaultConfig() *RuleConfig {
	return &RuleConfig{
		Rules: []Rule{
			{
				Name:        "no-latest-image",
				Description: "Disallow latest image tags",
				Severity:    "ERROR",
				Type:        "image",
				Conditions:  []string{"image_tag_equals:latest", "image_tag_missing"},
				Message:     "Container '{container}' uses 'latest' image tag",
				Help:        "use a specific version or digest",
			},
			{
				Name:        "require-resource-requests",
				Description: "Require CPU and memory requests",
				Severity:    "WARN",
				Type:        "resources",
				Conditions:  []string{"missing_cpu_requests", "missing_memory_requests"},
				Message:     "Container '{container}' missing resource requests",
				Help:        "set requests.cpu and requests.memory",
			},
			{
				Name:        "require-resource-limits",
				Description: "Require CPU and memory limits",
				Severity:    "WARN",
				Type:        "resources",
				Conditions:  []string{"missing_cpu_limits", "missing_memory_limits"},
				Message:     "Container '{container}' missing resource limits",
				Help:        "set limits.cpu and limits.memory",
			},
			{
				Name:        "no-root-containers",
				Description: "Containers must not run as root",
				Severity:    "ERROR",
				Type:        "security",
				Conditions:  []string{"missing_security_context", "run_as_non_root_false", "run_as_user_zero"},
				Message:     "Container '{container}' running as root or missing securityContext",
				Help:        "set runAsNonRoot: true and runAsUser to non-zero value",
			},
			{
				Name:        "no-privileged-containers",
				Description: "Containers must not run in privileged mode",
				Severity:    "ERROR",
				Type:        "security",
				Conditions:  []string{"privileged_true"},
				Message:     "Container '{container}' is running in privileged mode",
				Help:        "set securityContext.privileged: false or remove the field",
			},
			{
				Name:        "require-liveness-probe",
				Description: "Containers should define a liveness probe",
				Severity:    "WARN",
				Type:        "reliability",
				Conditions:  []string{"missing_liveness_probe"},
				Message:     "Container '{container}' is missing a liveness probe",
				Help:        "add a livenessProbe to detect and restart unhealthy containers",
			},
			{
				Name:        "require-readiness-probe",
				Description: "Containers should define a readiness probe",
				Severity:    "WARN",
				Type:        "reliability",
				Conditions:  []string{"missing_readiness_probe"},
				Message:     "Container '{container}' is missing a readiness probe",
				Help:        "add a readinessProbe to prevent traffic reaching unready containers",
			},
			{
				Name:        "require-image-pull-policy",
				Description: "Containers should explicitly set imagePullPolicy",
				Severity:    "WARN",
				Type:        "image",
				Conditions:  []string{"missing_image_pull_policy"},
				Message:     "Container '{container}' does not set imagePullPolicy",
				Help:        "set imagePullPolicy to Always, IfNotPresent, or Never",
			},
		},
	}
}
