package sensitive

import (
	"errors"
	"strings"
	"unicode"
)

// ValidatePublicText rejects secret-shaped, raw-diagnostic, and executable
// command content before it reaches any durable/public Shepherd surface. It is
// intentionally conservative: governed decision summaries are not a secret or
// command transport.
func ValidatePublicText(value string) error {
	return validate(value, false)
}

func ValidatePublicDocument(value string) error {
	return validate(value, true)
}

func ValidatePublicIdentifier(value string) error {
	if value == "" || len(value) > 256 || strings.HasPrefix(value, "/") || strings.HasSuffix(value, "/") ||
		strings.Contains(value, "//") {
		return errors.New("public identifier is empty, oversized, or malformed")
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return errors.New("public identifier contains an unsafe path segment")
		}
		for _, character := range segment {
			if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
				(character >= '0' && character <= '9') || strings.ContainsRune("._-", character) {
				continue
			}
			return errors.New("public identifier contains an unsafe character")
		}
	}
	return ValidatePublicText(value)
}

func validate(value string, allowNewlines bool) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return errors.New("public text is required")
	}
	for _, character := range trimmed {
		if character == 0 || character == '\r' || character == 0x7f ||
			(character == '\n' && !allowNewlines) || (character < 0x20 && character != '\t' && character != '\n') {
			return errors.New("public text contains control characters")
		}
	}
	lower := strings.ToLower(trimmed)
	for _, marker := range []string{
		"token", "password", "passwd", "secret", "credential", "authorization", "bearer ",
		"private key", "private_key", "api key", "api_key", "access key", "access_key",
		"client secret", "client_secret", "ghp_", "github_pat_", "gho_", "ghu_", "ghs_", "ghr_",
		"glpat-", "xoxb-", "xoxp-", "npm_", "pypi-", "sk-", "akia", "asiaj", "aiza", "ya29.",
		"-----begin ", "ssh-rsa ", "ssh-ed25519 ",
	} {
		if strings.Contains(lower, marker) {
			return errors.New("public text contains secret-shaped content")
		}
	}
	if strings.Contains(lower, "://") && strings.Contains(strings.SplitN(lower, "://", 2)[1], "@") {
		return errors.New("public text contains URI credential-shaped content")
	}
	for _, marker := range []string{
		"$(`", "$(", "`", ";", "&&", "||", "\x00", "rm -rf", "curl ", "wget ",
		"bash -c", "sh -c", "git push", "git merge", "gh api", "gh pr merge", "gh pr edit",
		"gh pr comment", "gh issue edit", "gh issue comment", "export ",
	} {
		if strings.Contains(lower, marker) {
			return errors.New("public text contains executable command-shaped content")
		}
	}
	if containsAccessKeyShape(trimmed) || containsOpaqueCredentialShape(trimmed) {
		return errors.New("public text contains access-key-shaped or opaque high-entropy content")
	}
	return nil
}

func containsOpaqueCredentialShape(value string) bool {
	start := -1
	for index := 0; index <= len(value); index++ {
		if index < len(value) && isOpaqueTokenCharacter(value[index]) {
			if start < 0 {
				start = index
			}
			continue
		}
		if start >= 0 && opaqueToken(value[start:index], value[:start]) {
			return true
		}
		start = -1
	}
	return false
}

func isOpaqueTokenCharacter(character byte) bool {
	return (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
		(character >= '0' && character <= '9') || strings.ContainsRune("_./+=-", rune(character))
}

func opaqueToken(value, prefix string) bool {
	if len(value) < 32 {
		return false
	}
	if looksPublicPath(value) {
		return false
	}
	categories := 0
	var lower, upper, digit, symbol, allHex bool
	allHex = true
	unique := make(map[byte]struct{}, 32)
	for index := 0; index < len(value); index++ {
		character := value[index]
		unique[character] = struct{}{}
		switch {
		case character >= 'a' && character <= 'z':
			lower = true
			if character > 'f' {
				allHex = false
			}
		case character >= 'A' && character <= 'Z':
			upper = true
			if character > 'F' {
				allHex = false
			}
		case character >= '0' && character <= '9':
			digit = true
		default:
			symbol, allHex = true, false
		}
	}
	if allHex {
		context := strings.ToLower(strings.TrimSpace(prefix))
		fields := strings.FieldsFunc(context, func(character rune) bool {
			return !unicode.IsLetter(character) && !unicode.IsDigit(character) && character != ':'
		})
		label := ""
		if len(fields) > 0 {
			label = fields[len(fields)-1]
		}
		if len(value) == 40 && (label == "commit" || label == "head") {
			return false
		}
		if len(value) == 64 && label == "sha256:" {
			return false
		}
		return len(unique) >= 12
	}
	if base32Token(value) && len(unique) >= 12 {
		return true
	}
	for _, present := range []bool{lower, upper, digit, symbol} {
		if present {
			categories++
		}
	}
	return categories >= 2 && len(unique) >= 16
}

func looksPublicPath(value string) bool {
	allowedPrefix := false
	for _, prefix := range []string{"internal/", "cmd/", "docs/", "website/", "agent-runtime/", "scripts/", ".planning/"} {
		if strings.HasPrefix(value, prefix) {
			allowedPrefix = true
			break
		}
	}
	if !allowedPrefix || strings.Contains(value, "//") {
		return false
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
		for _, character := range segment {
			if unicode.IsLetter(character) || unicode.IsDigit(character) || strings.ContainsRune("._-", character) {
				continue
			}
			return false
		}
	}
	return true
}

func base32Token(value string) bool {
	if len(value) < 32 || len(value)%8 != 0 {
		return false
	}
	for _, character := range value {
		if (character >= 'A' && character <= 'Z') || (character >= '2' && character <= '7') {
			continue
		}
		return false
	}
	return true
}

func containsAccessKeyShape(value string) bool {
	for index := 0; index+20 <= len(value); index++ {
		candidate := value[index : index+20]
		if !strings.HasPrefix(candidate, "AKIA") {
			continue
		}
		valid := true
		for _, character := range candidate[4:] {
			if !unicode.IsUpper(character) && !unicode.IsDigit(character) {
				valid = false
				break
			}
		}
		if valid {
			return true
		}
	}
	return false
}
