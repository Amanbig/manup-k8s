package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// Constants for ManUp annotations
const (
	AnnotationInject          = "manup.io/inject"
	AnnotationProject         = "manup.io/project"
	AnnotationEnv             = "manup.io/env"
	AnnotationServerURL       = "manup.io/server-url"
	AnnotationSecretName      = "manup.io/api-key-secret"
	AnnotationSecretKey       = "manup.io/api-key-secret-key"
	AnnotationMountPath       = "manup.io/mount-path"
	AnnotationFileName        = "manup.io/file-name"
	AnnotationInitImage       = "manup.io/init-image"
	AnnotationStatus          = "manup.io/status"
	DefaultServerURL          = "http://manup.manup.svc:7780"
	DefaultSecretName         = "manup-api-key"
	DefaultSecretKey          = "MANUP_API_KEY"
	DefaultMountPath          = "/app/secrets"
	DefaultFileName           = ".env"
	DefaultInitImage          = "curlimages/curl:8.11.1"
	VolumeName                = "manup-secrets"
)

// Kubernetes AdmissionReview types
type AdmissionReview struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Request    *AdmissionRequest  `json:"request,omitempty"`
	Response   *AdmissionResponse `json:"response,omitempty"`
}

type AdmissionRequest struct {
	UID       string          `json:"uid"`
	Kind      metav1GroupVersionKind `json:"kind"`
	Namespace string          `json:"namespace"`
	Operation string          `json:"operation"`
	Object    json.RawMessage `json:"object"`
}

type metav1GroupVersionKind struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

type AdmissionResponse struct {
	UID       string        `json:"uid"`
	Allowed   bool          `json:"allowed"`
	Patch     []byte        `json:"patch,omitempty"`
	PatchType string        `json:"patchType,omitempty"`
	Result    *metav1Status `json:"status,omitempty"`
}

type metav1Status struct {
	Message string `json:"message,omitempty"`
}

type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// Pod representation
type Pod struct {
	Metadata ObjectMeta `json:"metadata"`
	Spec     PodSpec    `json:"spec"`
}

type ObjectMeta struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels"`
}

type PodSpec struct {
	InitContainers []Container `json:"initContainers"`
	Containers     []Container `json:"containers"`
	Volumes        []Volume    `json:"volumes"`
}

type Container struct {
	Name         string        `json:"name"`
	Image        string        `json:"image"`
	Command      []string      `json:"command,omitempty"`
	Args         []string      `json:"args,omitempty"`
	Env          []EnvVar      `json:"env,omitempty"`
	VolumeMounts []VolumeMount `json:"volumeMounts,omitempty"`
}

type EnvVar struct {
	Name      string        `json:"name"`
	Value     string        `json:"value,omitempty"`
	ValueFrom *EnvVarSource `json:"valueFrom,omitempty"`
}

type EnvVarSource struct {
	SecretKeyRef *SecretKeySelector `json:"secretKeyRef,omitempty"`
}

type SecretKeySelector struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
	SubPath   string `json:"subPath,omitempty"`
	ReadOnly  bool   `json:"readOnly,omitempty"`
}

type Volume struct {
	Name     string    `json:"name"`
	EmptyDir *EmptyDir `json:"emptyDir,omitempty"`
}

type EmptyDir struct {
	Medium string `json:"medium,omitempty"`
}

// WebhookServer handles admission mutation requests
type WebhookServer struct {
	DefaultServerURL string
}

func (ws *WebhookServer) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (ws *WebhookServer) HandleMutate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, fmt.Sprintf("could not read body: %v", err), http.StatusBadRequest)
		return
	}

	var admissionReview AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		log.Printf("Error unmarshaling AdmissionReview: %v", err)
		http.Error(w, fmt.Sprintf("could not parse admission review: %v", err), http.StatusBadRequest)
		return
	}

	if admissionReview.Request == nil {
		http.Error(w, "AdmissionReview request is nil", http.StatusBadRequest)
		return
	}

	var pod Pod
	if err := json.Unmarshal(admissionReview.Request.Object, &pod); err != nil {
		log.Printf("Error unmarshaling Pod from request: %v", err)
		http.Error(w, fmt.Sprintf("could not unmarshal pod: %v", err), http.StatusBadRequest)
		return
	}

	admissionResponse := ws.mutatePod(&pod, admissionReview.Request.UID)
	admissionReview.Response = admissionResponse

	respBytes, err := json.Marshal(admissionReview)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		http.Error(w, fmt.Sprintf("could not marshal response: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(respBytes)
}

func (ws *WebhookServer) mutatePod(pod *Pod, uid string) *AdmissionResponse {
	annotations := pod.Metadata.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}

	// Check if injection is requested
	injectVal := strings.ToLower(annotations[AnnotationInject])
	if injectVal != "true" && injectVal != "enabled" && injectVal != "yes" {
		return &AdmissionResponse{
			UID:     uid,
			Allowed: true,
		}
	}

	// Prevent duplicate mutation
	if annotations[AnnotationStatus] == "injected" {
		return &AdmissionResponse{
			UID:     uid,
			Allowed: true,
		}
	}

	// Extract options with sensible defaults
	project := annotations[AnnotationProject]
	env := annotations[AnnotationEnv]
	if env == "" {
		env = "prod"
	}

	serverURL := annotations[AnnotationServerURL]
	if serverURL == "" {
		serverURL = ws.DefaultServerURL
		if serverURL == "" {
			serverURL = DefaultServerURL
		}
	}

	secretName := annotations[AnnotationSecretName]
	if secretName == "" {
		secretName = DefaultSecretName
	}

	secretKey := annotations[AnnotationSecretKey]
	if secretKey == "" {
		secretKey = DefaultSecretKey
	}

	mountPath := annotations[AnnotationMountPath]
	if mountPath == "" {
		mountPath = DefaultMountPath
	}

	fileName := annotations[AnnotationFileName]
	if fileName == "" {
		fileName = DefaultFileName
	}

	initImage := annotations[AnnotationInitImage]
	if initImage == "" {
		initImage = DefaultInitImage
	}

	log.Printf("Injecting secrets into pod %s/%s (Project: %s, Env: %s)",
		pod.Metadata.Namespace, pod.Metadata.Name, project, env)

	var patches []PatchOperation

	// 1. Add in-memory volume for secrets (tmpfs)
	secretVolume := Volume{
		Name: VolumeName,
		EmptyDir: &EmptyDir{
			Medium: "Memory",
		},
	}
	if len(pod.Spec.Volumes) == 0 {
		patches = append(patches, PatchOperation{
			Op:    "add",
			Path:  "/spec/volumes",
			Value: []Volume{secretVolume},
		})
	} else {
		patches = append(patches, PatchOperation{
			Op:    "add",
			Path:  "/spec/volumes/-",
			Value: secretVolume,
		})
	}

	// 2. Build the fetch script for the init container
	// Query ManUp REST endpoint directly: /api/secrets?project=...&env=...&format=dotenv
	url := fmt.Sprintf("%s/api/secrets?project=%s&env=%s&format=dotenv", serverURL, project, env)
	fetchScript := fmt.Sprintf(
		"echo '[manup-init] Fetching secrets for project=%s env=%s...'; "+
			"curl -sSf -H \"Authorization: Bearer $MANUP_API_KEY\" %q -o /manup-secrets/%s && "+
			"chmod 0600 /manup-secrets/%s && "+
			"echo '[manup-init] Secrets successfully injected to /manup-secrets/%s'",
		project, env, url, fileName, fileName, fileName,
	)

	initContainer := Container{
		Name:    "manup-secrets-init",
		Image:   initImage,
		Command: []string{"/bin/sh", "-c"},
		Args:    []string{fetchScript},
		Env: []EnvVar{
			{
				Name: "MANUP_API_KEY",
				ValueFrom: &EnvVarSource{
					SecretKeyRef: &SecretKeySelector{
						Name: secretName,
						Key:  secretKey,
					},
				},
			},
		},
		VolumeMounts: []VolumeMount{
			{
				Name:      VolumeName,
				MountPath: "/manup-secrets",
			},
		},
	}

	if len(pod.Spec.InitContainers) == 0 {
		patches = append(patches, PatchOperation{
			Op:    "add",
			Path:  "/spec/initContainers",
			Value: []Container{initContainer},
		})
	} else {
		patches = append(patches, PatchOperation{
			Op:    "add",
			Path:  "/spec/initContainers/-",
			Value: initContainer,
		})
	}

	// 3. Mount volume into target container(s)
	appVolumeMount := VolumeMount{
		Name:      VolumeName,
		MountPath: mountPath,
		ReadOnly:  true,
	}

	for i, container := range pod.Spec.Containers {
		if len(container.VolumeMounts) == 0 {
			patches = append(patches, PatchOperation{
				Op:    "add",
				Path:  fmt.Sprintf("/spec/containers/%d/volumeMounts", i),
				Value: []VolumeMount{appVolumeMount},
			})
		} else {
			patches = append(patches, PatchOperation{
				Op:    "add",
				Path:  fmt.Sprintf("/spec/containers/%d/volumeMounts/-", i),
				Value: appVolumeMount,
			})
		}
	}

	// 4. Update status annotation
	if pod.Metadata.Annotations == nil {
		patches = append(patches, PatchOperation{
			Op:   "add",
			Path: "/metadata/annotations",
			Value: map[string]string{
				AnnotationStatus: "injected",
			},
		})
	} else {
		// Escape JSON pointer key
		escapedKey := strings.ReplaceAll(AnnotationStatus, "/", "~1")
		patches = append(patches, PatchOperation{
			Op:    "add",
			Path:  fmt.Sprintf("/metadata/annotations/%s", escapedKey),
			Value: "injected",
		})
	}

	patchBytes, err := json.Marshal(patches)
	if err != nil {
		log.Printf("Error marshaling patch: %v", err)
		return &AdmissionResponse{
			UID:     uid,
			Allowed: false,
			Result: &metav1Status{
				Message: fmt.Sprintf("could not marshal JSON patch: %v", err),
			},
		}
	}

	patchType := "JSONPatch"
	return &AdmissionResponse{
		UID:       uid,
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: patchType,
	}
}
