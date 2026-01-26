package database

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo/scalingomock"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestManager_updateFirewallRules(t *testing.T) {
	t.Run("it does nothing when current rules are already created", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		firewallCustomRule := domain.FirewallRule{
			Type:  domain.FirewallRuleTypeCustomRange,
			CIDR:  "0.0.0.0/0",
			Label: "Allow all",
		}
		currentDB := domain.Database{
			ID:            databaseID,
			AddonID:       addonID,
			FireWallRules: []domain.FirewallRule{firewallCustomRule},
		}
		expectedRules := []domain.FirewallRule{firewallCustomRule}

		// When
		err := manager.updateFirewallRules(ctx, currentDB, expectedRules)

		// Then
		require.NoError(t, err)
	})

	t.Run("it does nothing when expected rules are empty", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		currentDB := domain.Database{
			ID:      databaseID,
			AddonID: addonID,
		}

		// Empty rules list - code will process the loop but with zero iterations
		expectedRules := []domain.FirewallRule{}

		// When
		err := manager.updateFirewallRules(ctx, currentDB, expectedRules)

		// Then
		require.NoError(t, err)
	})

	t.Run("it fails at creating firewall rule", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		firewallCustomRule := domain.FirewallRule{
			Type: domain.FirewallRuleTypeCustomRange,
			CIDR: "192.168.1.0/24",
		}
		firewallManagedRule := domain.FirewallRule{
			Type:    domain.FirewallRuleTypeManagedRange,
			RangeID: "man-osc-st-fr1-egress",
		}
		currentDB := domain.Database{
			ID:      databaseID,
			AddonID: addonID,
		}
		expectedRules := []domain.FirewallRule{firewallCustomRule, firewallManagedRule}

		errCreateRule := errors.New("create firewall rule")

		scClient.EXPECT().CreateFirewallRule(gomock.Any(), currentDB.ID, currentDB.AddonID, firewallCustomRule)
		scClient.EXPECT().CreateFirewallRule(gomock.Any(), currentDB.ID, currentDB.AddonID, firewallManagedRule).
			Return(errCreateRule)

		// When
		err := manager.updateFirewallRules(ctx, currentDB, expectedRules)

		// Then
		require.ErrorIs(t, err, errCreateRule)
	})

	t.Run("it successfully creates firewall rules", func(t *testing.T) {
		// Given
		ctx := t.Context()
		ctrl := gomock.NewController(t)
		scClient := scalingomock.NewMockClient(ctrl)

		manager := manager{
			scClient: scClient,
		}

		firewallCustomRule := domain.FirewallRule{
			Type: domain.FirewallRuleTypeCustomRange,
			CIDR: "192.168.1.0/24",
		}
		firewallManagedRule := domain.FirewallRule{
			Type:    domain.FirewallRuleTypeManagedRange,
			RangeID: "man-osc-st-fr1-egress",
		}
		currentDB := domain.Database{
			ID:      databaseID,
			AddonID: addonID,
		}
		expectedRules := []domain.FirewallRule{firewallCustomRule, firewallManagedRule}

		scClient.EXPECT().CreateFirewallRule(gomock.Any(), currentDB.ID, currentDB.AddonID, firewallCustomRule)
		scClient.EXPECT().CreateFirewallRule(gomock.Any(), currentDB.ID, currentDB.AddonID, firewallManagedRule)

		// When
		err := manager.updateFirewallRules(ctx, currentDB, expectedRules)

		// Then
		require.NoError(t, err)
	})
}

func TestManager_extractRulesDiff(t *testing.T) {
	t.Run("it extracts rules differences", func(t *testing.T) {
		// Given
		current := []domain.FirewallRule{
			{Type: domain.FirewallRuleTypeCustomRange, CIDR: "192.168.0.1/24", Label: "first"},
			{Type: domain.FirewallRuleTypeManagedRange, RangeID: "range-1"},
		}
		requested := []domain.FirewallRule{
			{Type: domain.FirewallRuleTypeManagedRange, RangeID: "range-2"},
			{Type: domain.FirewallRuleTypeCustomRange, CIDR: "192.168.0.1/24", Label: "redundant"},
			{Type: domain.FirewallRuleTypeCustomRange, CIDR: "0.0.0.0/0", Label: "all"},
		}

		expectedDiff := rulesDiff{
			newRules: []domain.FirewallRule{
				{Type: domain.FirewallRuleTypeCustomRange, CIDR: "0.0.0.0/0", Label: "all"},
				{Type: domain.FirewallRuleTypeManagedRange, RangeID: "range-2"},
			},
			oldRules: []domain.FirewallRule{
				{Type: domain.FirewallRuleTypeManagedRange, RangeID: "range-1"},
			},
		}

		// When
		res := extractRulesDiff(current, requested)

		// Then
		require.ElementsMatch(t, expectedDiff.newRules, res.newRules)
		require.ElementsMatch(t, expectedDiff.oldRules, res.oldRules)
	})
}
