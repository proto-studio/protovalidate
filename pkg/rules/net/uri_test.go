package net_test

import (
	"context"
	"fmt"
	"testing"

	"proto.zip/studio/validate/pkg/errors"
	"proto.zip/studio/validate/pkg/rulecontext"
	"proto.zip/studio/validate/pkg/rules/net"
	"proto.zip/studio/validate/pkg/testhelpers"
)

// uriPartRequiredMissingHelper evaluates a single missing value test
func uriPartRequiredMissingHelper(t testing.TB, name, value string, withRequired *net.URIRuleSet) {
	t.Helper()
	withDeepErrors := withRequired.WithDeepErrors()

	ctx := rulecontext.WithPathString(context.Background(), "uri")

	// Prepare the output variable for Apply
	var output string

	// Use Apply for the shallow error check
	err := withRequired.Apply(ctx, value, &output)

	if err == nil {
		t.Errorf("Expected shallow error to not be nil on %s", value)
	} else if code := err.First().Code(); code != errors.CodeRequired {
		t.Errorf("Expected shallow error code of %s, got %s (%s)", errors.CodeRequired, code, err)
	} else if path := err.First().Path(); path != "/uri" {
		t.Errorf("Expected shallow error path of %s, got %s (on %s)", "/uri/"+name, path, value)
	}

	// Use Apply for the deep error check
	err = withDeepErrors.Apply(ctx, value, &output)

	if err == nil {
		t.Errorf("Expected deep error to not be nil on %s", value)
	} else if code := err.First().Code(); code != errors.CodeRequired {
		t.Errorf("Expected deep error code of %s, got %s (%s on %s)", errors.CodeRequired, code, err, value)
	} else if path := err.First().Path(); path != "/uri/"+name {
		t.Errorf("Expected deep error path of %s, got %s (on %s)", "/uri/"+name, path, value)
	}
}

// uriPartRequiredHelper takes in two rule set (required and not) and three values:
// - valid
// - empty
// - missing
// It checks to make sure the following conditions are met:
// - If valid, always pass
// - If empty, always pass
// - If missing, only pass if it is not required
func uriPartRequiredHelper(t testing.TB, fnName, name string, withoutRequired, withRequired *net.URIRuleSet, identityCheck bool, valid, empty, missing string, additionalMissing ...string) {
	t.Helper()

	withRequiredAny := withRequired.Any()
	withoutRequiredAny := withoutRequired.Any()

	if !identityCheck {
		t.Errorf("Expected subsequent calls to %s to return the same RuleSet", fnName)
	}

	const expectedStringA = "URIRuleSet"
	if actual := withoutRequired.String(); expectedStringA != actual {
		t.Errorf("Expected String() without required to be `%s`, got: `%s`", expectedStringA, actual)
	}

	expectedStringB := fmt.Sprintf("URIRuleSet.%s()", fnName)
	if actual := withRequired.String(); expectedStringB != actual {
		t.Errorf("Expected String() with required to be `%s`, got: `%s`", expectedStringB, actual)
	}

	testhelpers.MustApply(t, withoutRequiredAny, valid)
	testhelpers.MustApply(t, withRequiredAny, valid)

	testhelpers.MustApply(t, withoutRequiredAny, empty)
	testhelpers.MustApply(t, withRequiredAny, empty)

	testhelpers.MustApply(t, withoutRequiredAny, missing)
	uriPartRequiredMissingHelper(t, name, missing, withRequired)

	for _, v := range additionalMissing {
		testhelpers.MustApply(t, withoutRequiredAny, v)
		uriPartRequiredMissingHelper(t, name, v, withRequired)
	}
}

// Requirements:
// - Default configuration doesn't return errors on valid value.
// - Implements interface.
func TestURIRuleSet(t *testing.T) {
	// Prepare the output variable for Apply
	var output string

	example := "https://example.com"

	// Use Apply instead of Run
	err := net.NewURI().Apply(context.TODO(), example, &output)

	if err != nil {
		t.Errorf("Expected errors to be empty, got: %s", err)
		return
	}

	if output != example {
		t.Error("Expected test URI to be returned")
		return
	}

	// Check if the rule set implements the expected interface
	ok := testhelpers.CheckRuleSetInterface[string](net.NewURI())
	if !ok {
		t.Error("Expected rule set to be implemented")
		return
	}

	testhelpers.MustApplyTypes[string](t, net.NewURI(), example)
}

// Requirements:
// - Required flag can be set.
// - Required flag can be read.
// - Required flag defaults to false.
// - Calling WithRequired on a rule set that already has it returns the identity.
func TestURIRequired(t *testing.T) {
	ruleSet := net.NewURI()

	if ruleSet.Required() {
		t.Error("Expected rule set to not be required")
	}

	ruleSet = ruleSet.WithRequired()

	if !ruleSet.Required() {
		t.Error("Expected rule set to be required")
	}

	ruleSet2 := ruleSet.WithRequired()

	if ruleSet2 != ruleSet {
		t.Error("Expected WithRequired to be idempotent")
	}
}

// Requirements:
// - Returns a coercion error if input is not a string.
func TestURICoercionFromUknown(t *testing.T) {
	val := new(struct {
		x int
	})

	testhelpers.MustNotApply(t, net.NewURI().Any(), &val, errors.CodeType)
}

// Requirements:
// - Scheme must start with a letter.
// - Scheme can contain . - and +.
func TestURISchemeCharacterSet(t *testing.T) {
	ruleSet := net.NewURI().Any()

	testhelpers.MustApply(t, ruleSet, "test://hello")
	testhelpers.MustApply(t, ruleSet, "test123://hello")
	testhelpers.MustApply(t, ruleSet, "test-123://hello")
	testhelpers.MustApply(t, ruleSet, "test.123://hello")
	testhelpers.MustApply(t, ruleSet, "test+123://hello")

	testhelpers.MustNotApply(t, ruleSet, "1test://hello", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "+test://hello", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, "-test://hello", errors.CodePattern)
	testhelpers.MustNotApply(t, ruleSet, ".test://hello", errors.CodePattern)
}

// Requirements:
// - Custom rules are validated and return errors.
// - Custom rule context contains URI parts:
//   - scheme
//   - authority
//   - path
//   - userinfo
//   - query
//   - fragment
//   - userinfo
//   - port
//   - user
//   - password
func TestURICustomContext(t *testing.T) {
	var ctxRef context.Context

	fn := func(ctx context.Context, value string) errors.ValidationErrorCollection {
		ctxRef = ctx
		return nil
	}

	ruleSet := net.NewURI().WithRuleFunc(fn)

	const testScheme = "https"
	const testHost = "example.com"
	const testPort = "8080"
	const testPath = "/test/path"
	const testQuery = "x=123"
	const testFragment = "section1"
	const testUser = "me"
	const testPassword = "qwerty"

	testUserinfo := fmt.Sprintf("%s:%s", testUser, testPassword)
	testAuthority := fmt.Sprintf("%s@%s:%s", testUserinfo, testHost, testPort)

	var output string
	err := ruleSet.Apply(context.TODO(), fmt.Sprintf("%s://%s%s?%s#%s", testScheme, testAuthority, testPath, testQuery, testFragment), &output)
	if err != nil {
		t.Fatalf("Expected error to not be nil, got: %s", err)
	}

	if ctxRef == nil {
		t.Fatal("Expected context to not be nil")
	}

	scheme := ctxRef.Value("scheme")
	if scheme == nil || scheme.(string) != testScheme {
		t.Errorf("Expected scheme to be `%s`, got `%s`", testScheme, scheme)
	}

	authority := ctxRef.Value("authority")
	if authority == nil || authority.(string) != testAuthority {
		t.Errorf("Expected authority to be `%s`, got `%s`", testAuthority, authority)
	}

	userinfo := ctxRef.Value("userinfo")
	if userinfo == nil || userinfo.(string) != testUserinfo {
		t.Errorf("Expected userinfo to be `%s`, got `%s`", testUserinfo, userinfo)
	}

	user := ctxRef.Value("user")
	if user == nil || user.(string) != testUser {
		t.Errorf("Expected user to be `%s`, got `%s`", testUser, user)
	}

	password := ctxRef.Value("password")
	if password == nil || password.(string) != testPassword {
		t.Errorf("Expected password to be `%s`, got `%s`", testPassword, password)
	}

	host := ctxRef.Value("host")
	if host == nil || host.(string) != testHost {
		t.Errorf("Expected host to be `%s`, got `%s`", testHost, host)
	}

	port := ctxRef.Value("port")
	if port == nil || port.(string) != testPort {
		t.Errorf("Expected port to be `%s`, got `%s`", testPort, port)
	}

	path := ctxRef.Value("path")
	if path == nil || path.(string) != testPath {
		t.Errorf("Expected path to be `%s`, got `%s`", testPath, path)
	}

	query := ctxRef.Value("query")
	if query == nil || query.(string) != testQuery {
		t.Errorf("Expected query to be `%s`, got `%s`", testQuery, query)
	}

	fragment := ctxRef.Value("fragment")
	if fragment == nil || fragment.(string) != testFragment {
		t.Errorf("Expected fragment to be `%s`, got `%s`", testFragment, fragment)
	}
}

// Requirements:
// - No negative ports.
// - No out of range ports.
// - Port must be a number.
func TestURIPort(t *testing.T) {
	ruleSet := net.NewURI().Any()

	testhelpers.MustNotApply(t, ruleSet, "https://example:-1", errors.CodeMin)
	testhelpers.MustNotApply(t, ruleSet, "https://example:65536", errors.CodeMax)
	testhelpers.MustNotApply(t, ruleSet, "https://example:notaport", errors.CodeType)
}

// Requirements:
// - When the deep errors flag is not set, all errors return the same path.
// - When the deep errors flag is set, all errors return a nested path.
// - Calling WithDeepErrors on a rule set that already has it returns the identity.
func TestURIDeepErrors(t *testing.T) {
	tests := map[string]string{
		"scheme":   "%://example.com",
		"user":     "https://%:@example.com",
		"password": "https://me:%@example.com",
		"host":     "https://%",
		"port":     "https://example.com:-1",
		"path":     "https://example.com/%",
		"query":    "https://example.com/?%",
		"fragment": "https://example.com/#%",
	}

	var output string

	ruleSet := net.NewURI()
	ctx := rulecontext.WithPathString(context.Background(), "url")

	if ruleSet.DeepErrors() {
		t.Errorf("Expected deep error to be false")
	}

	for path, value := range tests {
		errs := ruleSet.Apply(ctx, value, &output)

		if len(errs) != 1 {
			t.Errorf("Expected 1 error for %s, got: %d", path, len(errs))
		} else if errPath := errs.First().Path(); errPath != "/url" {
			t.Errorf("Expected path for %s to be `/url`, got: %s", path, errPath)
		}
	}

	ruleSet = ruleSet.WithDeepErrors()

	if !ruleSet.DeepErrors() {
		t.Errorf("Expected deep error to be true")
	}

	ruleSet2 := ruleSet.WithDeepErrors()
	if ruleSet != ruleSet2 {
		t.Errorf("Expected WithDeepErrors to be idempotent")
	}

	for path, value := range tests {
		errs := ruleSet.Apply(ctx, value, &output)

		if len(errs) != 1 {
			// We would have already printed this error
		} else if errPath := errs.First().Path(); errPath != "/url/"+path {
			t.Errorf("Expected path for %s to be `/url/%s`, got: %s", path, path, errPath)
		}
	}
}

// Requirements:
// - Relative flag can be set.
// - Relative flag can be read.
// - Relative flag defaults to false.
// - Calling WithRelative on a rule set that already has it returns the identity.
func TestURIRelative(t *testing.T) {
	ruleSet := net.NewURI()

	if ruleSet.Relative() {
		t.Error("Expected rule set to not allow relative URIs")
	}

	ruleSet = ruleSet.WithRelative()

	if !ruleSet.Relative() {
		t.Error("Expected rule set to allow relative URIs")
	}

	ruleSet2 := ruleSet.WithRelative()

	if ruleSet2 != ruleSet {
		t.Error("Expected WithRelative to be idempotent")
	}
}

// Requirement:
// - Only relative URIs can be zero length.
func TestURIZeroLength(t *testing.T) {
	ruleSet := net.NewURI()

	testhelpers.MustNotApply(t, ruleSet.Any(), "", errors.CodeRequired)

	ruleSet = ruleSet.WithRelative()

	testhelpers.MustApply(t, ruleSet.Any(), "")
}

// Requirement:
// - User can be required.
// - User can be empty even when required.
func TestURIWithUserRequired(t *testing.T) {
	withoutRequired := net.NewURI()
	withRequired := withoutRequired.WithUserRequired()
	withRequiredB := withRequired.WithUserRequired()

	uriPartRequiredHelper(
		t,
		"WithUserRequired",
		"user",
		withoutRequired,
		withRequired,
		withRequired == withRequiredB,
		"http://user:qwerty@example.com",
		"http://:qwerty@example.com",
		"http://example.com",
		"http:",
	)
}

// Requirement:
// - Password can be required.
// - Password can be empty even when required.
func TestURIWithPassword(t *testing.T) {
	withoutRequired := net.NewURI()
	withRequired := withoutRequired.WithPasswordRequired()
	withRequiredB := withRequired.WithPasswordRequired()

	uriPartRequiredHelper(
		t,
		"WithPasswordRequired",
		"password",
		withoutRequired,
		withRequired,
		withRequired == withRequiredB,
		"http://me:qwerty@example.com",
		"http://me:@example.com",
		"http://me@example.com",
		"http://example.com",
		"http:e",
		"http:",
	)
}

// Requirement:
// - Host can be required.
// - Host can be empty even when required.
func TestURIWithHostRequired(t *testing.T) {
	withoutRequired := net.NewURI()
	withRequired := withoutRequired.WithHostRequired()
	withRequiredB := withRequired.WithHostRequired()

	uriPartRequiredHelper(
		t,
		"WithHostRequired",
		"host",
		withoutRequired,
		withRequired,
		withRequired == withRequiredB,
		"http://example.com",
		"http://example.com",
		"http:e",
		"http:",
	)

	testhelpers.MustApply(t, withRequired.Any(), "http://")
}

// Requirement:
// - Port can be required.
func TestURIWithPortRequired(t *testing.T) {
	withoutRequired := net.NewURI()
	withRequired := withoutRequired.WithPortRequired()
	withRequiredB := withRequired.WithPortRequired()

	uriPartRequiredHelper(
		t,
		"WithPortRequired",
		"port",
		withoutRequired,
		withRequired,
		withRequired == withRequiredB,
		"http://example.com:8080",
		"http://example.com:0", // Empty will trigger an int conversion error so we can't test it with this helper
		"http://example.com",
		"http:e",
		"http:",
	)
}

// Requirement:
// - Query can be required.
// - Query can be empty even when required.
func TestURIWithQueryRequired(t *testing.T) {
	withoutRequired := net.NewURI()
	withRequired := withoutRequired.WithQueryRequired()
	withRequiredB := withRequired.WithQueryRequired()

	uriPartRequiredHelper(
		t,
		"WithQueryRequired",
		"query",
		withoutRequired,
		withRequired,
		withRequired == withRequiredB,
		"http://example.com?query=1",
		"http://example.com?",
		"http://example.com",
		"http:e",
		"http:",
	)
}

// Requirement:
// - Fragment can be required.
// - Fragment can be empty even when required.
func TestURIWithFragmentRequired(t *testing.T) {
	withoutRequired := net.NewURI()
	withRequired := withoutRequired.WithFragmentRequired()
	withRequiredB := withRequired.WithFragmentRequired()

	uriPartRequiredHelper(
		t,
		"WithFragmentRequired",
		"fragment",
		withoutRequired,
		withRequired,
		withRequired == withRequiredB,
		"http://example.com#fragment",
		"http://example.com#",
		"http://example.com",
		"http:e",
		"http:",
	)
}

// Requirement:
// - Bad URI escaping should cause an error.
// - Valid escaped URIs should pass validation.
func TestURIEscaping(t *testing.T) {
	ruleSet := net.NewURI()

	// Valid
	testhelpers.MustApply(t, ruleSet.Any(), "http://example.com/hello%20world")

	// Strings ends exactly on two hex characters
	testhelpers.MustApply(t, ruleSet.Any(), "http://example.com/hello%20")

	// String ends before reading two characters
	testhelpers.MustNotApply(t, ruleSet.Any(), "http://example.com/hello%2", errors.CodeEncoding)

	// Invalid hex for second character
	testhelpers.MustNotApply(t, ruleSet.Any(), "http://example.com/hello%2Zworld", errors.CodeEncoding)

	// Invalid hex for both characters
	testhelpers.MustNotApply(t, ruleSet.Any(), "http://example.com/hello%ZZworld", errors.CodeEncoding)
}

// Requirements
// - Custom validation rules are called.
// - All errors are returned.
func TestURICustom(t *testing.T) {
	testVal := "https://example.com"

	mock := testhelpers.NewMockRuleWithErrors[string](1)

	var output string
	err := net.NewURI().
		WithRuleFunc(mock.Function()).
		WithRuleFunc(mock.Function()).
		Apply(context.TODO(), testVal, &output)

	if err == nil {
		t.Error("Expected errors to not be nil")
	} else if len(err) != 2 {
		t.Errorf("Expected 2 errors, got: %d", len(err))
	}
}

// Requirements:
// - Conflicting rules are deduplicated
func TestURICustomConflict(t *testing.T) {
	testVal := "https://example.com"

	mockA := testhelpers.NewMockRule[string]()
	mockA.ConflictKey = "test"

	mockB := testhelpers.NewMockRule[string]()

	var output string
	err := net.NewURI().
		WithRule(mockB).
		WithRule(mockA).
		WithRule(mockB).
		WithRule(mockA).
		WithRule(mockB).
		Apply(context.TODO(), testVal, &output)

	if err != nil {
		t.Errorf("Expected errors to be nil, got: %s", err)
	}

	if mockA.EvaluateCallCount() != 1 {
		t.Errorf("Expected 1 call to Evaluate, got: %d", mockA.EvaluateCallCount())
	}

	if mockB.EvaluateCallCount() != 3 {
		t.Errorf("Expected 3 call to Evaluate, got: %d", mockB.EvaluateCallCount())
	}
}
