package engine

import (
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/connsdk"
)

func TestErrorWrapsHTTPErrorReachableViaErrorsAs(t *testing.T) {
	httpErr := &connsdk.HTTPError{Status: 401, URL: "https://api.example.com/widgets", Body: "unauthorized"}
	e := &Error{
		Connector:   "acme",
		Stream:      "widgets",
		Page:        2,
		RecordIndex: -1,
		Err:         httpErr,
	}

	var target *connsdk.HTTPError
	if !errors.As(e, &target) {
		t.Fatalf("errors.As did not reach the wrapped *connsdk.HTTPError")
	}
	if target.Status != 401 {
		t.Fatalf("unwrapped HTTPError.Status = %d, want 401", target.Status)
	}
}

func TestErrorUnwrap(t *testing.T) {
	inner := errors.New("boom")
	e := &Error{Connector: "acme", Err: inner}
	if !errors.Is(e, inner) {
		t.Fatalf("errors.Is did not find the inner sentinel via Unwrap")
	}
}

func TestErrorContextFields(t *testing.T) {
	e := &Error{
		Connector:   "acme",
		Stream:      "widgets",
		Page:        3,
		RecordIndex: -1,
		Err:         errors.New("network blip"),
	}
	msg := e.Error()
	if !strings.Contains(msg, "acme") {
		t.Fatalf("Error() %q does not mention connector", msg)
	}
	if !strings.Contains(msg, "widgets") {
		t.Fatalf("Error() %q does not mention stream", msg)
	}

	e2 := &Error{
		Connector:   "acme",
		Action:      "create_widget",
		Page:        -1,
		RecordIndex: 4,
		Err:         errors.New("validation failed"),
	}
	msg2 := e2.Error()
	if !strings.Contains(msg2, "create_widget") {
		t.Fatalf("Error() %q does not mention action", msg2)
	}
	if !strings.Contains(msg2, "4") {
		t.Fatalf("Error() %q does not mention record index", msg2)
	}
}

func TestErrorMapStatusMatch(t *testing.T) {
	rules := []ErrorRule{
		{Status: 401, Hint: "token is missing or expired; re-run pm credentials add acme"},
		{Status: 403, MatchBody: "rate limit", Class: "rate_limited"},
		{Status: 404, Hint: "not found"},
	}

	class, hint := applyErrorMap(rules, &connsdk.HTTPError{Status: 401, URL: "https://x", Body: "nope"})
	if hint != "token is missing or expired; re-run pm credentials add acme" {
		t.Fatalf("hint = %q", hint)
	}
	if class != "" {
		t.Fatalf("class = %q, want empty", class)
	}
}

func TestErrorMapStatusAndMatchBody(t *testing.T) {
	rules := []ErrorRule{
		{Status: 403, MatchBody: "rate limit", Class: "rate_limited"},
		{Status: 403, Hint: "forbidden"},
	}

	class, hint := applyErrorMap(rules, &connsdk.HTTPError{Status: 403, URL: "https://x", Body: "you hit the rate limit, slow down"})
	if class != "rate_limited" {
		t.Fatalf("class = %q, want rate_limited", class)
	}
	if hint != "" {
		t.Fatalf("hint = %q, want empty (first matching rule wins)", hint)
	}
}

func TestErrorMapMatchBodyMismatchFallsThrough(t *testing.T) {
	rules := []ErrorRule{
		{Status: 403, MatchBody: "rate limit", Class: "rate_limited"},
		{Status: 403, Hint: "forbidden"},
	}

	class, hint := applyErrorMap(rules, &connsdk.HTTPError{Status: 403, URL: "https://x", Body: "insufficient scope"})
	if class != "" {
		t.Fatalf("class = %q, want empty (match_body did not match)", class)
	}
	if hint != "forbidden" {
		t.Fatalf("hint = %q, want forbidden (second rule status-only match)", hint)
	}
}

func TestErrorMapNoMatch(t *testing.T) {
	rules := []ErrorRule{
		{Status: 401, Hint: "token expired"},
	}
	class, hint := applyErrorMap(rules, &connsdk.HTTPError{Status: 500, URL: "https://x", Body: "boom"})
	if class != "" || hint != "" {
		t.Fatalf("class=%q hint=%q, want both empty for non-HTTPError-matching status", class, hint)
	}
}

func TestErrorMapNonHTTPError(t *testing.T) {
	rules := []ErrorRule{{Status: 401, Hint: "token expired"}}
	class, hint := applyErrorMap(rules, errors.New("not an http error"))
	if class != "" || hint != "" {
		t.Fatalf("class=%q hint=%q, want both empty for non-HTTPError err", class, hint)
	}
}

func TestErrorHintSurfacesVerbatim(t *testing.T) {
	rules := []ErrorRule{
		{Status: 401, Hint: "token is missing or expired; re-run pm credentials add acme"},
	}
	httpErr := &connsdk.HTTPError{Status: 401, URL: "https://api.example.com/widgets", Body: "unauthorized"}
	class, hint := applyErrorMap(rules, httpErr)

	e := &Error{
		Connector:   "acme",
		Stream:      "widgets",
		Page:        1,
		RecordIndex: -1,
		Class:       class,
		Hint:        hint,
		Err:         httpErr,
	}

	msg := e.Error()
	if !strings.Contains(msg, "token is missing or expired; re-run pm credentials add acme") {
		t.Fatalf("Error() %q does not surface hint verbatim", msg)
	}
}

func TestErrorRedactsSecrets(t *testing.T) {
	token := "sk_live_abcdefghijklmnop"
	httpErr := &connsdk.HTTPError{
		Status: 400,
		URL:    "https://api.example.com/widgets?api_key=" + token,
		Body:   "token=" + token + " is invalid",
	}
	e := &Error{
		Connector:   "acme",
		Stream:      "widgets",
		Page:        1,
		RecordIndex: -1,
		Err:         httpErr,
	}

	msg := e.Error()
	if strings.Contains(msg, token) {
		t.Fatalf("Error() %q leaks the secret token", msg)
	}
}
