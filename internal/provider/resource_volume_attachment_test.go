package provider

import "testing"

func TestParseVolumeAttachmentImportID(t *testing.T) {
	vol, inst := parseVolumeAttachmentImportID("vol-1/inst-1")
	if vol != "vol-1" || inst != "inst-1" {
		t.Fatalf("unexpected parsed values: %q, %q", vol, inst)
	}

	vol, inst = parseVolumeAttachmentImportID("bad")
	if vol != "" || inst != "" {
		t.Fatalf("expected empty parse for invalid id")
	}
}

func TestFindAttachmentByInstance(t *testing.T) {
	attachments := []volumeAttachmentPayload{
		{ServerID: "inst-1", Device: "/dev/vdb"},
		{ServerID: "inst-2", Device: "/dev/vdc"},
	}

	matched := findAttachmentByInstance(attachments, "inst-2")
	if matched == nil {
		t.Fatalf("expected a matching attachment")
	}
	if matched.Device != "/dev/vdc" {
		t.Fatalf("unexpected matched device: %q", matched.Device)
	}

	if x := findAttachmentByInstance(attachments, "inst-3"); x != nil {
		t.Fatalf("expected no match")
	}
}

func TestVolumeAttachmentStateID(t *testing.T) {
	got := volumeAttachmentStateID("vol-1", "inst-1")
	if got != "vol-1/inst-1" {
		t.Fatalf("unexpected state id: %q", got)
	}
}
