package logging

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"time"
	"unicode"

	"polymetrics.ai/internal/safety"
)

const (
	maxDynamicAttrs = 64
	maxDynamicSlice = 64
	maxDynamicName  = 128
	redactedKeyName = "redacted_key"
	redactedGroup   = "redacted_group"
)

// RedactionOptions configures RedactingHandler.
type RedactionOptions struct {
	SensitiveKeys []string
	Registry      *ValueRegistry
}

// RedactingHandler redacts sensitive keys and registered values before records
// reach the wrapped handler. Place it as the outermost slog handler.
type RedactingHandler struct {
	next          slog.Handler
	registry      *ValueRegistry
	sensitiveKeys map[string]struct{}
	attrs         []groupedAttrs
	groups        []string
}

type groupedAttrs struct {
	groups []string
	attrs  []slog.Attr
}

// NewRedactingHandler wraps next with key, value, URL, and terminal sanitization.
func NewRedactingHandler(next slog.Handler, opts RedactionOptions) slog.Handler {
	if next == nil {
		next = discardHandler{}
	}
	registry := opts.Registry
	if registry == nil {
		registry = defaultRegistry
	}
	keys := fixedSensitiveKeys()
	for _, key := range opts.SensitiveKeys {
		keys[normalizeKey(key)] = struct{}{}
	}
	return &RedactingHandler{next: next, registry: registry, sensitiveKeys: keys}
}

func (h *RedactingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *RedactingHandler) Handle(ctx context.Context, record slog.Record) error {
	redacted := slog.NewRecord(record.Time, record.Level, h.redactText(ctx, record.Message), record.PC)
	entries := make([]groupedAttrs, 0, len(h.attrs)+1)
	entries = append(entries, h.attrs...)
	attrs := make([]slog.Attr, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return len(attrs) < maxDynamicAttrs
	})
	entries = append(entries, groupedAttrs{groups: cloneStrings(h.groups), attrs: attrs})
	for _, attr := range h.redactedEntries(ctx, entries) {
		redacted.AddAttrs(attr)
	}
	return h.next.Handle(ctx, redacted)
}

func (h *RedactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.groups = cloneStrings(h.groups)
	clone.attrs = append(cloneGroupedAttrs(h.attrs), groupedAttrs{groups: cloneStrings(h.groups), attrs: cloneAttrs(attrs)})
	return &clone
}

func (h *RedactingHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	clone := *h
	clone.attrs = cloneGroupedAttrs(h.attrs)
	clone.groups = append(cloneStrings(h.groups), name)
	return &clone
}

func (h *RedactingHandler) redactedEntries(ctx context.Context, entries []groupedAttrs) []slog.Attr {
	out := make([]slog.Attr, 0)
	for _, entry := range entries {
		wrapped := wrapGroups(entry.groups, entry.attrs)
		for _, attr := range wrapped {
			out = append(out, h.redactAttr(ctx, attr, false))
		}
	}
	return mergeGroupAttrs(out)
}

func wrapGroups(groups []string, attrs []slog.Attr) []slog.Attr {
	out := cloneAttrs(attrs)
	if len(out) > maxDynamicAttrs {
		out = append(out[:maxDynamicAttrs], slog.String(redactedKeyName, redactedValue))
	}
	for i := len(groups) - 1; i >= 0; i-- {
		group, _ := sanitizeGroupName(groups[i])
		if group == "" {
			continue
		}
		out = []slog.Attr{{Key: group, Value: slog.GroupValue(out...)}}
	}
	return out
}

func mergeGroupAttrs(attrs []slog.Attr) []slog.Attr {
	out := make([]slog.Attr, 0, len(attrs))
	groupIndex := map[string]int{}
	for _, attr := range attrs {
		attr.Value = attr.Value.Resolve()
		if attr.Key != "" && attr.Value.Kind() == slog.KindGroup {
			if idx, ok := groupIndex[attr.Key]; ok && out[idx].Value.Kind() == slog.KindGroup {
				merged := append(cloneAttrs(out[idx].Value.Group()), attr.Value.Group()...)
				out[idx].Value = slog.GroupValue(mergeGroupAttrs(merged)...)
				continue
			}
			groupIndex[attr.Key] = len(out)
		}
		out = append(out, attr)
	}
	return out
}

func (h *RedactingHandler) redactAttr(ctx context.Context, attr slog.Attr, parentSensitive bool) slog.Attr {
	var unsafeKey bool
	attr.Key, unsafeKey = sanitizeKeyName(attr.Key)
	attr.Value = attr.Value.Resolve()
	keySensitive := parentSensitive || unsafeKey || (attr.Key != "" && h.isSensitiveKey(attr.Key))
	if attr.Value.Kind() == slog.KindGroup {
		attrs := attr.Value.Group()
		redacted := make([]slog.Attr, 0, len(attrs))
		for i, child := range attrs {
			if i >= maxDynamicAttrs {
				redacted = append(redacted, slog.String(redactedKeyName, redactedValue))
				break
			}
			redacted = append(redacted, h.redactAttr(ctx, child, keySensitive))
		}
		attr.Value = slog.GroupValue(mergeGroupAttrs(redacted)...)
		return attr
	}
	if keySensitive {
		attr.Value = slog.StringValue(redactedValue)
		return attr
	}
	attr.Value = h.redactValue(ctx, attr.Value)
	return attr
}

func (h *RedactingHandler) redactValue(ctx context.Context, value slog.Value) slog.Value {
	value = value.Resolve()
	if value.Kind() == slog.KindGroup {
		attrs := value.Group()
		redacted := make([]slog.Attr, 0, len(attrs))
		for i, attr := range attrs {
			if i >= maxDynamicAttrs {
				redacted = append(redacted, slog.String(redactedKeyName, redactedValue))
				break
			}
			redacted = append(redacted, h.redactAttr(ctx, attr, false))
		}
		return slog.GroupValue(mergeGroupAttrs(redacted)...)
	}
	if value.Kind() == slog.KindString {
		return slog.StringValue(h.redactText(ctx, value.String()))
	}
	if value.Kind() != slog.KindAny {
		return value
	}
	return h.redactAny(ctx, value.Any())
}

func (h *RedactingHandler) redactAny(ctx context.Context, value any) slog.Value {
	switch v := value.(type) {
	case nil:
		return slog.AnyValue(nil)
	case error:
		return slog.StringValue(h.redactText(ctx, v.Error()))
	case url.URL:
		return slog.StringValue(h.redactText(ctx, safeURLString(v)))
	case *url.URL:
		if v == nil {
			return slog.AnyValue(nil)
		}
		return slog.StringValue(h.redactText(ctx, safeURLString(*v)))
	case json.RawMessage:
		return slog.StringValue(typeMarker(value))
	case []byte:
		return slog.StringValue(typeMarker(value))
	case http.Header:
		return slog.StringValue(typeMarker(value))
	case http.Request, *http.Request, http.Response, *http.Response:
		return slog.StringValue(typeMarker(value))
	case io.Reader:
		return slog.StringValue(typeMarker(value))
	case string:
		return slog.StringValue(h.redactText(ctx, v))
	case bool:
		return slog.BoolValue(v)
	case int:
		return slog.IntValue(v)
	case int8:
		return slog.Int64Value(int64(v))
	case int16:
		return slog.Int64Value(int64(v))
	case int32:
		return slog.Int64Value(int64(v))
	case int64:
		return slog.Int64Value(v)
	case uint:
		return slog.Uint64Value(uint64(v))
	case uint8:
		return slog.Uint64Value(uint64(v))
	case uint16:
		return slog.Uint64Value(uint64(v))
	case uint32:
		return slog.Uint64Value(uint64(v))
	case uint64:
		return slog.Uint64Value(v)
	case float32:
		return slog.Float64Value(float64(v))
	case float64:
		return slog.Float64Value(v)
	case time.Duration:
		return slog.DurationValue(v)
	case time.Time:
		return slog.TimeValue(v)
	case []string:
		out := make([]string, 0, min(len(v), maxDynamicSlice))
		for i, item := range v {
			if i >= maxDynamicSlice {
				out = append(out, redactedValue)
				break
			}
			out = append(out, h.redactText(ctx, item))
		}
		return slog.AnyValue(out)
	}
	ref := reflect.ValueOf(value)
	if ref.IsValid() && ref.Kind() == reflect.Map {
		return h.redactMap(ctx, ref)
	}
	return slog.StringValue(typeMarker(value))
}

func (h *RedactingHandler) redactMap(ctx context.Context, value reflect.Value) slog.Value {
	if value.IsNil() {
		return slog.GroupValue()
	}
	if value.Type().Key().Kind() != reflect.String {
		return slog.StringValue(typeMarker(value.Interface()))
	}
	keys := value.MapKeys()
	sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
	attrs := make([]slog.Attr, 0, min(len(keys), maxDynamicAttrs))
	for i, key := range keys {
		if i >= maxDynamicAttrs {
			attrs = append(attrs, slog.String(redactedKeyName, redactedValue))
			break
		}
		item := value.MapIndex(key)
		if !item.IsValid() {
			continue
		}
		attrs = append(attrs, h.redactAttr(ctx, slog.Any(key.String(), item.Interface()), false))
	}
	return slog.GroupValue(mergeGroupAttrs(attrs)...)
}

func (h *RedactingHandler) redactText(ctx context.Context, value string) string {
	return redactText(ctx, value, h.registry)
}

func (h *RedactingHandler) isSensitiveKey(key string) bool {
	normalized := normalizeKey(key)
	if _, ok := h.sensitiveKeys[normalized]; ok {
		return true
	}
	if registeredSensitiveKey(normalized) {
		return true
	}
	for _, marker := range []string{"token", "secret", "password", "apikey", "privatekey"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
}

func fixedSensitiveKeys() map[string]struct{} {
	keys := map[string]struct{}{}
	for _, key := range []string{
		"api_key", "apikey", "api-token", "access_token", "refresh_token", "token",
		"secret", "client_secret", "password", "passwd", "pwd", "private_key",
		"authorization", "cookie", "set-cookie", "x-api-key", "credential", "credentials",
		"header", "headers", "body", "request_body", "response_body", "argv", "args",
	} {
		keys[normalizeKey(key)] = struct{}{}
	}
	return keys
}

func normalizeKey(key string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(key) {
		if r == '_' || r == '-' || r == '.' || unicode.IsSpace(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func sanitizeKeyName(key string) (string, bool) {
	return sanitizeDynamicName(key, redactedKeyName)
}

func sanitizeGroupName(key string) (string, bool) {
	return sanitizeDynamicName(key, redactedGroup)
}

func sanitizeDynamicName(name, replacement string) (string, bool) {
	if name == "" {
		return "", false
	}
	unsafe := false
	var b strings.Builder
	for _, r := range name {
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) || safety.IsDangerousUnicode(r) {
			unsafe = true
			continue
		}
		b.WriteRune(r)
		if b.Len() > maxDynamicName {
			unsafe = true
			break
		}
	}
	sanitized := safety.SanitizeTerminalLine(b.String())
	if sanitized == "" {
		return replacement, true
	}
	if unsafe || sanitized != name {
		return replacement, true
	}
	return sanitized, false
}

func sanitizeKey(key string) string {
	name, _ := sanitizeKeyName(key)
	return name
}

func safeURLString(raw url.URL) string {
	raw.User = nil
	raw.RawQuery = ""
	raw.ForceQuery = false
	raw.Fragment = ""
	raw.RawPath = ""
	raw.Scheme = strings.ToLower(raw.Scheme)
	raw.Host = strings.ToLower(safety.SanitizeTerminalLine(raw.Host))
	raw.Path = safety.SanitizeTerminalLine(raw.Path)
	raw.Opaque = safety.SanitizeTerminalLine(raw.Opaque)
	return raw.String()
}

func typeMarker(value any) string {
	if value == nil {
		return "[redacted-type:<nil>]"
	}
	name := reflect.TypeOf(value).String()
	name = safety.SanitizeTerminalLine(name)
	if name == "" {
		name = "unknown"
	}
	return "[redacted-type:" + name + "]"
}

func cloneStrings(in []string) []string {
	return append([]string(nil), in...)
}

func cloneAttrs(in []slog.Attr) []slog.Attr {
	return append([]slog.Attr(nil), in...)
}

func cloneGroupedAttrs(in []groupedAttrs) []groupedAttrs {
	out := make([]groupedAttrs, 0, len(in))
	for _, entry := range in {
		out = append(out, groupedAttrs{groups: cloneStrings(entry.groups), attrs: cloneAttrs(entry.attrs)})
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
