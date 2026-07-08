package certify

import "testing"

func TestDefaultStreamName(t *testing.T) {
	if got := defaultStreamName("sample"); got != "customers" {
		t.Fatalf("defaultStreamName(sample) = %q, want customers", got)
	}
	if got := defaultStreamName("github"); got != "issues" {
		t.Fatalf("defaultStreamName(github) = %q, want issues", got)
	}
}
