package logging

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"unicode"

	"polymetrics.ai/internal/safety"
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
	attrs         []slog.Attr
	groups        []string
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
	for _, attr := range h.attrs {
		redacted.AddAttrs(h.redactAttr(ctx, attr, false))
	}
	var attrs []slog.Attr
	record.Attrs(func(attr slog.Attr) bool {
		attrs = append(attrs, attr)
		return true
	})
	for _, attr := range wrapGroups(h.groups, attrs) {
		redacted.AddAttrs(h.redactAttr(ctx, attr, false))
	}
	return h.next.Handle(ctx, redacted)
}

func (h *RedactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = append(append([]slog.Attr(nil), h.attrs...), wrapGroups(h.groups, attrs)...)
	clone.groups = append([]string(nil), h.groups...)
	return &clone
}

func (h *RedactingHandler) WithGroup(name string) slog.Handler {
	clone := *h
	clone.attrs = append([]slog.Attr(nil), h.attrs...)
	clone.groups = append(append([]string(nil), h.groups...), sanitizeKey(name))
	return &clone
}

func wrapGroups(groups []string, attrs []slog.Attr) []slog.Attr {
	out := append([]slog.Attr(nil), attrs...)
	for i := len(groups) - 1; i >= 0; i-- {
		group := sanitizeKey(groups[i])
		if group == "" {
			continue
		}
		out = []slog.Attr{{Key: group, Value: slog.GroupValue(out...)}}
	}
	return out
}

func (h *RedactingHandler) redactAttr(ctx context.Context, attr slog.Attr, parentSensitive bool) slog.Attr {
	attr.Key = sanitizeKey(attr.Key)
	attr.Value = attr.Value.Resolve()
	keySensitive := parentSensitive || (attr.Key != "" && h.isSensitiveKey(attr.Key))
	if attr.Value.Kind() == slog.KindGroup {
		attrs := attr.Value.Group()
		redacted := make([]slog.Attr, 0, len(attrs))
		for _, child := range attrs {
			redacted = append(redacted, h.redactAttr(ctx, child, keySensitive))
		}
		attr.Value = slog.GroupValue(redacted...)
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
		for _, attr := range attrs {
			redacted = append(redacted, h.redactAttr(ctx, attr, false))
		}
		return slog.GroupValue(redacted...)
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
	case map[string]string:
		attrs := make([]slog.Attr, 0, len(v))
		for key, item := range v {
			attrs = append(attrs, slog.String(key, item))
		}
		return h.redactValue(ctx, slog.GroupValue(attrs...))
	case map[string]any:
		attrs := make([]slog.Attr, 0, len(v))
		for key, item := range v {
			attrs = append(attrs, slog.Any(key, item))
		}
		return h.redactValue(ctx, slog.GroupValue(attrs...))
	case []string:
		out := make([]string, len(v))
		for i, item := range v {
			out[i] = h.redactText(ctx, item)
		}
		return slog.AnyValue(out)
	case fmt.Stringer:
		return slog.StringValue(h.redactText(ctx, v.String()))
	default:
		return slog.StringValue(h.redactText(ctx, fmt.Sprint(v)))
	}
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

func sanitizeKey(key string) string {
	return safety.SanitizeTerminalLine(key)
}

func safeURLString(raw url.URL) string {
	raw.User = nil
	raw.RawQuery = ""
	raw.ForceQuery = false
	raw.Fragment = ""
	raw.RawPath = ""
	raw.Host = sanitizeURLComponent(raw.Host)
	raw.Path = sanitizeURLComponent(raw.Path)
	raw.Opaque = sanitizeURLComponent(raw.Opaque)
	return raw.String()
}

func sanitizeURLComponent(value string) string {
	return safety.SanitizeTerminalLine(value)
}
