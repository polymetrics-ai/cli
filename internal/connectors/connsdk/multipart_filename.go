package connsdk

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

// MultipartFileNames derives a wire filename from its original logical name.
// Paths remain file custody inputs; the encoded name is never a source path.
func MultipartFileNames(file MultipartFile) (logical, wire string, err error) {
	logical = file.FileName
	if logical == "" {
		logical = filepath.Base(file.sourceName())
	}
	if logical == "" || logical == "." || logical == ".." || !utf8.ValidString(logical) || strings.ContainsAny(logical, "/\\") {
		return "", "", fmt.Errorf("multipart file %q has invalid logical filename", file.FieldName)
	}
	for _, r := range logical {
		if unicode.IsControl(r) {
			return "", "", fmt.Errorf("multipart file %q filename contains a control", file.FieldName)
		}
	}
	switch file.FilenameEncoding {
	case "", "identity":
		return logical, logical, nil
	case "url_percent_utf8":
		const hex = "0123456789ABCDEF"
		var encoded strings.Builder
		for i := 0; i < len(logical); i++ {
			c := logical[i]
			if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '.' || c == '_' || c == '~' {
				encoded.WriteByte(c)
			} else {
				encoded.WriteByte('%')
				encoded.WriteByte(hex[c>>4])
				encoded.WriteByte(hex[c&15])
			}
		}
		return logical, encoded.String(), nil
	default:
		return "", "", fmt.Errorf("multipart file %q has unsupported filename encoding", file.FieldName)
	}
}
