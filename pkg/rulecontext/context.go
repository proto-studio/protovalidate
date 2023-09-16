// Package rulecontext implements some helper functions to store values from the
// standard Go Context package.
package rulecontext

import (
	"context"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// Store the default message printer once it it initialized
var defaultPrinter *message.Printer

// Context keys to lookup values while avoiding conflicting keys
var printerContextKey int
var pathContextKey int
var RuleSetContextKey int

// init initialize any global variables needed
func init() {
	defaultPrinter = message.NewPrinter(language.AmericanEnglish)
}

// WithPrinters adds a printer key to a context
func WithPrinter(parent context.Context, printer *message.Printer) context.Context {
	if printer == nil {
		panic("expected printer to not be nil")
	}
	return context.WithValue(parent, &printerContextKey, printer)
}

// Printer returns the most recent printer from the context.
// If none is found it returns the default printer.
//
// This function never returns nil.
func Printer(ctx context.Context) *message.Printer {
	if ctx == nil {
		return defaultPrinter
	}

	printer := ctx.Value(&printerContextKey)

	if printer != nil {
		return printer.(*message.Printer)
	}

	return defaultPrinter
}

// WithRuleSet adds a rule set to the context.
func WithRuleSet(parent context.Context, ruleSet any) context.Context {
	if ruleSet == nil {
		panic("expected rule set to not be nil")
	}
	return context.WithValue(parent, &RuleSetContextKey, ruleSet)
}

// RuleSet returns the most resent rule set.
//
// For nested objects there may be more than one but only the most recent
// can be retrieved.
func RuleSet(ctx context.Context) any {
	if ctx == nil {
		return nil
	}

	return ctx.Value(&RuleSetContextKey)
}
