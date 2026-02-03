package domain

const (
	DatabaseFeatureForceSsl          = "force-ssl"
	DatabaseFeaturePubliclyAvailable = "publicly-available"
)

type DatabaseFeatures map[string]DatabaseFeatureStatus

type Database struct {
	ID        string
	AppID     string
	AddonID   string
	Name      string
	Type      DatabaseType
	Status    DatabaseStatus
	Plan      string
	ProjectID string

	Features      DatabaseFeatures
	FireWallRules []FirewallRule
}
