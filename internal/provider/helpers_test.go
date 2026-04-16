package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestQueryWithRegion(t *testing.T) {
	q := queryWithRegion("eu2")
	if got := q.Get("region"); got != "eu2" {
		t.Fatalf("expected region=eu2, got %q", got)
	}
}

func TestQueryWithRegion_Empty(t *testing.T) {
	q := queryWithRegion("")
	if got := q.Get("region"); got != "" {
		t.Fatalf("expected empty region, got %q", got)
	}
}

func TestStringSliceToTerraform_Empty(t *testing.T) {
	got := stringSliceToTerraform([]string{})
	if len(got) != 0 {
		t.Fatalf("expected 0 elements, got %d", len(got))
	}
}

func TestStringSliceToTerraform_WithValues(t *testing.T) {
	got := stringSliceToTerraform([]string{"a", "b", "c"})
	if len(got) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(got))
	}
	if got[0].ValueString() != "a" || got[1].ValueString() != "b" || got[2].ValueString() != "c" {
		t.Fatalf("unexpected values: %v", got)
	}
}

func TestPtrInt64ToTerraform_Nil(t *testing.T) {
	got := ptrInt64ToTerraform(nil)
	if !got.IsNull() {
		t.Fatalf("expected null for nil pointer, got %v", got)
	}
}

func TestPtrInt64ToTerraform_Value(t *testing.T) {
	v := int64(42)
	got := ptrInt64ToTerraform(&v)
	if got.ValueInt64() != 42 {
		t.Fatalf("expected 42, got %d", got.ValueInt64())
	}
}

func TestListStringToSlice_Null(t *testing.T) {
	result := listStringToSlice(types.ListNull(types.StringType))
	if len(result) != 0 {
		t.Fatalf("expected empty slice for null list, got %v", result)
	}
}

func TestListStringToSlice_WithValues(t *testing.T) {
	list, diags := types.ListValueFrom(context.Background(), types.StringType, []string{"x", "y"})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	result := listStringToSlice(list)
	if len(result) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(result))
	}
	if result[0] != "x" || result[1] != "y" {
		t.Fatalf("unexpected values: %v", result)
	}
}

func TestBoolOrDefault_Null(t *testing.T) {
	got := boolOrDefault(types.BoolNull(), true)
	if !got {
		t.Fatalf("expected default true for null bool")
	}
}

func TestBoolOrDefault_Set(t *testing.T) {
	got := boolOrDefault(types.BoolValue(false), true)
	if got {
		t.Fatalf("expected false for set bool, got true")
	}
}

func TestInt64OrZero_Null(t *testing.T) {
	got := int64OrZero(types.Int64Null())
	if got != nil {
		t.Fatalf("expected nil for null int64")
	}
}

func TestInt64OrZero_Zero(t *testing.T) {
	// Zero is a valid int64 value — int64OrZero returns a pointer to 0, not nil.
	got := int64OrZero(types.Int64Value(0))
	if got == nil || *got != 0 {
		t.Fatalf("expected pointer to 0, got %v", got)
	}
}

func TestInt64OrZero_Value(t *testing.T) {
	got := int64OrZero(types.Int64Value(22))
	if got == nil || *got != 22 {
		t.Fatalf("expected 22, got %v", got)
	}
}

func TestStringOrEmpty_Null(t *testing.T) {
	got := stringOrEmpty(types.StringNull())
	if got != "" {
		t.Fatalf("expected empty string for null, got %q", got)
	}
}

func TestStringOrEmpty_Value(t *testing.T) {
	got := stringOrEmpty(types.StringValue("tcp"))
	if got != "tcp" {
		t.Fatalf("expected %q, got %q", "tcp", got)
	}
}

func TestParseRuleImportID_Valid(t *testing.T) {
	sgID, ruleID := parseRuleImportID("sg-abc/rule-xyz")
	if sgID != "sg-abc" {
		t.Fatalf("expected sgID %q, got %q", "sg-abc", sgID)
	}
	if ruleID != "rule-xyz" {
		t.Fatalf("expected ruleID %q, got %q", "rule-xyz", ruleID)
	}
}

func TestParseRuleImportID_NoSlash(t *testing.T) {
	sgID, ruleID := parseRuleImportID("invalid")
	if sgID != "" || ruleID != "" {
		t.Fatalf("expected empty strings for invalid id, got %q %q", sgID, ruleID)
	}
}

func TestParseRuleImportID_Empty(t *testing.T) {
	sgID, ruleID := parseRuleImportID("")
	if sgID != "" || ruleID != "" {
		t.Fatalf("expected empty strings for empty input, got %q %q", sgID, ruleID)
	}
}

func TestParseRuleImportID_OnlySlash(t *testing.T) {
	sgID, ruleID := parseRuleImportID("/")
	if sgID != "" || ruleID != "" {
		t.Fatalf("expected empty strings for slash-only input, got %q %q", sgID, ruleID)
	}
}
