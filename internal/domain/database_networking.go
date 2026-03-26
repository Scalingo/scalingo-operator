package domain

type DatabaseNetworkConfiguration struct {
	OutscaleAccountID string
	OutscaleNetID     string
	IPRange           string
}

type DatabaseNetPeering struct {
	ID                   string
	OutscaleNetPeeringID string
}
