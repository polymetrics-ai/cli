package engine

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Vars is the variable environment available to templates: config values,
// secret values, the current raw record (nil outside record contexts), and
// the current cursor value.
type Vars struct {
	Config  map[string]string
	Secrets map[string]string
	Record  map[string]any
	Cursor  string
}

// templatePattern matches a single {{ ... }} expression, capturing its inner
// expression (a dotted reference optionally piped through one filter).
var templatePattern = regexp.MustCompile(`\{\{\s*(.*?)\s*\}\}`)

// Interpolate resolves every {{ }} expression in template against vars. No
// filter is applied to raw text; urlencode is opt-in via the explicit filter
// syntax ({{ config.x | urlencode }}).
func Interpolate(template string, vars Vars) (string, error) {
	return interpolate(template, vars, false)
}

// InterpolatePath resolves every {{ }} expression in template against vars,
// applying the urlencode filter by DEFAULT to every resolved value (path
// segments are the primary injection surface; THREAT-MODEL §2). An explicit
// filter still overrides the default.
func InterpolatePath(template string, vars Vars) (string, error) {
	return interpolate(template, vars, true)
}

// InterpolateHeader resolves every {{ }} expression in template against vars
// and rejects any resolved value containing CR or LF (header injection guard,
// THREAT-MODEL §2).
func InterpolateHeader(template string, vars Vars) (string, error) {
	out, err := interpolate(template, vars, false)
	if err != nil {
		return "", err
	}
	if strings.ContainsAny(out, "\r\n") {
		return "", fmt.Errorf("interpolate header: resolved value contains CR/LF")
	}
	return out, nil
}

func interpolate(template string, vars Vars, urlencodeDefault bool) (string, error) {
	var firstErr error
	out := templatePattern.ReplaceAllStringFunc(template, func(match string) string {
		if firstErr != nil {
			return ""
		}
		inner := templatePattern.FindStringSubmatch(match)[1]
		val, err := resolveExpr(inner, vars, urlencodeDefault)
		if err != nil {
			firstErr = err
			return ""
		}
		return val
	})
	if firstErr != nil {
		return "", firstErr
	}
	return out, nil
}

// resolveExpr resolves a single expression body (the text between {{ and
// }}), which is a dotted reference optionally followed by "| filter". The
// raw (pre-filter) resolved value is rejected outright if it carries CR/LF —
// header and path/query insertion are the injection surfaces (THREAT-MODEL
// §2) and no filter in this dialect is meant to legitimately produce or pass
// through newlines.
func resolveExpr(expr string, vars Vars, urlencodeDefault bool) (string, error) {
	parts := strings.Split(expr, "|")
	ref := strings.TrimSpace(parts[0])

	val, err := resolveRef(ref, vars)
	if err != nil {
		return "", err
	}
	if strings.ContainsAny(val, "\r\n") {
		return "", fmt.Errorf("interpolate: resolved value for %q contains CR/LF", ref)
	}

	filter := ""
	if len(parts) > 1 {
		filter = strings.TrimSpace(parts[1])
	} else if urlencodeDefault {
		filter = "urlencode"
	}

	return applyFilter(filter, val)
}

// resolveRef resolves a dotted reference like "config.base_url",
// "secrets.token", "record.user.login", or the bare "cursor".
func resolveRef(ref string, vars Vars) (string, error) {
	if ref == "cursor" {
		return vars.Cursor, nil
	}

	segs := strings.Split(ref, ".")
	if len(segs) < 2 {
		return "", fmt.Errorf("interpolate: unresolved reference %q", ref)
	}
	namespace, key := segs[0], segs[1]

	switch namespace {
	case "config":
		v, ok := vars.Config[key]
		if !ok {
			return "", fmt.Errorf("interpolate: unresolved key %q in config", key)
		}
		return v, nil
	case "secrets":
		v, ok := vars.Secrets[key]
		if !ok {
			return "", fmt.Errorf("interpolate: unresolved key %q in secrets", key)
		}
		return v, nil
	case "record":
		return resolveRecordPath(vars.Record, segs[1:])
	default:
		return "", fmt.Errorf("interpolate: unknown namespace %q in reference %q", namespace, ref)
	}
}

// resolveRecordPath walks a dotted path into a raw record. A missing
// intermediate value resolves to "" rather than panicking (read.go's
// computed_fields rely on this for optional nested fields).
func resolveRecordPath(record map[string]any, path []string) (string, error) {
	if record == nil {
		return "", fmt.Errorf("interpolate: unresolved key %q in record", strings.Join(path, "."))
	}
	var cur any = record
	for i, seg := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return "", fmt.Errorf("interpolate: unresolved key %q in record", strings.Join(path[:i+1], "."))
		}
		v, ok := m[seg]
		if !ok {
			return "", fmt.Errorf("interpolate: unresolved key %q in record", strings.Join(path[:i+1], "."))
		}
		cur = v
	}
	return stringify(cur), nil
}

func stringify(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprint(t)
	}
}

// applyFilter applies the named filter to val. "" means no filter.
func applyFilter(filter, val string) (string, error) {
	switch filter {
	case "":
		return val, nil
	case "urlencode":
		return urlencodeSegment(val), nil
	case "unix_seconds":
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return "", fmt.Errorf("interpolate: unix_seconds filter: invalid RFC3339 value %q: %w", val, err)
		}
		return strconv.FormatInt(t.Unix(), 10), nil
	case "base64":
		return base64.StdEncoding.EncodeToString([]byte(val)), nil
	default:
		return "", fmt.Errorf("interpolate: unknown filter %q", filter)
	}
}

// urlencodeSegment percent-encodes val for safe insertion as a single path
// segment: like url.QueryEscape (every reserved/metachar byte is percent-
// encoded, including '%' itself for the double-encode guard) but with a
// literal space encoded as "%20" rather than "+".
func urlencodeSegment(val string) string {
	return strings.ReplaceAll(url.QueryEscape(val), "+", "%20")
}

// EvalWhen evaluates a `when` condition template against vars. Supported
// grammar (design §B.3): equality (`config.k == 'v'`), membership
// (`config.k in ['a','b']`), and truthiness (`config.k` alone). Anything else
// is a compile error — no eval, no arithmetic, no user functions.
//
// Unlike general template interpolation (Interpolate/InterpolatePath/
// InterpolateHeader), a config/secrets reference whose key is entirely
// ABSENT at runtime does not error here: it resolves as an empty string, so
// truthiness is false, `==` compares against "", and `in [...]` treats it as
// not contained (unless the list itself contains the empty-string literal).
// This is what makes the OPTIONAL-credential pattern possible — e.g.
// `when: "{{ secrets.api_key }}"` gating an auth spec off when the caller
// never populated that secret — without requiring a separate "is this key
// present" primitive. Static validation (ResolveCheck, run by connectorgen
// validate) is unaffected: a when-template referencing a key that isn't even
// DECLARED in spec.json's properties is still a hard validate-time error;
// only RUNTIME absence of a spec-known key is tolerated here.
func EvalWhen(cond string, vars Vars) (bool, error) {
	inner, err := extractWhenExpr(cond)
	if err != nil {
		return false, err
	}

	if idx := strings.Index(inner, "=="); idx >= 0 {
		left := strings.TrimSpace(inner[:idx])
		right := strings.TrimSpace(inner[idx+2:])
		lv, err := resolveRefForWhen(left, vars)
		if err != nil {
			return false, err
		}
		rv, err := parseLiteral(right)
		if err != nil {
			return false, err
		}
		return lv == rv, nil
	}

	if idx := strings.Index(inner, " in "); idx >= 0 {
		left := strings.TrimSpace(inner[:idx])
		right := strings.TrimSpace(inner[idx+len(" in "):])
		lv, err := resolveRefForWhen(left, vars)
		if err != nil {
			return false, err
		}
		list, err := parseList(right)
		if err != nil {
			return false, err
		}
		for _, item := range list {
			if item == lv {
				return true, nil
			}
		}
		return false, nil
	}

	// Reject any other recognizable-but-unsupported operator explicitly so it
	// is a compile error, not silently treated as truthiness.
	for _, op := range []string{"!=", ">=", "<=", ">", "<", "&&", "||"} {
		if strings.Contains(inner, op) {
			return false, fmt.Errorf("when: unsupported operator %q in condition %q", op, cond)
		}
	}

	// Truthiness: a bare dotted reference.
	v, err := resolveRefForWhen(strings.TrimSpace(inner), vars)
	if err != nil {
		return false, err
	}
	return v != "", nil
}

// resolveRefForWhen resolves ref exactly like resolveRef, EXCEPT that a
// config.* or secrets.* reference whose key is absent from vars resolves to
// ("", nil) instead of propagating resolveRef's "unresolved key" error. This
// tolerance is intentionally scoped to when-condition evaluation only:
// resolveRef itself is untouched, so every other caller (Interpolate and its
// path/header variants, buildAuthenticator's per-mode field resolution)
// keeps hard-erroring on an absent key exactly as before. cursor and
// record.* references are delegated straight through (unaffected — EvalWhen
// is never invoked today with a non-nil vars.Record, since authVars, its
// only production vars-builder, never sets one).
func resolveRefForWhen(ref string, vars Vars) (string, error) {
	segs := strings.Split(ref, ".")
	if len(segs) >= 2 {
		switch segs[0] {
		case "config":
			if _, ok := vars.Config[segs[1]]; !ok {
				return "", nil
			}
		case "secrets":
			if _, ok := vars.Secrets[segs[1]]; !ok {
				return "", nil
			}
		}
	}
	return resolveRef(ref, vars)
}

// extractWhenExpr strips the {{ }} wrapper a `when` string is conventionally
// written with, tolerating a bare (unwrapped) expression too.
func extractWhenExpr(cond string) (string, error) {
	trimmed := strings.TrimSpace(cond)
	if strings.HasPrefix(trimmed, "{{") && strings.HasSuffix(trimmed, "}}") {
		return strings.TrimSpace(trimmed[2 : len(trimmed)-2]), nil
	}
	if trimmed == "" {
		return "", fmt.Errorf("when: empty condition")
	}
	return trimmed, nil
}

// parseLiteral parses a single-quoted string literal used on the right-hand
// side of `==`.
func parseLiteral(s string) (string, error) {
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		return s[1 : len(s)-1], nil
	}
	return "", fmt.Errorf("when: expected a quoted string literal, got %q", s)
}

// parseList parses a `['a', 'b']` literal list used with the `in` operator.
func parseList(s string) ([]string, error) {
	if len(s) < 2 || s[0] != '[' || s[len(s)-1] != ']' {
		return nil, fmt.Errorf("when: expected a list literal, got %q", s)
	}
	inner := strings.TrimSpace(s[1 : len(s)-1])
	if inner == "" {
		return nil, nil
	}
	items := strings.Split(inner, ",")
	out := make([]string, 0, len(items))
	for _, item := range items {
		lit, err := parseLiteral(strings.TrimSpace(item))
		if err != nil {
			return nil, err
		}
		out = append(out, lit)
	}
	return out, nil
}

// ResolveCheck statically validates every {{ }} reference in template against
// specKeys (the declared spec.json property names), used by connectorgen at
// build time. record/cursor/secrets references are not checked against
// specKeys since they are not spec-declared.
func ResolveCheck(template string, specKeys map[string]bool) error {
	matches := templatePattern.FindAllStringSubmatch(template, -1)
	for _, m := range matches {
		expr := strings.Split(m[1], "|")[0]
		ref := strings.TrimSpace(expr)
		if ref == "cursor" {
			continue
		}
		segs := strings.Split(ref, ".")
		if len(segs) < 2 {
			return fmt.Errorf("resolve check: malformed reference %q", ref)
		}
		namespace, key := segs[0], segs[1]
		switch namespace {
		case "config":
			if !specKeys[key] {
				return fmt.Errorf("resolve check: unknown spec key %q referenced as %q", key, ref)
			}
		case "secrets", "record":
			// not statically checkable against specKeys here.
		default:
			return fmt.Errorf("resolve check: unknown namespace %q in reference %q", namespace, ref)
		}
	}
	return nil
}
