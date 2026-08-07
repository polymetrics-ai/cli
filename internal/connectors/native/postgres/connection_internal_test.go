package postgres

import "testing"

// The captain's ruling is that MySQL and PostgreSQL must not drift into two
// spellings of the same transport-security choice. Existing libpq values keep
// their exact meaning; the canonical vocabulary is additionally accepted.
func TestSSLModeAcceptsTheSharedCanonicalVocabulary(t *testing.T) {
	for raw, want := range map[string]string{
		"disable":         "disable",
		"allow":           "allow",
		"prefer":          "prefer",
		"require":         "require",
		"verify-ca":       "verify-ca",
		"verify-full":     "verify-full",
		"disabled":        "disable",
		"preferred":       "prefer",
		"required":        "require",
		"verify-identity": "verify-full",
	} {
		got, err := libpqSSLMode(raw)
		if err != nil || got != want {
			t.Fatalf("libpqSSLMode(%q) = %q, %v, want %q", raw, got, err, want)
		}
	}
	if _, err := libpqSSLMode("not-a-mode"); err == nil {
		t.Fatal("libpqSSLMode() accepted an unknown sslmode")
	}
}
