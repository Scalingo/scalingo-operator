package domain

type Database struct {
	ID        string
	AppID     string
	AddonID   string
	Name      string
	Type      DatabaseType
	Status    DatabaseStatus
	Plan      string
	ProjectID string

	FireWallRules []FirewallRule
}
