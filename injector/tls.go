package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

// CertificatePair holds PEM-encoded certificate, private key, and CA certificate
type CertificatePair struct {
	CertPEM []byte
	KeyPEM  []byte
	CAPEM   []byte
}

// GenerateSelfSignedCert generates a CA and server certificate valid for the given DNS names
func GenerateSelfSignedCert(serviceName, namespace string) (*CertificatePair, error) {
	// 1. Generate Root CA
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate CA private key: %w", err)
	}

	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2026),
		Subject: pkix.Name{
			Organization: []string{"ManUp Security"},
			CommonName:   "ManUp Injector Root CA",
		},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create CA certificate: %w", err)
	}

	var caPEM bytes.Buffer
	if err := pem.Encode(&caPEM, &pem.Block{Type: "CERTIFICATE", Bytes: caDER}); err != nil {
		return nil, err
	}

	// 2. Generate Server Certificate
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate server private key: %w", err)
	}

	dnsNames := []string{
		serviceName,
		fmt.Sprintf("%s.%s", serviceName, namespace),
		fmt.Sprintf("%s.%s.svc", serviceName, namespace),
		fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace),
		"localhost",
	}

	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2027),
		Subject: pkix.Name{
			Organization: []string{"ManUp Security"},
			CommonName:   fmt.Sprintf("%s.%s.svc", serviceName, namespace),
		},
		NotBefore:    time.Now().Add(-1 * time.Hour),
		NotAfter:     time.Now().Add(5 * 365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     dnsNames,
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}

	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create server certificate: %w", err)
	}

	var serverCertPEM bytes.Buffer
	if err := pem.Encode(&serverCertPEM, &pem.Block{Type: "CERTIFICATE", Bytes: serverDER}); err != nil {
		return nil, err
	}

	var serverKeyPEM bytes.Buffer
	serverKeyBytes := x509.MarshalPKCS1PrivateKey(serverKey)
	if err := pem.Encode(&serverKeyPEM, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: serverKeyBytes}); err != nil {
		return nil, err
	}

	return &CertificatePair{
		CertPEM: serverCertPEM.Bytes(),
		KeyPEM:  serverKeyPEM.Bytes(),
		CAPEM:   caPEM.Bytes(),
	}, nil
}
