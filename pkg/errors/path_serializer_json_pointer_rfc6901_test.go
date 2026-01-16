package errors_test

import (
	"context"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
)

// TestJSONPointerSerializer_RFC6901Compliance tests:
// - JSON Pointer serializer correctly implements RFC 6901 escaping rules
// RFC 6901 specifies:
//   - "~" must be encoded as "~0"
//   - "/" must be encoded as "~1"
//   - These are the only two escape sequences
func TestJSONPointerSerializer_RFC6901Compliance(t *testing.T) {
	tests := []struct {
		name     string
		segment  string
		expected string
		desc     string
	}{
		{
			name:     "RFC 6901 example: m~n",
			segment:  "m~n",
			expected: "/m~0n",
			desc:     "From RFC 6901: key 'm~n' should serialize to '/m~0n'",
		},
		{
			name:     "RFC 6901 example: a/b",
			segment:  "a/b",
			expected: "/a~1b",
			desc:     "From RFC 6901: key 'a/b' should serialize to '/a~1b'",
		},
		{
			name:     "literal ~0 in input",
			segment:  "a~0b",
			expected: "/a~00b",
			desc:     "Literal '~0' should become '~00' (escaped tilde + 0)",
		},
		{
			name:     "literal ~1 in input",
			segment:  "a~1b",
			expected: "/a~01b",
			desc:     "Literal '~1' should become '~01' (escaped tilde + 1)",
		},
		{
			name:     "multiple tildes",
			segment:  "a~~b",
			expected: "/a~0~0b",
			desc:     "Multiple tildes should each be escaped",
		},
		{
			name:     "multiple slashes",
			segment:  "a//b",
			expected: "/a~1~1b",
			desc:     "Multiple slashes should each be escaped",
		},
		{
			name:     "tilde then slash",
			segment:  "a~/b",
			expected: "/a~0~1b",
			desc:     "Tilde then slash: '~' becomes '~0', '/' becomes '~1'",
		},
		{
			name:     "slash then tilde",
			segment:  "a/~b",
			expected: "/a~1~0b",
			desc:     "Slash then tilde: '/' becomes '~1', '~' becomes '~0'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := rulecontext.WithPathString(context.Background(), tt.segment)
			err := errors.Errorf(errors.CodeMin, ctx, "short", "message")

			serializer := errors.JSONPointerSerializer{}
			path := err.PathAs(serializer)

			if path != tt.expected {
				t.Errorf("%s\nExpected: '%s'\nGot:      '%s'", tt.desc, tt.expected, path)
			}
		})
	}
}
