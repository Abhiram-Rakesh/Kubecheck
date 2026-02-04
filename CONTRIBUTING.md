# Contributing to kubecheck

Thank you for your interest in contributing! This guide will help you get started.

## Architecture Overview

kubecheck is built entirely in Go with a YAML-configurable rule system:

```
┌──────────────────────────────────────────┐
│  Go CLI (cmd/kubecheck)                  │
│  - Argument parsing & file discovery     │
│  - YAML config loading                   │
│  - Rule engine evaluation                │
│  - Output formatting                     │
└──────────────────────────────────────────┘
```

## Development Setup

### Prerequisites

- Go ≥ 1.21
- Helm (optional, for Helm chart testing)

### Local Development

```bash
# Clone the repository
git clone https://github.com/Abhiram-Rakesh/kubecheck.git
cd kubecheck

# Build locally
cd cmd/kubecheck
go build -o kubecheck

# Test locally
./kubecheck ../../examples/deployment.yaml
```

## Adding New Validation Rules

### Option 1: YAML Configuration (Recommended)

For most use cases, you can add rules via YAML configuration without touching code:

**1. Edit `kubecheck.yaml`:**

```yaml
rules:
  - name: my-custom-rule
    description: Description of what this checks
    severity: ERROR # or WARN
    type: security # or image, resources, etc.
    conditions:
      - existing_condition_type
    message: "Container '{container}' violates custom rule"
    help: "How to fix this issue"
```

**2. Test it:**

```bash
kubecheck --config kubecheck.yaml examples/deployment.yaml
```

### Option 2: New Condition Type (Code Change)

If you need a new condition type, you'll need to modify Go code:

**1. Add condition check function** in `cmd/kubecheck/rule-engine.go`:

```go
// checkHostNetwork checks if container uses host network
func checkHostNetwork(c Container) bool {
    // Your validation logic here
    return c.HostNetwork != nil && *c.HostNetwork
}
```

**2. Add to condition switch** in `checkCondition()`:

```go
func (re *RuleEngine) checkCondition(condition string, container Container) bool {
    // ... existing code ...

    switch conditionType {
    // ... existing cases ...
    case "uses_host_network":
        return checkHostNetwork(container)
    default:
        return false
    }
}
```

**3. Update Container struct if needed** (add new fields):

```go
type Container struct {
    Name            string
    Image           string
    Resources       *Resources
    SecurityContext *SecurityContext
    HostNetwork     *bool  // Add new field
}
```

**4. Update parser** in `parseContainers()` to extract new field:

```go
if hostNet, ok := containerMap["hostNetwork"].(bool); ok {
    container.HostNetwork = &hostNet
}
```

**5. Document the new condition** in `CONFIG.md`:

```markdown
- `uses_host_network` - Container uses host network
```

**6. Create test case** in `examples/`:

```yaml
# examples/host-network.yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-host-network
spec:
  hostNetwork: true
  containers:
    - name: app
      image: nginx:1.21
```

**7. Test:**

```bash
cd cmd/kubecheck
go build
./kubecheck ../../examples/host-network.yaml
```

## Testing

### Run Tests

```bash
cd cmd/kubecheck
go test -v ./...
```

### Manual Testing

```bash
# Test single file
kubecheck examples/deployment.yaml

# Test directory
kubecheck examples/

# Test Helm chart
kubecheck examples/helm-chart/

# Test stdin
cat examples/pod.yaml | kubecheck -

# Test custom config
kubecheck --config custom-rules.yaml examples/
```

## Code Style

### Go Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `go vet` before committing
- Keep functions focused and testable
- Add comments for exported functions

**Example:**

```go
// checkImageTag validates the container image tag
func checkImageTag(image, tag string) bool {
    if !strings.Contains(image, ":") {
        return tag == "latest"
    }
    parts := strings.Split(image, ":")
    return len(parts) == 2 && parts[1] == tag
}
```

### YAML Configuration Style

```yaml
rules:
  - name: kebab-case-name
    description: Clear, concise description
    severity: ERROR # or WARN
    type: category
    conditions:
      - condition_one
      - condition_two
    message: "Clear error message with {container} placeholder"
    help: "Actionable fix suggestion"
```

## Pull Request Process

1. **Fork the repository**

2. **Create a feature branch**

   ```bash
   git checkout -b feature/my-new-condition
   ```

3. **Make your changes**
   - Add condition to `rule-engine.go`
   - Update documentation in `CONFIG.md`
   - Add test cases to `examples/`

4. **Test thoroughly**

   ```bash
   go build
   ./kubecheck examples/
   ```

5. **Commit with clear messages**

   ```bash
   git commit -m "feat: add host network validation condition"
   ```

6. **Push and create PR**
   ```bash
   git push origin feature/my-new-condition
   ```

## Condition Design Guidelines

### Severity Levels

**ERROR** - Production-critical violations

- Security issues (root containers, privileged mode)
- Non-deterministic configurations (`:latest` tags)
- Critical misconfigurations

**WARN** - Best practice violations

- Missing resource limits
- Missing health probes
- Suboptimal configurations

### Condition Characteristics

Good conditions should be:

1. **Specific** - Check one thing clearly
2. **Deterministic** - Same input = same result
3. **Documented** - Clear description and help text
4. **Testable** - Easy to create test cases
5. **Composable** - Works well with other conditions

### Condition Naming

Use descriptive names in snake_case:

- ✅ `missing_cpu_requests`
- ✅ `image_tag_equals:latest`
- ✅ `run_as_user_zero`
- ❌ `bad_cpu`
- ❌ `check1`

### Message Templates

Use `{container}` placeholder for container name:

```yaml
message: "Container '{container}' uses latest tag"
```

Provide actionable help:

```yaml
help: "use a specific version like nginx:1.21.0"
```

## Debugging

### Debug Go Code

```bash
cd cmd/kubecheck
go run . -v ../../examples/deployment.yaml
```

### Debug Config Loading

```bash
# See which config is loaded
kubecheck -v deployment.yaml

# Test custom config
kubecheck --config test-config.yaml -v deployment.yaml
```

### Common Issues

1. **Condition not triggering**
   - Check condition name matches exactly
   - Verify field is being parsed correctly
   - Add debug logging: `fmt.Printf("Debug: %+v\n", container)`

2. **Config not loading**
   - Verify YAML syntax: `yamllint kubecheck.yaml`
   - Check file location
   - Use `-v` flag to see which config is loaded

3. **Parser not extracting field**
   - Check YAML structure matches your code
   - Verify type assertions are correct
   - Test with simple example first

## Resources

- [Kubernetes API Concepts](https://kubernetes.io/docs/reference/using-api/api-concepts/)
- [Production Best Practices](https://learnk8s.io/production-best-practices)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

## Community

- GitHub Issues: Bug reports and feature requests
- GitHub Discussions: Questions and ideas

---

**Thank you for contributing to kubecheck!**

### Quick Reference

**Add YAML rule:** Edit `kubecheck.yaml`  
**Add condition type:** Edit `rule-engine.go` → `checkCondition()`  
**Add field to Container:** Edit `Container` struct → Update parser  
**Test changes:** `go build && ./kubecheck examples/`  
**Format code:** `gofmt -w .`  
**Run tests:** `go test -v ./...`
