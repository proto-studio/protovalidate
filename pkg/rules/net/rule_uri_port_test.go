package net_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rules/net"
	"proto.zip/studio/validate/pkg/testhelpers"
)

func TestWithMinPort(t *testing.T) {
	ruleSet := net.NewURI().WithMinPort(1000).Any()

	testhelpers.MustBeInvalid(t, ruleSet, "http://example.com:999", errors.CodeMin)
	testhelpers.MustBeValid(t, ruleSet, "http://example.com:1000", "http://example.com:1000")
}

func TestWithMaxPort(t *testing.T) {
	ruleSet := net.NewURI().WithMaxPort(10000).Any()

	testhelpers.MustBeInvalid(t, ruleSet, "http://example.com:10001", errors.CodeMax)
	testhelpers.MustBeValid(t, ruleSet, "http://example.com:9999", "http://example.com:9999")
}

func TestWithAllowedPorts(t *testing.T) {
	ruleSet := net.NewURI().WithAllowedPorts(100, 200).Any()

	testhelpers.MustBeInvalid(t, ruleSet, "http://example.com:150", errors.CodeNotAllowed)
	testhelpers.MustBeValid(t, ruleSet, "http://example.com:100", "http://example.com:100")
}
