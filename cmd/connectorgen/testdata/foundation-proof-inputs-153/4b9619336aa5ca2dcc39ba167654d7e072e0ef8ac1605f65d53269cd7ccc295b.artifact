package connsdk

import (
	"fmt"
	"strconv"
	"strings"
)

// ExactHTTPStatusRangeExecutionGap names the runtime capability required to
// execute a source-declared HTTP status class without widening it.
const ExactHTTPStatusRangeExecutionGap = "exact response-status range execution"

// ExactHTTPStatusNon2xxExecutionGap names the current runtime limitation for
// source-declared non-2xx statuses. The source contract remains retained and
// must be accounted for as blocked rather than widened into success-class
// execution.
const ExactHTTPStatusNon2xxExecutionGap = "exact response-status non-2xx execution"

// ExactHTTPStatusRangeError reports a source-declared HTTP status class such
// as 2XX. The source contract remains valid but an exact-status executor cannot
// represent it.
type ExactHTTPStatusRangeError struct {
	Declared string
}

func (e *ExactHTTPStatusRangeError) Error() string {
	return fmt.Sprintf("%q requires %s", e.Declared, ExactHTTPStatusRangeExecutionGap)
}

// NormalizeExactHTTPStatus returns the exact runtime status for an unambiguous
// textual HTTP code. It accepts whitespace, a leading plus, and leading zeroes
// while rejecting ranges and non-numeric values. Callers retain declared when
// source evidence needs its original spelling.
func NormalizeExactHTTPStatus(declared string) (int, error) {
	value := strings.TrimSpace(declared)
	if isExactHTTPStatusRange(value) {
		return 0, &ExactHTTPStatusRangeError{Declared: declared}
	}
	code, ok := exactHTTPStatusCode(value)
	if !ok {
		return 0, fmt.Errorf("must resolve to an HTTP status between 100 and 599")
	}
	return code, nil
}

func isExactHTTPStatusRange(value string) bool {
	if len(value) == 3 && value[0] >= '1' && value[0] <= '5' && (value[1] == 'x' || value[1] == 'X') && (value[2] == 'x' || value[2] == 'X') {
		return true
	}
	start, end, hasRange := strings.Cut(value, "-")
	if !hasRange || strings.Contains(end, "-") {
		return false
	}
	startCode, startOK := exactHTTPStatusCode(start)
	endCode, endOK := exactHTTPStatusCode(end)
	return startOK && endOK && startCode <= endCode
}

func exactHTTPStatusCode(value string) (int, bool) {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "+") {
		value = value[1:]
	}
	if value == "" {
		return 0, false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return 0, false
		}
	}
	code, err := strconv.ParseUint(value, 10, 16)
	if err != nil || code < 100 || code > 599 {
		return 0, false
	}
	return int(code), true
}
