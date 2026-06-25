package safety

import (
	"fmt"
	"path/filepath"
	"strings"
)

func IsDangerousUnicode(r rune) bool {
	switch {
	case r >= '\u200B' && r <= '\u200D':
		return true
	case r == '\uFEFF':
		return true
	case r >= '\u202A' && r <= '\u202E':
		return true
	case r >= '\u2028' && r <= '\u2029':
		return true
	case r >= '\u2066' && r <= '\u2069':
		return true
	default:
		return false
	}
}

func SanitizeTerminal(text string) string {
	var b strings.Builder
	for _, r := range text {
		if r == '\n' || r == '\t' {
			b.WriteRune(r)
			continue
		}
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) {
			continue
		}
		if IsDangerousUnicode(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func RejectDangerousChars(value, field string) error {
	for _, r := range value {
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) {
			return fmt.Errorf("%s contains invalid control characters", field)
		}
		if IsDangerousUnicode(r) {
			return fmt.Errorf("%s contains invalid unicode characters", field)
		}
	}
	return nil
}

func ValidateIdentifier(value, field string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}
	if err := RejectDangerousChars(value, field); err != nil {
		return err
	}
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-' || r == '.':
		default:
			return fmt.Errorf("%s contains invalid character %q", field, r)
		}
	}
	if strings.Contains(value, "..") {
		return fmt.Errorf("%s must not contain path traversal", field)
	}
	return nil
}

func ValidateRelativePath(value, field string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", field)
	}
	if err := RejectDangerousChars(value, field); err != nil {
		return err
	}
	if filepath.IsAbs(value) {
		return fmt.Errorf("%s must be relative", field)
	}
	clean := filepath.Clean(value)
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("%s must not escape the current directory", field)
	}
	return nil
}
