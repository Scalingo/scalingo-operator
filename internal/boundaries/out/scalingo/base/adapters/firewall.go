package adapters

import (
	"context"

	scalingoapi "github.com/Scalingo/go-scalingo/v9"
	errors "github.com/Scalingo/go-utils/errors/v3"

	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func ToFirewallRule(ctx context.Context, rule scalingoapi.FirewallRule) (domain.FirewallRule, error) {
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

func ToScalingoFirewallRuleType(ctx context.Context, ruleType domain.FirewallRuleType) (scalingoapi.FirewallRuleType, error) {
	switch ruleType {
	case domain.FirewallRuleTypeManagedRange:
		return scalingoapi.FirewallRuleTypeManagedRange, nil
	case domain.FirewallRuleTypeCustomRange:
		return scalingoapi.FirewallRuleTypeCustomRange, nil
	default:
		return scalingoapi.FirewallRuleType(""), errors.Newf(ctx, "invalid type %v", ruleType)
	}
}
