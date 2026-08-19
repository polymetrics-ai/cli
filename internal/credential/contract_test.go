package credential

import (
	"crypto/sha256"
	"testing"
)

func TestNormalizeStdinRemovesOnlyOneDocumentedTerminalDelimiter(t *testing.T) {
	canary := "credential-transport-canary"
	for _, tt := range []struct {
		name  string
		input string
		want  string
	}{
		{name: "LF", input: canary + "\n", want: canary},
		{name: "CRLF", input: canary + "\r\n", want: canary},
		{name: "extra LF remains", input: canary + "\n\n", want: canary + "\n"},
		{name: "extra CRLF remains", input: canary + "\r\n\r\n", want: canary + "\r\n"},
		{name: "bare CR remains", input: canary + "\r", want: canary + "\r"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeStdin(tt.input)
			if gotLength, wantLength := len(got), len(tt.want); gotLength != wantLength {
				t.Fatalf("normalized length = %d, want %d", gotLength, wantLength)
			}
			if gotHash, wantHash := sha256.Sum256([]byte(got)), sha256.Sum256([]byte(tt.want)); gotHash != wantHash {
				t.Fatalf("normalized SHA-256 = %x, want %x", gotHash, wantHash)
			}
		})
	}
}
