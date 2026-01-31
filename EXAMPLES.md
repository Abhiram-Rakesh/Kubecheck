# Usage Examples

Comprehensive examples of using `kubecheck` in various scenarios.

## Table of Contents

- [Basic Validation](#basic-validation)
- [CI/CD Integration](#cicd-integration)
- [Pre-Commit Hooks](#pre-commit-hooks)
- [Helm Charts](#helm-charts)
- [Advanced Workflows](#advanced-workflows)

## Basic Validation

### Single File Validation

```bash
$ kubecheck deployment.yaml

ERROR: deployment.yaml
  Resource: Deployment/nginx-deployment
  [ERROR] Container 'nginx' uses 'latest' image tag
  [WARN] Container 'nginx' missing resource requests/limits
  [ERROR] Container 'nginx' missing securityContext (should set runAsNonRoot: true)

─────────────────────────────────────
Summary: 1 files checked
  ✗ 1 with errors
Total violations: 3
─────────────────────────────────────
Exit code: 2
```

### Directory Validation

```bash
$ kubecheck k8s/

OK: k8s/service.yaml

WARN: k8s/deployment.yaml
  Resource: Deployment/api-server
  [WARN] Container 'api' missing resource limits

ERROR: k8s/cronjob.yaml
  Resource: CronJob/backup
  [ERROR] Container 'backup' uses 'latest' image tag

─────────────────────────────────────
Summary: 3 files checked
  ✓ 1 OK
  ⚠ 1 with warnings
  ✗ 1 with errors
Total violations: 2
─────────────────────────────────────
Exit code: 2
```

### Verbose Mode

```bash
$ kubecheck -v k8s/

OK: k8s/service.yaml
  Resource: Service/nginx-service

OK: k8s/deployment-good.yaml
  Resource: Deployment/secure-app

WARN: k8s/deployment.yaml
  Resource: Deployment/nginx
  [WARN] Container 'nginx' missing memory limits

─────────────────────────────────────
Summary: 3 files checked
  ✓ 2 OK
  ⚠ 1 with warnings
Total violations: 1
─────────────────────────────────────
```

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/k8s-validation.yml
name: Validate Kubernetes Manifests

on:
  pull_request:
    paths:
      - "k8s/**"
      - "charts/**"
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
          go-version: "1.21"

      - name: Setup Haskell
        uses: haskell/actions/setup@v2
        with:
          ghc-version: "9.6.6"
          cabal-version: "3.10"

      - name: Build kubecheck
        run: |
          git clone https://github.com/your-org/kubecheck
          cd kubecheck
          chmod +x build.sh
          ./build.sh

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
    - apt-get update
    - apt-get install -y ghc cabal-install
    - git clone https://github.com/your-org/kubecheck
    - cd kubecheck && ./build.sh && cd ..
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
        stage('Validate K8s Manifests') {
            steps {
                sh '''
                    git clone https://github.com/your-org/kubecheck
                    cd kubecheck
                    ./build.sh
                    cd ..
                    kubecheck k8s/
                '''
            }
        }
    }

    post {
        failure {
            echo 'Kubernetes manifests validation failed!'
        }
    }
}
```

### CircleCI

```yaml
# .circleci/config.yml
version: 2.1

jobs:
  validate:
    docker:
      - image: golang:1.21
    steps:
      - checkout
      - run:
          name: Install dependencies
          command: |
            apt-get update
            apt-get install -y ghc cabal-install
      - run:
          name: Build kubecheck
          command: |
            git clone https://github.com/your-org/kubecheck
            cd kubecheck
            ./build.sh
      - run:
          name: Validate manifests
          command: kubecheck k8s/

workflows:
  version: 2
  validate-manifests:
    jobs:
      - validate
```

## Pre-Commit Hooks

### Git Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash

echo "🔍 Validating Kubernetes manifests..."

# Find all YAML files in k8s/
changed_files=$(git diff --cached --name-only --diff-filter=ACM | grep -E "^k8s/.*\.ya?ml$")

if [ -z "$changed_files" ]; then
    echo "No Kubernetes manifests changed, skipping validation."
    exit 0
fi

# Validate changed files
for file in $changed_files; do
    echo "Checking $file..."
    if ! kubecheck "$file"; then
        echo "❌ Validation failed for $file"
        echo "Fix the issues above before committing."
        exit 1
    fi
done

echo "✅ All Kubernetes manifests are valid!"
exit 0
```

Make it executable:

```bash
chmod +x .git/hooks/pre-commit
```

### pre-commit Framework

Create `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: local
    hooks:
      - id: kubecheck
        name: Validate Kubernetes manifests
        entry: kubecheck
        language: system
        files: \.(yaml|yml)$
        pass_filenames: true
```

Install:

```bash
pip install pre-commit
pre-commit install
```

## Helm Charts

### Basic Helm Chart Validation

```bash
$ kubecheck ./charts/myapp/

Rendering Helm chart: ./charts/myapp/

ERROR: charts/myapp/templates/deployment.yaml
  Resource: Deployment/myapp
  [ERROR] Container 'myapp' uses 'latest' image tag

─────────────────────────────────────
Summary: 3 files checked
  ✓ 2 OK
  ✗ 1 with errors
Total violations: 1
─────────────────────────────────────
```

### Test Different Values Files

```bash
# Production values
helm template ./charts/myapp -f values-prod.yaml | kubecheck -

# Staging values
helm template ./charts/myapp -f values-staging.yaml | kubecheck -

# Development values
helm template ./charts/myapp -f values-dev.yaml | kubecheck -
```

### Validate with Overrides

```bash
helm template ./charts/myapp \
  --set image.tag=v1.2.3 \
  --set resources.limits.memory=1Gi \
  | kubecheck -
```

### Multiple Charts

```bash
#!/bin/bash
# validate-all-charts.sh

for chart in charts/*/; do
    echo "Validating $chart..."
    if ! kubecheck "$chart"; then
        echo "❌ Validation failed for $chart"
        exit 1
    fi
done

echo "✅ All charts validated successfully!"
```

## Advanced Workflows

### Integration with kubectl

```bash
# Validate before applying
kubectl get deployment nginx -o yaml | kubecheck - && kubectl apply -f deployment.yaml

# Validate current cluster resources
kubectl get deployments -o yaml | kubecheck -
```

### Watch Mode with entr

```bash
# Auto-validate on file changes
ls k8s/*.yaml | entr kubecheck k8s/
```

### Validate Kustomize Output

```bash
kustomize build overlays/production | kubecheck -
```

### Makefile Integration

```makefile
# Makefile
.PHONY: validate
validate:
	@echo "Validating Kubernetes manifests..."
	@kubecheck k8s/

.PHONY: validate-charts
validate-charts:
	@echo "Validating Helm charts..."
	@for chart in charts/*; do \
		echo "Checking $$chart..."; \
		kubecheck "$$chart" || exit 1; \
	done

.PHONY: deploy
deploy: validate
	kubectl apply -f k8s/
```

Usage:

```bash
make validate
make validate-charts
make deploy
```

### Parallel Validation (Large Repos)

```bash
#!/bin/bash
# parallel-validate.sh

find k8s/ -name "*.yaml" | parallel -j 4 kubecheck {}
```

### Custom Exit Code Handling

```bash
#!/bin/bash

kubecheck k8s/
exit_code=$?

case $exit_code in
    0)
        echo "✅ All checks passed"
        ;;
    1)
        echo "⚠️  Warnings found, proceeding anyway"
        exit 0  # Don't fail CI on warnings
        ;;
    2)
        echo "❌ Errors found, failing build"
        exit 1
        ;;
esac
```

### Slack Notifications

```bash
#!/bin/bash
# validate-and-notify.sh

if kubecheck k8s/; then
    curl -X POST -H 'Content-type: application/json' \
        --data '{"text":"✅ K8s manifests validation passed"}' \
        "$SLACK_WEBHOOK_URL"
else
    curl -X POST -H 'Content-type: application/json' \
        --data '{"text":"❌ K8s manifests validation failed"}' \
        "$SLACK_WEBHOOK_URL"
    exit 1
fi
```

### ArgoCD Integration

```yaml
# argocd-hook.yaml
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
          image: your-registry/kubecheck:latest
          command: ["kubecheck", "/manifests"]
          volumeMounts:
            - name: manifests
              mountPath: /manifests
      restartPolicy: Never
      volumes:
        - name: manifests
          configMap:
            name: k8s-manifests
```

### Terraform Integration

```hcl
# main.tf
resource "null_resource" "validate_k8s" {
  provisioner "local-exec" {
    command = "kubecheck k8s/"
  }

  triggers = {
    manifests = filemd5("k8s/")
  }
}
```

### VS Code Tasks

`.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Validate K8s Manifests",
      "type": "shell",
      "command": "kubecheck",
      "args": ["k8s/"],
      "presentation": {
        "reveal": "always",
        "panel": "new"
      },
      "problemMatcher": []
    }
  ]
}
```

Press `Ctrl+Shift+B` (or `Cmd+Shift+B` on Mac) to run.

---

## Tips and Tricks

### Ignore Specific Files

```bash
# Validate everything except test files
find k8s/ -name "*.yaml" ! -name "*-test.yaml" -exec kubecheck {} +
```

### Exit Codes in Scripts

```bash
#!/bin/bash
set -e  # Exit on error

kubecheck deployment.yaml    # Exit code 2 will stop the script
kubectl apply -f deployment.yaml
```

### Color Output in CI

Most CI systems support ANSI colors. If not:

```bash
kubecheck k8s/ 2>&1 | cat  # Strip colors
```

### Performance Tips

For large repos (1000+ files):

```bash
# Use parallel validation
find k8s/ -name "*.yaml" | xargs -P 8 -I {} kubecheck {}
```

---

**For more examples, see the `examples/` directory in the repository.**
