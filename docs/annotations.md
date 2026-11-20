# ManUp Kubernetes Annotations Reference

This document provides the complete specification for the annotations supported by the **ManUp Injector** Mutating Admission Controller.

---

## 📋 Full Annotations Reference

| Annotation | Type | Required | Default | Description |
|---|---|---|---|---|
| `manup.io/inject` | String | **Yes** | `"false"` | Set to `"true"`, `"enabled"`, or `"yes"` to activate automatic secret injection for this pod. |
| `manup.io/project` | String | **Yes** | `""` | Name or ID of the project in ManUp that owns the secrets. |
| `manup.io/env` | String | No | `"prod"` | Target environment name (e.g. `production`, `staging`, `dev`). |
| `manup.io/server-url` | String | No | `http://manup.manup.svc:7780` | URL of the ManUp backend service accessible within the cluster. |
| `manup.io/api-key-secret` | String | No | `manup-api-key` | Name of the Kubernetes `Secret` in the pod's namespace containing the ManUp API Key. |
| `manup.io/api-key-secret-key` | String | No | `MANUP_API_KEY` | The key name inside the Kubernetes Secret that holds the API key value. |
| `manup.io/mount-path` | String | No | `/app/secrets` | Directory path where the in-memory volume will be mounted inside your application containers. |
| `manup.io/file-name` | String | No | `.env` | File name written inside the mount path (e.g. `.env`, `config.env`). |
| `manup.io/init-image` | String | No | `curlimages/curl:8.11.1` | Container image used by the lightweight init container to fetch secrets. |

---

## 🎯 Usage Examples by Environment

### 1. Development Deployment
```yaml
metadata:
  annotations:
    manup.io/inject: "true"
    manup.io/project: "payments"
    manup.io/env: "development"
    manup.io/api-key-secret: "manup-dev-key"
    manup.io/mount-path: "/app/secrets"
    manup.io/file-name: ".env"
```

### 2. Staging Deployment
```yaml
metadata:
  annotations:
    manup.io/inject: "true"
    manup.io/project: "payments"
    manup.io/env: "staging"
    manup.io/api-key-secret: "manup-staging-key"
    manup.io/mount-path: "/app/secrets"
    manup.io/file-name: ".env"
```

### 3. Production Deployment
```yaml
metadata:
  annotations:
    manup.io/inject: "true"
    manup.io/project: "payments"
    manup.io/env: "production"
    manup.io/api-key-secret: "manup-prod-key"
    manup.io/mount-path: "/app/secrets"
    manup.io/file-name: ".env"
```

---

## 🔒 Scoped API Keys (Enforced Isolation)

In ManUp, API keys can be provisioned with **Project-level** and **Environment-level** scopes. 

When you use an environment-scoped API key:
- The API key itself is restricted by ManUp to only access that specific environment.
- If a developer sets `manup.io/env: "production"` using a staging API key, ManUp's backend authorization **rejects the query** with `HTTP 403 Forbidden`.
- This prevents cross-environment privilege escalation or accidental leakage.
