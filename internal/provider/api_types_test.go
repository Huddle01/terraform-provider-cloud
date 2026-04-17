package provider

import (
	"encoding/json"
	"testing"
)

func TestEffectiveEtherType_EtherTypeSet(t *testing.T) {
	r := securityRuleItem{EtherType: "IPv6", Ethertype: ""}
	if got := r.EffectiveEtherType(); got != "IPv6" {
		t.Fatalf("expected IPv6, got %q", got)
	}
}

func TestEffectiveEtherType_FallsBackToEthertype(t *testing.T) {
	r := securityRuleItem{EtherType: "", Ethertype: "IPv4"}
	if got := r.EffectiveEtherType(); got != "IPv4" {
		t.Fatalf("expected IPv4, got %q", got)
	}
}

func TestEffectiveEtherType_BothEmpty(t *testing.T) {
	r := securityRuleItem{EtherType: "", Ethertype: ""}
	if got := r.EffectiveEtherType(); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestSecurityRuleUnmarshalJSON_AllFields(t *testing.T) {
	min := int64(22)
	max := int64(22)
	raw, _ := json.Marshal(map[string]any{
		"id":                "rule-1",
		"direction":         "ingress",
		"ether_type":        "IPv4",
		"protocol":          "tcp",
		"port_range_min":    min,
		"port_range_max":    max,
		"remote_ip_prefix":  "0.0.0.0/0",
		"remote_group_id":   "",
		"security_group_id": "sg-1",
		"created_at":        "2024-01-01T00:00:00Z",
	})

	var r securityRuleItem
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if r.ID != "rule-1" {
		t.Errorf("ID: expected %q, got %q", "rule-1", r.ID)
	}
	if r.Direction != "ingress" {
		t.Errorf("Direction: expected %q, got %q", "ingress", r.Direction)
	}
	if r.EtherType != "IPv4" {
		t.Errorf("EtherType: expected %q, got %q", "IPv4", r.EtherType)
	}
	if r.Protocol != "tcp" {
		t.Errorf("Protocol: expected %q, got %q", "tcp", r.Protocol)
	}
	if r.PortRangeMin == nil || *r.PortRangeMin != 22 {
		t.Errorf("PortRangeMin: expected 22, got %v", r.PortRangeMin)
	}
	if r.PortRangeMax == nil || *r.PortRangeMax != 22 {
		t.Errorf("PortRangeMax: expected 22, got %v", r.PortRangeMax)
	}
	if r.RemoteIPPrefix != "0.0.0.0/0" {
		t.Errorf("RemoteIPPrefix: expected %q, got %q", "0.0.0.0/0", r.RemoteIPPrefix)
	}
}

func TestSecurityRuleUnmarshalJSON_NullPortRange(t *testing.T) {
	raw := []byte(`{"id":"rule-2","direction":"egress","ether_type":"IPv4","port_range_min":null,"port_range_max":null}`)

	var r securityRuleItem
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if r.PortRangeMin != nil {
		t.Errorf("expected nil PortRangeMin, got %v", r.PortRangeMin)
	}
	if r.PortRangeMax != nil {
		t.Errorf("expected nil PortRangeMax, got %v", r.PortRangeMax)
	}
}

func TestSecurityRuleUnmarshalJSON_MissingFields(t *testing.T) {
	raw := []byte(`{"id":"rule-3"}`)

	var r securityRuleItem
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if r.ID != "rule-3" {
		t.Errorf("ID: expected %q, got %q", "rule-3", r.ID)
	}
	if r.Direction != "" || r.Protocol != "" {
		t.Errorf("expected zero values for missing fields")
	}
	if r.PortRangeMin != nil || r.PortRangeMax != nil {
		t.Errorf("expected nil port ranges for missing fields")
	}
}

func TestParseOptionalInt64_StringEncoded(t *testing.T) {
	raw := json.RawMessage(`"80"`)
	got := parseOptionalInt64(raw)
	if got == nil || *got != 80 {
		t.Fatalf("expected 80, got %v", got)
	}
}

func TestParseOptionalInt64_Null(t *testing.T) {
	raw := json.RawMessage(`null`)
	got := parseOptionalInt64(raw)
	if got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}
