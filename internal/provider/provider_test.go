package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestStringOrEnv(t *testing.T) {
	t.Setenv("HUDDLE_API_KEY", "env-key")
	got := stringOrEnv(types.StringNull(), "HUDDLE_API_KEY")
	if got != "env-key" {
		t.Fatalf("expected env fallback, got %q", got)
	}

	got = stringOrEnv(types.StringValue("config-key"), "HUDDLE_API_KEY")
	if got != "config-key" {
		t.Fatalf("expected config value precedence, got %q", got)
	}
}

func TestStringOrDefault(t *testing.T) {
	got := stringOrDefault(types.StringNull(), "https://example.com")
	if got != "https://example.com" {
		t.Fatalf("unexpected default: %q", got)
	}
}

func TestInt64OrDefault(t *testing.T) {
	got := int64OrDefault(types.Int64Null(), 60)
	if got != 60 {
		t.Fatalf("unexpected default: %d", got)
	}
}

func TestGenerateIdempotencyKey(t *testing.T) {
	a := generateIdempotencyKey()
	b := generateIdempotencyKey()
	if len(a) != 32 || len(b) != 32 {
		t.Fatalf("unexpected key length: %d / %d", len(a), len(b))
	}
	if a == b {
		t.Fatalf("expected distinct keys")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
