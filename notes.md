## Handle of Explicit Nil

Nil-able values such as interfaces (including `any`) now have a `WithNil` method. If `WithNil` is called
then the value is allowed to be `nil` even if it is required. Otherwise a `nil` value will raise an error.

This is useful when you have an API that allows a field to be explicitly cleared. For example, you may have
a Json REST API that has a `PATCH` operation. A patch with a `null`/`nil` value has different meaning than
if the field is entirely absent.




This is a **huge** update.

After months of using ProtoValidate in production we've made several adjustments to the underlying
API and added a number of new features.

Not all these features are backwards compatible. When possible we've deprecated the old functions rather
than replace or remove them. However, if you are writing custom `Rule` or `RuleSet` implementations you
will need to immediately update your code if you want to continue working with `0.3` and beyond.

Of course, once we hit `1.0` we will be freezing the API and all backwards incompatible changes will be
a major version update.

# New Features

## More Ways to Handle Unknown Keys

You now have more options to handle unknown keys in maps and structs.

By using the new `WithDynamicKey` method you can specify rules for validating keys and their values.
Keys that match a dynamic key will not trigger an unexpected key error.

As you may expect, there is now also `WithDynamicConditionalKey` which behaves the same as
`WithConditionalKey` only with support for dynamic key values.

You can also specify what you want to do with the conditional keys using `WithDynamicBucket` and
`WithConditionalDynamicBucket` that will let you specify a map structure to put the dynamic keys into.

## "Apply" Rules Sets

You no longer `Validate` a rule set, you `Apply` it.

Going forward you should use the `Apply` method and both `Validate` and `ValidateWithContext` are removed.

This change makes the code cleaner as well as reduces confusion due to the words Validate and Evaluate
being very similar sounding.

Also, the new `Apply` method returns only the errors and accepts a output pointer similar to `json.Unmarshal`.
This should make it easier to Apply rule sets directly to a struct and also support additional use cases. For
instance, `TimeRuleSetString` was removed because `TimeRuleSet` can now apply directly to a `string` pointer.

In some cases this change may also decrease extra memory allocation which may be useful for large structs.

## URI Rule Set

The new URI rules set can be used to validate URIs and URLs.

This powerful new rule set not only validates the URI but also gives detailed errors as to
what part of the URI is causing issues.

Features for `v0.3` include:

- Requiring specific schemes (`http`, `https`, `tel`, `email`, etc).
- Port range validation.
- Host validation.
- Relative URIs.
- Query string validation.
- Fragment validation.

## Typed Keys for Maps

Map validators can now have key types besides just `string`.

## `WithJson` on Objects

Object validators can now set a flag using `WithJson` that will allow the input into the `Run`
function (`Validate` prior to `v0.3.0`) to accept a string encoded as Json. This effectively
eliminates the need to unmarshal Json prior to running the rule set.

This is off by default since allowing Json strings in arbitrary fields could have unintended
security and performance effects.

## More Test Helpers

New test helpers, `NewMockRule` and `NewMockError`, can be used to create mocks and also track
how many times they are called.

## Constant Rule Set

The new Constant Rule Set can be used to test if a value is exactly equal to the provided value.

## Interface Rule Set

The new Interface Rule Set can be used to quickly create rule sets for interfaces. This has an advantage
over `AnyRuleSet` in that the results of the rule set are typed to the interface type.

This is useful when you have a Json object that has multiple types under a key but and you don't know
which one will be there ahead of time. You can use an interface and validate it without sacrificing much
type safety.

## WithAllowedValues and WithRejectedValues on Numbers

Both floating point and int now support `WithAllowedValues` and `WithRejectedValues`. They behave identical
to the functions on `StringRuleSet`.

# Performance

## Idempotence

The version more consistently returns the identity if a function that is ineffective is called.

For example `WithRequired` on Object Rule Sets will no longer return a new rule set if the rule
set already has the required flag set.

Also, methods to get an empty Rule Set should now return the same Rule Set every time instead of
allocating a new one.

If you find any functions that should be idempotent but are not, please create an issue.

# Backwards Incompatible Changes

## Package Reorganization

All packages that test types that are part of the standard Go language (objects, numbers, slices) are now
part of the base `rules ` package.

Also, all `New` methods for Rule Sets no longer have the `New` prefix and are idempotent in most case.

## Naming of Standard Types

Several rule sets have been renamed to be more consistent with idiomatic Go.

`Object` and `ObjectMap` have been replaced with `Struct` and `Map`. `ObjectRuleSet` is used for both structs and
maps so its naming has stayed the same.

`Array` has been replaced with `Slice`.

## Rule Mutations

The biggest backwards incompatible change is that `Rule` implementations can no longer mutate the
values. The `Evaluate` signature has been changed to only return errors and not a new value.

If you are not writing custom rules, this change does not affect you.

This offers significant advantages, mainly:

- `Evaluate` is now thread safe.
- The code is now cleaner and more maintainable.
- Fewer side effects.

**Upgrading Tips:** If you are not doing mutations you can simply change your `Evaluate` methods to
no longer return a value. If you are doing mutations in your `Evaluate` you will need to move it
to a rule set instead, since `Apply` can mutate values.

## RuleSet Interface No Longer Includes `Validate` or `ValidateWithContext`

Both `Validate` and `ValidateWithContext` are no longer in the `RuleSet` interface in favor of `Apply`
and have been removed from all the standard rule sets.

**Upgrading Tips:** Upgrading code to use `Apply` is fairly straight forward. Instead of taking the
first return value, you will need to create a variable to store the output and pass it in as a pointer.
If you were previously setting the results to a `struct` this will result in less code.

## Typed Rules on Maps

`WithKey` and `WithConditionalKey` on maps now expects a rule set with the same type as the value.

This means that if you are using a type other than "any" the rule set needs to be for that exact
type. In most cases you can simply remove the call to `.Any()` when passing in the rule set.

The reason for this change is to catch potential type mismatches at compile time. Additionally,
this change reduces repetitive boilerplate code.

## `TimeStringRuleSet` Has Been Removed

The `TimeStringRuleSet` has been removed. You should use `TimeRuleSet` directly.

Instead of `NewTimeString(format)` you would do `NewTime().WithOutputFormat(format)` and pass a
`string` pointer into the `Apply` output. This is made possible because the output for `Apply` is `any`
so changing it to accept a string pointer was trivial.

## RuleSet Methods

`Validate` and `ValidateWithContext` are deprecated. Please switch to `Apply` instead.

## New / Removed Test Helpers

The `MustBeValid`, `MustBeValidAny`, `MustBeValidFunc`, and `MustBeInvalid` test helpers are now removed.
Please use `MustApply`, `MustApplyAny`, `MustApplyFunc` and `MustNotApply` instead.

Note that `MustApply` does not take an `expected` argument, it is assumed to be the same as input. If you wish
to check that the output is different, use `MustApplyMutation` or `MustApplyFunc` instead.

This change was made to reduce redundant code.  vast majority of calls to `MustBeValid` used the same value
for input and output. Additionally, the new names have more consistency with the library new rule set conventions.

Another new test helper `MustNotEvaluate` works similar to `MustNotApply` but for `Rule` implementations.

The new helper `MustApplyTypes` checks to ensure the `Apply` method of a `RuleSet` accepts all the appropriate output
types.
