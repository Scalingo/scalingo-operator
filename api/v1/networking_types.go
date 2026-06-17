package v1

type NetworkingSpec struct {
	// IPRange is the private network range to use when creating the database.
	// +kubebuilder:validation:Pattern=`^([0-9]{1,3}\.){3}[0-9]{1,3}\/[0-9]{1,2}$`
	// +optional
	IPRange string `json:"ip_range,omitempty"`

	// InternetAccess defines the external access through internet.
	// +kubebuilder:validation:Required
	InternetAccess InternetAccessSpec `json:"internet_access"`

	// Outscale defines the Outscale networking configuration.
	// +optional
	Outscale *OutscaleSpec `json:"outscale,omitempty"`

	// Firewall defines the firewall rules.
	// +optional
	Firewall *FirewallSpec `json:"firewall,omitempty"`
}

func (s NetworkingSpec) IsOutscaleOKSNetPeeringEnabled() bool {
	return s.Outscale != nil && s.Outscale.OKS != nil && s.Outscale.OKS.NetPeering
}

type OutscaleSpec struct {
	// OKS defines the Outscale Kubernetes Service networking configuration.
	// +optional
	OKS *OutscaleOKSSpec `json:"oks,omitempty"`
}

type OutscaleOKSSpec struct {
	// NetPeering enables Outscale Net Peering management for the database.
	// +optional
	NetPeering bool `json:"net_peering,omitempty"`
}

type InternetAccessSpec struct {
	// Enabled enables external access.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=true
	// +kubebuilder:default=true
	Enabled bool `json:"enabled"`
}

type FirewallSpec struct {
	// Rules is a list of firewall rules to be applied.
	// +kubebuilder:validation:MinItems=1
	Rules []FirewallRuleSpec `json:"rules"`
}

type FirewallRuleSpec struct {
	// Type of the firewall rule: custom or managed range.
	// +kubebuilder:validation:Enum=custom_range;managed_range
	// +kubebuilder:validation:Required
	Type string `json:"type"`

	// CIDR is the IP range in CIDR notation for "custom_range" type rules.
	// +kubebuilder:validation:Pattern=`^([0-9]{1,3}\.){3}[0-9]{1,3}\/[0-9]{1,2}$`
	// +optional
	CIDR string `json:"cidr,omitempty"`

	// Label is an optional label for the firewall rule.
	// +kubebuilder:validation:MinLength=5
	// +optional
	Label string `json:"label,omitempty"`

	// RangeID is the identifier of the managed range for "managed_range" type rules.
	// +kubebuilder:validation:MinLength=5
	// +optional
	RangeID string `json:"range_id,omitempty"`
}
