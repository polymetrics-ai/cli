package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	vNextPublicationControlRepairFile    = ".connectorgen-control-repair.json"
	vNextPublicationControlRepairVersion = 1

	vNextPublicationControlRepairPrepared            = "prepared"
	vNextPublicationControlRepairReplacementRetained = "replacement_retained"
	vNextPublicationControlRepairRestored            = "restored"
)

// vNextPublicationRecordedIdentity is the stable on-disk binding for one
// descriptor-relative publication member. It intentionally stores only the
// identity needed to reject a substituted member during repair recovery.
type vNextPublicationRecordedIdentity struct {
	Device uint64 `json:"device"`
	Inode  uint64 `json:"inode"`
	Mode   uint32 `json:"mode"`
}

func vNextPublicationRecordIdentity(identity vNextPublicationIdentity) vNextPublicationRecordedIdentity {
	return vNextPublicationRecordedIdentity{
		Device: identity.device,
		Inode:  identity.inode,
		Mode:   identity.mode,
	}
}

func (identity vNextPublicationRecordedIdentity) publicationIdentity(mode uint32) (vNextPublicationIdentity, error) {
	if identity.Inode == 0 || identity.Mode != mode {
		return vNextPublicationIdentity{}, fmt.Errorf("invalid recorded publication identity")
	}
	return vNextPublicationIdentity{
		device: identity.Device,
		inode:  identity.Inode,
		mode:   identity.Mode,
	}, nil
}

// vNextPublicationControlRepair is the durable authority for a final control
// transition. Until it is removed after a synced resolution, CURRENT and
// JOURNAL are deliberately not ordinary recovery inputs.
type vNextPublicationControlRepair struct {
	Version            int                               `json:"version"`
	Target             string                            `json:"target"`
	Phase              string                            `json:"phase"`
	Quarantine         string                            `json:"quarantine"`
	QuarantineIdentity vNextPublicationRecordedIdentity  `json:"quarantine_identity"`
	Expected           vNextPublicationRecordedIdentity  `json:"expected"`
	Prior              *vNextPublicationRecordedIdentity `json:"prior,omitempty"`
	Replacement        *vNextPublicationRecordedIdentity `json:"replacement,omitempty"`
}

type vNextPublicationControlRepairState struct {
	record             vNextPublicationControlRepair
	authorityIdentity  vNextPublicationIdentity
	quarantineName     string
	quarantine         *vNextPublicationDirectory
	quarantineIdentity vNextPublicationIdentity
	authorityCleared   bool
}

func vNextPublicationControlRepairTargetValid(target string) bool {
	return target == vNextPublicationCurrentFile || target == vNextPublicationJournalFile
}

func vNextPublicationControlRepairPhaseValid(phase string) bool {
	switch phase {
	case vNextPublicationControlRepairPrepared,
		vNextPublicationControlRepairReplacementRetained,
		vNextPublicationControlRepairRestored:
		return true
	default:
		return false
	}
}

func vNextPublicationControlRepairValid(repair vNextPublicationControlRepair) error {
	if repair.Version != vNextPublicationControlRepairVersion {
		return fmt.Errorf("invalid publication control repair version %d", repair.Version)
	}
	if !vNextPublicationControlRepairTargetValid(repair.Target) {
		return fmt.Errorf("invalid publication control repair target %q", repair.Target)
	}
	if !vNextPublicationControlRepairPhaseValid(repair.Phase) {
		return fmt.Errorf("invalid publication control repair phase %q", repair.Phase)
	}
	if !vNextPublicationDirectNameValid(repair.Quarantine) || !strings.HasPrefix(repair.Quarantine, ".connectorgen-quarantine-") {
		return fmt.Errorf("invalid publication control repair quarantine %q", repair.Quarantine)
	}
	if _, err := repair.QuarantineIdentity.publicationIdentity(unix.S_IFDIR); err != nil {
		return fmt.Errorf("invalid publication control repair quarantine identity: %w", err)
	}
	if _, err := repair.Expected.publicationIdentity(unix.S_IFREG); err != nil {
		return fmt.Errorf("invalid publication control repair expected identity: %w", err)
	}
	if repair.Prior != nil {
		if _, err := repair.Prior.publicationIdentity(unix.S_IFREG); err != nil {
			return fmt.Errorf("invalid publication control repair prior identity: %w", err)
		}
	}
	if repair.Replacement != nil {
		if _, err := repair.Replacement.publicationIdentity(unix.S_IFREG); err != nil {
			return fmt.Errorf("invalid publication control repair replacement identity: %w", err)
		}
	}
	if repair.Phase == vNextPublicationControlRepairPrepared && repair.Replacement != nil {
		return fmt.Errorf("prepared publication control repair retains a replacement")
	}
	if repair.Phase == vNextPublicationControlRepairReplacementRetained && repair.Replacement == nil {
		return fmt.Errorf("replacement-retained publication control repair has no replacement")
	}
	return nil
}

func (p *vNextGenerationPublisher) beginControlRepairLocked(operation *vNextPublicationOperation, target string, expected vNextPublicationIdentity) (*vNextPublicationControlRepairState, error) {
	if err := operation.assertLockBound(); err != nil {
		return nil, err
	}
	quarantineName, quarantine, quarantineIdentity, err := vNextPublicationCreateQuarantine(operation.connector)
	if err != nil {
		return nil, err
	}
	keepQuarantine := false
	defer func() {
		if !keepQuarantine {
			_ = quarantine.Close()
			_ = operation.connector.removeEmptyDirectoryBound(quarantineName, "publication control repair quarantine", quarantineIdentity)
		}
	}()

	prior, hasPrior, err := vNextPublicationBackupControl(operation, quarantine, target)
	if err != nil {
		return nil, err
	}
	if err := quarantine.Sync(); err != nil {
		return nil, fmt.Errorf("sync publication control repair backup: %w", err)
	}
	if err := p.hit(vNextPublicationAfterControlRepairBackupSync); err != nil {
		return nil, err
	}

	repair := vNextPublicationControlRepair{
		Version:            vNextPublicationControlRepairVersion,
		Target:             target,
		Phase:              vNextPublicationControlRepairPrepared,
		Quarantine:         quarantineName,
		QuarantineIdentity: vNextPublicationRecordIdentity(quarantineIdentity),
		Expected:           vNextPublicationRecordIdentity(expected),
	}
	if hasPrior {
		priorRecord := vNextPublicationRecordIdentity(prior)
		repair.Prior = &priorRecord
	}
	authorityIdentity, err := p.writeControlRepairLocked(operation, repair, nil)
	if err != nil {
		return nil, err
	}
	state := &vNextPublicationControlRepairState{
		record:             repair,
		authorityIdentity:  authorityIdentity,
		quarantineName:     quarantineName,
		quarantine:         quarantine,
		quarantineIdentity: quarantineIdentity,
	}
	keepQuarantine = true
	if err := p.hit(vNextPublicationAfterControlRepairPrepared); err != nil {
		return state, err
	}
	return state, nil
}

func (state *vNextPublicationControlRepairState) close(operation *vNextPublicationOperation) {
	if state == nil {
		return
	}
	_ = state.quarantine.Close()
	if state.authorityCleared {
		_ = operation.connector.removeEmptyDirectoryBound(state.quarantineName, "publication control repair quarantine", state.quarantineIdentity)
	}
}

func (p *vNextGenerationPublisher) writeControlRepairLocked(operation *vNextPublicationOperation, repair vNextPublicationControlRepair, expected *vNextPublicationIdentity) (vNextPublicationIdentity, error) {
	if err := operation.assertLockBound(); err != nil {
		return vNextPublicationIdentity{}, err
	}
	if err := vNextPublicationControlRepairValid(repair); err != nil {
		return vNextPublicationIdentity{}, err
	}
	payload, err := json.Marshal(repair)
	if err != nil {
		return vNextPublicationIdentity{}, err
	}
	payload = append(payload, '\n')

	temporaryName, temporaryRoot, temporary, err := vNextPublicationCreateTemp(operation.connector)
	if err != nil {
		return vNextPublicationIdentity{}, err
	}
	temporaryRootIdentity, err := vNextPublicationIdentityFromFile(temporaryRoot.file, "publication control repair temporary root")
	if err != nil {
		_ = temporary.Close()
		_ = temporaryRoot.Close()
		return vNextPublicationIdentity{}, err
	}
	moved := false
	defer func() {
		if !moved {
			if temporaryIdentity, identityErr := temporaryRoot.identityAt(vNextPublicationTemporaryFile, "publication control repair temporary"); identityErr == nil {
				_ = temporaryRoot.removeRegularBound(vNextPublicationTemporaryFile, "publication control repair temporary", temporaryIdentity)
			}
		}
		_ = temporary.Close()
		_ = temporaryRoot.Close()
		_ = operation.connector.removeEmptyDirectoryBound(temporaryName, "publication control repair temporary root", temporaryRootIdentity)
	}()
	if _, err := temporary.Write(payload); err != nil {
		return vNextPublicationIdentity{}, err
	}
	if err := temporary.Sync(); err != nil {
		return vNextPublicationIdentity{}, err
	}
	if err := temporaryRoot.Sync(); err != nil {
		return vNextPublicationIdentity{}, err
	}
	temporaryIdentity, err := vNextPublicationIdentityFromFile(temporary, "publication control repair temporary")
	if err != nil {
		return vNextPublicationIdentity{}, err
	}
	if expected == nil {
		if _, err := operation.connector.identityAt(vNextPublicationControlRepairFile, "publication control repair authority"); err == nil {
			return vNextPublicationIdentity{}, fmt.Errorf("publication control repair authority already exists")
		} else if !errors.Is(err, fs.ErrNotExist) {
			return vNextPublicationIdentity{}, err
		}
	} else if err := operation.connector.assertIdentity(vNextPublicationControlRepairFile, "publication control repair authority", *expected); err != nil {
		return vNextPublicationIdentity{}, err
	}
	if err := temporaryRoot.assertIdentity(vNextPublicationTemporaryFile, "publication control repair temporary", temporaryIdentity); err != nil {
		return vNextPublicationIdentity{}, err
	}
	if err := operation.connector.renameFrom(temporaryRoot, vNextPublicationTemporaryFile, vNextPublicationControlRepairFile); err != nil {
		return vNextPublicationIdentity{}, err
	}
	moved = true
	if err := operation.connector.assertIdentity(vNextPublicationControlRepairFile, "publication control repair authority", temporaryIdentity); err != nil {
		return vNextPublicationIdentity{}, err
	}
	if err := operation.connector.Sync(); err != nil {
		return vNextPublicationIdentity{}, fmt.Errorf("sync publication control repair authority: %w", err)
	}
	return temporaryIdentity, nil
}

func (p *vNextGenerationPublisher) updateControlRepairLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState) error {
	authorityIdentity, err := p.writeControlRepairLocked(operation, state.record, &state.authorityIdentity)
	if err != nil {
		return err
	}
	state.authorityIdentity = authorityIdentity
	return nil
}

func (p *vNextGenerationPublisher) readControlRepairLocked(operation *vNextPublicationOperation) (*vNextPublicationControlRepairState, bool, error) {
	if err := operation.assertLockBound(); err != nil {
		return nil, false, err
	}
	payload, found, authorityIdentity, err := vNextPublicationReadControlBound(operation.connector, vNextPublicationControlRepairFile, "publication control repair authority")
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}
	var repair vNextPublicationControlRepair
	if err := vNextPublicationDecode(payload, &repair); err != nil {
		return nil, false, fmt.Errorf("decode publication control repair authority: %w", err)
	}
	if err := vNextPublicationControlRepairValid(repair); err != nil {
		return nil, false, err
	}
	quarantineIdentity, err := repair.QuarantineIdentity.publicationIdentity(unix.S_IFDIR)
	if err != nil {
		return nil, false, err
	}
	if err := operation.connector.assertIdentity(repair.Quarantine, "publication control repair quarantine", quarantineIdentity); err != nil {
		return nil, false, err
	}
	quarantine, err := operation.connector.openDirectory(repair.Quarantine, "publication control repair quarantine")
	if err != nil {
		return nil, false, err
	}
	actualQuarantineIdentity, err := vNextPublicationIdentityFromFile(quarantine.file, "publication control repair quarantine")
	if err != nil {
		_ = quarantine.Close()
		return nil, false, err
	}
	if actualQuarantineIdentity != quarantineIdentity {
		_ = quarantine.Close()
		return nil, false, fmt.Errorf("publication control repair quarantine identity changed")
	}
	return &vNextPublicationControlRepairState{
		record:             repair,
		authorityIdentity:  authorityIdentity,
		quarantineName:     repair.Quarantine,
		quarantine:         quarantine,
		quarantineIdentity: quarantineIdentity,
	}, true, nil
}

func (p *vNextGenerationPublisher) recoverControlRepairLocked(operation *vNextPublicationOperation) error {
	state, found, err := p.readControlRepairLocked(operation)
	if err != nil || !found {
		return err
	}
	defer state.close(operation)
	return p.resolveControlRepairLocked(operation, state, false)
}

func vNextPublicationRepairPrior(state *vNextPublicationControlRepairState) (vNextPublicationIdentity, bool, error) {
	if state.record.Prior == nil {
		return vNextPublicationIdentity{}, false, nil
	}
	identity, err := state.record.Prior.publicationIdentity(unix.S_IFREG)
	if err != nil {
		return vNextPublicationIdentity{}, false, err
	}
	return identity, true, nil
}

func vNextPublicationRepairTargetIdentity(root *vNextPublicationDirectory, target string) (vNextPublicationIdentity, bool, error) {
	identity, err := root.identityAt(target, "publication control repair target")
	if errors.Is(err, fs.ErrNotExist) {
		return vNextPublicationIdentity{}, false, nil
	}
	if err != nil {
		return vNextPublicationIdentity{}, false, err
	}
	if identity.mode != unix.S_IFREG {
		return vNextPublicationIdentity{}, false, fmt.Errorf("publication control repair target is not a regular file")
	}
	return identity, true, nil
}

func (p *vNextGenerationPublisher) normalizeControlRepairReplacementLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState) error {
	replacementIdentity, err := state.quarantine.identityAt(vNextPublicationControlReplacementMember, "publication control repair replacement")
	if errors.Is(err, fs.ErrNotExist) {
		if state.record.Replacement != nil {
			return fmt.Errorf("publication control repair replacement disappeared")
		}
		return nil
	}
	if err != nil {
		return err
	}
	if replacementIdentity.mode != unix.S_IFREG {
		return fmt.Errorf("publication control repair replacement is not a regular file")
	}
	if state.record.Replacement != nil {
		expected, err := state.record.Replacement.publicationIdentity(unix.S_IFREG)
		if err != nil {
			return err
		}
		if replacementIdentity != expected {
			return fmt.Errorf("publication control repair replacement identity changed")
		}
		return nil
	}
	recorded := vNextPublicationRecordIdentity(replacementIdentity)
	state.record.Replacement = &recorded
	state.record.Phase = vNextPublicationControlRepairReplacementRetained
	return p.updateControlRepairLocked(operation, state)
}

func (p *vNextGenerationPublisher) retainControlRepairReplacementLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, actual vNextPublicationIdentity, injectFaults bool) (bool, error) {
	if err := p.normalizeControlRepairReplacementLocked(operation, state); err != nil {
		return false, err
	}
	if state.record.Replacement != nil {
		expected, err := state.record.Replacement.publicationIdentity(unix.S_IFREG)
		if err != nil {
			return false, err
		}
		if expected != actual {
			return false, fmt.Errorf("publication control repair observed a second replacement")
		}
		return false, nil
	}
	if err := state.quarantine.linkFromBound(operation.connector, state.record.Target, vNextPublicationControlReplacementMember, "publication control repair replacement", actual); err != nil {
		return false, err
	}
	if err := state.quarantine.Sync(); err != nil {
		return false, fmt.Errorf("sync publication control repair replacement: %w", err)
	}
	if injectFaults {
		if err := p.hit(vNextPublicationAfterControlRepairReplacementRetainSync); err != nil {
			return false, err
		}
	}
	recorded := vNextPublicationRecordIdentity(actual)
	state.record.Replacement = &recorded
	state.record.Phase = vNextPublicationControlRepairReplacementRetained
	if err := p.updateControlRepairLocked(operation, state); err != nil {
		return false, err
	}
	return true, nil
}

func (p *vNextGenerationPublisher) restoreControlRepairLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, actual vNextPublicationIdentity, found bool, prior vNextPublicationIdentity, hasPrior bool) error {
	if hasPrior {
		if found && actual == prior {
			return nil
		}
		if err := state.quarantine.assertIdentity(vNextPublicationControlBackupMember, "publication control repair backup", prior); err != nil {
			return err
		}
		if found {
			if err := operation.connector.assertIdentity(state.record.Target, "publication control repair target", actual); err != nil {
				return err
			}
			if err := operation.connector.renameFrom(state.quarantine, vNextPublicationControlBackupMember, state.record.Target); err != nil {
				return err
			}
		} else if err := operation.connector.linkFromBound(state.quarantine, vNextPublicationControlBackupMember, state.record.Target, "publication control repair backup", prior); err != nil {
			return err
		}
		return operation.connector.assertIdentity(state.record.Target, "publication control repair target", prior)
	}
	if !found {
		return nil
	}
	if err := operation.connector.removeRegularBound(state.record.Target, "publication control repair target", actual); err != nil {
		return err
	}
	return nil
}

func (p *vNextGenerationPublisher) clearControlRepairLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState) error {
	if err := operation.connector.removeRegularBound(vNextPublicationControlRepairFile, "publication control repair authority", state.authorityIdentity); err != nil {
		return err
	}
	if err := operation.connector.Sync(); err != nil {
		return fmt.Errorf("sync cleared publication control repair authority: %w", err)
	}
	state.authorityCleared = true
	return nil
}

func (p *vNextGenerationPublisher) removeControlRepairBackupLocked(state *vNextPublicationControlRepairState) error {
	prior, hasPrior, err := vNextPublicationRepairPrior(state)
	if err != nil || !hasPrior {
		return err
	}
	identity, err := state.quarantine.identityAt(vNextPublicationControlBackupMember, "publication control repair backup")
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if identity != prior {
		return fmt.Errorf("publication control repair backup identity changed")
	}
	if err := state.quarantine.removeRegularBound(vNextPublicationControlBackupMember, "publication control repair backup", prior); err != nil {
		return err
	}
	return state.quarantine.Sync()
}

func (p *vNextGenerationPublisher) resolveControlRepairLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, injectFaults bool) error {
	if err := p.normalizeControlRepairReplacementLocked(operation, state); err != nil {
		return err
	}
	prior, hasPrior, err := vNextPublicationRepairPrior(state)
	if err != nil {
		return err
	}
	actual, found, err := vNextPublicationRepairTargetIdentity(operation.connector, state.record.Target)
	if err != nil {
		return err
	}
	retained := false
	if found && (!hasPrior || actual != prior) {
		retained, err = p.retainControlRepairReplacementLocked(operation, state, actual, injectFaults)
		if err != nil {
			return err
		}
	}
	if retained && injectFaults {
		if err := p.hit(vNextPublicationAfterControlRepairReplacementSync); err != nil {
			return err
		}
	}
	if err := p.restoreControlRepairLocked(operation, state, actual, found, prior, hasPrior); err != nil {
		return err
	}
	if err := operation.connector.Sync(); err != nil {
		return fmt.Errorf("sync restored publication control: %w", err)
	}
	if injectFaults {
		if err := p.hit(vNextPublicationAfterControlRepairPublicRestoreSync); err != nil {
			return err
		}
	}
	if state.record.Phase != vNextPublicationControlRepairRestored {
		state.record.Phase = vNextPublicationControlRepairRestored
		if err := p.updateControlRepairLocked(operation, state); err != nil {
			return err
		}
	}
	if injectFaults {
		if err := p.hit(vNextPublicationAfterControlRepairRestoreSync); err != nil {
			return err
		}
	}
	if err := p.clearControlRepairLocked(operation, state); err != nil {
		return err
	}
	if injectFaults {
		if err := p.hit(vNextPublicationAfterControlRepairClearSync); err != nil {
			return err
		}
	}
	return p.removeControlRepairBackupLocked(state)
}

func (p *vNextGenerationPublisher) refuseMismatchedControlInstallLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState) error {
	cause := fmt.Errorf("publication control identity changed after namespace transition")
	if err := p.resolveControlRepairLocked(operation, state, true); err != nil {
		return fmt.Errorf("%w; durable control repair: %w", cause, err)
	}
	return cause
}

func (p *vNextGenerationPublisher) completeControlRepairLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState) error {
	if err := p.clearControlRepairLocked(operation, state); err != nil {
		return err
	}
	if err := p.hit(vNextPublicationAfterControlRepairClearSync); err != nil {
		return err
	}
	return p.removeControlRepairBackupLocked(state)
}
