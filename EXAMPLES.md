# Usage Examples

Comprehensive examples of using `kubecheck` in various scenarios.

## Table of Contents

- [Basic Validation](#basic-validation)
- [CI/CD Integration](#cicd-integration)
- [Helm Charts](#helm-charts)
- [Advanced Workflows](#advanced-workflows)

---

## Basic Validation

### Single File Validation

When validating a single file, `kubecheck` provides a detailed "Compiler-style" breakdown with line pointers and remediation tips.

```bash
$ kubecheck deployment.yaml

  ● File: deployment.yaml
  ┌─ Deployment: nginx-deployment ──────────────────────────────────────┐
  │                                                                     │
  │  ✖  Security Violation                                              │
  │     Container 'nginx' uses 'latest' image tag                       │
  │                                                                     │
  │     12 │   containers:                                              │
  │     13 │     - name: nginx                                          │
  │  >  14 │       image: nginx:latest                                  │
  │        │       ▲─── use a specific version or digest                │
  │                                                                     │
  │  ⚠  Resource Hygiene                                                │
  │     Container 'nginx' missing resource requests/limits              │
  │                                                                     │
  │  ✖  Security Violation                                              │
  │     Container 'nginx' missing securityContext                       │
  │     help: set 'runAsNonRoot: true' to improve pod security          │
  │                                                                     │
  └─────────────────────────────────────────────── [ 2 errors | 1 warn ]

  Summary ➔ 1 file checked. 3 violations found.

 $ kubecheck k8s/

  🔍 Scanning directory: ./k8s/
  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  ✔  k8s/service.yaml ......................................... PASSED

  ⚠  k8s/deployment.yaml ...................................... 1 WARN
     └─ [api-server] Container 'api' missing resource limits

  ✖  k8s/cronjob.yaml ......................................... 1 ERR
     └─ [backup] Container 'backup' uses 'latest' image tag
        > 08 │       image: backup-tool:latest
             │       ▲─── tag 'latest' is non-deterministic

  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Summary ➔ 3 files checked
  Result  ➔ 1 OK  |  1 Warning  |  1 Error
  Status  ➔ FAILED Exit code: 2

$ kubecheck -v k8s/

  ✔  k8s/service.yaml
     └─ Resource: Service/nginx-service

  ✔  k8s/deployment-good.yaml
     └─ Resource: Deployment/secure-app

  ⚠  k8s/deployment.yaml
     └─ [nginx] Container 'nginx' missing memory limits

  Summary ➔ 3 files checked. 1 violation found.

$ kubecheck ./charts/myapp/

  📦 Rendering Helm chart: ./charts/myapp/

  📂 templates/deployment.yaml
  ┌─ Deployment: myapp ────────────────────────────────────────────────┐
  │  ✖  Container 'myapp' uses 'latest' image tag                      │
  │     (Source: values.yaml -> .Values.image.tag)                     │
  └───────────────────────────────────────────────────── [ 1 violation ]

  Summary ➔ 1 violation in 3 templates.

---

### What changed in this format?
* **Logical Boxes:** Used box-drawing characters (`┌─`, `│`, `└─`) to group errors by resource.
* **Visual Symbols:** Replaced plain text `ERROR` with high-visibility symbols (`✖`, `⚠`, `✔`).
* **Action Pointers:** Included the `▲───` marker to show exactly where in the YAML the issue occurred
```
