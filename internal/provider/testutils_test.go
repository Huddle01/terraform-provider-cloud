package provider

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

// testAccProviderFactories wires the provider-under-test into every acceptance test step.
var testAccProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"huddle": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck fails fast if the minimum required env vars are absent.
// resource.Test already skips when TF_ACC is unset; this validates credentials.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	for _, env := range []string{"HUDDLE_API_KEY", "HUDDLE_REGION"} {
		if os.Getenv(env) == "" {
			t.Fatalf("env var %s must be set for acceptance tests", env)
		}
	}
}

// accName returns a deterministically unique name safe for cloud resource names.
func accName(base string) string {
	return fmt.Sprintf("tf-acc-%s-%s", base, acctest.RandString(6))
}

// testAccClient builds an apiClient from test env vars for use in CheckDestroy helpers.
func testAccClient() *apiClient {
	baseURL := os.Getenv("HUDDLE_LOCAL_BASE_URL")
	if baseURL == "" {
		baseURL = os.Getenv("HUDDLE_BASE_URL") // legacy fallback
	}
	if baseURL == "" {
		baseURL = "https://cloud.huddleapis.com/api/v1"
	}
	return newAPIClient(apiClientConfig{
		APIKey:        os.Getenv("HUDDLE_API_KEY"),
		BaseURL:       baseURL,
		DefaultRegion: os.Getenv("HUDDLE_REGION"),
		Timeout:       30 * time.Second,
	})
}

// testAccRegion returns the acceptance test region from env.
func testAccRegion() string {
	return os.Getenv("HUDDLE_REGION")
}

// testAccFlavorName returns the flavor name to use for instance tests (e.g. "anton-2").
// Set HUDDLE_FLAVOR_NAME to override; the test is skipped if unset.
func testAccFlavorName(t *testing.T) string {
	t.Helper()
	v := os.Getenv("HUDDLE_FLAVOR_NAME")
	if v == "" {
		t.Skip("HUDDLE_FLAVOR_NAME not set — skipping instance acceptance tests")
	}
	return strings.ToLower(v)
}

// testAccImageName returns the image name to use for instance tests (e.g. "ubuntu-22.04").
// Set HUDDLE_IMAGE_NAME to override; the test is skipped if unset.
func testAccImageName(t *testing.T) string {
	t.Helper()
	v := os.Getenv("HUDDLE_IMAGE_NAME")
	if v == "" {
		t.Skip("HUDDLE_IMAGE_NAME not set — skipping instance acceptance tests")
	}
	return strings.ToLower(v)
}

// testAccSGName returns the security group name to use for instance tests.
// Set HUDDLE_SG_NAME to override; defaults to "default".
func testAccSGName() string {
	if v := os.Getenv("HUDDLE_SG_NAME"); v != "" {
		return v
	}
	return "default"
}

// testAccSSHPublicKey returns an SSH public key for acceptance tests.
// Set HUDDLE_SSH_PUBLIC_KEY to use a real key.
func testAccSSHPublicKey() string {
	if v := os.Getenv("HUDDLE_SSH_PUBLIC_KEY"); v != "" {
		return v
	}
	// Deterministic test key — NOT for production use.
	return "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJBv1KfdsdR3VXQVX2K60Rb8VzrBTKn2FLnozM6C3qnr tf-acc-test"
}
