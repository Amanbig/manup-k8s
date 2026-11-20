# ManUp Kubernetes Architecture & Lifecycle

This document describes the architectural flow, component interactions, and security boundary when deploying ManUp and using the `manup-injector` controller.

---

## 🏗️ Architecture Diagram

```text
                               Kubernetes Control Plane
                             ┌─────────────────────────┐
                             │  kube-apiserver         │
                             └────────────┬────────────┘
                                          │
                    1. Pod Creation Event │ (AdmissionReview v1 over mTLS)
                                          ▼
                               ┌─────────────────────┐
                               │   manup-injector    │
                               │  (Admission Webhook)│
                               └──────────┬──────────┘
                                          │
                 2. Generates JSON Patch  │ (Adds RAM tmpfs volume, init-container,
                    & returns to API      │  and container volumeMounts)
                                          ▼
                             ┌─────────────────────────┐
                             │  kube-scheduler / kubelet│
                             └────────────┬────────────┘
                                          │
                                          ▼
                    3. Launches Pod on Worker Node
   ┌───────────────────────────────────────────────────────────────────────┐
   │ Pod: demo-service                                                    │
   │                                                                       │
   │  ┌──────────────────────────────┐                                     │
   │  │ initContainer:               │ 4. Calls ManUp REST API with        │
   │  │ manup-secrets-init           ├─────────────────┐                   │
   │  └──────────────┬───────────────┘                 │                   │
   │                 │ 5. Writes .env to RAM           ▼                   │
   │                 ▼                               ┌───────────────────┐ │
   │          ┌──────────────┐                       │   manup-server    │ │
   │          │ emptyDir     │                       │  (:7780 ClusterIP)│ │
   │          │ medium:Memory│                       └───────────────────┘ │
   │          └──────┬───────┘                                             │
   │                 │ 6. Mounts read-only                                 │
   │                 ▼                                                     │
   │  ┌──────────────────────────────┐                                     │
   │  │ app container                │                                     │
   │  │ (Loads /app/secrets/.env)    │                                     │
   │  └──────────────────────────────┘                                     │
   └───────────────────────────────────────────────────────────────────────┘
```

---

## 🔄 Pod Mutation Lifecycle

1. **Submission**:
   A developer or CI/CD pipeline applies a Deployment manifest to Kubernetes.
2. **Admission Interception**:
   Before writing the Pod to etcd, the Kubernetes API Server issues an HTTPS `AdmissionReview` request to `manup-injector.manup-system.svc:443/mutate`.
3. **Annotation Check**:
   - If `manup.io/inject: "true"` is absent, the webhook returns `allowed: true` with zero modifications.
   - If present, the webhook parses the project, environment, mount path, and secret parameters.
4. **JSON Patch Construction**:
   The webhook returns an RFC 6902 JSON Patch adding:
   - **Volume**: A dedicated `emptyDir` volume configured with `medium: Memory` (`tmpfs`).
   - **Init Container**: `manup-secrets-init` configured with the secret fetch script and `MANUP_API_KEY` bound from the specified Secret.
   - **Volume Mounts**: The `manup-secrets` volume is mounted read-only at `/app/secrets` across the application containers.
5. **Execution**:
   The Kubernetes kubelet creates the Pod. The init container runs first, downloads the decrypted environment into RAM, and exits cleanly (`0`). The application container starts with the `.env` file pre-loaded.
