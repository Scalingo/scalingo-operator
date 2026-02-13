package scalingo

import (
	"context"

	scalingoapi "github.com/Scalingo/go-scalingo/v9"
	errors "github.com/Scalingo/go-utils/errors/v3"

	"github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo/base/adapters"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func (c *client) CreateFirewallRule(ctx context.Context, dbID, addonID string, rule domain.FirewallRule) error {
	ruleType, err := adapters.ToScalingoFirewallRuleType(ctx, rule.Type)
	if err != nil {
		return errors.Wrap(ctx, err, "to scalingo firewall rule type")
	}

	_, err = c.scClient.Preview().FirewallRulesCreate(ctx, dbID, addonID, scalingoapi.FirewallRuleCreateParams{
		Type:    ruleType,
		CIDR:    rule.CIDR,
		Label:   rule.Label,
		RangeID: rule.RangeID,
	})
	if err != nil {
		return errors.Wrap(ctx, err, "create firewall rule")
	}
	return nil
}

func (c *client) ListFirewallRules(ctx context.Context, dbID, addonID string) ([]domain.FirewallRule, error) {
	scalingoRules, err := c.scClient.Preview().FirewallRulesList(ctx, dbID, addonID)
	if err != nil {
		return []domain.FirewallRule{}, errors.Wrap(ctx, err, "list firewall rules")
	}

	rules := make([]domain.FirewallRule, 0, len(scalingoRules))
	for _, scalingoRule := range scalingoRules {
		rule, err := adapters.ToFirewallRule(ctx, scalingoRule)
		if err != nil {
			return []domain.FirewallRule{}, errors.Wrap(ctx, err, "to firewall rule")
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (c *client) DeleteFirewallRule(ctx context.Context, dbID, addonID, firewallRuleID string) error {
	err := c.scClient.Preview().FirewallRulesDestroy(ctx, dbID, addonID, firewallRuleID)
	if err != nil {
		return errors.Wrap(ctx, err, "delete firewall rule")
	}
	return nil
}
