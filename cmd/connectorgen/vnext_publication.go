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

	"golang.org/x/sys/unix"

	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestindex"
)

const (
	vNextPublicationGenerationDirectory = "generations"
	vNextPublicationCurrentFile         = "CURRENT"
	vNextPublicationJournalFile         = "JOURNAL"
	vNextPublicationIntegrityFile       = "integrity.json"
	vNextPublicationControlMaxBytes     = 1 << 20
	vNextPublicationLeaseFile           = ".lease"
	vNextPublicationStageOwnerFile      = ".connectorgen-stage.json"
)

type vNextPublicationFaultPoint string

const (
	vNextPublicationBeforeFileSync                         vNextPublicationFaultPoint = "before_file_sync"
	vNextPublicationAfterFileSync                          vNextPublicationFaultPoint = "after_file_sync"
	vNextPublicationBeforeStageDirectory                   vNextPublicationFaultPoint = "before_stage_directory_sync"
	vNextPublicationAfterStageDirectory                    vNextPublicationFaultPoint = "after_stage_directory_sync"
	vNextPublicationBeforeJournalSync                      vNextPublicationFaultPoint = "before_journal_sync"
	vNextPublicationAfterJournalSync                       vNextPublicationFaultPoint = "after_journal_sync"
	vNextPublicationBeforeCurrentTempSync                  vNextPublicationFaultPoint = "before_current_temp_sync"
	vNextPublicationAfterCurrentTempSync                   vNextPublicationFaultPoint = "after_current_temp_sync"
	vNextPublicationBeforeCurrentRename                    vNextPublicationFaultPoint = "before_current_rename"
	vNextPublicationAfterCurrentRename                     vNextPublicationFaultPoint = "after_current_rename"
	vNextPublicationBeforeCurrentParent                    vNextPublicationFaultPoint = "before_current_parent_sync"
	vNextPublicationAfterCurrentParent                     vNextPublicationFaultPoint = "after_current_parent_sync"
	vNextPublicationBeforeActiveValidate                   vNextPublicationFaultPoint = "before_active_validation"
	vNextPublicationAfterActiveValidate                    vNextPublicationFaultPoint = "after_active_validation"
	vNextPublicationAfterOpenValidation                    vNextPublicationFaultPoint = "after_open_validation"
	vNextPublicationBeforeCommitSync                       vNextPublicationFaultPoint = "before_commit_sync"
	vNextPublicationAfterCommitSync                        vNextPublicationFaultPoint = "after_commit_sync"
	vNextPublicationBeforePrune                            vNextPublicationFaultPoint = "before_prune"
	vNextPublicationAfterPrune                             vNextPublicationFaultPoint = "after_prune"
	vNextPublicationBeforeStageCleanup                     vNextPublicationFaultPoint = "before_stage_cleanup"
	vNextPublicationBeforeLockAcquire                      vNextPublicationFaultPoint = "before_lock_acquire"
	vNextPublicationAfterLockAcquire                       vNextPublicationFaultPoint = "after_lock_acquire"
	vNextPublicationBeforeStageRename                      vNextPublicationFaultPoint = "before_stage_rename"
	vNextPublicationAfterStageRename                       vNextPublicationFaultPoint = "after_stage_rename"
	vNextPublicationBeforeStageRemoval                     vNextPublicationFaultPoint = "before_stage_removal"
	vNextPublicationBeforeGenerationRemoval                vNextPublicationFaultPoint = "before_generation_removal"
	vNextPublicationBeforeControlRemoval                   vNextPublicationFaultPoint = "before_control_removal"
	vNextPublicationAfterControlSourceIdentity             vNextPublicationFaultPoint = "after_control_source_identity"
	vNextPublicationAfterFinalControlSourceIdentity        vNextPublicationFaultPoint = "after_final_control_source_identity"
	vNextPublicationAfterControlRepairPrepared             vNextPublicationFaultPoint = "after_control_repair_prepared"
	vNextPublicationAfterControlRepairRecord               vNextPublicationFaultPoint = "after_control_repair_record"
	vNextPublicationAfterBaseControlRepairPrepared         vNextPublicationFaultPoint = "after_base_control_repair_prepared"
	vNextPublicationAfterControlRepairCaptureDirectory     vNextPublicationFaultPoint = "after_control_repair_capture_directory"
	vNextPublicationBeforeControlRepairCapture             vNextPublicationFaultPoint = "before_control_repair_capture"
	vNextPublicationAfterControlRepairCaptureRename        vNextPublicationFaultPoint = "after_control_repair_capture_rename"
	vNextPublicationAfterControlRepairCaptureDirectorySync vNextPublicationFaultPoint = "after_control_repair_capture_directory_sync"
	vNextPublicationAfterControlRepairCaptureRootSync      vNextPublicationFaultPoint = "after_control_repair_capture_root_sync"
	vNextPublicationAfterControlRepairCaptured             vNextPublicationFaultPoint = "after_control_repair_captured"
	vNextPublicationAfterControlRepairInstall              vNextPublicationFaultPoint = "after_control_repair_install"
	vNextPublicationAfterControlRepairInstallSync          vNextPublicationFaultPoint = "after_control_repair_install_sync"
	vNextPublicationAfterControlRepairSelected             vNextPublicationFaultPoint = "after_control_repair_selected"
	vNextPublicationAfterFinalControlRepairValidation      vNextPublicationFaultPoint = "after_final_control_repair_validation"
	vNextPublicationAfterControlAuthorityReadScan          vNextPublicationFaultPoint = "after_control_authority_read_scan"
	vNextPublicationAfterStageRemovalIdentity              vNextPublicationFaultPoint = "after_stage_removal_identity"
	vNextPublicationAfterGenerationRemovalIdentity         vNextPublicationFaultPoint = "after_generation_removal_identity"
	vNextPublicationAfterGenerationLeaseIdentity           vNextPublicationFaultPoint = "after_generation_lease_identity"
	vNextPublicationAfterControlRemovalIdentity            vNextPublicationFaultPoint = "after_control_removal_identity"
	vNextPublicationBeforeQuarantineRestore                vNextPublicationFaultPoint = "before_quarantine_restore"
)

// vNextPublicationHooks is test-only fault instrumentation for durable state
// boundaries. Production leaves At nil.
type vNextPublicationHooks struct {
	At                   func(vNextPublicationFaultPoint) error
	ControlCaptureRename func() error
	// LockContention is test-only instrumentation for the exact retained
	// directory descriptor after its nonblocking flock reports contention.
	// Production leaves it nil.
	LockContention func(vNextPublicationIdentity) error
	// AfterTemporaryOpen and AfterQuarantineOpen are nil-by-default test seams
	// at the owned allocation boundaries. Production leaves them nil.
	AfterTemporaryOpen  func(*vNextPublicationDirectory, string, *vNextPublicationDirectory) error
	AfterQuarantineOpen func(*vNextPublicationDirectory, string, *vNextPublicationDirectory, vNextPublicationIdentity) error
	// CloseAtomicTemporary is a narrowly scoped test seam. Its callback owns
	// exactly one real Close of the supplied temporary and may return a
	// completion error after that Close; production leaves it nil.
	CloseAtomicTemporary func(*os.File) error
	// CloseRepairPredecessor is the matching narrow test seam for an already
	// linked predecessor anchor during failed repair creation.
	CloseRepairPredecessor func(*vNextPublicationDirectory) error
	// CloseDirectory and CloseStageFile are nil-by-default test seams for
	// compound completion paths. Their callbacks must perform the real Close
	// before returning an injected completion error.
	CloseDirectory func(*os.File, string) error
	CloseStageFile func(*os.File) error
}
type vNextPublicationArtifacts struct {
	Files    map[string][]byte
	Validate func(root fs.FS) error
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

type vNextPublicationOperation struct {
	connector    *vNextPublicationDirectory
	generations  *vNextPublicationDirectory
	lock         *os.File
	lockIdentity vNextPublicationIdentity
	controls     map[string]vNextPublicationIdentity
}

func (operation *vNextPublicationOperation) bindControl(name string, identity vNextPublicationIdentity) {
	if operation.controls == nil {
		operation.controls = make(map[string]vNextPublicationIdentity)
	}
	operation.controls[name] = identity
}

func (operation *vNextPublicationOperation) clearControl(name string) {
	delete(operation.controls, name)
}

func (operation *vNextPublicationOperation) assertLockBound() error {
	if operation == nil || operation.connector == nil || operation.lock == nil {
		return fs.ErrClosed
	}
	return vNextPublicationAssertLockBound(operation.connector, operation.lock, operation.lockIdentity)
}

type vNextGenerationHandle struct {
	filesRoot *vNextPublicationDirectory
	pointer   vNextGenerationPointer
	files     map[string]struct{}
	hold      *os.File
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

func (p *vNextGenerationPublisher) openConnectorRoot(create bool) (*vNextPublicationDirectory, error) {
	definitions, err := vNextPublicationOpenDirectoryWithCloseForTest(p.root, "publication definitions root", p.hooks.CloseDirectory)
	if err != nil {
		return nil, err
	}
	label := fmt.Sprintf("connector publication root %q", p.connector)
	var connector *vNextPublicationDirectory
	if create {
		connector, err = definitions.ensureDirectory(p.connector, label)
	} else {
		connector, err = definitions.openDirectory(p.connector, label)
	}
	closeErr := definitions.Close()
	if err != nil {
		return nil, err
	}
	if closeErr != nil {
		if connectorCloseErr := connector.Close(); connectorCloseErr != nil {
			return nil, fmt.Errorf("close publication definitions root: %v; close connector root: %w", closeErr, connectorCloseErr)
		}
		return nil, fmt.Errorf("close publication definitions root: %w", closeErr)
	}
	return connector, nil
}

func (p *vNextGenerationPublisher) openOperationRoot(ctx context.Context, create bool) (*vNextPublicationOperation, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	connector, err := p.openConnectorRoot(create)
	if err != nil {
		return nil, err
	}
	return &vNextPublicationOperation{connector: connector}, nil
}

func (p *vNextGenerationPublisher) acquireOperation(ctx context.Context, operation *vNextPublicationOperation, mode int) error {
	if operation == nil || operation.connector == nil {
		return fs.ErrClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := p.hit(vNextPublicationBeforeLockAcquire); err != nil {
		return err
	}
	lock, identity, err := vNextPublicationOpenLock(operation.connector)
	if err != nil {
		return err
	}
	if err := vNextPublicationAcquireLock(ctx, lock, mode, "connector publication", func() error {
		if err := p.hit(vNextPublicationAfterLockAcquire); err != nil {
			return err
		}
		return vNextPublicationAssertLockBound(operation.connector, lock, identity)
	}, func() error {
		if p.hooks.LockContention == nil {
			return nil
		}
		return p.hooks.LockContention(identity)
	}); err != nil {
		_ = lock.Close()
		return err
	}
	if err := ctx.Err(); err != nil {
		unlockVNextPublicationFile(lock)
		_ = lock.Close()
		return err
	}
	operation.lock = lock
	operation.lockIdentity = identity
	if err := operation.assertLockBound(); err != nil {
		unlockVNextPublicationFile(lock)
		_ = lock.Close()
		operation.lock = nil
		operation.lockIdentity = vNextPublicationIdentity{}
		return err
	}
	return nil
}

func (operation *vNextPublicationOperation) openGenerations(ctx context.Context, create bool) error {
	if operation == nil || operation.connector == nil {
		return fs.ErrClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := operation.assertLockBound(); err != nil {
		return err
	}
	if operation.generations != nil {
		return nil
	}
	var (
		generations *vNextPublicationDirectory
		err         error
	)
	if create {
		generations, err = operation.connector.ensureDirectory(vNextPublicationGenerationDirectory, "generation root")
	} else {
		generations, err = operation.connector.openDirectory(vNextPublicationGenerationDirectory, "generation root")
	}
	if err != nil {
		return err
	}
	operation.generations = generations
	return nil
}

func (p *vNextGenerationPublisher) openOperation(ctx context.Context, mode int, create bool) (*vNextPublicationOperation, error) {
	operation, err := p.openOperationRoot(ctx, create)
	if err != nil {
		return nil, err
	}
	if err := p.acquireOperation(ctx, operation, mode); err != nil {
		operation.close()
		return nil, err
	}
	if err := operation.openGenerations(ctx, create); err != nil {
		operation.close()
		return nil, err
	}
	return operation, nil
}

func (operation *vNextPublicationOperation) close() {
	if operation == nil {
		return
	}
	unlockVNextPublicationFile(operation.lock)
	operation.lock = nil
	operation.lockIdentity = vNextPublicationIdentity{}
	_ = operation.generations.Close()
	operation.generations = nil
	_ = operation.connector.Close()
	operation.connector = nil
}

func (p *vNextGenerationPublisher) readSourceLock(operation *vNextPublicationOperation) ([]byte, error) {
	raw, err := operation.connector.readFile("source.lock.json", "source lock")
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (p *vNextGenerationPublisher) PublishContext(ctx context.Context, artifacts vNextPublicationArtifacts) (vNextGenerationPointer, error) {
	if err := ctx.Err(); err != nil {
		return vNextGenerationPointer{}, err
	}
	files, err := vNextPublicationFiles(artifacts.Files)
	if err != nil {
		return vNextGenerationPointer{}, err
	}
	operation, err := p.openOperation(ctx, syscall.LOCK_EX, true)
	if err != nil {
		return vNextGenerationPointer{}, err
	}
	defer operation.close()
	return p.publishLocked(operation, files, artifacts.Validate)
}

func (p *vNextGenerationPublisher) publishLocked(operation *vNextPublicationOperation, files map[string][]byte, validate func(fs.FS) error) (vNextGenerationPointer, error) {
	if err := operation.assertLockBound(); err != nil {
		return vNextGenerationPointer{}, err
	}
	if err := p.recoverLocked(operation); err != nil {
		return vNextGenerationPointer{}, err
	}
	old, hasOld, err := p.readCurrentLocked(operation)
	if err != nil {
		return vNextGenerationPointer{}, err
	}
	stage, pointer, err := p.stageLocked(operation, files, validate)
	if err != nil {
		return vNextGenerationPointer{}, err
	}
	if hasOld && old == pointer {
		return pointer, nil
	}
	if err := p.writeJournalLocked(operation, vNextGenerationJournal{Old: pointerOrNil(old, hasOld), New: pointer, State: "prepared"}); err != nil {
		return vNextGenerationPointer{}, err
	}
	if stage != "" {
		if err := p.activateStageLocked(operation, stage, pointer.Generation); err != nil {
			return vNextGenerationPointer{}, err
		}
	}
	if err := p.writeCurrentLocked(operation, pointer); err != nil {
		return vNextGenerationPointer{}, err
	}
	if err := p.hit(vNextPublicationBeforeActiveValidate); err != nil {
		return vNextGenerationPointer{}, err
	}
	_, validationErr := p.validatePointerLocked(operation, pointer, files, validate)
	if err := p.hit(vNextPublicationAfterActiveValidate); err != nil {
		return vNextGenerationPointer{}, err
	}
	if validationErr != nil {
		if rollbackErr := p.rollbackLocked(operation, old, hasOld, pointer); rollbackErr != nil {
			return vNextGenerationPointer{}, fmt.Errorf("validate active generation: %w; restore previous generation: %w", validationErr, rollbackErr)
		}
		return vNextGenerationPointer{}, fmt.Errorf("validate active generation: %w", validationErr)
	}
	if err := p.writeJournalLocked(operation, vNextGenerationJournal{Old: pointerOrNil(old, hasOld), New: pointer, State: "committed"}); err != nil {
		return vNextGenerationPointer{}, err
	}
	if err := p.pruneLocked(operation, pointer.Generation); err != nil {
		return vNextGenerationPointer{}, err
	}
	if err := p.removeJournalLocked(operation); err != nil {
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
	operation, err := p.openOperation(ctx, syscall.LOCK_SH, false)
	if err != nil {
		return err
	}
	defer operation.close()
	return p.checkLocked(operation, files, artifacts.Validate)
}

func (p *vNextGenerationPublisher) checkLocked(operation *vNextPublicationOperation, files map[string][]byte, validate func(fs.FS) error) error {
	if err := operation.assertLockBound(); err != nil {
		return err
	}
	if err := p.assertNoPendingJournalLocked(operation); err != nil {
		return err
	}
	pointer, found, err := p.readCurrentLocked(operation)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("connector %q has no active generation", p.connector)
	}
	if _, err := p.validatePointerLocked(operation, pointer, files, validate); err != nil {
		return err
	}
	return p.assertNoOrphansLocked(operation, pointer.Generation)
}

func (p *vNextGenerationPublisher) Recover(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	operation, err := p.openOperation(ctx, syscall.LOCK_EX, true)
	if err != nil {
		return err
	}
	defer operation.close()
	return p.recoverLocked(operation)
}

func (p *vNextGenerationPublisher) Open(ctx context.Context) (*vNextGenerationHandle, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	operation, err := p.openOperation(ctx, syscall.LOCK_EX, true)
	if err != nil {
		return nil, err
	}
	defer operation.close()
	if err := p.recoverLocked(operation); err != nil {
		return nil, err
	}
	return p.openLocked(operation)
}

func (p *vNextGenerationPublisher) openLocked(operation *vNextPublicationOperation) (*vNextGenerationHandle, error) {
	pointer, found, err := p.readCurrentLocked(operation)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("connector %q has no active generation", p.connector)
	}
	generation, err := operation.generations.openDirectory(pointer.Generation, fmt.Sprintf("generation %q", pointer.Generation))
	if err != nil {
		return nil, err
	}
	keepGeneration := false
	defer func() {
		if !keepGeneration {
			_ = generation.Close()
		}
	}()
	integrity, err := p.validatePointerRootLocked(operation, generation, pointer, nil, nil)
	if err != nil {
		return nil, err
	}
	if err := p.hit(vNextPublicationAfterOpenValidation); err != nil {
		return nil, err
	}
	hold, err := generation.duplicateFile(fmt.Sprintf("generation hold %q", pointer.Generation))
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(hold.Fd()), syscall.LOCK_SH); err != nil {
		_ = hold.Close()
		return nil, fmt.Errorf("hold generation directory %q: %w", pointer.Generation, err)
	}
	files := make(map[string]struct{}, len(integrity.Files))
	for _, file := range integrity.Files {
		files[file.Path] = struct{}{}
	}
	keepGeneration = true
	return &vNextGenerationHandle{filesRoot: generation, pointer: pointer, files: files, hold: hold}, nil
}

func (p *vNextGenerationPublisher) Prune(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	operation, err := p.openOperation(ctx, syscall.LOCK_EX, true)
	if err != nil {
		return err
	}
	defer operation.close()
	if err := p.recoverLocked(operation); err != nil {
		return err
	}
	pointer, found, err := p.readCurrentLocked(operation)
	if err != nil {
		return err
	}
	if !found {
		return p.pruneLocked(operation, "")
	}
	if _, err := p.validatePointerLocked(operation, pointer, nil, nil); err != nil {
		return fmt.Errorf("validate active generation before pruning: %w", err)
	}
	return p.pruneLocked(operation, pointer.Generation)
}

func (p *vNextGenerationPublisher) GenerationExists(generation string) bool {
	if !vNextPublicationGenerationIDValid(generation) {
		return false
	}
	connector, err := p.openConnectorRoot(false)
	if err != nil {
		return false
	}
	generations, err := connector.openDirectory(vNextPublicationGenerationDirectory, "generation root")
	if err != nil {
		if closeErr := connector.Close(); closeErr != nil {
			return false
		}
		return false
	}
	existing, err := generations.openDirectory(generation, fmt.Sprintf("generation %q", generation))
	if err != nil {
		if closeErr := generations.Close(); closeErr != nil {
			return false
		}
		if closeErr := connector.Close(); closeErr != nil {
			return false
		}
		return false
	}
	if closeErr := existing.Close(); closeErr != nil {
		return false
	}
	if closeErr := generations.Close(); closeErr != nil {
		return false
	}
	return connector.Close() == nil
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
	return h.filesRoot.readFile(name, "generation artifact")
}

func (h *vNextGenerationHandle) Release() {
	if h == nil || h.released {
		return
	}
	h.released = true
	unlockVNextPublicationFile(h.hold)
	_ = h.filesRoot.Close()
}

func (p *vNextGenerationPublisher) stageLocked(operation *vNextPublicationOperation, files map[string][]byte, validate func(fs.FS) error) (string, vNextGenerationPointer, error) {
	if err := operation.assertLockBound(); err != nil {
		return "", vNextGenerationPointer{}, err
	}
	generation := vNextPublicationGenerationID(files)
	existing, err := operation.generations.openDirectory(generation, fmt.Sprintf("generation %q", generation))
	if err == nil {
		_ = existing.Close()
		pointer, err := p.pointerForGenerationLocked(operation, generation)
		if err != nil {
			return "", vNextGenerationPointer{}, err
		}
		if _, err := p.validatePointerLocked(operation, pointer, files, validate); err != nil {
			return "", vNextGenerationPointer{}, fmt.Errorf("existing generation %q is not the requested closed set: %w", generation, err)
		}
		return "", pointer, nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return "", vNextGenerationPointer{}, err
	}

	stageName, stage, err := vNextPublicationCreateStage(operation.generations)
	if err != nil {
		return "", vNextGenerationPointer{}, err
	}
	stageIdentity, err := vNextPublicationIdentityFromFile(stage.file, "publication stage")
	if err != nil {
		_ = stage.Close()
		return "", vNextGenerationPointer{}, err
	}
	retained := false
	defer func() {
		if !retained {
			_ = p.removeStageBoundLocked(operation, stageName, stage, stageIdentity)
		}
		_ = stage.Close()
	}()
	marker, err := vNextPublicationJSON(vNextPublicationStageOwner{
		Version:    1,
		Connector:  p.connector,
		Generation: generation,
		Stage:      stageName,
	})
	if err != nil {
		return "", vNextGenerationPointer{}, fmt.Errorf("encode stage ownership marker: %w", err)
	}
	if err := p.writeStageFile(stage, vNextPublicationStageOwnerFile, marker); err != nil {
		return "", vNextGenerationPointer{}, err
	}
	if err := stage.Sync(); err != nil {
		return "", vNextGenerationPointer{}, fmt.Errorf("sync stage ownership marker: %w", err)
	}
	if err := operation.generations.Sync(); err != nil {
		return "", vNextGenerationPointer{}, fmt.Errorf("sync stage ownership parent: %w", err)
	}
	for _, name := range sortedVNextPublicationFiles(files) {
		if err := p.writeStageFile(stage, name, files[name]); err != nil {
			return "", vNextGenerationPointer{}, err
		}
	}
	if err := p.writeStageFile(stage, vNextPublicationLeaseFile, nil); err != nil {
		return "", vNextGenerationPointer{}, err
	}
	integrity, payload, pointer, err := vNextPublicationIntegrityForFiles(p.connector, generation, files)
	if err != nil {
		return "", vNextGenerationPointer{}, fmt.Errorf("encode generation integrity: %w", err)
	}
	if err := p.writeStageFile(stage, vNextPublicationIntegrityFile, payload); err != nil {
		return "", vNextGenerationPointer{}, err
	}
	if err := p.syncTreeDirectories(stage); err != nil {
		return "", vNextGenerationPointer{}, err
	}
	if err := p.hit(vNextPublicationBeforeStageDirectory); err != nil {
		return "", vNextGenerationPointer{}, err
	}
	if err := stage.Sync(); err != nil {
		return "", vNextGenerationPointer{}, fmt.Errorf("sync staged generation: %w", err)
	}
	if err := p.hit(vNextPublicationAfterStageDirectory); err != nil {
		return "", vNextGenerationPointer{}, err
	}
	if _, err := p.validateGenerationLocked(operation, stage, pointer); err != nil {
		return "", vNextGenerationPointer{}, err
	}
	if validate != nil {
		if err := validate(vNextPublicationDirectoryFS{root: stage}); err != nil {
			return "", vNextGenerationPointer{}, fmt.Errorf("validate staged generation: %w", err)
		}
	}
	if integrity.Generation != generation {
		return "", vNextGenerationPointer{}, fmt.Errorf("staged integrity generation %q does not match %q", integrity.Generation, generation)
	}
	retained = true
	return stageName, pointer, nil
}

func (p *vNextGenerationPublisher) activateStageLocked(operation *vNextPublicationOperation, stageName, generation string) error {
	if err := operation.assertLockBound(); err != nil {
		return err
	}
	if err := p.hit(vNextPublicationBeforeStageRename); err != nil {
		return err
	}
	if err := operation.generations.renameNoReplaceFrom(operation.generations, stageName, generation); err != nil {
		return fmt.Errorf("activate staged generation %q: %w", generation, err)
	}
	if err := operation.generations.Sync(); err != nil {
		return fmt.Errorf("sync generation root: %w", err)
	}
	return p.hit(vNextPublicationAfterStageRename)
}

func vNextPublicationCreateStage(generations *vNextPublicationDirectory) (string, *vNextPublicationDirectory, error) {
	var token [16]byte
	for range 128 {
		if _, err := cryptorand.Read(token[:]); err != nil {
			return "", nil, fmt.Errorf("generate publication stage name: %w", err)
		}
		name := ".stage-" + hex.EncodeToString(token[:])
		if err := unix.Mkdirat(int(generations.file.Fd()), name, 0o755); err != nil {
			if errors.Is(err, unix.EEXIST) {
				continue
			}
			return "", nil, fmt.Errorf("create same-filesystem stage: %w", err)
		}
		stage, err := generations.openDirectory(name, "publication stage")
		if err != nil {
			return "", nil, err
		}
		return name, stage, nil
	}
	return "", nil, fmt.Errorf("create same-filesystem stage: exhausted unique names")
}

func (p *vNextGenerationPublisher) writeStageFile(stage *vNextPublicationDirectory, name string, payload []byte) error {
	file, err := stage.openFile(name, "staged file "+name, unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o644, true)
	if err != nil {
		return err
	}
	if _, err := file.Write(payload); err != nil {
		_ = p.closeStageFile(file)
		return fmt.Errorf("write staged file %s: %w", name, err)
	}
	if err := p.hit(vNextPublicationBeforeFileSync); err != nil {
		_ = p.closeStageFile(file)
		return err
	}
	if err := file.Sync(); err != nil {
		_ = p.closeStageFile(file)
		return fmt.Errorf("sync staged file %s: %w", name, err)
	}
	if err := p.hit(vNextPublicationAfterFileSync); err != nil {
		_ = p.closeStageFile(file)
		return err
	}
	if err := p.closeStageFile(file); err != nil {
		return fmt.Errorf("close staged file %s: %w", name, err)
	}
	return nil
}

func (p *vNextGenerationPublisher) closeStageFile(file *os.File) error {
	if p.hooks.CloseStageFile != nil {
		return p.hooks.CloseStageFile(file)
	}
	return file.Close()
}

func (p *vNextGenerationPublisher) syncTreeDirectories(root *vNextPublicationDirectory) error {
	entries, err := root.readDir()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		stat, err := root.lstat(entry.Name(), "staged member "+entry.Name())
		if err != nil {
			return err
		}
		if vNextPublicationStatIsSymlink(stat) {
			return fmt.Errorf("staged member %q is a symlink", entry.Name())
		}
		if vNextPublicationStatIsDir(stat) {
			child, err := root.openDirectory(entry.Name(), "staged directory "+entry.Name())
			if err != nil {
				return err
			}
			err = p.syncTreeDirectories(child)
			closeErr := child.Close()
			if err != nil {
				return err
			}
			if closeErr != nil {
				return fmt.Errorf("close staged directory %s: %w", entry.Name(), closeErr)
			}
			continue
		}
		if !vNextPublicationStatIsRegular(stat) {
			return fmt.Errorf("staged member %q is not regular", entry.Name())
		}
	}
	return root.Sync()
}

func (p *vNextGenerationPublisher) writeJournalLocked(operation *vNextPublicationOperation, journal vNextGenerationJournal) error {
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
	return p.writeAtomicLocked(operation, vNextPublicationJournalFile, payload, before, after, "", "", "", "")
}

func (p *vNextGenerationPublisher) writeCurrentLocked(operation *vNextPublicationOperation, pointer vNextGenerationPointer) error {
	if err := vNextPublicationPointerValid(pointer); err != nil {
		return err
	}
	payload, err := json.Marshal(pointer)
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	return p.writeAtomicLocked(operation, vNextPublicationCurrentFile, payload, vNextPublicationBeforeCurrentTempSync, vNextPublicationAfterCurrentTempSync, vNextPublicationBeforeCurrentRename, vNextPublicationAfterCurrentRename, vNextPublicationBeforeCurrentParent, vNextPublicationAfterCurrentParent)
}

func (p *vNextGenerationPublisher) closeAtomicTemporary(file *os.File) error {
	if p.hooks.CloseAtomicTemporary != nil {
		return p.hooks.CloseAtomicTemporary(file)
	}
	return file.Close()
}

func (p *vNextGenerationPublisher) closeRepairPredecessor(directory *vNextPublicationDirectory) error {
	if p.hooks.CloseRepairPredecessor != nil {
		return p.hooks.CloseRepairPredecessor(directory)
	}
	return directory.Close()
}

func (p *vNextGenerationPublisher) writeAtomicLocked(operation *vNextPublicationOperation, target string, payload []byte, beforeSync, afterSync, beforeRename, afterRename, beforeParent, afterParent vNextPublicationFaultPoint) (resultErr error) {
	if err := operation.assertLockBound(); err != nil {
		return err
	}
	temporaryName, temporaryRoot, temporary, err := vNextPublicationCreateTemp(operation.connector, p.hooks.AfterTemporaryOpen)
	if err != nil {
		return err
	}
	temporaryRootIdentity, err := vNextPublicationIdentityFromFile(temporaryRoot.file, "publication temporary root")
	if err != nil {
		resultErr = err
		vNextPublicationRecordError(&resultErr, "close publication temporary", p.closeAtomicTemporary(temporary))
		vNextPublicationCloseAfter(&resultErr, temporaryRoot, "publication temporary root")
		return resultErr
	}
	identity, err := vNextPublicationIdentityFromFile(temporary, "publication temporary")
	if err != nil {
		resultErr = err
		vNextPublicationRecordError(&resultErr, "close publication temporary", p.closeAtomicTemporary(temporary))
		vNextPublicationCloseAfter(&resultErr, temporaryRoot, "publication temporary root")
		return resultErr
	}
	defer func() {
		vNextPublicationRecordError(&resultErr, "remove publication temporary", p.removeRegularQuarantinedLocked(temporaryRoot, vNextPublicationTemporaryFile, "publication temporary", temporary, identity, ""))
		vNextPublicationRecordError(&resultErr, "close publication temporary", p.closeAtomicTemporary(temporary))
		vNextPublicationCloseAfter(&resultErr, temporaryRoot, "publication temporary root")
		vNextPublicationRecordError(&resultErr, "remove publication temporary root", operation.connector.removeEmptyDirectoryBound(temporaryName, "publication temporary root", temporaryRootIdentity))
	}()
	if err := temporary.Chmod(0o644); err != nil {
		return err
	}
	if _, err := temporary.Write(payload); err != nil {
		return err
	}
	if beforeSync != "" {
		if err := p.hit(beforeSync); err != nil {
			return err
		}
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if afterSync != "" {
		if err := p.hit(afterSync); err != nil {
			return err
		}
	}
	if beforeRename != "" {
		if err := p.hit(beforeRename); err != nil {
			return err
		}
	}
	if err := temporaryRoot.assertIdentity(vNextPublicationTemporaryFile, "publication temporary", identity); err != nil {
		return err
	}
	if err := p.hit(vNextPublicationAfterControlSourceIdentity); err != nil {
		return err
	}
	if err := p.hit(vNextPublicationAfterFinalControlSourceIdentity); err != nil {
		return err
	}
	intended := vNextPublicationPresentControlState(vNextPublicationControlReplacementMember, identity)
	selected, err := p.transitionControlLocked(operation, target, intended, &vNextPublicationControlAnchorSource{
		directory: temporaryRoot,
		name:      vNextPublicationTemporaryFile,
		identity:  identity,
	})
	if err != nil {
		return err
	}
	if !selected.sameLogical(intended) {
		return &vNextPublicationControlConflictError{target: target}
	}
	operation.bindControl(target, identity)
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
	if err := operation.connector.Sync(); err != nil {
		return err
	}
	if afterParent != "" {
		if err := p.hit(afterParent); err != nil {
			return err
		}
	}
	return nil
}

const vNextPublicationTemporaryFile = "control"

func vNextPublicationCreateTemp(root *vNextPublicationDirectory, afterOpen func(*vNextPublicationDirectory, string, *vNextPublicationDirectory) error) (string, *vNextPublicationDirectory, *os.File, error) {
	var token [16]byte
	for range 128 {
		if _, err := cryptorand.Read(token[:]); err != nil {
			return "", nil, nil, fmt.Errorf("generate publication temporary name: %w", err)
		}
		name := ".connectorgen-publication-" + hex.EncodeToString(token[:])
		if err := unix.Mkdirat(int(root.file.Fd()), name, 0o700); err != nil {
			if errors.Is(err, fs.ErrExist) {
				continue
			}
			return "", nil, nil, fmt.Errorf("create publication temporary root: %w", err)
		}
		temporaryRoot, err := root.openDirectory(name, "publication temporary root")
		if err != nil {
			return "", nil, nil, err
		}
		if afterOpen != nil {
			if err := afterOpen(root, name, temporaryRoot); err != nil {
				_ = temporaryRoot.Close()
				return "", nil, nil, err
			}
		}
		file, err := temporaryRoot.openFile(vNextPublicationTemporaryFile, "publication temporary", unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o644, false)
		if err == nil {
			return name, temporaryRoot, file, nil
		}
		_ = temporaryRoot.Close()
		if rootIdentity, identityErr := root.identityAt(name, "publication temporary root"); identityErr == nil {
			_ = root.removeEmptyDirectoryBound(name, "publication temporary root", rootIdentity)
		}
		if !errors.Is(err, fs.ErrExist) {
			return "", nil, nil, err
		}
	}
	return "", nil, nil, fmt.Errorf("create publication temporary root: exhausted unique names")
}

const vNextPublicationQuarantineMember = "candidate"

type vNextPublicationRemovalBinding struct {
	name     string
	label    string
	identity vNextPublicationIdentity
}

func vNextPublicationCreateQuarantine(root *vNextPublicationDirectory, afterOpen func(*vNextPublicationDirectory, string, *vNextPublicationDirectory, vNextPublicationIdentity) error) (string, *vNextPublicationDirectory, vNextPublicationIdentity, error) {
	var token [16]byte
	for range 128 {
		if _, err := cryptorand.Read(token[:]); err != nil {
			return "", nil, vNextPublicationIdentity{}, fmt.Errorf("generate publication quarantine name: %w", err)
		}
		name := ".connectorgen-quarantine-" + hex.EncodeToString(token[:])
		if err := unix.Mkdirat(int(root.file.Fd()), name, 0o700); err != nil {
			if errors.Is(err, fs.ErrExist) {
				continue
			}
			return "", nil, vNextPublicationIdentity{}, fmt.Errorf("create publication quarantine: %w", err)
		}
		quarantine, err := root.openDirectory(name, "publication quarantine")
		if err != nil {
			return "", nil, vNextPublicationIdentity{}, err
		}
		identity, err := vNextPublicationIdentityFromFile(quarantine.file, "publication quarantine")
		if err == nil {
			if afterOpen != nil {
				if hookErr := afterOpen(root, name, quarantine, identity); hookErr != nil {
					err = hookErr
				} else {
					return name, quarantine, identity, nil
				}
			} else {
				return name, quarantine, identity, nil
			}
		}
		_ = quarantine.Close()
		if rootIdentity, identityErr := root.identityAt(name, "publication quarantine"); identityErr == nil {
			_ = root.removeEmptyDirectoryBound(name, "publication quarantine", rootIdentity)
		}
		return "", nil, vNextPublicationIdentity{}, err
	}
	return "", nil, vNextPublicationIdentity{}, fmt.Errorf("create publication quarantine: exhausted unique names")
}

const (
	vNextPublicationControlBackupMember      = "prior"
	vNextPublicationControlReplacementMember = "replacement"
	vNextPublicationControlCaptureMember     = "candidate"
)

func vNextPublicationAssertRemovalBindings(root *vNextPublicationDirectory, bindings []vNextPublicationRemovalBinding) error {
	for _, binding := range bindings {
		if err := root.assertIdentity(binding.name, binding.label, binding.identity); err != nil {
			return err
		}
	}
	return nil
}

func (p *vNextGenerationPublisher) restoreQuarantinedLocked(parent *vNextPublicationDirectory, name, label string, quarantine *vNextPublicationDirectory, candidateIdentity vNextPublicationIdentity) error {
	if err := quarantine.assertIdentity(vNextPublicationQuarantineMember, label, candidateIdentity); err != nil {
		return err
	}
	if _, err := parent.identityAt(name, label); err == nil {
		return fmt.Errorf("%s identity changed while restoring quarantined candidate", label)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	if err := p.hit(vNextPublicationBeforeQuarantineRestore); err != nil {
		return err
	}
	return parent.renameNoReplaceFrom(quarantine, vNextPublicationQuarantineMember, name)
}

func (p *vNextGenerationPublisher) refuseQuarantinedReplacementLocked(parent *vNextPublicationDirectory, name, label string, quarantine *vNextPublicationDirectory, candidateIdentity vNextPublicationIdentity) error {
	cause := fmt.Errorf("%s identity changed", label)
	if err := p.restoreQuarantinedLocked(parent, name, label, quarantine, candidateIdentity); err != nil {
		return fmt.Errorf("%w; retain quarantined candidate: %w", cause, err)
	}
	return cause
}

func (p *vNextGenerationPublisher) removeRegularQuarantinedLocked(parent *vNextPublicationDirectory, name, label string, file *os.File, identity vNextPublicationIdentity, afterIdentity vNextPublicationFaultPoint) (resultErr error) {
	actual, err := vNextPublicationIdentityFromFile(file, label)
	if err != nil {
		return err
	}
	if actual != identity {
		return fmt.Errorf("%s identity changed", label)
	}
	if err := parent.assertIdentity(name, label, identity); err != nil {
		return err
	}
	if afterIdentity != "" {
		if err := p.hit(afterIdentity); err != nil {
			return err
		}
	}
	quarantineName, quarantine, quarantineIdentity, err := vNextPublicationCreateQuarantine(parent, p.hooks.AfterQuarantineOpen)
	if err != nil {
		return err
	}
	defer func() {
		_ = parent.removeEmptyDirectoryBound(quarantineName, "publication quarantine", quarantineIdentity)
	}()
	defer vNextPublicationCloseAfter(&resultErr, quarantine, "publication quarantine")

	if err := quarantine.renameFrom(parent, name, vNextPublicationQuarantineMember); err != nil {
		return err
	}
	candidateIdentity, err := quarantine.identityAt(vNextPublicationQuarantineMember, label)
	if err != nil {
		return err
	}
	if candidateIdentity != identity {
		return p.refuseQuarantinedReplacementLocked(parent, name, label, quarantine, candidateIdentity)
	}
	return quarantine.removeRegularBound(vNextPublicationQuarantineMember, label, identity)
}

func (p *vNextGenerationPublisher) removeTreeQuarantinedLocked(parent *vNextPublicationDirectory, name, label string, root *vNextPublicationDirectory, identity vNextPublicationIdentity, bindings []vNextPublicationRemovalBinding, afterBindings, afterEntry vNextPublicationFaultPoint) (resultErr error) {
	actual, err := vNextPublicationIdentityFromFile(root.file, label)
	if err != nil {
		return err
	}
	if actual != identity {
		return fmt.Errorf("%s identity changed", label)
	}
	if err := vNextPublicationAssertRemovalBindings(root, bindings); err != nil {
		return err
	}
	if afterBindings != "" {
		if err := p.hit(afterBindings); err != nil {
			return err
		}
	}
	if err := parent.assertIdentity(name, label, identity); err != nil {
		return err
	}
	if afterEntry != "" {
		if err := p.hit(afterEntry); err != nil {
			return err
		}
	}
	quarantineName, quarantine, quarantineIdentity, err := vNextPublicationCreateQuarantine(parent, p.hooks.AfterQuarantineOpen)
	if err != nil {
		return err
	}
	defer func() {
		_ = parent.removeEmptyDirectoryBound(quarantineName, "publication quarantine", quarantineIdentity)
	}()
	defer vNextPublicationCloseAfter(&resultErr, quarantine, "publication quarantine")

	if err := quarantine.renameFrom(parent, name, vNextPublicationQuarantineMember); err != nil {
		return err
	}
	candidateIdentity, err := quarantine.identityAt(vNextPublicationQuarantineMember, label)
	if err != nil {
		return err
	}
	if candidateIdentity != identity {
		return p.refuseQuarantinedReplacementLocked(parent, name, label, quarantine, candidateIdentity)
	}
	candidate, err := quarantine.openDirectory(vNextPublicationQuarantineMember, label)
	if err != nil {
		return p.refuseQuarantinedReplacementLocked(parent, name, label, quarantine, candidateIdentity)
	}
	defer vNextPublicationCloseAfter(&resultErr, candidate, "publication quarantine candidate")
	if err := vNextPublicationAssertRemovalBindings(candidate, bindings); err != nil {
		return p.refuseQuarantinedReplacementLocked(parent, name, label, quarantine, candidateIdentity)
	}
	return quarantine.removeTreeBound(vNextPublicationQuarantineMember, label, candidate, identity)
}

func (p *vNextGenerationPublisher) readCurrentLocked(operation *vNextPublicationOperation) (vNextGenerationPointer, bool, error) {
	if err := operation.assertLockBound(); err != nil {
		return vNextGenerationPointer{}, false, err
	}
	graph, err := p.controlAuthorityForReadLocked(operation)
	if err != nil {
		return vNextGenerationPointer{}, false, err
	}
	defer graph.close()
	payload, found, identity, err := vNextPublicationReadAuthorizedControlLocked(operation, graph, vNextPublicationCurrentFile, "CURRENT")
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
	operation.bindControl(vNextPublicationCurrentFile, identity)
	return pointer, true, nil
}

func (p *vNextGenerationPublisher) pointerForGenerationLocked(operation *vNextPublicationOperation, generation string) (vNextGenerationPointer, error) {
	if !vNextPublicationGenerationIDValid(generation) {
		return vNextGenerationPointer{}, fmt.Errorf("invalid generation %q", generation)
	}
	generationRoot, err := operation.generations.openDirectory(generation, fmt.Sprintf("generation %q", generation))
	if err != nil {
		return vNextGenerationPointer{}, err
	}
	pointer, resultErr := p.pointerForGenerationRootLocked(generationRoot, generation)
	closeErr := generationRoot.Close()
	if resultErr != nil {
		return vNextGenerationPointer{}, resultErr
	}
	if closeErr != nil {
		return vNextGenerationPointer{}, fmt.Errorf("close generation %q: %w", generation, closeErr)
	}
	return pointer, nil
}

func (p *vNextGenerationPublisher) pointerForGenerationRootLocked(generationRoot *vNextPublicationDirectory, generation string) (vNextGenerationPointer, error) {
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

func vNextPublicationReadControl(root *vNextPublicationDirectory, name, label string) ([]byte, bool, error) {
	payload, found, _, err := vNextPublicationReadControlBound(root, name, label)
	return payload, found, err
}

func vNextPublicationReadControlBound(root *vNextPublicationDirectory, name, label string) ([]byte, bool, vNextPublicationIdentity, error) {
	handle, err := root.openRegular(name, label, unix.O_RDONLY)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, vNextPublicationIdentity{}, nil
	}
	if err != nil {
		return nil, true, vNextPublicationIdentity{}, err
	}
	identity, err := vNextPublicationIdentityFromFile(handle, label)
	if err != nil {
		if closeErr := handle.Close(); closeErr != nil {
			return nil, true, vNextPublicationIdentity{}, fmt.Errorf("stat %s: %v; close: %w", label, err, closeErr)
		}
		return nil, true, vNextPublicationIdentity{}, err
	}
	payload, err := vNextPublicationReadOpenControl(handle, label)
	closeErr := handle.Close()
	if err != nil {
		return nil, true, vNextPublicationIdentity{}, err
	}
	if closeErr != nil {
		return nil, true, vNextPublicationIdentity{}, fmt.Errorf("close %s: %w", label, closeErr)
	}
	return payload, true, identity, nil
}

func vNextPublicationReadOpenControl(handle *os.File, label string) ([]byte, error) {
	info, err := handle.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", label, err)
	}
	if info.Size() < 0 || info.Size() > vNextPublicationControlMaxBytes {
		return nil, fmt.Errorf("%s exceeds %d-byte limit", label, vNextPublicationControlMaxBytes)
	}
	payload, err := io.ReadAll(io.LimitReader(handle, vNextPublicationControlMaxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", label, err)
	}
	if len(payload) > vNextPublicationControlMaxBytes {
		return nil, fmt.Errorf("%s exceeds %d-byte limit", label, vNextPublicationControlMaxBytes)
	}
	return payload, nil
}

func (p *vNextGenerationPublisher) validatePointerLocked(operation *vNextPublicationOperation, pointer vNextGenerationPointer, expected map[string][]byte, validate func(fs.FS) error) (vNextGenerationIntegrity, error) {
	if err := vNextPublicationPointerValid(pointer); err != nil {
		return vNextGenerationIntegrity{}, err
	}
	generationRoot, err := operation.generations.openDirectory(pointer.Generation, fmt.Sprintf("generation %q", pointer.Generation))
	if err != nil {
		return vNextGenerationIntegrity{}, err
	}
	integrity, resultErr := p.validatePointerRootLocked(operation, generationRoot, pointer, expected, validate)
	closeErr := generationRoot.Close()
	if resultErr != nil {
		return vNextGenerationIntegrity{}, resultErr
	}
	if closeErr != nil {
		return vNextGenerationIntegrity{}, fmt.Errorf("close generation %q: %w", pointer.Generation, closeErr)
	}
	return integrity, nil
}

func (p *vNextGenerationPublisher) validatePointerRootLocked(operation *vNextPublicationOperation, generationRoot *vNextPublicationDirectory, pointer vNextGenerationPointer, expected map[string][]byte, validate func(fs.FS) error) (vNextGenerationIntegrity, error) {
	if err := vNextPublicationPointerValid(pointer); err != nil {
		return vNextGenerationIntegrity{}, err
	}
	integrity, err := p.validateGenerationLocked(operation, generationRoot, pointer)
	if err != nil {
		return vNextGenerationIntegrity{}, err
	}
	if expected != nil {
		if err := vNextPublicationCompareFiles(integrity.Files, expected); err != nil {
			return vNextGenerationIntegrity{}, err
		}
	}
	if validate != nil {
		if err := validate(vNextPublicationDirectoryFS{root: generationRoot}); err != nil {
			return vNextGenerationIntegrity{}, fmt.Errorf("validate active generation: %w", err)
		}
	}
	return integrity, nil
}

func (p *vNextGenerationPublisher) validateGenerationLocked(operation *vNextPublicationOperation, generationRoot *vNextPublicationDirectory, pointer vNextGenerationPointer) (vNextGenerationIntegrity, error) {
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
	if err := p.validateStageOwnerLocked(generationRoot, "", pointer.Generation); err != nil {
		return vNextGenerationIntegrity{}, err
	}
	computedGeneration, err := vNextPublicationValidateFileDigests(generationRoot, integrity.Files)
	if err != nil {
		return vNextGenerationIntegrity{}, err
	}
	if computedGeneration != pointer.Generation {
		return vNextGenerationIntegrity{}, fmt.Errorf("generation content address %q does not match selected generation %q", computedGeneration, pointer.Generation)
	}
	if err := vNextPublicationValidateClosedTree(generationRoot, integrity.Files); err != nil {
		return vNextGenerationIntegrity{}, err
	}
	return integrity, nil
}

func (p *vNextGenerationPublisher) recoverLocked(operation *vNextPublicationOperation) error {
	if err := operation.assertLockBound(); err != nil {
		return err
	}
	if err := p.recoverControlRepairLocked(operation); err != nil {
		return err
	}
	graph, err := p.controlAuthorityForReadLocked(operation)
	if err != nil {
		return err
	}
	payload, foundJournal, journalIdentity, readErr := vNextPublicationReadAuthorizedControlLocked(operation, graph, vNextPublicationJournalFile, "publication journal")
	graph.close()
	if readErr != nil {
		return readErr
	}
	if foundJournal {
		operation.bindControl(vNextPublicationJournalFile, journalIdentity)
	}
	if !foundJournal {
		if err := p.removeStagesLocked(operation); err != nil {
			return err
		}
		current, found, readErr := p.readCurrentLocked(operation)
		if readErr != nil {
			return readErr
		}
		active := ""
		if found {
			if _, validationErr := p.validatePointerLocked(operation, current, nil, nil); validationErr != nil {
				return fmt.Errorf("recover active generation: %w", validationErr)
			}
			active = current.Generation
		}
		return p.pruneRecoveredLocked(operation, active)
	}
	var journal vNextGenerationJournal
	if err := vNextPublicationDecode(payload, &journal); err != nil {
		return fmt.Errorf("decode publication journal: %w", err)
	}
	if err := vNextPublicationJournalValid(journal); err != nil {
		return err
	}
	current, hasCurrent, err := p.readCurrentLocked(operation)
	if err != nil {
		return err
	}
	if hasCurrent && current != journal.New && (journal.Old == nil || current != *journal.Old) {
		return fmt.Errorf("recover journal: CURRENT does not match either journal generation")
	}
	if hasCurrent && current == journal.New {
		if _, err := p.validatePointerLocked(operation, journal.New, nil, nil); err == nil {
			if err := p.removeStagesLocked(operation); err != nil {
				return err
			}
			if err := p.pruneRecoveredLocked(operation, journal.New.Generation); err != nil {
				return err
			}
			return p.removeJournalLocked(operation)
		}
	}
	if journal.Old != nil {
		if _, err := p.validatePointerLocked(operation, *journal.Old, nil, nil); err != nil {
			return fmt.Errorf("recover journal: old generation is invalid: %w", err)
		}
		if !hasCurrent || current != *journal.Old {
			if err := p.writeCurrentLocked(operation, *journal.Old); err != nil {
				return fmt.Errorf("restore old CURRENT: %w", err)
			}
		}
	} else if hasCurrent && current == journal.New {
		if err := p.removeCurrentLocked(operation); err != nil {
			return err
		}
	}
	if err := p.removeGenerationLocked(operation, journal.New.Generation); err != nil {
		return err
	}
	if err := p.removeStagesLocked(operation); err != nil {
		return err
	}
	if err := operation.generations.Sync(); err != nil {
		return fmt.Errorf("sync recovered generation removal: %w", err)
	}
	return p.removeJournalLocked(operation)
}

func (p *vNextGenerationPublisher) rollbackLocked(operation *vNextPublicationOperation, old vNextGenerationPointer, hasOld bool, rejected vNextGenerationPointer) error {
	if hasOld {
		if err := p.writeCurrentLocked(operation, old); err != nil {
			return err
		}
	} else if err := p.removeCurrentLocked(operation); err != nil {
		return err
	}
	if err := p.removeGenerationLocked(operation, rejected.Generation); err != nil {
		return err
	}
	if err := operation.generations.Sync(); err != nil {
		return fmt.Errorf("sync rejected generation removal: %w", err)
	}
	return p.removeJournalLocked(operation)
}

func (p *vNextGenerationPublisher) removeCurrentLocked(operation *vNextPublicationOperation) error {
	return p.removeControlLocked(operation, vNextPublicationCurrentFile)
}

func (p *vNextGenerationPublisher) removeJournalLocked(operation *vNextPublicationOperation) error {
	return p.removeControlLocked(operation, vNextPublicationJournalFile)
}

func (p *vNextGenerationPublisher) removeControlLocked(operation *vNextPublicationOperation, name string) error {
	if err := operation.assertLockBound(); err != nil {
		return err
	}
	if err := p.hit(vNextPublicationBeforeControlRemoval); err != nil {
		return err
	}
	selected, err := p.transitionControlLocked(operation, name, vNextPublicationAbsentControlState(), nil)
	if err != nil {
		return err
	}
	if selected.Present {
		return &vNextPublicationControlConflictError{target: name}
	}
	operation.clearControl(name)
	return nil
}

func (p *vNextGenerationPublisher) pruneLocked(operation *vNextPublicationOperation, active string) error {
	return p.pruneWithFaultsLocked(operation, active, true)
}

func (p *vNextGenerationPublisher) pruneRecoveredLocked(operation *vNextPublicationOperation, active string) error {
	return p.pruneWithFaultsLocked(operation, active, false)
}

func (p *vNextGenerationPublisher) pruneWithFaultsLocked(operation *vNextPublicationOperation, active string, injectFaults bool) error {
	if err := operation.assertLockBound(); err != nil {
		return err
	}
	if injectFaults {
		if err := p.hit(vNextPublicationBeforePrune); err != nil {
			return err
		}
	}
	entries, err := operation.generations.readDir()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".stage-") {
			if err := p.removeStageLocked(operation, name); err != nil {
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
		if err := p.removeGenerationLocked(operation, name); err != nil {
			return err
		}
	}
	if err := operation.generations.Sync(); err != nil {
		return err
	}
	if injectFaults {
		return p.hit(vNextPublicationAfterPrune)
	}
	return nil
}

func (p *vNextGenerationPublisher) removeStagesLocked(operation *vNextPublicationOperation) error {
	entries, err := operation.generations.readDir()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".stage-") {
			if err := p.removeStageLocked(operation, entry.Name()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *vNextGenerationPublisher) removeStageLocked(operation *vNextPublicationOperation, name string) (resultErr error) {
	if err := operation.assertLockBound(); err != nil {
		return err
	}
	if !vNextPublicationStageNameValid(name) {
		return fmt.Errorf("invalid staging directory %q", name)
	}
	if err := p.hit(vNextPublicationBeforeStageCleanup); err != nil {
		return err
	}
	stage, err := operation.generations.openDirectory(name, fmt.Sprintf("stale staging directory %q", name))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer vNextPublicationCloseAfter(&resultErr, stage, "stale staging directory "+name)
	identity, err := vNextPublicationIdentityFromFile(stage.file, fmt.Sprintf("stale staging directory %q", name))
	if err != nil {
		return err
	}
	return p.removeStageBoundLocked(operation, name, stage, identity)
}

func (p *vNextGenerationPublisher) removeStageBoundLocked(operation *vNextPublicationOperation, name string, stage *vNextPublicationDirectory, identity vNextPublicationIdentity) error {
	if err := p.validateStageOwnerLocked(stage, name, ""); err != nil {
		return fmt.Errorf("refuse remove stale staging directory %q without ownership proof: %w", name, err)
	}
	if err := p.hit(vNextPublicationBeforeStageRemoval); err != nil {
		return err
	}
	return p.removeTreeQuarantinedLocked(operation.generations, name, fmt.Sprintf("stale staging directory %q", name), stage, identity, nil, "", vNextPublicationAfterStageRemovalIdentity)
}

func (p *vNextGenerationPublisher) validateStageOwnerLocked(root *vNextPublicationDirectory, stage, generation string) error {
	payload, found, err := vNextPublicationReadControl(root, vNextPublicationStageOwnerFile, "stage ownership marker")
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

func (p *vNextGenerationPublisher) removeGenerationLocked(operation *vNextPublicationOperation, generation string) (resultErr error) {
	if err := operation.assertLockBound(); err != nil {
		return err
	}
	if !vNextPublicationGenerationIDValid(generation) {
		return fmt.Errorf("invalid generation %q", generation)
	}
	generationRoot, err := operation.generations.openDirectory(generation, fmt.Sprintf("generation %q", generation))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer vNextPublicationCloseAfter(&resultErr, generationRoot, "generation "+generation)
	identity, err := vNextPublicationIdentityFromFile(generationRoot.file, fmt.Sprintf("generation %q", generation))
	if err != nil {
		return err
	}
	pointer, err := p.pointerForGenerationRootLocked(generationRoot, generation)
	if err != nil {
		return fmt.Errorf("refuse prune generation %q without validated publication ownership: %w", generation, err)
	}
	if _, err := p.validateGenerationLocked(operation, generationRoot, pointer); err != nil {
		return fmt.Errorf("refuse prune generation %q without validated publication ownership: %w", generation, err)
	}
	hold, err := generationRoot.duplicateFile(fmt.Sprintf("generation cleanup hold %q", generation))
	if err != nil {
		return err
	}
	defer unlockVNextPublicationFile(hold)
	if err := syscall.Flock(int(hold.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil
		}
		return fmt.Errorf("lock generation directory %q: %w", generation, err)
	}
	lease, err := generationRoot.openRegular(vNextPublicationLeaseFile, fmt.Sprintf("generation lease %q", generation), unix.O_RDONLY)
	if err != nil {
		return err
	}
	defer vNextPublicationCloseAfter(&resultErr, lease, "generation lease "+generation)
	leaseIdentity, err := vNextPublicationIdentityFromFile(lease, fmt.Sprintf("generation lease %q", generation))
	if err != nil {
		return err
	}
	if err := p.hit(vNextPublicationBeforeGenerationRemoval); err != nil {
		return err
	}
	return p.removeTreeQuarantinedLocked(
		operation.generations,
		generation,
		fmt.Sprintf("generation %q", generation),
		generationRoot,
		identity,
		[]vNextPublicationRemovalBinding{{
			name:     vNextPublicationLeaseFile,
			label:    fmt.Sprintf("generation lease %q", generation),
			identity: leaseIdentity,
		}},
		vNextPublicationAfterGenerationLeaseIdentity,
		vNextPublicationAfterGenerationRemovalIdentity,
	)
}

func (p *vNextGenerationPublisher) assertNoOrphansLocked(operation *vNextPublicationOperation, active string) error {
	entries, err := operation.generations.readDir()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		stat, err := operation.generations.lstat(entry.Name(), "generation member "+entry.Name())
		if err != nil {
			return err
		}
		if vNextPublicationStatIsSymlink(stat) {
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

func (p *vNextGenerationPublisher) assertNoPendingJournalLocked(operation *vNextPublicationOperation) error {
	graph, err := p.controlAuthorityForReadLocked(operation)
	if err != nil {
		return err
	}
	defer graph.close()
	_, found, _, err := vNextPublicationReadAuthorizedControlLocked(operation, graph, vNextPublicationJournalFile, "publication journal")
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	return fmt.Errorf("connector %q has a pending publication journal; recover before checking", p.connector)
}

func vNextPublicationAcquireLock(ctx context.Context, lock *os.File, mode int, label string, afterAcquire, onContention func() error) error {
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
			if afterAcquire != nil {
				if err := afterAcquire(); err != nil {
					_ = syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
					return err
				}
			}
			if err := ctx.Err(); err != nil {
				_ = syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
				return err
			}
			return nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EAGAIN) {
			return fmt.Errorf("lock %s: %w", label, err)
		}
		if onContention != nil {
			if err := onContention(); err != nil {
				return err
			}
		}
		retry.Reset(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-retry.C:
		}
	}
}

func vNextPublicationOpenLock(root *vNextPublicationDirectory) (*os.File, vNextPublicationIdentity, error) {
	lock, err := root.duplicateFile("connector publication lock")
	if err != nil {
		return nil, vNextPublicationIdentity{}, err
	}
	identity, err := vNextPublicationIdentityFromFile(lock, "connector publication lock")
	if err != nil {
		_ = lock.Close()
		return nil, vNextPublicationIdentity{}, err
	}
	rootIdentity, err := vNextPublicationIdentityFromFile(root.file, "connector publication root")
	if err != nil {
		_ = lock.Close()
		return nil, vNextPublicationIdentity{}, err
	}
	if rootIdentity != identity {
		_ = lock.Close()
		return nil, vNextPublicationIdentity{}, fmt.Errorf("connector publication root identity changed")
	}
	return lock, identity, nil
}

// vNextPublicationAssertLockBound keeps the complete operation on the retained
// connector-directory inode. Sibling lock files cannot establish another Flock
// domain because this check never treats them as serialization authority.
func vNextPublicationAssertLockBound(root *vNextPublicationDirectory, lock *os.File, identity vNextPublicationIdentity) error {
	actual, err := vNextPublicationIdentityFromFile(lock, "connector publication lock")
	if err != nil {
		return err
	}
	if actual != identity {
		return fmt.Errorf("connector publication lock identity changed")
	}
	rootIdentity, err := vNextPublicationIdentityFromFile(root.file, "connector publication root")
	if err != nil {
		return err
	}
	if rootIdentity != identity {
		return fmt.Errorf("connector publication root identity changed")
	}
	return nil
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
		if rune < '0' || rune > '9' && (rune < 'a' || rune > 'f') {
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
		if rune < '0' || rune > '9' && (rune < 'a' || rune > 'f') {
			return false
		}
	}
	return true
}

func vNextPublicationValidateFileDigests(root *vNextPublicationDirectory, files []vNextGenerationFileDigest) (string, error) {
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
		handle, err := root.openRegular(file.Path, "generation artifact "+file.Path, unix.O_RDONLY)
		if err != nil {
			return "", err
		}
		info, err := handle.Stat()
		if err != nil {
			_ = handle.Close()
			return "", fmt.Errorf("stat generation artifact %s: %w", file.Path, err)
		}
		if info.Size() != int64(file.Bytes) {
			_ = handle.Close()
			return "", fmt.Errorf("generation artifact %s is not the declared regular file", file.Path)
		}
		vNextPublicationWriteLength(content, len(file.Path))
		_, _ = io.WriteString(content, file.Path)
		vNextPublicationWriteLength(content, file.Bytes)

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

func vNextPublicationValidateClosedTree(root *vNextPublicationDirectory, files []vNextGenerationFileDigest) error {
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
	var walk func(*vNextPublicationDirectory, string) error
	walk = func(directory *vNextPublicationDirectory, prefix string) error {
		entries, err := directory.readDir()
		if err != nil {
			return err
		}
		for _, entry := range entries {
			relative := entry.Name()
			if prefix != "" {
				relative = prefix + "/" + relative
			}
			stat, err := directory.lstat(entry.Name(), "generation member "+relative)
			if err != nil {
				return err
			}
			if vNextPublicationStatIsSymlink(stat) {
				return fmt.Errorf("generation contains symlink %s", relative)
			}
			if vNextPublicationStatIsDir(stat) {
				if _, exists := expectedDirectories[relative]; !exists {
					return fmt.Errorf("generation contains unexpected directory %q", relative)
				}
				child, err := directory.openDirectory(entry.Name(), "generation directory "+relative)
				if err != nil {
					return err
				}
				err = walk(child, relative)
				closeErr := child.Close()
				if err != nil {
					return err
				}
				if closeErr != nil {
					return fmt.Errorf("close generation directory %s: %w", relative, closeErr)
				}
				continue
			}
			if !vNextPublicationStatIsRegular(stat) {
				return fmt.Errorf("generation contains nonregular member %s", relative)
			}
			if _, exists := expectedFiles[relative]; !exists {
				return fmt.Errorf("generation contains unexpected member %q", relative)
			}
			if relative == vNextPublicationLeaseFile && stat.Size != 0 {
				return fmt.Errorf("generation lease must be empty")
			}
			seen[relative] = struct{}{}
		}
		return nil
	}
	if err := walk(root, ""); err != nil {
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
		Validate: func(root fs.FS) error {
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

func vNextValidatePublishedStage(root fs.FS, expected vNextStagedGeneration) error {
	bundle, err := engine.Load(vNextPublicationStageFS{connector: expected.Manifest.Connector, root: root}, expected.Manifest.Connector)
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
