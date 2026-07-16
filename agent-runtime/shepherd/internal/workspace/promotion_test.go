package workspace

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	_ "modernc.org/sqlite"
)

func TestStageGSDStateSnapshotsCommittedWALAndInstallsStandaloneDatabase(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	source := filepath.Join(root, "candidate-gsd")
	if err := os.MkdirAll(source, 0o700); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(source, "gsd.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	for _, statement := range []string{`PRAGMA journal_mode=WAL`, `PRAGMA wal_autocheckpoint=0`, `CREATE TABLE proof(value TEXT)`, `INSERT INTO proof(value) VALUES ('committed-through-wal')`} {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := os.Stat(dbPath + "-wal"); err != nil {
		t.Fatalf("WAL fixture missing: %v", err)
	}
	repo := filepath.Join(root, "repo")
	paths, err := PlanPromotionPaths(repo, "j1")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(paths.Canonical, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(paths.Canonical, "STATE.md"), []byte("old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	validated, err := SnapshotGSDManifest(ctx, source, root)
	if err != nil {
		t.Fatal(err)
	}
	staged, err := StageGSDState(ctx, source, paths)
	if err != nil {
		t.Fatal(err)
	}
	if staged.ManifestHash != validated.Hash {
		t.Fatalf("normalized staged hash=%s want validator-bound %s", staged.ManifestHash, validated.Hash)
	}
	if _, err := os.Stat(filepath.Join(paths.Stage, "gsd.db-wal")); !os.IsNotExist(err) {
		t.Fatalf("staged WAL exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(paths.Stage, "gsd.db-shm")); !os.IsNotExist(err) {
		t.Fatalf("staged SHM exists: %v", err)
	}
	if err := InstallGSDState(ctx, paths, staged.ManifestHash, staged.BackupManifestHash, nil); err != nil {
		t.Fatal(err)
	}
	installed, err := sql.Open("sqlite", filepath.Join(paths.Canonical, "gsd.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = installed.Close() }()
	var value, integrity string
	if err := installed.QueryRowContext(ctx, `SELECT value FROM proof`).Scan(&value); err != nil || value != "committed-through-wal" {
		t.Fatalf("value=%q err=%v", value, err)
	}
	if err := installed.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&integrity); err != nil || integrity != "ok" {
		t.Fatalf("integrity=%q err=%v", integrity, err)
	}
}

func TestSnapshotGSDManifestRejectsOrphanSQLiteSidecars(t *testing.T) {
	for _, fixture := range []struct {
		name    string
		content []byte
	}{
		{name: "zero-length WAL", content: nil},
		{name: "zero-length SHM", content: nil},
		{name: "non-empty WAL", content: []byte("orphan")},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(root, "candidate")
			protected := filepath.Join(root, "protected")
			if err := os.MkdirAll(source, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(protected, 0o700); err != nil {
				t.Fatal(err)
			}
			name := "gsd.db-wal"
			if strings.Contains(fixture.name, "SHM") {
				name = "gsd.db-shm"
			}
			if err := os.WriteFile(filepath.Join(source, name), fixture.content, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := SnapshotGSDManifest(context.Background(), source, protected); err == nil ||
				!strings.Contains(err.Error(), "without gsd.db") {
				t.Fatalf("orphan sidecar err=%v", err)
			}
		})
	}
}

func TestGSDManifestDeterministicBoundedAndRejectsUnsafeFiles(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "phases", "M001"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "STATE.md"), []byte("state\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "phases", "M001", "PLAN.md"), []byte("plan\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	first, err := BuildGSDManifest(ctx, root)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildGSDManifest(ctx, root)
	if err != nil || first.Hash != second.Hash || first.JSON != second.JSON {
		t.Fatalf("manifest nondeterministic: first=%+v second=%+v err=%v", first, second, err)
	}
	if !strings.Contains(first.JSON, `"path":"STATE.md"`) || !strings.Contains(first.JSON, `"type":"file"`) {
		t.Fatalf("manifest=%s", first.JSON)
	}
	if err := os.Symlink(filepath.Join(root, "STATE.md"), filepath.Join(root, "unsafe-link")); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildGSDManifest(ctx, root); err == nil {
		t.Fatal("symlink accepted")
	}
}

func TestGSDManifestRejectsSpecialAndOversizedTrees(t *testing.T) {
	ctx := context.Background()
	t.Run("special", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, "special"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "STATE.md"), []byte("ok"), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := syscall.Mkfifo(filepath.Join(root, "special", "pipe"), 0o600); err != nil {
			t.Skipf("fifo unavailable: %v", err)
		}
		if _, err := BuildGSDManifest(ctx, root); err == nil {
			t.Fatal("special file accepted")
		}
	})
	t.Run("oversized", func(t *testing.T) {
		root := t.TempDir()
		if err := os.WriteFile(filepath.Join(root, "too-large"), make([]byte, MaxGSDManifestFileBytes+1), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := BuildGSDManifest(ctx, root); err == nil {
			t.Fatal("oversized file accepted")
		}
	})
}

func TestInstallGSDStateFailureBeforeBackupRenameRecovers(t *testing.T) {
	testPromotionSwapBoundary(t, PromotionBoundaryBeforeBackupRename)
}

func TestInstallGSDStateFailureAfterBackupRenameRecovers(t *testing.T) {
	testPromotionSwapBoundary(t, PromotionBoundaryAfterBackupRename)
}

func TestInstallGSDStateFailureAfterStageInstallRecovers(t *testing.T) {
	testPromotionSwapBoundary(t, PromotionBoundaryAfterStageInstall)
}

func TestRecoverGSDStateRepeatedlyIsIdempotent(t *testing.T) {
	ctx, paths, staged := promotionSwapFixture(t)
	if err := InstallGSDState(ctx, paths, staged.ManifestHash, staged.BackupManifestHash, nil); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := RecoverGSDState(ctx, paths, staged.ManifestHash, staged.BackupManifestHash, nil); err != nil {
			t.Fatalf("recovery %d: %v", i, err)
		}
	}
	assertManifestHash(t, ctx, paths.Canonical, staged.ManifestHash)
	assertManifestHash(t, ctx, paths.Backup, staged.BackupManifestHash)
}

func TestRecoverGSDStateBlocksMissingChangedOrCorruptOwnedResources(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*testing.T, PromotionPaths)
	}{
		{name: "missing stage", mutate: func(t *testing.T, paths PromotionPaths) { removeTreeForTest(t, paths.Stage) }},
		{name: "changed stage", mutate: func(t *testing.T, paths PromotionPaths) {
			if err := os.WriteFile(filepath.Join(paths.Stage, "STATE.md"), []byte("tampered\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "symlinked backup", mutate: func(t *testing.T, paths PromotionPaths) {
			removeTreeForTest(t, paths.Backup)
			if err := os.Symlink(t.TempDir(), paths.Backup); err != nil {
				t.Fatal(err)
			}
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, paths, staged := promotionSwapFixture(t)
			test.mutate(t, paths)
			if err := RecoverGSDState(ctx, paths, staged.ManifestHash, staged.BackupManifestHash, nil); err == nil {
				t.Fatal("corrupt promotion resources recovered")
			}
		})
	}
}

func TestResetPromotionStageRemovesOnlyJournalOwnedPartialStage(t *testing.T) {
	ctx, paths, _ := promotionSwapFixture(t)
	removeTreeForTest(t, paths.Stage)
	if err := os.MkdirAll(filepath.Join(paths.Stage, "partial"), 0o700); err != nil {
		t.Fatal(err)
	}
	unknown := paths.Stage + ".unknown"
	if err := os.MkdirAll(unknown, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := ResetPromotionStage(ctx, paths); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(paths.Stage); !os.IsNotExist(err) {
		t.Fatalf("partial stage remains: %v", err)
	}
	if _, err := os.Stat(unknown); err != nil {
		t.Fatalf("unknown sibling changed: %v", err)
	}
}

func TestPromotionCleanupResumesOwnedTombstone(t *testing.T) {
	ctx, paths, staged := promotionSwapFixture(t)
	if err := InstallGSDState(ctx, paths, staged.ManifestHash, staged.BackupManifestHash, nil); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(paths.Backup, paths.Backup+".delete"); err != nil {
		t.Fatal(err)
	}
	if err := CleanupPromotionArtifacts(ctx, paths, staged.ManifestHash, staged.BackupManifestHash); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(paths.Backup + ".delete"); !os.IsNotExist(err) {
		t.Fatalf("cleanup tombstone remains: %v", err)
	}
}

func TestPromotionCleanupLeavesUnknownDirectoriesUntouched(t *testing.T) {
	ctx, paths, staged := promotionSwapFixture(t)
	unknown := filepath.Join(filepath.Dir(paths.Canonical), ".gsd.stage-unknown")
	if err := os.MkdirAll(unknown, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := InstallGSDState(ctx, paths, staged.ManifestHash, staged.BackupManifestHash, nil); err != nil {
		t.Fatal(err)
	}
	if err := CleanupPromotionArtifacts(ctx, paths, staged.ManifestHash, staged.BackupManifestHash); err != nil {
		t.Fatal(err)
	}
	if err := CleanupPromotionArtifacts(ctx, paths, staged.ManifestHash, staged.BackupManifestHash); err != nil {
		t.Fatalf("idempotent cleanup: %v", err)
	}
	if _, err := os.Stat(unknown); err != nil {
		t.Fatalf("unknown directory changed: %v", err)
	}
}

func testPromotionSwapBoundary(t *testing.T, boundary PromotionBoundary) {
	t.Helper()
	ctx, paths, staged := promotionSwapFixture(t)
	injected := errors.New("injected promotion stop")
	failpoint := func(observed PromotionBoundary) error {
		if observed == boundary {
			return injected
		}
		return nil
	}
	if err := InstallGSDState(ctx, paths, staged.ManifestHash, staged.BackupManifestHash, failpoint); !errors.Is(err, injected) {
		t.Fatalf("boundary=%s err=%v", boundary, err)
	}
	if err := RecoverGSDState(ctx, paths, staged.ManifestHash, staged.BackupManifestHash, nil); err != nil {
		t.Fatal(err)
	}
	assertManifestHash(t, ctx, paths.Canonical, staged.ManifestHash)
	assertManifestHash(t, ctx, paths.Backup, staged.BackupManifestHash)
}

func promotionSwapFixture(t *testing.T) (context.Context, PromotionPaths, StagedGSDState) {
	t.Helper()
	ctx := context.Background()
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	canonical := filepath.Join(repo, ".gsd")
	source := filepath.Join(root, "candidate")
	for path, content := range map[string]string{filepath.Join(canonical, "STATE.md"): "old\n", filepath.Join(source, "STATE.md"): "new\n"} {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	paths, err := PlanPromotionPaths(repo, "j1")
	if err != nil {
		t.Fatal(err)
	}
	staged, err := StageGSDState(ctx, source, paths)
	if err != nil {
		t.Fatal(err)
	}
	return ctx, paths, staged
}

func assertManifestHash(t *testing.T, ctx context.Context, root, want string) {
	t.Helper()
	manifest, err := BuildGSDManifest(ctx, root)
	if err != nil || manifest.Hash != want {
		t.Fatalf("manifest root=%s got=%+v err=%v want=%s", root, manifest, err, want)
	}
}

func removeTreeForTest(t *testing.T, root string) {
	t.Helper()
	if err := os.RemoveAll(root); err != nil { // test fixture only
		t.Fatal(err)
	}
}
