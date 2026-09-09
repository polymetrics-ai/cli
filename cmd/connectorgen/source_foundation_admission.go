package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"slices"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/manifestidentity"
)

// Admission describes authoring observations. It never changes the manifest
// whose invalid execution generation and missing references it explains.
type sourceFoundationAdmission struct {
	Kind               string                       `json:"kind"`
	CurrentValidation  sourceLaneValidation         `json:"current_validation"`
	CurrentDiagnostics []sourceLaneDiagnostic       `json:"current_diagnostics"`
	MissingCLI         []sourceFoundationMissingCLI `json:"missing_cli"`
}

type sourceFoundationMissingCLI struct {
	Connector           string                           `json:"connector"`
	Path                string                           `json:"path"`
	SourceLock          sourceArtifactPin                `json:"source_lock"`
	PartialGeneration   string                           `json:"partial_generation"`
	CanonicalGeneration string                           `json:"canonical_generation"`
	IntendedCLI         sourceArtifactPin                `json:"intended_cli"`
	Affected            []sourceFoundationMissingCommand `json:"affected"`
	DiagnosticIndices   []int                            `json:"diagnostic_indices"`
}

type sourceFoundationMissingCommand struct {
	Identity            sourceFoundationCell `json:"identity"`
	Binding             sourceLaneTargetRef  `json:"binding"`
	LocalRequirementIDs []string             `json:"local_requirement_ids"`
}

type sourceFoundationAdmissionCustody struct {
	observation sourceFoundationAdmission
	pins        []sourceArtifactPin
	identities  map[string]os.FileInfo
	namespaces  map[string][]string
	cohort      sourceLaneCohort
	annotations []sourceSemanticAnnotation
}

func sourceFoundationCurrentAdmission(manifest sourceLaneManifest) sourceFoundationAdmission {
	return sourceFoundationAdmission{Kind: "current_valid", CurrentValidation: manifest.Validation,
		CurrentDiagnostics: append([]sourceLaneDiagnostic{}, manifest.Diagnostics...), MissingCLI: []sourceFoundationMissingCLI{}}
}

// The canonical view has fresh maps and a freshly loaded complete generation.
// It is used only for full source-reference checks, never as current execution.
func sourceFoundationCanonicalFacts(facts sourceFacts, connector string) (sourceFacts, error) {
	if facts.bindings == nil {
		return facts, fmt.Errorf("canonical source bindings unavailable")
	}
	descriptor, exists := facts.bindings.Canonical[connector]
	if !exists || descriptor.Connector != connector {
		return facts, fmt.Errorf("canonical connector unavailable")
	}
	bundle, err := engine.Load(newVNextExecutionFS(connector, descriptor.Staged.Outputs), connector)
	if err != nil || bundle.Identity.Digest != descriptor.Staged.Identity.Digest {
		return facts, fmt.Errorf("complete canonical generation invalid: %w", err)
	}
	prefix := "internal/connectors/defs/" + connector + "/"
	artifacts := map[string][]byte{}
	authoring := map[string][]byte{}
	for name, raw := range descriptor.Staged.Outputs {
		artifacts[prefix+name] = append([]byte{}, raw...)
	}
	for name, raw := range facts.bindings.Authoring {
		authoring[name] = append([]byte{}, raw...)
	}
	facts.bindings = &sourceLaneBindingInputs{Authoring: authoring, Artifacts: artifacts, Canonical: map[string]vNextCanonicalDescriptor{connector: descriptor}, Bundles: map[string]engine.Bundle{connector: bundle}, Observations: []sourceLaneBindingObservation{}}
	return facts, nil
}

// This validates immutable source identities independently of assessments and
// counts. Only exact actual reference diagnostics can enter the exception.
func sourceFoundationAdmissionSource(manifest sourceLaneManifest, cohort sourceLaneCohort) error {
	if err := validateSourceLaneCohort(cohort); err != nil {
		return err
	}
	expected := map[sourceOperationKey]string{}
	for _, anchor := range cohort.Inventories {
		for _, id := range anchor.ExpectedIDs {
			key := sourceOperationKey{Connector: anchor.Connector, Inventory: anchor.Inventory, ID: id}
			if _, exists := expected[key]; exists {
				return fmt.Errorf("admission duplicate source identity")
			}
			expected[key] = anchor.Class
		}
	}
	if len(expected) != len(manifest.SourceOperations) {
		return fmt.Errorf("admission source membership mismatch")
	}
	for _, row := range manifest.SourceOperations {
		class, exists := expected[row.Source.Key]
		if !exists || class != row.Source.Class || !row.Source.Observed || row.Facts.Status == "unavailable" || len(row.Lanes) != 7 {
			return fmt.Errorf("admission source identity unavailable")
		}
		delete(expected, row.Source.Key)
		for i, lane := range sourceLaneNames() {
			if row.Lanes[i].Lane != lane {
				return fmt.Errorf("admission source lane membership mismatch")
			}
		}
	}
	if len(expected) != 0 || len(validateSourceLaneFactCitations(manifest)) != 0 || len(validateSourceLaneInterpretationEvidence(manifest, manifest.annotationInputs)) != 0 {
		return fmt.Errorf("admission source citation or membership invalid")
	}
	validation := sourceLaneValidation{Status: "valid"}
	for _, d := range manifest.Diagnostics {
		if d.Severity == "error" {
			validation.Errors++
			validation.Status = "invalid"
		}
		if d.Severity == "deficit" {
			validation.Deficits++
		}
		if d.Severity == "error" && (d.Stage != "reference" || d.Code != "execution_generation_mismatch") {
			return fmt.Errorf("admission unrelated source error")
		}
	}
	if validation != manifest.Validation {
		return fmt.Errorf("admission source diagnostic accounting mismatch")
	}
	return nil
}

func sourceFoundationDiagnosticKey(d sourceLaneDiagnostic) string {
	raw, _ := json.Marshal(d)
	return string(raw)
}

// Derive every group and its complete lost command set from current/canonical
// source. Authored foundation requirements have no input to this function.
func deriveSourceFoundationAdmission(manifest sourceLaneManifest, cohort sourceLaneCohort, annotations []sourceSemanticAnnotation) (sourceFoundationAdmission, error) {
	result := sourceFoundationCurrentAdmission(manifest)
	if err := sourceFoundationAdmissionSource(manifest, cohort); err != nil {
		return result, err
	}
	byKey := map[sourceOperationKey]sourceSemanticAnnotation{}
	for _, a := range annotations {
		if _, exists := byKey[a.Key]; exists {
			return result, fmt.Errorf("admission duplicate annotation")
		}
		byKey[a.Key] = a
	}
	groups := map[string]*sourceFoundationMissingCLI{}
	canonicalFacts := map[string]sourceFacts{}
	for _, row := range manifest.SourceOperations {
		binding := row.Facts.bindings
		if binding == nil {
			return result, fmt.Errorf("admission actual bindings unavailable")
		}
		connector := row.Source.Key.Connector
		prefix := "internal/connectors/defs/" + connector + "/"
		missing := prefix + "cli_surface.json"
		descriptor, exists := binding.Canonical[connector]
		if !exists {
			continue
		}
		if _, present := binding.Artifacts[missing]; present {
			continue
		}
		if _, staged := descriptor.Staged.Outputs["cli_surface.json"]; !staged {
			continue
		}
		if groups[connector] != nil {
			continue
		}
		for _, o := range binding.Observations {
			if o.Connector == connector {
				return result, fmt.Errorf("admission invalid partial input: %s", o.Code)
			}
		}
		actual, loaded := binding.Bundles[connector]
		if !loaded {
			return result, fmt.Errorf("admission partial bundle absent")
		}
		partial := map[string][]byte{}
		for name, raw := range descriptor.Staged.Outputs {
			if name == "cli_surface.json" {
				continue
			}
			present, exists := binding.Artifacts[prefix+name]
			if !exists || !bytes.Equal(present, raw) {
				return result, fmt.Errorf("admission supporting artifact missing or changed")
			}
			partial[name] = present
		}
		for name, raw := range binding.Artifacts {
			if strings.HasPrefix(name, prefix) {
				canonical, present := descriptor.Staged.Outputs[strings.TrimPrefix(name, prefix)]
				if !present || !bytes.Equal(raw, canonical) {
					return result, fmt.Errorf("admission extra or mismatched execution artifact")
				}
			}
		}
		partialBundle, err := engine.Load(newVNextExecutionFS(connector, partial), connector)
		if err != nil || partialBundle.Identity != actual.Identity || actual.Identity.Digest == descriptor.Staged.Identity.Digest {
			return result, fmt.Errorf("admission exact partial generation mismatch: %w", err)
		}
		facts, err := sourceFoundationCanonicalFacts(row.Facts, connector)
		if err != nil {
			return result, err
		}
		complete := facts.bindings.Bundles[connector]
		if complete.CLISurface == nil || len(complete.CLISurface.Commands) == 0 {
			return result, fmt.Errorf("admission missing canonical command surface")
		}
		lock, exists := binding.Authoring[prefix+"source.lock.json"]
		if !exists {
			return result, fmt.Errorf("admission canonical source lock absent")
		}
		cli := descriptor.Staged.Outputs["cli_surface.json"]
		groups[connector] = &sourceFoundationMissingCLI{Connector: connector, Path: missing, SourceLock: sourceArtifactPin{Path: prefix + "source.lock.json", SHA256: sourceBytesHash(lock), Bytes: int64(len(lock))}, PartialGeneration: actual.Identity.Digest, CanonicalGeneration: descriptor.Staged.Identity.Digest, IntendedCLI: sourceArtifactPin{Path: missing, SHA256: sourceBytesHash(cli), Bytes: int64(len(cli))}, Affected: []sourceFoundationMissingCommand{}, DiagnosticIndices: []int{}}
		canonicalFacts[connector] = facts
	}
	if len(groups) == 0 {
		return result, fmt.Errorf("admission no exact missing canonical CLI")
	}
	expectedDiagnostics := map[string]int{}
	for _, row := range manifest.SourceOperations {
		group := groups[row.Source.Key.Connector]
		if group == nil {
			continue
		}
		a, exists := byKey[row.Source.Key]
		if !exists {
			continue
		}
		if len(a.MaterializedBindings) != 0 {
			return result, fmt.Errorf("admission conflicting materialized claims")
		}
		facts, err := sourceFoundationCanonicalFacts(row.Facts, row.Source.Key.Connector)
		if err != nil {
			return result, err
		}
		complete := classifySourceLanes(row.Source.Key, facts, &a)
		current := classifySourceLanes(row.Source.Key, row.Facts, &a)
		for i, cell := range complete {
			if cell.Lane != current[i].Lane || cell.Applicability != current[i].Applicability {
				return result, fmt.Errorf("admission canonical projection changed source applicability")
			}
			for _, d := range cell.Diagnostics {
				if d.Stage == "reference" {
					return result, fmt.Errorf("admission canonical reference invalid: %s", d.Code)
				}
			}
		}
		for _, ref := range a.IntendedBindings {
			var completeLane sourceLaneCell
			for _, cell := range complete {
				if cell.Lane == ref.Lane {
					completeLane = cell
				}
			}
			if !slices.ContainsFunc(completeLane.References, func(r sourceLaneTargetRef) bool { return sourceLaneTargetRefEqual(r, ref) }) {
				return result, fmt.Errorf("admission unjoined canonical intended binding")
			}
			if ref.Artifact == group.Path {
				if ref.Kind != "command" {
					return result, fmt.Errorf("admission missing target is not command")
				}
				group.Affected = append(group.Affected, sourceFoundationMissingCommand{Identity: sourceFoundationCell{Key: row.Source.Key, Lane: ref.Lane}, Binding: ref, LocalRequirementIDs: []string{}})
			}
		}
		for _, cell := range current {
			for _, d := range cell.Diagnostics {
				if d.Stage == "reference" {
					valid := d.Code == "execution_generation_mismatch" && d.Severity == "error" || d.Code == "target_absent" && d.Severity == "deficit" && strings.HasPrefix(d.Pointer, group.Path+"#")
					if !valid {
						return result, fmt.Errorf("admission unexplained current reference: %s", d.Code)
					}
					expectedDiagnostics[sourceFoundationDiagnosticKey(d)]++
				}
			}
		}
	}
	for i, d := range manifest.Diagnostics {
		if d.Stage == "reference" {
			key := sourceFoundationDiagnosticKey(d)
			if expectedDiagnostics[key] == 0 {
				return result, fmt.Errorf("admission unrelated reference diagnostic")
			}
			expectedDiagnostics[key]--
			group := groups[d.Key.Connector]
			if group == nil {
				return result, fmt.Errorf("admission reference outside missing group")
			}
			group.DiagnosticIndices = append(group.DiagnosticIndices, i)
		} else if d.Severity == "error" || d.Stage == "collection" || d.Stage == "execution_load" || d.Stage == "canonical_import" {
			return result, fmt.Errorf("admission unrelated input failure")
		}
	}
	for _, count := range expectedDiagnostics {
		if count != 0 {
			return result, fmt.Errorf("admission missing actual diagnostic")
		}
	}
	connectors := make([]string, 0, len(groups))
	for name := range groups {
		connectors = append(connectors, name)
	}
	sort.Strings(connectors)
	for _, name := range connectors {
		group := groups[name]
		bundle := canonicalFacts[name].bindings.Bundles[name]
		if len(group.Affected) != len(bundle.CLISurface.Commands) {
			return result, fmt.Errorf("admission lost surface ownership incomplete")
		}
		for _, command := range bundle.CLISurface.Commands {
			if command.Availability != "implemented" {
				return result, fmt.Errorf("admission unsupported lost command")
			}
			matches := 0
			for _, affected := range group.Affected {
				if affected.Binding.ID == command.Path {
					matches++
				}
			}
			if matches != 1 {
				return result, fmt.Errorf("admission ambiguous lost command ownership")
			}
		}
		result.MissingCLI = append(result.MissingCLI, *group)
	}
	result.Kind = "canonical_intended_missing_cli"
	return result, nil
}

func sourceFoundationAdmissionEqual(a, b sourceFoundationAdmission) bool {
	return reflect.DeepEqual(a, b)
}

// Enumerate only the existing closed execution namespace, with a finite entry
// budget. Directory symlinks are refused rather than followed.
func sourceFoundationExecutionNames(ctx context.Context, root *os.Root, connector string) ([]string, error) {
	prefix := "internal/connectors/defs/" + connector
	if !sourceProofSafePath(prefix) {
		return nil, fmt.Errorf("admission namespace path invalid")
	}
	names := []string{}
	entries := 0
	var walk func(string) error
	walk = func(dir string) (err error) {
		if err := ctx.Err(); err != nil {
			return err
		}
		info, err := root.Lstat(dir)
		if err != nil {
			return err
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("admission namespace directory invalid")
		}
		f, err := root.Open(dir)
		if err != nil {
			return err
		}
		defer func() { err = errors.Join(err, f.Close()) }()
		opened, statErr := f.Stat()
		if statErr != nil || !os.SameFile(info, opened) {
			return fmt.Errorf("admission namespace directory replaced: %w", statErr)
		}
		for {
			batch, readErr := f.ReadDir(256)
			for _, entry := range batch {
				entries++
				if entries > 65536 {
					return fmt.Errorf("admission namespace capacity exceeded")
				}
				if err := ctx.Err(); err != nil {
					return err
				}
				name := dir + "/" + entry.Name()
				relative := strings.TrimPrefix(name, prefix+"/")
				if entry.IsDir() && (relative == "schemas" || strings.HasPrefix(relative, "schemas/")) {
					if err := walk(name); err != nil {
						return err
					}
					continue
				}
				if relative == "schemas" && !entry.IsDir() {
					return fmt.Errorf("admission schema namespace invalid")
				}
				if !manifestidentity.IsExecutionJSONFile(relative) {
					continue
				}
				if _, code := sourceProofFileInfo(root, name); code != "" {
					return fmt.Errorf("admission namespace file invalid")
				}
				names = append(names, name)
			}
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				return readErr
			}
		}
		return nil
	}
	if err := walk(prefix); err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

func sourceFoundationAdmissionPins(manifest sourceLaneManifest) ([]sourceArtifactPin, error) {
	pins := map[string]sourceArtifactPin{}
	add := func(pin sourceArtifactPin) error {
		if old, exists := pins[pin.Path]; exists && old != pin {
			return fmt.Errorf("admission inconsistent source input observations")
		}
		pins[pin.Path] = pin
		return nil
	}
	for _, pin := range manifest.Inputs {
		if err := add(pin); err != nil {
			return nil, err
		}
	}
	seen := map[*sourceLaneBindingInputs]bool{}
	for _, row := range manifest.SourceOperations {
		b := row.Facts.bindings
		if b == nil || seen[b] {
			continue
		}
		seen[b] = true
		for _, files := range []map[string][]byte{b.Authoring, b.Artifacts} {
			for name, raw := range files {
				if err := add(sourceArtifactPin{Path: name, SHA256: sourceBytesHash(raw), Bytes: int64(len(raw))}); err != nil {
					return nil, err
				}
			}
		}
	}
	result := make([]sourceArtifactPin, 0, len(pins))
	for _, pin := range pins {
		result = append(result, pin)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result, nil
}

func observeSourceFoundationAdmission(cache *sourceProofFileCache, manifest sourceLaneManifest, cohort sourceLaneCohort, annotations []sourceSemanticAnnotation) (*sourceFoundationAdmissionCustody, error) {
	observed, err := deriveSourceFoundationAdmission(manifest, cohort, annotations)
	if err != nil {
		return nil, err
	}
	pins, err := sourceFoundationAdmissionPins(manifest)
	if err != nil {
		return nil, err
	}
	result := &sourceFoundationAdmissionCustody{observation: observed, pins: pins, identities: map[string]os.FileInfo{}, namespaces: map[string][]string{}, cohort: cohort, annotations: annotations}
	for _, pin := range pins {
		_, file := cache.get(pin.Path, 64<<20, false)
		if file.code != "" || file.hash != pin.SHA256 || file.size != pin.Bytes {
			return nil, fmt.Errorf("admission required source input changed: %s", pin.Path)
		}
		result.identities[pin.Path] = file.info
	}
	for _, group := range observed.MissingCLI {
		_, file := cache.get(group.Path, 64<<20, false)
		if file.code != "missing" {
			return nil, fmt.Errorf("admission CLI is not initially absent")
		}
		names, err := sourceFoundationExecutionNames(cache.ctx, cache.root, group.Connector)
		if err != nil {
			return nil, err
		}
		expected := []string{}
		prefix := "internal/connectors/defs/" + group.Connector + "/"
		for _, pin := range pins {
			if strings.HasPrefix(pin.Path, prefix) && manifestidentity.IsExecutionJSONFile(strings.TrimPrefix(pin.Path, prefix)) {
				expected = append(expected, pin.Path)
			}
		}
		if !slices.Equal(names, expected) {
			return nil, fmt.Errorf("admission current execution namespace differs")
		}
		result.namespaces[group.Connector] = names
	}
	return result, nil
}

func revalidateSourceFoundationAdmission(cache *sourceProofFileCache, custody *sourceFoundationAdmissionCustody) error {
	if custody == nil {
		return nil
	}
	if err := cache.ctx.Err(); err != nil {
		return err
	}
	for _, pin := range custody.pins {
		_, file := cache.get(pin.Path, 64<<20, false)
		if file.code != "" || file.hash != pin.SHA256 || file.size != pin.Bytes || !os.SameFile(file.info, custody.identities[pin.Path]) {
			return fmt.Errorf("admission source input custody changed: %s", pin.Path)
		}
	}
	for _, group := range custody.observation.MissingCLI {
		if _, code := sourceProofFileInfo(cache.root, group.Path); code != "missing" {
			return fmt.Errorf("admission required CLI absence changed")
		}
		names, err := sourceFoundationExecutionNames(cache.ctx, cache.root, group.Connector)
		if err != nil {
			return err
		}
		if !slices.Equal(names, custody.namespaces[group.Connector]) {
			return fmt.Errorf("admission namespace changed")
		}
	}
	return cache.ctx.Err()
}

func sourceFoundationAdmissionWithRoot(ctx context.Context, repo string, custody *sourceFoundationAdmissionCustody, fn func() error) (err error) {
	defer func() { err = errors.Join(err, ctx.Err()) }()
	if custody == nil {
		return fn()
	}
	root, err := os.OpenRoot(repo)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, root.Close()) }()
	cache := sourceProofFileCache{ctx: ctx, root: root, limits: sourceProofFileLimits{UniqueBytes: 512 << 20, Files: 65536}, files: map[string]*sourceProofFile{}}
	if err = revalidateSourceFoundationAdmission(&cache, custody); err != nil {
		return err
	}
	if err = fn(); err != nil {
		return err
	}
	cache.finalize()
	if err = revalidateSourceFoundationAdmission(&cache, custody); err != nil {
		return err
	}
	return nil
}

func buildSourceFoundationAdmission(ctx context.Context, repo string, manifest sourceLaneManifest, cohort sourceLaneCohort, annotations []sourceSemanticAnnotation, cache *sourceProofFileCache) (sourceLaneManifest, *sourceFoundationAdmissionCustody, error) {
	custody, err := observeSourceFoundationAdmission(cache, manifest, cohort, annotations)
	if err != nil {
		return manifest, nil, err
	}
	// The extra build is the actual source producer under established custody,
	// not a reconstructed valid execution report or an assessment projection.
	replay, err := buildSourceLaneManifest(ctx, repo, cohort, annotations)
	if err != nil {
		return manifest, nil, err
	}
	if len(validateSourceLaneManifest(replay, manifest)) != 0 {
		return manifest, nil, fmt.Errorf("admission source replay changed")
	}
	replayPins, err := sourceFoundationAdmissionPins(replay)
	if err != nil || !reflect.DeepEqual(replayPins, custody.pins) {
		return manifest, nil, fmt.Errorf("admission replay binding inputs changed")
	}
	derived, err := deriveSourceFoundationAdmission(replay, cohort, annotations)
	if err != nil || !sourceFoundationAdmissionEqual(derived, custody.observation) {
		return manifest, nil, fmt.Errorf("admission replay relationship changed: %w", err)
	}
	if err = revalidateSourceFoundationAdmission(cache, custody); err != nil {
		return manifest, nil, err
	}
	return replay, custody, nil
}

func sourceFoundationRequiredAbsent(custody *sourceFoundationAdmissionCustody, name string) bool {
	if custody != nil {
		for _, group := range custody.observation.MissingCLI {
			if group.Path == name {
				return true
			}
		}
	}
	return false
}

func sourceFoundationAdmissionFit(custody *sourceFoundationAdmissionCustody, cell sourceFoundationCell, ref sourceLaneTargetRef) bool {
	if custody == nil {
		return false
	}
	for _, group := range custody.observation.MissingCLI {
		if group.Connector != ref.Connector || group.CanonicalGeneration != ref.Generation {
			continue
		}
		for _, affected := range group.Affected {
			if affected.Identity == cell && affected.Binding.CanonicalID == ref.CanonicalID && len(affected.LocalRequirementIDs) > 0 {
				if ref.Kind != "command" || sourceLaneTargetRefEqual(affected.Binding, ref) {
					return true
				}
			}
		}
	}
	return false
}

// Stage B enriches a newly returned observation. U remains independent of
// editable requirements; unresolved rows account for work but confer no fit.
func sourceFoundationAdmissionAssessments(ctx context.Context, repo string, universe sourceFoundationUniverse, authored []sourceFoundationCellAssessment, cache *sourceProofFileCache) (sourceFoundationUniverse, error) {
	needed := universe.admission != nil
	if !needed {
		rows := map[sourceOperationKey]sourceLaneManifestRow{}
		for _, row := range universe.manifest.SourceOperations {
			rows[row.Source.Key] = row
		}
		for _, cell := range authored {
			for _, r := range cell.Requirements {
				if r.Assessment != "existing_shared_capability" && r.Assessment != "connector_local_configuration" {
					continue
				}
				for _, ref := range r.FitBindings {
					for _, lane := range rows[cell.Key].Lanes {
						if lane.Lane == cell.Lane {
							member := func(other sourceLaneTargetRef) bool { return sourceLaneTargetRefEqual(ref, other) }
							if slices.ContainsFunc(lane.IntendedBindings, member) && !slices.ContainsFunc(lane.References, member) {
								needed = true
							}
						}
					}
				}
			}
		}
	}
	if !needed {
		return universe, nil
	}
	custody := universe.admission
	if custody == nil {
		_, observed, err := buildSourceFoundationAdmissionObserved(ctx, repo, universe.producer, universe.cohort, universe.annotations, cache)
		if err != nil {
			return universe, err
		}
		custody = observed
	}
	copyCustody := *custody
	raw, err := json.Marshal(custody.observation)
	if err != nil {
		return universe, err
	}
	if err = json.Unmarshal(raw, &copyCustody.observation); err != nil {
		return universe, err
	}
	for g := range copyCustody.observation.MissingCLI {
		group := &copyCustody.observation.MissingCLI[g]
		for a := range group.Affected {
			affected := &group.Affected[a]
			affected.LocalRequirementIDs = []string{}
			for _, cell := range authored {
				if (sourceFoundationCell{Key: cell.Key, Lane: cell.Lane}) != affected.Identity {
					continue
				}
				for _, r := range cell.Requirements {
					if r.Assessment != "connector_local_configuration" && r.Assessment != "unresolved" {
						continue
					}
					if !slices.Contains(r.AffectedArtifacts, group.Path) || len(r.EvidenceRequirements) == 0 {
						continue
					}
					if slices.ContainsFunc(r.FitBindings, func(ref sourceLaneTargetRef) bool { return sourceLaneTargetRefEqual(ref, affected.Binding) }) {
						affected.LocalRequirementIDs = append(affected.LocalRequirementIDs, r.ID)
					}
				}
			}
			sort.Strings(affected.LocalRequirementIDs)
			if len(affected.LocalRequirementIDs) == 0 {
				return universe, fmt.Errorf("admission missing local command accounting: %s", affected.Binding.ID)
			}
		}
	}
	universe.admission = &copyCustody
	return universe, nil
}
