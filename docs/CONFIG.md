# Configuration Guide

kubecheck supports YAML-based configuration files, allowing organizations to define custom validation rules.

## Configuration File Locations

kubecheck will automatically look for configuration files in the following order:

1. `./kubecheck.yaml` (current directory)
2. `./kubecheck.yml` (current directory)
3. `~/.kubecheck/config.yaml` (user home directory)
4. `~/.kubecheck/config.yml` (user home directory)

You can also specify a custom config file:

```bash
kubecheck --config /path/to/custom-config.yaml deployment.yaml
```

If no config file is found, kubecheck uses built-in default rules.

## Configuration Format

```yaml
rules:
  - name: rule-name
    description: Human-readable description
    severity: ERROR  # or WARN
    type: category   # image, resources, security, etc.
    conditions:
      - condition_type:value
      - another_condition
    message: "Error message with {container} placeholder"
    help: "Helpful suggestion for fixing the issue"
```

## Available Conditions

### Image Conditions

- `image_tag_equals:TAG` - Image tag equals specified value
- `image_tag_missing` - No tag specified (implicit :latest)

### Resource Conditions

- `missing_cpu_requests` - No CPU requests specified
- `missing_memory_requests` - No memory requests specified
- `missing_cpu_limits` - No CPU limits specified
- `missing_memory_limits` - No memory limits specified

### Security Conditions

- `missing_security_context` - No securityContext defined
- `run_as_non_root_false` - runAsNonRoot is set to false
- `run_as_user_zero` - runAsUser is set to 0 (root)

## Example Configuration

### Minimal Configuration

```yaml
rules:
  - name: no-latest
    description: Prevent latest tags
    severity: ERROR
    type: image
    conditions:
      - image_tag_equals:latest
    message: "Container '{container}' uses latest tag"
```

### Organization-Specific Rules

```yaml
rules:
  # Security: No latest tags
  - name: no-latest-image
    description: Disallow latest image tags
    severity: ERROR
    type: image
    conditions:
      - image_tag_equals:latest
      - image_tag_missing
    message: "Container '{container}' uses 'latest' image tag"
    help: "use a specific version or digest"

  # Security: Non-root containers
  - name: no-root-containers
    description: Containers must not run as root
    severity: ERROR
    type: security
    conditions:
      - missing_security_context
      - run_as_non_root_false
      - run_as_user_zero
    message: "Container '{container}' may run as root"
    help: "set runAsNonRoot: true"

  # Resources: Require requests
  - name: require-resource-requests
    description: All containers must have resource requests
    severity: WARN
    type: resources
    conditions:
      - missing_cpu_requests
      - missing_memory_requests
    message: "Container '{container}' missing resource requests"
    help: "set resources.requests.cpu and resources.requests.memory"

  # Resources: Require limits
  - name: require-resource-limits
    description: All containers must have resource limits
    severity: WARN
    type: resources
    conditions:
      - missing_cpu_limits
      - missing_memory_limits
    message: "Container '{container}' missing resource limits"
    help: "set resources.limits.cpu and resources.limits.memory"
```

### Team-Specific Configuration

Different teams can have different rules:

```bash
# DevOps team (strict)
kubecheck --config .kubecheck/devops-rules.yaml k8s/

# Development team (relaxed)
kubecheck --config .kubecheck/dev-rules.yaml k8s/
```

## Severity Levels

### ERROR
- Causes kubecheck to exit with code 2
- Fails CI/CD pipelines
- Should be used for critical production requirements

### WARN
- Causes kubecheck to exit with code 1
- Can be configured to pass in CI/CD
- Should be used for best practices and recommendations

## Default Rules

If no config file is found, kubecheck uses these default rules:

1. **no-latest-image** (ERROR) - Disallow :latest tags
2. **no-root-containers** (ERROR) - Containers must not run as root
3. **require-resource-requests** (WARN) - CPU and memory requests required
4. **require-resource-limits** (WARN) - CPU and memory limits required

## Usage Examples

### Using Default Rules

```bash
# Uses built-in defaults
kubecheck deployment.yaml
```

### Using Custom Config

```bash
# Specify config file
kubecheck --config prod-rules.yaml deployment.yaml

# Use config from current directory (auto-detected)
kubecheck deployment.yaml
```

### Organization Setup

Create a shared config repository:

```bash
# Clone org config
git clone https://github.com/myorg/kubecheck-config
cd kubecheck-config

# Validate against org rules
kubecheck --config rules/production.yaml ../myapp/k8s/
```

### CI/CD Integration

```yaml
# .github/workflows/validate.yml
- name: Validate Kubernetes manifests
  run: |
    # Download org-wide rules
    curl -o kubecheck.yaml https://config.company.com/kubecheck.yaml
    
    # Validate
    kubecheck k8s/
```

## Best Practices

1. **Start with defaults** - Begin with built-in rules and customize gradually
2. **Version control** - Keep config files in git alongside manifests
3. **Team alignment** - Share configs across teams for consistency
4. **Document custom rules** - Add clear descriptions and help text
5. **Test changes** - Validate config changes in non-prod first
6. **Gradual rollout** - Start with WARN severity, move to ERROR after adoption

## Extending with New Conditions

To add new condition types, edit `cmd/kubecheck/rule-engine.go`:

```go
// Add to the checkCondition function
case "my_custom_condition":
    return checkMyCustomCondition(container)

// Implement the check function
func checkMyCustomCondition(c Container) bool {
    // Your validation logic
    return false
}
```

Then use it in your config:

```yaml
rules:
  - name: my-custom-rule
    conditions:
      - my_custom_condition
    message: "Custom validation failed"
```

## Troubleshooting

### Config file not found

```bash
# Check current directory
ls -la kubecheck.yaml

# Check home directory
ls -la ~/.kubecheck/

# Specify explicitly
kubecheck --config /path/to/config.yaml deployment.yaml
```

### Invalid config format

```bash
# Validate YAML syntax
yamllint kubecheck.yaml

# Check for typos in condition names
```

### Rules not triggering

- Verify condition names match exactly (see Available Conditions)
- Check that severity is either ERROR or WARN
- Ensure message includes {container} placeholder

---

For more information, see:
- [README.md](README.md) - General usage
- [CONTRIBUTING.md](CONTRIBUTING.md) - Adding new condition types
- [EXAMPLES.md](EXAMPLES.md) - Real-world configurations
