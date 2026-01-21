package database

import (
	"context"

	"golang.org/x/sync/errgroup"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	errors "github.com/Scalingo/go-utils/errors/v3"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func (m *manager) updateFirewallRules(ctx context.Context, currentDB domain.Database, expectedRules []domain.FirewallRule) error {
	log := logf.FromContext(ctx)

	nbRules := len(expectedRules)
	if nbRules == 0 {
		log.Info("No firewall rule to create")
		return nil
	}

	// TODO: apply a diff to gather: 1/ rules to delete, 2/ rules to add, 3/ rules to update (label)

	g, ctx := errgroup.WithContext(ctx)
	for _, rule := range expectedRules {
		g.Go(func() error {
			// Remark: DB ID is the app ID for DBNG.
			err := m.scClient.CreateFirewallRule(ctx, currentDB.ID, currentDB.AddonID, rule)
			if err == nil {
				log.Info("Created firewall rule", "rule", rule)
			} else {
				log.Error(err, "Fail to create firewall rule", "AppID", currentDB.ID, "AddonID", currentDB.AddonID, "rule", rule)
			}
			return err
		})
	}

	err := g.Wait()
	if err != nil {
		return errors.Wrap(ctx, err, "update firewall rules")
	}
	return nil
}

func (m *manager) deleteFirewallRules(ctx context.Context, dbID string) error {
	return domain.ErrNotImplemented
}
