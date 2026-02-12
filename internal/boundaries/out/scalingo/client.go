package scalingo

import (
	"context"

	"github.com/Scalingo/scalingo-operator/internal/domain"
)

// Wrapper for go-scalingo client.
type Client interface {
	// Database.
	CreateDatabase(ctx context.Context, db domain.Database) (domain.Database, error)
	GetDatabase(ctx context.Context, dbID string) (domain.Database, error)
	UpdateDatabasePlan(ctx context.Context, db domain.Database) error
	DeleteDatabase(ctx context.Context, dbID string) error

	// Firewall.
	CreateFirewallRule(ctx context.Context, dbID, addonID string, rule domain.FirewallRule) error
	ListFirewallRules(ctx context.Context, dbID, addonID string) ([]domain.FirewallRule, error)
	DeleteFirewallRule(ctx context.Context, dbID, addonID, firewallRuleID string) error

	// Application.
	FindApplicationVariable(ctx context.Context, appID, varName string) (string, error)
}
