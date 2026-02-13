package rules

import (
	"testing"
)

// TestFormatInt tests formatInt for coverage (used when formatting int for error messages/output).
func TestFormatInt(t *testing.T) {
	tests := []struct {
		name  string
		value any
		base  int
		want  string
	}{
		{"int decimal", 42, 10, "42"},
		{"int hex", 255, 16, "ff"},
		{"int8", int8(100), 10, "100"},
		{"int64 negative", int64(-1), 10, "-1"},
		{"uint", uint(99), 10, "99"},
		{"uint32 hex", uint32(0xAB), 16, "ab"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.value.(type) {
			case int:
				got := formatInt(v, tt.base)
				if got != tt.want {
					t.Errorf("formatInt(%v, %d) = %q, want %q", v, tt.base, got, tt.want)
				}
			case int8:
				got := formatInt(v, tt.base)
				if got != tt.want {
					t.Errorf("formatInt(%v, %d) = %q, want %q", v, tt.base, got, tt.want)
				}
			case int64:
				got := formatInt(v, tt.base)
				if got != tt.want {
					t.Errorf("formatInt(%v, %d) = %q, want %q", v, tt.base, got, tt.want)
				}
			case uint:
				got := formatInt(v, tt.base)
				if got != tt.want {
					t.Errorf("formatInt(%v, %d) = %q, want %q", v, tt.base, got, tt.want)
				}
			case uint32:
				got := formatInt(v, tt.base)
				if got != tt.want {
					t.Errorf("formatInt(%v, %d) = %q, want %q", v, tt.base, got, tt.want)
				}
			}
		})
	}
}

// TestFormatFloat tests formatFloat for coverage.
func TestFormatFloat(t *testing.T) {
	// No WithFixedOutput: uses default 'g' format
	rs := Float64()
	got := formatFloat(rs, 123.456)
	if got == "" {
		t.Error("formatFloat: expected non-empty string")
	}

	// WithFixedOutput
	rs2 := Float64().WithFixedOutput(2)
	got2 := formatFloat(rs2, 123.456)
	if got2 != "123.46" {
		t.Errorf("formatFloat with FixedOutput(2) = %q, want 123.46", got2)
	}

	// WithRounding (roundingPrecision path)
	rs3 := Float64().WithRounding(RoundingHalfUp, 2)
	got3 := formatFloat(rs3, 123.456)
	if got3 == "" {
		t.Error("formatFloat with rounding: expected non-empty string")
	}

	// Float32 with WithFixedOutput (outputPrecision path, bits=32)
	rs32 := Float32().WithFixedOutput(1)
	got32 := formatFloat(rs32, float32(99.99))
	if got32 != "100.0" {
		t.Errorf("formatFloat float32 = %q, want 100.0", got32)
	}

	// Float32 default 'g' format (bits == 32 branch: sigDigits = 7)
	rs32Default := Float32()
	got32Default := formatFloat(rs32Default, float32(123.456))
	if got32Default == "" {
		t.Error("formatFloat float32 default: expected non-empty string")
	}
}

// TestTrimTrailingZeros tests trimTrailingZeros for coverage.
func TestTrimTrailingZeros(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"123.46000", "123.46"},
		{"123.000", "123"},
		{"123", "123"},
		{"0.0", "0"},
		{"1.5", "1.5"},
	}
	for _, tt := range tests {
		got := trimTrailingZeros(tt.in)
		if got != tt.want {
			t.Errorf("trimTrailingZeros(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
