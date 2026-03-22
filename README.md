# Kubecheck

**Kubernetes static analysis CLI tool**

A CLI tool that validates Kubernetes YAML files against production best practices without connecting to a cluster. Designed for CI/CD pipelines, pre-commit hooks, and local developer validation.

### Examples

<table>
  <tr>
    <td align="center" width="50%">
      <b>Single file — with violations</b><br/>
      <img src="https://github.com/user-attachments/assets/786fff77-b97a-4f0e-93ed-4dc68457efc5"/>
    </td>
    <td align="center" width="50%">
      <b>Single file — all checks passed</b><br/>
      <img src="https://github.com/user-attachments/assets/e2e34113-d944-4a01-a3cd-6c72488b36b7"/>
    </td>
  </tr>
  <tr>
    <td align="center" width="50%">
      <b>Multi-document validation</b><br/>
      <img src="https://github.com/user-attachments/assets/b7fce799-5095-4069-ba44-b474f32ef38b"/>
    </td>
    <td align="center" width="50%">
      <b>Full directory scan</b><br/>
      <img src="https://github.com/user-attachments/assets/9fb39273-c5de-4998-a444-2e9270094f1c"/>
    </td>
  </tr>
  <tr>
    <td align="center" width="50%">
      <b>Stdin piping</b><br/>
      <img src="https://github.com/user-attachments/assets/5b7a23e3-6e0f-4493-b679-4a6a5dde3e32"/>
    </td>
    <td align="center" width="50%">
      <b>Helm chart + stdin</b><br/>
      <img src="https://github.com/user-attachments/assets/05e76ea6-ca38-4e4e-86d5-f5db7616b7f6"/>
    </td>
  </tr>
</table>

## How It Works

kubecheck parses YAML manifests, extracts container specs from supported resource types, and evaluates each container against a configurable set of rules. Violations are reported with severity levels and actionable help text.

**Supported resource types:** Deployment, StatefulSet, DaemonSet, ReplicaSet, Job, CronJob, Pod

## Features

### Input Support

- Single Kubernetes YAML files
- Directories (recursive scanning)
- Multi-document YAML files (`---` separated)
- Helm charts (via `helm template`)
- Stdin piping

### YAML-Configurable Rules

Organizations can define custom validation rules via YAML configuration:

```yaml
rules:
  - name: no-latest-image
    severity: ERROR
    conditions:
      - image_tag_equals:latest
    message: "Container '{container}' uses 'latest' image tag"
```

See [docs/CONFIG.md](docs/CONFIG.md) for complete documentation.

### Default Validation Rules

| Rule                          | Severity | Description                           |
| ----------------------------- | -------- | ------------------------------------- |
| `no-latest-image`             | ERROR    | Disallow `image: latest` tags         |
| `no-root-containers`          | ERROR    | Detect containers running as root     |
| `no-privileged-containers`    | ERROR    | Detect containers in privileged mode  |
| `require-resource-requests`   | WARN     | Require CPU/memory requests           |
| `require-resource-limits`     | WARN     | Require CPU/memory limits             |
| `require-liveness-probe`      | WARN     | Require a liveness probe              |
| `require-readiness-probe`     | WARN     | Require a readiness probe             |
| `require-image-pull-policy`   | WARN     | Require explicit imagePullPolicy      |

### Exit Codes

```
0 - OK    (all checks passed)
1 - WARN  (warnings found)
2 - ERROR (errors found)
```

The CLI exits with the highest severity found, making it CI-friendly.

## Installation

### Pre-built Binary (Recommended)

Download the latest binary for your platform from the [Releases page](https://github.com/Abhiram-Rakesh/Kubecheck/releases) and place it in your `PATH`.

### Build from Source

**Prerequisites:** Go ≥ 1.21, Helm (optional)

```bash
git clone https://github.com/Abhiram-Rakesh/Kubecheck.git
cd Kubecheck
chmod +x *.sh
./build.sh
```

This installs the `kubecheck` binary to `/usr/local/bin`.

### Uninstall

```bash
./uninstall.sh
```

## Usage

### Basic Usage

```bash
# Validate a single file
kubecheck deployment.yaml

# Validate a directory (recursive)
kubecheck k8s/

# Validate a Helm chart
kubecheck ./my-chart/

# Pipe from stdin
helm template ./my-chart | kubecheck -

# Verbose output (shows which config file was loaded)
kubecheck -v deployment.yaml

# Use custom config
kubecheck --config my-rules.yaml deployment.yaml
```

### Configuration

kubecheck looks for configuration files in this order:

1. `--config` flag (highest priority)
2. `./kubecheck.yaml` (current directory)
3. `./kubecheck.yml` (current directory)
4. `~/.kubecheck/config.yaml` (home directory)
5. Built-in defaults (if no config found)

Create a custom config:

```yaml
# kubecheck.yaml
rules:
  - name: require-company-registry
    description: All images must use company registry
    severity: ERROR
    type: image
    conditions:
      - image_not_from_registry:registry.company.com
    message: "Container '{container}' uses external registry"
    help: "use images from registry.company.com"
```

See [docs/CONFIG.md](docs/CONFIG.md) for the complete configuration guide.

### CI/CD Integration

**GitHub Actions:**

```yaml
- name: Validate Kubernetes manifests
  run: |
    git clone https://github.com/Abhiram-Rakesh/Kubecheck.git
    cd Kubecheck && ./build.sh && cd ..
    kubecheck k8s/
```

**GitLab CI:**

```yaml
validate-manifests:
  stage: test
  script:
    - git clone https://github.com/Abhiram-Rakesh/Kubecheck.git
    - cd Kubecheck && ./build.sh && cd ..
    - kubecheck k8s/
```

## Documentation

- [docs/CONFIG.md](docs/CONFIG.md) - Configuration guide
- [docs/QUICKSTART.md](docs/QUICKSTART.md) - Get started in 5 minutes
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) - System design
- [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) - How to contribute
- [docs/EXAMPLES.md](docs/EXAMPLES.md) - Real-world usage examples

## Acknowledgments

Built with best practices from:

- [Kubernetes Production Best Practices](https://learnk8s.io/production-best-practices)
- [NSA Kubernetes Hardening Guide](https://www.nsa.gov/Press-Room/News-Highlights/Article/Article/2716980/)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)

---
