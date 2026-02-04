# Usage Examples

Comprehensive examples of using `kubecheck` in various scenarios.

## Table of Contents

- [Basic Validation](#basic-validation)
- [Configuration Examples](#configuration-examples)
- [CI/CD Integration](#cicd-integration)
- [Advanced Workflows](#advanced-workflows)

---

## Basic Validation

### Single File Validation

```bash
$ kubecheck deployment.yaml

  ● File: deployment.yaml
  ┌─ Deployment: nginx-deployment ──────────────────────┐
  │  ✖  Security Violation
  │     Container 'nginx' uses 'latest' image tag
  │     ▲─── use a specific version or digest
  │
  │  ⚠  Resource Hygiene
  │     Container 'nginx' missing resource requests
  └──────────────────────────────────── [ 1 errors | 1 warns ]

  Summary ➔ 1 file checked. 2 violations found.
```

### Directory Validation

```bash
$ kubecheck k8s/

  🔍 Scanning directory: ./k8s/
  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  ✔  k8s/service.yaml ..................... PASSED
  ⚠  k8s/deployment.yaml .................. 1 WARN
     └─ [api] Container 'api' missing resource limits
  ✖  k8s/cronjob.yaml ..................... 1 ERR
     └─ [backup] Container 'backup' uses 'latest' tag

  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Summary ➔ 3 files checked
  Result  ➔ 1 OK  |  1 Warning  |  1 Error
  Status  ➔ FAILED Exit code: 2
```

### Helm Chart Validation

```bash
$ kubecheck ./charts/myapp/

  📦 Rendering Helm chart: ./charts/myapp/

  ✔  templates/service.yaml ............... PASSED
  ✖  templates/deployment.yaml ............ 1 ERR
     └─ [myapp] Container 'myapp' uses 'latest' tag
```

## Configuration Examples

### Minimal Configuration

```yaml
# kubecheck.yaml
rules:
  - name: no-latest
    severity: ERROR
    type: image
    conditions:
      - image_tag_equals:latest
    message: "Container '{container}' uses latest tag"
```

### Production Configuration

```yaml
# kubecheck.yaml
rules:
  # Security: No latest tags
  - name: no-latest-image
    description: Prevent non-deterministic deployments
    severity: ERROR
    type: image
    conditions:
      - image_tag_equals:latest
      - image_tag_missing
    message: "Container '{container}' uses 'latest' image tag"
    help: "use specific version like nginx:1.21.0"

  # Security: Non-root containers
  - name: no-root-containers
    description: Containers must not run as root
    severity: ERROR
    type: security
    conditions:
      - missing_security_context
      - run_as_non_root_false
      - run_as_user_zero
    message: "Container '{container}' running as root"
    help: "set runAsNonRoot: true and runAsUser: 1000"

  # Resources: CPU requests
  - name: require-cpu-requests
    description: CPU requests for scheduling
    severity: WARN
    type: resources
    conditions:
      - missing_cpu_requests
    message: "Container '{container}' missing CPU requests"
    help: "set resources.requests.cpu (e.g., 100m)"

  # Resources: Memory requests
  - name: require-memory-requests
    description: Memory requests for scheduling
    severity: WARN
    type: resources
    conditions:
      - missing_memory_requests
    message: "Container '{container}' missing memory requests"
    help: "set resources.requests.memory (e.g., 128Mi)"

  # Resources: CPU limits
  - name: require-cpu-limits
    description: CPU limits to prevent noisy neighbors
    severity: WARN
    type: resources
    conditions:
      - missing_cpu_limits
    message: "Container '{container}' missing CPU limits"
    help: "set resources.limits.cpu (e.g., 500m)"

  # Resources: Memory limits
  - name: require-memory-limits
    description: Memory limits to prevent OOM
    severity: WARN
    type: resources
    conditions:
      - missing_memory_limits
    message: "Container '{container}' missing memory limits"
    help: "set resources.limits.memory (e.g., 512Mi)"
```

### Team-Specific Configs

**DevOps Team (Strict):**
```yaml
# .kubecheck/devops-rules.yaml
rules:
  - name: no-latest-image
    severity: ERROR
    conditions:
      - image_tag_equals:latest
    message: "Latest tags not allowed in production"

  - name: require-all-resources
    severity: ERROR
    conditions:
      - missing_cpu_requests
      - missing_memory_requests
      - missing_cpu_limits
      - missing_memory_limits
    message: "All resource specs required"
```

**Development Team (Relaxed):**
```yaml
# .kubecheck/dev-rules.yaml
rules:
  - name: no-latest-image
    severity: WARN
    conditions:
      - image_tag_equals:latest
    message: "Consider using specific tags"

  - name: suggest-resources
    severity: WARN
    conditions:
      - missing_cpu_requests
      - missing_memory_requests
    message: "Resource requests recommended"
```

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/k8s-validation.yml
name: Validate Kubernetes Manifests

on:
  pull_request:
    paths:
      - 'k8s/**'
      - 'charts/**'
  push:
    branches:
      - main

jobs:
  validate:
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install kubecheck
        run: |
          git clone https://github.com/your-org/kubecheck
          cd kubecheck
          ./build.sh
      
      - name: Download org config
        run: |
          curl -o kubecheck.yaml \
            https://config.company.com/kubecheck.yaml
      
      - name: Validate manifests
        run: kubecheck k8s/
      
      - name: Validate Helm charts
        run: |
          for chart in charts/*; do
            kubecheck "$chart"
          done
```

### GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - validate

validate:k8s:
  stage: validate
  image: golang:1.21
  before_script:
    - git clone https://github.com/your-org/kubecheck
    - cd kubecheck && ./build.sh && cd ..
    - curl -o kubecheck.yaml https://config.company.com/kubecheck.yaml
  script:
    - kubecheck k8s/
  only:
    changes:
      - k8s/**
      - charts/**
```

### Jenkins Pipeline

```groovy
// Jenkinsfile
pipeline {
    agent any
    
    stages {
        stage('Setup') {
            steps {
                sh '''
                    git clone https://github.com/your-org/kubecheck
                    cd kubecheck && ./build.sh
                '''
            }
        }
        
        stage('Validate') {
            steps {
                sh '''
                    curl -o kubecheck.yaml https://config.company.com/kubecheck.yaml
                    kubecheck k8s/
                '''
            }
        }
    }
}
```

### ArgoCD Pre-Sync Hook

```yaml
# argo-validation-hook.yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: kubecheck-validator
  annotations:
    argocd.argoproj.io/hook: PreSync
    argocd.argoproj.io/hook-delete-policy: BeforeHookCreation
spec:
  template:
    spec:
      containers:
      - name: validate
        image: golang:1.21
        command:
          - sh
          - -c
          - |
            git clone https://github.com/your-org/kubecheck
            cd kubecheck && ./build.sh
            kubecheck /manifests
        volumeMounts:
        - name: manifests
          mountPath: /manifests
      restartPolicy: Never
      volumes:
      - name: manifests
        configMap:
          name: k8s-manifests
```

## Advanced Workflows

### Pre-Commit Hook

```bash
# .git/hooks/pre-commit
#!/bin/bash

echo "🔍 Validating Kubernetes manifests..."

# Find changed YAML files
changed_files=$(git diff --cached --name-only --diff-filter=ACM | grep -E "^k8s/.*\.ya?ml$")

if [ -z "$changed_files" ]; then
    echo "No K8s manifests changed"
    exit 0
fi

# Validate each file
for file in $changed_files; do
    echo "Checking $file..."
    if ! kubecheck "$file"; then
        echo "❌ Validation failed for $file"
        echo "Fix the issues before committing"
        exit 1
    fi
done

echo "✅ All manifests valid!"
exit 0
```

### Makefile Integration

```makefile
# Makefile
.PHONY: validate validate-strict validate-charts

validate:
	@echo "Validating with default rules..."
	@kubecheck k8s/

validate-strict:
	@echo "Validating with strict rules..."
	@kubecheck --config .kubecheck/strict-rules.yaml k8s/

validate-charts:
	@echo "Validating Helm charts..."
	@for chart in charts/*; do \
		echo "Checking $$chart..."; \
		kubecheck "$$chart" || exit 1; \
	done

deploy: validate
	kubectl apply -f k8s/
```

### Organization Config Repository

```bash
# Setup org-wide config
git clone https://github.com/myorg/kubecheck-config
cd myapp

# Use org config
kubecheck --config ../kubecheck-config/production.yaml k8s/

# Or link it
ln -s ../kubecheck-config/production.yaml kubecheck.yaml
kubecheck k8s/
```

### Custom Wrapper Script

```bash
#!/bin/bash
# kube-validate.sh - Wrapper for kubecheck with org defaults

ORG_CONFIG="https://config.company.com/kubecheck.yaml"
LOCAL_CONFIG=".kubecheck.yaml"

# Download org config if not present
if [ ! -f "$LOCAL_CONFIG" ]; then
    echo "Downloading org config..."
    curl -s -o "$LOCAL_CONFIG" "$ORG_CONFIG"
fi

# Run validation
kubecheck --config "$LOCAL_CONFIG" "$@"
```

Usage:
```bash
./kube-validate.sh k8s/
./kube-validate.sh deployment.yaml
./kube-validate.sh --help
```

### Multi-Environment Validation

```bash
#!/bin/bash
# validate-all-envs.sh

ENVS="dev staging production"

for env in $ENVS; do
    echo "Validating $env environment..."
    if ! kubecheck --config ".kubecheck/${env}-rules.yaml" "k8s/${env}/"; then
        echo "❌ $env validation failed"
        exit 1
    fi
done

echo "✅ All environments validated!"
```

---

For more information, see:
- [README.md](README.md) - Overview and getting started
- [CONFIG.md](CONFIG.md) - Configuration reference
- [CONTRIBUTING.md](CONTRIBUTING.md) - Adding new rules
