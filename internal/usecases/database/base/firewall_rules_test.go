package database

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/Scalingo/scalingo-operator/internal/boundaries/out/scalingo/scalingomock"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

func TestManager_UpdateFirewallRules(t *testing.T) {
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
