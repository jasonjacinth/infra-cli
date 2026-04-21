# Infra-CLI

A lightweight, single-binary command-line tool that abstracts `kubectl` and `docker` commands into a simplified, developer-friendly workflow. Built to standardize infrastructure operations and reduce onboarding time for engineering teams.

---

## Why This Exists

In large-scale microservice environments, application developers often spend too much time fighting infrastructure tooling instead of writing business logic. Infra-CLI was developed to treat the internal platform as a product, abstracting the complexity of local and remote container orchestration into a secure, "paved path" developer experience. It reduces developer cognitive load, abstracts environment parity, and enforces operational guardrails (like namespace protection and production confirmations), drastically reducing L1 support tickets for the SRE team.

## Features

| Feature | Description |
|---|---|
| **Environment Abstraction** | Automatically routes commands to Docker (local) or Kubernetes (production) via the `-e` flag. |
| **Pod Auto-Resolution** | `infra-cli logs --app my-service` automatically finds the pod `my-service-7d4b8c6f5-xk9zn` — no copy-pasting pod names. |
| **Dependency Checker** | `infra-cli setup` validates your toolchain and provides install links for anything missing. |
| **Namespace Protection** | Blocks operations against `kube-system` and other restricted namespaces at the guardrail layer — not at the cluster level. |
| **Production Confirmation** | `deploy`, `rollback`, and `cleanup` require explicit `y/N` consent (or `--force`) when targeting production. |
| **Cross-Platform Binary** | Compiles to a self-contained binary for macOS (ARM/Intel) and Linux — distribute via any internal repository. |

---

## Installation

### Prerequisites

- [Go 1.21+](https://go.dev/dl/) (for building from source)
- [Docker Desktop](https://docs.docker.com/get-docker/) (for local workflows)
- [kubectl](https://kubernetes.io/docs/tasks/tools/) (for Kubernetes workflows)

### Build from Source

```bash
git clone https://github.com/jasonjacinth/infra-cli.git
cd infra-cli/project
make build
```

The binary is compiled to `./bin/infra-cli`.

### Install System-Wide

```bash
sudo cp ./bin/infra-cli /usr/local/bin/
infra-cli --help
```

### Cross-Compile for Distribution

```bash
make build-all
```

This produces binaries for:
- `bin/infra-cli-darwin-arm64` — macOS (Apple Silicon)
- `bin/infra-cli-darwin-amd64` — macOS (Intel)
- `bin/infra-cli-linux-amd64` — Linux

---

## Quick Start

### 1. Verify Dependencies

```bash
infra-cli setup
```

```
Checking system dependencies...

  Docker:  Docker version 27.x.x
  kubectl: Client Version: v1.34.1

All dependencies are installed. You're ready to go!
```

### 2. Deploy the Sample Workload

```bash
infra-cli deploy --app nginx -e production
```

```
Deploying 'nginx' to Kubernetes using k8s/deployment.yaml...

deployment.apps/nginx created
service/nginx created

'nginx' deployed to Kubernetes.
```

### 3. Check Status

```bash
infra-cli status -e production
```

```
Kubernetes Pod Status
──────────────────────────────────────────────────
NAME                     READY   STATUS    RESTARTS   AGE
nginx-6b7f6db5c7-4xr2k  1/1     Running   0          30s
nginx-6b7f6db5c7-m9plj  1/1     Running   0          30s
```

### 4. Tail Logs

```bash
infra-cli logs --app nginx -e production
```

The CLI automatically resolves the pod name from `nginx` — no need to copy-paste `nginx-6b7f6db5c7-4xr2k`.

### 5. Rollback a Deployment

```bash
infra-cli rollback --app nginx -e production
```

### 6. Clean Up

```bash
infra-cli cleanup
```

### 7. Enforcing Guardrails (Safety)

Infra-CLI actively prevents accidental damage to the cluster.

**Namespace Protection:**
```bash
infra-cli deploy --app nginx -e production -n kube-system
```
```
[Guardrail Violation] Operations in the 'kube-system' namespace are restricted to cluster administrators
```

**Production Confirmation:**
```bash
infra-cli rollback --app nginx -e production
```
```
You are about to run 'rollback' against the PRODUCTION environment.
Are you sure? [y/N]: 
```

---

## Command Reference

| Command | Description | Key Flags |
|---|---|---|
| `infra-cli setup` | Check if Docker and kubectl are installed | — |
| `infra-cli deploy` | Deploy an application | `--app`, `-e` |
| `infra-cli status` | Show health of running services | `--app` (optional), `-e` |
| `infra-cli logs` | Tail application logs (auto-resolves pod names) | `--app`, `-e` |
| `infra-cli rollback` | Revert the last deployment | `--app`, `-e` |
| `infra-cli cleanup` | Remove deployed test workloads | `--dir` (default: `k8s/`) |

### Global Flags

| Flag | Short | Default | Description |
|---|---|---|---|
| `--environment` | `-e` | `local` | Target environment: `local` (Docker) or `production` (Kubernetes) |
| `--namespace` | `-n` | `default` | Kubernetes namespace to operate in (`kube-system` is restricted) |
| `--force` | — | `false` | Bypass production confirmation prompts (for CI/CD pipelines) |
| `--help` | `-h` | — | Show help for any command |

---

## Project Structure

```
project/
├── main.go                      # Entry point
├── Makefile                     # Build targets (build, build-all, clean, run, test, vet)
├── go.mod                       # Go module definition
├── cmd/
│   ├── root.go                  # Root command + global flags (env, namespace, force)
│   ├── version.go               # Version command (build metadata via ldflags)
│   ├── setup.go                 # Dependency checker
│   ├── deploy.go                # Deploy via kubectl apply (namespace-aware, prod-gated)
│   ├── status.go                # Service health via docker ps / kubectl get pods
│   ├── logs.go                  # Log tailing with pod auto-resolution
│   ├── rollback.go              # Rollback (namespace-aware, prod-gated)
│   ├── cleanup.go               # Tear down workloads (namespace-aware, prod-gated)
│   ├── root_test.go             # Tests: subcommand registration
│   └── deploy_test.go           # Tests: deploy flag validation
├── internal/
│   ├── guardrail/
│   │   ├── guardrail.go         # Namespace protection + production confirmation logic
│   │   └── guardrail_test.go    # Tests: allowed/restricted namespace enforcement
│   ├── shell/
│   │   ├── shell.go             # os/exec wrapper with error handling
│   │   └── shell_test.go        # Tests: IsInstalled, Run success/failure
│   └── style/
│       └── style.go             # Centralized lipgloss terminal styles
└── k8s/
    ├── deployment.yaml          # Sample Nginx deployment (2 replicas)
    └── service.yaml             # NodePort service on port 30080
```

---

## Tech Stack

- **Go** — compiles to a static, self-contained binary with zero runtime dependencies.
- **[Cobra](https://github.com/spf13/cobra)** — industry-standard CLI framework for Go.
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** — terminal styling library for colored, formatted output.
- **os/exec** — wraps existing `docker` and `kubectl` binaries rather than importing full SDKs, keeping the codebase lean and uncomplicated.

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

Developed and maintained by Jason Jacinth