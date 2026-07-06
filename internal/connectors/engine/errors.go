package engine

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/safety"
)

// Error is the engine's typed error: it wraps the underlying failure (most
// commonly a *connsdk.HTTPError) with the context needed to act on it —
// which connector/stream/action, which page or record index, and an
// optional error_map-derived class/hint.
type Error struct {
	Connector, Stream, Action string
	Page, RecordIndex         int // -1 when not applicable
	Class, Hint               string
	Err                       error
}

// Error renders a redacted, context-carrying message. error_map hints
// surface verbatim (design §F.4); everything else is passed through
// safety.RedactErrorText so secrets never reach logs/CLI output.
func (e *Error) Error() string {
	var b strings.Builder
	b.WriteString(e.Connector)
	if e.Stream != "" {
		b.WriteString(" stream=")
		b.WriteString(e.Stream)
	}
	if e.Action != "" {
		b.WriteString(" action=")
		b.WriteString(e.Action)
	}
	if e.Page >= 0 {
		b.WriteString(" page=")
		b.WriteString(strconv.Itoa(e.Page))
	}
	if e.RecordIndex >= 0 {
		b.WriteString(" record=")
		b.WriteString(strconv.Itoa(e.RecordIndex))
	}
	if e.Class != "" {
		b.WriteString(" class=")
		b.WriteString(e.Class)
	}

	msg := safety.RedactErrorText(fmt.Sprintf("%s: %v", b.String(), e.Err))

	if e.Hint != "" {
		// Hints are operator-authored guidance (error_map.hint in streams.json)
		// and must surface verbatim, never redacted or truncated.
		msg = msg + " (" + e.Hint + ")"
	}
	return msg
}

// Unwrap exposes the wrapped error so errors.Is/errors.As can reach it (and
// in turn reach a wrapped *connsdk.HTTPError).
func (e *Error) Unwrap() error {
	return e.Err
}

// applyErrorMap evaluates rules in declared order against err (typically a
// *connsdk.HTTPError) and returns the class/hint from the first rule whose
// status matches AND, if match_body is set, whose body substring matches
// too. A rule with only "status" matches any body. No match returns ("","").
// A non-HTTPError err never matches (error_map is an HTTP-status concept).
func applyErrorMap(rules []ErrorRule, err error) (class, hint string) {
	var httpErr *connsdk.HTTPError
	if !errors.As(err, &httpErr) {
		return "", ""
	}

	for _, rule := range rules {
		if rule.Status != httpErr.Status {
			continue
		}
		if rule.MatchBody != "" && !strings.Contains(httpErr.Body, rule.MatchBody) {
			continue
		}
		return rule.Class, rule.Hint
	}
	return "", ""
}
