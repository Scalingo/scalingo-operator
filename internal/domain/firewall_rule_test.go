package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFirewallRuleType_Validate(t *testing.T) {
	t.Run("it successfully validates managed_range type", func(t *testing.T) {
		require.NoError(t, FirewallRuleTypeManagedRange.Validate())
	})

	t.Run("it successfully validates custom_range type", func(t *testing.T) {
		require.NoError(t, FirewallRuleTypeCustomRange.Validate())
	})

	t.Run("it returns error for invalid type", func(t *testing.T) {
		require.ErrorContains(t, FirewallRuleType("invalid").Validate(), "invalid firewall rule type")
	})

	t.Run("it returns error for empty type", func(t *testing.T) {
		require.ErrorContains(t, FirewallRuleType("").Validate(), "invalid firewall rule type")
	})
}

func TestFirewallRule_String(t *testing.T) {
	t.Run("it returns string representation with all fields", func(t *testing.T) {
		rule := FirewallRule{
			ID:      "rule-123",
			Type:    FirewallRuleTypeManagedRange,
			CIDR:    "192.168.1.0/24",
			Label:   "Test Rule",
			RangeID: "range-456",
		}
		expected := "{ ID: 'rule-123', Type: managed_range, CIDR: '192.168.1.0/24', Label: 'Test Rule', RangeID: 'range-456' }"
		require.Equal(t, expected, rule.String())
	})

	t.Run("it returns string representation with empty fields", func(t *testing.T) {
		rule := FirewallRule{}
		expected := "{ ID: '', Type: , CIDR: '', Label: '', RangeID: '' }"
		require.Equal(t, expected, rule.String())
	})
}

func TestFirewallRule_Validate(t *testing.T) {
	t.Run("it successfully validates managed_range rule with range_id", func(t *testing.T) {
		rule := FirewallRule{
			Type:    FirewallRuleTypeManagedRange,
			RangeID: "range-123",
		}
		require.NoError(t, rule.Validate())
	})

	t.Run("it successfully validates custom_range rule with cidr", func(t *testing.T) {
		rule := FirewallRule{
			Type: FirewallRuleTypeCustomRange,
			CIDR: "10.0.0.0/16",
		}
		require.NoError(t, rule.Validate())
	})

	t.Run("it returns error for invalid type", func(t *testing.T) {
		rule := FirewallRule{
			Type: FirewallRuleType("invalid"),
		}
		require.ErrorContains(t, rule.Validate(), "invalid firewall rule type")
	})

	t.Run("it returns error for managed_range without range_id", func(t *testing.T) {
		rule := FirewallRule{
			Type:    FirewallRuleTypeManagedRange,
			RangeID: "",
		}
		require.ErrorContains(t, rule.Validate(), "missing range_id")
	})

	t.Run("it returns error for custom_range without cidr", func(t *testing.T) {
		rule := FirewallRule{
			Type: FirewallRuleTypeCustomRange,
			CIDR: "",
		}
		require.ErrorContains(t, rule.Validate(), "missing cidr")
	})
}

func TestFirewallRulesCompare(t *testing.T) {

	t.Run("it compares rules by types only", func(t *testing.T) {
		ruleManagedRange := FirewallRule{Type: FirewallRuleTypeManagedRange}
		ruleManagedCustom := FirewallRule{Type: FirewallRuleTypeCustomRange}

		require.Less(t, CompareFirewallRules(ruleManagedCustom, ruleManagedRange), 0)
		require.Greater(t, CompareFirewallRules(ruleManagedRange, ruleManagedCustom), 0)
		require.Equal(t, CompareFirewallRules(ruleManagedCustom, ruleManagedCustom), 0)
		require.Equal(t, CompareFirewallRules(ruleManagedRange, ruleManagedRange), 0)
	})

	t.Run("it compares custom rules", func(t *testing.T) {
		rule1 := FirewallRule{Type: FirewallRuleTypeCustomRange, CIDR: "0.0.0.0/0"}
		rule2 := FirewallRule{Type: FirewallRuleTypeCustomRange, CIDR: "192.168.0.1/24"}
		rule3 := FirewallRule{Type: FirewallRuleTypeCustomRange, CIDR: "192.168.0.1/24", Label: "label"}

		require.Less(t, CompareFirewallRules(rule1, rule2), 0)
		require.Greater(t, CompareFirewallRules(rule2, rule1), 0)
		require.Equal(t, CompareFirewallRules(rule2, rule3), 0)
	})

	t.Run("it compares managed rules", func(t *testing.T) {
		rule1 := FirewallRule{Type: FirewallRuleTypeManagedRange, RangeID: "man-osc-fr1-egress"}
		rule2 := FirewallRule{Type: FirewallRuleTypeManagedRange, RangeID: "man-osc-secnum-fr1-egress"}

		require.Less(t, CompareFirewallRules(rule1, rule2), 0)
		require.Greater(t, CompareFirewallRules(rule2, rule1), 0)
		require.Equal(t, CompareFirewallRules(rule2, rule2), 0)
	})

	t.Run("it compares mixed rules", func(t *testing.T) {
		rule1 := FirewallRule{Type: FirewallRuleTypeCustomRange, CIDR: "0.0.0.0/0"}
		rule2 := FirewallRule{Type: FirewallRuleTypeCustomRange, CIDR: "192.168.0.1/24"}
		rule3 := FirewallRule{Type: FirewallRuleTypeManagedRange, RangeID: "man-osc-fr1-egress"}

		require.Less(t, CompareFirewallRules(rule1, rule2), 0)
		require.Less(t, CompareFirewallRules(rule2, rule3), 0)
	})
}
