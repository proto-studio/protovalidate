package net_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/net"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithAllowedSchemes(t *testing.T) {
	ruleSet := net.NewURI().WithAllowedSchemes("http", "https").Any()

	testhelpers.MustNotRun(t, ruleSet, "ftp://example.com", errors.CodeNotAllowed)
	testhelpers.MustRun(t, ruleSet, "http://example.com")
}
