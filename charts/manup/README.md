# ManUp Helm Chart

Production-ready Helm chart for deploying [ManUp](https://github.com/procoder588/manup) on Kubernetes, with native support for **HashiCorp Vault direct secrets injection**, embedded database persistence (PGLite), and external PostgreSQL databases.

---

## 🚀 Quick Start

### 1. Standard Installation (using Kubernetes Secrets)

```bash
helm install manup ./manup-chart \
  --namespace manup --create-namespace \
  --set secrets.masterKey="a-secure-32-byte-master-key-here" \
  --set secrets.jwtSecret="your-jwt-secret-here" \
  --set secrets.refreshTokenSecret="your-refresh-secret-here" \
  --set secrets.defaultAdminPassword="YourStrongPassword123"
```

---

## 🔒 HashiCorp Vault Direct Secret Injection

This chart supports **zero-secret storage in Kubernetes etcd** by directly integrating with HashiCorp Vault using the **Vault Agent Sidecar Injector**.

### How It Works:
1. When `vault.enabled: true`, the chart annotates the Pod with Vault agent injector annotations.
2. The Vault Agent sidecar authenticates against Vault using the Pod's Kubernetes `ServiceAccount`.
3. Secrets from Vault (`secret/data/manup`) are rendered to an in-memory volume at `/vault/secrets/config.env`.
4. The container automatically sources `/vault/secrets/config.env` upon startup before starting the Node process. No plain secrets are kept in Kubernetes etcd!

### Prerequisites in Vault:
1. Enable Kubernetes auth method in Vault:
   ```bash
   vault auth enable kubernetes
   ```
2. Configure Vault Kubernetes Auth with your cluster API & token reviewer:
   ```bash
   vault write auth/kubernetes/config \
     kubernetes_host="https://kubernetes.default.svc:443"
   ```
3. Create a Vault policy `manup-policy`:
   ```hcl
   path "secret/data/manup" {
     capabilities = ["read"]
   }
   ```
4. Create a Vault role binding the policy to ManUp's ServiceAccount:
   ```bash
   vault write auth/kubernetes/role/manup-role \
     bound_service_account_names=manup \
     bound_service_account_namespaces=manup \
     policies=manup-policy \
     ttl=1h
   ```
5. Store your secrets in Vault:
   ```bash
   vault kv put secret/manup \
     MASTER_KEY="9a8b7c6d5e4f3g2h1i0j9k8l7m6n5o4p" \
     JWT_SECRET="super_secret_jwt_sign_key_manup_2026" \
     REFRESH_TOKEN_SECRET="a_different_refresh_token_sign_key_2026" \
     DEFAULT_ADMIN_EMAIL="admin@manup.io" \
     DEFAULT_ADMIN_PASSWORD="adminpassword123"
   ```

### Deploying ManUp with Vault:
```bash
helm install manup ./manup-chart \
  --namespace manup --create-namespace \
  --set vault.enabled=true \
  --set vault.role="manup-role" \
  --set vault.secretPath="secret/data/manup"
```

---

## ⚙️ Configuration Reference

### Common Parameters

| Parameter | Description | Default |
|---|---|---|
| `image.repository` | Container image repository | `procoder588/manup` |
| `image.tag` | Container image tag | `""` (defaults to chart `appVersion`) |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `service.type` | Kubernetes Service type | `ClusterIP` |
| `service.port` | Service port | `7780` |
| `persistence.enabled` | Persistent Volume Claim for embedded PGLite database | `true` |
| `persistence.size` | PVC storage size | `5Gi` |
| `config.dbType` | Database type: `PGLITE` or `POSTGRES` | `PGLITE` |
| `config.port` | Application internal port | `7780` |
| `injector.enabled` | Deploy built-in Mutating Admission Webhook Injector | `false` |
| `injector.image.repository` | Webhook controller image | `procoder588/manup-injector` |
| `injector.image.tag` | Webhook controller image tag | `latest` |
| `livenessProbe.httpGet.path` | Health check endpoint | `/health` |

### Vault Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `vault.enabled` | Enable HashiCorp Vault Agent Sidecar injection | `false` |
| `vault.role` | Vault Kubernetes auth role name | `manup-role` |
| `vault.secretPath` | Vault secret KV path | `secret/data/manup` |
| `vault.authPath` | Custom Kubernetes auth mount path | `""` |
| `vault.serverUrl` | Optional Vault server address | `""` |
| `vault.secretFileName` | Secrets file inside `/vault/secrets/` | `config.env` |
| `vault.sourceEnvFile` | Pre-load env file before starting Node.js | `true` |
| `vault.customTemplate` | Override default Vault template | `""` |

---

## 🗄️ Database Options

### 1. Default (Embedded PGLite)
By default, `config.dbType: "PGLITE"` and `persistence.enabled: true`. A PersistentVolumeClaim will be bound to `/app/manup` to keep your database data across pod restarts.

### 2. External PostgreSQL
If using an external PostgreSQL database (AWS RDS, Cloud SQL, Supabase, etc.):
```bash
helm install manup ./manup-chart \
  --namespace manup \
  --set config.dbType="POSTGRES" \
  --set persistence.enabled=false \
  --set secrets.databaseUrl="postgres://username:password@postgres-host:5432/manup"
```
Or when using Vault, simply store `DATABASE_URL` in Vault under `secret/data/manup`.
