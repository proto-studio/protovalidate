package validate_test

import (
	"testing"

	"proto.zip/studio/validate"
)

type testStruct struct{}

func TestArray(t *testing.T) {
	ruleSet := validate.Array[any]()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestArrayAny(t *testing.T) {
	ruleSet := validate.ArrayAny()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestConstant(t *testing.T) {
	ruleSet := validate.Constant[int](42)
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestInterface(t *testing.T) {
	ruleSet := validate.Interface[any]()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestInt(t *testing.T) {
	ruleSet := validate.Int()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestUint(t *testing.T) {
	ruleSet := validate.Uint()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestInt8(t *testing.T) {
	ruleSet := validate.Int8()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestUint8(t *testing.T) {
	ruleSet := validate.Uint8()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestInt16(t *testing.T) {
	ruleSet := validate.Int16()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestUint16(t *testing.T) {
	ruleSet := validate.Uint16()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestInt32(t *testing.T) {
	ruleSet := validate.Int32()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestUint32(t *testing.T) {
	ruleSet := validate.Uint32()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestInt64(t *testing.T) {
	ruleSet := validate.Int64()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestUint64(t *testing.T) {
	ruleSet := validate.Uint64()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestFloat32(t *testing.T) {
	ruleSet := validate.Float32()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestFloat64(t *testing.T) {
	ruleSet := validate.Float64()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestMap(t *testing.T) {
	ruleSet := validate.Map[int]()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestMapAny(t *testing.T) {
	ruleSet := validate.MapAny()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestObject(t *testing.T) {
	ruleSet := validate.Object[*testStruct]()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestString(t *testing.T) {
	ruleSet := validate.String()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestDomain(t *testing.T) {
	ruleSet := validate.Domain()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestURI(t *testing.T) {
	ruleSet := validate.URI()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestEmail(t *testing.T) {
	ruleSet := validate.Email()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}

func TestTime(t *testing.T) {
	ruleSet := validate.Time()
	if ruleSet == nil {
		t.Error("Expected rule set to not be nil")
	}
}
