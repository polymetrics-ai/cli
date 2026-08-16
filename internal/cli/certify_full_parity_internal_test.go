package cli

import (
	"strings"
	"testing"
)

func TestCertifyOptionsFullParityEnablesFullWriteAndRejectsWriteSkip(t *testing.T) {
	opts, err := certifyOptionsFromFlags("github", parseFlags([]string{"github", "--full-parity"}))
	if err != nil {
		t.Fatalf("certifyOptionsFromFlags(--full-parity) = %v", err)
	}
	if !opts.Full || !opts.Write || !opts.RequireFullParity {
		t.Fatalf("--full-parity options = %+v, want full, write, and parity requirement", opts)
	}

	_, err = certifyOptionsFromFlags("github", parseFlags([]string{"github", "--full-parity", "--skip", "write"}))
	if err == nil || !strings.Contains(err.Error(), "cannot skip write") {
		t.Fatalf("--full-parity --skip write error = %v, want refusal", err)
	}
}
