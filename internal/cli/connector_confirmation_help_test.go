package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

// The typed confirmation is stated once, by the help, resolved from the bound
// write action. It used to be stated twice on ten github commands: a derived
// CONFIRMATION section and a hand-authored NOTES string three lines below it.
//
// The notes are gone, so these cases are chosen to fail if the help ever
// regresses to reading them: none of these commands carries a note now, and two
// of them (`repo delete`, `repo archive`) are ones that did.
func TestConnectorCommandHelpStatesTheTypedConfirmation(t *testing.T) {
	const want = "execution requires the typed confirmation --confirm destructive"

	for _, path := range [][]string{
		{"repo", "delete"},
		{"repo", "archive"},
		{"repo", "unarchive"},
		{"transfer", "create"},
		{"release", "delete"},
		{"repo", "deploy-key", "delete"},
	} {
		t.Run(strings.Join(path, " "), func(t *testing.T) {
			help := connectorCommandHelp(t, path)
			if !strings.Contains(help, "CONFIRMATION") || !strings.Contains(help, want) {
				t.Fatalf("pm github %s help omits the confirmation requirement:\n%s",
					strings.Join(path, " "), help)
			}
			if strings.Contains(help, "requires typed confirmation --confirm") {
				t.Fatalf("pm github %s help restates the confirmation as a note; the derived section is the only source:\n%s",
					strings.Join(path, " "), help)
			}
		})
	}
}

// The converse: a command that reaches no typed gate must not claim one, or the
// marker stops meaning anything. `issue create`, `repo create` and `secret set`
// are approval-only writes and `issue list` reads.
func TestConnectorCommandHelpOmitsConfirmationWhereNoneIsDemanded(t *testing.T) {
	for _, path := range [][]string{{"issue", "create"}, {"repo", "create"}, {"secret", "set"}, {"issue", "list"}} {
		t.Run(strings.Join(path, " "), func(t *testing.T) {
			help := connectorCommandHelp(t, path)
			if strings.Contains(help, "CONFIRMATION") {
				t.Fatalf("pm github %s help claims a confirmation it does not demand:\n%s",
					strings.Join(path, " "), help)
			}
		})
	}
}

func connectorCommandHelp(t *testing.T, path []string) string {
	t.Helper()
	var stdout, stderr bytes.Buffer
	args := append(append([]string{"github"}, path...), "--help")
	if code := cli.Run(args, &stdout, &stderr); code != 0 {
		t.Fatalf("cli.Run(%v) code = %d stderr = %s", args, code, stderr.String())
	}
	return stdout.String()
}
