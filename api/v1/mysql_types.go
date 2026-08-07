/*
Copyright 2026.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MySQLSpec defines the desired state of MySQL.
type MySQLSpec struct {
	// AuthSecret contains the references to the authentication details needed to connect to Scalingo.
	// +kubebuilder:validation:Required
	AuthSecret AuthSecretSpec `json:"authSecret"`

	// ConnInfoSecretTarget defines where to store the connection information secret.
	// +kubebuilder:validation:Required
	ConnInfoSecretTarget SecretTargetSpec `json:"connInfoSecretTarget"`

	// Networking defines the networking configuration.
	// +kubebuilder:validation:Required
	Networking NetworkingSpec `json:"networking"`

	// Name is the name of the MySQL database to create on Scalingo. Falls back to metadata.name if empty.
	// +kubebuilder:validation:MinLength=5
	// +optional
	Name string `json:"name,omitempty"`

	// Plan is the plan to use for the MySQL database.
	// +kubebuilder:validation:MinLength=10
	// +kubebuilder:validation:Required
	Plan string `json:"plan"`

	// Region is the Scalingo region where the MySQL database will be created.
	// +kubebuilder:default="osc-fr1"
	// +kubebuilder:validation:MinLength=5
	Region string `json:"region"`

	// ProjectID is the Scalingo project ID where the MySQL database will be created.
	// If not specified, the default project associated with the authentication token will be used.
	// +optional
	ProjectID string `json:"projectID,omitempty"`
}

// MySQLStatus defines the observed state of MySQL.
type MySQLStatus struct {
	// conditions represent the current state of the MySQL resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ScalingoDatabaseID is the unique identifier of the MySQL database on Scalingo.
	ScalingoDatabaseID string `json:"scalingoDatabaseID,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MySQL is the Schema for the mysqls API.
type MySQL struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of MySQL
	// +required
	Spec MySQLSpec `json:"spec"`

	// status defines the observed state of MySQL
	// +optional
	Status MySQLStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// MySQLList contains a list of MySQL.
type MySQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []MySQL `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MySQL{}, &MySQLList{})
}
