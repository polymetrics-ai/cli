package connectors

import (
	"fmt"
	"net/url"
	"unicode/utf8"

	"polymetrics.ai/internal/safety"
)

// MaxDirectReadPageCursorBytes is the decoded-navigation admission bound. A
// continuation is provider output replayed by a caller, not a free-form query
// channel, so it remains small enough to validate before auth/signing.
const MaxDirectReadPageCursorBytes = 16 << 10

// ValidateDirectReadPageCursor is shared by the installed CLI, commandrunner,
// declarative engine, and native operation readers. It rejects values that
// could change terminal/log parsing or exceed either their literal or
// form-encoded transport budget before any provider I/O occurs.
func ValidateDirectReadPageCursor(cursor string) error {
	if cursor == "" {
		return nil
	}
	if !utf8.ValidString(cursor) {
		return fmt.Errorf("page cursor is not valid UTF-8")
	}
	if len(cursor) > MaxDirectReadPageCursorBytes {
		return fmt.Errorf("page cursor exceeds %d bytes", MaxDirectReadPageCursorBytes)
	}
	if err := safety.RejectDangerousChars(cursor, "page cursor"); err != nil {
		return err
	}
	// Query escaping is what cursor-token strategies put on the wire. Cap it
	// independently so a Unicode-heavy token cannot expand past the bounded
	// request surface after validation.
	if encoded := url.QueryEscape(cursor); len(encoded) > MaxDirectReadPageCursorBytes {
		return fmt.Errorf("page cursor percent-encoded form exceeds %d bytes", MaxDirectReadPageCursorBytes)
	}
	return nil
}
