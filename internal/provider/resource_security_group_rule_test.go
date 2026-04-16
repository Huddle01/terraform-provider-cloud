package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestParseRuleImportID_ValidComposite(t *testing.T) {
	sgID, ruleID := parseRuleImportID("sg-0a1b2c/rl-9z8y7x")
	if sgID != "sg-0a1b2c" {
		t.Fatalf("expected sgID %q, got %q", "sg-0a1b2c", sgID)
	}
	if ruleID != "rl-9z8y7x" {
		t.Fatalf("expected ruleID %q, got %q", "rl-9z8y7x", ruleID)
	}
}

func TestSameRule_Matches(t *testing.T) {
	min := int64(22)
	max := int64(22)
	rule := &securityRuleItem{
		Direction:      "ingress",
		EtherType:      "IPv4",
		Protocol:       "tcp",
		PortRangeMin:   &min,
		PortRangeMax:   &max,
		RemoteIPPrefix: "0.0.0.0/0",
		RemoteGroupID:  "",
	}
	plan := securityGroupRuleResourceModel{
		Direction:      types.StringValue("ingress"),
		EtherType:      types.StringValue("IPv4"),
		Protocol:       types.StringValue("tcp"),
		PortRangeMin:   types.Int64Value(22),
		PortRangeMax:   types.Int64Value(22),
		RemoteIPPrefix: types.StringValue("0.0.0.0/0"),
		RemoteGroupID:  types.StringNull(),
	}
	if !sameRule(rule, plan) {
		t.Fatal("expected rules to match")
	}
}

func TestSameRule_DirectionMismatch(t *testing.T) {
	rule := &securityRuleItem{Direction: "egress", EtherType: "IPv4"}
	plan := securityGroupRuleResourceModel{
		Direction: types.StringValue("ingress"),
		EtherType: types.StringValue("IPv4"),
	}
	if sameRule(rule, plan) {
		t.Fatal("expected rules not to match on direction")
	}
}

func TestSameRule_ProtocolMismatch(t *testing.T) {
	rule := &securityRuleItem{Direction: "ingress", EtherType: "IPv4", Protocol: "udp"}
	plan := securityGroupRuleResourceModel{
		Direction: types.StringValue("ingress"),
		EtherType: types.StringValue("IPv4"),
		Protocol:  types.StringValue("tcp"),
	}
	if sameRule(rule, plan) {
		t.Fatal("expected rules not to match on protocol")
	}
}

func TestSameInt64PtrValue_BothNil(t *testing.T) {
	if !sameInt64PtrValue(nil, nil) {
		t.Fatal("expected nil == nil")
	}
}

func TestSameInt64PtrValue_OneNil(t *testing.T) {
	v := int64(22)
	if sameInt64PtrValue(nil, &v) {
		t.Fatal("expected nil != &22")
	}
	if sameInt64PtrValue(&v, nil) {
		t.Fatal("expected &22 != nil")
	}
}

func TestSameInt64PtrValue_Equal(t *testing.T) {
	a, b := int64(80), int64(80)
	if !sameInt64PtrValue(&a, &b) {
		t.Fatal("expected 80 == 80")
	}
}

func TestSameInt64PtrValue_NotEqual(t *testing.T) {
	a, b := int64(80), int64(443)
	if sameInt64PtrValue(&a, &b) {
		t.Fatal("expected 80 != 443")
	}
}
