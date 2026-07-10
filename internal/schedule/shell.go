package schedule

import "strings"

func shellArg(s string) string {
	if s == "" {
		return "''"
	}
	if strings.IndexFunc(s, func(r rune) bool {
		return (r < 'A' || r > 'Z') &&
			(r < 'a' || r > 'z') &&
			(r < '0' || r > '9') &&
			!strings.ContainsRune("/._:-", r)
	}) == -1 {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
