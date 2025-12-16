package v1alpha1

type SecretTargetSpec struct {
	// The name of the secret to create or update with the connection information.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Prefix for the secret keys.
	// Added as is without any transformation.
	// By default, the prefix uses the following format: "SCALINGO_<DB_TYPE>_"
	Prefix string `json:"prefix,omitempty"`
}
