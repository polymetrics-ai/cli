package coordination

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"polymetrics.ai/internal/connectors"
	statestore "polymetrics.ai/internal/state"
)

const coordinationStoreSchemaVersion = 1

var (
	// ErrCoordinationStoreSchema refuses state written with a schema this
	// binary cannot interpret. Refusal leaves the file byte-for-byte intact.
	ErrCoordinationStoreSchema = errors.New("coordination store schema is unsupported")
	// ErrRateParkingConflict means a duplicate run identifier carried different
	// durable evidence and therefore cannot safely replace the original record.
	ErrRateParkingConflict = errors.New("parked rate-limit run conflicts with durable evidence")
	// ErrRateParkingClaimLost means another process owns or removed the durable
	// resume claim. The caller must not acknowledge or delete the record.
	ErrRateParkingClaimLost = errors.New("parked rate-limit resume claim is unavailable")
)

type authCohortFileState struct {
	SchemaVersion int                         `json:"schema_version"`
	Records       map[string]AuthCohortHealth `json:"records"`
}

type FileAuthCohortHealthStore struct {
	store statestore.JSONStore[authCohortFileState]
}

// OpenFileAuthCohortHealthStore opens a dependency-free, atomically replaced
// JSON store guarded by a process-crash-safe advisory lock.
func OpenFileAuthCohortHealthStore(path string) (*FileAuthCohortHealthStore, error) {
	if path == "" {
		return nil, errors.New("authentication cohort store path is required")
	}
	s := &FileAuthCohortHealthStore{store: statestore.JSONStore[authCohortFileState]{
		Path: path,
		Initial: func() authCohortFileState {
			return authCohortFileState{SchemaVersion: coordinationStoreSchemaVersion, Records: make(map[string]AuthCohortHealth)}
		},
		Locker: statestore.AdvisoryFileLock{Path: path + ".lock"},
	}}
	state, err := s.store.Load()
	if err != nil {
		return nil, fmt.Errorf("open authentication cohort store: %w", err)
	}
	if err := validateAuthCohortFileState(state); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *FileAuthCohortHealthStore) Load(cohort connectors.AuthCohortKey) (AuthCohortHealth, bool, error) {
	if s == nil {
		return AuthCohortHealth{}, false, errAuthCohortHealthStoreUnavailable
	}
	state, err := s.store.Load()
	if err != nil {
		return AuthCohortHealth{}, false, err
	}
	if err := validateAuthCohortFileState(state); err != nil {
		return AuthCohortHealth{}, false, err
	}
	health, found := state.Records[string(cohort)]
	return health, found, nil
}

func (s *FileAuthCohortHealthStore) Initialize(cohort connectors.AuthCohortKey, initial AuthCohortHealth) (AuthCohortHealth, error) {
	if s == nil {
		return AuthCohortHealth{}, errAuthCohortHealthStoreUnavailable
	}
	var result AuthCohortHealth
	_, err := s.store.Update(func(state authCohortFileState) (authCohortFileState, error) {
		if err := validateAuthCohortFileState(state); err != nil {
			return state, err
		}
		if existing, found := state.Records[string(cohort)]; found {
			result = existing
			return state, nil
		}
		if err := validateAuthCohortHealth(initial); err != nil {
			return state, err
		}
		state.Records[string(cohort)] = initial
		result = initial
		return state, nil
	})
	return result, err
}

func (s *FileAuthCohortHealthStore) CompareAndSwap(cohort connectors.AuthCohortKey, old, next AuthCohortHealth) (bool, error) {
	if s == nil {
		return false, errAuthCohortHealthStoreUnavailable
	}
	swapped := false
	_, err := s.store.Update(func(state authCohortFileState) (authCohortFileState, error) {
		if err := validateAuthCohortFileState(state); err != nil {
			return state, err
		}
		current, found := state.Records[string(cohort)]
		if !found || current != old {
			return state, nil
		}
		if err := validateAuthCohortHealth(next); err != nil {
			return state, err
		}
		state.Records[string(cohort)] = next
		swapped = true
		return state, nil
	})
	return swapped, err
}

func validateAuthCohortFileState(state authCohortFileState) error {
	if state.SchemaVersion != coordinationStoreSchemaVersion {
		return ErrCoordinationStoreSchema
	}
	if state.Records == nil {
		return errors.New("authentication cohort store records are unavailable")
	}
	for key, health := range state.Records {
		if key == "" {
			return errors.New("authentication cohort store key is unavailable")
		}
		if err := validateAuthCohortHealth(health); err != nil {
			return err
		}
	}
	return nil
}

func validateAuthCohortHealth(health AuthCohortHealth) error {
	if health.Epoch == 0 || health.LastFencedEpoch > health.Epoch || (health.Fenced && health.LastFencedEpoch != health.Epoch) {
		return errors.New("authentication cohort health state is invalid")
	}
	return nil
}

type rateParkingFileRecord struct {
	Run        ParkedRateLimitRun `json:"run"`
	ClaimOwner string             `json:"claim_owner,omitempty"`
	ClaimUntil time.Time          `json:"claim_until,omitempty"`
}

type rateParkingFileState struct {
	SchemaVersion int                              `json:"schema_version"`
	Records       map[string]rateParkingFileRecord `json:"records"`
}

type FileRateParkingStore struct {
	store statestore.JSONStore[rateParkingFileState]
}

func OpenFileRateParkingStore(path string) (*FileRateParkingStore, error) {
	if path == "" {
		return nil, errors.New("rate parking store path is required")
	}
	s := &FileRateParkingStore{store: statestore.JSONStore[rateParkingFileState]{
		Path: path,
		Initial: func() rateParkingFileState {
			return rateParkingFileState{SchemaVersion: coordinationStoreSchemaVersion, Records: make(map[string]rateParkingFileRecord)}
		},
		Locker: statestore.AdvisoryFileLock{Path: path + ".lock"},
	}}
	state, err := s.store.Load()
	if err != nil {
		return nil, fmt.Errorf("open rate parking store: %w", err)
	}
	if err := validateRateParkingFileState(state); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *FileRateParkingStore) List() ([]ParkedRateLimitRun, error) {
	state, err := s.load()
	if err != nil {
		return nil, err
	}
	runs := make([]ParkedRateLimitRun, 0, len(state.Records))
	for _, record := range state.Records {
		runs = append(runs, record.Run.Clone())
	}
	sort.Slice(runs, func(i, j int) bool { return runs[i].RunID < runs[j].RunID })
	return runs, nil
}

func (s *FileRateParkingStore) Create(run ParkedRateLimitRun) (ParkedRateLimitRun, bool, error) {
	var result ParkedRateLimitRun
	created := false
	_, err := s.store.Update(func(state rateParkingFileState) (rateParkingFileState, error) {
		if err := validateRateParkingFileState(state); err != nil {
			return state, err
		}
		if existing, found := state.Records[run.RunID]; found {
			if !parkedRateLimitRunEqual(existing.Run, run) {
				return state, ErrRateParkingConflict
			}
			result = existing.Run.Clone()
			return state, nil
		}
		if err := validateParkedRateLimitRun(run); err != nil {
			return state, err
		}
		state.Records[run.RunID] = rateParkingFileRecord{Run: run.Clone()}
		result = run.Clone()
		created = true
		return state, nil
	})
	return result, created, err
}

func (s *FileRateParkingStore) Rearm(run ParkedRateLimitRun, owner string) (ParkedRateLimitRun, error) {
	var result ParkedRateLimitRun
	_, err := s.store.Update(func(state rateParkingFileState) (rateParkingFileState, error) {
		if err := validateRateParkingFileState(state); err != nil {
			return state, err
		}
		if err := validateParkedRateLimitRun(run); err != nil {
			return state, err
		}
		record, found := state.Records[run.RunID]
		if !found || record.ClaimOwner != owner {
			return state, ErrRateParkingClaimLost
		}
		if record.Run.Scope != run.Scope {
			return state, ErrRateParkingConflict
		}
		state.Records[run.RunID] = rateParkingFileRecord{Run: run.Clone()}
		result = run.Clone()
		return state, nil
	})
	return result, err
}

func (s *FileRateParkingStore) HasScope(scope connectors.RateLimitScopeKey) (bool, error) {
	state, err := s.load()
	if err != nil {
		return false, err
	}
	for _, record := range state.Records {
		if record.Run.Scope == scope {
			return true, nil
		}
	}
	return false, nil
}

func (s *FileRateParkingStore) Claim(runID, owner string, now, until time.Time) (ParkedRateLimitRun, bool, time.Time, error) {
	var run ParkedRateLimitRun
	var retryAt time.Time
	claimed := false
	_, err := s.store.Update(func(state rateParkingFileState) (rateParkingFileState, error) {
		if err := validateRateParkingFileState(state); err != nil {
			return state, err
		}
		record, found := state.Records[runID]
		if !found {
			return state, ErrRateParkingClaimLost
		}
		run = record.Run.Clone()
		if record.ClaimOwner != "" && record.ClaimOwner != owner && record.ClaimUntil.After(now) {
			retryAt = record.ClaimUntil
			return state, nil
		}
		record.ClaimOwner = owner
		record.ClaimUntil = until.UTC()
		state.Records[runID] = record
		claimed = true
		return state, nil
	})
	return run, claimed, retryAt, err
}

func (s *FileRateParkingStore) RenewClaim(runID, owner string, until time.Time) (bool, error) {
	renewed := false
	_, err := s.store.Update(func(state rateParkingFileState) (rateParkingFileState, error) {
		if err := validateRateParkingFileState(state); err != nil {
			return state, err
		}
		record, found := state.Records[runID]
		if !found || record.ClaimOwner != owner {
			return state, nil
		}
		record.ClaimUntil = until.UTC()
		state.Records[runID] = record
		renewed = true
		return state, nil
	})
	return renewed, err
}

func (s *FileRateParkingStore) ReleaseClaim(runID, owner string) error {
	_, err := s.store.Update(func(state rateParkingFileState) (rateParkingFileState, error) {
		if err := validateRateParkingFileState(state); err != nil {
			return state, err
		}
		record, found := state.Records[runID]
		if !found {
			return state, nil
		}
		if record.ClaimOwner != owner {
			return state, ErrRateParkingClaimLost
		}
		record.ClaimOwner = ""
		record.ClaimUntil = time.Time{}
		state.Records[runID] = record
		return state, nil
	})
	return err
}

func (s *FileRateParkingStore) Complete(runID, owner string) error {
	_, err := s.store.Update(func(state rateParkingFileState) (rateParkingFileState, error) {
		if err := validateRateParkingFileState(state); err != nil {
			return state, err
		}
		record, found := state.Records[runID]
		if !found || record.ClaimOwner != owner {
			return state, ErrRateParkingClaimLost
		}
		delete(state.Records, runID)
		return state, nil
	})
	return err
}

func (s *FileRateParkingStore) Delete(runID string, now time.Time) error {
	_, err := s.store.Update(func(state rateParkingFileState) (rateParkingFileState, error) {
		if err := validateRateParkingFileState(state); err != nil {
			return state, err
		}
		record, found := state.Records[runID]
		if found && record.ClaimOwner != "" && record.ClaimUntil.After(now) {
			return state, ErrRateParkingClaimLost
		}
		delete(state.Records, runID)
		return state, nil
	})
	return err
}

func (s *FileRateParkingStore) load() (rateParkingFileState, error) {
	if s == nil {
		return rateParkingFileState{}, errRateParkingUnavailable
	}
	state, err := s.store.Load()
	if err != nil {
		return state, err
	}
	return state, validateRateParkingFileState(state)
}

func validateRateParkingFileState(state rateParkingFileState) error {
	if state.SchemaVersion != coordinationStoreSchemaVersion {
		return ErrCoordinationStoreSchema
	}
	if state.Records == nil {
		return errors.New("rate parking store records are unavailable")
	}
	for runID, record := range state.Records {
		if runID != record.Run.RunID {
			return errors.New("rate parking store run identifier is inconsistent")
		}
		if err := validateParkedRateLimitRun(record.Run); err != nil {
			return err
		}
		if (record.ClaimOwner == "") != record.ClaimUntil.IsZero() {
			return errors.New("rate parking store claim is incomplete")
		}
	}
	return nil
}
