package scalingo

import (
	"context"

	errors "github.com/Scalingo/go-utils/errors/v3"
	"github.com/Scalingo/scalingo-operator/internal/domain"

	scalingoapi "github.com/Scalingo/go-scalingo/v9"
)

func (c *client) CreateFirewallRule(ctx context.Context, appID, addonID string, rule domain.FirewallRule) error {
	_, err := c.scClient.Preview().FirewallRulesCreate(ctx, appID, addonID, scalingoapi.FirewallRuleCreateParams{
		Type:    toScalingoFirewallRuleType(rule.Type),
		CIDR:    rule.CIDR,
		Label:   rule.Label,
		RangeID: rule.RangeID,
	})
	if err != nil {
		return errors.Wrap(ctx, err, "create firewall rule")
	}
	return nil
}

func (c *client) ListFirewallRules(ctx context.Context, appID, addonID string) ([]domain.FirewallRule, error) {
	scalingoRules, err := c.scClient.Preview().FirewallRulesList(ctx, appID, addonID)
	if err != nil {
		return []domain.FirewallRule{}, errors.Wrap(ctx, err, "list firewall rules")
	}

	rules := make([]domain.FirewallRule, 0, len(scalingoRules))
	for _, scalingoRule := range scalingoRules {
		rules = append(rules, toFirewallRule(scalingoRule))
	}
	return rules, nil
}

func (c *client) DeleteFirewallRule(ctx context.Context, appID, addonID, firewallRuleID string) error {
	err := c.scClient.Preview().FirewallRulesDestroy(ctx, appID, addonID, firewallRuleID)
	if err != nil {
		return errors.Wrap(ctx, err, "delete firewall rule")
	}
	return nil
}

func toFirewallRule(rule scalingoapi.FirewallRule) domain.FirewallRule {
	return domain.FirewallRule{
		ID:      rule.ID,
		Type:    toFirewallRuleType(rule.Type),
		CIDR:    rule.CIDR,
		Label:   rule.Label,
		RangeID: rule.RangeID,
	}
}

func toFirewallRuleType(ruleType scalingoapi.FirewallRuleType) domain.FirewallRuleType {
	switch ruleType {
	case scalingoapi.FirewallRuleTypeManagedRange:
		return domain.FirewallRuleTypeManagedRange
	default:
		return domain.FirewallRuleTypeCustomRange
	}
}

func toScalingoFirewallRuleType(ruleType domain.FirewallRuleType) scalingoapi.FirewallRuleType {
	switch ruleType {
	case domain.FirewallRuleTypeManagedRange:
		return scalingoapi.FirewallRuleTypeManagedRange
	default:
		return scalingoapi.FirewallRuleTypeCustomRange
	}
}
