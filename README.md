# ManUp Kubernetes Suite (`manup-k8s`)

This repository contains the Kubernetes deployment tooling for [ManUp](https://github.com/procoder588/manup), including the official **Helm Chart** and the **Mutating Admission Controller (`manup-injector`)**.

---

## 📁 Repository Structure

```text
manup-k8s/
├── charts/
│   └── manup/                    # Official Helm Chart (ManUp Server + Injector)
├── injector/                     # Go Mutating Admission Webhook Controller
│   ├── main.go
│   ├── webhook.go
│   ├── tls.go
│   ├── Dockerfile
│   └── deploy/                   # Standalone raw manifests
└── .github/
    └── workflows/
        └── docker-publish.yml    # Multi-arch automated tag build & push
```

---

## 🚀 Quick Start: Deploying ManUp to Kubernetes

### 1. Install via Helm (with Auto Secret Injector)

```bash
helm install manup ./charts/manup \
  --namespace manup --create-namespace \
  --set injector.enabled=true \
  --set secrets.masterKey="9a8b7c6d5e4f3g2h1i0j9k8l7m6n5o4p" \
  --set secrets.jwtSecret="super_secret_jwt_sign_key_manup_2026" \
  --set secrets.refreshTokenSecret="a_different_refresh_token_sign_key_2026" \
  --set secrets.defaultAdminPassword="adminpassword123"
```

### 2. Auto-Inject Secrets into Any Application

Once installed, annotate any Pod or Deployment in your cluster:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: demo-service
spec:
  template:
    metadata:
      annotations:
        manup.io/inject: "true"
        manup.io/project: "payments"
        manup.io/env: "production"
        manup.io/api-key-secret: "manup-api-key"
        manup.io/mount-path: "/app/secrets"
    spec:
      containers:
        - name: app
          image: node:20-alpine
          command: ["sh", "-c", "cat /app/secrets/.env && node index.js"]
```

The controller intercepts the pod creation, mounts an in-memory `tmpfs` volume, queries ManUp via the scoped API key, and mounts the decrypted `.env` file directly into your container!

---

## 🔒 HashiCorp Vault Integration (Alternative)

If you use HashiCorp Vault as an external key management system, ManUp itself can have its secrets directly injected via HashiCorp Vault Agent sidecar:

```bash
helm install manup ./charts/manup \
  --namespace manup --create-namespace \
  --set vault.enabled=true \
  --set vault.role="manup-role" \
  --set vault.secretPath="secret/data/manup"
```

---

## 🛠️ Automated CI/CD Releases

Docker images for `manup-injector` are built automatically on tag pushes:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Required GitHub Secrets:
- `DOCKERHUB_USERNAME`
- `DOCKERHUB_TOKEN`
