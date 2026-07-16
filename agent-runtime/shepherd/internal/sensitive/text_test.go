package sensitive

import "testing"

func TestValidatePublicTextRejectsSecretAndCommandShapes(t *testing.T) {
	t.Parallel()
	for _, value := range []string{
		"password: example", "TOKEN example", "github_pat_example", "ghp_example",
		"-----BEGIN PRIVATE KEY-----", "https://user:pass@example.invalid", "AKIAABCDEFGHIJKLMNOP",
		"curl https://example.invalid", "continue && publish", "$(whoami)",
		"git push origin main", "gh pr merge 390", "gh issue comment 389", "export KEY=value",
		"glpat-example", "xoxb-example", "sk-example",
		"mF9xQ2vL7pR4sT8uW1yZ6aB3cD5eG0hJ2kN8qP4rS7tV",
		"0123456789abcdef89abcdef01234567",
		"0123456789abcdef89abcdef012345670123456789abcdef89abcdef01234567",
		"MFRGGZDFMZTWQ2LKNNWG23TPOI5DA7BR",
		"ABCDEFGHIJKLMNOPQRSTUVWXYZABCDEF",
		"AbCdEfGhIjKlMnOpQrStUvWx/Yz0123456789abcdef",
		"overhead 0123456789abcdef0123456789abcdef01234567",
	} {
		if err := ValidatePublicText(value); err == nil {
			t.Fatalf("unsafe public text accepted: %q", value)
		}
	}
	for _, value := range []string{
		"Continue with the bounded implementation", "Add only internal/connectors/defs/defs_test.go",
		"retry budget exhausted", "approved issue context",
		"commit aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"head 0123456789abcdef0123456789abcdef01234567",
		"sha256:0123456789abcdef89abcdef012345670123456789abcdef89abcdef01234567",
	} {
		if err := ValidatePublicText(value); err != nil {
			t.Fatalf("safe public text %q rejected: %v", value, err)
		}
	}
}
