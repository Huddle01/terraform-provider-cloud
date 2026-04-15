package provider

import (
	"os"
	"testing"
)

func TestAcceptancePrereqs(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run acceptance tests")
	}

	required := []string{
		"HUDDLE_API_KEY",
		"HUDDLE_REGION",
	}
	for _, key := range required {
		if os.Getenv(key) == "" {
			t.Fatalf("missing %s", key)
		}
	}
}
