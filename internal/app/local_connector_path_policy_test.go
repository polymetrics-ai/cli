package app_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/app"
)

func TestResolvedLocalConnectorRelativePathUsesSelectedProjectRoot(t *testing.T) {
	ctx := context.Background()
	cwd := t.TempDir()
	t.Chdir(cwd)

	for _, connectorName := range []string{"warehouse", "outbox"} {
		t.Run(connectorName, func(t *testing.T) {
			root := t.TempDir()
			if err := app.InitProject(root); err != nil {
				t.Fatalf("InitProject() error = %v", err)
			}
			a, err := app.Open(root)
			if err != nil {
				t.Fatalf("Open() error = %v", err)
			}
			credentialName := connectorName + "-relative"
			if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
				Name:      credentialName,
				Connector: connectorName,
				Config:    map[string]string{"path": "relative-effect"},
			}); err != nil {
				t.Fatalf("AddCredential() error = %v", err)
			}
			connector, runtime, err := a.ResolveConnectorCredential(ctx, connectorName, credentialName, nil)
			if err != nil {
				t.Fatalf("ResolveConnectorCredential() error = %v", err)
			}
			if err := connector.Check(ctx, runtime); err != nil {
				t.Fatalf("Check() error = %v", err)
			}
			if _, err := os.Stat(filepath.Join(root, "relative-effect")); err != nil {
				t.Fatalf("selected-root effect missing: %v", err)
			}
			if _, err := os.Stat(filepath.Join(cwd, "relative-effect")); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("relative effect escaped to cwd: %v", err)
			}
		})
	}
}

func TestLocalConnectorCheckRevalidatesPathAfterCredentialResolution(t *testing.T) {
	ctx := context.Background()
	for _, connectorName := range []string{"warehouse", "outbox"} {
		for _, allowExternal := range []bool{false, true} {
			name := connectorName + "-denied"
			if allowExternal {
				name = connectorName + "-allowed"
			}
			t.Run(name, func(t *testing.T) {
				root := t.TempDir()
				external := t.TempDir()
				redirect := filepath.Join(root, "redirect")
				if err := os.Mkdir(redirect, 0o700); err != nil {
					t.Fatalf("create initial path: %v", err)
				}
				if err := app.InitProject(root); err != nil {
					t.Fatalf("InitProject() error = %v", err)
				}
				a, err := app.Open(root)
				if err != nil {
					t.Fatalf("Open() error = %v", err)
				}
				config := map[string]string{"path": filepath.Join(redirect, "effect")}
				if allowExternal {
					config["allow_external_path"] = "true"
				}
				if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
					Name:      name,
					Connector: connectorName,
					Config:    config,
				}); err != nil {
					t.Fatalf("AddCredential() error = %v", err)
				}
				connector, runtime, err := a.ResolveConnectorCredential(ctx, connectorName, name, nil)
				if err != nil {
					t.Fatalf("ResolveConnectorCredential() error = %v", err)
				}
				if err := os.Remove(redirect); err != nil {
					t.Fatalf("remove initial path: %v", err)
				}
				if err := os.Symlink(external, redirect); err != nil {
					t.Skipf("symlinks unavailable on this platform: %v", err)
				}

				err = connector.Check(ctx, runtime)
				externalEffect := filepath.Join(external, "effect")
				if allowExternal {
					if err != nil {
						t.Fatalf("explicit external policy rejected Check(): %v", err)
					}
					if _, err := os.Stat(externalEffect); err != nil {
						t.Fatalf("explicit external policy effect missing: %v", err)
					}
					return
				}
				if err == nil {
					t.Fatal("Check() accepted a retargeted path without explicit external policy")
				}
				if _, err := os.Stat(externalEffect); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("denied Check() created external effect: %v", err)
				}
			})
		}
	}
}
