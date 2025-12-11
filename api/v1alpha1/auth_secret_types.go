package v1alpha1

type AuthSecretSpec struct {
	// SecretName is the name of the Kubernetes Secret that contains authentication details.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// SecretKey is the key within the Secret that holds the authentication information.
	// If not specified, it defaults to "token".
	// +kubebuilder:default="token"
	Key string `json:"key,omitempty"`
}
