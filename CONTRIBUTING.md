# Contributing to ManUp Kubernetes Suite

Thank you for your interest in contributing to **ManUp Kubernetes Suite** (`manup-k8s`)! We welcome contributions to the Helm Chart, the Mutating Injector Controller, and the documentation.

---

## 🛠️ Local Development Setup

### Prerequisites
- [Go](https://golang.org) 1.24+
- [Docker](https://docs.docker.com)
- [Helm](https://helm.sh) v3+
- A local Kubernetes cluster ([k3s](https://k3s.io), [kind](https://kind.sigs.k8s.io), or [minikube](https://minikube.sigs.k8s.io))

---

## 🧪 Testing Guidelines

### 1. Webhook Controller Tests
The controller logic is located in `injector/`:

```bash
cd injector
go test -v ./...
```

### 2. Helm Chart Validation
The Helm chart is located in `charts/manup/`:

```bash
cd charts/manup
helm lint .
helm template test-release . --set injector.enabled=true
```

---

## 📝 Pull Request Workflow

1. **Fork** the repository and create a feature branch (`git checkout -b feature/my-feature`).
2. **Commit** your changes following conventional commits (`feat: ...`, `fix: ...`, `docs: ...`).
3. Ensure all tests and linters pass.
4. **Push** to your fork and submit a Pull Request.

---

## 📜 Code of Conduct
Please adhere to our [Code of Conduct](CODE_OF_CONDUCT.md) in all project interactions.
