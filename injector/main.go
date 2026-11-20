package main

import (
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		port            int
		certFile        string
		keyFile         string
		serviceName     string
		namespace       string
		defaultManUpURL string
		printCABundle   bool
	)

	flag.IntVar(&port, "port", 8443, "Secure port for webhook server")
	flag.StringVar(&certFile, "tls-cert-file", "", "File containing the x509 Certificate for HTTPS")
	flag.StringVar(&keyFile, "tls-key-file", "", "File containing the x509 private key")
	flag.StringVar(&serviceName, "service-name", "manup-injector", "Kubernetes service name for the webhook")
	flag.StringVar(&namespace, "namespace", "manup-system", "Kubernetes namespace for the webhook")
	flag.StringVar(&defaultManUpURL, "default-manup-url", "http://manup.manup.svc:7780", "Default ManUp server URL")
	flag.BoolVar(&printCABundle, "print-ca-bundle", false, "Print base64-encoded CA bundle and exit")
	flag.Parse()

	// If environment variables are set, use them
	if envURL := os.Getenv("DEFAULT_MANUP_URL"); envURL != "" {
		defaultManUpURL = envURL
	}
	if envNS := os.Getenv("POD_NAMESPACE"); envNS != "" {
		namespace = envNS
	}

	var tlsConfig *tls.Config

	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.Fatalf("Failed to load provided TLS key pair: %v", err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		log.Printf("Loaded TLS certificate from files %s and %s", certFile, keyFile)
	} else {
		log.Printf("Generating self-signed TLS certificate for service %s in namespace %s...", serviceName, namespace)
		certPair, err := GenerateSelfSignedCert(serviceName, namespace)
		if err != nil {
			log.Fatalf("Failed to generate self-signed certificate: %v", err)
		}

		caBase64 := base64.StdEncoding.EncodeToString(certPair.CAPEM)
		if printCABundle {
			fmt.Println(caBase64)
			os.Exit(0)
		}

		log.Printf("----------------------------------------------------------------")
		log.Printf("MutatingWebhookConfiguration caBundle:")
		log.Printf("%s", caBase64)
		log.Printf("----------------------------------------------------------------")

		cert, err := tls.X509KeyPair(certPair.CertPEM, certPair.KeyPEM)
		if err != nil {
			log.Fatalf("Failed to parse generated TLS keypair: %v", err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	webhookServer := &WebhookServer{
		DefaultServerURL: defaultManUpURL,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", webhookServer.HandleMutate)
	mux.HandleFunc("/healthz", webhookServer.HandleHealthz)

	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	go func() {
		log.Printf("Starting ManUp Mutating Admission Webhook server on port %d...", port)
		if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Listen for termination signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down ManUp Webhook server gracefully...")
}
