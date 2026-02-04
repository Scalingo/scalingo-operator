package adapters

import (
	"context"

	"github.com/Scalingo/go-utils/errors/v3"
	"github.com/Scalingo/scalingo-operator/api/v1alpha1"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func toFirewallRules(ctx context.Context, networkSpec v1alpha1.NetworkingSpec) ([]domain.FirewallRule, error) {
	if networkSpec.Firewall == nil || len(networkSpec.Firewall.Rules) == 0 {
		return nil, nil
	}

	newRules := make([]domain.FirewallRule, 0, len(networkSpec.Firewall.Rules))
	for _, rule := range networkSpec.Firewall.Rules {
		newRule, err := toFirewallRule(ctx, rule)
		if err != nil {
			return nil, err
		}
		newRules = append(newRules, newRule)
	}
	return newRules, nil
}

func toFirewallRule(ctx context.Context, rule v1alpha1.FirewallRuleSpec) (domain.FirewallRule, error) {
	newRule := domain.FirewallRule{
		Type:    domain.FirewallRuleType(rule.Type),
		CIDR:    rule.CIDR,
		Label:   rule.Label,
		RangeID: rule.RangeID,
	}

	err := newRule.Validate()
	if err != nil {
		return domain.FirewallRule{}, errors.Wrap(ctx, err, "validate firewall rule")
	}
	return newRule, nil
}
