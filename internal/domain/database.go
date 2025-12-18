package domain

type Database struct {
	ID        string
	AppID     string
	Name      string
	Type      DatabaseType
	Status    AddonStatus
	Plan      string
	ProjectID string
}
