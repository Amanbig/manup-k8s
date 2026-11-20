package main

import (
	"encoding/json"
	"testing"
)

func TestMutatePodInjection(t *testing.T) {
	ws := &WebhookServer{
		DefaultServerURL: "http://manup.manup.svc:7780",
	}

	pod := &Pod{
		Metadata: ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			Annotations: map[string]string{
				AnnotationInject:  "true",
				AnnotationProject: "billing",
				AnnotationEnv:     "production",
			},
		},
		Spec: PodSpec{
			Containers: []Container{
				{
					Name:  "app",
					Image: "my-app:1.0",
				},
			},
		},
	}

	resp := ws.mutatePod(pod, "test-uid-12345")
	if !resp.Allowed {
		t.Fatalf("Expected pod to be allowed, but was rejected")
	}

	if resp.PatchType != "JSONPatch" {
		t.Fatalf("Expected PatchType 'JSONPatch', got %s", resp.PatchType)
	}

	if len(resp.Patch) == 0 {
		t.Fatalf("Expected non-empty patch bytes")
	}

	var patches []PatchOperation
	if err := json.Unmarshal(resp.Patch, &patches); err != nil {
		t.Fatalf("Failed to parse JSON patches: %v", err)
	}

	// Verify we got the expected patches: volume, initContainer, volumeMount, annotation
	var hasVolume, hasInitContainer, hasVolumeMount, hasAnnotation bool
	for _, p := range patches {
		if p.Path == "/spec/volumes" || p.Path == "/spec/volumes/-" {
			hasVolume = true
		}
		if p.Path == "/spec/initContainers" || p.Path == "/spec/initContainers/-" {
			hasInitContainer = true
		}
		if p.Path == "/spec/containers/0/volumeMounts" || p.Path == "/spec/containers/0/volumeMounts/-" {
			hasVolumeMount = true
		}
		if p.Path == "/metadata/annotations/manup.io~1status" {
			hasAnnotation = true
		}
	}

	if !hasVolume {
		t.Errorf("Patch missing /spec/volumes")
	}
	if !hasInitContainer {
		t.Errorf("Patch missing /spec/initContainers")
	}
	if !hasVolumeMount {
		t.Errorf("Patch missing /spec/containers/0/volumeMounts")
	}
	if !hasAnnotation {
		t.Errorf("Patch missing /metadata/annotations/manup.io~1status")
	}
}

func TestNoInjectionWhenAnnotationMissing(t *testing.T) {
	ws := &WebhookServer{
		DefaultServerURL: "http://manup.manup.svc:7780",
	}

	pod := &Pod{
		Metadata: ObjectMeta{
			Name:      "plain-pod",
			Namespace: "default",
			Annotations: map[string]string{
				"some.other/annotation": "foo",
			},
		},
		Spec: PodSpec{
			Containers: []Container{
				{
					Name:  "app",
					Image: "nginx:latest",
				},
			},
		},
	}

	resp := ws.mutatePod(pod, "plain-uid")
	if !resp.Allowed {
		t.Fatalf("Expected allowed")
	}
	if len(resp.Patch) != 0 {
		t.Fatalf("Expected no patch for unannotated pod, got %s", string(resp.Patch))
	}
}

func TestTLSGeneration(t *testing.T) {
	certPair, err := GenerateSelfSignedCert("manup-injector", "manup-system")
	if err != nil {
		t.Fatalf("GenerateSelfSignedCert failed: %v", err)
	}

	if len(certPair.CertPEM) == 0 || len(certPair.KeyPEM) == 0 || len(certPair.CAPEM) == 0 {
		t.Fatalf("CertPair has empty fields")
	}
}
