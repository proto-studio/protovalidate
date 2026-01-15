// A simple example of using translations for your rules.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	_ "proto.zip/studio/validate/_examples/i18n/translations"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

// standardRuleSet uses built-in messages from the error dictionary.
// These messages are automatically translatable via i18n.
var standardRuleSet = rules.String().
	WithMinLen(3).
	WithMaxLen(7)

// customRuleSet demonstrates using WithErrorMessage for custom error messages.
// Custom messages are also translatable via i18n when using the printer from context.
var customRuleSet = rules.String().
	WithMinLen(5).
	WithErrorMessage("too short", "username must be at least %d characters")

func checkAll(w io.Writer, locale *string, str ...string) {
	lang := language.MustParse(*locale)
	printer := message.NewPrinter(lang)

	if len(str) == 0 {
		printer.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		printer.Fprintf(os.Stderr, "\nTrailing arguments:\n")
		printer.Fprintf(os.Stderr, "  [strings ...]  list of strings to validate.\n")
		os.Exit(1)
	}

	ctx := rulecontext.WithPrinter(context.Background(), printer)

	var output string

	printer.Fprintf(w, "\n=== Standard Messages (from error dictionary) ===\n")
	printer.Fprintf(w, "Rule: string with length 3-7 characters\n\n")

	for _, s := range str {
		err := standardRuleSet.Apply(ctx, s, &output)
		if err == nil {
			printer.Fprintf(w, "'%s' is valid\n", s)
		} else {
			fmt.Fprintf(w, "'%s' is invalid: %s\n", s, err)
		}
	}

	printer.Fprintf(w, "\n=== Custom Messages (using WithErrorMessage) ===\n")
	printer.Fprintf(w, "Rule: username with minimum 5 characters\n\n")

	for _, s := range str {
		err := customRuleSet.Apply(ctx, s, &output)
		if err == nil {
			printer.Fprintf(w, "'%s' is valid\n", s)
		} else {
			fmt.Fprintf(w, "'%s' is invalid: %s\n", s, err)
		}
	}
}

// Try changing the string to see different results.
func main() {
	locale := flag.String("locale", "en-US", "Locale to display messages in.")

	flag.Parse()
	checkAll(os.Stdout, locale, flag.Args()...)
}
