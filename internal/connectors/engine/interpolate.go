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
// filter still overrides the default. A resolved value that is exactly ".."
// (or, after percent-decoding, still resolves to a ".." path segment) is
// rejected outright (F9b/m3): urlencodeSegment intentionally leaves a bare
// "." unescaped (it is not itself a metacharacter needing escaping for
// safe insertion), which means a literal ".." value would otherwise survive
// as an intact same-value path segment even though slashes are encoded.
func InterpolatePath(template string, vars Vars) (string, error) {
	out, err := interpolate(template, vars, true)
	if err != nil {
		return "", err
	}
	if containsDotDotSegment(out) {
		return "", fmt.Errorf("interpolate path: resolved template %q contains a \"..\" path segment", template)
	}
	return out, nil
}

// containsDotDotSegment reports whether any "/"-delimited segment of path is
// exactly "..", checking both the raw (possibly percent-encoded) segments
// and their percent-decoded form so an encoded traversal segment (e.g.
// "%2e%2e") is caught too.
func containsDotDotSegment(path string) bool {
	for _, seg := range strings.Split(path, "/") {
		if seg == ".." {
			return true
		}
		if decoded, err := url.PathUnescape(seg); err == nil && decoded == ".." {
			return true
		}
	}
	return false
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
// }}), which is a dotted reference optionally followed by one or more
// "| filter" stages, applied left-to-right (F9: a filter chain like
// "| urlencode | base64" must apply BOTH filters in sequence, never silently
// truncate after the first). The raw (pre-filter) resolved value is rejected
// outright if it carries CR/LF — header and path/query insertion are the
// injection surfaces (THREAT-MODEL §2) and no filter in this dialect is
// meant to legitimately produce or pass through newlines.
func resolveExpr(expr string, vars Vars, urlencodeDefault bool) (string, error) {
	parts := strings.Split(expr, "|")
	ref := strings.TrimSpace(parts[0])

	rawVal, err := resolveRefValue(ref, vars)
	if err != nil {
		return "", err
	}
	val := stringify(rawVal)
	if strings.ContainsAny(val, "\r\n") {
		return "", fmt.Errorf("interpolate: resolved value for %q contains CR/LF", ref)
	}

	filters := make([]string, 0, len(parts)-1)
	for _, p := range parts[1:] {
		if f := strings.TrimSpace(p); f != "" {
			filters = append(filters, f)
		}
	}
	if len(filters) == 0 && urlencodeDefault {
		filters = []string{"urlencode"}
	}

	cur := val
	for _, filter := range filters {
		next, err := applyFilterValue(filter, cur, rawVal)
		if err != nil {
			return "", err
		}
		cur = next
		// After the first filter stage the "raw" array-shaped value (if any)
		// no longer applies — only the FIRST filter in a chain may consume
		// the pre-stringify raw value (e.g. "record.tags | join:,"); every
		// subsequent filter operates on the running string result like
		// urlencode/base64/unix_seconds always have.
		rawVal = cur
	}
	return cur, nil
}

// unresolvedKeyError is the typed sentinel for "reference resolved to
// nothing because the key is absent" (F4, REVIEW.md: error-classification by
// substring matching is brittle — read.go's resolveHeaders/computed_fields
// callers use errors.As against this type instead of scanning err.Error()
// text). Namespace is "config", "secrets", or "record"; Key is the dotted
// key/path that could not be resolved.
type unresolvedKeyError struct {
	Namespace string
	Key       string
}

func (e *unresolvedKeyError) Error() string {
	return fmt.Sprintf("interpolate: unresolved key %q in %s", e.Key, e.Namespace)
}

// resolveRef resolves a dotted reference like "config.base_url",
// "secrets.token", "record.user.login", or the bare "cursor".
func resolveRef(ref string, vars Vars) (string, error) {
	v, err := resolveRefValue(ref, vars)
	if err != nil {
		return "", err
	}
	return stringify(v), nil
}

// resolveRefValue is resolveRef's raw-value counterpart: config/secrets/
// cursor resolve to a string (as always), but a "record.*" reference returns
// the RAW (possibly array/object/number) value rather than an
// already-stringified one, so filters that need the pre-stringify shape
// (join:<sep> on an array) can operate on it. Every other caller that only
// ever wants a string (resolveRef, and therefore EvalWhen's
// resolveRefForWhen) is unaffected — resolveRef just stringifies afterward,
// exactly matching its prior behavior byte-for-byte.
func resolveRefValue(ref string, vars Vars) (any, error) {
	if ref == "cursor" {
		return vars.Cursor, nil
	}

	segs := strings.Split(ref, ".")
	if len(segs) < 2 {
		return nil, fmt.Errorf("interpolate: unresolved reference %q", ref)
	}
	namespace, key := segs[0], segs[1]

	switch namespace {
	case "config":
		v, ok := vars.Config[key]
		if !ok {
			return nil, &unresolvedKeyError{Namespace: "config", Key: key}
		}
		return v, nil
	case "secrets":
		v, ok := vars.Secrets[key]
		if !ok {
			return nil, &unresolvedKeyError{Namespace: "secrets", Key: key}
		}
		return v, nil
	case "record":
		return resolveRecordPathValue(vars.Record, segs[1:])
	default:
		return nil, fmt.Errorf("interpolate: unknown namespace %q in reference %q", namespace, ref)
	}
}

// resolveRecordPathValue walks a dotted path into a raw record, returning the
// RAW value found (string/number/bool/array/object/nil) rather than an
// already-stringified one. A missing intermediate value is an
// *unresolvedKeyError with Namespace "record" (read.go's computed_fields
// relies on this — via errors.As, not string matching — to treat an absent
// optional nested field as "omit the computed field" rather than a hard
// error).
func resolveRecordPathValue(record map[string]any, path []string) (any, error) {
	if record == nil {
		return nil, &unresolvedKeyError{Namespace: "record", Key: strings.Join(path, ".")}
	}
	var cur any = record
	for i, seg := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, &unresolvedKeyError{Namespace: "record", Key: strings.Join(path[:i+1], ".")}
		}
		v, ok := m[seg]
		if !ok {
			return nil, &unresolvedKeyError{Namespace: "record", Key: strings.Join(path[:i+1], ".")}
		}
		cur = v
	}
	return cur, nil
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

// applyFilter applies the named filter to val ("" means no filter). Kept as
// the string-only entry point for callers/tests that never need the
// raw-value-aware join:<sep> filter; it delegates to applyFilterValue with
// no raw value available (so "join" on a string is always the
// not-an-array error, exactly as if no array had ever been in scope).
func applyFilter(filter, val string) (string, error) {
	return applyFilterValue(filter, val, val)
}

// applyFilterValue applies filter to val (the running string result of any
// prior filter stage), with rawVal available for filters that need the
// PRE-STRINGIFY shape of the value (currently only join:<sep>, which
// requires an actual []any — not its already-stringified form). Every other
// filter ignores rawVal and behaves exactly as before.
func applyFilterValue(filter, val string, rawVal any) (string, error) {
	switch {
	case filter == "":
		return val, nil
	case filter == "urlencode":
		return urlencodeSegment(val), nil
	case filter == "unix_seconds":
		t, err := time.Parse(time.RFC3339, val)
		if err != nil {
			return "", fmt.Errorf("interpolate: unix_seconds filter: invalid RFC3339 value %q: %w", val, err)
		}
		return strconv.FormatInt(t.Unix(), 10), nil
	case filter == "base64":
		return base64.StdEncoding.EncodeToString([]byte(val)), nil
	case filter == "last_path_segment":
		return lastPathSegment(val), nil
	case strings.HasPrefix(filter, "join:"):
		sep := strings.TrimPrefix(filter, "join:")
		return applyJoinFilter(sep, rawVal)
	default:
		return "", fmt.Errorf("interpolate: unknown filter %q", filter)
	}
}

// applyJoinFilter joins an array-valued rawVal with sep (F7 meta-rule
// enablement: e.g. searxng's "engines" declared as an array, joined into the
// comma-separated wire form legacy sends, without changing the emitted
// RECORD shape — only the outgoing request representation). A non-array
// rawVal is a hard error rather than a silent fmt.Sprint stringification,
// since joining a scalar is never the intended use of this filter. Note:
// sep cannot itself be "|" — the filter-chain delimiter takes precedence
// during the outer "|"-split, so "join:|" would be parsed as "join:" piped
// into a next (empty) filter stage rather than a literal pipe separator; any
// other separator (including multi-character ones, since sep is everything
// after the first ":") is unambiguous.
func applyJoinFilter(sep string, rawVal any) (string, error) {
	arr, ok := rawVal.([]any)
	if !ok {
		return "", fmt.Errorf("interpolate: join filter requires an array value, got %T", rawVal)
	}
	parts := make([]string, len(arr))
	for i, v := range arr {
		parts[i] = stringify(v)
	}
	return strings.Join(parts, sep), nil
}

// urlencodeSegment percent-encodes val for safe insertion as a single path
// segment: like url.QueryEscape (every reserved/metachar byte is percent-
// encoded, including '%' itself for the double-encode guard) but with a
// literal space encoded as "%20" rather than "+".
func urlencodeSegment(val string) string {
	return strings.ReplaceAll(url.QueryEscape(val), "+", "%20")
}

// lastPathSegment returns the final "/"-delimited, non-empty segment of val
// (gap-loop item 4, REVIEW-B.md finding 1 / cross-cutting adjudication 1):
// the trailing-URI-segment convention a legacy HAL/URI-keyed API commonly
// derives its record id from (calendly's idFromURI(uri)). A trailing slash
// is ignored (does not produce an empty final segment); a value with no "/"
// at all passes through unchanged (nothing to split); an entirely empty
// value returns "". This mirrors the semantics of trimming a trailing slash
// then taking strings.LastIndex(val, "/")+1: onward, never an error — a
// computed_fields template using this filter on a genuinely malformed source
// value degrades to returning that value's own trailing text rather than
// hard-failing the whole record.
func lastPathSegment(val string) string {
	trimmed := strings.TrimRight(val, "/")
	if trimmed == "" {
		return ""
	}
	if idx := strings.LastIndex(trimmed, "/"); idx >= 0 {
		return trimmed[idx+1:]
	}
	return trimmed
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

// knownFilterNames is the set of filter names ResolveCheck accepts
// statically (F9: a typo'd filter name should fail `connectorgen validate`,
// not silently error only at runtime). "join:<sep>" is a prefix form, not a
// fixed name, and is checked separately in isKnownFilter.
var knownFilterNames = map[string]bool{
	"urlencode":         true,
	"unix_seconds":      true,
	"base64":            true,
	"last_path_segment": true,
}

func isKnownFilter(filter string) bool {
	if knownFilterNames[filter] {
		return true
	}
	return strings.HasPrefix(filter, "join:")
}

// ResolveCheck statically validates every {{ }} reference in template against
// specKeys (the declared spec.json property names), used by connectorgen at
// build time. record/cursor/secrets references are not checked against
// specKeys since they are not spec-declared. Every filter stage in a
// (possibly chained) "| filter1 | filter2" pipeline must be a known filter
// name (F9: an unknown filter name is a validate-time error, not just a
// runtime one).
func ResolveCheck(template string, specKeys map[string]bool) error {
	matches := templatePattern.FindAllStringSubmatch(template, -1)
	for _, m := range matches {
		segs := strings.Split(m[1], "|")
		ref := strings.TrimSpace(segs[0])
		for _, f := range segs[1:] {
			filter := strings.TrimSpace(f)
			if filter == "" {
				continue
			}
			if !isKnownFilter(filter) {
				return fmt.Errorf("resolve check: unknown filter %q referenced in %q", filter, strings.TrimSpace(m[1]))
			}
		}
		if ref == "cursor" {
			continue
		}
		refSegs := strings.Split(ref, ".")
		if len(refSegs) < 2 {
			return fmt.Errorf("resolve check: malformed reference %q", ref)
		}
		namespace, key := refSegs[0], refSegs[1]
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

// ResolveCheckAuthSpec statically validates EVERY templated field of an
// AuthSpec against specKeys (F9, REVIEW.md: cmd/connectorgen/validate.go's
// checkInterpolations only checked Token/Value/When, leaving username/
// password/token_url/client_id/client_secret/scopes typos to pass static
// validation and fail only at runtime). Wiring this into `connectorgen
// validate` itself is a follow-up (cmd/connectorgen is outside this task's
// editable file set); this is the engine-side building block for it.
func ResolveCheckAuthSpec(spec AuthSpec, specKeys map[string]bool) error {
	fields := []struct {
		name, tmpl string
	}{
		{"token", spec.Token},
		{"username", spec.Username},
		{"password", spec.Password},
		{"value", spec.Value},
		{"token_url", spec.TokenURL},
		{"client_id", spec.ClientID},
		{"client_secret", spec.ClientSecret},
		{"scopes", spec.Scopes},
		{"when", spec.When},
	}
	for _, f := range fields {
		if f.tmpl == "" {
			continue
		}
		if err := ResolveCheck(f.tmpl, specKeys); err != nil {
			return fmt.Errorf("auth spec (mode %q) field %q: %w", spec.Mode, f.name, err)
		}
	}
	return nil
}
