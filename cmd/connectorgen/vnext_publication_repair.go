package main

import (
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"golang.org/x/sys/unix"
)

const (
	vNextPublicationControlRepairVersion = 3

	vNextPublicationControlAuthorityMarkerFile      = ".connectorgen-control-authority-v3.json"
	vNextPublicationControlRepairDirectoryPrefix    = ".connectorgen-control-repair-"
	vNextPublicationControlRepairPreparedFile       = "prepared.json"
	vNextPublicationControlRepairPhasePrefix        = "phase-"
	vNextPublicationControlRepairPhaseSuffix        = ".json"
	vNextPublicationControlRepairMaxCaptureAttempts = 4
	vNextPublicationControlRepairMaxPhases          = vNextPublicationControlRepairMaxCaptureAttempts*2 + 2

	vNextPublicationControlCapturePrefix = "capture-"

	vNextPublicationControlRepairCaptureIntent = "capture_intent"
	vNextPublicationControlRepairCaptured      = "captured"
	vNextPublicationControlRepairSelected      = "selected"
	vNextPublicationControlRepairTerminal      = "terminal"

	vNextPublicationControlRepairCommitted     = "committed"
	vNextPublicationControlRepairRolledBack    = "rolled_back"
	vNextPublicationControlRepairRetryRequired = "retry_required"
)

var errVNextPublicationControlConflict = errors.New("publication control transition requires retry")

// vNextPublicationControlConflictError means that the publisher retained every
// object it observed but could not select an old-or-new public control after a
// bounded number of no-replace captures.
type vNextPublicationControlConflictError struct {
	target string
}

func (err *vNextPublicationControlConflictError) Error() string {
	return fmt.Sprintf("%s: %s", errVNextPublicationControlConflict, err.target)
}

func (err *vNextPublicationControlConflictError) Unwrap() error {
	return errVNextPublicationControlConflict
}

// vNextPublicationRecordedIdentity is the stable on-disk binding for one
// descriptor-relative publication member.
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
	return vNextPublicationIdentity{device: identity.Device, inode: identity.Inode, mode: identity.Mode}, nil
}

// vNextPublicationControlState is deliberately a tagged union. An absent
// public control is not represented by a synthetic zero inode.
type vNextPublicationControlState struct {
	Present  bool                              `json:"present"`
	Identity *vNextPublicationRecordedIdentity `json:"identity,omitempty"`
	Member   string                            `json:"member,omitempty"`
}

func vNextPublicationAbsentControlState() vNextPublicationControlState {
	return vNextPublicationControlState{}
}

func vNextPublicationPresentControlState(member string, identity vNextPublicationIdentity) vNextPublicationControlState {
	recorded := vNextPublicationRecordIdentity(identity)
	return vNextPublicationControlState{Present: true, Identity: &recorded, Member: member}
}

func (state vNextPublicationControlState) identity(expectedMember string) (vNextPublicationIdentity, bool, error) {
	if !state.Present {
		if state.Identity != nil || state.Member != "" {
			return vNextPublicationIdentity{}, false, fmt.Errorf("absent publication control state retains an identity")
		}
		return vNextPublicationIdentity{}, false, nil
	}
	if state.Identity == nil || state.Member != expectedMember {
		return vNextPublicationIdentity{}, false, fmt.Errorf("invalid publication control state member %q", state.Member)
	}
	identity, err := state.Identity.publicationIdentity(unix.S_IFREG)
	if err != nil {
		return vNextPublicationIdentity{}, false, err
	}
	return identity, true, nil
}

func (state vNextPublicationControlState) sameLogical(other vNextPublicationControlState) bool {
	if state.Present != other.Present {
		return false
	}
	if !state.Present {
		return true
	}
	return state.Identity != nil && other.Identity != nil && *state.Identity == *other.Identity
}

func vNextPublicationControlStateWithMember(state vNextPublicationControlState, member string) vNextPublicationControlState {
	if !state.Present {
		return vNextPublicationAbsentControlState()
	}
	return vNextPublicationControlState{Present: true, Identity: state.Identity, Member: member}
}

type vNextPublicationControlAuthorityMarker struct {
	Version int      `json:"version"`
	Targets []string `json:"targets"`
}

type vNextPublicationControlRepairPredecessor struct {
	Transaction         string                                      `json:"transaction"`
	TransactionIdentity vNextPublicationRecordedIdentity            `json:"transaction_identity"`
	Terminal            vNextPublicationControlRepairPhaseReference `json:"terminal"`
	Selected            vNextPublicationControlState                `json:"selected"`
}

// vNextPublicationControlRepair is immutable transaction authority. Its
// predecessor binds the prior terminal, so an unrelated transaction cannot
// become a second head merely by being present in the connector directory.
type vNextPublicationControlRepair struct {
	Version             int                                       `json:"version"`
	Target              string                                    `json:"target"`
	Transaction         string                                    `json:"transaction"`
	TransactionIdentity vNextPublicationRecordedIdentity          `json:"transaction_identity"`
	Predecessor         *vNextPublicationControlRepairPredecessor `json:"predecessor,omitempty"`
	Prior               vNextPublicationControlState              `json:"prior"`
	Intended            vNextPublicationControlState              `json:"intended"`
	MaxCaptureAttempts  int                                       `json:"max_capture_attempts"`
}

type vNextPublicationControlRepairCapture struct {
	Attempt  int                              `json:"attempt"`
	Name     string                           `json:"name"`
	Identity vNextPublicationRecordedIdentity `json:"identity"`
}

// vNextPublicationControlRepairPhase remains append-only. Every phase binds
// both the immutable prepared record and the immediately preceding record.
type vNextPublicationControlRepairPhase struct {
	Version          int                                         `json:"version"`
	Sequence         int                                         `json:"sequence"`
	State            string                                      `json:"state"`
	PreparedDigest   string                                      `json:"prepared_digest"`
	PreparedIdentity vNextPublicationRecordedIdentity            `json:"prepared_identity"`
	Previous         vNextPublicationControlRepairPhaseReference `json:"previous"`
	Capture          *vNextPublicationControlRepairCapture       `json:"capture,omitempty"`
	Candidate        *vNextPublicationControlState               `json:"candidate,omitempty"`
	Selected         *vNextPublicationControlState               `json:"selected,omitempty"`
	Outcome          string                                      `json:"outcome,omitempty"`
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
	transactionIdentity vNextPublicationIdentity
	phases              []vNextPublicationControlRepairPhaseState
}

type vNextPublicationControlAuthorityHead struct {
	state    *vNextPublicationControlRepairState
	terminal vNextPublicationControlRepairPhaseState
	selected vNextPublicationControlState
	outcome  string
}

type vNextPublicationControlAuthorityGraph struct {
	marker         bool
	markerIdentity vNextPublicationIdentity
	states         map[string]*vNextPublicationControlRepairState
	heads          map[string]*vNextPublicationControlAuthorityHead
}

func (graph *vNextPublicationControlAuthorityGraph) assertPrivateIdentity(operation *vNextPublicationOperation) error {
	if graph == nil {
		return fs.ErrClosed
	}
	if err := operation.assertLockBound(); err != nil {
		return err
	}
	if !graph.marker {
		return nil
	}
	if err := operation.connector.assertIdentity(vNextPublicationControlAuthorityMarkerFile, "publication control authority marker", graph.markerIdentity); err != nil {
		return err
	}
	for _, state := range graph.states {
		if err := state.assertPrivateIdentity(operation); err != nil {
			return err
		}
	}
	return nil
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

func vNextPublicationControlRepairPhaseName(sequence int) string {
	return fmt.Sprintf("%s%04d%s", vNextPublicationControlRepairPhasePrefix, sequence, vNextPublicationControlRepairPhaseSuffix)
}

func vNextPublicationControlCaptureName(attempt int) string {
	return fmt.Sprintf("%s%04d", vNextPublicationControlCapturePrefix, attempt)
}

func vNextPublicationControlCaptureNameValid(name string) bool {
	if !vNextPublicationDirectNameValid(name) || !strings.HasPrefix(name, vNextPublicationControlCapturePrefix) {
		return false
	}
	var attempt int
	if _, err := fmt.Sscanf(name, vNextPublicationControlCapturePrefix+"%04d", &attempt); err != nil {
		return false
	}
	return attempt >= 1 && attempt <= vNextPublicationControlRepairMaxCaptureAttempts && name == vNextPublicationControlCaptureName(attempt)
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

func vNextPublicationControlAuthorityMarkerValid(marker vNextPublicationControlAuthorityMarker) error {
	if marker.Version != vNextPublicationControlRepairVersion || len(marker.Targets) != 2 {
		return fmt.Errorf("invalid publication control authority marker")
	}
	seen := make(map[string]struct{}, len(marker.Targets))
	for _, target := range marker.Targets {
		if !vNextPublicationControlRepairTargetValid(target) {
			return fmt.Errorf("invalid publication control authority target %q", target)
		}
		if _, duplicate := seen[target]; duplicate {
			return fmt.Errorf("duplicate publication control authority target %q", target)
		}
		seen[target] = struct{}{}
	}
	if _, ok := seen[vNextPublicationCurrentFile]; !ok {
		return fmt.Errorf("publication control authority marker omits CURRENT")
	}
	if _, ok := seen[vNextPublicationJournalFile]; !ok {
		return fmt.Errorf("publication control authority marker omits JOURNAL")
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
	if _, _, err := repair.Prior.identity(vNextPublicationControlBackupMember); err != nil {
		return fmt.Errorf("invalid publication control repair prior: %w", err)
	}
	if _, _, err := repair.Intended.identity(vNextPublicationControlReplacementMember); err != nil {
		return fmt.Errorf("invalid publication control repair intended: %w", err)
	}
	if repair.MaxCaptureAttempts != vNextPublicationControlRepairMaxCaptureAttempts {
		return fmt.Errorf("invalid publication control repair capture limit %d", repair.MaxCaptureAttempts)
	}
	if repair.Predecessor == nil {
		return nil
	}
	predecessor := repair.Predecessor
	if !vNextPublicationControlRepairTransactionNameValid(predecessor.Transaction) {
		return fmt.Errorf("invalid publication control repair predecessor transaction %q", predecessor.Transaction)
	}
	if _, err := predecessor.TransactionIdentity.publicationIdentity(unix.S_IFDIR); err != nil {
		return fmt.Errorf("invalid publication control repair predecessor transaction identity: %w", err)
	}
	if err := vNextPublicationControlRepairPhaseReferenceValid(predecessor.Terminal); err != nil {
		return fmt.Errorf("invalid publication control repair predecessor terminal: %w", err)
	}
	if predecessor.Selected.Present {
		if _, _, err := predecessor.Selected.identity(predecessor.Selected.Member); err != nil {
			return fmt.Errorf("invalid publication control repair predecessor selection: %w", err)
		}
	} else if _, _, err := predecessor.Selected.identity(""); err != nil {
		return fmt.Errorf("invalid publication control repair predecessor absence: %w", err)
	}
	return nil
}

func vNextPublicationControlRepairCaptureValid(capture vNextPublicationControlRepairCapture) error {
	if capture.Attempt < 1 || capture.Attempt > vNextPublicationControlRepairMaxCaptureAttempts || capture.Name != vNextPublicationControlCaptureName(capture.Attempt) {
		return fmt.Errorf("invalid publication control capture %q", capture.Name)
	}
	if _, err := capture.Identity.publicationIdentity(unix.S_IFDIR); err != nil {
		return fmt.Errorf("invalid publication control capture identity: %w", err)
	}
	return nil
}

func vNextPublicationControlRepairPhaseStateValid(state string) bool {
	switch state {
	case vNextPublicationControlRepairCaptureIntent,
		vNextPublicationControlRepairCaptured,
		vNextPublicationControlRepairSelected,
		vNextPublicationControlRepairTerminal:
		return true
	default:
		return false
	}
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
	case vNextPublicationControlRepairCaptureIntent:
		if phase.Capture == nil || phase.Candidate != nil || phase.Selected != nil || phase.Outcome != "" {
			return fmt.Errorf("invalid publication capture intent phase")
		}
		return vNextPublicationControlRepairCaptureValid(*phase.Capture)
	case vNextPublicationControlRepairCaptured:
		if phase.Capture == nil || phase.Candidate == nil || phase.Selected != nil || phase.Outcome != "" {
			return fmt.Errorf("invalid publication captured phase")
		}
		if err := vNextPublicationControlRepairCaptureValid(*phase.Capture); err != nil {
			return err
		}
		_, _, err := phase.Candidate.identity(vNextPublicationControlCaptureMember)
		return err
	case vNextPublicationControlRepairSelected:
		if phase.Capture != nil || phase.Candidate != nil || phase.Selected == nil || phase.Outcome != "" {
			return fmt.Errorf("invalid publication selected phase")
		}
		if phase.Selected.Present {
			_, _, err := phase.Selected.identity(phase.Selected.Member)
			return err
		}
		_, _, err := phase.Selected.identity("")
		return err
	case vNextPublicationControlRepairTerminal:
		if phase.Capture != nil || phase.Candidate != nil || phase.Selected == nil {
			return fmt.Errorf("invalid publication terminal phase")
		}
		switch phase.Outcome {
		case vNextPublicationControlRepairCommitted, vNextPublicationControlRepairRolledBack, vNextPublicationControlRepairRetryRequired:
		default:
			return fmt.Errorf("invalid publication terminal outcome %q", phase.Outcome)
		}
		if phase.Selected.Present {
			_, _, err := phase.Selected.identity(phase.Selected.Member)
			return err
		}
		_, _, err := phase.Selected.identity("")
		return err
	}
	return nil
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
		return "", nil, vNextPublicationIdentity{}, err
	}
	return "", nil, vNextPublicationIdentity{}, fmt.Errorf("create publication control repair transaction: exhausted unique names")
}

type vNextPublicationRecordDisposition uint8

const (
	vNextPublicationRecordUnknown vNextPublicationRecordDisposition = iota
	vNextPublicationRecordNotCreated
	vNextPublicationRecordRemoved
	vNextPublicationRecordRetainedIncomplete
	vNextPublicationRecordRetainedComplete
)

type vNextPublicationRecordResult struct {
	created         bool
	identity        vNextPublicationIdentity
	contentComplete bool
	disposition     vNextPublicationRecordDisposition
}

func vNextPublicationWriteControlRepairRecord(directory *vNextPublicationDirectory, name, label string, payload []byte, hooks vNextPublicationControlRecordHooks) (result vNextPublicationRecordResult, err error) {
	opened := directory.openFileResult(name, label, unix.O_CREAT|unix.O_EXCL|unix.O_WRONLY, 0o600, false)
	err = errors.Join(opened.openErr, opened.parentCloseErr)
	if !opened.opened {
		result.disposition = vNextPublicationRecordNotCreated
		return result, err
	}
	result.created = true
	file := os.NewFile(uintptr(opened.fd), label)
	if file == nil {
		vNextPublicationRecordError(&err, "adopt "+label, fmt.Errorf("invalid file descriptor"))
		vNextPublicationRecordError(&err, "close opened "+label, vNextPublicationCloseOpenedFileAfterParentClose(opened.fd, label))
		return result, err
	}
	// Install the unique Close owner before identity acquisition or any hook.
	defer func() {
		if hooks.Close != nil {
			vNextPublicationRecordError(&err, "close "+label, hooks.Close(file, label))
		} else {
			vNextPublicationRecordError(&err, "close "+label, file.Close())
		}
		if result.contentComplete {
			if bindErr := directory.assertIdentity(name, label, result.identity); bindErr != nil {
				vNextPublicationRecordError(&err, "bind completed "+label, bindErr)
				result.disposition = vNextPublicationRecordUnknown
			} else {
				result.disposition = vNextPublicationRecordRetainedComplete
			}
		}
	}()
	identity, identityErr := vNextPublicationIdentityFromFile(file, label)
	if identityErr != nil {
		vNextPublicationRecordError(&err, "identify "+label, identityErr)
		return result, err
	}
	result.identity = identity
	result.disposition = vNextPublicationRecordRetainedIncomplete
	// Incomplete disposal happens before Close, while this fd pins our inode.
	defer func() {
		if result.contentComplete {
			return
		}
		removeErr := directory.removeRegularBound(name, label, result.identity)
		vNextPublicationRecordError(&err, "remove incomplete "+label, removeErr)
		if removeErr == nil {
			result.disposition = vNextPublicationRecordRemoved
			vNextPublicationRecordError(&err, "sync removed "+label, directory.Sync())
		}
	}()
	if err != nil {
		return result, err
	}
	write := file.Write
	if hooks.Write != nil {
		write = func(b []byte) (int, error) { return hooks.Write(file, label, b) }
	}
	written, writeErr := write(payload)
	result.contentComplete = written == len(payload)
	if writeErr != nil {
		vNextPublicationRecordError(&err, "write "+label, writeErr)
	}
	if !result.contentComplete {
		vNextPublicationRecordError(&err, "write "+label, fmt.Errorf("wrote %d of %d bytes: %w", written, len(payload), io.ErrShortWrite))
		return result, err
	}
	if err != nil {
		return result, err
	}
	sync := file.Sync
	if hooks.Sync != nil {
		sync = func() error { return hooks.Sync(file, label) }
	}
	vNextPublicationRecordError(&err, "sync "+label, sync())
	vNextPublicationRecordError(&err, "bind "+label, directory.assertIdentity(name, label, result.identity))
	return result, err
}

func (state *vNextPublicationControlRepairState) close() {
	// Authority graph states retain only immutable identity/digest metadata.
	// Traversal descriptors are opened and closed at each dependent operation.
}

func (graph *vNextPublicationControlAuthorityGraph) close() {
	if graph == nil {
		return
	}
	for _, state := range graph.states {
		state.close()
	}
}

func (state *vNextPublicationControlRepairState) preparedReference() vNextPublicationControlRepairPhaseReference {
	return vNextPublicationControlRepairPhaseReference{
		Name: vNextPublicationControlRepairPreparedFile, Identity: vNextPublicationRecordIdentity(state.preparedIdentity), Digest: state.preparedDigest,
	}
}

func (state *vNextPublicationControlRepairState) latestReference() vNextPublicationControlRepairPhaseReference {
	if len(state.phases) == 0 {
		return state.preparedReference()
	}
	phase := state.phases[len(state.phases)-1]
	return vNextPublicationControlRepairPhaseReference{Name: phase.name, Identity: vNextPublicationRecordIdentity(phase.identity), Digest: phase.digest}
}

func (state *vNextPublicationControlRepairState) latestPhase() string {
	if len(state.phases) == 0 {
		return ""
	}
	return state.phases[len(state.phases)-1].record.State
}

func (state *vNextPublicationControlRepairState) terminal() (vNextPublicationControlRepairPhaseState, vNextPublicationControlState, string, bool) {
	if len(state.phases) == 0 {
		return vNextPublicationControlRepairPhaseState{}, vNextPublicationControlState{}, "", false
	}
	phase := state.phases[len(state.phases)-1]
	if phase.record.State != vNextPublicationControlRepairTerminal || phase.record.Selected == nil {
		return vNextPublicationControlRepairPhaseState{}, vNextPublicationControlState{}, "", false
	}
	return phase, *phase.record.Selected, phase.record.Outcome, true
}

func vNextPublicationRecordError(resultErr *error, action string, cause error) {
	if cause == nil {
		return
	}
	wrapped := fmt.Errorf("%s: %w", action, cause)
	if *resultErr == nil {
		*resultErr = wrapped
		return
	}
	*resultErr = errors.Join(*resultErr, wrapped)
}

func vNextPublicationCloseAfter(resultErr *error, closer interface{ Close() error }, label string) {
	vNextPublicationRecordError(resultErr, "close "+label, closer.Close())
}

func (state *vNextPublicationControlRepairState) assertPrivateIdentity(operation *vNextPublicationOperation) (resultErr error) {
	transaction, err := state.openTransaction(operation)
	if err != nil {
		return err
	}
	defer vNextPublicationCloseAfter(&resultErr, transaction, "publication control repair transaction")
	return state.assertPrivateIdentityAt(operation, transaction)
}

func (state *vNextPublicationControlRepairState) openTransaction(operation *vNextPublicationOperation) (*vNextPublicationDirectory, error) {
	if state == nil {
		return nil, fs.ErrClosed
	}
	if err := operation.assertLockBound(); err != nil {
		return nil, err
	}
	if err := operation.connector.assertIdentity(state.transactionName, "publication control repair transaction", state.transactionIdentity); err != nil {
		return nil, err
	}
	transaction, err := operation.connector.openDirectory(state.transactionName, "publication control repair transaction")
	if err != nil {
		return nil, err
	}
	actual, err := vNextPublicationIdentityFromFile(transaction.file, "publication control repair transaction")
	if err != nil {
		_ = transaction.Close()
		return nil, err
	}
	if actual != state.transactionIdentity {
		_ = transaction.Close()
		return nil, fmt.Errorf("publication control repair transaction identity changed")
	}
	return transaction, nil
}

func (state *vNextPublicationControlRepairState) assertPrivateIdentityAt(operation *vNextPublicationOperation, transaction *vNextPublicationDirectory) error {
	if state == nil || transaction == nil {
		return fs.ErrClosed
	}
	prepared, found, identity, err := vNextPublicationReadControlBound(transaction, vNextPublicationControlRepairPreparedFile, "publication control repair prepared authority")
	if err != nil {
		return err
	}
	if !found || identity != state.preparedIdentity || vNextPublicationDigest(prepared) != state.preparedDigest {
		return fmt.Errorf("publication control repair prepared authority identity changed")
	}
	if err := state.validateAnchors(transaction); err != nil {
		return err
	}
	for _, phase := range state.phases {
		payload, found, identity, err := vNextPublicationReadControlBound(transaction, phase.name, "publication control repair phase")
		if err != nil {
			return err
		}
		if !found || identity != phase.identity || vNextPublicationDigest(payload) != phase.digest {
			return fmt.Errorf("publication control repair phase %q identity changed", phase.name)
		}
		switch phase.record.State {
		case vNextPublicationControlRepairCaptureIntent:
			if err := vNextPublicationValidateCaptureLocked(transaction, *phase.record.Capture); err != nil {
				return err
			}
		case vNextPublicationControlRepairCaptured:
			if err := vNextPublicationValidateCapturedCandidateLocked(transaction, *phase.record.Capture, *phase.record.Candidate); err != nil {
				return err
			}
		}
	}
	if err := state.assertPredecessorIdentity(operation); err != nil {
		return err
	}
	return nil
}

func (state *vNextPublicationControlRepairState) assertPredecessorIdentity(operation *vNextPublicationOperation) (resultErr error) {
	predecessor := state.record.Predecessor
	if predecessor == nil {
		return nil
	}
	expectedTransaction, err := predecessor.TransactionIdentity.publicationIdentity(unix.S_IFDIR)
	if err != nil {
		return err
	}
	if err := operation.connector.assertIdentity(predecessor.Transaction, "publication control predecessor transaction", expectedTransaction); err != nil {
		return err
	}
	transaction, err := operation.connector.openDirectory(predecessor.Transaction, "publication control predecessor transaction")
	if err != nil {
		return err
	}
	defer vNextPublicationCloseAfter(&resultErr, transaction, "publication control predecessor transaction")
	actualTransaction, err := vNextPublicationIdentityFromFile(transaction.file, "publication control predecessor transaction")
	if err != nil {
		return err
	}
	if actualTransaction != expectedTransaction {
		return fmt.Errorf("publication control predecessor transaction identity changed")
	}
	expectedTerminal, err := predecessor.Terminal.Identity.publicationIdentity(unix.S_IFREG)
	if err != nil {
		return err
	}
	payload, found, identity, err := vNextPublicationReadControlBound(transaction, predecessor.Terminal.Name, "publication control predecessor terminal")
	if err != nil {
		return err
	}
	if !found || identity != expectedTerminal || vNextPublicationDigest(payload) != predecessor.Terminal.Digest {
		return fmt.Errorf("publication control predecessor terminal identity changed")
	}
	var phase vNextPublicationControlRepairPhase
	if err := vNextPublicationDecode(payload, &phase); err != nil {
		return fmt.Errorf("decode publication control predecessor terminal: %w", err)
	}
	if err := vNextPublicationControlRepairPhaseValid(phase); err != nil {
		return err
	}
	if phase.State != vNextPublicationControlRepairTerminal || phase.Selected == nil || !phase.Selected.sameLogical(predecessor.Selected) {
		return fmt.Errorf("publication control predecessor terminal changed")
	}
	return nil
}

func (state *vNextPublicationControlRepairState) anchor(transaction *vNextPublicationDirectory, control vNextPublicationControlState, member string) (vNextPublicationIdentity, bool, error) {
	identity, present, err := control.identity(member)
	if err != nil || !present {
		return identity, present, err
	}
	if err := transaction.assertIdentity(member, "publication control anchor", identity); err != nil {
		return vNextPublicationIdentity{}, false, err
	}
	return identity, true, nil
}

func vNextPublicationReadControlState(directory *vNextPublicationDirectory, name, label, member string) (vNextPublicationControlState, error) {
	_, found, identity, err := vNextPublicationReadControlBound(directory, name, label)
	if err != nil {
		return vNextPublicationControlState{}, err
	}
	if !found {
		return vNextPublicationAbsentControlState(), nil
	}
	return vNextPublicationPresentControlState(member, identity), nil
}

func vNextPublicationWriteAuthorityMarker(root *vNextPublicationDirectory, hooks vNextPublicationControlRecordHooks) (vNextPublicationIdentity, error) {
	marker := vNextPublicationControlAuthorityMarker{Version: vNextPublicationControlRepairVersion, Targets: []string{vNextPublicationCurrentFile, vNextPublicationJournalFile}}
	payload, err := json.Marshal(marker)
	if err != nil {
		return vNextPublicationIdentity{}, err
	}
	payload = append(payload, '\n')
	record, err := vNextPublicationWriteControlRepairRecord(root, vNextPublicationControlAuthorityMarkerFile, "publication control authority marker", payload, hooks)
	if err != nil {
		return vNextPublicationIdentity{}, err
	}
	if err := root.Sync(); err != nil {
		return vNextPublicationIdentity{}, fmt.Errorf("sync publication control authority marker: %w", err)
	}
	return record.identity, nil
}

func (p *vNextGenerationPublisher) scanControlAuthorityLocked(operation *vNextPublicationOperation) (*vNextPublicationControlAuthorityGraph, error) {
	if err := operation.assertLockBound(); err != nil {
		return nil, err
	}
	graph := &vNextPublicationControlAuthorityGraph{states: make(map[string]*vNextPublicationControlRepairState), heads: make(map[string]*vNextPublicationControlAuthorityHead)}
	markerPayload, markerFound, markerIdentity, err := vNextPublicationReadControlBound(operation.connector, vNextPublicationControlAuthorityMarkerFile, "publication control authority marker")
	if err != nil {
		return nil, err
	}
	if markerFound {
		var marker vNextPublicationControlAuthorityMarker
		if err := vNextPublicationDecode(markerPayload, &marker); err != nil {
			return nil, fmt.Errorf("decode publication control authority marker: %w", err)
		}
		if err := vNextPublicationControlAuthorityMarkerValid(marker); err != nil {
			return nil, err
		}
		graph.marker = true
		graph.markerIdentity = markerIdentity
	}
	entries, err := operation.connector.readDir()
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, vNextPublicationControlRepairDirectoryPrefix) {
			continue
		}
		if !vNextPublicationControlRepairTransactionNameValid(name) {
			graph.close()
			return nil, fmt.Errorf("invalid publication control repair transaction entry %q", name)
		}
		state, err := p.readControlRepairStateLocked(operation, name)
		if err != nil {
			graph.close()
			return nil, err
		}
		graph.states[name] = state
	}
	if err := vNextPublicationBuildAuthorityHeads(graph); err != nil {
		graph.close()
		return nil, err
	}
	return graph, nil
}

func (p *vNextGenerationPublisher) readControlRepairStateLocked(operation *vNextPublicationOperation, name string) (state *vNextPublicationControlRepairState, resultErr error) {
	transactionIdentity, err := operation.connector.identityAt(name, "publication control repair transaction")
	if err != nil {
		return nil, err
	}
	if transactionIdentity.mode != unix.S_IFDIR {
		return nil, fmt.Errorf("publication control repair transaction is not a directory")
	}
	transaction, err := operation.connector.openDirectory(name, "publication control repair transaction")
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := transaction.Close(); closeErr != nil && resultErr == nil {
			state = nil
			resultErr = fmt.Errorf("close publication control repair transaction: %w", closeErr)
		}
	}()
	fail := func(err error) (*vNextPublicationControlRepairState, error) {
		return nil, err
	}
	actualTransactionIdentity, err := vNextPublicationIdentityFromFile(transaction.file, "publication control repair transaction")
	if err != nil {
		return fail(err)
	}
	if actualTransactionIdentity != transactionIdentity {
		return fail(fmt.Errorf("publication control repair transaction identity changed"))
	}
	payload, found, preparedIdentity, err := vNextPublicationReadControlBound(transaction, vNextPublicationControlRepairPreparedFile, "publication control repair prepared authority")
	if err != nil {
		return fail(err)
	}
	if !found {
		return fail(fmt.Errorf("publication control repair transaction %q has no prepared authority", name))
	}
	var repair vNextPublicationControlRepair
	if err := vNextPublicationDecode(payload, &repair); err != nil {
		return fail(fmt.Errorf("decode publication control repair prepared authority: %w", err))
	}
	if err := vNextPublicationControlRepairValid(repair); err != nil {
		return fail(err)
	}
	if repair.Transaction != name || repair.TransactionIdentity != vNextPublicationRecordIdentity(transactionIdentity) {
		return fail(fmt.Errorf("publication control repair prepared authority does not bind its transaction"))
	}
	state = &vNextPublicationControlRepairState{
		record: repair, preparedIdentity: preparedIdentity, preparedDigest: vNextPublicationDigest(payload), transactionName: name, transactionIdentity: transactionIdentity,
	}
	if err := state.validateAnchors(transaction); err != nil {
		return nil, err
	}
	if err := p.readControlRepairPhasesLocked(operation, state, transaction); err != nil {
		return nil, err
	}
	return state, nil
}

func (state *vNextPublicationControlRepairState) validateAnchors(transaction *vNextPublicationDirectory) error {
	if _, _, err := state.anchor(transaction, state.record.Prior, vNextPublicationControlBackupMember); err != nil {
		return fmt.Errorf("validate publication control repair prior anchor: %w", err)
	}
	if _, _, err := state.anchor(transaction, state.record.Intended, vNextPublicationControlReplacementMember); err != nil {
		return fmt.Errorf("validate publication control repair intended anchor: %w", err)
	}
	return nil
}

func (p *vNextGenerationPublisher) readControlRepairPhasesLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, transaction *vNextPublicationDirectory) error {
	if err := state.assertPrivateIdentityAt(operation, transaction); err != nil {
		return err
	}
	entries, err := transaction.readDir()
	if err != nil {
		return err
	}
	allowed := map[string]struct{}{vNextPublicationControlRepairPreparedFile: {}}
	if state.record.Prior.Present {
		allowed[vNextPublicationControlBackupMember] = struct{}{}
	}
	if state.record.Intended.Present {
		allowed[vNextPublicationControlReplacementMember] = struct{}{}
	}
	for sequence := 1; sequence <= vNextPublicationControlRepairMaxPhases; sequence++ {
		allowed[vNextPublicationControlRepairPhaseName(sequence)] = struct{}{}
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, vNextPublicationControlCapturePrefix) {
			if !vNextPublicationControlCaptureNameValid(name) {
				return fmt.Errorf("invalid publication control capture entry %q", name)
			}
			continue
		}
		if _, ok := allowed[name]; !ok {
			return fmt.Errorf("unexpected publication control repair transaction member %q", name)
		}
	}

	previous := state.preparedReference()
	missing := false
	captures := make(map[string]vNextPublicationControlRepairCapture)
	lastCapture := ""
	lastState := ""
	for sequence := 1; sequence <= vNextPublicationControlRepairMaxPhases; sequence++ {
		name := vNextPublicationControlRepairPhaseName(sequence)
		payload, found, identity, err := vNextPublicationReadControlBound(transaction, name, "publication control repair phase")
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
			return fmt.Errorf("publication control repair phase %q does not bind its predecessor", name)
		}
		if err := vNextPublicationValidateControlRepairPhaseSemanticLocked(state, transaction, phase, lastState, captures, lastCapture); err != nil {
			return err
		}
		if phase.State == vNextPublicationControlRepairCaptureIntent {
			captures[phase.Capture.Name] = *phase.Capture
			lastCapture = phase.Capture.Name
		}
		previous = vNextPublicationControlRepairPhaseReference{Name: name, Identity: vNextPublicationRecordIdentity(identity), Digest: vNextPublicationDigest(payload)}
		state.phases = append(state.phases, vNextPublicationControlRepairPhaseState{record: phase, name: name, identity: identity, digest: vNextPublicationDigest(payload)})
		lastState = phase.State
	}
	unrecordedCapture := ""
	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), vNextPublicationControlCapturePrefix) {
			continue
		}
		capture, ok := captures[entry.Name()]
		if !ok {
			if unrecordedCapture != "" {
				return fmt.Errorf("multiple unrecorded publication control captures %q and %q", unrecordedCapture, entry.Name())
			}
			unrecordedCapture = entry.Name()
			continue
		}
		if err := vNextPublicationValidateCaptureLocked(transaction, capture); err != nil {
			return err
		}
	}
	if unrecordedCapture == "" {
		return nil
	}
	if state.latestPhase() != "" && state.latestPhase() != vNextPublicationControlRepairCaptured {
		return fmt.Errorf("unreferenced publication control capture %q follows %q", unrecordedCapture, state.latestPhase())
	}
	expected := vNextPublicationControlCaptureName(len(captures) + 1)
	if unrecordedCapture != expected {
		return fmt.Errorf("unreferenced publication control capture %q is not next expected %q", unrecordedCapture, expected)
	}
	_, err = vNextPublicationReadEmptyControlCaptureLocked(transaction, unrecordedCapture)
	return err
}

func vNextPublicationValidateControlRepairPhaseSemanticLocked(state *vNextPublicationControlRepairState, transaction *vNextPublicationDirectory, phase vNextPublicationControlRepairPhase, previous string, captures map[string]vNextPublicationControlRepairCapture, lastCapture string) error {
	if len(state.phases) == 0 && phase.State == vNextPublicationControlRepairTerminal {
		if state.record.Predecessor != nil || !state.record.Prior.sameLogical(state.record.Intended) || phase.Selected == nil || !phase.Selected.sameLogical(state.record.Intended) || phase.Outcome != vNextPublicationControlRepairCommitted {
			return fmt.Errorf("invalid publication control base terminal")
		}
		return nil
	}
	switch previous {
	case "":
		if phase.State != vNextPublicationControlRepairCaptureIntent {
			return fmt.Errorf("publication control repair must begin with capture intent")
		}
	case vNextPublicationControlRepairCaptureIntent:
		if phase.State != vNextPublicationControlRepairCaptured || phase.Capture == nil || phase.Capture.Name != lastCapture {
			return fmt.Errorf("publication control capture intent is not followed by its capture")
		}
	case vNextPublicationControlRepairCaptured:
		if phase.State == vNextPublicationControlRepairCaptureIntent || phase.State == vNextPublicationControlRepairSelected {
			break
		}
		if phase.State == vNextPublicationControlRepairTerminal && phase.Outcome == vNextPublicationControlRepairRetryRequired && len(captures) == vNextPublicationControlRepairMaxCaptureAttempts {
			break
		}
		return fmt.Errorf("invalid publication control transition after capture")
	case vNextPublicationControlRepairSelected:
		if phase.State != vNextPublicationControlRepairTerminal {
			return fmt.Errorf("publication control selection is not terminalized")
		}
	default:
		return fmt.Errorf("publication control repair phase follows terminal state")
	}
	switch phase.State {
	case vNextPublicationControlRepairCaptureIntent:
		if _, duplicate := captures[phase.Capture.Name]; duplicate || phase.Capture.Attempt != len(captures)+1 {
			return fmt.Errorf("invalid publication control capture sequence")
		}
		return vNextPublicationValidateCaptureLocked(transaction, *phase.Capture)
	case vNextPublicationControlRepairCaptured:
		if _, ok := captures[phase.Capture.Name]; !ok {
			return fmt.Errorf("publication capture phase has no prior intent")
		}
		return vNextPublicationValidateCapturedCandidateLocked(transaction, *phase.Capture, *phase.Candidate)
	case vNextPublicationControlRepairSelected, vNextPublicationControlRepairTerminal:
		if phase.Selected == nil || (!phase.Selected.sameLogical(state.record.Prior) && !phase.Selected.sameLogical(state.record.Intended)) {
			return fmt.Errorf("publication control phase selects neither prior nor intended state")
		}
		if phase.State == vNextPublicationControlRepairTerminal {
			switch phase.Outcome {
			case vNextPublicationControlRepairCommitted:
				if !phase.Selected.sameLogical(state.record.Intended) {
					return fmt.Errorf("committed publication terminal does not select intended state")
				}
			case vNextPublicationControlRepairRolledBack, vNextPublicationControlRepairRetryRequired:
				if !phase.Selected.sameLogical(state.record.Prior) {
					return fmt.Errorf("rollback publication terminal does not select prior state")
				}
			}
		}
	}
	return nil
}

// vNextPublicationBeforeRecordedCaptureOpenForTest is an inert, exact-boundary
// seam for the CP11 capture proof. It runs immediately before the real
// descriptor-relative open; production leaves it nil.
var vNextPublicationBeforeRecordedCaptureOpenForTest func(*vNextPublicationDirectory, vNextPublicationControlRepairCapture)

func vNextPublicationOpenRecordedCaptureLocked(transaction *vNextPublicationDirectory, capture vNextPublicationControlRepairCapture) (directory *vNextPublicationDirectory, resultErr error) {
	expected, err := capture.Identity.publicationIdentity(unix.S_IFDIR)
	if err != nil {
		return nil, err
	}
	if vNextPublicationBeforeRecordedCaptureOpenForTest != nil {
		vNextPublicationBeforeRecordedCaptureOpenForTest(transaction, capture)
	}
	directory, err = transaction.openDirectory(capture.Name, "publication control capture")
	if err != nil {
		return nil, err
	}
	actual, err := vNextPublicationIdentityFromFile(directory.file, "publication control capture")
	if err == nil && actual != expected {
		err = fmt.Errorf("publication control capture identity changed")
	}
	if err != nil {
		vNextPublicationCloseAfter(&err, directory, "publication control capture")
		return nil, err
	}
	return directory, nil
}

func vNextPublicationValidateCaptureLocked(transaction *vNextPublicationDirectory, capture vNextPublicationControlRepairCapture) (resultErr error) {
	directory, err := vNextPublicationOpenRecordedCaptureLocked(transaction, capture)
	if err != nil {
		return err
	}
	defer vNextPublicationCloseAfter(&resultErr, directory, "publication control capture")
	entries, err := directory.readDir()
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Name() != vNextPublicationControlCaptureMember {
			return fmt.Errorf("unexpected publication control capture member %q", entry.Name())
		}
	}
	return nil
}

func vNextPublicationReadEmptyControlCaptureLocked(transaction *vNextPublicationDirectory, name string) (identity vNextPublicationIdentity, resultErr error) {
	directory, err := transaction.openDirectory(name, "publication control capture")
	if err != nil {
		return vNextPublicationIdentity{}, err
	}
	defer func() {
		if closeErr := directory.Close(); closeErr != nil && resultErr == nil {
			identity = vNextPublicationIdentity{}
			resultErr = fmt.Errorf("close publication control capture: %w", closeErr)
		}
	}()
	identity, err = vNextPublicationIdentityFromFile(directory.file, "publication control capture")
	if err != nil {
		return vNextPublicationIdentity{}, err
	}
	entries, err := directory.readDir()
	if err != nil {
		return vNextPublicationIdentity{}, err
	}
	if len(entries) != 0 {
		return vNextPublicationIdentity{}, fmt.Errorf("unrecorded publication control capture %q is not empty", name)
	}
	return identity, nil
}

func vNextPublicationValidateCapturedCandidateLocked(transaction *vNextPublicationDirectory, capture vNextPublicationControlRepairCapture, candidate vNextPublicationControlState) (resultErr error) {
	if err := vNextPublicationValidateCaptureLocked(transaction, capture); err != nil {
		return err
	}
	directory, err := vNextPublicationOpenRecordedCaptureLocked(transaction, capture)
	if err != nil {
		return err
	}
	defer vNextPublicationCloseAfter(&resultErr, directory, "publication control capture")
	identity, present, err := candidate.identity(vNextPublicationControlCaptureMember)
	if err != nil {
		return err
	}
	actual, err := directory.identityAt(vNextPublicationControlCaptureMember, "publication control capture candidate")
	if errors.Is(err, fs.ErrNotExist) {
		if present {
			return fmt.Errorf("publication control capture candidate is missing")
		}
		return nil
	}
	if err != nil {
		return err
	}
	if !present || actual != identity {
		return fmt.Errorf("publication control capture candidate identity changed")
	}
	return nil
}

func vNextPublicationBuildAuthorityHeads(graph *vNextPublicationControlAuthorityGraph) error {
	children := make(map[string]int, len(graph.states))
	for name, state := range graph.states {
		predecessor := state.record.Predecessor
		if predecessor == nil {
			continue
		}
		parent, ok := graph.states[predecessor.Transaction]
		if !ok {
			return fmt.Errorf("publication control repair %q references missing predecessor %q", name, predecessor.Transaction)
		}
		if parent.transactionIdentity != mustVNextPublicationRecordedDirectoryIdentity(predecessor.TransactionIdentity) || parent.record.Target != state.record.Target {
			return fmt.Errorf("publication control repair %q predecessor does not bind target identity", name)
		}
		terminal, selected, _, ok := parent.terminal()
		expectedTerminal := vNextPublicationControlRepairPhaseReference{Name: terminal.name, Identity: vNextPublicationRecordIdentity(terminal.identity), Digest: terminal.digest}
		if !ok || predecessor.Terminal != expectedTerminal || !predecessor.Selected.sameLogical(selected) {
			return fmt.Errorf("publication control repair %q predecessor is not the retained terminal", name)
		}
		children[predecessor.Transaction]++
		if children[predecessor.Transaction] > 1 {
			return fmt.Errorf("publication control repair predecessor %q has multiple successors", predecessor.Transaction)
		}
	}
	for name := range graph.states {
		seen := make(map[string]struct{})
		current := name
		for {
			if _, repeated := seen[current]; repeated {
				return fmt.Errorf("publication control repair predecessor cycle at %q", current)
			}
			seen[current] = struct{}{}
			predecessor := graph.states[current].record.Predecessor
			if predecessor == nil {
				break
			}
			current = predecessor.Transaction
		}
	}
	candidateHeads := make(map[string][]*vNextPublicationControlAuthorityHead)
	for name, state := range graph.states {
		if children[name] != 0 {
			continue
		}
		terminal, selected, outcome, terminalized := state.terminal()
		if !terminalized {
			candidateHeads[state.record.Target] = append(candidateHeads[state.record.Target], &vNextPublicationControlAuthorityHead{state: state})
			continue
		}
		candidateHeads[state.record.Target] = append(candidateHeads[state.record.Target], &vNextPublicationControlAuthorityHead{state: state, terminal: terminal, selected: selected, outcome: outcome})
	}
	for target, heads := range candidateHeads {
		if len(heads) != 1 {
			return fmt.Errorf("publication control %q has %d authority heads", target, len(heads))
		}
		graph.heads[target] = heads[0]
	}
	if graph.marker {
		for _, target := range []string{vNextPublicationCurrentFile, vNextPublicationJournalFile} {
			if _, ok := graph.heads[target]; !ok {
				return fmt.Errorf("publication control authority marker has no %s head", target)
			}
		}
	}
	return nil
}

func mustVNextPublicationRecordedDirectoryIdentity(recorded vNextPublicationRecordedIdentity) vNextPublicationIdentity {
	identity, err := recorded.publicationIdentity(unix.S_IFDIR)
	if err != nil {
		return vNextPublicationIdentity{}
	}
	return identity
}

func (p *vNextGenerationPublisher) ensureControlAuthorityLocked(operation *vNextPublicationOperation) error {
	for range 4 {
		graph, err := p.scanControlAuthorityLocked(operation)
		if err != nil {
			return err
		}
		if graph.marker {
			graph.close()
			return nil
		}

		var pending *vNextPublicationControlRepairState
		for _, state := range graph.states {
			if state.record.Predecessor != nil {
				graph.close()
				return fmt.Errorf("publication control authority marker is missing from successor graph")
			}
			if _, _, _, terminal := state.terminal(); terminal {
				continue
			}
			if !vNextPublicationResumableBaseControlAuthority(state) || pending != nil {
				graph.close()
				return fmt.Errorf("publication control authority marker is missing from pending repair")
			}
			pending = state
		}
		if pending != nil {
			err := p.resumeBaseControlAuthorityLocked(operation, pending)
			graph.close()
			if err != nil {
				return err
			}
			continue
		}

		missing := ""
		for _, target := range []string{vNextPublicationCurrentFile, vNextPublicationJournalFile} {
			head, found := graph.heads[target]
			if !found {
				missing = target
				break
			}
			if _, _, _, terminal := head.state.terminal(); !terminal {
				graph.close()
				return fmt.Errorf("publication control authority marker is missing from pending repair")
			}
		}
		graph.close()
		if missing == "" {
			_, err := vNextPublicationWriteAuthorityMarker(operation.connector, p.hooks.ControlRecord)
			return err
		}
		state, err := p.createBaseControlAuthorityLocked(operation, missing)
		if err != nil {
			return err
		}
		state.close()
	}
	return fmt.Errorf("publication control authority bootstrap did not create marker")
}

func vNextPublicationResumableBaseControlAuthority(state *vNextPublicationControlRepairState) bool {
	return state != nil &&
		state.record.Predecessor == nil &&
		len(state.phases) == 0 &&
		state.record.Prior.sameLogical(state.record.Intended)
}

func (p *vNextGenerationPublisher) resumeBaseControlAuthorityLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState) error {
	if !vNextPublicationResumableBaseControlAuthority(state) {
		return fmt.Errorf("publication control authority marker is missing from pending repair")
	}
	if err := state.assertPrivateIdentity(operation); err != nil {
		return err
	}
	selected := state.record.Intended
	return p.appendControlRepairPhaseLocked(operation, state, vNextPublicationControlRepairPhase{
		State:    vNextPublicationControlRepairTerminal,
		Selected: &selected,
		Outcome:  vNextPublicationControlRepairCommitted,
	})
}

func (p *vNextGenerationPublisher) createBaseControlAuthorityLocked(operation *vNextPublicationOperation, target string) (*vNextPublicationControlRepairState, error) {
	observed, err := vNextPublicationReadControlState(operation.connector, target, "publication control", vNextPublicationControlBackupMember)
	if err != nil {
		return nil, err
	}
	var source *vNextPublicationControlAnchorSource
	if observed.Present {
		identity, _, err := observed.identity(vNextPublicationControlBackupMember)
		if err != nil {
			return nil, err
		}
		source = &vNextPublicationControlAnchorSource{directory: operation.connector, name: target, identity: identity}
	}
	return p.createControlRepairLocked(operation, target, nil, observed, source, true)
}

type vNextPublicationControlAnchorSource struct {
	directory *vNextPublicationDirectory
	name      string
	identity  vNextPublicationIdentity
}

type vNextPublicationControlRepairAnchor struct {
	name     string
	identity vNextPublicationIdentity
}

func (p *vNextGenerationPublisher) createControlRepairLocked(operation *vNextPublicationOperation, target string, predecessor *vNextPublicationControlAuthorityHead, intended vNextPublicationControlState, intendedSource *vNextPublicationControlAnchorSource, base bool) (state *vNextPublicationControlRepairState, resultErr error) {
	if err := operation.assertLockBound(); err != nil {
		return nil, err
	}
	transactionName, transaction, transactionIdentity, err := vNextPublicationCreateControlRepairTransaction(operation.connector)
	if err != nil {
		return nil, err
	}
	keep := false
	anchors := make([]vNextPublicationControlRepairAnchor, 0, 2)
	var preparedIdentity vNextPublicationIdentity
	defer func() {
		if keep {
			if closeErr := transaction.Close(); closeErr != nil {
				state = nil
				vNextPublicationRecordError(&resultErr, "close publication control repair transaction", closeErr)
			}
			return
		}
		// A failed exclusive open does not establish absence. Never remove
		// dependencies while any unknown/retained record still occupies the name.
		if _, absenceErr := transaction.identityAt(vNextPublicationControlRepairPreparedFile, "unprepared authority"); !vNextPublicationPureNotExist(absenceErr) {
			if absenceErr != nil {
				vNextPublicationRecordError(&resultErr, "check unprepared authority absence", absenceErr)
			}
			vNextPublicationCloseAfter(&resultErr, transaction, "unprepared publication control repair transaction")
			return
		}

		for index := len(anchors) - 1; index >= 0; index-- {
			anchor := anchors[index]
			vNextPublicationRecordError(&resultErr, "remove unprepared publication control anchor", transaction.removeRegularBound(anchor.name, "unprepared publication control anchor", anchor.identity))
		}
		vNextPublicationRecordError(&resultErr, "sync unprepared publication control repair transaction", transaction.Sync())
		vNextPublicationCloseAfter(&resultErr, transaction, "unprepared publication control repair transaction")
		vNextPublicationRecordError(&resultErr, "remove unprepared publication control repair transaction", operation.connector.removeEmptyDirectoryBound(transactionName, "unprepared publication control repair transaction", transactionIdentity))
		vNextPublicationRecordError(&resultErr, "sync publication connector after unprepared repair cleanup", operation.connector.Sync())
	}()

	prior := vNextPublicationAbsentControlState()
	if predecessor != nil {
		prior = vNextPublicationControlStateWithMember(predecessor.selected, vNextPublicationControlBackupMember)
		if identity, present, err := prior.identity(vNextPublicationControlBackupMember); err != nil {
			return nil, err
		} else if present {
			predecessorTransaction, err := predecessor.state.openTransaction(operation)
			if err != nil {
				return nil, err
			}
			linkErr := transaction.linkFromBound(predecessorTransaction, predecessor.selected.Member, vNextPublicationControlBackupMember, "publication control prior anchor", identity)
			if linkErr == nil {
				// The link is a known owned creation before the later fallible
				// predecessor Close returns through !keep cleanup.
				anchors = append(anchors, vNextPublicationControlRepairAnchor{name: vNextPublicationControlBackupMember, identity: identity})
			}
			closeErr := p.closeRepairPredecessor(predecessorTransaction)
			if linkErr != nil {
				if closeErr != nil {
					return nil, errors.Join(linkErr, fmt.Errorf("close publication control predecessor transaction: %w", closeErr))
				}
				return nil, linkErr
			}
			if closeErr != nil {
				return nil, fmt.Errorf("close publication control predecessor transaction: %w", closeErr)
			}
			if err := transaction.assertIdentity(vNextPublicationControlBackupMember, "publication control prior anchor", identity); err != nil {
				return nil, err
			}
		}
	}
	if base {
		prior = vNextPublicationControlStateWithMember(intended, vNextPublicationControlBackupMember)
		if identity, present, err := prior.identity(vNextPublicationControlBackupMember); err != nil {
			return nil, err
		} else if present {
			if intendedSource == nil {
				return nil, fmt.Errorf("publication control base authority has no source")
			}
			if err := transaction.linkFromBound(intendedSource.directory, intendedSource.name, vNextPublicationControlBackupMember, "publication control prior anchor", identity); err != nil {
				return nil, err
			}
			anchors = append(anchors, vNextPublicationControlRepairAnchor{name: vNextPublicationControlBackupMember, identity: identity})
		}
	}

	newIntended := vNextPublicationControlStateWithMember(intended, vNextPublicationControlReplacementMember)
	if identity, present, err := newIntended.identity(vNextPublicationControlReplacementMember); err != nil {
		return nil, err
	} else if present {
		if base {
			if err := transaction.linkFromBound(transaction, vNextPublicationControlBackupMember, vNextPublicationControlReplacementMember, "publication control intended anchor", identity); err != nil {
				return nil, err
			}
			anchors = append(anchors, vNextPublicationControlRepairAnchor{name: vNextPublicationControlReplacementMember, identity: identity})
		} else {
			if intendedSource == nil {
				return nil, fmt.Errorf("publication control intended state has no source")
			}
			if err := transaction.linkFromBound(intendedSource.directory, intendedSource.name, vNextPublicationControlReplacementMember, "publication control intended anchor", identity); err != nil {
				return nil, err
			}
			anchors = append(anchors, vNextPublicationControlRepairAnchor{name: vNextPublicationControlReplacementMember, identity: identity})
		}
	}
	if err := transaction.Sync(); err != nil {
		return nil, fmt.Errorf("sync publication control repair anchors: %w", err)
	}

	repair := vNextPublicationControlRepair{Version: vNextPublicationControlRepairVersion, Target: target, Transaction: transactionName, TransactionIdentity: vNextPublicationRecordIdentity(transactionIdentity), Prior: prior, Intended: newIntended, MaxCaptureAttempts: vNextPublicationControlRepairMaxCaptureAttempts}
	if predecessor != nil {
		repair.Predecessor = &vNextPublicationControlRepairPredecessor{Transaction: predecessor.state.transactionName, TransactionIdentity: vNextPublicationRecordIdentity(predecessor.state.transactionIdentity), Terminal: vNextPublicationControlRepairPhaseReference{Name: predecessor.terminal.name, Identity: vNextPublicationRecordIdentity(predecessor.terminal.identity), Digest: predecessor.terminal.digest}, Selected: predecessor.selected}
	}
	if err := vNextPublicationControlRepairValid(repair); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(repair)
	if err != nil {
		return nil, err
	}
	payload = append(payload, '\n')
	if err := p.hit(vNextPublicationBeforeControlRepairRecord); err != nil {
		return nil, err
	}
	prepared, err := vNextPublicationWriteControlRepairRecord(transaction, vNextPublicationControlRepairPreparedFile, "publication control repair prepared authority", payload, p.hooks.ControlRecord)
	preparedIdentity = prepared.identity
	keep = prepared.disposition == vNextPublicationRecordRetainedComplete
	if err != nil {
		return nil, err
	}
	// The prepared record has closed, synced, and bound its own identity. Every
	// later frontier retains the complete authority graph for fresh recovery.
	keep = true
	if err := p.hit(vNextPublicationAfterControlRepairRecord); err != nil {
		return nil, err
	}
	if err := transaction.Sync(); err != nil {
		return nil, fmt.Errorf("sync publication control repair prepared authority: %w", err)
	}
	if err := p.hit(vNextPublicationAfterControlRepairTransactionSync); err != nil {
		return nil, err
	}
	if err := operation.connector.Sync(); err != nil {
		return nil, fmt.Errorf("sync publication control repair transaction: %w", err)
	}
	if err := p.hit(vNextPublicationAfterControlRepairConnectorSync); err != nil {
		return nil, err
	}
	state = &vNextPublicationControlRepairState{record: repair, preparedIdentity: preparedIdentity, preparedDigest: vNextPublicationDigest(payload), transactionName: transactionName, transactionIdentity: transactionIdentity}
	if base {
		if err := p.hit(vNextPublicationAfterBaseControlRepairPrepared); err != nil {
			state.close()
			return nil, err
		}
		selected := newIntended
		if err := p.appendControlRepairPhaseLocked(operation, state, vNextPublicationControlRepairPhase{State: vNextPublicationControlRepairTerminal, Selected: &selected, Outcome: vNextPublicationControlRepairCommitted}); err != nil {
			state.close()
			return nil, err
		}
		return state, nil
	}
	if err := p.hit(vNextPublicationAfterControlRepairPrepared); err != nil {
		return state, err
	}
	return state, nil
}

func (p *vNextGenerationPublisher) appendControlRepairPhaseLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, phase vNextPublicationControlRepairPhase) (resultErr error) {
	transaction, err := state.openTransaction(operation)
	if err != nil {
		return err
	}
	defer vNextPublicationCloseAfter(&resultErr, transaction, "publication control repair transaction")
	if err := state.assertPrivateIdentityAt(operation, transaction); err != nil {
		return err
	}
	phase.Version = vNextPublicationControlRepairVersion
	phase.Sequence = len(state.phases) + 1
	phase.PreparedDigest = state.preparedDigest
	phase.PreparedIdentity = vNextPublicationRecordIdentity(state.preparedIdentity)
	phase.Previous = state.latestReference()
	if err := vNextPublicationControlRepairPhaseValid(phase); err != nil {
		return err
	}
	payload, err := json.Marshal(phase)
	if err != nil {
		return err
	}
	payload = append(payload, '\n')
	name := vNextPublicationControlRepairPhaseName(phase.Sequence)
	record, err := vNextPublicationWriteControlRepairRecord(transaction, name, "publication control repair phase", payload, p.hooks.ControlRecord)
	if record.disposition == vNextPublicationRecordRetainedComplete {
		state.phases = append(state.phases, vNextPublicationControlRepairPhaseState{record: phase, name: name, identity: record.identity, digest: vNextPublicationDigest(payload)})
	}
	if err != nil {
		return err
	}
	if err := transaction.Sync(); err != nil {
		return fmt.Errorf("sync publication control repair phase: %w", err)
	}
	return nil
}

func (p *vNextGenerationPublisher) beginControlCaptureLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState) (resultErr error) {
	attempt := 1
	for _, phase := range state.phases {
		if phase.record.State == vNextPublicationControlRepairCaptureIntent {
			attempt++
		}
	}
	if attempt > vNextPublicationControlRepairMaxCaptureAttempts {
		return p.terminalizeRetryLocked(operation, state)
	}
	transaction, err := state.openTransaction(operation)
	if err != nil {
		return err
	}
	defer vNextPublicationCloseAfter(&resultErr, transaction, "publication control repair transaction")
	if err := state.assertPrivateIdentityAt(operation, transaction); err != nil {
		return err
	}
	name := vNextPublicationControlCaptureName(attempt)
	created := false
	identity, err := transaction.identityAt(name, "publication control capture")
	if errors.Is(err, fs.ErrNotExist) {
		if err := unix.Mkdirat(int(transaction.file.Fd()), name, 0o700); err != nil {
			return fmt.Errorf("create publication control capture: %w", err)
		}
		created = true
		identity, err = transaction.identityAt(name, "publication control capture")
	}
	if err != nil {
		return err
	}
	if identity.mode != unix.S_IFDIR {
		return fmt.Errorf("publication control capture is not a directory")
	}
	capture, err := transaction.openDirectory(name, "publication control capture")
	if err != nil {
		return err
	}
	actual, err := vNextPublicationIdentityFromFile(capture.file, "publication control capture")
	if err == nil && actual != identity {
		err = fmt.Errorf("publication control capture identity changed")
	}
	if err == nil {
		entries, readErr := capture.readDir()
		if readErr != nil {
			err = readErr
		} else if len(entries) != 0 {
			err = fmt.Errorf("unrecorded publication control capture %q is not empty", name)
		}
	}
	vNextPublicationRecordError(&err, "prepare publication control capture", p.hit(vNextPublicationBeforeControlRepairCaptureClose))
	vNextPublicationRecordError(&err, "sync publication control capture", capture.Sync())
	vNextPublicationRecordError(&err, "complete publication control capture sync", p.hit(vNextPublicationAfterControlRepairCaptureSync))
	vNextPublicationCloseAfter(&err, capture, "publication control capture")
	if err != nil {
		return err
	}
	if err := transaction.Sync(); err != nil {
		return fmt.Errorf("sync publication control capture parent: %w", err)
	}
	if created {
		if err := p.hit(vNextPublicationAfterControlRepairCaptureDirectory); err != nil {
			return err
		}
	}
	if err := transaction.assertIdentity(name, "publication control capture", identity); err != nil {
		return err
	}
	captureRecord := vNextPublicationControlRepairCapture{Attempt: attempt, Name: name, Identity: vNextPublicationRecordIdentity(identity)}
	if err := p.appendControlRepairPhaseLocked(operation, state, vNextPublicationControlRepairPhase{State: vNextPublicationControlRepairCaptureIntent, Capture: &captureRecord}); err != nil {
		return err
	}
	return p.hit(vNextPublicationBeforeControlRepairCapture)
}

func (p *vNextGenerationPublisher) captureControlNoReplaceLocked(destination, source *vNextPublicationDirectory, sourceName, destinationName string) error {
	if p.hooks.ControlCaptureRename != nil {
		return vNextPublicationNoReplaceRenameError(p.hooks.ControlCaptureRename())
	}
	return destination.renameNoReplaceFrom(source, sourceName, destinationName)
}

func (p *vNextGenerationPublisher) completeControlCaptureLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState) (resultErr error) {
	if state.latestPhase() != vNextPublicationControlRepairCaptureIntent {
		return fmt.Errorf("publication control repair has no capture intent")
	}
	intent := state.phases[len(state.phases)-1].record
	capture := *intent.Capture
	transaction, err := state.openTransaction(operation)
	if err != nil {
		return err
	}
	defer vNextPublicationCloseAfter(&resultErr, transaction, "publication control repair transaction")
	if err := state.assertPrivateIdentityAt(operation, transaction); err != nil {
		return err
	}
	if err := vNextPublicationValidateCaptureLocked(transaction, capture); err != nil {
		return err
	}
	directory, err := vNextPublicationOpenRecordedCaptureLocked(transaction, capture)
	if err != nil {
		return err
	}
	defer vNextPublicationCloseAfter(&resultErr, directory, "publication control capture")
	candidateIdentity, err := directory.identityAt(vNextPublicationControlCaptureMember, "publication control capture candidate")
	candidate := vNextPublicationAbsentControlState()
	if err == nil {
		if candidateIdentity.mode != unix.S_IFREG {
			return fmt.Errorf("publication control capture candidate is not a regular file")
		}
		candidate = vNextPublicationPresentControlState(vNextPublicationControlCaptureMember, candidateIdentity)
	} else if errors.Is(err, fs.ErrNotExist) {
		if err := p.captureControlNoReplaceLocked(directory, operation.connector, state.record.Target, vNextPublicationControlCaptureMember); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return err
			}
		} else {
			candidateIdentity, err = directory.identityAt(vNextPublicationControlCaptureMember, "publication control capture candidate")
			if err != nil {
				return err
			}
			if candidateIdentity.mode != unix.S_IFREG {
				return fmt.Errorf("publication control capture candidate is not a regular file")
			}
			candidate = vNextPublicationPresentControlState(vNextPublicationControlCaptureMember, candidateIdentity)
		}
	} else {
		return err
	}
	if err := p.hit(vNextPublicationAfterControlRepairCaptureRename); err != nil {
		return err
	}
	if err := directory.Sync(); err != nil {
		return fmt.Errorf("sync publication control capture: %w", err)
	}
	if err := p.hit(vNextPublicationAfterControlRepairCaptureDirectorySync); err != nil {
		return err
	}
	if err := operation.connector.Sync(); err != nil {
		return fmt.Errorf("sync publication control after capture: %w", err)
	}
	if err := p.hit(vNextPublicationAfterControlRepairCaptureRootSync); err != nil {
		return err
	}
	if err := p.appendControlRepairPhaseLocked(operation, state, vNextPublicationControlRepairPhase{State: vNextPublicationControlRepairCaptured, Capture: &capture, Candidate: &candidate}); err != nil {
		return err
	}
	return p.hit(vNextPublicationAfterControlRepairCaptured)
}

func (p *vNextGenerationPublisher) selectControlStateLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState, selected vNextPublicationControlState) (installed bool, resultErr error) {
	if selected.Present {
		transaction, err := state.openTransaction(operation)
		if err != nil {
			return false, err
		}
		defer vNextPublicationCloseAfter(&resultErr, transaction, "publication control repair transaction")
		if err := state.assertPrivateIdentityAt(operation, transaction); err != nil {
			return false, err
		}
		identity, _, err := state.anchor(transaction, selected, selected.Member)
		if err != nil {
			return false, err
		}
		if err := operation.connector.linkFromBound(transaction, selected.Member, state.record.Target, "publication control selection", identity); err != nil {
			if errors.Is(err, fs.ErrExist) {
				return true, nil
			}
			return false, err
		}
	} else {
		_, err := operation.connector.identityAt(state.record.Target, "publication control selection")
		if err == nil {
			return true, nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return false, err
		}
	}
	if err := p.hit(vNextPublicationAfterControlRepairInstall); err != nil {
		return false, err
	}
	if err := operation.connector.Sync(); err != nil {
		return false, fmt.Errorf("sync publication control selection: %w", err)
	}
	if err := p.hit(vNextPublicationAfterControlRepairInstallSync); err != nil {
		return false, err
	}
	copy := selected
	if err := p.appendControlRepairPhaseLocked(operation, state, vNextPublicationControlRepairPhase{State: vNextPublicationControlRepairSelected, Selected: &copy}); err != nil {
		return false, err
	}
	if err := p.hit(vNextPublicationAfterControlRepairSelected); err != nil {
		return false, err
	}
	return false, nil
}

func (p *vNextGenerationPublisher) terminalizeRetryLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState) error {
	if _, _, _, terminal := state.terminal(); terminal {
		return &vNextPublicationControlConflictError{target: state.record.Target}
	}
	selected := state.record.Prior
	if err := p.appendControlRepairPhaseLocked(operation, state, vNextPublicationControlRepairPhase{State: vNextPublicationControlRepairTerminal, Selected: &selected, Outcome: vNextPublicationControlRepairRetryRequired}); err != nil {
		return err
	}
	return &vNextPublicationControlConflictError{target: state.record.Target}
}

func (p *vNextGenerationPublisher) resolveControlRepairLocked(operation *vNextPublicationOperation, state *vNextPublicationControlRepairState) error {
	for {
		if err := state.assertPrivateIdentity(operation); err != nil {
			return err
		}
		switch state.latestPhase() {
		case "":
			if err := p.beginControlCaptureLocked(operation, state); err != nil {
				return err
			}
		case vNextPublicationControlRepairCaptureIntent:
			if err := p.completeControlCaptureLocked(operation, state); err != nil {
				return err
			}
		case vNextPublicationControlRepairCaptured:
			captured := state.phases[len(state.phases)-1].record
			if captured.Candidate == nil {
				return fmt.Errorf("publication control capture has no candidate")
			}
			selected := state.record.Intended
			if !captured.Candidate.sameLogical(state.record.Prior) {
				selected = state.record.Prior
			}
			late, err := p.selectControlStateLocked(operation, state, selected)
			if err != nil {
				return err
			}
			if late {
				if err := p.beginControlCaptureLocked(operation, state); err != nil {
					return err
				}
			}
		case vNextPublicationControlRepairSelected:
			selected := *state.phases[len(state.phases)-1].record.Selected
			matches, err := vNextPublicationControlStateMatchesPublic(operation.connector, state.record.Target, selected)
			if err != nil {
				return err
			}
			if !matches {
				return p.terminalizeRetryLocked(operation, state)
			}
			outcome := vNextPublicationControlRepairRolledBack
			if selected.sameLogical(state.record.Intended) {
				outcome = vNextPublicationControlRepairCommitted
			}
			if err := p.appendControlRepairPhaseLocked(operation, state, vNextPublicationControlRepairPhase{State: vNextPublicationControlRepairTerminal, Selected: &selected, Outcome: outcome}); err != nil {
				return err
			}
			if err := p.hit(vNextPublicationAfterFinalControlRepairValidation); err != nil {
				return err
			}
		case vNextPublicationControlRepairTerminal:
			_, selected, outcome, _ := state.terminal()
			if outcome == vNextPublicationControlRepairRetryRequired {
				return &vNextPublicationControlConflictError{target: state.record.Target}
			}
			matches, err := vNextPublicationControlStateMatchesPublic(operation.connector, state.record.Target, selected)
			if err != nil {
				return err
			}
			if !matches {
				return &vNextPublicationControlConflictError{target: state.record.Target}
			}
			return nil
		default:
			return fmt.Errorf("invalid publication control repair phase %q", state.latestPhase())
		}
	}
}

func vNextPublicationControlStateMatchesPublic(root *vNextPublicationDirectory, target string, selected vNextPublicationControlState) (bool, error) {
	identity, present, err := selected.identity(selected.Member)
	if err != nil {
		return false, err
	}
	_, found, actual, err := vNextPublicationReadControlBound(root, target, "publication control")
	if err != nil {
		return false, err
	}
	if found != present {
		return false, nil
	}
	return !present || actual == identity, nil
}

func (p *vNextGenerationPublisher) recoverControlRepairLocked(operation *vNextPublicationOperation) error {
	if err := p.ensureControlAuthorityLocked(operation); err != nil {
		return err
	}
	for range 16 {
		graph, err := p.scanControlAuthorityLocked(operation)
		if err != nil {
			return err
		}
		if !graph.marker {
			graph.close()
			return fmt.Errorf("publication control authority bootstrap did not create marker")
		}
		var repair *vNextPublicationControlRepairState
		var reconcile *vNextPublicationControlAuthorityHead
		for _, target := range []string{vNextPublicationCurrentFile, vNextPublicationJournalFile} {
			head := graph.heads[target]
			if head == nil {
				graph.close()
				return fmt.Errorf("publication control authority has no %s head", target)
			}
			if _, _, _, terminal := head.state.terminal(); !terminal {
				repair = head.state
				break
			}
			matches, matchErr := vNextPublicationControlStateMatchesPublic(operation.connector, target, head.selected)
			if matchErr != nil {
				graph.close()
				return matchErr
			}
			if head.outcome == vNextPublicationControlRepairRetryRequired || !matches {
				reconcile = head
				break
			}
		}
		if repair != nil {
			err = p.resolveControlRepairLocked(operation, repair)
			graph.close()
			if err != nil && !errors.Is(err, errVNextPublicationControlConflict) {
				return err
			}
			continue
		}
		if reconcile != nil {
			selected := reconcile.selected
			var source *vNextPublicationControlAnchorSource
			var sourceTransaction *vNextPublicationDirectory
			if selected.Present {
				sourceTransaction, err = reconcile.state.openTransaction(operation)
				if err != nil {
					graph.close()
					return err
				}
				identity, _, identityErr := reconcile.state.anchor(sourceTransaction, selected, selected.Member)
				if identityErr != nil {
					_ = sourceTransaction.Close()
					graph.close()
					return identityErr
				}
				source = &vNextPublicationControlAnchorSource{directory: sourceTransaction, name: selected.Member, identity: identity}
			}
			state, createErr := p.createControlRepairLocked(operation, reconcile.state.record.Target, reconcile, selected, source, false)
			if sourceTransaction != nil {
				if closeErr := sourceTransaction.Close(); closeErr != nil && createErr == nil {
					createErr = fmt.Errorf("close publication control reconciliation source: %w", closeErr)
				}
			}
			graph.close()
			if createErr != nil {
				return createErr
			}
			err = p.resolveControlRepairLocked(operation, state)
			state.close()
			if err != nil && !errors.Is(err, errVNextPublicationControlConflict) {
				return err
			}
			continue
		}
		graph.close()
		return nil
	}
	return &vNextPublicationControlConflictError{target: "CURRENT/JOURNAL"}
}

func (p *vNextGenerationPublisher) transitionControlLocked(operation *vNextPublicationOperation, target string, intended vNextPublicationControlState, source *vNextPublicationControlAnchorSource) (vNextPublicationControlState, error) {
	if err := p.recoverControlRepairLocked(operation); err != nil {
		return vNextPublicationControlState{}, err
	}
	graph, err := p.scanControlAuthorityLocked(operation)
	if err != nil {
		return vNextPublicationControlState{}, err
	}
	head := graph.heads[target]
	if head == nil || head.outcome == vNextPublicationControlRepairRetryRequired {
		graph.close()
		return vNextPublicationControlState{}, &vNextPublicationControlConflictError{target: target}
	}
	state, err := p.createControlRepairLocked(operation, target, head, intended, source, false)
	graph.close()
	if err != nil {
		return vNextPublicationControlState{}, err
	}
	defer state.close()
	if err := p.resolveControlRepairLocked(operation, state); err != nil {
		return vNextPublicationControlState{}, err
	}
	_, selected, _, terminal := state.terminal()
	if !terminal {
		return vNextPublicationControlState{}, fmt.Errorf("publication control transition did not terminalize")
	}
	return selected, nil
}

func (p *vNextGenerationPublisher) controlAuthorityForReadLocked(operation *vNextPublicationOperation) (*vNextPublicationControlAuthorityGraph, error) {
	graph, err := p.scanControlAuthorityLocked(operation)
	if err != nil {
		return nil, err
	}
	if !graph.marker {
		if len(graph.states) != 0 {
			graph.close()
			return nil, fmt.Errorf("publication control repair exists before authority bootstrap")
		}
		return graph, nil
	}
	for _, target := range []string{vNextPublicationCurrentFile, vNextPublicationJournalFile} {
		head := graph.heads[target]
		if head == nil || head.outcome == vNextPublicationControlRepairRetryRequired {
			graph.close()
			return nil, &vNextPublicationControlConflictError{target: target}
		}
		if _, _, _, terminal := head.state.terminal(); !terminal {
			graph.close()
			return nil, fmt.Errorf("publication control %q has a pending repair", target)
		}
	}
	if err := p.hit(vNextPublicationAfterControlAuthorityReadScan); err != nil {
		graph.close()
		return nil, err
	}
	return graph, nil
}

func vNextPublicationReadAuthorizedControlLocked(operation *vNextPublicationOperation, graph *vNextPublicationControlAuthorityGraph, target, label string) ([]byte, bool, vNextPublicationIdentity, error) {
	if graph == nil || !graph.marker {
		return vNextPublicationReadControlBound(operation.connector, target, label)
	}
	if err := graph.assertPrivateIdentity(operation); err != nil {
		return nil, false, vNextPublicationIdentity{}, err
	}
	head := graph.heads[target]
	if head == nil {
		return nil, false, vNextPublicationIdentity{}, fmt.Errorf("publication control authority has no %s head", target)
	}
	payload, found, identity, err := vNextPublicationReadControlBound(operation.connector, target, label)
	if err != nil {
		return nil, false, vNextPublicationIdentity{}, err
	}
	expected, present, err := head.selected.identity(head.selected.Member)
	if err != nil {
		return nil, false, vNextPublicationIdentity{}, err
	}
	if found != present || (present && identity != expected) {
		return nil, false, vNextPublicationIdentity{}, fmt.Errorf("publication control %q diverges from terminal authority", target)
	}
	return payload, found, identity, nil
}
