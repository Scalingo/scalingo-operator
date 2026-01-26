package database

import (
	"context"

	"github.com/Scalingo/scalingo-operator/internal/domain"
)

type Manager interface {
	CreateDatabase(ctx context.Context, db domain.Database) (domain.Database, error)
	GetDatabase(ctx context.Context, dbID string) (domain.Database, error)
	GetDatabaseURL(ctx context.Context, db domain.Database) (domain.DatabaseURL, error)
	UpdateDatabase(ctx context.Context, dbID string, expectedDB domain.Database) error
	DeleteDatabase(ctx context.Context, dbID string) error
}
