// File: api/v1alpha1/microfrontend_types.go
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime"
)

// MicroFrontendSpec defines the desired state of MicroFrontend
type MicroFrontendSpec struct {
	OCIArtifact string   `json:"ociArtifact"`
	CDNTarget   string   `json:"cdnTarget"`
	EntryPoint  string   `json:"entryPoint"`
	ExposedModules []string `json:"exposedModules"`
}

// MicroFrontendStatus defines the observed state of MicroFrontend
type MicroFrontendStatus struct {
	Synced       bool   `json:"synced"`
	LastSyncedAt string `json:"lastSyncedAt,omitempty"`
	Message      string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MicroFrontend is the Schema for the microfrontends API
type MicroFrontend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MicroFrontendSpec   `json:"spec,omitempty"`
	Status MicroFrontendStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MicroFrontendList contains a list of MicroFrontend
type MicroFrontendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MicroFrontend `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MicroFrontend{}, &MicroFrontendList{})
}

func (in *MicroFrontend) GetObjectKind() schema.ObjectKind {
	return &in.TypeMeta
}

func (in *MicroFrontend) DeepCopyObject() runtime.Object {
	return in.DeepCopy()
}
