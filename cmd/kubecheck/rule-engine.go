package main

import (
	"strings"
)

// RuleEngine evaluates YAML-defined rules against Kubernetes resources
type RuleEngine struct {
	config *RuleConfig
}

// NewRuleEngine creates a new rule engine with the given config
func NewRuleEngine(config *RuleConfig) *RuleEngine {
	return &RuleEngine{
		config: config,
	}
}

// EvaluateResource evaluates all rules against a Kubernetes resource
func (re *RuleEngine) EvaluateResource(resource K8sResource) []Violation {
	var violations []Violation

	// Extract containers from the resource
	containers := extractContainersFromResource(resource)

	// Evaluate each rule
	for _, rule := range re.config.Rules {
		for _, container := range containers {
			containerViolations := re.evaluateRule(rule, container)
			violations = append(violations, containerViolations...)
		}
	}

	return violations
}

// evaluateRule evaluates a single rule against a container
func (re *RuleEngine) evaluateRule(rule Rule, container Container) []Violation {
	var violations []Violation

	for _, condition := range rule.Conditions {
		if re.checkCondition(condition, container) {
			// Replace {container} placeholder in message
			message := strings.ReplaceAll(rule.Message, "{container}", container.Name)

			violation := Violation{
				Severity: rule.Severity,
				Message:  message,
				Rule:     rule.Name,
			}
			violations = append(violations, violation)
			break // Only report one violation per rule per container
		}
	}

	return violations
}

// checkCondition evaluates a single condition
func (re *RuleEngine) checkCondition(condition string, container Container) bool {
	parts := strings.Split(condition, ":")
	conditionType := parts[0]
	var conditionValue string
	if len(parts) > 1 {
		conditionValue = parts[1]
	}

	switch conditionType {
	case "image_tag_equals":
		return imageTagEquals(container.Image, conditionValue)
	case "image_tag_missing":
		return imageTagMissing(container.Image)
	case "missing_cpu_requests":
		return missingCPURequests(container)
	case "missing_memory_requests":
		return missingMemoryRequests(container)
	case "missing_cpu_limits":
		return missingCPULimits(container)
	case "missing_memory_limits":
		return missingMemoryLimits(container)
	case "missing_security_context":
		return missingSecurityContext(container)
	case "run_as_non_root_false":
		return runAsNonRootFalse(container)
	case "run_as_user_zero":
		return runAsUserZero(container)
	default:
		return false
	}
}

// Container represents a Kubernetes container spec
type Container struct {
	Name            string
	Image           string
	Resources       *Resources
	SecurityContext *SecurityContext
}

// Resources represents resource requirements
type Resources struct {
	Requests *ResourceSpec
	Limits   *ResourceSpec
}

// ResourceSpec represents CPU and memory specs
type ResourceSpec struct {
	CPU    string
	Memory string
}

// SecurityContext represents security settings
type SecurityContext struct {
	RunAsNonRoot *bool
	RunAsUser    *int
}

// Condition evaluation functions
func imageTagEquals(image, tag string) bool {
	if !strings.Contains(image, ":") {
		return tag == "latest" // No tag means implicit :latest
	}
	parts := strings.Split(image, ":")
	return len(parts) == 2 && parts[1] == tag
}

func imageTagMissing(image string) bool {
	return !strings.Contains(image, ":")
}

func missingCPURequests(c Container) bool {
	return c.Resources == nil || c.Resources.Requests == nil || c.Resources.Requests.CPU == ""
}

func missingMemoryRequests(c Container) bool {
	return c.Resources == nil || c.Resources.Requests == nil || c.Resources.Requests.Memory == ""
}

func missingCPULimits(c Container) bool {
	return c.Resources == nil || c.Resources.Limits == nil || c.Resources.Limits.CPU == ""
}

func missingMemoryLimits(c Container) bool {
	return c.Resources == nil || c.Resources.Limits == nil || c.Resources.Limits.Memory == ""
}

func missingSecurityContext(c Container) bool {
	return c.SecurityContext == nil
}

func runAsNonRootFalse(c Container) bool {
	return c.SecurityContext != nil && c.SecurityContext.RunAsNonRoot != nil && !*c.SecurityContext.RunAsNonRoot
}

func runAsUserZero(c Container) bool {
	return c.SecurityContext != nil && c.SecurityContext.RunAsUser != nil && *c.SecurityContext.RunAsUser == 0
}

// extractContainersFromResource extracts containers from a K8s resource
func extractContainersFromResource(resource K8sResource) []Container {
	var containers []Container

	// Navigate through the spec to find containers
	if resource.Spec == nil {
		return containers
	}

	// Try to find containers in spec.template.spec.containers (Deployment, StatefulSet, etc.)
	if template, ok := resource.Spec["template"].(map[string]interface{}); ok {
		if spec, ok := template["spec"].(map[string]interface{}); ok {
			if containerList, ok := spec["containers"].([]interface{}); ok {
				containers = parseContainers(containerList)
				return containers
			}
		}
	}

	// Try to find containers directly in spec.containers (Pod)
	if containerList, ok := resource.Spec["containers"].([]interface{}); ok {
		containers = parseContainers(containerList)
		return containers
	}

	return containers
}

// parseContainers converts interface{} to Container structs
func parseContainers(containerList []interface{}) []Container {
	var containers []Container

	for _, c := range containerList {
		containerMap, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		container := Container{
			Name:  getStringValue(containerMap, "name"),
			Image: getStringValue(containerMap, "image"),
		}

		// Parse resources
		if resourcesMap, ok := containerMap["resources"].(map[string]interface{}); ok {
			container.Resources = parseResources(resourcesMap)
		}

		// Parse security context
		if securityMap, ok := containerMap["securityContext"].(map[string]interface{}); ok {
			container.SecurityContext = parseSecurityContext(securityMap)
		}

		containers = append(containers, container)
	}

	return containers
}

// parseResources parses resource requirements
func parseResources(resourcesMap map[string]interface{}) *Resources {
	resources := &Resources{}

	if requestsMap, ok := resourcesMap["requests"].(map[string]interface{}); ok {
		resources.Requests = &ResourceSpec{
			CPU:    getStringValue(requestsMap, "cpu"),
			Memory: getStringValue(requestsMap, "memory"),
		}
	}

	if limitsMap, ok := resourcesMap["limits"].(map[string]interface{}); ok {
		resources.Limits = &ResourceSpec{
			CPU:    getStringValue(limitsMap, "cpu"),
			Memory: getStringValue(limitsMap, "memory"),
		}
	}

	return resources
}

// parseSecurityContext parses security context
func parseSecurityContext(securityMap map[string]interface{}) *SecurityContext {
	sc := &SecurityContext{}

	if runAsNonRoot, ok := securityMap["runAsNonRoot"].(bool); ok {
		sc.RunAsNonRoot = &runAsNonRoot
	}

	if runAsUser, ok := securityMap["runAsUser"].(int); ok {
		sc.RunAsUser = &runAsUser
	}

	return sc
}

// getStringValue safely gets a string value from a map
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
