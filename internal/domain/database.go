package domain

type Database struct {
	ID        string
	Name      string
	Type      DatabaseType
	Status    AddonStatus
	Plan      string
	ProjectID string
}
