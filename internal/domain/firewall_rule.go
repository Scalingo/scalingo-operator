package domain

import (
	"errors"
	"fmt"
)

type FirewallRuleType string

const (
	FirewallRuleTypeManagedRange FirewallRuleType = "managed_range"
	FirewallRuleTypeCustomRange  FirewallRuleType = "custom_range"
)

type FirewallRule struct {
	ID      string
	Type    FirewallRuleType
	CIDR    string
	Label   string
	RangeID string
}

func (t FirewallRuleType) Validate() error {
	switch t {
	case FirewallRuleTypeManagedRange, FirewallRuleTypeCustomRange:
		return nil
	default:
		return fmt.Errorf("invalid firewall rule type: %s", t)
	}
}

func (r FirewallRule) String() string {
	return fmt.Sprintf("{ ID: '%s', Type: %s, CIDR: '%s', Label: '%s', RangeID: '%s' }",
		r.ID,
		r.Type,
		r.CIDR,
		r.Label,
		r.RangeID)
}

func (r FirewallRule) Validate() error {
	err := r.Type.Validate()
	if err != nil {
		return err
	}

	switch r.Type {
	case FirewallRuleTypeManagedRange:
		if r.RangeID == "" {
			return errors.New("missing range_id")
		}
	case FirewallRuleTypeCustomRange:
		if r.CIDR == "" {
			return errors.New("missing cidr")
		}
	}
	return nil
}
