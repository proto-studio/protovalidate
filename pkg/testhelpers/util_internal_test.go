package testhelpers

import "testing"

// TestNeverAssignableImpl_priv tests that the priv() method works.
// This is needed for coverage even though it's an internal implementation detail.
func TestNeverAssignableImpl_priv(t *testing.T) {
	na := &neverAssignableImpl{}
	// Just call it - it's a no-op but needed for coverage
	na.priv()
}

