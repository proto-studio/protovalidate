// A simple example of using rule sets to validate strings.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"proto.zip/studio/validate"
	"proto.zip/studio/validate/pkg/rules"
)

var ruleSet rules.RuleSet[string] = validate.String().
	WithMinLen(3).
	WithMaxLen(7)

// checkAll iterates over an array of strings and calls the rule set for each one.
func checkAll(w io.Writer, str ...string) {
	if len(str) == 0 {
		fmt.Fprintf(w, "Enter 1 or more strings on the command line.\n")
		return
	}

	for _, s := range str {
		var output string

		// Use Apply instead of Run to validate the string
		err := ruleSet.Apply(context.TODO(), s, &output)
		if err == nil {
			fmt.Fprintf(w, "'%s' is valid\n", s)
		} else {
			fmt.Fprintf(w, "'%s' is invalid: %s\n", s, err)
		}
	}
}

// Try changing the string to see different results.
func main() {
	flag.Parse()
	checkAll(os.Stdout, flag.Args()...)
}
