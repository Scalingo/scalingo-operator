package adapters

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Scalingo/scalingo-operator/api/v1alpha1"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestToFirewallRules(t *testing.T) {
	t.Run("it returns nil when firewall is nil", func(t *testing.T) {
		spec := v1alpha1.NetworkingSpec{
			Firewall: nil,
		}

		rules, err := toFirewallRules(t.Context(), spec)

		require.NoError(t, err)
		require.Nil(t, rules)
	})

	t.Run("it returns nil when rules are empty", func(t *testing.T) {
		spec := v1alpha1.NetworkingSpec{
			Firewall: &v1alpha1.FirewallSpec{
				Rules: []v1alpha1.FirewallRuleSpec{},
			},
		}

		rules, err := toFirewallRules(t.Context(), spec)

		require.NoError(t, err)
		require.Nil(t, rules)
	})

	t.Run("it converts single custom_range rule", func(t *testing.T) {
		spec := v1alpha1.NetworkingSpec{
			Firewall: &v1alpha1.FirewallSpec{
				Rules: []v1alpha1.FirewallRuleSpec{
					{
						Type:  "custom_range",
						CIDR:  "192.168.1.0/24",
						Label: "Test Label",
					},
				},
			},
		}

		rules, err := toFirewallRules(t.Context(), spec)

		require.NoError(t, err)
		require.Len(t, rules, 1)
		require.Equal(t, domain.FirewallRuleTypeCustomRange, rules[0].Type)
		require.Equal(t, "192.168.1.0/24", rules[0].CIDR)
		require.Equal(t, "Test Label", rules[0].Label)
		require.Empty(t, rules[0].RangeID)
	})

	t.Run("it converts single managed_range rule", func(t *testing.T) {
		spec := v1alpha1.NetworkingSpec{
			Firewall: &v1alpha1.FirewallSpec{
				Rules: []v1alpha1.FirewallRuleSpec{
					{
						Type:    "managed_range",
						RangeID: "range-123",
						Label:   "Managed Label",
					},
				},
			},
		}

		rules, err := toFirewallRules(t.Context(), spec)

		require.NoError(t, err)
		require.Len(t, rules, 1)
		require.Equal(t, domain.FirewallRuleTypeManagedRange, rules[0].Type)
		require.Equal(t, "range-123", rules[0].RangeID)
		require.Equal(t, "Managed Label", rules[0].Label)
		require.Empty(t, rules[0].CIDR)
	})

	t.Run("it converts multiple rules", func(t *testing.T) {
		spec := v1alpha1.NetworkingSpec{
			Firewall: &v1alpha1.FirewallSpec{
				Rules: []v1alpha1.FirewallRuleSpec{
					{
						Type:  "custom_range",
						CIDR:  "10.0.0.0/8",
						Label: "Custom 1",
					},
					{
						Type:    "managed_range",
						RangeID: "range-456",
					},
					{
						Type: "custom_range",
						CIDR: "172.16.0.0/12",
					},
				},
			},
		}

		rules, err := toFirewallRules(t.Context(), spec)

		require.NoError(t, err)
		require.Len(t, rules, 3)
		require.Equal(t, domain.FirewallRuleTypeCustomRange, rules[0].Type)
		require.Equal(t, "10.0.0.0/8", rules[0].CIDR)
		require.Equal(t, "Custom 1", rules[0].Label)
		require.Equal(t, domain.FirewallRuleTypeManagedRange, rules[1].Type)
		require.Equal(t, "range-456", rules[1].RangeID)
		require.Equal(t, domain.FirewallRuleTypeCustomRange, rules[2].Type)
		require.Equal(t, "172.16.0.0/12", rules[2].CIDR)
	})

	t.Run("it returns error for invalid rule", func(t *testing.T) {
		spec := v1alpha1.NetworkingSpec{
			Firewall: &v1alpha1.FirewallSpec{
				Rules: []v1alpha1.FirewallRuleSpec{
					{
						Type: "custom_range",
						// Missing CIDR
					},
				},
			},
		}

		rules, err := toFirewallRules(t.Context(), spec)

		require.Error(t, err)
		require.Nil(t, rules)
		require.ErrorContains(t, err, "missing cidr")
	})

	t.Run("it returns error when one of multiple rules is invalid", func(t *testing.T) {
		spec := v1alpha1.NetworkingSpec{
			Firewall: &v1alpha1.FirewallSpec{
				Rules: []v1alpha1.FirewallRuleSpec{
					{
						Type: "custom_range",
						CIDR: "10.0.0.0/8",
					},
					{
						Type: "managed_range",
						// Missing RangeID
					},
				},
			},
		}

		rules, err := toFirewallRules(t.Context(), spec)

		require.Error(t, err)
		require.Nil(t, rules)
		require.ErrorContains(t, err, "missing range_id")
	})
}

func TestToFirewallRule(t *testing.T) {
	t.Run("it converts custom_range rule with all fields", func(t *testing.T) {
		spec := v1alpha1.FirewallRuleSpec{
			Type:  "custom_range",
			CIDR:  "192.168.1.0/24",
			Label: "My Custom Rule",
		}

		rule, err := toFirewallRule(t.Context(), spec)

		require.NoError(t, err)
		require.Equal(t, domain.FirewallRuleTypeCustomRange, rule.Type)
		require.Equal(t, "192.168.1.0/24", rule.CIDR)
		require.Equal(t, "My Custom Rule", rule.Label)
		require.Empty(t, rule.RangeID)
		require.Empty(t, rule.ID)
	})

	t.Run("it converts custom_range rule without label", func(t *testing.T) {
		spec := v1alpha1.FirewallRuleSpec{
			Type: "custom_range",
			CIDR: "10.0.0.0/16",
		}

		rule, err := toFirewallRule(t.Context(), spec)

		require.NoError(t, err)
		require.Equal(t, domain.FirewallRuleTypeCustomRange, rule.Type)
		require.Equal(t, "10.0.0.0/16", rule.CIDR)
		require.Empty(t, rule.Label)
		require.Empty(t, rule.RangeID)
	})

	t.Run("it converts managed_range rule with all fields", func(t *testing.T) {
		spec := v1alpha1.FirewallRuleSpec{
			Type:    "managed_range",
			RangeID: "range-789",
			Label:   "My Managed Rule",
		}

		rule, err := toFirewallRule(t.Context(), spec)

		require.NoError(t, err)
		require.Equal(t, domain.FirewallRuleTypeManagedRange, rule.Type)
		require.Equal(t, "range-789", rule.RangeID)
		require.Equal(t, "My Managed Rule", rule.Label)
		require.Empty(t, rule.CIDR)
		require.Empty(t, rule.ID)
	})

	t.Run("it converts managed_range rule without label", func(t *testing.T) {
		spec := v1alpha1.FirewallRuleSpec{
			Type:    "managed_range",
			RangeID: "range-xyz",
		}

		rule, err := toFirewallRule(t.Context(), spec)

		require.NoError(t, err)
		require.Equal(t, domain.FirewallRuleTypeManagedRange, rule.Type)
		require.Equal(t, "range-xyz", rule.RangeID)
		require.Empty(t, rule.Label)
		require.Empty(t, rule.CIDR)
	})

	t.Run("it returns error for invalid type", func(t *testing.T) {
		spec := v1alpha1.FirewallRuleSpec{
			Type: "invalid_type",
		}

		rule, err := toFirewallRule(t.Context(), spec)

		require.Error(t, err)
		require.ErrorContains(t, err, "invalid firewall rule type")
		require.Equal(t, domain.FirewallRule{}, rule)
	})

	t.Run("it returns error for custom_range without CIDR", func(t *testing.T) {
		spec := v1alpha1.FirewallRuleSpec{
			Type:  "custom_range",
			Label: "Missing CIDR",
		}

		rule, err := toFirewallRule(t.Context(), spec)

		require.Error(t, err)
		require.ErrorContains(t, err, "missing cidr")
		require.Equal(t, domain.FirewallRule{}, rule)
	})

	t.Run("it returns error for managed_range without RangeID", func(t *testing.T) {
		spec := v1alpha1.FirewallRuleSpec{
			Type:  "managed_range",
			Label: "Missing RangeID",
		}

		rule, err := toFirewallRule(t.Context(), spec)

		require.Error(t, err)
		require.ErrorContains(t, err, "missing range_id")
		require.Equal(t, domain.FirewallRule{}, rule)
	})

	t.Run("it returns error with wrapped context", func(t *testing.T) {
		spec := v1alpha1.FirewallRuleSpec{
			Type: "custom_range",
		}

		rule, err := toFirewallRule(t.Context(), spec)

		require.Error(t, err)
		require.ErrorContains(t, err, "validate firewall rule")
		require.Equal(t, domain.FirewallRule{}, rule)
	})
}
