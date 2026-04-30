# Infra-CLI

A high-performance developer utility designed to standardize infrastructure operations. It abstracts the complexity of Kubernetes and Docker through a unified interface, enforcing strict operational guardrails and environment-aware safety checks to prevent misconfigurations and minimize manual support overhead.

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

## SRE Features

| Feature | SRE Concept | Description |
|---|---|---|
| **Canary Deployments** | Deployment Strategy | `deploy --strategy canary` scales to 1 replica, validates health, then prompts to promote or abort with automatic rollback. |
| **Health-Validated Deploys** | Rollback Strategy | Every deploy waits for rollout completion and scans for unhealthy pods. Automatically rolls back on `CrashLoopBackOff`, `ImagePullBackOff`, or other failure states. |
| **Chaos Engineering** | Resilience Testing | `chaos pod-kill --app nginx` randomly kills a pod, waits for Kubernetes to recreate it, and reports whether self-healing succeeded. |
| **Blameless Postmortems** | Incident Management | `postmortem create --title "..." --severity critical` generates a structured postmortem document with timeline, 5-Whys root cause analysis, action items, and auto-captured cluster context. |
| **SLO Validation** | SLI/SLO/SLA | `slo validate --app nginx` checks availability (replica readiness), restart budget (max 5 restarts), and pod stability (running for at least 5 minutes). |
| **Capacity Analysis** | Capacity Planning | `capacity --app nginx` displays resource requests, limits, and actual usage with utilization percentages for CPU and memory. |

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
| `infra-cli deploy` | Deploy an application (with health validation) | `--app`, `-e`, `--strategy` |
| `infra-cli status` | Show health of running services | `--app` (optional), `-e` |
| `infra-cli logs` | Tail application logs (auto-resolves pod names) | `--app`, `-e` |
| `infra-cli rollback` | Revert the last deployment | `--app`, `-e` |
| `infra-cli cleanup` | Remove deployed test workloads | `--dir` (default: `k8s/`) |
| `infra-cli chaos pod-kill` | Kill a random pod to test self-healing | `--app`, `-e` |
| `infra-cli postmortem create` | Generate a blameless postmortem document | `--title`, `--severity`, `--output` |
| `infra-cli slo validate` | Validate SLOs for an application | `--app`, `-e` |
| `infra-cli capacity` | Analyze resource usage and capacity | `--app`, `-e` |

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
│   ├── deploy.go                # Deploy with health validation, canary strategy, auto-rollback
│   ├── status.go                # Service health via docker ps / kubectl get pods
│   ├── logs.go                  # Log tailing with pod auto-resolution
│   ├── rollback.go              # Rollback (namespace-aware, prod-gated)
│   ├── cleanup.go               # Tear down workloads (namespace-aware, prod-gated)
│   ├── chaos.go                 # Chaos engineering: pod-kill with recovery reporting
│   ├── postmortem.go            # Blameless postmortem generator with cluster context
│   ├── slo.go                   # SLO validation: availability, restarts, stability
│   ├── capacity.go              # Capacity analysis: usage vs requests/limits
│   ├── root_test.go             # Tests: subcommand registration
│   ├── deploy_test.go           # Tests: deploy flag and strategy validation
│   ├── chaos_test.go            # Tests: chaos subcommand and flag registration
│   ├── postmortem_test.go       # Tests: postmortem subcommand and flag registration
│   ├── slo_test.go              # Tests: SLO subcommand and flag registration
│   └── capacity_test.go         # Tests: capacity command and flag registration
├── internal/
│   ├── guardrail/
│   │   ├── guardrail.go         # Namespace protection + production/canary confirmation logic
│   │   └── guardrail_test.go    # Tests: allowed/restricted namespace enforcement
│   ├── shell/
│   │   ├── shell.go             # os/exec wrapper with error handling
│   │   └── shell_test.go        # Tests: IsInstalled, Run success/failure
│   └── style/
│       └── style.go             # Centralized lipgloss terminal styles
└── k8s/
    ├── deployment.yaml          # Sample Nginx deployment (2 replicas, probes, resource limits)
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