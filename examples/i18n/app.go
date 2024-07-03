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
	"proto.zip/studio/validate"
	_ "proto.zip/studio/validate/examples/i18n/translations"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules"
)

var ruleSet rules.RuleSet[string] = validate.String().
	WithMinLen(3).
	WithMaxLen(7)

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

	for _, s := range str {
		_, err := ruleSet.Run(ctx, s)
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
