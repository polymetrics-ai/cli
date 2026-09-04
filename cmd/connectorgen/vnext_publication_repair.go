package main

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	vNextPublicationControlRepairVersion = 2

	vNextPublicationControlRepairDirectoryPrefix = ".connectorgen-control-repair-"
	vNextPublicationControlRepairPreparedFile    = "prepared.json"
	vNextPublicationControlRepairPhasePrefix     = "phase-"
	vNextPublicationControlRepairPhaseSuffix     = ".json"
	vNextPublicationControlRepairMaxPhases       = 3

	vNextPublicationControlRepairInstalled           = "installed"
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

// vNextPublicationControlRepair is immutable prepared recovery authority for
// one public control transition. It lives only inside its own private,
// identity-bound transaction directory. CURRENT and JOURNAL stay untrusted
// until this record and every verified phase have resolved.
type vNextPublicationControlRepair struct {
	Version             int                               `json:"version"`
	Target              string                            `json:"target"`
	Transaction         string                            `json:"transaction"`
	TransactionIdentity vNextPublicationRecordedIdentity  `json:"transaction_identity"`
	Expected            vNextPublicationRecordedIdentity  `json:"expected"`
	Prior               *vNextPublicationRecordedIdentity `json:"prior,omitempty"`
}

// vNextPublicationControlRepairPhase is an append-only state transition. The
// prepared record is immutable; each phase names and binds its immediate
// predecessor so recovery can select only a contiguous verified chain.
type vNextPublicationControlRepairPhase struct {
	Version          int                                         `json:"version"`
	Sequence         int                                         `json:"sequence"`
	State            string                                      `json:"state"`
	PreparedDigest   string                                      `json:"prepared_digest"`
	PreparedIdentity vNextPublicationRecordedIdentity            `json:"prepared_identity"`
	Previous         vNextPublicationControlRepairPhaseReference `json:"previous"`
	Control          *vNextPublicationRecordedIdentity           `json:"control,omitempty"`
}

type vNextPublicationControlRepairPhaseReference struct {
	Name     string                           `json:"name"`
	Identity vNextPublicationRecordedIdentity `json:"identity"`
	Digest   string                           `json:"digest"`
}

type vNextPublicationControlRepairPhaseState struct {
	record   vNextPublicationControlRepairPhase
	name     string
	identity vNextPublicationIdentity
	digest   string
}

type vNextPublicationControlRepairState struct {
	record              vNextPublicationControlRepair
	preparedIdentity    vNextPublicationIdentity
	preparedDigest      string
	transactionName     string
	transaction         *vNextPublicationDirectory
	transactionIdentity vNextPublicationIdentity
	phases              []vNextPublicationControlRepairPhaseState
	authorityCleared    bool
}

func vNextPublicationControlRepairTargetValid(target string) bool {
	return target == vNextPublicationCurrentFile || target == vNextPublicationJournalFile
}

func vNextPublicationControlRepairTransactionNameValid(name string) bool {
	if !vNextPublicationDirectNameValid(name) || !strings.HasPrefix(name, vNextPublicationControlRepairDirectoryPrefix) {
		return false
	}
	token := strings.TrimPrefix(name, vNextPublicationControlRepairDirectoryPrefix)
	if len(token) != 32 {
		return false
	}
	decoded, err := hex.DecodeString(token)
	return err == nil && len(decoded) == 16
}

func vNextPublicationControlRepairPhaseStateValid(state string) bool {
	switch state {
	case vNextPublicationControlRepairInstalled,
		vNextPublicationControlRepairReplacementRetained,
		vNextPublicationControlRepairRestored:
		return true
	default:
		return false
	}
}

func vNextPublicationControlRepairPhaseName(sequence int) string {
	return fmt.Sprintf("%s%04d%s", vNextPublicationControlRepairPhasePrefix, sequence, vNextPublicationControlRepairPhaseSuffix)
}

func vNextPublicationControlRepairPhaseReferenceValid(reference vNextPublicationControlRepairPhaseReference) error {
	if !vNextPublicationDirectNameValid(reference.Name) {
		return fmt.Errorf("invalid publication control repair phase predecessor name %q", reference.Name)
	}
	if _, err := reference.Identity.publicationIdentity(unix.S_IFREG); err != nil {
		return fmt.Errorf("invalid publication control repair phase predecessor identity: %w", err)
	}
	if !vNextPublicationDigestValid(reference.Digest) {
		return fmt.Errorf("invalid publication control repair phase predecessor digest")
	}
	return nil
}

func vNextPublicationControlRepairValid(repair vNextPublicationControlRepair) error {
	if repair.Version != vNextPublicationControlRepairVersion {
		return fmt.Errorf("invalid publication control repair version %d", repair.Version)
	}
	if !vNextPublicationControlRepairTargetValid(repair.Target) {
		return fmt.Errorf("invalid publication control repair target %q", repair.Target)
	}
	if !vNextPublicationControlRepairTransactionNameValid(repair.Transaction) {
		return fmt.Errorf("invalid publication control repair transaction %q", repair.Transaction)
	}
	if _, err := repair.TransactionIdentity.publicationIdentity(unix.S_IFDIR); err != nil {
		return fmt.Errorf("invalid publication control repair transaction identity: %w", err)
	}
	if _, err := repair.Expected.publicationIdentity(unix.S_IFREG); err != nil {
		return fmt.Errorf("invalid publication control repair expected identity: %w", err)
	}
	if repair.Prior != nil {
		if _, err := repair.Prior.publicationIdentity(unix.S_IFREG); err != nil {
			return fmt.Errorf("invalid publication control repair prior identity: %w", err)
		}
	}
	return nil
}

func vNextPublicationControlRepairPhaseValid(phase vNextPublicationControlRepairPhase) error {
	if phase.Version != vNextPublicationControlRepairVersion {
		return fmt.Errorf("invalid publication control repair phase version %d", phase.Version)
	}
	if phase.Sequence < 1 || phase.Sequence > vNextPublicationControlRepairMaxPhases {
		return fmt.Errorf("invalid publication control repair phase sequence %d", phase.Sequence)
	}
	if !vNextPublicationControlRepairPhaseStateValid(phase.State) {
		return fmt.Errorf("invalid publication control repair phase state %q", phase.State)
	}
	if !vNextPublicationDigestValid(phase.PreparedDigest) {
		return fmt.Errorf("invalid publication control repair prepared digest")
	}
	if _, err := phase.PreparedIdentity.publicationIdentity(unix.S_IFREG); err != nil {
		return fmt.Errorf("invalid publication control repair prepared identity: %w", err)
	}
	if err := vNextPublicationControlRepairPhaseReferenceValid(phase.Previous); err != nil {
		return err
	}
	switch phase.State {
	case vNextPublicationControlRepairInstalled, vNextPublicationControlRepairReplacementRetained:
		if phase.Control == nil {
			return fmt.Errorf("publication control repair phase %q has no control identity", phase.State)
		}
		if _, err := phase.Control.publicationIdentity(unix.S_IFREG); err != nil {
			return fmt.Errorf("invalid publication control repair phase control identity: %w", err)
		}
	case vNextPublicationControlRepairRestored:
		if phase.Control != nil {
			if _, err := phase.Control.publicationIdentity(unix.S_IFREG); err != nil {
				return fmt.Errorf("invalid publication control repair restored control identity: %w", err)
			}
		}
	}
	return nil
}

func vNextPublicationControlRepairPhaseTransitionValid(previous, next string) bool {
	switch previous {
	case "":
		return next == vNextPublicationControlRepairInstalled || next == vNextPublicationControlRepairReplacementRetained || next == vNextPublicationControlRepairRestored
	case vNextPublicationControlRepairInstalled:
		return next == vNextPublicationControlRepairReplacementRetained || next == vNextPublicationControlRepairRestored
	case vNextPublicationControlRepairReplacementRetained:
		return next == vNextPublicationControlRepairRestored
	default:
		return false
	}
}

func vNextPublicationCreateControlRepairTransaction(root *vNextPublicationDirectory) (string, *vNextPublicationDirectory, vNextPublicationIdentity, error) {
	var token [16]byte
	for range 128 {
		if _, err := cryptorand.Read(token[:]); err != nil {
			return "", nil, vNextPublicationIdentity{}, fmt.Errorf("generate publication control repair transaction name: %w", err)
		}
		name := vNextPublicationControlRepairDirectoryPrefix + hex.EncodeToString(token[:])
		if err := unix.Mkdirat(int(root.file.Fd()), name, 0o700); err != nil {
			if errors.Is(err, fs.ErrExist) {
				continue
			}
			return "", nil, vNextPublicationIdentity{}, fmt.Errorf("create publication control repair transaction: %w", err)
		}
		transaction, err := root.openDirectory(name, "publication control repair transaction")
		if err != nil {
			return "", nil, vNextPublicationIdentity{}, err
		}
		identity, err := vNextPublicationIdentityFromFile(transaction.file, "publication control repair transaction")
		if err == nil {
			return name, transaction, identity, nil
		}
		_ = transaction.Close()
		if rootIdentity, identityErr := root.identityAt(name, "publication control repair transaction"); identityErr == nil {
			_ = root.removeEmptyDirectoryBound(name, "publication control repair transaction", rootIdentity)
		}
		return "", nil, vNextPublicationIdentity{}, err
	}
	return "", nil, vNextPublicationIdentity{}, fmt.Errorf("create publication control repair transaction: exhausted unique names")
}

func vNextPublicationWriteControlRepairRecord(directory *vNextPublicationDirectory, name, label string, payload []byte) (identity vNextPublicationIdentity, err error) {
	file, err := directory.openFile(name, label, unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o600, false)
	if err != nil {
		return vNextPublicationIdentity{}, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close %s: %w", label, closeErr)
		}
	}()
	written, err := file.Write(payload)
	if err != nil {
		return vNextPublicationIdentity{}, fmt.Errorf("write %s: %w", label, err)
	}
	if written != len(payload) {
		return vNextPublicationIdentity{}, fmt.Errorf("write %s: short write", label)
	}
	if err := file.Sync(); err != nil {
		return vNextPublicationIdentity{}, fmt.Errorf("sync %s: %w", label, err)
	}
	identity, err = vNextPublicationIdentityFromFile(file, label)
	if err != nil {
		return vNextPublicationIdentity{}, err
	}
	if err := directory.assertIdentity(name, label, identity); err != nil {
		return vNextPublicationIdentity{}, err
	}
	return identity, nil
}

func (state *vNextPublicationControlRepairState) close() {
	if state == nil || state.transaction == nil {
		return
	}
	_ = state.transaction.Close()
	state.transaction = nil
}

func (state *vNextPublicationControlRepairState) preparedReference() vNextPublicationControlRepairPhaseReference {
	return vNextPublicationControlRepairPhaseReference{
		Name:     vNextPublicationControlRepairPreparedFile,
		Identity: vNextPublicationRecordIdentity(state.preparedIdentity),
		Digest:   state.preparedDigest,
	}
}

func (state *vNextPublicationControlRepairState) latestReference() vNextPublicationControlRepairPhaseReference {
	if len(state.phases) == 0 {
		return state.preparedReference()
	}
	phase := state.phases[len(state.phases)-1]
	return vNextPublicationControlRepairPhaseReference{
		Name:     phase.name,
		Identity: vNextPublicationRecordIdentity(phase.identity),
		Digest:   phase.digest,
	}
}

func (state *vNextPublicationControlRepairState) latestPhase() string {
	if len(state.phases) == 0 {
		return ""
	}
	return state.phases[len(state.phases)-1].record.State
}

func (state *vNextPublicationControlRepairState) latestControl() *vNextPublicationRecordedIdentity {
	if len(state.phases) == 0 {
		return nil
	}
	return state.phases[len(state.phases)-1].record.Control
}

func (state *vNextPublicationControlRepairState) assertPrivateIdentity(operation *vNextPublicationOperation) error {
	if state == nil || state.transaction == nil || state.authorityCleared {
		return fs.ErrClosed
	}
	if err := operation.assertLockBound(); err != nil {
		return err
	}
	if err := operation.connector.assertIdentity(state.transactionName, "publication control repair transaction", state.transactionIdentity); err != nil {
		return err
	}
	actualTransactionIdentity, err := vNextPublicationIdentityFromFile(state.transaction.file, "publication control repair transaction")
	if err != nil {
		return err
	}
	if actualTransactionIdentity != state.transactionIdentity {
		return fmt.Errorf("publication control repair transaction identity changed")
	}
	return state.transaction.assertIdentity(vNextPublicationControlRepairPreparedFile, "publication control repair prepared authority", state.preparedIdentity)
}

func (p *vNextGenerationPublisher) beginControlRepairLocked(operation *vNextPublicationOperation, target string, expected vNextPublicationIdentity) (*vNextPublicationControlRepairState, error) {
	if err := operation.assertLockBound(); err != nil {
		return nil, err
	}
	transactionName, transaction, transactionIdentity, err := vNextPublicationCreateControlRepairTransaction(operation.connector)
	if err != nil {
		return nil, err
	}
	keepTransaction := false
	defer func() {
		if !keepTransaction {
			_ = transaction.Close()
			_ = operation.connector.removeEmptyDirectoryBound(transactionName, "publication control repair transaction", transactionIdentity)
		}
	}()

	prior, hasPrior, err := vNextPublicationBackupControl(operation, transaction, target)
	if err != nil {
		return nil, err
	}
	if err := transaction.Sync(); err != nil {
		return nil, fmt.Errorf("sync publication control repair backup: %w", err)
	}
	if err := p.hit(vNextPublicationAfterControlRepairBackupSync); err != nil {
		return nil, err
	}

	repair := vNextPublicationControlRepair{
		Version:             vNextPublicationControlRepairVersion,
		Target:              target,
		Transaction:         transactionName,
		TransactionIdentity: vNextPublicationRecordIdentity(transactionIdentity),
		Expected:            vNextPublicationRecordIdentity(expected),
	}
	if hasPrior {
		priorRecord := vNextPublicationRecordIdentity(prior)
		repair.Prior = &priorRecord
	}
	if err := vNextPublicationControlRepairValid(repair); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(repair)
	if err != nil {
		return nil, err
	}
	payload = append(payload, '\n')
	preparedIdentity, err := vNextPublicationWriteControlRepairRecord(transaction, vNextPublicationControlRepairPreparedFile, "publication control repair prepared authority", payload)
	if err != nil {
		return nil, err
	}
	if err := transaction.Sync(); err != nil {
		return nil, fmt.Errorf("sync publication control repair prepared authority: %w", err)
	}
	if err := operation.connector.Sync(); err != nil {
		return nil, fmt.Errorf("sync publication control repair transaction: %w", err)
	}
	state := &vNextPublicationControlRepairState{
		record:              repair,
		preparedIdentity:    preparedIdentity,
		preparedDigest:      vNextPublicationDigest(payload),
		transactionName:     transactionName,
		transaction:         transaction,
		transactionIdentity: transactionIdentity,
	}
	keepTransaction = true
	if err := p.hit(vNextPublicationAfterControlRepairPrepared); err != nil {
		return state, err
	}
	return state, nil
}

func (p *vNextGenerationPublisher) appendControlRepairPhaseLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, phaseState string, control *vNextPublicationRecordedIdentity) error {
	if err := state.assertPrivateIdentity(operation); err != nil {
		return err
	}
	previousState := state.latestPhase()
	if !vNextPublicationControlRepairPhaseTransitionValid(previousState, phaseState) {
		return fmt.Errorf("invalid publication control repair phase transition %q to %q", previousState, phaseState)
	}
	sequence := len(state.phases) + 1
	if sequence > vNextPublicationControlRepairMaxPhases {
		return fmt.Errorf("publication control repair phase sequence exhausted")
	}
	phase := vNextPublicationControlRepairPhase{
		Version:          vNextPublicationControlRepairVersion,
		Sequence:         sequence,
		State:            phaseState,
		PreparedDigest:   state.preparedDigest,
		PreparedIdentity: vNextPublicationRecordIdentity(state.preparedIdentity),
		Previous:         state.latestReference(),
		Control:          control,
	}
	if err := vNextPublicationControlRepairPhaseValid(phase); err != nil {
		return err
	}
	payload, err := json.Marshal(phase)
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	name := vNextPublicationControlRepairPhaseName(sequence)
	identity, err := vNextPublicationWriteControlRepairRecord(state.transaction, name, "publication control repair phase", payload)
	if err != nil {
		return err
	}
	if err := state.transaction.Sync(); err != nil {
		return fmt.Errorf("sync publication control repair phase: %w", err)
	}
	state.phases = append(state.phases, vNextPublicationControlRepairPhaseState{
		record:   phase,
		name:     name,
		identity: identity,
		digest:   vNextPublicationDigest(payload),
	})
	return nil
}

func (p *vNextGenerationPublisher) readControlRepairLocked(operation *vNextPublicationOperation) (*vNextPublicationControlRepairState, bool, error) {
	if err := operation.assertLockBound(); err != nil {
		return nil, false, err
	}
	entries, err := operation.connector.readDir()
	if err != nil {
		return nil, false, err
	}
	var recovered *vNextPublicationControlRepairState
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, vNextPublicationControlRepairDirectoryPrefix) {
			continue
		}
		if !vNextPublicationControlRepairTransactionNameValid(name) {
			return nil, false, fmt.Errorf("invalid publication control repair transaction entry %q", name)
		}
		transactionIdentity, err := operation.connector.identityAt(name, "publication control repair transaction")
		if err != nil {
			return nil, false, err
		}
		if transactionIdentity.mode != unix.S_IFDIR {
			return nil, false, fmt.Errorf("publication control repair transaction is not a directory")
		}
		transaction, err := operation.connector.openDirectory(name, "publication control repair transaction")
		if err != nil {
			return nil, false, err
		}
		actualTransactionIdentity, err := vNextPublicationIdentityFromFile(transaction.file, "publication control repair transaction")
		if err != nil {
			_ = transaction.Close()
			return nil, false, err
		}
		if actualTransactionIdentity != transactionIdentity {
			_ = transaction.Close()
			return nil, false, fmt.Errorf("publication control repair transaction identity changed")
		}
		payload, found, preparedIdentity, err := vNextPublicationReadControlBound(transaction, vNextPublicationControlRepairPreparedFile, "publication control repair prepared authority")
		if err != nil {
			_ = transaction.Close()
			return nil, false, err
		}
		if !found {
			_ = transaction.Close()
			continue
		}
		var repair vNextPublicationControlRepair
		if err := vNextPublicationDecode(payload, &repair); err != nil {
			_ = transaction.Close()
			return nil, false, fmt.Errorf("decode publication control repair prepared authority: %w", err)
		}
		if err := vNextPublicationControlRepairValid(repair); err != nil {
			_ = transaction.Close()
			return nil, false, err
		}
		recordedTransactionIdentity, err := repair.TransactionIdentity.publicationIdentity(unix.S_IFDIR)
		if err != nil {
			_ = transaction.Close()
			return nil, false, err
		}
		if repair.Transaction != name || recordedTransactionIdentity != transactionIdentity {
			_ = transaction.Close()
			return nil, false, fmt.Errorf("publication control repair prepared authority does not bind its transaction")
		}
		state := &vNextPublicationControlRepairState{
			record:              repair,
			preparedIdentity:    preparedIdentity,
			preparedDigest:      vNextPublicationDigest(payload),
			transactionName:     name,
			transaction:         transaction,
			transactionIdentity: transactionIdentity,
		}
		if err := p.readControlRepairPhasesLocked(operation, state); err != nil {
			state.close()
			return nil, false, err
		}
		if recovered != nil {
			state.close()
			return nil, false, fmt.Errorf("multiple publication control repair authorities remain")
		}
		recovered = state
	}
	if recovered == nil {
		return nil, false, nil
	}
	return recovered, true, nil
}

func (p *vNextGenerationPublisher) readControlRepairPhasesLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState) error {
	if err := state.assertPrivateIdentity(operation); err != nil {
		return err
	}
	entries, err := state.transaction.readDir()
	if err != nil {
		return err
	}
	allowed := map[string]struct{}{
		vNextPublicationControlRepairPreparedFile: {},
		vNextPublicationControlBackupMember:       {},
		vNextPublicationControlReplacementMember:  {},
	}
	for sequence := 1; sequence <= vNextPublicationControlRepairMaxPhases; sequence++ {
		allowed[vNextPublicationControlRepairPhaseName(sequence)] = struct{}{}
	}
	for _, entry := range entries {
		if _, ok := allowed[entry.Name()]; !ok {
			return fmt.Errorf("unexpected publication control repair transaction member %q", entry.Name())
		}
	}

	previous := state.preparedReference()
	previousState := ""
	missing := false
	for sequence := 1; sequence <= vNextPublicationControlRepairMaxPhases; sequence++ {
		name := vNextPublicationControlRepairPhaseName(sequence)
		payload, found, identity, err := vNextPublicationReadControlBound(state.transaction, name, "publication control repair phase")
		if err != nil {
			return err
		}
		if !found {
			missing = true
			continue
		}
		if missing {
			return fmt.Errorf("publication control repair phase chain has a gap before %q", name)
		}
		var phase vNextPublicationControlRepairPhase
		if err := vNextPublicationDecode(payload, &phase); err != nil {
			return fmt.Errorf("decode publication control repair phase: %w", err)
		}
		if err := vNextPublicationControlRepairPhaseValid(phase); err != nil {
			return err
		}
		if phase.Sequence != sequence || phase.PreparedDigest != state.preparedDigest || phase.PreparedIdentity != vNextPublicationRecordIdentity(state.preparedIdentity) || phase.Previous != previous {
			return fmt.Errorf("publication control repair phase %q is not bound to its prepared predecessor", name)
		}
		if !vNextPublicationControlRepairPhaseTransitionValid(previousState, phase.State) {
			return fmt.Errorf("invalid publication control repair phase chain transition %q to %q", previousState, phase.State)
		}
		if err := vNextPublicationValidateControlRepairPhaseSemanticLocked(state, phase); err != nil {
			return err
		}
		phaseState := vNextPublicationControlRepairPhaseState{
			record:   phase,
			name:     name,
			identity: identity,
			digest:   vNextPublicationDigest(payload),
		}
		state.phases = append(state.phases, phaseState)
		previous = vNextPublicationControlRepairPhaseReference{
			Name:     name,
			Identity: vNextPublicationRecordIdentity(identity),
			Digest:   phaseState.digest,
		}
		previousState = phase.State
	}
	return nil
}

func vNextPublicationValidateControlRepairPhaseSemanticLocked(state *vNextPublicationControlRepairState, phase vNextPublicationControlRepairPhase) error {
	expected := state.record.Expected
	prior, hasPrior, err := vNextPublicationRepairPrior(state)
	if err != nil {
		return err
	}
	switch phase.State {
	case vNextPublicationControlRepairInstalled:
		if phase.Control == nil || *phase.Control != expected {
			return fmt.Errorf("publication control repair installed phase does not bind the expected control")
		}
	case vNextPublicationControlRepairReplacementRetained:
		if phase.Control == nil {
			return fmt.Errorf("publication control repair replacement phase has no replacement")
		}
		replacement, err := phase.Control.publicationIdentity(unix.S_IFREG)
		if err != nil {
			return err
		}
		if err := state.transaction.assertIdentity(vNextPublicationControlReplacementMember, "publication control repair replacement", replacement); err != nil {
			return err
		}
	case vNextPublicationControlRepairRestored:
		if hasPrior {
			if phase.Control == nil || *phase.Control != vNextPublicationRecordIdentity(prior) {
				return fmt.Errorf("publication control repair restored phase does not bind the prior control")
			}
		} else if phase.Control != nil {
			return fmt.Errorf("publication control repair restored no-prior phase retains a control identity")
		}
	}
	return nil
}

func (p *vNextGenerationPublisher) recoverControlRepairLocked(operation *vNextPublicationOperation) error {
	state, found, err := p.readControlRepairLocked(operation)
	if err != nil || !found {
		return err
	}
	defer state.close()
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

func vNextPublicationRepairExpected(state *vNextPublicationControlRepairState) (vNextPublicationIdentity, error) {
	return state.record.Expected.publicationIdentity(unix.S_IFREG)
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

func vNextPublicationRepairTargetMatchesDesired(actual vNextPublicationIdentity, found bool, prior vNextPublicationIdentity, hasPrior bool) bool {
	if hasPrior {
		return found && actual == prior
	}
	return !found
}

func (p *vNextGenerationPublisher) retainControlRepairReplacementLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, actual vNextPublicationIdentity, injectFaults bool) error {
	if err := state.assertPrivateIdentity(operation); err != nil {
		return err
	}
	replacement, err := state.transaction.identityAt(vNextPublicationControlReplacementMember, "publication control repair replacement")
	if err == nil {
		if replacement != actual {
			return fmt.Errorf("publication control repair observed a second replacement")
		}
	} else if errors.Is(err, fs.ErrNotExist) {
		if err := state.transaction.linkFromBound(operation.connector, state.record.Target, vNextPublicationControlReplacementMember, "publication control repair replacement", actual); err != nil {
			return err
		}
		if err := state.transaction.Sync(); err != nil {
			return fmt.Errorf("sync publication control repair replacement: %w", err)
		}
		if injectFaults {
			if err := p.hit(vNextPublicationAfterControlRepairReplacementRetainSync); err != nil {
				return err
			}
		}
	} else {
		return err
	}
	recorded := vNextPublicationRecordIdentity(actual)
	if state.latestPhase() != vNextPublicationControlRepairReplacementRetained {
		if err := p.appendControlRepairPhaseLocked(operation, state, vNextPublicationControlRepairReplacementRetained, &recorded); err != nil {
			return err
		}
	}
	if injectFaults {
		if err := p.hit(vNextPublicationAfterControlRepairReplacementSync); err != nil {
			return err
		}
	}
	return nil
}

func (p *vNextGenerationPublisher) restoreControlRepairLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, actual vNextPublicationIdentity, found bool, prior vNextPublicationIdentity, hasPrior bool) error {
	if err := state.assertPrivateIdentity(operation); err != nil {
		return err
	}
	if hasPrior {
		if found && actual == prior {
			return nil
		}
		if err := state.transaction.assertIdentity(vNextPublicationControlBackupMember, "publication control repair backup", prior); err != nil {
			return err
		}
		if found {
			if err := operation.connector.assertIdentity(state.record.Target, "publication control repair target", actual); err != nil {
				return err
			}
			if err := operation.connector.renameFrom(state.transaction, vNextPublicationControlBackupMember, state.record.Target); err != nil {
				return err
			}
		} else if err := operation.connector.linkFromBound(state.transaction, vNextPublicationControlBackupMember, state.record.Target, "publication control repair backup", prior); err != nil {
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

func (p *vNextGenerationPublisher) appendRestoredControlRepairPhaseLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, prior vNextPublicationIdentity, hasPrior bool, injectFaults bool) error {
	if state.latestPhase() != vNextPublicationControlRepairRestored {
		var control *vNextPublicationRecordedIdentity
		if hasPrior {
			recorded := vNextPublicationRecordIdentity(prior)
			control = &recorded
		}
		if err := p.appendControlRepairPhaseLocked(operation, state, vNextPublicationControlRepairRestored, control); err != nil {
			return err
		}
	}
	if injectFaults {
		if err := p.hit(vNextPublicationAfterControlRepairRestoreSync); err != nil {
			return err
		}
	}
	return nil
}

func (p *vNextGenerationPublisher) restoreAndRecordControlRepairLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, actual vNextPublicationIdentity, found bool, prior vNextPublicationIdentity, hasPrior bool, injectFaults bool) error {
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
	if err := p.appendRestoredControlRepairPhaseLocked(operation, state, prior, hasPrior, injectFaults); err != nil {
		return err
	}
	return p.finishControlRepairLocked(operation, state, injectFaults)
}

func (p *vNextGenerationPublisher) resolveUnexpectedControlRepairLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, actual vNextPublicationIdentity, found bool, prior vNextPublicationIdentity, hasPrior bool, injectFaults bool) error {
	if found {
		if err := p.retainControlRepairReplacementLocked(operation, state, actual, injectFaults); err != nil {
			return err
		}
	}
	return p.restoreAndRecordControlRepairLocked(operation, state, actual, found, prior, hasPrior, injectFaults)
}

func (p *vNextGenerationPublisher) resolveControlRepairLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, injectFaults bool) error {
	if err := state.assertPrivateIdentity(operation); err != nil {
		return err
	}
	prior, hasPrior, err := vNextPublicationRepairPrior(state)
	if err != nil {
		return err
	}
	expected, err := vNextPublicationRepairExpected(state)
	if err != nil {
		return err
	}
	actual, found, err := vNextPublicationRepairTargetIdentity(operation.connector, state.record.Target)
	if err != nil {
		return err
	}
	desired := vNextPublicationRepairTargetMatchesDesired(actual, found, prior, hasPrior)

	switch state.latestPhase() {
	case "":
		if found && actual == expected {
			recorded := vNextPublicationRecordIdentity(expected)
			if err := p.appendControlRepairPhaseLocked(operation, state, vNextPublicationControlRepairInstalled, &recorded); err != nil {
				return err
			}
			if injectFaults {
				if err := p.hit(vNextPublicationAfterControlRepairInstalledPhaseSync); err != nil {
					return err
				}
			}
			return p.finishControlRepairLocked(operation, state, injectFaults)
		}
		if desired {
			return p.restoreAndRecordControlRepairLocked(operation, state, actual, found, prior, hasPrior, injectFaults)
		}
		return p.resolveUnexpectedControlRepairLocked(operation, state, actual, found, prior, hasPrior, injectFaults)

	case vNextPublicationControlRepairInstalled:
		if found && actual == expected {
			return p.finishControlRepairLocked(operation, state, injectFaults)
		}
		if desired {
			return p.restoreAndRecordControlRepairLocked(operation, state, actual, found, prior, hasPrior, injectFaults)
		}
		return p.resolveUnexpectedControlRepairLocked(operation, state, actual, found, prior, hasPrior, injectFaults)

	case vNextPublicationControlRepairReplacementRetained:
		control := state.latestControl()
		if control == nil {
			return fmt.Errorf("publication control repair replacement phase has no control")
		}
		replacement, err := control.publicationIdentity(unix.S_IFREG)
		if err != nil {
			return err
		}
		if desired {
			return p.restoreAndRecordControlRepairLocked(operation, state, actual, found, prior, hasPrior, injectFaults)
		}
		if !found || actual != replacement {
			return fmt.Errorf("publication control repair observed a second replacement")
		}
		return p.restoreAndRecordControlRepairLocked(operation, state, actual, found, prior, hasPrior, injectFaults)

	case vNextPublicationControlRepairRestored:
		if !desired {
			return fmt.Errorf("publication control changed after restored repair phase")
		}
		return p.finishControlRepairLocked(operation, state, injectFaults)
	default:
		return fmt.Errorf("invalid publication control repair phase %q", state.latestPhase())
	}
}

func (p *vNextGenerationPublisher) finishControlRepairLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, injectFaults bool) error {
	if err := p.clearControlRepairLocked(operation, state); err != nil {
		return err
	}
	if injectFaults {
		if err := p.hit(vNextPublicationAfterControlRepairClearSync); err != nil {
			return err
		}
	}
	return nil
}

func (p *vNextGenerationPublisher) clearControlRepairLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState) error {
	if err := state.assertPrivateIdentity(operation); err != nil {
		return err
	}
	prior, hasPrior, err := vNextPublicationRepairPrior(state)
	if err != nil {
		return err
	}
	expected, err := vNextPublicationRepairExpected(state)
	if err != nil {
		return err
	}
	actual, found, err := vNextPublicationRepairTargetIdentity(operation.connector, state.record.Target)
	if err != nil {
		return err
	}
	switch state.latestPhase() {
	case vNextPublicationControlRepairInstalled:
		if !found || actual != expected {
			return fmt.Errorf("publication control changed before installed repair cleanup")
		}
	case vNextPublicationControlRepairRestored:
		if !vNextPublicationRepairTargetMatchesDesired(actual, found, prior, hasPrior) {
			return fmt.Errorf("publication control changed before restored repair cleanup")
		}
	default:
		return fmt.Errorf("publication control repair cannot clear phase %q", state.latestPhase())
	}

	for _, phase := range state.phases {
		if phase.record.State != vNextPublicationControlRepairReplacementRetained || phase.record.Control == nil {
			continue
		}
		replacement, err := phase.record.Control.publicationIdentity(unix.S_IFREG)
		if err != nil {
			return err
		}
		if err := state.transaction.assertIdentity(vNextPublicationControlReplacementMember, "publication control repair replacement", replacement); err != nil {
			return err
		}
	}
	if err := p.hit(vNextPublicationBeforeControlRepairAuthorityClear); err != nil {
		return err
	}
	if err := state.assertPrivateIdentity(operation); err != nil {
		return err
	}
	actual, found, err = vNextPublicationRepairTargetIdentity(operation.connector, state.record.Target)
	if err != nil {
		return err
	}
	switch state.latestPhase() {
	case vNextPublicationControlRepairInstalled:
		if !found || actual != expected {
			return fmt.Errorf("publication control changed before final installed repair cleanup")
		}
	case vNextPublicationControlRepairRestored:
		if !vNextPublicationRepairTargetMatchesDesired(actual, found, prior, hasPrior) {
			return fmt.Errorf("publication control changed before final restored repair cleanup")
		}
	}

	// Retire the sole recovery authority before deleting phase or backup data:
	// a crash after this sync leaves a valid public control and only private
	// garbage, never a pending authority whose restoration material is gone.
	if err := state.transaction.removeRegularBound(vNextPublicationControlRepairPreparedFile, "publication control repair prepared authority", state.preparedIdentity); err != nil {
		return err
	}
	if err := state.transaction.Sync(); err != nil {
		return fmt.Errorf("sync cleared publication control repair authority: %w", err)
	}
	state.authorityCleared = true
	if err := p.hit(vNextPublicationAfterControlRepairAuthorityRetireSync); err != nil {
		return err
	}

	for _, phase := range state.phases {
		if err := state.transaction.removeRegularBound(phase.name, "publication control repair phase", phase.identity); err != nil {
			return err
		}
	}
	if hasPrior {
		backup, backupErr := state.transaction.identityAt(vNextPublicationControlBackupMember, "publication control repair backup")
		if backupErr == nil {
			if backup != prior {
				return fmt.Errorf("publication control repair backup identity changed")
			}
			if err := state.transaction.removeRegularBound(vNextPublicationControlBackupMember, "publication control repair backup", prior); err != nil {
				return err
			}
		} else if !errors.Is(backupErr, fs.ErrNotExist) {
			return backupErr
		}
	}
	if err := state.transaction.Sync(); err != nil {
		return fmt.Errorf("sync cleaned publication control repair transaction: %w", err)
	}

	entries, err := state.transaction.readDir()
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return nil
	}
	if err := state.transaction.Close(); err != nil {
		return err
	}
	state.transaction = nil
	if err := operation.connector.removeEmptyDirectoryBound(state.transactionName, "publication control repair transaction", state.transactionIdentity); err != nil {
		return err
	}
	return operation.connector.Sync()
}

func (p *vNextGenerationPublisher) refuseMismatchedControlInstallLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState) error {
	cause := fmt.Errorf("publication control identity changed after namespace transition")
	if err := p.resolveControlRepairLocked(operation, state, true); err != nil {
		return fmt.Errorf("%w; durable control repair: %w", cause, err)
	}
	return cause
}

func (p *vNextGenerationPublisher) completeControlRepairLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState) error {
	return p.resolveControlRepairLocked(operation, state, true)
}
