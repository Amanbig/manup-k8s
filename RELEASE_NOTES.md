# ManUp Kubernetes Suite Release Notes

> Current Version: **v0.1.1**  
> For previous versions, explore the [Release History](releases/).

---

## 🚀 What's New in v0.1.1

Release **v0.1.1** introduces release automation, zero-configuration TLS certificate provisioning, and unified version synchronization across Docker images and Helm packages.

---

### 🛡️ 1. Automated Version Synchronization
- **Tag-Driven Build Pipeline**: Pushing a Git tag (`v*`) triggers automated multi-arch builds for `linux/amd64` and `linux/arm64`.
- **Synchronized Helm Packaging**: The chart version in `Chart.yaml` is dynamically packaged to match the Git release tag (`manup-0.1.1.tgz`).
- **Default Image Matching**: The webhook injector deployment automatically defaults its image tag to match the Helm chart version.

---

### 🔒 2. Zero-Config Webhook TLS Provisioning
- **Native Cryptographic Generation**: The Helm chart uses Helm's built-in Sprig functions (`genCA` and `genSignedCert`) to issue mutual TLS certificates on the fly.
- **Zero Third-Party Dependencies**: No need for `cert-manager` or external OpenSSL certificate management scripts.
- **CA Bundle Registration**: The generated CA bundle is automatically embedded into the Kubernetes `MutatingWebhookConfiguration`.

---

### 📦 3. Dual Release Distribution (Direct URL & OCI)
- **Direct Asset Installation**: Charts are automatically packaged and attached to GitHub Releases for direct URL installation.
- **OCI Registry Support**: Charts are pushed to GitHub Container Registry (`ghcr.io/Amanbig/charts/manup`).

---

## 📜 Historical Releases

| Release | Date | Description | Changelog |
|---|---|---|---|
| **v0.1.1** | Nov 20, 2026 | Version sync, zero-config TLS, OCI & direct release assets | [v0.1.1.md](releases/v0.1.1.md) |
| **v0.1.0** | Nov 20, 2026 | Initial public release of Helm Chart & Injector Controller | [v0.1.0.md](releases/v0.1.0.md) |
