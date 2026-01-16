package main

import (
	"context"
	"fmt"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

func main() {
	fmt.Println("=== Path Serialization Examples ===\n")

	// Example 1: Simple nested path
	fmt.Println("Example 1: Simple nested path (users.profile.name)")
	example1()

	// Example 2: Path with array indices
	fmt.Println("\nExample 2: Path with array indices (users[0].emails[1])")
	example2()

	// Example 3: Complex mixed path
	fmt.Println("\nExample 3: Complex mixed path (data.items[2].metadata.tags[0])")
	example3()

	// Example 4: Path starting with index
	fmt.Println("\nExample 4: Path starting with index ([5].value)")
	example4()

	// Example 5: Single segment
	fmt.Println("\nExample 5: Single segment (username)")
	example5()
}

func example1() {
	ctx := rulecontext.WithPathString(context.Background(), "users")
	ctx = rulecontext.WithPathString(ctx, "profile")
	ctx = rulecontext.WithPathString(ctx, "name")
	err := errors.Errorf(errors.CodeMin, ctx, "below minimum", "must be at least %d", 10)

	printAllFormats(err)
}

func example2() {
	ctx := rulecontext.WithPathString(context.Background(), "users")
	ctx = rulecontext.WithPathIndex(ctx, 0)
	ctx = rulecontext.WithPathString(ctx, "emails")
	ctx = rulecontext.WithPathIndex(ctx, 1)
	err := errors.Errorf(errors.CodeMin, ctx, "below minimum", "must be at least %d", 10)

	printAllFormats(err)
}

func example3() {
	ctx := rulecontext.WithPathString(context.Background(), "data")
	ctx = rulecontext.WithPathString(ctx, "items")
	ctx = rulecontext.WithPathIndex(ctx, 2)
	ctx = rulecontext.WithPathString(ctx, "metadata")
	ctx = rulecontext.WithPathString(ctx, "tags")
	ctx = rulecontext.WithPathIndex(ctx, 0)
	err := errors.Errorf(errors.CodeMin, ctx, "below minimum", "must be at least %d", 10)

	printAllFormats(err)
}

func example4() {
	ctx := rulecontext.WithPathIndex(context.Background(), 5)
	ctx = rulecontext.WithPathString(ctx, "value")
	err := errors.Errorf(errors.CodeMin, ctx, "below minimum", "must be at least %d", 10)

	printAllFormats(err)
}

func example5() {
	ctx := rulecontext.WithPathString(context.Background(), "username")
	err := errors.Errorf(errors.CodeMin, ctx, "below minimum", "must be at least %d", 10)

	printAllFormats(err)
}

func printAllFormats(err errors.ValidationError) {
	defaultSerializer := errors.DefaultPathSerializer{}
	jsonPointerSerializer := errors.JSONPointerSerializer{}
	jsonPathSerializer := errors.JSONPathSerializer{}
	dotNotationSerializer := errors.DotNotationSerializer{}

	fmt.Printf("  Default (Path()):        %s\n", err.Path())
	fmt.Printf("  Default (PathAs):        %s\n", err.PathAs(defaultSerializer))
	fmt.Printf("  JSON Pointer (RFC 6901): %s\n", err.PathAs(jsonPointerSerializer))
	fmt.Printf("  JSONPath:                %s\n", err.PathAs(jsonPathSerializer))
	fmt.Printf("  Dot Notation:            %s\n", err.PathAs(dotNotationSerializer))
}
