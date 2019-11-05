package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/joshrwolf/rad/pkg/mutators"

	log "github.com/sirupsen/logrus"
	v1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	universalDeserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
)

type admitFunc func(*v1beta1.AdmissionRequest) ([]mutators.PatchOperation, error)

// AdmitFuncHandler wraps TODO
func AdmitFuncHandler(admit admitFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveAdmitFunc(w, r, admit)
	})
}

// doServeAdmitFunc parses an HTTP request for an admission controller webhook and validates it
// upon a successfully validated request, it passes the request along to admitFunc to recieve a response body
func doServeAdmitFunc(w http.ResponseWriter, r *http.Request, admit admitFunc) ([]byte, error) {
	// Do some request validation, only make sure we respond to POST requests and json content type
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil, fmt.Errorf("invalid method %s, only POST methods are allowed", r.Method)
	}

	// Read and validate the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("failed to read the response body: %v", err)
	}

	// Make sure it's a json body type
	if contentType := r.Header.Get("Content-Type"); contentType != `application/json` {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("content type %s is not supported, only supports application/json", contentType)
	}

	// Parse the AdmissionReview request
	admissionReviewReq := v1beta1.AdmissionReview{}
	if _, _, err := universalDeserializer.Decode(body, nil, &admissionReviewReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("could not deserialize request: %v", err)
	} else if admissionReviewReq.Request == nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, errors.New("admission review request is nil")
	}

	// Build an AdmissionReview response
	admissionReviewResponse := v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID: admissionReviewReq.Request.UID,

			// Add any extra annotations for debugging or just general explicitness
			AuditAnnotations: map[string]string{
				"allasso": "image has been mutated",
			},
		},
	}

	var patchOps []mutators.PatchOperation
	// Apply admit() to get patch operations
	patchOps, err = admit(admissionReviewReq.Request)

	if err != nil {
		admissionReviewResponse.Response.Allowed = false
		admissionReviewResponse.Response.Result = &metav1.Status{
			Message: err.Error(),
		}
	} else {
		// return a positive response
		patchBytes, err := json.Marshal(patchOps)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return nil, fmt.Errorf("failed to marshal JSON patch objects: %v", err)
		}
		admissionReviewResponse.Response.Allowed = true
		admissionReviewResponse.Response.Patch = patchBytes
	}

	// FINALLY, return a valid patch response
	bytes, err := json.Marshal(&admissionReviewResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %v", err)
	}
	return bytes, nil
}

func serveAdmitFunc(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	log.Printf("Handling webhook request...")

	var writeErr error
	if bytes, err := doServeAdmitFunc(w, r, admit); err != nil {
		log.Printf("Error handling webhook request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, writeErr = w.Write([]byte(err.Error()))
	} else {
		log.Print("Successfully handled webhook request")
		_, writeErr = w.Write(bytes)
	}

	if writeErr != nil {
		log.Printf("Failed to write response: %v", writeErr)
	}
}
