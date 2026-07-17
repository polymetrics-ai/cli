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
	redacted := slog.NewRecord(record.Time, record.Level, h.redactText(record.Message), record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		redacted.AddAttrs(h.redactAttr(attr))
		return true
	})
	return h.next.Handle(ctx, redacted)
}

func (h *RedactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	redacted := make([]slog.Attr, 0, len(attrs))
	for _, attr := range attrs {
		redacted = append(redacted, h.redactAttr(attr))
	}
	return &RedactingHandler{next: h.next.WithAttrs(redacted), registry: h.registry, sensitiveKeys: h.sensitiveKeys}
}

func (h *RedactingHandler) WithGroup(name string) slog.Handler {
	return &RedactingHandler{next: h.next.WithGroup(sanitizeKey(name)), registry: h.registry, sensitiveKeys: h.sensitiveKeys}
}

func (h *RedactingHandler) redactAttr(attr slog.Attr) slog.Attr {
	attr.Key = sanitizeKey(attr.Key)
	if attr.Key == "" {
		return attr
	}
	if h.isSensitiveKey(attr.Key) {
		attr.Value = slog.StringValue(redactedValue)
		return attr
	}
	attr.Value = h.redactValue(attr.Value)
	return attr
}

func (h *RedactingHandler) redactValue(value slog.Value) slog.Value {
	value = value.Resolve()
	if value.Kind() == slog.KindGroup {
		attrs := value.Group()
		redacted := make([]slog.Attr, 0, len(attrs))
		for _, attr := range attrs {
			redacted = append(redacted, h.redactAttr(attr))
		}
		return slog.GroupValue(redacted...)
	}
	if value.Kind() == slog.KindString {
		return slog.StringValue(h.redactText(value.String()))
	}
	if value.Kind() != slog.KindAny {
		return value
	}
	return h.redactAny(value.Any())
}

func (h *RedactingHandler) redactAny(value any) slog.Value {
	switch v := value.(type) {
	case nil:
		return slog.AnyValue(nil)
	case error:
		return slog.StringValue(h.redactText(v.Error()))
	case url.URL:
		return slog.StringValue(safeURLString(v))
	case *url.URL:
		if v == nil {
			return slog.AnyValue(nil)
		}
		return slog.StringValue(safeURLString(*v))
	case map[string]string:
		attrs := make([]slog.Attr, 0, len(v))
		for key, item := range v {
			attrs = append(attrs, slog.String(key, item))
		}
		return h.redactValue(slog.GroupValue(attrs...))
	case map[string]any:
		attrs := make([]slog.Attr, 0, len(v))
		for key, item := range v {
			attrs = append(attrs, slog.Any(key, item))
		}
		return h.redactValue(slog.GroupValue(attrs...))
	case []string:
		out := make([]string, len(v))
		for i, item := range v {
			out[i] = h.redactText(item)
		}
		return slog.AnyValue(out)
	case fmt.Stringer:
		return slog.StringValue(h.redactText(v.String()))
	default:
		return slog.StringValue(h.redactText(fmt.Sprint(v)))
	}
}

func (h *RedactingHandler) redactText(value string) string {
	value = safety.SanitizeTerminal(value)
	value = safety.RedactErrorText(value)
	value = h.registry.redactString(value)
	return value
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
	return safety.SanitizeTerminal(key)
}

func safeURLString(raw url.URL) string {
	raw.User = nil
	raw.RawQuery = ""
	raw.ForceQuery = false
	raw.Fragment = ""
	return raw.String()
}
