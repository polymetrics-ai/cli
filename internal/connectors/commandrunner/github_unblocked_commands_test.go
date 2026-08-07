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
			if tc.destructive {
				kind := connectors.ConfirmationForWriteAction(action).Kind
				if kind != connectors.ConfirmationKindDestructive {
					t.Fatalf("github %q reaches %s %s with confirmation %q; a destructive write must keep its typed confirmation",
						tc.command, action.Method, action.Path, kind)
				}
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

// The captain held two commands back explicitly. This test is the durable
// record of that hold: `auth token` prints a credential, which contradicts a
// standing rule, and `api` is the arbitrary-request escape hatch that bypasses
// every declared surface. Neither may drift to `implemented` without the
// captain's own decision.
func TestGitHubHeldCommandsStayBlocked(t *testing.T) {
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
		if cmd.Availability != "unsafe_or_disallowed" {
			t.Fatalf("github %q availability = %q, want unsafe_or_disallowed until the captain decides",
				path, cmd.Availability)
		}
	}
}
