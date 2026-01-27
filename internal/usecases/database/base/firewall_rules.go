package database

import (
	"context"
	"slices"

	"golang.org/x/sync/errgroup"

	logf "sigs.k8s.io/controller-runtime/pkg/log"

	errors "github.com/Scalingo/go-utils/errors/v3"
	"github.com/Scalingo/scalingo-operator/internal/domain"
)

// updateFirewallRules adds or delete rules so as to bring the current firewall rules to the expected ones.
func (m *manager) updateFirewallRules(ctx context.Context, currentDB domain.Database, expectedRules []domain.FirewallRule) error {
	rulesToApply := establishRulesToApply(currentDB.FireWallRules, expectedRules)

	g, ctx := errgroup.WithContext(ctx)
	for _, ruleToApply := range rulesToApply {
		g.Go(func() error {
			switch ruleToApply.action {
			case addRuleAction:
				return m.addFirewallRule(ctx, currentDB.ID, currentDB.AddonID, ruleToApply.rule)
			case deleteRuleAction:
				return m.deleteFirewallRule(ctx, currentDB.ID, currentDB.AddonID, ruleToApply.rule)
			default:
				return errors.Newf(ctx, "undefined rule action %v", ruleToApply.action)
			}
		})
	}
	err := g.Wait()
	if err != nil {
		return errors.Wrap(ctx, err, "update firewall rules")
	}
	return nil
}

func (m *manager) addFirewallRule(ctx context.Context, dbID, addonID string, rule domain.FirewallRule) error {
	log := logf.FromContext(ctx)

	err := m.scClient.CreateFirewallRule(ctx, dbID, addonID, rule)
	if err == nil {
		log.Info("Add firewall rule", "rule", rule)
	} else {
		log.Error(err, "Fail to add firewall rule", "AppID", dbID, "AddonID", addonID, "rule", rule)
	}
	return err
}

func (m *manager) deleteFirewallRule(ctx context.Context, dbID, addonID string, rule domain.FirewallRule) error {
	log := logf.FromContext(ctx)

	err := m.scClient.DeleteFirewallRule(ctx, dbID, addonID, rule.ID)
	if err == nil {
		log.Info("Delete firewall rule", "rule", rule)
	} else {
		log.Error(err, "Fail to delete firewall rule", "AppID", dbID, "AddonID", addonID, "rule", rule)
	}
	return err
}

type ruleToApplyAction int

const (
	addRuleAction ruleToApplyAction = iota
	deleteRuleAction
)

type ruleToApply struct {
	rule   domain.FirewallRule
	action ruleToApplyAction
}

type rulesToApply []ruleToApply

// establishRulesToApply compares two FirewallRule slices so as to establish the rules to apply.
func establishRulesToApply(currentRules []domain.FirewallRule, expectedRules []domain.FirewallRule) rulesToApply {
	// Sort slices to further use BinarySearchFunc.
	slices.SortFunc(currentRules, domain.CompareFirewallRules)
	slices.SortFunc(expectedRules, domain.CompareFirewallRules)

	res := make(rulesToApply, 0, len(expectedRules)+len(currentRules))

	for _, rule := range expectedRules {
		_, found := slices.BinarySearchFunc(currentRules, rule, domain.CompareFirewallRules)
		if !found {
			res = append(res, ruleToApply{rule, addRuleAction})
		}
	}
	for _, rule := range currentRules {
		_, found := slices.BinarySearchFunc(expectedRules, rule, domain.CompareFirewallRules)
		if !found {
			res = append(res, ruleToApply{rule, deleteRuleAction})
		}
	}
	return res
}
