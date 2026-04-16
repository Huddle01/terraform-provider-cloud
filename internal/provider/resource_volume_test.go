package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAttachedServerIDs(t *testing.T) {
	attachments, diags := flattenVolumeAttachments([]volumeAttachmentPayload{
		{ServerID: "inst-1", Device: "/dev/vdb"},
		{ServerID: "inst-2", Device: "/dev/vdc"},
	})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	ids := attachedServerIDs(attachments)
	if len(ids) != 2 {
		t.Fatalf("expected 2 attached server ids, got %d", len(ids))
	}
	if ids[0] != "inst-1" || ids[1] != "inst-2" {
		t.Fatalf("unexpected attached ids: %#v", ids)
	}
}

func TestAttachedServerIDsNullList(t *testing.T) {
	ids := attachedServerIDs(types.ListNull(types.ObjectType{AttrTypes: map[string]attr.Type{
		"server_id": types.StringType,
		"device":    types.StringType,
	}}))
	if len(ids) != 0 {
		t.Fatalf("expected no ids, got %#v", ids)
	}
}

func TestFlattenVolumeAttachments(t *testing.T) {
	attachments, diags := flattenVolumeAttachments([]volumeAttachmentPayload{{ServerID: "inst-1", Device: "/dev/vdb"}})
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if attachments.IsNull() || attachments.IsUnknown() {
		t.Fatalf("expected concrete attachment list")
	}
	if len(attachments.Elements()) != 1 {
		t.Fatalf("expected one attachment")
	}
}
