package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

// A command's cli_surface notes are prose one author wrote on one command, so
// only ten of github's destructive commands carry a confirmation marker. The
// other 163 reach the identical typed gate, and an absent note reads exactly
// like "no confirmation needed" — which is the reading that sends an operator
// into an unexplained failure at run.
//
// So the help states the requirement from the bound write action instead. These
// cases are chosen to fail if that ever regresses to reading notes: `repo
// delete` is annotated, `release delete` and `repo deploy-key delete` are not,
// and all three must say the same thing.
func TestConnectorCommandHelpStatesConfirmationWithoutABundleNote(t *testing.T) {
	const want = "execution requires the typed confirmation --confirm destructive"

	for _, tc := range []struct {
		name  string
		path  []string
		noted bool
	}{
		{name: "annotated delete", path: []string{"repo", "delete"}, noted: true},
		{name: "unannotated delete", path: []string{"release", "delete"}},
		{name: "unannotated nested delete", path: []string{"repo", "deploy-key", "delete"}},
		{name: "unannotated declared destructive", path: []string{"repo", "archive"}, noted: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			help := connectorCommandHelp(t, tc.path)
			if !strings.Contains(help, "CONFIRMATION") || !strings.Contains(help, want) {
				t.Fatalf("pm github %s help omits the confirmation requirement:\n%s",
					strings.Join(tc.path, " "), help)
			}
			if noted := strings.Contains(help, "NOTES"); noted != tc.noted {
				t.Fatalf("pm github %s help NOTES presence = %v, want %v; the confirmation must not depend on it",
					strings.Join(tc.path, " "), noted, tc.noted)
			}
		})
	}
}

// The converse: a command that reaches no typed gate must not claim one, or the
// marker stops meaning anything. `issue create` is a plain approval-only write
// and `issue list` reads.
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
