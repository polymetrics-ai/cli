package dbtest

import (
	"context"
	"errors"
	"testing"
)

// These cases call the real Start/Close path and inspect the existing runtime
// recorder. Expected immutable IDs are fixed fixture values, not harness output.
func TestImageCachePolicy199(t *testing.T) {
	for _, tc := range []struct {
		name, policy                                                   string
		missing, invalid, mismatch, retag, badDaemon, readinessFailure bool
		wantError, wantPull                                            bool
	}{
		{name: "cached", policy: "cache-only"},
		{name: "missing", policy: "cache-only", missing: true, wantError: true},
		{name: "invalid_identity", policy: "cache-only", invalid: true, wantError: true},
		{name: "tag_identity_mismatch", policy: "cache-only", mismatch: true, wantError: true},
		{name: "source_tag_changes_after_selected_inspection", policy: "cache-only", retag: true},
		{name: "daemon_refused", policy: "cache-only", badDaemon: true, wantError: true},
		{name: "readiness_failure_cleanup", policy: "cache-only", readinessFailure: true},
		{name: "default_pull", wantPull: true},
		{name: "explicit_pull", policy: "pull", wantPull: true},
		{name: "invalid_policy", policy: "sometimes", wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			recorder := &scriptedRunner{raceForeignRunImage: tc.mismatch}
			config := testConfig(recorder)
			config.ImagePolicy = tc.policy
			if !tc.missing {
				recorder.setImageWithID(config.Image, scriptedSourceImageID, true)
			}
			if tc.invalid {
				recorder.setImageWithID(config.Image, "malformed-image-id", true)
			}
			if tc.badDaemon {
				recorder.infoErr = errors.New("daemon rejected")
			}
			// Interpose only the existing command seam to replace the tag AFTER the
			// selected immutable-ID read; unformatted existence reads do not trigger it.
			wrapper := &cacheRetagRunner199{scriptedRunner: recorder, source: config.Image, retag: tc.retag}
			config.Run = wrapper
			h, err := New(config)
			if err == nil {
				t.Cleanup(func() {
					if e := h.Close(context.Background()); e != nil {
						t.Errorf("Close: %v", e)
					}
				})
				_, err = h.Start(context.Background())
			}
			if (err != nil) != tc.wantError {
				t.Errorf("Start/New error=%v, wantError=%v", err, tc.wantError)
			}
			if commandsContain(recorder.commands, "pull") != tc.wantPull {
				t.Errorf("pull observed=%v, want=%v", commandsContain(recorder.commands, "pull"), tc.wantPull)
			}
			start := commandIndex(recorder.commands, "run")
			if tc.wantError {
				if start >= 0 {
					t.Errorf("refusal created container: %v", recorder.commands[start])
				}
				return
			}
			if err != nil {
				return
			}
			if start < 0 {
				t.Fatal("no database start")
			}
			args := recorder.commands[start]
			if tc.policy == "cache-only" && !commandsContain([][]string{args}, "run", "--pull=never") {
				t.Error("cache-only start must explicitly forbid runtime implicit pulls")
			}
			if args[len(args)-1] != scriptedSourceImageID {
				t.Errorf("started image=%q, want original immutable ID", args[len(args)-1])
			}
			report := h.Report()
			wantPolicy := tc.policy
			if wantPolicy == "" {
				wantPolicy = "pull"
			}
			if report.ImagePolicy != wantPolicy || report.ImageID != scriptedSourceImageID {
				t.Errorf("image evidence=%+v", report)
			}
			if tc.readinessFailure {
				// Readiness belongs to the caller. Its failure must still run owned Close.
				if !recorder.containerLive {
					t.Fatal("readiness failure must follow real start")
				}
				if err := h.Close(context.Background()); err != nil {
					t.Fatal(err)
				}
				if recorder.containerLive || recorder.anonymousVolumePresent {
					t.Fatal("owned resources survived readiness failure cleanup")
				}
				if !recorder.imagePresent(config.Image) {
					t.Fatal("cleanup removed shared source image")
				}
				if commandsContain(recorder.commands, "volume") {
					t.Fatal("cleanup addressed arbitrary volume")
				}
			}
		})
	}
}

type cacheRetagRunner199 struct {
	*scriptedRunner
	source string
	retag  bool
}

func (r *cacheRetagRunner199) Run(ctx context.Context, endpoint string, args ...string) (string, error) {
	result, err := r.scriptedRunner.Run(ctx, endpoint, args...)
	if r.retag && commandHasPrefix(args, "image", "inspect") && args[len(args)-1] == r.source {
		if _, ok := commandFlagValue(args, "--format"); ok {
			r.setImageWithID(r.source, scriptedForeignImageID, true)
			r.retag = false
		}
	}
	return result, err
}

func TestImageCachePolicyRefusesOverrides199(t *testing.T) {
	for _, args := range [][]string{{"--pull=always"}, {"--pull", "always"}, {"--pull=missing"}} {
		runner := &scriptedRunner{}
		config := testConfig(runner)
		config.ImagePolicy = "cache-only"
		config.ContainerArgs = args
		if _, err := New(config); err == nil {
			t.Errorf("cache-only accepted competing pull arguments %q", args)
		}
		if len(runner.commands) != 0 {
			t.Fatal("invalid config performed runtime commands")
		}
	}
}
