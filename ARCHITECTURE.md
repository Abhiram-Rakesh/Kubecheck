# Architecture Documentation

## System Overview

`kubecheck` is a production-grade static analysis tool for Kubernetes manifests, built entirely in Go with a YAML-configurable rule system.

## Design Principles

### 1. Simplicity First

**Pure Go Implementation**
- No complex language dependencies (removed Haskell)
- Easy to install and distribute
- Fast compilation and execution
- Cross-platform compatibility

**YAML-Based Configuration**
- Declarative rule definitions
- No code changes needed for new rules
- Organization-specific customization
- Version-controlled policy

### 2. Data Flow

```
User Input
    ↓
Go CLI (main.go)
    ↓
Config Loader (config.go) → Load rules from YAML or defaults
    ↓
File Discovery (parser.go, helm.go)
    ↓
YAML Parsing → K8sResource structs
    ↓
Rule Engine (rule-engine.go)
    ↓
Condition Evaluation → Match against containers
    ↓
Violation List
    ↓
Reporter (reporter.go)
    ↓
Formatted Output (with colors & box-drawing)
    ↓
Exit Code (0, 1, 2)
```

### 3. Why Pure Go?

**Advantages**
- ✅ Single binary distribution
- ✅ No runtime dependencies
- ✅ Easy CI/CD integration
- ✅ Fast startup time
- ✅ Simple installation (just Go)
- ✅ Cross-platform builds

**Previous Haskell Approach**
- ❌ Required GHC + Cabal installation
- ❌ Complex build process
- ❌ Difficult for non-Haskell developers
- ❌ Slower builds

## Component Details

### Go Components

#### `main.go`
- Entry point for CLI
- Parses flags: `-v` for verbose, `--config` for custom config
- Determines input type (file, directory, Helm chart, stdin)
- Loads rule configuration
- Orchestrates validation pipeline
- Manages exit codes based on severity

#### `config.go`
- Loads YAML configuration files
- Provides default built-in rules
- Searches multiple config locations
- Validates config structure

#### `rule-engine.go`
- Evaluates YAML-defined rules
- Extracts containers from resources
- Checks conditions against containers
- Generates violations with messages
- Supports extensible condition system

#### `parser.go`
- Reads YAML files
- Handles multi-document YAML (--- separators)
- Unmarshals into Go structs
- Recursively scans directories for .yaml/.yml files

#### `helm.go`
- Detects Helm charts (looks for Chart.yaml)
- Executes `helm template` to render manifests
- Writes rendered output to temporary files
- Returns file paths for validation

#### `reporter.go`
- Formats validation results with colors and box-drawing
- Tracks statistics (OK, WARN, ERROR counts)
- Provides two output modes:
  - **Single file**: Detailed boxes with inline help
  - **Directory**: Compact tree format
- Prints styled summary

## Configuration System

### Config File Structure

```yaml
rules:
  - name: string           # Unique identifier
    description: string    # Human-readable description
    severity: ERROR|WARN   # Violation severity
    type: string          # Category (image, resources, security)
    conditions: []string  # List of conditions to check
    message: string       # Error message (supports {container} placeholder)
    help: string          # Optional remediation guidance
```

### Condition Evaluation

Conditions are strings in the format `condition_type:value`.

**Evaluation Flow:**
1. Parse condition string
2. Extract condition type and optional value
3. Match against switch statement
4. Execute condition-specific check function
5. Return boolean result

**Example:**
```yaml
conditions:
  - image_tag_equals:latest  # condition_type:value
  - missing_cpu_requests     # condition_type (no value)
```

### Extensibility

Adding new conditions:

1. **Add condition check function** in `rule-engine.go`:
   ```go
   func checkNewCondition(c Container) bool {
       // Your validation logic
       return false
   }
   ```

2. **Add case to switch statement**:
   ```go
   case "new_condition":
       return checkNewCondition(container)
   ```

3. **Use in configuration**:
   ```yaml
   conditions:
     - new_condition
   ```

## Data Structures

### K8sResource

```go
type K8sResource struct {
    APIVersion string
    Kind       string
    Metadata   map[string]interface{}
    Spec       map[string]interface{}
}
```

Flexible structure using `map[string]interface{}` to handle various Kubernetes resource types.

### Container

```go
type Container struct {
    Name            string
    Image           string
    Resources       *Resources
    SecurityContext *SecurityContext
}
```

Parsed from resource spec for validation.

### Violation

```go
type Violation struct {
    Severity string  // ERROR or WARN
    Message  string  // Error message
    Rule     string  // Rule identifier
}
```

## Build System

### Development Build
```bash
cd cmd/kubecheck
go build
```

### Production Build
```bash
./build.sh
```

This script:
1. Checks prerequisites (Go ≥ 1.21)
2. Builds Go CLI
3. Installs to `/usr/local/bin/kubecheck`

### Uninstallation
```bash
./uninstall.sh
```

Removes installed binary.

## Testing Strategy

### Unit Tests
- Test YAML parsing
- Test file discovery
- Test condition evaluation
- Test config loading

### Integration Tests
- End-to-end validation of example manifests
- Helm chart rendering and validation
- Multi-document YAML handling
- Stdin piping

### CI/CD Integration
```yaml
# Example GitHub Actions
- name: Validate manifests
  run: kubecheck k8s/
  # Fails on exit code 2 (ERROR)
```

## Performance Considerations

### File Processing
- Recursive directory scanning is efficient
- Multi-document YAML is streamed
- Helm rendering uses temp files

### Rule Evaluation
- O(n*m) complexity where n=containers, m=rules
- Conditions short-circuit on first match
- Minimal memory overhead

## Security Considerations

1. **No Cluster Access**
   - Tool never connects to Kubernetes API
   - No credentials required
   - Safe to run in CI/CD

2. **Sandboxed Helm Execution**
   - Only runs `helm template` (read-only)
   - Uses temporary directories
   - No network access required

3. **Input Validation**
   - YAML parsing is safe (no code execution)
   - Bounded resource usage
   - Config validation prevents injection

## Future Enhancements

### Planned Features
- Custom condition plugins via Go plugins
- Regex-based conditions
- Resource-type specific rules
- Remote config loading (HTTP/S3)
- JSON/SARIF output formats
- Parallel validation

### Non-Goals
- Runtime validation (use OPA/Kyverno)
- Cluster state analysis (use kubectl/k9s)
- Remediation/auto-fix (keep tool focused)
- Web UI (CLI-first philosophy)

---

## Comparison: Old vs New Architecture

### Old (Haskell + Go)
```
Go CLI → JSON → Haskell Rule Engine → JSON → Go Reporter
```

**Issues:**
- Complex installation (Go + GHC + Cabal)
- Slow builds
- Hard to extend for non-Haskell devs
- IPC overhead

### New (Pure Go + YAML)
```
Go CLI → YAML Config → Go Rule Engine → Go Reporter
```

**Benefits:**
- Simple installation (just Go)
- Fast builds
- Easy to extend (just edit YAML)
- No IPC overhead
- Organization-friendly

---

For more details, see:
- [README.md](README.md) - Overview
- [CONFIG.md](CONFIG.md) - Configuration system
- [CONTRIBUTING.md](CONTRIBUTING.md) - How to contribute
