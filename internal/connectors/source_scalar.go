package connectors

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ParseExactJSONNumber accepts one JSON numeric lexeme and preserves its exact
// rational value. Callers apply their input and work budgets before comparison.
func ParseExactJSONNumber(value string) (*big.Rat, bool) {
	if !json.Valid([]byte(value)) {
		return nil, false
	}
	return new(big.Rat).SetString(value)
}

// ParseExactJSONInteger accepts the integer syntax used by executable flags.
// Decimal and exponent syntax remain distinct even when numerically integral.
func ParseExactJSONInteger(value string) (*big.Int, bool) {
	if !json.Valid([]byte(value)) {
		return nil, false
	}
	return new(big.Int).SetString(value, 10)
}

// DecodeSourceScalar preserves source scalar spelling and presence. maxBytes is
// the caller's finite PM input/work budget, not a provider schema constraint.
// Numeric exponent magnitude is bounded by the same budget before any caller
// converts the lexeme to a rational for schema comparisons.
func DecodeSourceScalar(kind, raw string, maxBytes int) (any, error) {
	if maxBytes <= 0 || len(raw) > maxBytes || !utf8.ValidString(raw) {
		return nil, fmt.Errorf("source scalar exceeds input budget or has invalid UTF-8")
	}
	switch kind {
	case "string":
		return raw, nil
	case "boolean":
		switch raw {
		case "true":
			return true, nil
		case "false":
			return false, nil
		}
	case "integer", "number":
		if raw == "" || strings.TrimSpace(raw) != raw {
			break
		}
		number, err := decodeExactNumber([]byte(raw))
		if err != nil {
			break
		}
		if kind == "integer" && strings.ContainsAny(raw, ".eE") {
			break
		}
		if index := strings.IndexAny(raw, "eE"); index >= 0 {
			exponent, err := strconv.ParseInt(raw[index+1:], 10, 64)
			if err != nil || exponent > int64(maxBytes) || exponent < -int64(maxBytes) {
				return nil, fmt.Errorf("source scalar exceeds numeric work budget")
			}
		}
		return number, nil
	default:
		return nil, fmt.Errorf("unsupported source scalar type %q", kind)
	}
	return nil, fmt.Errorf("invalid source %s scalar", kind)
}
