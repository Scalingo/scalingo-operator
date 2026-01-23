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

	// TODO: apply a diff to gather: 1/ rules to delete, 2/ rules to add, then apply these diff rules.
	// 			-> this will ensure firewall rules application idempotency.
	//
	// Actually, rules are added only once, right after DB creation.

	if len(currentDB.FireWallRules) != 0 {
		log.Info("Firewall rules already created, nothing to do")
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, rule := range expectedRules {
		g.Go(func() error {
			err := m.scClient.CreateFirewallRule(ctx, currentDB.ID, currentDB.AddonID, rule)
			if err == nil {
				log.Info("Add firewall rule", "rule", rule)
			} else {
				log.Error(err, "Fail to add firewall rule", "AppID", currentDB.ID, "AddonID", currentDB.AddonID, "rule", rule)
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
