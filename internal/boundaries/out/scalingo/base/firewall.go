package scalingo

import (
	"context"

	errors "github.com/Scalingo/go-utils/errors/v3"
	"github.com/Scalingo/scalingo-operator/internal/domain"

	scalingoapi "github.com/Scalingo/go-scalingo/v9"
)

func (c *client) CreateFirewallRule(ctx context.Context, dbID, addonID string, rule domain.FirewallRule) error {
	ruleType, err := toScalingoFirewallRuleType(ctx, rule.Type)
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
		rule, err := toFirewallRule(ctx, scalingoRule)
		if err != nil {
			return []domain.FirewallRule{}, errors.Wrap(ctx, err, "to firewall rule")
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func toFirewallRule(ctx context.Context, rule scalingoapi.FirewallRule) (domain.FirewallRule, error) {
	ruleType, err := toFirewallRuleType(ctx, rule.Type)
	if err != nil {
		return domain.FirewallRule{}, errors.Wrap(ctx, err, "to firewall rule type")
	}
	return domain.FirewallRule{
		ID:      rule.ID,
		Type:    ruleType,
		CIDR:    rule.CIDR,
		Label:   rule.Label,
		RangeID: rule.RangeID,
	}, nil
}

func toFirewallRuleType(ctx context.Context, ruleType scalingoapi.FirewallRuleType) (domain.FirewallRuleType, error) {
	switch ruleType {
	case scalingoapi.FirewallRuleTypeManagedRange:
		return domain.FirewallRuleTypeManagedRange, nil
	case scalingoapi.FirewallRuleTypeCustomRange:
		return domain.FirewallRuleTypeCustomRange, nil
	default:
		return domain.FirewallRuleType(""), errors.Newf(ctx, "invalid type %v", ruleType)
	}
}

func toScalingoFirewallRuleType(ctx context.Context, ruleType domain.FirewallRuleType) (scalingoapi.FirewallRuleType, error) {
	switch ruleType {
	case domain.FirewallRuleTypeManagedRange:
		return scalingoapi.FirewallRuleTypeManagedRange, nil
	case domain.FirewallRuleTypeCustomRange:
		return scalingoapi.FirewallRuleTypeCustomRange, nil
	default:
		return scalingoapi.FirewallRuleType(""), errors.Newf(ctx, "invalid type %v", ruleType)
	}
}
