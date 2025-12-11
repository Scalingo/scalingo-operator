package goscalingo

import (
	"context"

	"github.com/Scalingo/scalingo-operator/internal/domain"
)

// Wrapper for go-scalingo client.
type Client interface {
	CreateDatabase(ctx context.Context, db domain.Database) (domain.Database, error)
	GetDatabase(ctx context.Context, dbID string) (domain.Database, error)
	UpdateDatabase(ctx context.Context, db domain.Database) (domain.Database, error)
	DeleteDatabase(ctx context.Context, dbID string) error
}
