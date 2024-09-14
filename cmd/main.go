package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func (p *PatchOperation) String() string {
	return fmt.Sprintf("op: %s, path: %s, value: %s", p.Op, p.Path, p.Value)
}

func main() {

	r := gin.Default()

	r.POST("/mutate", func(c *gin.Context) {
		var admissionReview admissionv1.AdmissionReview
		if err := c.ShouldBindJSON(&admissionReview); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		admissionResponse := handleMutation(admissionReview)

		admissionReview.Response = &admissionResponse
		c.JSON(http.StatusOK, admissionReview)
	})

	fmt.Println("Starting webhook server on port 8443...")
	if err := r.RunTLS(":8443", "/certs/tls.crt", "/certs/tls.key"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}

func handleMutation(ar admissionv1.AdmissionReview) admissionv1.AdmissionResponse {
	pod := corev1.Pod{}

	if err := json.Unmarshal(ar.Request.Object.Raw, &pod); err != nil {
		return admissionv1.AdmissionResponse{
			UID:     ar.Request.UID,
			Allowed: false,
			Result:  &metav1.Status{Message: err.Error()},
		}
	}

	patchBytesV2, err := createPodsPathchOps(&pod)
	if err != nil {
		return admissionv1.AdmissionResponse{
			UID:     ar.Request.UID,
			Allowed: false,
			Result:  &metav1.Status{Message: err.Error()},
		}
	}

	return admissionv1.AdmissionResponse{
		UID:     ar.Request.UID,
		Allowed: true,
		Patch:   patchBytesV2,
		PatchType: func() *admissionv1.PatchType {
			pt := admissionv1.PatchTypeJSONPatch
			return &pt
		}(),
	}
}

func createPodsPathchOps(pod *corev1.Pod) ([]byte, error) {
	var patchOps []PatchOperation

	for i := range pod.Spec.Containers {
		patchOps = append(patchOps, PatchOperation{
			Op:    "replace",
			Path:  fmt.Sprintf("/spec/containers/%d/imagePullPolicy", i),
			Value: "Always",
		})
	}

	for i := range pod.Spec.InitContainers {
		patchOps = append(patchOps, PatchOperation{
			Op:    "replace",
			Path:  fmt.Sprintf("/spec/initContainers/%d/imagePullPolicy", i),
			Value: "Always",
		})
	}

	for _, ops := range patchOps {
		fmt.Println(ops.String())
	}

	return json.Marshal(patchOps)
}
