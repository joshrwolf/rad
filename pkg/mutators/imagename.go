package mutators

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	v1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	podResource           = metav1.GroupVersionResource{Version: "v1", Resource: "pods"}
	universalDeserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
)

type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func PrependRegistry(req *v1beta1.AdmissionRequest) ([]PatchOperation, error) {
	log.Info("PrependRegistry")
	// Bail out if for whatever reason this gets called on a non pod object
	if req.Resource != podResource {
		log.Printf("PrependRegistry() cannot be called on a non pod resource, expected type: %s", podResource)
		return nil, nil
	}

	// Parse the pod object
	raw := req.Object.Raw
	pod := corev1.Pod{}
	if _, _, err := universalDeserializer.Decode(raw, nil, &pod); err != nil {
		return nil, fmt.Errorf("failed to deserialize pod object: %v", err)
	}

	// Create the patch object(s)
	var patches []PatchOperation
	for i, c := range pod.Spec.Containers {
		patch := PatchOperation{
			Op:    "replace",
			Path:  fmt.Sprintf("/spec/containers/%d/image", i),
			Value: prepend(c.Image),
		}
		patches = append(patches, patch)
	}

	return patches, nil
}

func prepend(imageName string) string {
	// Check if environment variable exists first
	registryToPrepend := os.Getenv("PREPEND_REGISTRY")

	if registryToPrepend != "" {
		return fmt.Sprintf("%s/%s", registryToPrepend, imageName)
	}

	// Just cat together image name and tag
	return imageName
}
