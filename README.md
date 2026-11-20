# ManUp Kubernetes Suite (`manup-k8s`)

[![CI](https://github.com/Amanbig/manup-k8s/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/Amanbig/manup-k8s/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go)](https://golang.org)
[![Helm](https://img.shields.io/badge/Helm-v3-0F1689?logo=helm)](https://helm.sh)
[![Docker](https://img.shields.io/badge/Docker-Multi--Arch-2496ED?logo=docker)](https://hub.docker.com/r/procoder588/manup)

Official Kubernetes integration suite for **[ManUp](https://github.com/procoder588/manup)** — the modern self-hosted secrets & credential management vault.

This repository provides:
1. **Official Helm Chart** (`charts/manup`): Production-grade deployment for the ManUp server with embedded persistence (PGLite) or external PostgreSQL, Ingress, and health monitoring.
2. **Mutating Admission Webhook Controller** (`injector`): An admission controller (similar to HashiCorp's `vault-k8s`) that automatically intercepts Pod creation and injects decrypted secrets into in-memory RAM volumes.

---

## 🏗️ Architecture

```text
 1. Developer applies Deployment with annotations:
    annotations:
      manup.io/inject: "true"
      manup.io/project: "payments"
      manup.io/env: "production"
              │
              ▼
 2. Kubernetes API Server intercepts Pod creation (AdmissionReview)
              │
              ▼
 3. manup-injector Controller patches the Pod:
    • Adds in-memory RAM volume (tmpfs: medium: Memory)
    • Injects lightweight init-container (manup-secrets-init)
    • Mounts RAM volume read-only into application container(s)
              │
              ▼
 4. Pod runs: Secrets loaded in RAM before application code starts!
```

---

## ⚡ Highlights & Features

- **Zero-Storage in Kubernetes etcd**: Secrets are fetched dynamically and kept exclusively in a `tmpfs` RAM volume (`medium: Memory`). Plaintext secrets never touch physical disks or Kubernetes etcd.
- **Zero-Config TLS**: The Helm chart automatically provisions internal CA and webhook server certificates on the fly using Helm's native cryptographic engine—no `cert-manager` required.
- **Scoped API Key Isolation**: Supports project-level and environment-level API keys. An API key provisioned for `dev` is rejected if requested for `production`.
- **Developer-Friendly**: No manual volume mounts or curl commands needed in application manifests. Add 3 annotations and deploy.
- **Dual Support**: Supports both **ManUp direct injection** and **HashiCorp Vault sidecar injection**.

---

## 📁 Repository Structure

```text
manup-k8s/
├── charts/
│   └── manup/                      # Production Helm Chart
│       ├── Chart.yaml
│       ├── values.yaml             # Complete configuration reference
│       └── templates/              # Deployment, PVC, Service, Webhook manifests
├── injector/                       # Go Mutating Admission Webhook Controller
│   ├── main.go                     # Webhook server & entrypoint
│   ├── webhook.go                  # Mutation logic & JSON patch generator
│   ├── tls.go                      # Self-signed certificate generator
│   ├── webhook_test.go             # Unit tests
│   ├── Dockerfile                  # Lightweight multi-stage build (~6.5MB)
│   └── deploy/                     # Raw manifests & install script
├── docs/                           # Deep-dive documentation
│   ├── annotations.md              # Complete annotation reference & options
│   ├── architecture.md             # Architectural lifecycle & flow diagrams
│   └── application-guides.md       # Integration guide for Node, Python, Go, Java
└── .github/
    └── workflows/
        └── docker-publish.yml      # Tag-based multi-arch build & release workflow
```

---

## 🚀 Installation & Quick Start

You can install ManUp and the Secret Injector either directly from the **published GitHub Release package** or from **cloned source**. Both use the pre-built Docker Hub images (`procoder588/manup:latest` and `procoder588/manup-injector:latest`).

### Option 1: Direct Install via GitHub Release (Recommended)

No git clone needed! Install directly using the released chart archive:

```bash
helm install manup https://github.com/Amanbig/manup-k8s/releases/download/v0.1.1/manup-0.1.1.tgz \
  --namespace manup --create-namespace \
  --set injector.enabled=true \
  --set secrets.masterKey="9a8b7c6d5e4f3g2h1i0j9k8l7m6n5o4p" \
  --set secrets.jwtSecret="super_secret_jwt_sign_key_manup_2026" \
  --set secrets.refreshTokenSecret="a_different_refresh_token_sign_key_2026" \
  --set secrets.defaultAdminPassword="adminpassword123"
```

### Option 2: Install from Cloned Repository

```bash
git clone https://github.com/Amanbig/manup-k8s.git
cd manup-k8s

helm install manup ./charts/manup \
  --namespace manup --create-namespace \
  --set injector.enabled=true \
  --set secrets.masterKey="9a8b7c6d5e4f3g2h1i0j9k8l7m6n5o4p" \
  --set secrets.jwtSecret="super_secret_jwt_sign_key_manup_2026" \
  --set secrets.refreshTokenSecret="a_different_refresh_token_sign_key_2026" \
  --set secrets.defaultAdminPassword="adminpassword123"
```

```bash
git clone https://github.com/Amanbig/manup-k8s.git
cd manup-k8s

helm install manup ./charts/manup \
  --namespace manup --create-namespace \
  --set injector.enabled=true \
  --set secrets.masterKey="9a8b7c6d5e4f3g2h1i0j9k8l7m6n5o4p" \
  --set secrets.jwtSecret="super_secret_jwt_sign_key_manup_2026" \
  --set secrets.refreshTokenSecret="a_different_refresh_token_sign_key_2026" \
  --set secrets.defaultAdminPassword="adminpassword123"
```

---

### Step 2: Store Your ManUp API Key in Kubernetes

In ManUp, create a project (e.g. `billing`) and an API Key. Save the API key in a secret in your application's namespace:

```bash
kubectl create secret generic manup-api-key \
  --from-literal=MANUP_API_KEY="mp_your_manup_api_key_here"
```

### 3. Deploy Your Application (With Automatic Secret Injection)

Annotate your application Deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: billing-service
spec:
  replicas: 1
  selector:
    matchLabels:
      app: billing-service
  template:
    metadata:
      labels:
        app: billing-service
      annotations:
        # Enable secret injection
        manup.io/inject: "true"
        # Project and environment in ManUp
        manup.io/project: "billing"
        manup.io/env: "production"
        # Name of secret containing MANUP_API_KEY
        manup.io/api-key-secret: "manup-api-key"
        # Target in-memory mount path inside your app
        manup.io/mount-path: "/app/secrets"
        manup.io/file-name: ".env"
    spec:
      containers:
        - name: app
          image: my-company/billing-service:latest
          # Application automatically reads /app/secrets/.env
```

---

## 📋 Annotation Reference

| Annotation | Description | Default |
|---|---|---|
| `manup.io/inject` | Set to `"true"` to trigger automatic secret injection | `"false"` |
| `manup.io/project` | Name of the project in ManUp | *(Required)* |
| `manup.io/env` | Target environment in ManUp (`production`, `staging`, `dev`) | `"prod"` |
| `manup.io/server-url` | Internal URL of ManUp service | `http://manup.manup.svc:7780` |
| `manup.io/api-key-secret` | Kubernetes secret name containing your ManUp API key | `manup-api-key` |
| `manup.io/api-key-secret-key` | Key inside the secret containing the API key | `MANUP_API_KEY` |
| `manup.io/mount-path` | Directory mounted in RAM inside your application container | `/app/secrets` |
| `manup.io/file-name` | File name inside the mount path | `.env` |

👉 *For the full annotations guide and cross-environment isolation, see [docs/annotations.md](docs/annotations.md).*

---

## 💻 Application Integration

Secrets are formatted as standard `.env` key-value pairs in memory. Loading them in popular frameworks requires zero custom code:

- **Node.js**: `import 'dotenv/config'` or `require('dotenv').config({ path: '/app/secrets/.env' })`
- **Python**: `from dotenv import load_dotenv; load_dotenv('/app/secrets/.env')`
- **Go**: `godotenv.Load("/app/secrets/.env")`
- **Shell / Any Binary**: `. /app/secrets/.env && exec ./my-binary`

👉 *For code examples in Node.js, Python, Go, and Java, see [docs/application-guides.md](docs/application-guides.md).*

---

## 🔒 Alternative: HashiCorp Vault Direct Secret Injection

If your organization uses **HashiCorp Vault**, you can deploy ManUp with direct Vault agent injection by enabling it in `values.yaml`:

```bash
helm install manup ./charts/manup \
  --namespace manup --create-namespace \
  --set vault.enabled=true \
  --set vault.role="manup-role" \
  --set vault.secretPath="secret/data/manup"
```

---

## 📦 Automated Release Pipeline

The repository includes a GitHub Actions workflow that synchronizes versions across **both the Docker image and Helm Chart**:

```bash
git tag v0.1.0
git push origin v0.1.0
```

1. **Builds multi-arch images** (`linux/amd64`, `linux/arm64`) and pushes `procoder588/manup-injector:0.1.0` and `:latest` to Docker Hub and GHCR.
2. **Packages the Helm Chart** as `manup-0.1.0.tgz`.
3. **Creates a draft GitHub Release** with the packaged chart attached.

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
