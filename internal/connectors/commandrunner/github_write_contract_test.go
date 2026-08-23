package commandrunner

import (
	"context"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
)

// These are the two contracts the runtime actually enforces for a github
// reverse-ETL command, and the approval text is where `pm github <cmd> --help`,
// MANUAL.md and SKILL.md publish them.
//
// A safe write does not require a preview. PlanConnectorCommand mints the
// approval token at plan time for any command with no typed confirmation and no
// bound operation, and RunReverseETL's planRequiresPersistedPreview is false for
// exactly that command, so `plan` then the stdin approval marker executes. Telling the
// operator a preview is required is not a harmless over-statement: it describes
// a gate that is not there, which is the same class of error as describing a
// gate that is there as absent.
const (
	githubSafeWriteApproval        = "Reverse ETL writes require plan, approval, execute; preview is optional."
	githubDestructiveWriteApproval = "Reverse ETL writes require plan, preview, approval, execute."
	githubBinaryUploadApproval     = "Binary uploads require plan, preview, approval, execute."
)

func githubCommandSurface(t *testing.T) (connectors.Connector, []connectors.CommandSurfaceCommand) {
	t.Helper()
	registry := bundleregistry.New()
	connector, ok := registry.Get("github")
	if !ok {
		t.Fatal("github connector is not registered")
	}
	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		t.Fatal("github connector exposes no command surface")
	}
	return connector, provider.CommandSurface().Commands
}

// The bundle shipped one blanket approval sentence on all 525 reverse-ETL
// commands, claiming every write requires a preview. It is true of the 176 that
// carry a typed confirmation and false of the other 349, which hand the caller
// an approval token straight out of `plan`.
//
// The assertion is on the derived confirmation rather than on a list of command
// names, so a command that gains or loses its typed confirmation is required to
// move its approval text with it.
func TestGitHubApprovalTextMatchesTheEnforcedWriteContract(t *testing.T) {
	connector, commands := githubCommandSurface(t)

	safe, destructive := 0, 0
	for _, cmd := range commands {
		if cmd.Intent != "reverse_etl" {
			continue
		}
		want := githubSafeWriteApproval
		if ConfirmationChallengeForCommand(connector, cmd) != "" {
			want = githubDestructiveWriteApproval
			destructive++
		} else {
			safe++
		}
		if cmd.Approval != want {
			t.Errorf("github %q approval = %q, want %q", cmd.Path, cmd.Approval, want)
		}
	}
	if safe == 0 || destructive == 0 {
		t.Fatalf("github reverse_etl commands: safe=%d destructive=%d; both classes must exist for this test to mean anything", safe, destructive)
	}
}

// This is the installed GitHub bundle's public upload path, rather than a
// hand-built test connector. It proves the hand-rolled command surface maps
// only its declared source field into the existing approval-bound write plan,
// and that Run cannot bypass that plan to send an arbitrary body.
func TestGitHubReleaseAssetUploadBuildsBoundBinaryPlan(t *testing.T) {
	connector, commands := githubCommandSurface(t)
	var surface connectors.CommandSurfaceCommand
	for _, command := range commands {
		if command.Path == "releases assets upload" {
			surface = command
			break
		}
	}
	if surface.Path == "" {
		t.Fatal("github release asset upload command is not declared")
	}
	if surface.Intent != "binary_upload" || surface.Availability != "implemented" || surface.Write != "releases_release_id_assets2" {
		t.Fatalf("github release asset upload surface = %+v, want an implemented declared binary-upload write", surface)
	}
	if surface.Approval != githubBinaryUploadApproval {
		t.Fatalf("github release asset upload approval = %q, want %q", surface.Approval, githubBinaryUploadApproval)
	}

	request := Request{
		Path: strings.Fields(surface.Path),
		Flags: map[string][]string{
			"file-path":  {"release.tgz"},
			"name":       {"release.tgz"},
			"release-id": {"42"},
		},
	}
	plan, err := BuildWriteCommand(context.Background(), connector, request)
	if err != nil {
		t.Fatalf("BuildWriteCommand(github release asset upload): %v", err)
	}
	if plan.Write != surface.Write || !plan.ApprovalRequired || plan.Record["file_path"] != "release.tgz" {
		t.Fatalf("binary upload plan = %+v, want declared source in an approval-bound plan", plan)
	}
	if _, err := Run(context.Background(), connector, request, func(connectors.Record) error {
		t.Fatal("github release asset upload bypassed plan, preview, approval, execute")
		return nil
	}); err == nil || !strings.Contains(err.Error(), "plan, preview, approval, execute") {
		t.Fatalf("Run(github release asset upload) error = %v, want lifecycle block", err)
	}
}

// A gh-familiar alias and the documented-surface command generated for the same
// endpoint are the same DELETE under two names. `repo delete` was withheld while
// `repo delete-2` ran the identical request, so the pair is the exact shape the
// restoration had to fix: whatever contract one carries, the other must carry,
// or blocking a name goes back to hiding a capability instead of removing it.
//
// The pairs are discovered from the bundle rather than listed, so a future
// generated twin is covered the day it lands.
func TestGitHubGeneratedTwinsShareTheirAliasWriteContract(t *testing.T) {
	connector, commands := githubCommandSurface(t)

	byWrite := map[string][]connectors.CommandSurfaceCommand{}
	for _, cmd := range commands {
		if cmd.Intent != "reverse_etl" || strings.TrimSpace(cmd.Write) == "" {
			continue
		}
		byWrite[cmd.Write] = append(byWrite[cmd.Write], cmd)
	}

	pairs := 0
	for write, group := range byWrite {
		if len(group) < 2 {
			continue
		}
		pairs++
		first := group[0]
		wantConfirmation := ConfirmationChallengeForCommand(connector, first)
		for _, cmd := range group[1:] {
			if got := ConfirmationChallengeForCommand(connector, cmd); got != wantConfirmation {
				t.Errorf("github %q and %q both write %q but resolve confirmations %q and %q",
					first.Path, cmd.Path, write, wantConfirmation, got)
			}
			if cmd.Availability != first.Availability {
				t.Errorf("github %q and %q both write %q but declare availability %q and %q",
					first.Path, cmd.Path, write, first.Availability, cmd.Availability)
			}
			if cmd.Approval != first.Approval {
				t.Errorf("github %q and %q both write %q but declare approval %q and %q",
					first.Path, cmd.Path, write, first.Approval, cmd.Approval)
			}
			// `risk` is deliberately not compared. Github states it as free prose
			// on the gh-familiar aliases ("Locks issue conversation for repository
			// users.") and as a level on the generated twins, and two names for
			// one endpoint may honestly describe different callers' exposure.
			// Confirmation, availability and approval are the contract; risk is
			// the commentary on it.
		}
	}
	if pairs == 0 {
		t.Fatal("no github write action is reachable under two command names; the drift this test guards cannot occur, which means the surface changed shape")
	}

	// Pin the pair the restoration turned on by name. Discovery above proves the
	// invariant holds across the surface; this proves it holds for the command
	// the captain's decision is actually about, so a rename cannot quietly drop
	// it from coverage.
	byPath := map[string]connectors.CommandSurfaceCommand{}
	for _, cmd := range commands {
		byPath[cmd.Path] = cmd
	}
	alias, ok := byPath["repo delete"]
	if !ok {
		t.Fatal("github command \"repo delete\" is not declared")
	}
	twin, ok := byPath["repo delete-2"]
	if !ok {
		t.Fatal("github command \"repo delete-2\" is not declared")
	}
	if alias.Write != twin.Write {
		t.Fatalf("github \"repo delete\" writes %q but \"repo delete-2\" writes %q; the two names must reach one action", alias.Write, twin.Write)
	}
	for _, cmd := range []connectors.CommandSurfaceCommand{alias, twin} {
		if got := ConfirmationChallengeForCommand(connector, cmd); got != string(connectors.ConfirmationKindDestructive) {
			t.Errorf("github %q resolves confirmation %q, want %q", cmd.Path, got, connectors.ConfirmationKindDestructive)
		}
		if cmd.Approval != githubDestructiveWriteApproval {
			t.Errorf("github %q approval = %q, want %q", cmd.Path, cmd.Approval, githubDestructiveWriteApproval)
		}
	}
}
