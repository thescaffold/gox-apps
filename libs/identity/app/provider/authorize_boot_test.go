package provider

import "testing"

// TestAuthorizeController_Boot_SelfWiresRegistry is an internal-package test
// (not tests/identity_test.go) because it needs the unexported `registry`
// field to verify the fix directly: before this Boot() existed, `registry`
// stayed nil forever (PLAN M0-27a — nothing anywhere called SetRegistry),
// so every /provider/:provider/authorize call 400'd with "oauth registry
// not configured" regardless of credentials.
func TestAuthorizeController_Boot_SelfWiresRegistry(t *testing.T) {
	c := &AuthorizeController{}

	if c.registry != nil {
		t.Fatalf("registry should start nil")
	}

	if err := c.Boot(nil); err != nil {
		t.Fatalf("Boot returned error: %v", err)
	}

	if c.registry == nil {
		t.Fatalf("Boot did not wire a registry")
	}

	for _, name := range []string{"github", "gitlab", "bitbucket", "google"} {
		if c.registry.Get(name) == nil {
			t.Errorf("provider %q not registered after Boot", name)
		}
	}
}

func TestAuthorizeController_Boot_DoesNotClobberExistingRegistry(t *testing.T) {
	c := &AuthorizeController{}

	if err := c.Boot(nil); err != nil {
		t.Fatalf("first Boot returned error: %v", err)
	}
	wired := c.registry

	if err := c.Boot(nil); err != nil {
		t.Fatalf("second Boot returned error: %v", err)
	}

	if c.registry != wired {
		t.Fatalf("second Boot() replaced an already-wired registry")
	}
}
