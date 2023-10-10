package objects

import "proto.zip/studio/validate/pkg/rules"

// Conditional interface must be implemented for rules that are passed into WithConditionalKey.
// They must implement all of the standard rule methods as well as a method Keys which should return
// an array of all the keys names that must be present and error free for the rule to evaluate.
//
// ObjectRuleSet[T] implements this interface out of the box.
type Conditional[T any] interface {
	rules.Rule[T]
	Keys() []string // Return all key names that the rule depends on
}
