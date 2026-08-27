package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// batch1DeferredManifest is the authoritative source-operation census for the
// Batch 1 declaration bridge. It deliberately records source facts and a
// concrete unavailable foundation; it is not an executable-operation catalog.
type batch1DeferredManifest struct {
	SchemaVersion json.RawMessage                  `json:"schema_version"`
	Invariants    batch1DeferredManifestInvariants `json:"invariants"`
	Records       []batch1DeferredRecord           `json:"records"`
}

type batch1DeferredManifestInvariants struct {
	SourceOperations                              int      `json:"source_operations"`
	Runnable                                      int      `json:"runnable"`
	Declarable                                    int      `json:"declarable"`
	Deferred                                      int      `json:"deferred"`
	EverySourceOperationExactlyOnce               bool     `json:"every_source_operation_exactly_once"`
	EveryRecordHasCLIPath                         bool     `json:"every_record_has_cli_path"`
	DeferredHasExactlyOneConcreteMissingComponent bool     `json:"deferred_has_exactly_one_concrete_missing_component"`
	ZeroDenominatorForbidden                      bool     `json:"zero_denominator_forbidden"`
	PolicyOnlyTermsForbiddenAsComponents          []string `json:"policy_only_terms_forbidden_as_components"`
}

type batch1DeferredRecord struct {
	Provider              string                               `json:"provider"`
	RecordKey             string                               `json:"record_key"`
	MappingState          string                               `json:"mapping_state"`
	DeclarationState      string                               `json:"declaration_state"`
	Lane                  string                               `json:"lane"`
	IntendedCLIPath       batch1DeferredIntendedCLIPath        `json:"intended_cli_path"`
	CanonicalTarget       batch1DeferredCanonicalTarget        `json:"canonical_target"`
	Source                batch1DeferredSource                 `json:"source"`
	MissingImplementation *batch1DeferredMissingImplementation `json:"missing_implementation"`
}

type batch1DeferredIntendedCLIPath struct {
	Path                string `json:"path"`
	Source              string `json:"source"`
	CurrentAvailability string `json:"current_availability"`
}

type batch1DeferredCanonicalTarget struct {
	OperationKey string                          `json:"operation_key"`
	Endpoint     batch1DeferredCanonicalEndpoint `json:"endpoint"`
}

type batch1DeferredCanonicalEndpoint struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type batch1DeferredSource struct {
	Protocol            string `json:"protocol"`
	OperationID         string `json:"operation_id"`
	ProviderOperationID string `json:"provider_operation_id"`
	Method              string `json:"method"`
	Path                string `json:"path"`
	Lock                string `json:"source_lock"`
	URL                 string `json:"source_url"`
	Location            string `json:"source_location"`
}

type batch1DeferredMissingImplementation struct {
	Component              string                               `json:"component"`
	Foundation             string                               `json:"foundation"`
	AdditionalFoundation   string                               `json:"additional_foundation"`
	Evidence               string                               `json:"evidence"`
	ProjectionPrerequisite batch1DeferredProjectionPrerequisite `json:"projection_prerequisite"`
}

type batch1DeferredProjectionPrerequisite struct {
	Kind                string `json:"kind"`
	SameCLIPathRequired bool   `json:"same_cli_path_required"`
	Status              string `json:"status"`
}

type batch1DeferredCensus struct {
	SourceOperations int
	Runnable         int
	Declarable       int
	Deferred         int
}

// batch1DeferredCatalogs is the source/declaration triple generated from the
// manifest. Each row is intentionally operation-free: it makes the source
// target discoverable while reserving the actual typed lane for its later
// foundation PR.
type batch1DeferredCatalogs struct {
	Inventory    []declarationAdmissionInventoryOperation
	Sources      []declarationAdmissionSourceOperation
	Declarations []declarationAdmissionDeclaration
}

func decodeBatch1DeferredManifest(raw []byte) (batch1DeferredManifest, error) {
	var manifest batch1DeferredManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return batch1DeferredManifest{}, fmt.Errorf("decode Batch 1 source-operation mapping manifest: %w", err)
	}
	if err := manifest.validate(); err != nil {
		return batch1DeferredManifest{}, err
	}
	return manifest, nil
}

func (manifest batch1DeferredManifest) validate() error {
	if !isBatch1ManifestSchemaVersion(manifest.SchemaVersion) {
		return fmt.Errorf("Batch 1 source-operation mapping manifest: unsupported schema_version %s", strings.TrimSpace(string(manifest.SchemaVersion)))
	}
	if manifest.Invariants.SourceOperations <= 0 || manifest.Invariants.Runnable < 0 || manifest.Invariants.Declarable < 0 || manifest.Invariants.Deferred < 0 {
		return fmt.Errorf("Batch 1 source-operation mapping manifest: invalid published census")
	}
	if !manifest.Invariants.EverySourceOperationExactlyOnce || !manifest.Invariants.EveryRecordHasCLIPath || !manifest.Invariants.DeferredHasExactlyOneConcreteMissingComponent || !manifest.Invariants.ZeroDenominatorForbidden {
		return fmt.Errorf("Batch 1 source-operation mapping manifest: required invariants are not all true")
	}
	if len(manifest.Invariants.PolicyOnlyTermsForbiddenAsComponents) == 0 {
		return fmt.Errorf("Batch 1 source-operation mapping manifest: policy-only foundation terms are required")
	}
	if len(manifest.Records) == 0 {
		return fmt.Errorf("Batch 1 source-operation mapping manifest: records are required")
	}

	seenRecordKeys := make(map[string]struct{}, len(manifest.Records))
	seenCLIIdentities := make(map[string]struct{}, len(manifest.Records))
	for index, record := range manifest.Records {
		prefix := fmt.Sprintf("Batch 1 source-operation mapping manifest record %d", index)
		if strings.TrimSpace(record.RecordKey) == "" {
			return fmt.Errorf("%s: record_key is required", prefix)
		}
		if _, exists := seenRecordKeys[record.RecordKey]; exists {
			return fmt.Errorf("%s: duplicate record_key %q", prefix, record.RecordKey)
		}
		seenRecordKeys[record.RecordKey] = struct{}{}
		if strings.TrimSpace(record.Provider) == "" {
			return fmt.Errorf("%s: provider is required", prefix)
		}
		if !batch1DeferredMappingState(record.MappingState) {
			return fmt.Errorf("%s: unsupported mapping_state %q", prefix, record.MappingState)
		}
		if !batch1DeferredDeclarationState(record.DeclarationState) {
			return fmt.Errorf("%s: unsupported declaration_state %q", prefix, record.DeclarationState)
		}
		if !batch1DeferredLane(record.Lane) {
			return fmt.Errorf("%s: unsupported source lane %q", prefix, record.Lane)
		}
		if strings.TrimSpace(record.IntendedCLIPath.Path) == "" {
			return fmt.Errorf("%s: stable CLI path is required", prefix)
		}
		cliIdentity := record.Provider + "\x00" + record.IntendedCLIPath.Path
		if _, exists := seenCLIIdentities[cliIdentity]; exists {
			return fmt.Errorf("%s: duplicate provider CLI identity %q", prefix, record.IntendedCLIPath.Path)
		}
		seenCLIIdentities[cliIdentity] = struct{}{}
		if err := validateBatch1DeferredTarget(record.CanonicalTarget, prefix); err != nil {
			return err
		}
		if err := validateBatch1DeferredSource(record.Source, record.Provider, record.CanonicalTarget, prefix); err != nil {
			return err
		}
		if record.MappingState == "deferred" {
			if err := validateBatch1DeferredFoundation(record.MissingImplementation, manifest.Invariants.PolicyOnlyTermsForbiddenAsComponents, prefix); err != nil {
				return err
			}
		}
	}
	return nil
}

func (manifest batch1DeferredManifest) reconcilePublishedCensus() (batch1DeferredCensus, error) {
	if err := manifest.validate(); err != nil {
		return batch1DeferredCensus{}, err
	}
	stats := batch1DeferredCensus{SourceOperations: len(manifest.Records)}
	for _, record := range manifest.Records {
		switch record.MappingState {
		case "runnable":
			stats.Runnable++
		case "declarable":
			stats.Declarable++
		case "deferred":
			stats.Deferred++
		}
	}
	want := manifest.Invariants
	if stats.SourceOperations != want.SourceOperations || stats.Runnable != want.Runnable || stats.Declarable != want.Declarable || stats.Deferred != want.Deferred {
		return batch1DeferredCensus{}, fmt.Errorf("Batch 1 source-operation mapping manifest: published census drift: got source_operations=%d runnable=%d declarable=%d deferred=%d, want source_operations=%d runnable=%d declarable=%d deferred=%d", stats.SourceOperations, stats.Runnable, stats.Declarable, stats.Deferred, want.SourceOperations, want.Runnable, want.Declarable, want.Deferred)
	}
	return stats, nil
}

func (manifest batch1DeferredManifest) deferredRecords() []batch1DeferredRecord {
	deferred := make([]batch1DeferredRecord, 0, manifest.Invariants.Deferred)
	for _, record := range manifest.Records {
		if record.MappingState == "deferred" {
			deferred = append(deferred, record)
		}
	}
	return deferred
}

func buildBatch1DeferredCatalogs(manifest batch1DeferredManifest, defsRoot string) (batch1DeferredCatalogs, error) {
	if _, err := manifest.reconcilePublishedCensus(); err != nil {
		return batch1DeferredCatalogs{}, err
	}
	catalogs := batch1DeferredCatalogs{
		Inventory:    make([]declarationAdmissionInventoryOperation, 0, manifest.Invariants.Deferred),
		Sources:      make([]declarationAdmissionSourceOperation, 0, manifest.Invariants.Deferred),
		Declarations: make([]declarationAdmissionDeclaration, 0, manifest.Invariants.Deferred),
	}
	reviewedLocks := map[string]declarationAdmissionReviewedSourceLock{}
	for _, record := range manifest.deferredRecords() {
		lockRelative, err := batch1DeferredRelativeLockPath(record)
		if err != nil {
			return batch1DeferredCatalogs{}, err
		}
		lock, found := reviewedLocks[lockRelative]
		if !found {
			raw, err := os.ReadFile(filepath.Join(defsRoot, filepath.FromSlash(lockRelative)))
			if err != nil {
				return batch1DeferredCatalogs{}, fmt.Errorf("read Batch 1 source lock %q: %w", lockRelative, err)
			}
			lock, err = parseDeclarationAdmissionSourceLock(raw, record.Provider)
			if err != nil {
				return batch1DeferredCatalogs{}, fmt.Errorf("validate Batch 1 source lock %q: %w", lockRelative, err)
			}
			reviewedLocks[lockRelative] = lock
		}
		reviewed, found := lock.Operations[record.Source.OperationID]
		if !found {
			return batch1DeferredCatalogs{}, fmt.Errorf("Batch 1 source record %q is absent from its reviewed source lock", record.RecordKey)
		}
		if err := batch1DeferredReviewedSourceMatchesManifest(record, reviewed); err != nil {
			return batch1DeferredCatalogs{}, err
		}

		destructiveKind := batch1DeferredDestructiveKind(record)
		binding := declarationAdmissionBinding{Kind: "command", ID: record.IntendedCLIPath.Path}
		foundation := batch1DeferredAdmissionFoundation(record)
		catalogs.Inventory = append(catalogs.Inventory, declarationAdmissionInventoryOperation{
			Connector:         record.Provider,
			SourceID:          record.Source.OperationID,
			SourceLock:        lockRelative,
			SourceOperationID: record.Source.OperationID,
		})
		catalogs.Sources = append(catalogs.Sources, declarationAdmissionSourceOperation{
			Connector:           record.Provider,
			ID:                  record.Source.OperationID,
			Protocol:            record.Source.Protocol,
			SourceURL:           record.Source.URL,
			Location:            record.Source.Location,
			ProviderOperationID: record.Source.ProviderOperationID,
			Method:              record.Source.Method,
			Path:                record.Source.Path,
			Binding:             binding,
			DestructiveKind:     destructiveKind,
		})
		catalogs.Declarations = append(catalogs.Declarations, declarationAdmissionDeclaration{
			Connector:  record.Provider,
			SourceID:   record.Source.OperationID,
			Lane:       batch1DeferredAdmissionLane(record.Lane),
			Command:    record.IntendedCLIPath.Path,
			State:      declarationAdmissionStateDeferred,
			Canonical:  declarationAdmissionEndpoint{Method: record.CanonicalTarget.Endpoint.Method, Path: record.CanonicalTarget.Endpoint.Path},
			Binding:    binding,
			Foundation: &foundation,
		})
	}
	sort.Slice(catalogs.Inventory, func(i, j int) bool {
		if catalogs.Inventory[i].Connector != catalogs.Inventory[j].Connector {
			return catalogs.Inventory[i].Connector < catalogs.Inventory[j].Connector
		}
		return catalogs.Inventory[i].SourceID < catalogs.Inventory[j].SourceID
	})
	sort.Slice(catalogs.Sources, func(i, j int) bool {
		if catalogs.Sources[i].Connector != catalogs.Sources[j].Connector {
			return catalogs.Sources[i].Connector < catalogs.Sources[j].Connector
		}
		return catalogs.Sources[i].ID < catalogs.Sources[j].ID
	})
	sort.Slice(catalogs.Declarations, func(i, j int) bool {
		if catalogs.Declarations[i].Connector != catalogs.Declarations[j].Connector {
			return catalogs.Declarations[i].Connector < catalogs.Declarations[j].Connector
		}
		return catalogs.Declarations[i].SourceID < catalogs.Declarations[j].SourceID
	})
	return catalogs, nil
}

func batch1DeferredRelativeLockPath(record batch1DeferredRecord) (string, error) {
	const defsPrefix = "internal/connectors/defs/"
	if !strings.HasPrefix(record.Source.Lock, defsPrefix) {
		return "", fmt.Errorf("Batch 1 source record %q: source_lock is outside connector definitions", record.RecordKey)
	}
	relative := strings.TrimPrefix(record.Source.Lock, defsPrefix)
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(relative)))
	if clean != relative || !strings.HasPrefix(relative, record.Provider+"/sources/") || !strings.HasSuffix(relative, "-operation-source-lock.json") {
		return "", fmt.Errorf("Batch 1 source record %q: source_lock is not a connector-owned operation lock", record.RecordKey)
	}
	return relative, nil
}

func batch1DeferredReviewedSourceMatchesManifest(record batch1DeferredRecord, reviewed declarationAdmissionReviewedOperation) error {
	if reviewed.Protocol != record.Source.Protocol || reviewed.SourceURL != record.Source.URL || reviewed.Location != record.Source.Location ||
		reviewed.ProviderOperationID != record.Source.ProviderOperationID || !strings.EqualFold(reviewed.Method, record.Source.Method) || reviewed.Path != record.Source.Path {
		return fmt.Errorf("Batch 1 source record %q does not match its reviewed provider citation and target", record.RecordKey)
	}
	return nil
}

func batch1DeferredDestructiveKind(record batch1DeferredRecord) string {
	if record.CanonicalTarget.Endpoint.Method == "DELETE" {
		return "delete"
	}
	return "none"
}

func batch1DeferredAdmissionLane(lane string) string {
	switch lane {
	case "binary_read":
		return declarationAdmissionLaneBinaryDownload
	case "binary_write":
		return declarationAdmissionLaneBinaryUpload
	default:
		return lane
	}
}

func batch1DeferredAdmissionFoundation(record batch1DeferredRecord) declarationAdmissionFoundation {
	return declarationAdmissionFoundation{
		ID:        record.MissingImplementation.Foundation,
		Reason:    record.MissingImplementation.Component + " unavailable: " + record.MissingImplementation.Evidence,
		Component: "runtime_executor",
		Evidence:  "runtime_executor_absent",
		Target: declarationAdmissionEndpoint{
			Method: record.CanonicalTarget.Endpoint.Method,
			Path:   record.CanonicalTarget.Endpoint.Path,
		},
	}
}

func isBatch1ManifestSchemaVersion(raw json.RawMessage) bool {
	var version string
	if err := json.Unmarshal(raw, &version); err == nil {
		return version == "batch1-source-operation-mapping-manifest/v1"
	}
	var legacyVersion int
	return json.Unmarshal(raw, &legacyVersion) == nil && legacyVersion == 1
}

func batch1DeferredMappingState(value string) bool {
	return value == "runnable" || value == "declarable" || value == "deferred"
}

func batch1DeferredDeclarationState(value string) bool {
	return value == "implemented" || value == "deferred"
}

func batch1DeferredLane(value string) bool {
	switch value {
	case "direct_read", "direct_write", "etl", "binary_read", "binary_write":
		return true
	default:
		return false
	}
}

func validateBatch1DeferredTarget(target batch1DeferredCanonicalTarget, prefix string) error {
	if strings.TrimSpace(target.OperationKey) == "" {
		return fmt.Errorf("%s: canonical operation_key is required", prefix)
	}
	if !batch1DeferredHTTPMethod(target.Endpoint.Method) {
		return fmt.Errorf("%s: unsupported canonical target method %q", prefix, target.Endpoint.Method)
	}
	if !batch1DeferredRelativePath(target.Endpoint.Path) {
		return fmt.Errorf("%s: canonical target path must be connector-relative", prefix)
	}
	return nil
}

func validateBatch1DeferredSource(source batch1DeferredSource, provider string, target batch1DeferredCanonicalTarget, prefix string) error {
	if source.Protocol != "rest" {
		return fmt.Errorf("%s: unsupported source protocol %q", prefix, source.Protocol)
	}
	if strings.TrimSpace(source.OperationID) == "" {
		return fmt.Errorf("%s: source operation_id is required", prefix)
	}
	if source.Method != target.Endpoint.Method || source.Path != target.Endpoint.Path {
		return fmt.Errorf("%s: source target must match canonical target", prefix)
	}
	if !batch1DeferredRelativePath(source.Path) {
		return fmt.Errorf("%s: source path must be connector-relative", prefix)
	}
	if strings.TrimSpace(source.Lock) == "" || !strings.HasPrefix(source.Lock, "internal/connectors/defs/"+provider+"/sources/") {
		return fmt.Errorf("%s: source_lock must identify the provider-owned source lock", prefix)
	}
	if strings.TrimSpace(source.Location) == "" {
		return fmt.Errorf("%s: source_location is required", prefix)
	}
	parsedURL, err := url.Parse(source.URL)
	if err != nil || parsedURL.Scheme != "https" || parsedURL.Host == "" {
		return fmt.Errorf("%s: source_url must be an exact HTTPS provider citation", prefix)
	}
	return nil
}

func validateBatch1DeferredFoundation(missing *batch1DeferredMissingImplementation, policyOnlyTerms []string, prefix string) error {
	if missing == nil {
		return fmt.Errorf("%s: deferred record requires exactly one concrete missing foundation", prefix)
	}
	if strings.TrimSpace(missing.Foundation) == "" || strings.TrimSpace(missing.AdditionalFoundation) != "" {
		return fmt.Errorf("%s: deferred record requires exactly one concrete missing foundation", prefix)
	}
	for _, forbidden := range policyOnlyTerms {
		if strings.EqualFold(strings.TrimSpace(missing.Foundation), strings.TrimSpace(forbidden)) {
			return fmt.Errorf("%s: policy-only foundation %q is not a concrete missing foundation", prefix, missing.Foundation)
		}
	}
	if missing.Component != "source_descriptor" && missing.Component != "typed_input_schema" && missing.Component != "cli_projection" {
		return fmt.Errorf("%s: unsupported concrete missing component %q", prefix, missing.Component)
	}
	if strings.TrimSpace(missing.Evidence) == "" {
		return fmt.Errorf("%s: foundation evidence is required", prefix)
	}
	prerequisite := missing.ProjectionPrerequisite
	if prerequisite.Kind != "runtime_deferred_command_projection" || !prerequisite.SameCLIPathRequired || prerequisite.Status != "required" {
		return fmt.Errorf("%s: deferred projection prerequisite must require the same stable CLI path", prefix)
	}
	return nil
}

func batch1DeferredHTTPMethod(value string) bool {
	switch value {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	default:
		return false
	}
}

func batch1DeferredRelativePath(value string) bool {
	return strings.HasPrefix(value, "/") && !strings.HasPrefix(value, "//") && !strings.Contains(value, "://") && !strings.ContainsAny(value, "\r\n")
}
