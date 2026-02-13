package adapters

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

		result, err := ToFirewallRule(t.Context(), scalingoRule)

		require.NoError(t, err)
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

		result, err := ToFirewallRule(t.Context(), scalingoRule)

		require.NoError(t, err)
		require.Equal(t, expected, result)
	})

	t.Run("it fails converting firewall rule with unknown range", func(t *testing.T) {
		scalingoRule := scalingoapi.FirewallRule{
			ID:    "rule-123",
			Type:  scalingoapi.FirewallRuleType("new unknown range"),
			CIDR:  "10.0.0.0/8",
			Label: "custom-label",
		}

		_, err := ToFirewallRule(t.Context(), scalingoRule)

		require.ErrorContains(t, err, "to firewall rule type: invalid type new unknown range")
	})
}

func TestToFirewallRuleType(t *testing.T) {
	tests := map[string]struct {
		scalingoType  scalingoapi.FirewallRuleType
		expectedType  domain.FirewallRuleType
		expectedError string
	}{
		"it converts managed range type": {
			scalingoType: scalingoapi.FirewallRuleTypeManagedRange,
			expectedType: domain.FirewallRuleTypeManagedRange,
		},
		"it converts custom range type": {
			scalingoType: scalingoapi.FirewallRuleTypeCustomRange,
			expectedType: domain.FirewallRuleTypeCustomRange,
		},
		"it fails for unknown range type": {
			scalingoType:  "UNKNOWN",
			expectedError: "invalid type UNKNOWN",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := toFirewallRuleType(t.Context(), test.scalingoType)

			if test.expectedError != "" {
				require.ErrorContains(t, err, test.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedType, result)
			}
		})
	}
}

func TestToScalingoFirewallRuleType(t *testing.T) {
	tests := map[string]struct {
		domainType    domain.FirewallRuleType
		expectedType  scalingoapi.FirewallRuleType
		expectedError string
	}{
		"it converts managed range type": {
			domainType:   domain.FirewallRuleTypeManagedRange,
			expectedType: scalingoapi.FirewallRuleTypeManagedRange,
		},
		"it converts custom range type": {
			domainType:   domain.FirewallRuleTypeCustomRange,
			expectedType: scalingoapi.FirewallRuleTypeCustomRange,
		},
		"it fails for unknown range type": {
			domainType:    "UNKNOWN",
			expectedError: "invalid type UNKNOWN",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := ToScalingoFirewallRuleType(t.Context(), test.domainType)
			if test.expectedError != "" {
				require.ErrorContains(t, err, test.expectedError)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.expectedType, result)
			}
		})
	}
}
