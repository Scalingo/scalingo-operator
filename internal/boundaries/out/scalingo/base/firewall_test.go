package scalingo

import (
	"testing"

	"github.com/stretchr/testify/require"

	scalingoapi "github.com/Scalingo/go-scalingo/v9"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestToFirewallRule(t *testing.T) {
	t.Run("it converts custom range firewall rule", func(t *testing.T) {
		scalingoRule := scalingoapi.FirewallRule{
			ID:    "rule-123",
			Type:  scalingoapi.FirewallRuleTypeCustomRange,
			CIDR:  "10.0.0.0/8",
			Label: "custom-label",
		}

		expected := domain.FirewallRule{
			ID:    "rule-123",
			Type:  domain.FirewallRuleTypeCustomRange,
			CIDR:  "10.0.0.0/8",
			Label: "custom-label",
		}

		result := toFirewallRule(scalingoRule)

		require.Equal(t, expected, result)
	})

	t.Run("it converts managed range firewall rule", func(t *testing.T) {
		scalingoRule := scalingoapi.FirewallRule{
			ID:      "rule-456",
			Type:    scalingoapi.FirewallRuleTypeManagedRange,
			Label:   "managed-label",
			RangeID: "range-789",
		}

		expected := domain.FirewallRule{
			ID:      "rule-456",
			Type:    domain.FirewallRuleTypeManagedRange,
			Label:   "managed-label",
			RangeID: "range-789",
		}

		result := toFirewallRule(scalingoRule)

		require.Equal(t, expected, result)
	})
}

func TestToFirewallRuleType(t *testing.T) {
	tests := map[string]struct {
		scalingoType scalingoapi.FirewallRuleType
		expectedType domain.FirewallRuleType
	}{
		"it converts managed range type": {
			scalingoType: scalingoapi.FirewallRuleTypeManagedRange,
			expectedType: domain.FirewallRuleTypeManagedRange,
		},
		"it converts custom range type": {
			scalingoType: scalingoapi.FirewallRuleTypeCustomRange,
			expectedType: domain.FirewallRuleTypeCustomRange,
		},
		"it defaults to custom range for unknown type": {
			scalingoType: "unknown",
			expectedType: domain.FirewallRuleTypeCustomRange,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := toFirewallRuleType(test.scalingoType)
			require.Equal(t, test.expectedType, result)
		})
	}
}

func TestToScalingoFirewallRuleType(t *testing.T) {
	tests := map[string]struct {
		domainType   domain.FirewallRuleType
		expectedType scalingoapi.FirewallRuleType
	}{
		"it converts managed range type": {
			domainType:   domain.FirewallRuleTypeManagedRange,
			expectedType: scalingoapi.FirewallRuleTypeManagedRange,
		},
		"it converts custom range type": {
			domainType:   domain.FirewallRuleTypeCustomRange,
			expectedType: scalingoapi.FirewallRuleTypeCustomRange,
		},
		"it defaults to custom range for unknown type": {
			domainType:   "unknown",
			expectedType: scalingoapi.FirewallRuleTypeCustomRange,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := toScalingoFirewallRuleType(test.domainType)
			require.Equal(t, test.expectedType, result)
		})
	}
}
