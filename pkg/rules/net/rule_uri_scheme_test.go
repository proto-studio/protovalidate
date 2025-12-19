package net_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/net"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestURIRuleSet_WithAllowedSchemes(t *testing.T) {
	ruleSet := net.URI().WithAllowedSchemes("http", "https").Any()

	testhelpers.MustNotApply(t, ruleSet, "ftp://example.com", errors.CodeNotAllowed)
	testhelpers.MustApply(t, ruleSet, "http://example.com")
}
