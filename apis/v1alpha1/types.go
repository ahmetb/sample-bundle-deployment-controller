// Copyright 2022 Ahmet Alp Balkan
//
// +kubebuilder:object:generate=true
// +groupName=ahmet.dev
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "ahmet.dev", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

type BundleDeploymentSpec struct {
	Replicas      int    `json:"replicas,omitempty"`
	ReleaseBundle string `json:"releaseBundle,omitempty"`
}

type BundleDeploymentStatus struct {
	ObservedGeneration         int64  `json:"observedGeneration,omitempty"`
	LastAppliedReleaseChecksum string `json:"lastAppliedReleaseChecksum,omitempty"`
	Ready                      bool   `json:"ready,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="BUNDLE",type="string",JSONPath=".spec.releaseBundle"
// +kubebuilder:printcolumn:name="GENERATION",type="integer",JSONPath=".metadata.generation"
// +kubebuilder:printcolumn:name="OBSERVED_GEN",type="integer",JSONPath=".status.observedGeneration"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

type BundleDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BundleDeploymentSpec   `json:"spec,omitempty"`
	Status BundleDeploymentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type BundleDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BundleDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BundleDeployment{}, &BundleDeploymentList{})
}
