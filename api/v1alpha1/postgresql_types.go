/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PostgreSQLSpec defines the desired state of PostgreSQL
type PostgreSQLSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// Auth contains the references to the authentication details needed to connect to Scalingo.
	// +kubebuilder:validation:Required
	AuthSecret AuthSecretSpec `json:"authSecret"`

	ConnInfoSecretTarget SecretTargetSpec `json:"connInfoSecretTarget"`

	// Name is the name of the PostgreSQL database to create on Scalingo.
	// +kubebuilder:validation:MinLength=5
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Plan is the plan to use for the PostgreSQL database.
	// +kubebuilder:validation:MinLength=10
	// +kubebuilder:validation:Required
	Plan string `json:"plan"`

	// Region is the Scalingo region where the PostgreSQL database will be created.
	// +kubebuilder:default="osc-fr1"
	Region string `json:"region"`

	// ProjectID is the Scalingo project ID where the PostgreSQL database will be created.
	// If not specified, the default project associated with the authentication token will be used.
	ProjectID string `json:"projectID,omitempty"`
}

// PostgreSQLStatus defines the observed state of PostgreSQL.
type PostgreSQLStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the PostgreSQL resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ScalingoDatabaseID is the unique identifier of the PostgreSQL database on Scalingo.
	ScalingoDatabaseID string `json:"scalingoDatabaseID,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PostgreSQL is the Schema for the postgresqls API
type PostgreSQL struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of PostgreSQL
	// +required
	Spec PostgreSQLSpec `json:"spec"`

	// status defines the observed state of PostgreSQL
	// +optional
	Status PostgreSQLStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// PostgreSQLList contains a list of PostgreSQL
type PostgreSQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgreSQL `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PostgreSQL{}, &PostgreSQLList{})
}
