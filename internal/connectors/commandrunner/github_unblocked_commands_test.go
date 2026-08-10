package commandrunner

import (
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
)

// The captain's order restores gh-familiar command names that were shipped
// `unsafe_or_disallowed` without his authorisation. Each one is restored by
// pointing it at a write action the connector already executes, so this test
// asserts the two things that make the restoration honest rather than
// cosmetic: the command reaches the real runtime preflight, and every
// destructive endpoint still arrives at the typed-confirmation gate it would
// have had if it had never been blocked.
//
// `repo delete` is deliberately in this list. Blocking it never made a
// repository safer — DELETE /repos/{owner}/{repo} is reachable today as
// `repo delete-2` — it only hid the destructive path behind a name nobody
// would guess.
func TestGitHubRestoredCommandsAreExecutable(t *testing.T) {
	cases := []struct {
		command     string
		write       string
		destructive bool
	}{
		{command: "repo create", write: "repos_create_for_authenticated_user"},
		{command: "repo delete", write: "repo", destructive: true},
		{command: "repo archive", write: "archive_repo"},
		{command: "repo unarchive", write: "unarchive_repo"},
		{command: "cache delete", write: "actions_caches_cache_id", destructive: true},
		{command: "secret set", write: "actions_secrets_secret_name3"},
		{command: "secret delete", write: "actions_secrets_secret_name", destructive: true},
	}

	registry := bundleregistry.New()
	connector, ok := registry.Get("github")
	if !ok {
		t.Fatal("github connector is not registered")
	}
	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		t.Fatal("github connector exposes no command surface")
	}

	commands := map[string]connectors.CommandSurfaceCommand{}
	for _, cmd := range provider.CommandSurface().Commands {
		commands[cmd.Path] = cmd
	}

	manifest := connectors.ManifestOf(connector)
	if len(manifest.WriteActions) == 0 {
		t.Fatal("github connector exposes no write actions")
	}

	for _, tc := range cases {
		t.Run(tc.command, func(t *testing.T) {
			cmd, ok := commands[tc.command]
			if !ok {
				t.Fatalf("github command %q is not declared", tc.command)
			}
			if cmd.Availability != "implemented" {
				t.Fatalf("github %q availability = %q, want implemented", tc.command, cmd.Availability)
			}
			if cmd.Intent != "reverse_etl" {
				t.Fatalf("github %q intent = %q, want reverse_etl", tc.command, cmd.Intent)
			}
			if cmd.Write != tc.write {
				t.Fatalf("github %q write = %q, want %q", tc.command, cmd.Write, tc.write)
			}
			if strings.TrimSpace(cmd.Approval) == "" {
				t.Fatalf("github %q must declare an approval requirement", tc.command)
			}

			action, ok := findWriteAction(manifest, tc.write)
			if !ok {
				t.Fatalf("github write action %q is not exposed by the connector", tc.write)
			}

			// A destructive endpoint must still demand the closed typed
			// confirmation. Assert against the same resolver the runtime uses,
			// not against the raw metadata field, so an action that inherits
			// the confirmation from its DELETE method counts and an action
			// that quietly loses it fails.
			kind := connectors.ConfirmationForWriteAction(action).Kind
			if tc.destructive {
				if kind != connectors.ConfirmationKindDestructive {
					t.Fatalf("github %q reaches %s %s with confirmation %q; a destructive write must keep its typed confirmation",
						tc.command, action.Method, action.Path, kind)
				}
			} else if kind != "" {
				// The captain classified `repo create`, `secret set`,
				// `repo archive` and `repo unarchive` as approval-only writes.
				// Adding a typed challenge to one is as much a drift from that
				// decision as dropping one from a delete, and `repo archive`
				// carried exactly that drift: it declared `confirm: destructive`
				// on an update the decision names as non-destructive.
				t.Fatalf("github %q reaches %s %s with confirmation %q; an approval-only write must not gain a typed challenge",
					tc.command, action.Method, action.Path, kind)
			}

			if err := Preflight(connector, strings.Fields(tc.command)); err != nil {
				reason := err.Error()
				var blocked *BlockedCommandError
				if errors.As(err, &blocked) {
					reason = blocked.Reason
				}
				t.Fatalf("github %q does not pass runtime preflight: %s", tc.command, reason)
			}
		})
	}
}

// `auth token` and `api` are not provider operations: one would print a
// stored credential and the other would be a generic authenticated transport
// escape hatch. They must remain non-executable, but they are classified as
// unsupported_local rather than unsafe_or_disallowed so the GitHub documented
// surface has no safety-label loophole.
func TestGitHubCapabilityEscapesStayNonExecutableWithoutUnsafeClassification(t *testing.T) {
	held := []string{"auth token", "api"}

	registry := bundleregistry.New()
	connector, ok := registry.Get("github")
	if !ok {
		t.Fatal("github connector is not registered")
	}
	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		t.Fatal("github connector exposes no command surface")
	}

	commands := map[string]connectors.CommandSurfaceCommand{}
	for _, cmd := range provider.CommandSurface().Commands {
		commands[cmd.Path] = cmd
	}

	for _, path := range held {
		cmd, ok := commands[path]
		if !ok {
			t.Fatalf("github command %q is not declared", path)
		}
		if cmd.Availability != "unsupported_local" {
			t.Fatalf("github %q availability = %q, want unsupported_local for a deliberately absent local capability",
				path, cmd.Availability)
		}
		if strings.TrimSpace(cmd.Write) != "" {
			t.Fatalf("github %q binds write action %q; a held command must reach no write executor",
				path, cmd.Write)
		}
		if err := Preflight(connector, strings.Fields(path)); err == nil {
			t.Fatalf("github %q passes runtime preflight; a held command must be refused", path)
		}
	}
}

// The historical gh-style aliases below now reuse exact declared REST or
// fixed-document GraphQL contracts. Preflight is the runtime-owned admission
// seam: it proves the aliases are not help-only strings and cannot drift from
// the same command dispatch the binary invokes.
func TestGitHubLegacyAliasesPassRuntimePreflight(t *testing.T) {
	aliases := []string{
		"issue view", "pr view", "release view", "ruleset view", "run view", "workflow view",
		"discussion create", "issue status", "pr checks", "pr status", "project create", "ruleset check", "search prs", "status",
		"issue delete", "issue transfer", "pr revert",
	}

	registry := bundleregistry.New()
	connector, ok := registry.Get("github")
	if !ok {
		t.Fatal("github connector is not registered")
	}
	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		t.Fatal("github connector exposes no command surface")
	}
	commands := map[string]connectors.CommandSurfaceCommand{}
	for _, cmd := range provider.CommandSurface().Commands {
		commands[cmd.Path] = cmd
	}
	for _, path := range aliases {
		t.Run(path, func(t *testing.T) {
			cmd, found := commands[path]
			if !found {
				t.Fatalf("github command %q is not declared", path)
			}
			if cmd.Availability != "implemented" {
				t.Fatalf("github %q availability = %q, want implemented", path, cmd.Availability)
			}
			if err := Preflight(connector, strings.Fields(path)); err != nil {
				t.Fatalf("github %q does not pass runtime preflight: %v", path, err)
			}
		})
	}
}

func TestGitHubGraphQLDestructiveAliasesRequireTypedConfirmation(t *testing.T) {
	aliases := []string{"issue delete", "issue transfer", "pr revert"}
	registry := bundleregistry.New()
	connector, ok := registry.Get("github")
	if !ok {
		t.Fatal("github connector is not registered")
	}
	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		t.Fatal("github connector exposes no command surface")
	}
	commands := map[string]connectors.CommandSurfaceCommand{}
	for _, cmd := range provider.CommandSurface().Commands {
		commands[cmd.Path] = cmd
	}
	for _, path := range aliases {
		t.Run(path, func(t *testing.T) {
			cmd := commands[path]
			if got := ConfirmationChallengeForCommand(connector, cmd); got != string(connectors.ConfirmationKindDestructive) {
				t.Fatalf("github %q confirmation challenge = %q, want typed destructive confirmation", path, got)
			}
		})
	}
}

// A note that restates the typed confirmation is a second source for a fact the
// help already derives from the write action, and the two drift: the note was
// carried by ten of github's 173 destructive commands, which made its absence
// on the other 163 read as "no confirmation needed". The derived
// ConfirmationChallengeForCommand is now the only source, so no bundle may
// reintroduce the copy.
//
// The assertion is on every command, not just the ten, because the failure this
// prevents is a note appearing somewhere new.
func TestGitHubNotesDoNotRestateTheTypedConfirmation(t *testing.T) {
	registry := bundleregistry.New()
	connector, ok := registry.Get("github")
	if !ok {
		t.Fatal("github connector is not registered")
	}
	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		t.Fatal("github connector exposes no command surface")
	}

	confirmable := 0
	for _, cmd := range provider.CommandSurface().Commands {
		if strings.Contains(strings.ToLower(cmd.Notes), "--confirm") {
			t.Errorf("github %q restates the typed confirmation in notes (%q); the derived help owns that fact",
				cmd.Path, cmd.Notes)
		}
		if ConfirmationChallengeForCommand(connector, cmd) != "" {
			confirmable++
		}
	}
	if confirmable == 0 {
		t.Fatal("no github command resolves a typed confirmation; the resolver the help depends on is not working")
	}
}
