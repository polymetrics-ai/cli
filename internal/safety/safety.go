package safety

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var rawURLPattern = regexp.MustCompile(`(?i)\b[a-z][a-z0-9+.-]*://[^\s<>"']+`)
var jsonBodyPattern = regexp.MustCompile(`: \{.*\}$`)
var secretAssignmentPattern = regexp.MustCompile(`(?i)(api[_-]?key|access[_-]?token|token|secret|password)=([^\s&]+)`)

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

func SanitizeTerminalLine(text string) string {
	text = SanitizeTerminal(text)
	var b strings.Builder
	lastSpace := false
	for _, r := range text {
		if r == '\n' || r == '\t' {
			if !lastSpace {
				b.WriteByte(' ')
				lastSpace = true
			}
			continue
		}
		b.WriteRune(r)
		lastSpace = r == ' '
	}
	return strings.TrimSpace(b.String())
}

func RedactErrorText(text string) string {
	text = rawURLPattern.ReplaceAllStringFunc(text, redactURL)
	text = jsonBodyPattern.ReplaceAllString(text, ": [redacted]")
	text = secretAssignmentPattern.ReplaceAllString(text, "$1=[redacted]")
	return text
}

func redactURL(raw string) string {
	suffix := ""
	for len(raw) > 0 {
		last := raw[len(raw)-1]
		if !strings.ContainsRune(":,.!?;)]}", rune(last)) {
			break
		}
		suffix = string(last) + suffix
		raw = raw[:len(raw)-1]
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "[redacted-url]" + suffix
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	parsed.RawPath = ""
	parsed.Host = strings.ToLower(SanitizeTerminalLine(parsed.Host))
	parsed.Path = SanitizeTerminalLine(parsed.Path)
	parsed.Opaque = SanitizeTerminalLine(parsed.Opaque)
	if parsed.Host == "" {
		return "[redacted-url]" + suffix
	}
	return parsed.String() + suffix
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

func ValidateLocalWritePath(projectRoot, value, field string, allowExternal bool) error {
	_, err := ResolveLocalWritePath(projectRoot, value, field, allowExternal)
	return err
}

func ResolveLocalWritePath(projectRoot, value, field string, allowExternal bool) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	if err := RejectDangerousChars(value, field); err != nil {
		return "", err
	}
	rootAbs, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", fmt.Errorf("resolve project root: %w", err)
	}
	var pathAbs string
	if filepath.IsAbs(value) {
		pathAbs = filepath.Clean(value)
	} else {
		clean := filepath.Clean(value)
		if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
			return "", fmt.Errorf("%s must not escape the project root", field)
		}
		pathAbs = filepath.Join(rootAbs, clean)
	}
	insideProject, err := pathWithin(rootAbs, pathAbs)
	if err != nil {
		return "", fmt.Errorf("compare %s to project root: %w", field, err)
	}
	if !insideProject {
		if allowExternal {
			return pathAbs, nil
		}
		return "", fmt.Errorf("%s outside the project root requires allow_external_path=true", field)
	}
	if allowExternal {
		return pathAbs, nil
	}

	resolvedRoot, err := resolveThroughNearestExistingAncestor(rootAbs)
	if err != nil {
		return "", fmt.Errorf("resolve project root: %w", err)
	}
	resolvedPath, err := resolveThroughNearestExistingAncestor(pathAbs)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", field, err)
	}
	insideProject, err = pathWithin(resolvedRoot, resolvedPath)
	if err != nil {
		return "", fmt.Errorf("compare resolved %s to project root: %w", field, err)
	}
	if !insideProject {
		return "", fmt.Errorf("%s resolves outside the project root and requires allow_external_path=true", field)
	}
	return pathAbs, nil
}

func pathWithin(root, path string) (bool, error) {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false, err
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)), nil
}

func resolveThroughNearestExistingAncestor(path string) (string, error) {
	current := filepath.Clean(path)
	var missing []string
	for {
		_, err := os.Lstat(current)
		if err == nil {
			resolved, err := filepath.EvalSymlinks(current)
			if err != nil {
				return "", err
			}
			parts := make([]string, 0, len(missing)+1)
			parts = append(parts, resolved)
			parts = append(parts, missing...)
			return filepath.Join(parts...), nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", err
		}
		missing = append([]string{filepath.Base(current)}, missing...)
		current = parent
	}
}
