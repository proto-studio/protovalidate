package util

import (
	"strings"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
)

// TestTruncateString tests:
//   - TruncateString returns original string if length <= maxStringDisplayLength
//   - TruncateString truncates and adds ellipsis if length > maxStringDisplayLength
func TestTruncateString(t *testing.T) {
	// Test short string (no truncation)
	short := "short"
	result := TruncateString(short)
	if result != short {
		t.Errorf("Expected %q, got %q", short, result)
	}

	// Test string at max length (no truncation)
	atMax := ""
	for i := 0; i < maxStringDisplayLength; i++ {
		atMax += "a"
	}
	result = TruncateString(atMax)
	if result != atMax {
		t.Errorf("Expected string at max length to not be truncated, got %q", result)
	}

	// Test long string (truncation)
	long := ""
	for i := 0; i < maxStringDisplayLength+10; i++ {
		long += "a"
	}
	result = TruncateString(long)
	if len(result) != maxStringDisplayLength+3 {
		t.Errorf("Expected truncated string length to be %d, got %d", maxStringDisplayLength+3, len(result))
	}
	if result[:maxStringDisplayLength] != long[:maxStringDisplayLength] {
		t.Errorf("Expected truncated string to match prefix, got %q", result)
	}
	if result[maxStringDisplayLength:] != "..." {
		t.Errorf("Expected truncated string to end with ..., got %q", result[maxStringDisplayLength:])
	}
}

// TestFormatErrorMessageLabel tests:
//   - FormatErrorMessageLabel formats with both short and long messages
//   - FormatErrorMessageLabel truncates long strings
func TestFormatErrorMessageLabel(t *testing.T) {
	// Test normal strings
	short := "short"
	long := "long message"
	result := FormatErrorMessageLabel(short, long)
	expected := `WithErrorMessage("short", "long message")`
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Test with truncation
	longStr := ""
	for i := 0; i < maxStringDisplayLength+10; i++ {
		longStr += "a"
	}
	result = FormatErrorMessageLabel(short, longStr)
	if !strings.Contains(result, "...") {
		t.Errorf("Expected result to contain ellipsis for long string, got %q", result)
	}
	if !strings.Contains(result, short) {
		t.Errorf("Expected result to contain short message, got %q", result)
	}
}

// TestFormatStringArgLabel tests:
//   - FormatStringArgLabel formats with method name and value
//   - FormatStringArgLabel truncates long strings
func TestFormatStringArgLabel(t *testing.T) {
	// Test normal string
	methodName := "WithDocsURI"
	value := "https://example.com"
	result := FormatStringArgLabel(methodName, value)
	expected := `WithDocsURI("https://example.com")`
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Test with truncation
	longValue := ""
	for i := 0; i < maxStringDisplayLength+10; i++ {
		longValue += "a"
	}
	result = FormatStringArgLabel(methodName, longValue)
	if !strings.Contains(result, "...") {
		t.Errorf("Expected result to contain ellipsis for long string, got %q", result)
	}
	if !strings.Contains(result, methodName) {
		t.Errorf("Expected result to contain method name, got %q", result)
	}
}

// TestFormatErrorCodeLabel tests:
//   - FormatErrorCodeLabel formats with error code value
func TestFormatErrorCodeLabel(t *testing.T) {
	code := errors.CodeRequired
	result := FormatErrorCodeLabel(code)
	expected := "WithErrorCode(REQUIRED)"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	code = errors.CodePattern
	result = FormatErrorCodeLabel(code)
	expected = "WithErrorCode(PATTERN)"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestFormatErrorMetaLabel tests:
//   - FormatErrorMetaLabel formats with key and string value
//   - FormatErrorMetaLabel formats with key and non-string value
//   - FormatErrorMetaLabel truncates long strings
func TestFormatErrorMetaLabel(t *testing.T) {
	// Test with string value
	key := "key"
	value := "value"
	result := FormatErrorMetaLabel(key, value)
	expected := `WithErrorMeta("key", "value")`
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Test with non-string value
	valueInt := 42
	result = FormatErrorMetaLabel(key, valueInt)
	expected = `WithErrorMeta("key", 42)`
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Test with long string value (truncation)
	longValue := ""
	for i := 0; i < maxStringDisplayLength+10; i++ {
		longValue += "a"
	}
	result = FormatErrorMetaLabel(key, longValue)
	if !strings.Contains(result, "...") {
		t.Errorf("Expected result to contain ellipsis for long string, got %q", result)
	}

	// Test with long key (truncation)
	longKey := ""
	for i := 0; i < maxStringDisplayLength+10; i++ {
		longKey += "k"
	}
	result = FormatErrorMetaLabel(longKey, value)
	if !strings.Contains(result, "...") {
		t.Errorf("Expected result to contain ellipsis for long key, got %q", result)
	}
}

// TestFormatErrorCallbackLabel tests:
//   - FormatErrorCallbackLabel returns the expected format
func TestFormatErrorCallbackLabel(t *testing.T) {
	result := FormatErrorCallbackLabel()
	expected := "WithErrorCallback(<func>)"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}
