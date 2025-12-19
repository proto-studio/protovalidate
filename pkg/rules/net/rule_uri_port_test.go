package net_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/net"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestURIRuleSet_WithMinPort(t *testing.T) {
	ruleSet := net.URI().WithMinPort(1000).Any()

	testhelpers.MustNotApply(t, ruleSet, "http://example.com:999", errors.CodeMin)
	testhelpers.MustApply(t, ruleSet, "http://example.com:1000")
}

func TestURIRuleSet_WithMaxPort(t *testing.T) {
	ruleSet := net.URI().WithMaxPort(10000).Any()

	testhelpers.MustNotApply(t, ruleSet, "http://example.com:10001", errors.CodeMax)
	testhelpers.MustApply(t, ruleSet, "http://example.com:9999")
}

func TestURIRuleSet_WithAllowedPorts(t *testing.T) {
	ruleSet := net.URI().WithAllowedPorts(100, 200).Any()

	testhelpers.MustNotApply(t, ruleSet, "http://example.com:150", errors.CodeNotAllowed)
	testhelpers.MustApply(t, ruleSet, "http://example.com:100")
}
