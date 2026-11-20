#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NAMESPACE="manup-system"
SERVICE_NAME="manup-injector"
SECRET_NAME="manup-injector-tls"

echo "==> Creating namespace ${NAMESPACE}..."
kubectl apply -f "${SCRIPT_DIR}/00-namespace.yaml"

echo "==> Generating TLS certificates for ${SERVICE_NAME}.${NAMESPACE}.svc..."
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

# Generate CA key and certificate
openssl genrsa -out "${TMP_DIR}/ca.key" 2048
openssl req -x509 -new -nodes -key "${TMP_DIR}/ca.key" -subj "/CN=ManUp Injector CA" -days 3650 -out "${TMP_DIR}/ca.crt"

# Generate Server key and CSR with SANs
openssl genrsa -out "${TMP_DIR}/server.key" 2048
cat <<EOF > "${TMP_DIR}/csr.conf"
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = ${SERVICE_NAME}
DNS.2 = ${SERVICE_NAME}.${NAMESPACE}
DNS.3 = ${SERVICE_NAME}.${NAMESPACE}.svc
DNS.4 = ${SERVICE_NAME}.${NAMESPACE}.svc.cluster.local
EOF

openssl req -new -key "${TMP_DIR}/server.key" -subj "/CN=${SERVICE_NAME}.${NAMESPACE}.svc" -out "${TMP_DIR}/server.csr" -config "${TMP_DIR}/csr.conf"
openssl x509 -req -in "${TMP_DIR}/server.csr" -CA "${TMP_DIR}/ca.crt" -CAkey "${TMP_DIR}/ca.key" -CAcreateserial -out "${TMP_DIR}/server.crt" -days 1825 -extensions v3_req -extfile "${TMP_DIR}/csr.conf"

echo "==> Storing TLS certificate in Kubernetes secret..."
kubectl create secret tls "${SECRET_NAME}" \
  --cert="${TMP_DIR}/server.crt" \
  --key="${TMP_DIR}/server.key" \
  --namespace="${NAMESPACE}" \
  --dry-run=client -o yaml | kubectl apply -f -

CA_BUNDLE=$(base64 -w 0 < "${TMP_DIR}/ca.crt" || base64 < "${TMP_DIR}/ca.crt")

echo "==> Applying RBAC..."
kubectl apply -f "${SCRIPT_DIR}/01-rbac.yaml"

echo "==> Applying Service..."
kubectl apply -f "${SCRIPT_DIR}/03-service.yaml"

echo "==> Applying Deployment..."
kubectl apply -f "${SCRIPT_DIR}/02-deployment.yaml"

echo "==> Applying MutatingWebhookConfiguration..."
sed "s|\${CA_BUNDLE}|${CA_BUNDLE}|g" "${SCRIPT_DIR}/04-webhook.yaml" | kubectl apply -f -

echo "==> Waiting for manup-injector deployment to be ready..."
kubectl rollout status deployment/${SERVICE_NAME} -n ${NAMESPACE} --timeout=60s

echo "=========================================================="
echo "✅ ManUp Mutating Admission Controller successfully installed!"
echo "Any Pod with 'manup.io/inject: \"true\"' will now be mutated."
echo "=========================================================="
