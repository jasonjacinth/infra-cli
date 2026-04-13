# Infra-CLI

A lightweight, single-binary command-line tool that abstracts `kubectl` and `docker` commands into a simplified, developer-friendly workflow. Built to standardize infrastructure operations and reduce onboarding time for engineering teams.

---

## Why This Exists

Managing Kubernetes clusters and Docker containers requires memorizing long, flag-heavy commands that vary between environments. **Infra-CLI** was designed to solve this by providing a unified interface that:

- **Standardizes workflows** — one set of commands works across local Docker and remote Kubernetes environments.
- **Reduces onboarding time** — new engineers can deploy, debug, and monitor services on day one with `infra-cli setup` and a handful of intuitive commands.
- **Eliminates context-switching** — no need to remember whether you're running `docker ps` or `kubectl get pods`; just run `infra-cli status`.

## Features

| Feature | Description |
|---|---|
| **Environment Abstraction** | Automatically routes commands to Docker (local) or Kubernetes (production) via the `-e` flag. |
| **Pod Auto-Resolution** | `infra-cli logs --app my-service` automatically finds the pod `my-service-7d4b8c6f5-xk9zn` — no copy-pasting pod names. |
| **Dependency Checker** | `infra-cli setup` validates your toolchain and provides install links for anything missing. |
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
🔍 Checking system dependencies...

  ✅ Docker:  Docker version 27.x.x
  ✅ kubectl: Client Version: v1.34.1

🎉 All dependencies are installed. You're ready to go!
```

### 2. Deploy the Sample Workload

```bash
infra-cli deploy --app nginx -e production
```

```
🚀 Deploying 'nginx' to Kubernetes using k8s/deployment.yaml...

deployment.apps/nginx created
service/nginx created

✅ 'nginx' deployed to Kubernetes.
```

### 3. Check Status

```bash
infra-cli status -e production
```

```
☸️  Kubernetes Pod Status
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
| `--help` | `-h` | — | Show help for any command |

---

## Project Structure

```
project/
├── main.go                  # Entry point
├── Makefile                 # Build targets (build, build-all, clean, run)
├── go.mod                   # Go module definition
├── cmd/
│   ├── root.go              # Root command + global flags
│   ├── setup.go             # Dependency checker
│   ├── deploy.go            # Deploy via docker-compose / kubectl apply
│   ├── status.go            # Service health via docker ps / kubectl get pods
│   ├── logs.go              # Log tailing with pod auto-resolution
│   ├── rollback.go          # Rollback via docker restart / kubectl rollout undo
│   └── cleanup.go           # Tear down test workloads
├── internal/
│   └── shell/
│       └── shell.go         # os/exec wrapper with error handling
└── k8s/
    ├── deployment.yaml      # Sample Nginx deployment (2 replicas)
    └── service.yaml         # NodePort service on port 30080
```

---

## Tech Stack

- **Go** — compiles to a static, self-contained binary with zero runtime dependencies.
- **[Cobra](https://github.com/spf13/cobra)** — industry-standard CLI framework for Go.
- **os/exec** — wraps existing `docker` and `kubectl` binaries rather than importing full SDKs, keeping the codebase lean and uncomplicated.

---

## License

This project is for educational and portfolio purposes.
