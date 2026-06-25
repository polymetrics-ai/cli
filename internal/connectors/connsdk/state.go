package connsdk

import (
	"strconv"
	"strings"
	"time"
)

// CursorStateKey is the conventional key under which an incremental cursor is
// stored in a connector's state map.
const CursorStateKey = "cursor"

// Cursor returns the stored incremental cursor value (empty for a first sync).
func Cursor(state map[string]string) string {
	if state == nil {
		return ""
	}
	return strings.TrimSpace(state[CursorStateKey])
}

// WithCursor returns a copy of state with the cursor set.
func WithCursor(state map[string]string, value string) map[string]string {
	out := make(map[string]string, len(state)+1)
	for k, v := range state {
		out[k] = v
	}
	out[CursorStateKey] = value
	return out
}

// MaxCursor returns the larger of two cursor values. It compares numerically when
// both parse as numbers, as time when both parse as RFC3339 timestamps, and
// lexicographically otherwise. Empty values lose to non-empty.
func MaxCursor(a, b string) string {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	if an, aerr := strconv.ParseFloat(a, 64); aerr == nil {
		if bn, berr := strconv.ParseFloat(b, 64); berr == nil {
			if bn > an {
				return b
			}
			return a
		}
	}
	if at, aerr := parseTime(a); aerr == nil {
		if bt, berr := parseTime(b); berr == nil {
			if bt.After(at) {
				return b
			}
			return a
		}
	}
	if b > a {
		return b
	}
	return a
}

func parseTime(value string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errNotTime
}

var errNotTime = &timeParseError{}

type timeParseError struct{}

func (*timeParseError) Error() string { return "not a recognized time format" }
