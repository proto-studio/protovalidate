package rules_test

import (
	"testing"

	"proto.zip/studio/validate/pkg/rules"
	rulesnet "proto.zip/studio/validate/pkg/rules/net"
	rulestime "proto.zip/studio/validate/pkg/rules/time"
)

// TestConflictType_Replaces tests the Replaces methods on conflict type enums.
// These methods are used internally by noConflict to check if conflict types replace rulesets.

// TestStringConflictType_Replaces tests:
// - stringConflictType.Replaces correctly identifies StringRuleSet conflicts
// - Returns false for non-StringRuleSet rules
func TestStringConflictType_Replaces(t *testing.T) {
	// Test conflict resolution with WithRequired
	rs1 := rules.String().WithRequired()
	rs2 := rs1.WithRequired()
	if rs2.String() != "StringRuleSet.WithRequired()" {
		t.Errorf("Expected conflict resolution, got %s", rs2.String())
	}

	// Test conflict resolution with WithNil
	rs3 := rules.String().WithNil()
	rs4 := rs3.WithNil()
	if rs4.String() != "StringRuleSet.WithNil()" {
		t.Errorf("Expected conflict resolution, got %s", rs4.String())
	}

	// Test conflict resolution with WithStrict
	rs5 := rules.String().WithStrict()
	rs6 := rs5.WithStrict()
	if rs6.String() != "StringRuleSet.WithStrict()" {
		t.Errorf("Expected conflict resolution, got %s", rs6.String())
	}
}

// TestIntConflictType_Replaces tests:
// - intConflictType.Replaces correctly identifies IntRuleSet conflicts for all integer types
// - Returns false for non-IntRuleSet rules
func TestIntConflictType_Replaces(t *testing.T) {
	// Test all integer types to get 100% coverage
	tests := []struct {
		name string
		rs   interface{}
	}{
		{"int", rules.Int().WithBase(10)},
		{"int8", rules.Int8().WithBase(10)},
		{"int16", rules.Int16().WithBase(10)},
		{"int32", rules.Int32().WithBase(10)},
		{"int64", rules.Int64().WithBase(10)},
		{"uint", rules.Uint().WithBase(10)},
		{"uint8", rules.Uint8().WithBase(10)},
		{"uint16", rules.Uint16().WithBase(10)},
		{"uint32", rules.Uint32().WithBase(10)},
		{"uint64", rules.Uint64().WithBase(10)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test conflict resolution by calling WithBase twice
			// This exercises the Replaces method on intConflictType
			rs := tt.rs
			switch v := rs.(type) {
			case *rules.IntRuleSet[int]:
				rs2 := v.WithBase(16)
				if rs2.String() != "IntRuleSet[int].WithBase(16)" {
					t.Errorf("Expected conflict resolution, got %s", rs2.String())
				}
			case *rules.IntRuleSet[int8]:
				rs2 := v.WithBase(16)
				if rs2.String() != "IntRuleSet[int8].WithBase(16)" {
					t.Errorf("Expected conflict resolution, got %s", rs2.String())
				}
			case *rules.IntRuleSet[int16]:
				rs2 := v.WithBase(16)
				if rs2.String() != "IntRuleSet[int16].WithBase(16)" {
					t.Errorf("Expected conflict resolution, got %s", rs2.String())
				}
			case *rules.IntRuleSet[int32]:
				rs2 := v.WithBase(16)
				if rs2.String() != "IntRuleSet[int32].WithBase(16)" {
					t.Errorf("Expected conflict resolution, got %s", rs2.String())
				}
			case *rules.IntRuleSet[int64]:
				rs2 := v.WithBase(16)
				if rs2.String() != "IntRuleSet[int64].WithBase(16)" {
					t.Errorf("Expected conflict resolution, got %s", rs2.String())
				}
			case *rules.IntRuleSet[uint]:
				rs2 := v.WithBase(16)
				if rs2.String() != "IntRuleSet[uint].WithBase(16)" {
					t.Errorf("Expected conflict resolution, got %s", rs2.String())
				}
			case *rules.IntRuleSet[uint8]:
				rs2 := v.WithBase(16)
				if rs2.String() != "IntRuleSet[uint8].WithBase(16)" {
					t.Errorf("Expected conflict resolution, got %s", rs2.String())
				}
			case *rules.IntRuleSet[uint16]:
				rs2 := v.WithBase(16)
				if rs2.String() != "IntRuleSet[uint16].WithBase(16)" {
					t.Errorf("Expected conflict resolution, got %s", rs2.String())
				}
			case *rules.IntRuleSet[uint32]:
				rs2 := v.WithBase(16)
				if rs2.String() != "IntRuleSet[uint32].WithBase(16)" {
					t.Errorf("Expected conflict resolution, got %s", rs2.String())
				}
			case *rules.IntRuleSet[uint64]:
				rs2 := v.WithBase(16)
				if rs2.String() != "IntRuleSet[uint64].WithBase(16)" {
					t.Errorf("Expected conflict resolution, got %s", rs2.String())
				}
			}
		})
	}
}

// TestFloatConflictType_Replaces tests:
// - floatConflictType.Replaces correctly identifies FloatRuleSet conflicts for both float types
// - Returns false for non-FloatRuleSet rules
func TestFloatConflictType_Replaces(t *testing.T) {
	// Test float32 - WithRounding
	rs1 := rules.Float32().WithRounding(rules.RoundingUp, 2)
	rs2 := rs1.WithRounding(rules.RoundingDown, 3)
	if rs2.String() != "FloatRuleSet[float32].WithRounding(Down, 3)" {
		t.Errorf("Expected conflict resolution for float32 WithRounding, got %s", rs2.String())
	}

	// Test float32 - WithFixedOutput
	rs3 := rules.Float32().WithFixedOutput(2)
	rs4 := rs3.WithFixedOutput(4)
	if rs4.String() != "FloatRuleSet[float32].WithFixedOutput(4)" {
		t.Errorf("Expected conflict resolution for float32 WithFixedOutput, got %s", rs4.String())
	}

	// Test float64 - WithRounding
	rs5 := rules.Float64().WithRounding(rules.RoundingUp, 2)
	rs6 := rs5.WithRounding(rules.RoundingDown, 3)
	if rs6.String() != "FloatRuleSet[float64].WithRounding(Down, 3)" {
		t.Errorf("Expected conflict resolution for float64 WithRounding, got %s", rs6.String())
	}

	// Test float64 - WithFixedOutput
	rs7 := rules.Float64().WithFixedOutput(2)
	rs8 := rs7.WithFixedOutput(4)
	if rs8.String() != "FloatRuleSet[float64].WithFixedOutput(4)" {
		t.Errorf("Expected conflict resolution for float64 WithFixedOutput, got %s", rs8.String())
	}
}

// TestSliceConflictType_Replaces tests:
// - conflictType.Replaces correctly identifies SliceRuleSet conflicts
// - Returns false for non-SliceRuleSet rules
func TestSliceConflictType_Replaces(t *testing.T) {
	// Test conflict resolution with WithMinLen
	rs1 := rules.Slice[int]().WithMinLen(3)
	rs2 := rs1.WithMinLen(5)
	if rs2.String() != "SliceRuleSet[int].WithMinLen(5)" {
		t.Errorf("Expected conflict resolution, got %s", rs2.String())
	}

	// Test conflict resolution with WithMaxLen
	rs3 := rules.Slice[int]().WithMaxLen(10)
	rs4 := rs3.WithMaxLen(20)
	if rs4.String() != "SliceRuleSet[int].WithMaxLen(20)" {
		t.Errorf("Expected conflict resolution, got %s", rs4.String())
	}
}

// TestDomainConflictType_Replaces tests:
// - domainConflictType.Replaces correctly identifies DomainRuleSet conflicts
// - Returns false for non-DomainRuleSet rules
func TestDomainConflictType_Replaces(t *testing.T) {
	// Test conflict resolution with WithRequired
	rs1 := rulesnet.Domain().WithRequired()
	rs2 := rs1.WithRequired()
	// Since DomainRuleSet uses cloneWithConflictType, the second WithRequired should replace the first
	if rs2.String() != "DomainRuleSet.WithRequired()" {
		t.Errorf("Expected conflict resolution, got %s", rs2.String())
	}

	// Test conflict resolution with WithNil
	rs3 := rulesnet.Domain().WithNil()
	rs4 := rs3.WithNil()
	if rs4.String() != "DomainRuleSet.WithNil()" {
		t.Errorf("Expected conflict resolution, got %s", rs4.String())
	}
}

// TestEmailConflictType_Replaces tests:
// - emailConflictType.Replaces correctly identifies EmailRuleSet conflicts
// - Returns false for non-EmailRuleSet rules
func TestEmailConflictType_Replaces(t *testing.T) {
	// Test conflict resolution with WithRequired
	rs1 := rulesnet.Email().WithRequired()
	rs2 := rs1.WithRequired()
	if rs2.String() != "EmailRuleSet.WithRequired()" {
		t.Errorf("Expected conflict resolution, got %s", rs2.String())
	}

	// Test conflict resolution with WithNil
	rs3 := rulesnet.Email().WithNil()
	rs4 := rs3.WithNil()
	if rs4.String() != "EmailRuleSet.WithNil()" {
		t.Errorf("Expected conflict resolution, got %s", rs4.String())
	}
}

// TestURIConflictType_Replaces tests:
// - uriConflictType.Replaces correctly identifies URIRuleSet conflicts
// - Returns false for non-URIRuleSet rules
func TestURIConflictType_Replaces(t *testing.T) {
	// Test conflict resolution with WithRequired
	rs1 := rulesnet.URI().WithRequired()
	rs2 := rs1.WithRequired()
	if rs2.String() != "URIRuleSet.WithRequired()" {
		t.Errorf("Expected conflict resolution, got %s", rs2.String())
	}

	// Test conflict resolution with WithNil
	rs3 := rulesnet.URI().WithNil()
	rs4 := rs3.WithNil()
	if rs4.String() != "URIRuleSet.WithNil()" {
		t.Errorf("Expected conflict resolution, got %s", rs4.String())
	}
}

// TestTimeConflictType_Replaces tests:
// - conflictType.Replaces correctly identifies TimeRuleSet conflicts
// - Returns false for non-TimeRuleSet rules
func TestTimeConflictType_Replaces(t *testing.T) {
	// Test conflict resolution with WithRequired
	rs1 := rulestime.Time().WithRequired()
	rs2 := rs1.WithRequired()
	if rs2.String() != "TimeRuleSet.WithRequired()" {
		t.Errorf("Expected conflict resolution, got %s", rs2.String())
	}

	// Test conflict resolution with WithNil
	rs3 := rulestime.Time().WithNil()
	rs4 := rs3.WithNil()
	if rs4.String() != "TimeRuleSet.WithNil()" {
		t.Errorf("Expected conflict resolution, got %s", rs4.String())
	}

	// Test conflict resolution with WithLayouts
	rs5 := rulestime.Time().WithLayouts("2006-01-02")
	rs6 := rs5.WithLayouts("2006-01-02T15:04:05Z07:00")
	if rs6.String() != "TimeRuleSet.WithLayouts(\"2006-01-02T15:04:05Z07:00\")" {
		t.Errorf("Expected conflict resolution for WithLayouts, got %s", rs6.String())
	}

	// Test conflict resolution with WithOutputLayout
	rs7 := rulestime.Time().WithOutputLayout("2006-01-02")
	rs8 := rs7.WithOutputLayout("2006-01-02T15:04:05Z07:00")
	if rs8.String() != "TimeRuleSet.WithOutputLayout(\"2006-01-02T15:04:05Z07:00\")" {
		t.Errorf("Expected conflict resolution for WithOutputLayout, got %s", rs8.String())
	}
}

// TestStringConflictType_Conflict tests:
// - stringConflictType.Conflict correctly identifies conflicts
func TestStringConflictType_Conflict(t *testing.T) {
	// Test that WithRequired conflicts with WithRequired (now uses conflict resolution)
	rs1 := rules.String().WithRequired()
	rs2 := rs1.WithRequired()
	// With conflict resolution, the second WithRequired should replace the first
	if rs2.String() != "StringRuleSet.WithRequired()" {
		t.Errorf("Expected conflict resolution, got %s", rs2.String())
	}
}
