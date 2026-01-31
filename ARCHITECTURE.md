# Architecture Documentation

## System Overview

`kubecheck` is a production-grade static analysis tool for Kubernetes manifests, designed with a clear separation of concerns between orchestration (Go) and validation logic (Haskell).

## Design Principles

### 1. Separation of Concerns

**Go Layer (Orchestration)**
- Handles system integration (file I/O, process execution)
- Parses command-line arguments
- Discovers and reads YAML files
- Integrates with Helm for chart rendering
- Formats output for human consumption
- Manages process exit codes

**Haskell Layer (Validation)**
- Defines validation rules in a pure, functional style
- Provides strong type safety for Kubernetes resources
- Ensures deterministic validation logic
- Maintains rule extensibility through composition

### 2. Data Flow

```
User Input
    ↓
Go CLI (main.go)
    ↓
File Discovery (parser.go, helm.go)
    ↓
YAML Parsing → K8sResource structs
    ↓
JSON Serialization
    ↓
Haskell Rule Engine (stdin)
    ↓
Rule Application (Rules.hs)
    ↓
Violation List (JSON stdout)
    ↓
Go Reporter (reporter.go)
    ↓
Formatted Output
    ↓
Exit Code (0, 1, 2)
```

### 3. Why Haskell for Rules?

**Type Safety**
- Compile-time guarantees prevent runtime errors
- Impossible to apply wrong rule to wrong resource type
- Type-driven development catches bugs early

**Pure Functions**
- No side effects in validation logic
- Deterministic behavior (same input → same output)
- Easy to test and reason about

**Declarative Style**
- Rules read like specifications
- Easy to audit and review
- Natural fit for policy-as-code

**Composability**
- Rules are just functions that return violations
- New rules compose naturally with existing ones
- No complex frameworks or inheritance hierarchies

## Component Details

### Go Components

#### `main.go`
- Entry point for CLI
- Parses flags: `-v` for verbose
- Determines input type (file, directory, Helm chart, stdin)
- Orchestrates validation pipeline
- Manages exit codes based on severity

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
- Formats validation results
- Tracks statistics (OK, WARN, ERROR counts)
- Provides verbose and non-verbose output modes
- Prints summary at the end

#### `engine.go`
- Calls the Haskell rule engine binary
- Pipes JSON via stdin
- Reads JSON violations from stdout
- Handles error cases

### Haskell Components

#### `Types.hs`
- Defines core data types
- `K8sResource`: Top-level Kubernetes resource
- `Container`: Container specification
- `SecurityContext`, `Resources`: Nested configurations
- `Violation`: Rule violation with severity and message
- Aeson instances for JSON serialization

#### `Rules.hs`
- Contains individual validation rules
- Each rule is a pure function: `Container -> [Violation]`
- Currently implements:
  - `checkNoLatestImage`: Disallow `:latest` tags
  - `checkRequireResources`: Require resource requests/limits
  - `checkNoRootContainers`: Enforce non-root containers

#### `Validator.hs`
- Applies all rules to a resource
- Extracts containers from various resource types
- Handles Deployments, Pods, StatefulSets, etc.
- Composes rule results into final violation list

#### `app/Main.hs`
- CLI wrapper for the rule engine
- Reads JSON from stdin
- Decodes into `K8sResource`
- Validates and encodes violations to JSON
- Writes to stdout

## Inter-Process Communication

### JSON Schema

**Input (Go → Haskell):**
```json
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "nginx",
    "namespace": "default"
  },
  "spec": {
    "template": {
      "spec": {
        "containers": [
          {
            "name": "nginx",
            "image": "nginx:latest",
            "securityContext": {
              "runAsNonRoot": false
            }
          }
        ]
      }
    }
  }
}
```

**Output (Haskell → Go):**
```json
[
  {
    "severity": "ERROR",
    "message": "Container 'nginx' uses 'latest' image tag",
    "rule": "no-latest-image"
  },
  {
    "severity": "ERROR",
    "message": "Container 'nginx' has runAsNonRoot set to false",
    "rule": "no-root-containers"
  }
]
```

## Extension Points

### Adding New Rules

1. **Define in `Rules.hs`:**
   ```haskell
   checkNewRule :: Container -> [Violation]
   checkNewRule container = 
       if condition
           then [Violation severity message "rule-name"]
           else []
   ```

2. **Apply in `Validator.hs`:**
   ```haskell
   validateContainer container =
       concat [ ...
              , checkNewRule container
              ]
   ```

3. **Update types if needed** in `Types.hs`

### Adding New Resource Types

1. **Extend `Spec` type** in `Types.hs`
2. **Update `extractContainers`** in `Validator.hs`
3. Test with example manifests

## Build System

### Development Build
```bash
# Go CLI
cd cmd/kubecheck && go build

# Haskell engine
cd haskell && cabal build
```

### Production Build
```bash
./build.sh
```

This script:
1. Checks prerequisites (Go, GHC, Cabal)
2. Builds Haskell rule engine
3. Builds Go CLI
4. Installs both to system paths:
   - `/usr/local/bin/kubecheck`
   - `/usr/local/lib/kubecheck/kubecheck-rules`

### Uninstallation
```bash
./uninstall.sh
```

Removes all installed artifacts.

## Testing Strategy

### Unit Tests
- **Go**: Test YAML parsing, file discovery
- **Haskell**: Test individual rules, type conversions

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
- Recursive directory scanning is efficient for typical K8s repos
- Multi-document YAML is streamed to avoid memory issues
- Helm rendering uses temp files to avoid memory pressure

### Rule Engine
- Pure functions enable parallelization (future enhancement)
- Type safety eliminates runtime checks
- Minimal memory overhead for violation lists

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
   - JSON schema validation in Haskell
   - Bounded resource usage

## Future Enhancements

### Planned Features
- Custom rule plugins
- SARIF output format for IDE integration
- JSON/YAML output modes
- Rule severity overrides via config file
- Parallel validation for large repositories
- Support for Kustomize overlays

### Non-Goals
- Runtime validation (use OPA/Kyverno)
- Cluster state analysis (use kubectl/k9s)
- Remediation/auto-fix (keep tool focused)
- Web UI (CLI-first philosophy)

---

## References

- [Kubernetes API Concepts](https://kubernetes.io/docs/reference/using-api/api-concepts/)
- [Production Best Practices](https://learnk8s.io/production-best-practices)
- [Haskell Aeson Library](https://hackage.haskell.org/package/aeson)
- [Go YAML Library](https://github.com/go-yaml/yaml)
