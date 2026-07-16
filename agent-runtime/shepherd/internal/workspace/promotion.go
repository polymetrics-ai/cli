package workspace

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	sqlite "modernc.org/sqlite"
)

const (
	MaxGSDManifestEntries    = 4096
	MaxGSDManifestFileBytes  = 64 * 1024 * 1024
	MaxGSDManifestTotalBytes = 128 * 1024 * 1024
)

type GSDManifestEntry struct {
	Path string `json:"path"`
	Type string `json:"type"`
	Size int64  `json:"size"`
	Hash string `json:"hash,omitempty"`
}

type GSDManifest struct {
	JSON    string
	Hash    string
	Entries []GSDManifestEntry
}

type PromotionPaths struct {
	Canonical string
	Stage     string
	Backup    string
}

type StagedGSDState struct {
	ManifestJSON       string
	ManifestHash       string
	BackupManifestJSON string
	BackupManifestHash string
}

type PromotionBoundary string

const (
	PromotionBoundaryBeforeBackupRename PromotionBoundary = "before_backup_rename"
	PromotionBoundaryAfterBackupRename  PromotionBoundary = "after_backup_rename"
	PromotionBoundaryAfterStageInstall  PromotionBoundary = "after_stage_install"
)

type PromotionFailpoint func(PromotionBoundary) error

func PlanPromotionPaths(repoRoot, journalID string) (PromotionPaths, error) {
	if !filepath.IsAbs(repoRoot) || strings.TrimSpace(journalID) == "" || strings.ContainsAny(journalID, "/\\\r\n\x00") {
		return PromotionPaths{}, errors.New("absolute repository and safe journal identity are required")
	}
	repoRoot = filepath.Clean(repoRoot)
	parent := filepath.Dir(repoRoot)
	prefix := "." + safePromotionPathPart(filepath.Base(repoRoot)) + ".shepherd-gsd-" + safePromotionPathPart(journalID)
	return PromotionPaths{
		Canonical: filepath.Join(repoRoot, ".gsd"),
		Stage:     filepath.Join(parent, prefix+".stage"),
		Backup:    filepath.Join(parent, prefix+".backup"),
	}, nil
}

func BuildGSDManifest(ctx context.Context, root string) (GSDManifest, error) {
	info, err := os.Lstat(root)
	if err != nil {
		return GSDManifest{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return GSDManifest{}, errors.New("GSD state root must be a non-symlink directory")
	}
	root = filepath.Clean(root)
	entries := make([]GSDManifestEntry, 0, 64)
	var total int64
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil || !filepath.IsLocal(rel) || rel == "." || strings.ContainsAny(rel, "\r\n\x00") {
			return errors.New("unsafe GSD manifest path")
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if len(entries) >= MaxGSDManifestEntries {
			return errors.New("GSD manifest entry limit exceeded")
		}
		manifestPath := filepath.ToSlash(rel)
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			return fmt.Errorf("GSD manifest rejects symlink %q", manifestPath)
		case info.IsDir():
			entries = append(entries, GSDManifestEntry{Path: manifestPath, Type: "dir"})
			return nil
		case info.Mode().IsRegular():
			if info.Size() > MaxGSDManifestFileBytes || total > MaxGSDManifestTotalBytes-info.Size() {
				return errors.New("GSD manifest byte limit exceeded")
			}
			total += info.Size()
			hash, err := hashRegularFile(path, info.Size())
			if err != nil {
				return err
			}
			entries = append(entries, GSDManifestEntry{Path: manifestPath, Type: "file", Size: info.Size(), Hash: hash})
			return nil
		default:
			return fmt.Errorf("GSD manifest rejects special file %q", manifestPath)
		}
	})
	if err != nil {
		return GSDManifest{}, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	encoded, err := json.Marshal(entries)
	if err != nil {
		return GSDManifest{}, err
	}
	return GSDManifest{JSON: string(encoded), Hash: digestBytes(encoded), Entries: entries}, nil
}

func ResetPromotionStage(ctx context.Context, paths PromotionPaths) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validatePromotionPaths(paths); err != nil {
		return err
	}
	if err := requireAbsent(paths.Backup); err != nil {
		return fmt.Errorf("backup ownership: %w", err)
	}
	if _, err := os.Lstat(paths.Stage); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	root, err := os.OpenRoot(filepath.Dir(paths.Stage))
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()
	if err := root.RemoveAll(filepath.Base(paths.Stage)); err != nil {
		return fmt.Errorf("remove exact journal-owned partial stage: %w", err)
	}
	return syncDirectory(filepath.Dir(paths.Stage))
}

// SnapshotGSDManifest creates a bounded normalized copy of candidate GSD state in
// controller-owned storage. SQLite WAL state is folded through the same online
// backup path used by promotion, so the returned hash binds the exact installable
// representation rather than transient WAL/SHM bytes.
func SnapshotGSDManifest(ctx context.Context, source, protectedRoot string) (GSDManifest, error) {
	if !filepath.IsAbs(source) || !filepath.IsAbs(protectedRoot) || filepath.Clean(source) != source ||
		filepath.Clean(protectedRoot) != protectedRoot {
		return GSDManifest{}, errors.New("absolute clean snapshot paths are required")
	}
	info, err := os.Lstat(protectedRoot)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return GSDManifest{}, errors.New("protected snapshot root must be a real directory")
	}
	root, err := os.MkdirTemp(protectedRoot, ".validated-gsd-")
	if err != nil {
		return GSDManifest{}, err
	}
	defer func() { _ = os.RemoveAll(root) }()
	stage := filepath.Join(root, "snapshot")
	if err := copyGSDTreeToStage(ctx, source, stage); err != nil {
		return GSDManifest{}, err
	}
	manifest, err := BuildGSDManifest(ctx, stage)
	if err != nil {
		return GSDManifest{}, err
	}
	return manifest, nil
}

func StageGSDState(ctx context.Context, source string, paths PromotionPaths) (StagedGSDState, error) {
	if err := validatePromotionPaths(paths); err != nil {
		return StagedGSDState{}, err
	}
	if err := requireAbsent(paths.Stage); err != nil {
		return StagedGSDState{}, fmt.Errorf("stage ownership: %w", err)
	}
	if err := requireAbsent(paths.Backup); err != nil {
		return StagedGSDState{}, fmt.Errorf("backup ownership: %w", err)
	}
	backupManifest, err := BuildGSDManifest(ctx, paths.Canonical)
	if err != nil {
		return StagedGSDState{}, fmt.Errorf("manifest canonical GSD state: %w", err)
	}
	if err := copyGSDTreeToStage(ctx, source, paths.Stage); err != nil {
		return StagedGSDState{}, err
	}
	manifest, err := BuildGSDManifest(ctx, paths.Stage)
	if err != nil {
		return StagedGSDState{}, fmt.Errorf("manifest staged GSD state: %w", err)
	}
	if err := syncTree(paths.Stage); err != nil {
		return StagedGSDState{}, fmt.Errorf("sync staged GSD state: %w", err)
	}
	if err := syncDirectory(filepath.Dir(paths.Stage)); err != nil {
		return StagedGSDState{}, fmt.Errorf("sync stage parent: %w", err)
	}
	return StagedGSDState{ManifestJSON: manifest.JSON, ManifestHash: manifest.Hash, BackupManifestJSON: backupManifest.JSON, BackupManifestHash: backupManifest.Hash}, nil
}

func InstallGSDState(ctx context.Context, paths PromotionPaths, manifestHash, backupManifestHash string, failpoint PromotionFailpoint) error {
	if err := validatePromotionPaths(paths); err != nil {
		return err
	}
	if err := verifyManifestHash(ctx, paths.Stage, manifestHash); err != nil {
		return fmt.Errorf("verify staged GSD state: %w", err)
	}
	if err := verifyManifestHash(ctx, paths.Canonical, backupManifestHash); err != nil {
		return fmt.Errorf("verify canonical GSD state: %w", err)
	}
	if err := requireAbsent(paths.Backup); err != nil {
		return fmt.Errorf("backup ownership: %w", err)
	}
	if err := hitPromotionFailpoint(failpoint, PromotionBoundaryBeforeBackupRename); err != nil {
		return err
	}
	if err := os.Rename(paths.Canonical, paths.Backup); err != nil {
		return fmt.Errorf("rename canonical GSD state to backup: %w", err)
	}
	if err := syncRenameParents(paths.Canonical, paths.Backup); err != nil {
		return fmt.Errorf("sync backup rename: %w", err)
	}
	if err := hitPromotionFailpoint(failpoint, PromotionBoundaryAfterBackupRename); err != nil {
		return err
	}
	if err := verifyManifestHash(ctx, paths.Backup, backupManifestHash); err != nil {
		return fmt.Errorf("verify installed backup: %w", err)
	}
	if err := os.Rename(paths.Stage, paths.Canonical); err != nil {
		return fmt.Errorf("rename staged GSD state to canonical: %w", err)
	}
	if err := syncRenameParents(paths.Stage, paths.Canonical); err != nil {
		return fmt.Errorf("sync state installation: %w", err)
	}
	if err := hitPromotionFailpoint(failpoint, PromotionBoundaryAfterStageInstall); err != nil {
		return err
	}
	if err := verifyManifestHash(ctx, paths.Canonical, manifestHash); err != nil {
		return fmt.Errorf("verify installed GSD state: %w", err)
	}
	return nil
}

func RecoverGSDState(ctx context.Context, paths PromotionPaths, manifestHash, backupManifestHash string, failpoint PromotionFailpoint) error {
	if err := validatePromotionPaths(paths); err != nil {
		return err
	}
	canonicalHash, canonicalExists, err := observedManifestHash(ctx, paths.Canonical)
	if err != nil {
		return fmt.Errorf("inspect canonical GSD state: %w", err)
	}
	stageHash, stageExists, err := observedManifestHash(ctx, paths.Stage)
	if err != nil {
		return fmt.Errorf("inspect staged GSD state: %w", err)
	}
	backupHash, backupExists, err := observedManifestHash(ctx, paths.Backup)
	if err != nil {
		return fmt.Errorf("inspect backup GSD state: %w", err)
	}
	switch {
	case canonicalExists && canonicalHash == backupManifestHash && stageExists && stageHash == manifestHash && !backupExists:
		return InstallGSDState(ctx, paths, manifestHash, backupManifestHash, failpoint)
	case !canonicalExists && stageExists && stageHash == manifestHash && backupExists && backupHash == backupManifestHash:
		if err := verifyManifestHash(ctx, paths.Backup, backupManifestHash); err != nil {
			return err
		}
		if err := os.Rename(paths.Stage, paths.Canonical); err != nil {
			return fmt.Errorf("finish GSD state installation: %w", err)
		}
		if err := syncRenameParents(paths.Stage, paths.Canonical); err != nil {
			return fmt.Errorf("sync recovered GSD state: %w", err)
		}
		if err := hitPromotionFailpoint(failpoint, PromotionBoundaryAfterStageInstall); err != nil {
			return err
		}
		return verifyManifestHash(ctx, paths.Canonical, manifestHash)
	case canonicalExists && canonicalHash == manifestHash && !stageExists && backupExists && backupHash == backupManifestHash:
		return nil
	default:
		return fmt.Errorf("promotion resources do not match journal ownership: canonical=%t stage=%t backup=%t", canonicalExists, stageExists, backupExists)
	}
}

func ValidateStagedGSDState(ctx context.Context, paths PromotionPaths, manifestHash, backupManifestHash string) error {
	if err := validatePromotionPaths(paths); err != nil {
		return err
	}
	if err := verifyManifestHash(ctx, paths.Stage, manifestHash); err != nil {
		return fmt.Errorf("verify staged GSD state: %w", err)
	}
	if err := verifyManifestHash(ctx, paths.Canonical, backupManifestHash); err != nil {
		return fmt.Errorf("verify pre-promotion canonical GSD state: %w", err)
	}
	return requireAbsent(paths.Backup)
}

func ValidateInstalledGSDState(ctx context.Context, paths PromotionPaths, manifestHash, backupManifestHash string) error {
	if err := validatePromotionPaths(paths); err != nil {
		return err
	}
	if err := verifyManifestHash(ctx, paths.Canonical, manifestHash); err != nil {
		return fmt.Errorf("verify installed canonical GSD state: %w", err)
	}
	if err := verifyManifestHash(ctx, paths.Backup, backupManifestHash); err != nil {
		return fmt.Errorf("verify retained GSD backup: %w", err)
	}
	return requireAbsent(paths.Stage)
}

func CleanupPromotionArtifacts(ctx context.Context, paths PromotionPaths, manifestHash, backupManifestHash string) error {
	if err := validatePromotionPaths(paths); err != nil {
		return err
	}
	if err := verifyManifestHash(ctx, paths.Canonical, manifestHash); err != nil {
		return fmt.Errorf("refuse cleanup without installed state: %w", err)
	}
	if err := requireAbsent(paths.Stage); err != nil {
		return fmt.Errorf("refuse cleanup with unexpected stage: %w", err)
	}
	tombstone := paths.Backup + ".delete"
	backupExists, err := pathExists(paths.Backup)
	if err != nil {
		return fmt.Errorf("inspect promotion backup: %w", err)
	}
	tombstoneExists, err := pathExists(tombstone)
	if err != nil {
		return fmt.Errorf("inspect promotion cleanup tombstone: %w", err)
	}
	if backupExists && tombstoneExists {
		return errors.New("promotion backup and cleanup tombstone both exist")
	}
	if backupExists {
		if err := verifyManifestHash(ctx, paths.Backup, backupManifestHash); err != nil {
			return fmt.Errorf("refuse cleanup of unowned backup: %w", err)
		}
		if err := os.Rename(paths.Backup, tombstone); err != nil {
			return fmt.Errorf("rename verified backup for cleanup: %w", err)
		}
		if err := syncDirectory(filepath.Dir(paths.Backup)); err != nil {
			return fmt.Errorf("sync promotion cleanup rename: %w", err)
		}
		tombstoneExists = true
	}
	if !tombstoneExists {
		return nil
	}
	root, err := os.OpenRoot(filepath.Dir(tombstone))
	if err != nil {
		return err
	}
	removeErr := root.RemoveAll(filepath.Base(tombstone))
	closeErr := root.Close()
	if removeErr != nil || closeErr != nil {
		return errors.Join(removeErr, closeErr)
	}
	return syncDirectory(filepath.Dir(tombstone))
}

func copyGSDTreeToStage(ctx context.Context, source, stage string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return fmt.Errorf("inspect candidate GSD state: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return errors.New("candidate GSD state must be a non-symlink directory")
	}
	if err := os.Mkdir(stage, 0o700); err != nil {
		return fmt.Errorf("create GSD stage: %w", err)
	}
	sourceRoot, err := os.OpenRoot(source)
	if err != nil {
		return fmt.Errorf("open candidate GSD root: %w", err)
	}
	defer func() { _ = sourceRoot.Close() }()
	var sourceDB string
	var sourceDBInfo os.FileInfo
	var entries int
	var totalBytes, sqliteBytes int64
	var sqliteSidecarFound bool
	err = filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if path == source {
			return nil
		}
		rel, err := filepath.Rel(source, path)
		if err != nil || !filepath.IsLocal(rel) || strings.ContainsAny(rel, "\r\n\x00") {
			return errors.New("unsafe candidate GSD path")
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		entries++
		if entries > MaxGSDManifestEntries {
			return errors.New("candidate GSD entry limit exceeded")
		}
		manifestPath := filepath.ToSlash(rel)
		if manifestPath == "gsd.db-journal" || strings.HasSuffix(manifestPath, "/gsd.db-journal") ||
			strings.HasSuffix(manifestPath, "/gsd.db-wal") || strings.HasSuffix(manifestPath, "/gsd.db-shm") {
			return fmt.Errorf("candidate GSD state contains unexpected SQLite sidecar %q", manifestPath)
		}
		if info.Mode().IsRegular() {
			if info.Size() > MaxGSDManifestFileBytes || totalBytes > MaxGSDManifestTotalBytes-info.Size() {
				return errors.New("candidate GSD byte limit exceeded")
			}
			totalBytes += info.Size()
		}
		if manifestPath == "gsd.db-wal" || manifestPath == "gsd.db-shm" {
			sqliteSidecarFound = true
			if !info.Mode().IsRegular() {
				return fmt.Errorf("SQLite sidecar is not regular: %q", manifestPath)
			}
			sqliteBytes += info.Size()
			if sqliteBytes > MaxGSDManifestFileBytes {
				return errors.New("candidate SQLite snapshot byte limit exceeded")
			}
			return nil
		}
		if manifestPath == "gsd.db" {
			if !info.Mode().IsRegular() {
				return errors.New("gsd.db is not a regular file")
			}
			sqliteBytes += info.Size()
			if sqliteBytes > MaxGSDManifestFileBytes {
				return errors.New("candidate SQLite snapshot byte limit exceeded")
			}
			sourceDB, sourceDBInfo = path, info
			return nil
		}
		destination := filepath.Join(stage, rel)
		switch {
		case info.Mode()&os.ModeSymlink != 0:
			return fmt.Errorf("candidate GSD state contains symlink %q", manifestPath)
		case info.IsDir():
			return os.Mkdir(destination, 0o700)
		case info.Mode().IsRegular():
			if info.Size() > MaxGSDManifestFileBytes {
				return errors.New("candidate GSD file limit exceeded")
			}
			return copyManifestRegularFile(sourceRoot, rel, destination, info)
		default:
			return fmt.Errorf("candidate GSD state contains special file %q", manifestPath)
		}
	})
	if err != nil {
		return fmt.Errorf("stage candidate GSD state: %w", err)
	}
	if sourceDB == "" && sqliteSidecarFound {
		return errors.New("candidate GSD state contains SQLite sidecars without gsd.db")
	}
	if sourceDB != "" {
		if err := snapshotSQLite(ctx, sourceDB, sourceDBInfo, filepath.Join(stage, "gsd.db")); err != nil {
			return fmt.Errorf("snapshot gsd.db: %w", err)
		}
	}
	return nil
}

func snapshotSQLite(ctx context.Context, source string, expected os.FileInfo, destination string) error {
	db, err := sql.Open("sqlite", source)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	connection, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = connection.Close() }()
	type backuper interface {
		NewBackup(string) (*sqlite.Backup, error)
	}
	err = connection.Raw(func(driverConnection any) error {
		provider, ok := driverConnection.(backuper)
		if !ok {
			return errors.New("SQLite driver does not support online backup")
		}
		backup, err := provider.NewBackup(destination)
		if err != nil {
			return err
		}
		finished := false
		defer func() {
			if !finished {
				_ = backup.Finish()
			}
		}()
		for more := true; more; {
			more, err = backup.Step(-1)
			if err != nil {
				return err
			}
		}
		if err := backup.Finish(); err != nil {
			return err
		}
		finished = true
		return nil
	})
	if err != nil {
		return err
	}
	observed, err := os.Lstat(source)
	if err != nil || observed.Mode()&os.ModeSymlink != 0 || !observed.Mode().IsRegular() || !os.SameFile(expected, observed) {
		return errors.New("candidate gsd.db changed during online backup")
	}
	return verifySQLiteIntegrity(ctx, destination)
}

func verifySQLiteIntegrity(ctx context.Context, path string) error {
	db, err := sql.Open("sqlite", path+"?mode=ro")
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	var result string
	if err := db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result); err != nil {
		return err
	}
	if result != "ok" {
		return fmt.Errorf("SQLite integrity check returned %q", result)
	}
	return nil
}

func validatePromotionPaths(paths PromotionPaths) error {
	for _, path := range []string{paths.Canonical, paths.Stage, paths.Backup} {
		if !filepath.IsAbs(path) || filepath.Clean(path) != path {
			return errors.New("promotion paths must be absolute and clean")
		}
	}
	if filepath.Base(paths.Canonical) != ".gsd" || paths.Canonical == paths.Stage || paths.Canonical == paths.Backup || paths.Stage == paths.Backup {
		return errors.New("promotion paths are not distinct canonical GSD resources")
	}
	if filepath.Dir(paths.Stage) != filepath.Dir(paths.Backup) || filepath.Dir(paths.Stage) != filepath.Dir(filepath.Dir(paths.Canonical)) {
		return errors.New("promotion resources must share the repository parent filesystem")
	}
	stageBase, backupBase := filepath.Base(paths.Stage), filepath.Base(paths.Backup)
	if !strings.Contains(stageBase, ".shepherd-gsd-") || !strings.HasSuffix(stageBase, ".stage") ||
		!strings.HasSuffix(backupBase, ".backup") || strings.TrimSuffix(stageBase, ".stage") != strings.TrimSuffix(backupBase, ".backup") {
		return errors.New("promotion stage or backup path is not journal-scoped")
	}
	parentInfo, err := os.Stat(filepath.Dir(paths.Stage))
	if err != nil {
		return fmt.Errorf("inspect promotion parent filesystem: %w", err)
	}
	repoInfo, err := os.Stat(filepath.Dir(paths.Canonical))
	if err != nil {
		return fmt.Errorf("inspect canonical repository filesystem: %w", err)
	}
	parentStat, parentOK := parentInfo.Sys().(*syscall.Stat_t)
	repoStat, repoOK := repoInfo.Sys().(*syscall.Stat_t)
	if !parentOK || !repoOK || parentStat.Dev != repoStat.Dev {
		return errors.New("promotion stage, backup, and canonical state must share a filesystem")
	}
	return nil
}

func verifyManifestHash(ctx context.Context, root, want string) error {
	if !validDigest(want) {
		return errors.New("invalid expected manifest hash")
	}
	manifest, err := BuildGSDManifest(ctx, root)
	if err != nil {
		return err
	}
	if manifest.Hash != want {
		return fmt.Errorf("manifest mismatch: got %s", manifest.Hash)
	}
	return nil
}

func observedManifestHash(ctx context.Context, path string) (string, bool, error) {
	_, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	manifest, err := BuildGSDManifest(ctx, path)
	if err != nil {
		return "", true, err
	}
	return manifest.Hash, true, nil
}

func requireAbsent(path string) error {
	_, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	return errors.New("path already exists")
}

func copyManifestRegularFile(root *os.Root, source, destination string, expected os.FileInfo) error {
	input, err := root.Open(source)
	if err != nil {
		return err
	}
	defer func() { _ = input.Close() }()
	opened, err := input.Stat()
	if err != nil || !opened.Mode().IsRegular() || !os.SameFile(expected, opened) || opened.Size() != expected.Size() {
		return errors.New("candidate GSD file changed before bounded copy")
	}
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	written, copyErr := io.Copy(output, io.LimitReader(input, MaxGSDManifestFileBytes+1))
	if copyErr == nil && written != expected.Size() {
		copyErr = errors.New("source changed while staging")
	}
	if syncErr := output.Sync(); copyErr == nil {
		copyErr = syncErr
	}
	if closeErr := output.Close(); copyErr == nil {
		copyErr = closeErr
	}
	return copyErr
}

func hashRegularFile(path string, expectedSize int64) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	hash := sha256.New()
	written, err := io.Copy(hash, io.LimitReader(file, MaxGSDManifestFileBytes+1))
	if err != nil {
		return "", err
	}
	if written != expectedSize {
		return "", errors.New("file changed while hashing")
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func syncTree(root string) error {
	var directories []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return errors.New("refuse to sync symlink")
		}
		if info.IsDir() {
			directories = append(directories, path)
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		err = file.Sync()
		closeErr := file.Close()
		if err != nil {
			return err
		}
		return closeErr
	})
	if err != nil {
		return err
	}
	for i := len(directories) - 1; i >= 0; i-- {
		if err := syncDirectory(directories[i]); err != nil {
			return err
		}
	}
	return nil
}

func syncRenameParents(source, destination string) error {
	sourceParent, destinationParent := filepath.Dir(source), filepath.Dir(destination)
	if err := syncDirectory(sourceParent); err != nil {
		return err
	}
	if destinationParent != sourceParent {
		return syncDirectory(destinationParent)
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = directory.Close() }()
	return directory.Sync()
}

func pathExists(path string) (bool, error) {
	_, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return err == nil, err
}

func hitPromotionFailpoint(failpoint PromotionFailpoint, boundary PromotionBoundary) error {
	if failpoint == nil {
		return nil
	}
	return failpoint(boundary)
}

func digestBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func safePromotionPathPart(value string) string {
	var builder strings.Builder
	for _, char := range value {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '-' || char == '_' {
			builder.WriteRune(char)
		}
	}
	if builder.Len() == 0 {
		return "unknown"
	}
	return builder.String()
}
