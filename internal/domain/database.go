package domain

const (
	// Databases finalizers names.
	OpenSearchFinalizerName = "databases.scalingo.com/OpenSearchFinalizer"
	PostgreSQLFinalizerName = "databases.scalingo.com/PostgresFinalizer"
)

type Database struct {
	ID        string
	Name      string
	Type      DatabaseType
	Status    AddonStatus
	Plan      string
	ProjectID string
}
