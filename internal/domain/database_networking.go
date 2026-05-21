package domain

type DatabaseEndpointType string

const (
	DatabaseEndpointTypePublicRW         DatabaseEndpointType = "public-rw"
	DatabaseEndpointTypePrivatePeeringRW DatabaseEndpointType = "private-peering-rw"
)

type DatabaseEndpoint struct {
	ID         string
	DatabaseID string
	Hostname   string
	Port       int
	Type       DatabaseEndpointType
}

type DatabaseNetworkConfiguration struct {
	OutscaleAccountID string
	OutscaleNetID     string
	IPRange           string
}

type DatabaseNetPeering struct {
	ID                   string
	OutscaleNetPeeringID string
}
