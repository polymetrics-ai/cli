package main

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"syscall"
	"time"

	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestindex"
)

const (
	vNextPublicationGenerationDirectory = "generations"
	vNextPublicationCurrentFile         = "CURRENT"
	vNextPublicationJournalFile         = "JOURNAL"
	vNextPublicationLockFile            = ".connectorgen.lock"
	vNextPublicationIntegrityFile       = "integrity.json"
	vNextPublicationControlMaxBytes     = 1 << 20
	vNextPublicationLeaseFile           = ".lease"
	vNextPublicationStageOwnerFile      = ".connectorgen-stage.json"
)

type vNextPublicationFaultPoint string

const (
	vNextPublicationBeforeFileSync        vNextPublicationFaultPoint = "before_file_sync"
	vNextPublicationAfterFileSync         vNextPublicationFaultPoint = "after_file_sync"
	vNextPublicationBeforeStageDirectory  vNextPublicationFaultPoint = "before_stage_directory_sync"
	vNextPublicationAfterStageDirectory   vNextPublicationFaultPoint = "after_stage_directory_sync"
	vNextPublicationBeforeJournalSync     vNextPublicationFaultPoint = "before_journal_sync"
	vNextPublicationAfterJournalSync      vNextPublicationFaultPoint = "after_journal_sync"
	vNextPublicationBeforeCurrentTempSync vNextPublicationFaultPoint = "before_current_temp_sync"
	vNextPublicationAfterCurrentTempSync  vNextPublicationFaultPoint = "after_current_temp_sync"
	vNextPublicationBeforeCurrentRename   vNextPublicationFaultPoint = "before_current_rename"
	vNextPublicationAfterCurrentRename    vNextPublicationFaultPoint = "after_current_rename"
	vNextPublicationBeforeCurrentParent   vNextPublicationFaultPoint = "before_current_parent_sync"
	vNextPublicationAfterCurrentParent    vNextPublicationFaultPoint = "after_current_parent_sync"
	vNextPublicationBeforeActiveValidate  vNextPublicationFaultPoint = "before_active_validation"
	vNextPublicationAfterActiveValidate   vNextPublicationFaultPoint = "after_active_validation"
	vNextPublicationBeforeCommitSync      vNextPublicationFaultPoint = "before_commit_sync"
	vNextPublicationAfterCommitSync       vNextPublicationFaultPoint = "after_commit_sync"
	vNextPublicationBeforePrune           vNextPublicationFaultPoint = "before_prune"
	vNextPublicationAfterPrune            vNextPublicationFaultPoint = "after_prune"
)

// vNextPublicationHooks is test-only fault instrumentation for durable state
// boundaries. Production leaves At nil.
type vNextPublicationHooks struct {
	At func(vNextPublicationFaultPoint) error
}

// vNextPublicationArtifacts is one already-admitted, immutable generation.
// Files excludes publication control files; the publisher writes those itself.
type vNextPublicationArtifacts struct {
	Files    map[string][]byte
	Validate func(root string) error
}

type vNextGenerationPointer struct {
	Generation      string `json:"generation"`
	IntegrityDigest string `json:"integrity_digest"`
}

type vNextGenerationJournal struct {
	Old   *vNextGenerationPointer `json:"old,omitempty"`
	New   vNextGenerationPointer  `json:"new"`
	State string                  `json:"state"`
}

type vNextPublicationStageOwner struct {
	Version    int    `json:"version"`
	Connector  string `json:"connector"`
	Generation string `json:"generation"`
	Stage      string `json:"stage"`
}

type vNextGenerationIntegrity struct {
	Version    int                         `json:"version"`
	Connector  string                      `json:"connector"`
	Generation string                      `json:"generation"`
	Files      []vNextGenerationFileDigest `json:"files"`
}

type vNextGenerationFileDigest struct {
	Path   string `json:"path"`
	Digest string `json:"digest"`
	Bytes  int    `json:"bytes"`
}

type vNextGenerationPublisher struct {
	root      string
	connector string
	hooks     vNextPublicationHooks
}

type vNextGenerationHandle struct {
	filesRoot *os.Root
	pointer   vNextGenerationPointer
	files     map[string]struct{}
	lease     *os.File
	released  bool
}

func newVNextGenerationPublisher(root, connector string, hooks vNextPublicationHooks) (*vNextGenerationPublisher, error) {
	if !namePattern.MatchString(connector) {
		return nil, fmt.Errorf("invalid publication connector %q", connector)
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve publication root: %w", err)
	}
	return &vNextGenerationPublisher{root: absoluteRoot, connector: connector, hooks: hooks}, nil
}

func (p *vNextGenerationPublisher) connectorRoot() string {
	return filepath.Join(p.root, p.connector)
}

func (p *vNextGenerationPublisher) generationsRoot() string {
	return filepath.Join(p.connectorRoot(), vNextPublicationGenerationDirectory)
}

func (p *vNextGenerationPublisher) generationPath(generation string) string {
	return filepath.Join(p.generationsRoot(), generation)
}

func (p *vNextGenerationPublisher) openConnectorRoot(create bool) (*os.Root, error) {
	defsRoot, err := os.OpenRoot(p.root)
	if err != nil {
		return nil, fmt.Errorf("open publication definitions root: %w", err)
	}
	defer defsRoot.Close()

	info, err := defsRoot.Lstat(p.connector)
	if errors.Is(err, fs.ErrNotExist) && create {
		if err := defsRoot.Mkdir(p.connector, 0o755); err != nil && !errors.Is(err, fs.ErrExist) {
			return nil, fmt.Errorf("create connector publication root: %w", err)
		}
		info, err = defsRoot.Lstat(p.connector)
	}
	if err != nil {
		return nil, fmt.Errorf("stat connector publication root: %w", err)
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		return nil, fmt.Errorf("connector publication root %q is a symlink", p.connector)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("connector publication root %q is not a directory", p.connector)
	}
	root, err := defsRoot.OpenRoot(p.connector)
	if err != nil {
		return nil, fmt.Errorf("open connector publication root: %w", err)
	}
	return root, nil
}

func (p *vNextGenerationPublisher) readSourceLock() ([]byte, error) {
	root, err := p.openConnectorRoot(false)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	info, err := root.Lstat("source.lock.json")
	if err != nil {
		return nil, fmt.Errorf("stat source lock: %w", err)
	}
	if info.Mode()&fs.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("source lock is not a regular file")
	}
	raw, err := root.ReadFile("source.lock.json")
	if err != nil {
		return nil, fmt.Errorf("read source lock: %w", err)
	}
	return raw, nil
}

func (p *vNextGenerationPublisher) ensureGenerationsRootLocked() error {
	root, err := p.openConnectorRoot(true)
	if err != nil {
		return err
	}
	defer root.Close()
	info, err := root.Lstat(vNextPublicationGenerationDirectory)
	if errors.Is(err, fs.ErrNotExist) {
		if err := root.Mkdir(vNextPublicationGenerationDirectory, 0o755); err != nil && !errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("create generation root: %w", err)
		}
		info, err = root.Lstat(vNextPublicationGenerationDirectory)
	}
	if err != nil {
		return fmt.Errorf("stat generation root: %w", err)
	}
	if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		return fmt.Errorf("generation root is not a directory")
	}
	return nil
}

func (p *vNextGenerationPublisher) assertGenerationsRootLocked() error {
	root, err := p.openConnectorRoot(false)
	if err != nil {
		return err
	}
	defer root.Close()
	info, err := root.Lstat(vNextPublicationGenerationDirectory)
	if err != nil {
		return fmt.Errorf("stat generation root: %w", err)
	}
	if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		return fmt.Errorf("generation root is not a directory")
	}
	return nil
}

func (p *vNextGenerationPublisher) PublishContext(ctx context.Context, artifacts vNextPublicationArtifacts) (vNextGenerationPointer, error) {
	if err := ctx.Err(); err != nil {
		return vNextGenerationPointer{}, err
	}
	files, err := vNextPublicationFiles(artifacts.Files)
	if err != nil {
		return vNextGenerationPointer{}, err
	}
	lock, err := p.lock(ctx, syscall.LOCK_EX)
	if err != nil {
		return vNextGenerationPointer{}, err
	}
	defer unlockVNextPublicationFile(lock)
	if err := p.recoverLocked(); err != nil {
		return vNextGenerationPointer{}, err
	}

	old, hasOld, err := p.readCurrentLocked()
	if err != nil {
		return vNextGenerationPointer{}, err
	}
	pointer, err := p.stageLocked(files, artifacts.Validate)
	if err != nil {
		return vNextGenerationPointer{}, err
	}
	if hasOld && old == pointer {
		return pointer, nil
	}
	if err := p.writeJournalLocked(vNextGenerationJournal{Old: pointerOrNil(old, hasOld), New: pointer, State: "prepared"}); err != nil {
		return vNextGenerationPointer{}, err
	}
	if err := p.writeCurrentLocked(pointer); err != nil {
		return vNextGenerationPointer{}, err
	}
	if err := p.hit(vNextPublicationBeforeActiveValidate); err != nil {
		return vNextGenerationPointer{}, err
	}
	_, validationErr := p.validatePointerLocked(pointer, files, artifacts.Validate)
	if err := p.hit(vNextPublicationAfterActiveValidate); err != nil {
		return vNextGenerationPointer{}, err
	}
	if validationErr != nil {
		if rollbackErr := p.rollbackLocked(old, hasOld, pointer); rollbackErr != nil {
			return vNextGenerationPointer{}, fmt.Errorf("validate active generation: %v; restore previous generation: %w", validationErr, rollbackErr)
		}
		return vNextGenerationPointer{}, fmt.Errorf("validate active generation: %w", validationErr)
	}
	if err := p.writeJournalLocked(vNextGenerationJournal{Old: pointerOrNil(old, hasOld), New: pointer, State: "committed"}); err != nil {
		return vNextGenerationPointer{}, err
	}
	if err := p.pruneLocked(pointer.Generation); err != nil {
		return vNextGenerationPointer{}, err
	}
	if err := p.removeJournalLocked(); err != nil {
		return vNextGenerationPointer{}, err
	}
	return pointer, nil
}

func (p *vNextGenerationPublisher) CheckContext(ctx context.Context, artifacts vNextPublicationArtifacts) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	files, err := vNextPublicationFiles(artifacts.Files)
	if err != nil {
		return err
	}
	lock, err := p.readLock(ctx, syscall.LOCK_SH)
	if err != nil {
		return err
	}
	defer unlockVNextPublicationFile(lock)
	if err := p.assertNoPendingJournalLocked(); err != nil {
		return err
	}
	if err := p.assertGenerationsRootLocked(); err != nil {
		return err
	}
	pointer, found, err := p.readCurrentLocked()
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("connector %q has no active generation", p.connector)
	}
	if _, err := p.validatePointerLocked(pointer, files, artifacts.Validate); err != nil {
		return err
	}
	return p.assertNoOrphansLocked(pointer.Generation)
}

func (p *vNextGenerationPublisher) Recover(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	lock, err := p.lock(ctx, syscall.LOCK_EX)
	if err != nil {
		return err
	}
	defer unlockVNextPublicationFile(lock)
	return p.recoverLocked()
}

func (p *vNextGenerationPublisher) Open(ctx context.Context) (*vNextGenerationHandle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := p.Recover(ctx); err != nil {
		return nil, err
	}
	lock, err := p.lock(ctx, syscall.LOCK_SH)
	if err != nil {
		return nil, err
	}
	defer unlockVNextPublicationFile(lock)
	if err := p.assertGenerationsRootLocked(); err != nil {
		return nil, err
	}
	pointer, found, err := p.readCurrentLocked()
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("connector %q has no active generation", p.connector)
	}
	integrity, err := p.validatePointerLocked(pointer, nil, nil)
	if err != nil {
		return nil, err
	}
	generationRoot, err := os.OpenRoot(p.generationPath(pointer.Generation))
	if err != nil {
		return nil, fmt.Errorf("open generation root %q: %w", pointer.Generation, err)
	}
	lease, err := generationRoot.OpenFile(vNextPublicationLeaseFile, os.O_RDWR, 0)
	if err != nil {
		_ = generationRoot.Close()
		return nil, fmt.Errorf("open generation lease %q: %w", pointer.Generation, err)
	}
	if err := syscall.Flock(int(lease.Fd()), syscall.LOCK_SH); err != nil {
		_ = lease.Close()
		_ = generationRoot.Close()
		return nil, fmt.Errorf("hold generation lease %q: %w", pointer.Generation, err)
	}
	files := make(map[string]struct{}, len(integrity.Files))
	for _, file := range integrity.Files {
		files[file.Path] = struct{}{}
	}
	return &vNextGenerationHandle{filesRoot: generationRoot, pointer: pointer, files: files, lease: lease}, nil
}

func (p *vNextGenerationPublisher) Prune(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	lock, err := p.lock(ctx, syscall.LOCK_EX)
	if err != nil {
		return err
	}
	defer unlockVNextPublicationFile(lock)
	if err := p.recoverLocked(); err != nil {
		return err
	}
	pointer, found, err := p.readCurrentLocked()
	if err != nil {
		return err
	}
	if !found {
		return p.pruneLocked("")
	}
	if _, err := p.validatePointerLocked(pointer, nil, nil); err != nil {
		return fmt.Errorf("validate active generation before pruning: %w", err)
	}
	return p.pruneLocked(pointer.Generation)
}

func (p *vNextGenerationPublisher) GenerationExists(generation string) bool {
	if !vNextPublicationGenerationIDValid(generation) {
		return false
	}
	root, err := p.openConnectorRoot(false)
	if err != nil {
		return false
	}
	defer root.Close()
	generations, err := root.Lstat(vNextPublicationGenerationDirectory)
	if err != nil || !generations.IsDir() || generations.Mode()&fs.ModeSymlink != 0 {
		return false
	}
	info, err := root.Lstat(path.Join(vNextPublicationGenerationDirectory, generation))
	return err == nil && info.IsDir() && info.Mode()&fs.ModeSymlink == 0
}

func (h *vNextGenerationHandle) Generation() string {
	if h == nil {
		return ""
	}
	return h.pointer.Generation
}

func (h *vNextGenerationHandle) Files() []string {
	if h == nil {
		return nil
	}
	files := make([]string, 0, len(h.files))
	for name := range h.files {
		files = append(files, name)
	}
	sort.Strings(files)
	return files
}

func (h *vNextGenerationHandle) ReadFile(name string) ([]byte, error) {
	if h == nil || h.released {
		return nil, fs.ErrClosed
	}
	if _, exists := h.files[name]; !exists {
		return nil, fs.ErrNotExist
	}
	return h.filesRoot.ReadFile(name)
}

func (h *vNextGenerationHandle) Release() {
	if h == nil || h.released {
		return
	}
	h.released = true
	unlockVNextPublicationFile(h.lease)
	_ = h.filesRoot.Close()
}

func (p *vNextGenerationPublisher) stageLocked(files map[string][]byte, validate func(string) error) (vNextGenerationPointer, error) {
	generation := vNextPublicationGenerationID(files)
	if err := p.ensureGenerationsRootLocked(); err != nil {
		return vNextGenerationPointer{}, err
	}
	final := p.generationPath(generation)
	if _, err := os.Stat(final); err == nil {
		pointer, err := p.pointerForGenerationLocked(generation)
		if err != nil {
			return vNextGenerationPointer{}, err
		}
		if _, err := p.validatePointerLocked(pointer, files, validate); err != nil {
			return vNextGenerationPointer{}, fmt.Errorf("existing generation %q is not the requested closed set: %w", generation, err)
		}
		return pointer, nil
	} else if !errors.Is(err, fs.ErrNotExist) {
		return vNextGenerationPointer{}, fmt.Errorf("stat generation %q: %w", generation, err)
	}

	stage, err := os.MkdirTemp(p.generationsRoot(), ".stage-")
	if err != nil {
		return vNextGenerationPointer{}, fmt.Errorf("create same-filesystem stage: %w", err)
	}
	stageName := filepath.Base(stage)
	marker, err := vNextPublicationJSON(vNextPublicationStageOwner{
		Version:    1,
		Connector:  p.connector,
		Generation: generation,
		Stage:      stageName,
	})
	if err != nil {
		return vNextGenerationPointer{}, fmt.Errorf("encode stage ownership marker: %w", err)
	}
	if err := p.writeStageFile(stage, vNextPublicationStageOwnerFile, marker); err != nil {
		return vNextGenerationPointer{}, err
	}
	if err := syncVNextPublicationDirectory(stage); err != nil {
		return vNextGenerationPointer{}, fmt.Errorf("sync stage ownership marker: %w", err)
	}
	if err := syncVNextPublicationDirectory(p.generationsRoot()); err != nil {
		return vNextGenerationPointer{}, fmt.Errorf("sync stage ownership parent: %w", err)
	}
	published := false
	defer func() {
		if !published {
			_ = p.removeStageLocked(stageName)
		}
	}()
	for _, name := range sortedVNextPublicationFiles(files) {
		if err := p.writeStageFile(stage, name, files[name]); err != nil {
			return vNextGenerationPointer{}, err
		}
	}
	if err := p.writeStageFile(stage, vNextPublicationLeaseFile, nil); err != nil {
		return vNextGenerationPointer{}, err
	}
	integrity, payload, pointer, err := vNextPublicationIntegrityForFiles(p.connector, generation, files)
	if err != nil {
		return vNextGenerationPointer{}, fmt.Errorf("encode generation integrity: %w", err)
	}
	if err := p.writeStageFile(stage, vNextPublicationIntegrityFile, payload); err != nil {
		return vNextGenerationPointer{}, err
	}
	if err := p.syncTreeDirectories(stage); err != nil {
		return vNextGenerationPointer{}, err
	}
	if err := p.hit(vNextPublicationBeforeStageDirectory); err != nil {
		return vNextGenerationPointer{}, err
	}
	if err := syncVNextPublicationDirectory(stage); err != nil {
		return vNextGenerationPointer{}, fmt.Errorf("sync staged generation: %w", err)
	}
	if err := p.hit(vNextPublicationAfterStageDirectory); err != nil {
		return vNextGenerationPointer{}, err
	}
	if _, err := p.validateGenerationLocked(stage, pointer); err != nil {
		return vNextGenerationPointer{}, err
	}
	if validate != nil {
		if err := validate(stage); err != nil {
			return vNextGenerationPointer{}, fmt.Errorf("validate staged generation: %w", err)
		}
	}
	if integrity.Generation != generation {
		return vNextGenerationPointer{}, fmt.Errorf("staged integrity generation %q does not match %q", integrity.Generation, generation)
	}
	if err := os.Rename(stage, final); err != nil {
		return vNextGenerationPointer{}, fmt.Errorf("activate staged generation %q: %w", generation, err)
	}
	published = true
	if err := syncVNextPublicationDirectory(p.generationsRoot()); err != nil {
		return vNextGenerationPointer{}, fmt.Errorf("sync generation root: %w", err)
	}
	return pointer, nil
}

func (p *vNextGenerationPublisher) writeStageFile(stage, name string, payload []byte) error {
	path := filepath.Join(stage, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create staged parent for %s: %w", name, err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("create staged file %s: %w", name, err)
	}
	if _, err := file.Write(payload); err != nil {
		_ = file.Close()
		return fmt.Errorf("write staged file %s: %w", name, err)
	}
	if err := p.hit(vNextPublicationBeforeFileSync); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync staged file %s: %w", name, err)
	}
	if err := p.hit(vNextPublicationAfterFileSync); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close staged file %s: %w", name, err)
	}
	return nil
}

func (p *vNextGenerationPublisher) syncTreeDirectories(root string) error {
	var directories []string
	err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			directories = append(directories, name)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk staged directories: %w", err)
	}
	sort.Slice(directories, func(left, right int) bool { return len(directories[left]) > len(directories[right]) })
	for _, directory := range directories {
		if err := syncVNextPublicationDirectory(directory); err != nil {
			return fmt.Errorf("sync staged directory %s: %w", directory, err)
		}
	}
	return nil
}

func (p *vNextGenerationPublisher) writeJournalLocked(journal vNextGenerationJournal) error {
	if journal.State != "prepared" && journal.State != "committed" {
		return fmt.Errorf("invalid generation journal state %q", journal.State)
	}
	payload, err := json.Marshal(journal)
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	before, after := vNextPublicationBeforeJournalSync, vNextPublicationAfterJournalSync
	if journal.State == "committed" {
		before, after = vNextPublicationBeforeCommitSync, vNextPublicationAfterCommitSync
	}
	return p.writeAtomicLocked(vNextPublicationJournalFile, payload, before, after, "", "", "", "")
}

func (p *vNextGenerationPublisher) writeCurrentLocked(pointer vNextGenerationPointer) error {
	if err := vNextPublicationPointerValid(pointer); err != nil {
		return err
	}
	payload, err := json.Marshal(pointer)
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return p.writeAtomicLocked(vNextPublicationCurrentFile, payload, vNextPublicationBeforeCurrentTempSync, vNextPublicationAfterCurrentTempSync, vNextPublicationBeforeCurrentRename, vNextPublicationAfterCurrentRename, vNextPublicationBeforeCurrentParent, vNextPublicationAfterCurrentParent)
}

func (p *vNextGenerationPublisher) writeAtomicLocked(target string, payload []byte, beforeSync, afterSync, beforeRename, afterRename, beforeParent, afterParent vNextPublicationFaultPoint) error {
	root, err := p.openConnectorRoot(true)
	if err != nil {
		return err
	}
	defer root.Close()
	temporaryName, temporary, err := vNextPublicationCreateTemp(root)
	if err != nil {
		return err
	}
	defer func() { _ = root.Remove(temporaryName) }()
	if err := temporary.Chmod(0o644); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(payload); err != nil {
		_ = temporary.Close()
		return err
	}
	if beforeSync != "" {
		if err := p.hit(beforeSync); err != nil {
			_ = temporary.Close()
			return err
		}
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if afterSync != "" {
		if err := p.hit(afterSync); err != nil {
			_ = temporary.Close()
			return err
		}
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if beforeRename != "" {
		if err := p.hit(beforeRename); err != nil {
			return err
		}
	}
	if err := root.Rename(temporaryName, target); err != nil {
		return err
	}
	if afterRename != "" {
		if err := p.hit(afterRename); err != nil {
			return err
		}
	}
	if beforeParent != "" {
		if err := p.hit(beforeParent); err != nil {
			return err
		}
	}
	if err := syncVNextPublicationRoot(root); err != nil {
		return err
	}
	if afterParent != "" {
		if err := p.hit(afterParent); err != nil {
			return err
		}
	}
	return nil
}

func vNextPublicationCreateTemp(root *os.Root) (string, *os.File, error) {
	var token [16]byte
	for range 128 {
		if _, err := cryptorand.Read(token[:]); err != nil {
			return "", nil, fmt.Errorf("generate publication temporary name: %w", err)
		}
		name := ".connectorgen-publication-" + hex.EncodeToString(token[:])
		file, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			return name, file, nil
		}
		if !errors.Is(err, fs.ErrExist) {
			return "", nil, fmt.Errorf("create publication temporary file: %w", err)
		}
	}
	return "", nil, fmt.Errorf("create publication temporary file: exhausted unique names")
}

func syncVNextPublicationRoot(root *os.Root) error {
	directory, err := root.Open(".")
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func (p *vNextGenerationPublisher) readCurrentLocked() (vNextGenerationPointer, bool, error) {
	root, err := p.openConnectorRoot(false)
	if err != nil {
		return vNextGenerationPointer{}, false, err
	}
	defer root.Close()
	payload, found, err := vNextPublicationReadControl(root, vNextPublicationCurrentFile, "CURRENT")
	if err != nil {
		return vNextGenerationPointer{}, false, err
	}
	if !found {
		return vNextGenerationPointer{}, false, nil
	}
	var pointer vNextGenerationPointer
	if err := vNextPublicationDecode(payload, &pointer); err != nil {
		return vNextGenerationPointer{}, false, fmt.Errorf("decode CURRENT: %w", err)
	}
	if err := vNextPublicationPointerValid(pointer); err != nil {
		return vNextGenerationPointer{}, false, fmt.Errorf("invalid CURRENT: %w", err)
	}
	return pointer, true, nil
}

func (p *vNextGenerationPublisher) pointerForGenerationLocked(generation string) (vNextGenerationPointer, error) {
	if !vNextPublicationGenerationIDValid(generation) {
		return vNextGenerationPointer{}, fmt.Errorf("invalid generation %q", generation)
	}
	root := p.generationPath(generation)
	info, err := os.Lstat(root)
	if err != nil {
		return vNextGenerationPointer{}, fmt.Errorf("stat generation %q: %w", generation, err)
	}
	if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		return vNextGenerationPointer{}, fmt.Errorf("generation %q is not a directory", generation)
	}
	generationRoot, err := os.OpenRoot(root)
	if err != nil {
		return vNextGenerationPointer{}, fmt.Errorf("open generation %q root: %w", generation, err)
	}
	defer generationRoot.Close()
	payload, found, err := vNextPublicationReadControl(generationRoot, vNextPublicationIntegrityFile, "generation integrity")
	if err != nil {
		return vNextGenerationPointer{}, fmt.Errorf("read generation %q integrity: %w", generation, err)
	}
	if !found {
		return vNextGenerationPointer{}, fmt.Errorf("generation %q has no integrity file", generation)
	}
	var integrity vNextGenerationIntegrity
	if err := vNextPublicationDecode(payload, &integrity); err != nil {
		return vNextGenerationPointer{}, fmt.Errorf("decode generation %q integrity: %w", generation, err)
	}
	return vNextGenerationPointer{Generation: generation, IntegrityDigest: vNextPublicationDigest(payload)}, nil
}

func vNextPublicationReadControl(root *os.Root, name, label string) ([]byte, bool, error) {
	info, err := root.Lstat(name)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("stat %s: %w", label, err)
	}
	if info.Mode()&fs.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, true, fmt.Errorf("%s is not a regular file", label)
	}
	if info.Size() < 0 || info.Size() > vNextPublicationControlMaxBytes {
		return nil, true, fmt.Errorf("%s exceeds %d-byte limit", label, vNextPublicationControlMaxBytes)
	}
	handle, err := root.Open(name)
	if err != nil {
		return nil, true, fmt.Errorf("open %s: %w", label, err)
	}
	defer handle.Close()
	payload, err := io.ReadAll(io.LimitReader(handle, vNextPublicationControlMaxBytes+1))
	if err != nil {
		return nil, true, fmt.Errorf("read %s: %w", label, err)
	}
	if len(payload) > vNextPublicationControlMaxBytes {
		return nil, true, fmt.Errorf("%s exceeds %d-byte limit", label, vNextPublicationControlMaxBytes)
	}
	return payload, true, nil
}

func (p *vNextGenerationPublisher) validatePointerLocked(pointer vNextGenerationPointer, expected map[string][]byte, validate func(string) error) (vNextGenerationIntegrity, error) {
	if err := vNextPublicationPointerValid(pointer); err != nil {
		return vNextGenerationIntegrity{}, err
	}
	root := p.generationPath(pointer.Generation)
	integrity, err := p.validateGenerationLocked(root, pointer)
	if err != nil {
		return vNextGenerationIntegrity{}, err
	}
	if expected != nil {
		if err := vNextPublicationCompareFiles(integrity.Files, expected); err != nil {
			return vNextGenerationIntegrity{}, err
		}
	}
	if validate != nil {
		if err := validate(root); err != nil {
			return vNextGenerationIntegrity{}, fmt.Errorf("validate active generation: %w", err)
		}
	}
	return integrity, nil
}

func (p *vNextGenerationPublisher) validateGenerationLocked(root string, pointer vNextGenerationPointer) (vNextGenerationIntegrity, error) {
	rootInfo, err := os.Lstat(root)
	if err != nil {
		return vNextGenerationIntegrity{}, fmt.Errorf("stat generation root: %w", err)
	}
	if !rootInfo.IsDir() || rootInfo.Mode()&fs.ModeSymlink != 0 {
		return vNextGenerationIntegrity{}, fmt.Errorf("generation root is not a directory")
	}
	generationRoot, err := os.OpenRoot(root)
	if err != nil {
		return vNextGenerationIntegrity{}, fmt.Errorf("open generation root: %w", err)
	}
	defer generationRoot.Close()
	payload, found, err := vNextPublicationReadControl(generationRoot, vNextPublicationIntegrityFile, "integrity")
	if err != nil {
		return vNextGenerationIntegrity{}, err
	}
	if !found {
		return vNextGenerationIntegrity{}, fmt.Errorf("integrity file is missing")
	}
	if vNextPublicationDigest(payload) != pointer.IntegrityDigest {
		return vNextGenerationIntegrity{}, fmt.Errorf("integrity digest does not match CURRENT")
	}
	var integrity vNextGenerationIntegrity
	if err := vNextPublicationDecode(payload, &integrity); err != nil {
		return vNextGenerationIntegrity{}, fmt.Errorf("decode integrity: %w", err)
	}
	if integrity.Version != 1 || integrity.Connector != p.connector || integrity.Generation != pointer.Generation {
		return vNextGenerationIntegrity{}, fmt.Errorf("integrity identity does not match active connector generation")
	}
	if !vNextPublicationGenerationIDValid(integrity.Generation) {
		return vNextGenerationIntegrity{}, fmt.Errorf("integrity generation is invalid")
	}
	if err := p.validateStageOwnerLocked(root, "", pointer.Generation); err != nil {
		return vNextGenerationIntegrity{}, err
	}
	computedGeneration, err := vNextPublicationValidateFileDigests(root, integrity.Files)
	if err != nil {
		return vNextGenerationIntegrity{}, err
	}
	if computedGeneration != pointer.Generation {
		return vNextGenerationIntegrity{}, fmt.Errorf("generation content address %q does not match selected generation %q", computedGeneration, pointer.Generation)
	}
	if err := vNextPublicationValidateClosedTree(root, integrity.Files); err != nil {
		return vNextGenerationIntegrity{}, err
	}
	return integrity, nil
}

func (p *vNextGenerationPublisher) recoverLocked() error {
	if err := p.ensureGenerationsRootLocked(); err != nil {
		return err
	}
	if err := p.removeStagesLocked(); err != nil {
		return err
	}
	root, err := p.openConnectorRoot(false)
	if err != nil {
		return err
	}
	defer root.Close()
	payload, foundJournal, err := vNextPublicationReadControl(root, vNextPublicationJournalFile, "publication journal")
	if err != nil {
		return err
	}
	if !foundJournal {
		current, found, readErr := p.readCurrentLocked()
		if readErr != nil {
			return readErr
		}
		active := ""
		if found {
			if _, validationErr := p.validatePointerLocked(current, nil, nil); validationErr != nil {
				return fmt.Errorf("recover active generation: %w", validationErr)
			}
			active = current.Generation
		}
		return p.pruneRecoveredLocked(active)
	}
	var journal vNextGenerationJournal
	if err := vNextPublicationDecode(payload, &journal); err != nil {
		return fmt.Errorf("decode publication journal: %w", err)
	}
	if err := vNextPublicationJournalValid(journal); err != nil {
		return err
	}
	current, hasCurrent, err := p.readCurrentLocked()
	if err != nil {
		return err
	}
	if hasCurrent && current != journal.New && (journal.Old == nil || current != *journal.Old) {
		return fmt.Errorf("recover journal: CURRENT does not match either journal generation")
	}
	if hasCurrent && current == journal.New {
		if _, err := p.validatePointerLocked(journal.New, nil, nil); err == nil {
			if err := p.pruneRecoveredLocked(journal.New.Generation); err != nil {
				return err
			}
			return p.removeJournalLocked()
		}
	}
	if journal.Old != nil {
		if _, err := p.validatePointerLocked(*journal.Old, nil, nil); err != nil {
			return fmt.Errorf("recover journal: old generation is invalid: %w", err)
		}
		if !hasCurrent || current != *journal.Old {
			if err := p.writeCurrentLocked(*journal.Old); err != nil {
				return fmt.Errorf("restore old CURRENT: %w", err)
			}
		}
	} else if hasCurrent && current == journal.New {
		if err := p.removeCurrentLocked(); err != nil {
			return err
		}
	}
	if err := p.removeGenerationLocked(journal.New.Generation); err != nil {
		return err
	}
	if err := syncVNextPublicationDirectory(p.generationsRoot()); err != nil {
		return fmt.Errorf("sync recovered generation removal: %w", err)
	}
	return p.removeJournalLocked()
}

func (p *vNextGenerationPublisher) rollbackLocked(old vNextGenerationPointer, hasOld bool, rejected vNextGenerationPointer) error {
	if hasOld {
		if err := p.writeCurrentLocked(old); err != nil {
			return err
		}
	} else if err := p.removeCurrentLocked(); err != nil {
		return err
	}
	if err := p.removeGenerationLocked(rejected.Generation); err != nil {
		return err
	}
	if err := syncVNextPublicationDirectory(p.generationsRoot()); err != nil {
		return fmt.Errorf("sync rejected generation removal: %w", err)
	}
	return p.removeJournalLocked()
}

func (p *vNextGenerationPublisher) removeCurrentLocked() error {
	return p.removeControlLocked(vNextPublicationCurrentFile)
}

func (p *vNextGenerationPublisher) removeJournalLocked() error {
	return p.removeControlLocked(vNextPublicationJournalFile)
}

func (p *vNextGenerationPublisher) removeControlLocked(name string) error {
	root, err := p.openConnectorRoot(false)
	if err != nil {
		return err
	}
	defer root.Close()
	if err := root.Remove(name); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return syncVNextPublicationRoot(root)
}

func (p *vNextGenerationPublisher) pruneLocked(active string) error {
	return p.pruneWithFaultsLocked(active, true)
}

func (p *vNextGenerationPublisher) pruneRecoveredLocked(active string) error {
	return p.pruneWithFaultsLocked(active, false)
}

func (p *vNextGenerationPublisher) pruneWithFaultsLocked(active string, injectFaults bool) error {
	if injectFaults {
		if err := p.hit(vNextPublicationBeforePrune); err != nil {
			return err
		}
	}
	entries, err := os.ReadDir(p.generationsRoot())
	if err != nil {
		return fmt.Errorf("read generation root: %w", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".stage-") {
			if err := p.removeStageLocked(name); err != nil {
				return err
			}
			continue
		}
		if name == active {
			continue
		}
		if !vNextPublicationGenerationIDValid(name) {
			return fmt.Errorf("unexpected generation member %q", name)
		}
		if err := p.removeGenerationLocked(name); err != nil {
			return err
		}
	}
	if err := syncVNextPublicationDirectory(p.generationsRoot()); err != nil {
		return err
	}
	if injectFaults {
		return p.hit(vNextPublicationAfterPrune)
	}
	return nil
}

func (p *vNextGenerationPublisher) removeStagesLocked() error {
	entries, err := os.ReadDir(p.generationsRoot())
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".stage-") {
			if err := p.removeStageLocked(entry.Name()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *vNextGenerationPublisher) removeStageLocked(name string) error {
	if !vNextPublicationStageNameValid(name) {
		return fmt.Errorf("invalid staging directory %q", name)
	}
	stage := filepath.Join(p.generationsRoot(), name)
	info, err := os.Lstat(stage)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat stale staging directory %q: %w", name, err)
	}
	if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		return fmt.Errorf("stale staging entry %q is not a directory", name)
	}
	if err := p.validateStageOwnerLocked(stage, name, ""); err != nil {
		return fmt.Errorf("refuse remove stale staging directory %q without ownership proof: %w", name, err)
	}
	if err := os.RemoveAll(stage); err != nil {
		return fmt.Errorf("remove stale staging directory %q: %w", name, err)
	}
	return nil
}

func (p *vNextGenerationPublisher) validateStageOwnerLocked(root, stage, generation string) error {
	stageRoot, err := os.OpenRoot(root)
	if err != nil {
		return fmt.Errorf("open stage ownership root: %w", err)
	}
	defer stageRoot.Close()
	payload, found, err := vNextPublicationReadControl(stageRoot, vNextPublicationStageOwnerFile, "stage ownership marker")
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("stage ownership marker is missing")
	}
	var owner vNextPublicationStageOwner
	if err := vNextPublicationDecode(payload, &owner); err != nil {
		return fmt.Errorf("decode stage ownership marker: %w", err)
	}
	if owner.Version != 1 || owner.Connector != p.connector || !vNextPublicationGenerationIDValid(owner.Generation) || !vNextPublicationStageNameValid(owner.Stage) {
		return fmt.Errorf("stage ownership marker is invalid")
	}
	if stage != "" && owner.Stage != stage {
		return fmt.Errorf("stage ownership marker does not match staging directory")
	}
	if generation != "" && owner.Generation != generation {
		return fmt.Errorf("stage ownership marker does not match generation")
	}
	return nil
}

func vNextPublicationStageNameValid(name string) bool {
	if !strings.HasPrefix(name, ".stage-") || len(name) == len(".stage-") || strings.ContainsAny(name, `/\`) {
		return false
	}
	for _, value := range name {
		if value < 0x20 || value == 0x7f {
			return false
		}
	}
	return true
}

func (p *vNextGenerationPublisher) removeGenerationLocked(generation string) error {
	if !vNextPublicationGenerationIDValid(generation) {
		return fmt.Errorf("invalid generation %q", generation)
	}
	root := p.generationPath(generation)
	info, err := os.Lstat(root)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat generation %q: %w", generation, err)
	}
	if !info.IsDir() || info.Mode()&fs.ModeSymlink != 0 {
		return fmt.Errorf("generation %q is not a directory", generation)
	}
	pointer, err := p.pointerForGenerationLocked(generation)
	if err != nil {
		return fmt.Errorf("refuse prune generation %q without validated publication ownership: %w", generation, err)
	}
	if _, err := p.validatePointerLocked(pointer, nil, nil); err != nil {
		return fmt.Errorf("refuse prune generation %q without validated publication ownership: %w", generation, err)
	}
	leasePath := filepath.Join(root, vNextPublicationLeaseFile)
	leaseInfo, err := os.Lstat(leasePath)
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("generation %q has no lease file", generation)
	}
	if err != nil {
		return fmt.Errorf("stat generation lease %q: %w", generation, err)
	}
	if !leaseInfo.Mode().IsRegular() {
		return fmt.Errorf("generation %q lease is not regular", generation)
	}
	lease, err := os.OpenFile(leasePath, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open generation lease %q: %w", generation, err)
	}
	defer lease.Close()
	if err := syscall.Flock(int(lease.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil
		}
		return fmt.Errorf("lock generation lease %q: %w", generation, err)
	}
	defer func() { _ = syscall.Flock(int(lease.Fd()), syscall.LOCK_UN) }()
	if err := os.RemoveAll(root); err != nil {
		return fmt.Errorf("prune generation %q: %w", generation, err)
	}
	return nil
}

func (p *vNextGenerationPublisher) assertNoOrphansLocked(active string) error {
	entries, err := os.ReadDir(p.generationsRoot())
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("generation root contains symlink %q", entry.Name())
		}
		if entry.Name() != active {
			if strings.HasPrefix(entry.Name(), ".stage-") {
				return fmt.Errorf("stale staging directory %q remains", entry.Name())
			}
			if !vNextPublicationGenerationIDValid(entry.Name()) {
				return fmt.Errorf("unexpected generation member %q", entry.Name())
			}
		}
	}
	return nil
}

func (p *vNextGenerationPublisher) assertNoPendingJournalLocked() error {
	root, err := p.openConnectorRoot(false)
	if err != nil {
		return err
	}
	defer root.Close()
	_, found, err := vNextPublicationReadControl(root, vNextPublicationJournalFile, "publication journal")
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	return fmt.Errorf("connector %q has a pending publication journal; recover before checking", p.connector)
}

func (p *vNextGenerationPublisher) lock(ctx context.Context, mode int) (*os.File, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	root, err := p.openConnectorRoot(true)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	lock, err := vNextPublicationOpenLock(root, true)
	if err != nil {
		return nil, err
	}
	if err := vNextPublicationAcquireLock(ctx, lock, mode, "connector publication"); err != nil {
		_ = lock.Close()
		return nil, err
	}
	return lock, nil
}

func (p *vNextGenerationPublisher) readLock(ctx context.Context, mode int) (*os.File, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	root, err := p.openConnectorRoot(false)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	lock, err := vNextPublicationOpenLock(root, false)
	if err != nil {
		return nil, err
	}
	if err := vNextPublicationAcquireLock(ctx, lock, mode, "existing connector publication"); err != nil {
		_ = lock.Close()
		return nil, err
	}
	return lock, nil
}

func vNextPublicationAcquireLock(ctx context.Context, lock *os.File, mode int, label string) error {
	retry := time.NewTimer(time.Hour)
	if !retry.Stop() {
		<-retry.C
	}
	defer retry.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := syscall.Flock(int(lock.Fd()), mode|syscall.LOCK_NB)
		if err == nil {
			return nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			return fmt.Errorf("lock %s: %w", label, err)
		}
		retry.Reset(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-retry.C:
		}
	}
}

func vNextPublicationOpenLock(root *os.Root, create bool) (*os.File, error) {
	info, err := root.Lstat(vNextPublicationLockFile)
	if errors.Is(err, fs.ErrNotExist) {
		if !create {
			return nil, fmt.Errorf("open existing connector publication lock: %w", err)
		}
		lock, createErr := root.OpenFile(vNextPublicationLockFile, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
		if createErr == nil {
			return lock, nil
		}
		if !errors.Is(createErr, fs.ErrExist) {
			return nil, fmt.Errorf("create connector publication lock: %w", createErr)
		}
		info, err = root.Lstat(vNextPublicationLockFile)
	}
	if err != nil {
		return nil, fmt.Errorf("stat connector publication lock: %w", err)
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		return nil, fmt.Errorf("connector publication lock is a symlink")
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("connector publication lock is not a regular file")
	}
	lock, err := root.OpenFile(vNextPublicationLockFile, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open connector publication lock: %w", err)
	}
	return lock, nil
}

func (p *vNextGenerationPublisher) hit(point vNextPublicationFaultPoint) error {
	if p.hooks.At == nil {
		return nil
	}
	if err := p.hooks.At(point); err != nil {
		return fmt.Errorf("publication fault at %s: %w", point, err)
	}
	return nil
}

func vNextPublicationFiles(files map[string][]byte) (map[string][]byte, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("publication requires at least one artifact")
	}
	out := make(map[string][]byte, len(files))
	for name, payload := range files {
		if err := vNextPublicationArtifactPathValid(name); err != nil {
			return nil, err
		}
		out[name] = append([]byte(nil), payload...)
	}
	return out, nil
}

func vNextPublicationArtifactPathValid(name string) error {
	if name == vNextPublicationIntegrityFile || name == vNextPublicationLeaseFile {
		return fmt.Errorf("publication artifact %q is reserved", name)
	}
	if name == "" || strings.Contains(name, `\`) || path.IsAbs(name) || path.Clean(name) != name || name == "." || strings.HasPrefix(name, "../") || strings.Contains(name, "/../") {
		return fmt.Errorf("publication artifact path %q is invalid", name)
	}
	for _, component := range strings.Split(name, "/") {
		if component == "" || strings.HasPrefix(component, ".") {
			return fmt.Errorf("publication artifact path %q is invalid", name)
		}
	}
	for _, value := range name {
		if value < 0x20 || value == 0x7f {
			return fmt.Errorf("publication artifact path %q contains a control character", name)
		}
	}
	return nil
}

func vNextPublicationGenerationID(files map[string][]byte) string {
	hash := sha256.New()
	for _, name := range sortedVNextPublicationFiles(files) {
		vNextPublicationWriteLength(hash, len(name))
		_, _ = io.WriteString(hash, name)
		payload := files[name]
		vNextPublicationWriteLength(hash, len(payload))
		_, _ = hash.Write(payload)
	}
	return "g-" + hex.EncodeToString(hash.Sum(nil))
}

func vNextPublicationIntegrityForFiles(connector, generation string, files map[string][]byte) (vNextGenerationIntegrity, []byte, vNextGenerationPointer, error) {
	integrity := vNextGenerationIntegrity{Version: 1, Connector: connector, Generation: generation, Files: make([]vNextGenerationFileDigest, 0, len(files))}
	for _, name := range sortedVNextPublicationFiles(files) {
		integrity.Files = append(integrity.Files, vNextGenerationFileDigest{Path: name, Digest: vNextPublicationDigest(files[name]), Bytes: len(files[name])})
	}
	payload, err := json.Marshal(integrity)
	if err != nil {
		return vNextGenerationIntegrity{}, nil, vNextGenerationPointer{}, err
	}
	payload = append(payload, '\n')
	return integrity, payload, vNextGenerationPointer{Generation: generation, IntegrityDigest: vNextPublicationDigest(payload)}, nil
}

func vNextPublicationDigest(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func vNextPublicationPointerValid(pointer vNextGenerationPointer) error {
	if !vNextPublicationGenerationIDValid(pointer.Generation) {
		return fmt.Errorf("generation %q is invalid", pointer.Generation)
	}
	if !vNextPublicationDigestValid(pointer.IntegrityDigest) {
		return fmt.Errorf("integrity digest is invalid")
	}
	return nil
}

func vNextPublicationGenerationIDValid(value string) bool {
	if len(value) != len("g-")+64 || !strings.HasPrefix(value, "g-") {
		return false
	}
	for _, rune := range value[len("g-"):] {
		if !(rune >= '0' && rune <= '9') && !(rune >= 'a' && rune <= 'f') {
			return false
		}
	}
	return true
}

func vNextPublicationDigestValid(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	for _, rune := range value[len("sha256:"):] {
		if !(rune >= '0' && rune <= '9') && !(rune >= 'a' && rune <= 'f') {
			return false
		}
	}
	return true
}

func vNextPublicationValidateFileDigests(root string, files []vNextGenerationFileDigest) (string, error) {
	generationRoot, err := os.OpenRoot(root)
	if err != nil {
		return "", fmt.Errorf("open generation artifact root: %w", err)
	}
	defer generationRoot.Close()
	content := sha256.New()
	previous := ""
	for _, file := range files {
		if err := vNextPublicationArtifactPathValid(file.Path); err != nil {
			return "", err
		}
		if previous >= file.Path {
			return "", fmt.Errorf("integrity files are not strictly sorted")
		}
		previous = file.Path
		if file.Bytes < 0 || !vNextPublicationDigestValid(file.Digest) {
			return "", fmt.Errorf("integrity entry %q is invalid", file.Path)
		}
		info, err := generationRoot.Lstat(file.Path)
		if err != nil {
			return "", fmt.Errorf("stat generation artifact %s: %w", file.Path, err)
		}
		if !info.Mode().IsRegular() || info.Size() != int64(file.Bytes) {
			return "", fmt.Errorf("generation artifact %s is not the declared regular file", file.Path)
		}
		vNextPublicationWriteLength(content, len(file.Path))
		_, _ = io.WriteString(content, file.Path)
		vNextPublicationWriteLength(content, file.Bytes)

		handle, err := generationRoot.Open(file.Path)
		if err != nil {
			return "", fmt.Errorf("open generation artifact %s: %w", file.Path, err)
		}
		digest := sha256.New()
		var buffer [32 * 1024]byte
		var bytesRead int64
		for {
			count, readErr := handle.Read(buffer[:])
			if count > 0 {
				bytesRead += int64(count)
				_, _ = content.Write(buffer[:count])
				_, _ = digest.Write(buffer[:count])
			}
			if errors.Is(readErr, io.EOF) {
				break
			}
			if readErr != nil {
				_ = handle.Close()
				return "", fmt.Errorf("read generation artifact %s: %w", file.Path, readErr)
			}
		}
		if err := handle.Close(); err != nil {
			return "", fmt.Errorf("close generation artifact %s: %w", file.Path, err)
		}
		actualDigest := "sha256:" + hex.EncodeToString(digest.Sum(nil))
		if bytesRead != int64(file.Bytes) || actualDigest != file.Digest {
			return "", fmt.Errorf("generation artifact %s differs from integrity", file.Path)
		}
	}
	return "g-" + hex.EncodeToString(content.Sum(nil)), nil
}

func vNextPublicationWriteLength(writer io.Writer, length int) {
	var frame [8]byte
	binary.BigEndian.PutUint64(frame[:], uint64(length))
	_, _ = writer.Write(frame[:])
}

func vNextPublicationValidateClosedTree(root string, files []vNextGenerationFileDigest) error {
	expectedFiles := make(map[string]struct{}, len(files)+3)
	expectedFiles[vNextPublicationIntegrityFile] = struct{}{}
	expectedFiles[vNextPublicationLeaseFile] = struct{}{}
	expectedFiles[vNextPublicationStageOwnerFile] = struct{}{}
	expectedDirectories := map[string]struct{}{".": {}}
	for _, file := range files {
		expectedFiles[file.Path] = struct{}{}
		parts := strings.Split(file.Path, "/")
		for index := 1; index < len(parts); index++ {
			expectedDirectories[strings.Join(parts[:index], "/")] = struct{}{}
		}
	}
	seen := make(map[string]struct{}, len(expectedFiles))
	err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, name)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("generation contains symlink %s", name)
		}
		if entry.IsDir() {
			if _, exists := expectedDirectories[relative]; !exists {
				return fmt.Errorf("generation contains unexpected directory %q", relative)
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("generation contains nonregular member %s", name)
		}
		if _, exists := expectedFiles[relative]; !exists {
			return fmt.Errorf("generation contains unexpected member %q", relative)
		}
		if relative == vNextPublicationLeaseFile {
			info, err := entry.Info()
			if err != nil {
				return fmt.Errorf("stat generation lease: %w", err)
			}
			if info.Size() != 0 {
				return fmt.Errorf("generation lease must be empty")
			}
		}
		seen[relative] = struct{}{}
		return nil
	})
	if err != nil {
		return err
	}
	for name := range expectedFiles {
		if _, found := seen[name]; !found {
			return fmt.Errorf("generation is missing expected member %q", name)
		}
	}
	return nil
}

func vNextPublicationCompareFiles(actual []vNextGenerationFileDigest, expected map[string][]byte) error {
	if len(actual) != len(expected) {
		return fmt.Errorf("active generation has %d artifacts, expected %d", len(actual), len(expected))
	}
	for _, entry := range actual {
		payload, found := expected[entry.Path]
		if !found {
			return fmt.Errorf("active generation has unexpected artifact %q", entry.Path)
		}
		if len(payload) != entry.Bytes || vNextPublicationDigest(payload) != entry.Digest {
			return fmt.Errorf("active generation artifact %q differs from expected publication", entry.Path)
		}

	}
	return nil
}

func vNextPublicationJournalValid(journal vNextGenerationJournal) error {
	if journal.State != "prepared" && journal.State != "committed" {
		return fmt.Errorf("invalid publication journal state %q", journal.State)
	}
	if err := vNextPublicationPointerValid(journal.New); err != nil {
		return fmt.Errorf("invalid new journal pointer: %w", err)
	}
	if journal.Old != nil {
		if err := vNextPublicationPointerValid(*journal.Old); err != nil {
			return fmt.Errorf("invalid old journal pointer: %w", err)
		}
		if *journal.Old == journal.New {
			return fmt.Errorf("journal old and new generation are equal")
		}
	}
	return nil
}

func vNextPublicationDecode(payload []byte, destination any) error {
	return decodeStrictJSON(payload, destination)
}

func sortedVNextPublicationFiles(files map[string][]byte) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func pointerOrNil(pointer vNextGenerationPointer, found bool) *vNextGenerationPointer {
	if !found {
		return nil
	}
	copy := pointer
	return &copy
}

func syncVNextPublicationDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func unlockVNextPublicationFile(file *os.File) {
	if file == nil {
		return
	}
	_ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	_ = file.Close()
}

type vNextPublishedGenerationManifest struct {
	Version   int                 `json:"version"`
	Execution manifestindex.Entry `json:"execution"`
}

type vNextPublishedGenerationIndex struct {
	Version int                   `json:"version"`
	Entries []manifestindex.Entry `json:"entries"`
}

type vNextPublishedAtlasReference struct {
	ID        string `json:"id"`
	Digest    string `json:"digest"`
	Available bool   `json:"available"`
	Reference string `json:"reference"`
}

type vNextPublishedGenerationProof struct {
	Version          int    `json:"version"`
	SourceLockDigest string `json:"source_lock_digest"`
	ExecutionDigest  string `json:"execution_digest"`
	ProvenanceDigest string `json:"provenance_digest"`
}

// vNextPublicationArtifactsForStage turns the one CP10-admitted in-memory
// generation into the complete physical publication set. It neither reads a
// second source lock nor writes a connector root.
func vNextPublicationArtifactsForStage(raw []byte, connector string, stage vNextStagedGeneration) (vNextPublicationArtifacts, error) {
	if stage.Identity.Connector != connector || stage.Manifest.Connector != connector {
		return vNextPublicationArtifacts{}, fmt.Errorf("staged generation connector does not match publication connector %q", connector)
	}
	files, err := vNextPublicationFiles(stage.Outputs)
	if err != nil {
		return vNextPublicationArtifacts{}, err
	}
	provenance, err := vNextPublicationJSON(stage.Provenance)
	if err != nil {
		return vNextPublicationArtifacts{}, fmt.Errorf("encode staged provenance: %w", err)
	}
	atlas, err := vNextPublicationJSON(vNextPublishedAtlasReferences(stage.Sync))
	if err != nil {
		return vNextPublicationArtifacts{}, fmt.Errorf("encode staged Atlas references: %w", err)
	}
	manifest, err := vNextPublicationJSON(vNextPublishedGenerationManifest{Version: 1, Execution: stage.Manifest})
	if err != nil {
		return vNextPublicationArtifacts{}, fmt.Errorf("encode staged manifest: %w", err)
	}
	index, err := vNextPublicationJSON(vNextPublishedGenerationIndex{Version: 1, Entries: stage.Index.List()})
	if err != nil {
		return vNextPublicationArtifacts{}, fmt.Errorf("encode staged compact index: %w", err)
	}
	proof, err := vNextPublicationJSON(vNextPublishedGenerationProof{
		Version:          1,
		SourceLockDigest: vNextPublicationDigest(raw),
		ExecutionDigest:  stage.Identity.Digest,
		ProvenanceDigest: vNextPublicationDigest(provenance),
	})
	if err != nil {
		return vNextPublicationArtifacts{}, fmt.Errorf("encode staged proof metadata: %w", err)
	}
	for name, payload := range map[string][]byte{
		"manifest.json":   manifest,
		"provenance.json": provenance,
		"atlas.json":      atlas,
		"index.json":      index,
		"proof.json":      proof,
	} {
		if _, exists := files[name]; exists {
			return vNextPublicationArtifacts{}, fmt.Errorf("staged execution set reserves publication artifact %q", name)
		}
		files[name] = payload
	}
	return vNextPublicationArtifacts{
		Files: files,
		Validate: func(root string) error {
			return vNextValidatePublishedStage(root, stage)
		},
	}, nil
}

func vNextPublishedAtlasReferences(admissions []vNextResolvedSyncAdmission) []vNextPublishedAtlasReference {
	references := make([]vNextPublishedAtlasReference, 0, len(admissions))
	for _, admission := range admissions {
		if admission.Result.Plan == nil {
			continue
		}
		foundation := admission.Result.Plan.Foundation
		references = append(references, vNextPublishedAtlasReference{
			ID: foundation.ID, Digest: foundation.Digest, Available: foundation.Available, Reference: foundation.Reference,
		})
	}
	sort.Slice(references, func(left, right int) bool {
		if references[left].ID != references[right].ID {
			return references[left].ID < references[right].ID
		}
		if references[left].Digest != references[right].Digest {
			return references[left].Digest < references[right].Digest
		}
		if references[left].Reference != references[right].Reference {
			return references[left].Reference < references[right].Reference
		}
		return !references[left].Available && references[right].Available
	})
	out := references[:0]
	for _, reference := range references {
		if len(out) > 0 && out[len(out)-1] == reference {
			continue
		}
		out = append(out, reference)
	}
	return out
}

func vNextPublicationJSON(value any) ([]byte, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

// vNextPublicationStageFS gives the existing loader its required
// connector-named directory while every opened byte still comes from the
// physical staged generation directory.
type vNextPublicationStageFS struct {
	connector string
	root      fs.FS
}

func (f vNextPublicationStageFS) Open(name string) (fs.File, error) {
	switch {
	case name == "." || name == f.connector:
		return f.root.Open(".")
	case strings.HasPrefix(name, f.connector+"/"):
		return f.root.Open(strings.TrimPrefix(name, f.connector+"/"))
	default:
		return nil, fs.ErrNotExist
	}
}

func vNextValidatePublishedStage(root string, expected vNextStagedGeneration) error {
	bundle, err := engine.Load(vNextPublicationStageFS{connector: expected.Manifest.Connector, root: os.DirFS(root)}, expected.Manifest.Connector)
	if err != nil {
		return fmt.Errorf("load staged execution bundle: %w", err)
	}
	if bundle.Identity != expected.Identity {
		return fmt.Errorf("staged execution identity does not match admitted identity")
	}
	selection, err := vNextSelectedRuntime(expected.Manifest.Connector)
	if err != nil {
		return fmt.Errorf("select staged runtime: %w", err)
	}
	runtime := engine.New(bundle, selection.Hooks)
	actual := vNextManifestEntry(bundle, runtime, selection)
	if !reflect.DeepEqual(actual, expected.Manifest) {
		return fmt.Errorf("staged manifest does not match admitted selection")
	}
	index, err := manifestindex.New([]manifestindex.Entry{actual}, 1)
	if err != nil {
		return fmt.Errorf("admit staged compact index: %w", err)
	}
	if !reflect.DeepEqual(index.List(), expected.Index.List()) {
		return fmt.Errorf("staged compact index does not match admitted index")
	}
	if surface := runtime.CommandSurface(); surface != nil {
		for _, command := range surface.Commands {
			if command.Availability != "implemented" {
				continue
			}
			if err := commandrunner.Preflight(runtime, strings.Fields(command.Path)); err != nil {
				return fmt.Errorf("preflight staged command %q: %w", command.Path, err)
			}
		}
	}
	return nil
}
