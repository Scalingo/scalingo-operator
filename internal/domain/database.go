package domain

type Database struct {
	ID        string
	Name      string
	Type      DatabaseType
	Status    DatabaseStatus
	Plan      string
	ProjectID string
}
