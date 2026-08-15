package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/flow"
)

var reversePlanJobReferencePattern = regexp.MustCompile(`^rplan_[0-9a-f]{16}$`)

func resolveManifestJobs(ctx context.Context, a *app.App, manifest flow.FlowManifest) (flow.FlowManifest, error) {
	if a == nil {
		return flow.FlowManifest{}, errors.New("flow job resolution requires an app")
	}
	resolved := manifest
	resolved.Steps = append([]flow.FlowStep(nil), manifest.Steps...)
	for i := range resolved.Steps {
		step := &resolved.Steps[i]
		switch step.Kind {
		case flow.KindSync:
			reference := strings.TrimSpace(step.Job)
			if reference == "" {
				reference = strings.TrimSpace(step.Connection)
			}
			if malformedFlowJobReference(reference) {
				return flow.FlowManifest{}, flowJobReferenceError(manifest, *step, reference, flow.JobReferenceMalformed, nil)
			}
			connection, err := a.GetConnection(reference)
			if err != nil {
				return flow.FlowManifest{}, flowJobReferenceError(manifest, *step, reference, flow.JobReferenceMissing, err)
			}
			if step.Connection != "" && step.Connection != connection.Name {
				return flow.FlowManifest{}, flowJobReferenceError(manifest, *step, reference, flow.JobReferenceUnrecognized, errors.New("connection field disagrees with job reference"))
			}
			step.Job = reference
			step.Connection = connection.Name

		case flow.KindAction:
			reference := strings.TrimSpace(step.Job)
			if malformedFlowJobReference(reference) || !reversePlanJobReferencePattern.MatchString(reference) {
				return flow.FlowManifest{}, flowJobReferenceError(manifest, *step, reference, flow.JobReferenceMalformed, nil)
			}
			plan, err := a.GetReversePlan(reference)
			if err != nil {
				return flow.FlowManifest{}, flowJobReferenceError(manifest, *step, reference, flow.JobReferenceMissing, err)
			}
			if plan.ConnectorCommandOperation != "" || plan.ConnectorCommand != "" {
				return flow.FlowManifest{}, flowJobReferenceError(manifest, *step, reference, flow.JobReferenceUnrecognized, errors.New("connector-command reverse plans are not flow action jobs"))
			}
			if plan.AuthorizationReference == "" {
				return flow.FlowManifest{}, flowJobReferenceError(manifest, *step, reference, flow.JobReferenceUnapproved, errors.New("reverse plan has no standing authorization"))
			}
			if step.ActionCfg == nil {
				return flow.FlowManifest{}, flowJobReferenceError(manifest, *step, reference, flow.JobReferenceMalformed, errors.New("action_cfg is required"))
			}
			if hasInlineActionScope(*step.ActionCfg) {
				return flow.FlowManifest{}, flowJobReferenceError(manifest, *step, reference, flow.JobReferenceMalformed, errors.New("action scope must be derived from the referenced job"))
			}
			readBack := strings.TrimSpace(step.ActionCfg.ReadBackStream)
			if readBack == "" {
				return flow.FlowManifest{}, flowJobReferenceError(manifest, *step, reference, flow.JobReferenceMalformed, errors.New("action_cfg.read_back_stream is required"))
			}
			cfg := &flow.ActionConfig{
				SourceTable: plan.SourceTable, SourceConnection: plan.SourceConnection,
				DestinationConnector: plan.DestinationConnector, DestinationCredential: plan.DestinationCredential,
				DestinationConfig: cloneFlowStringMap(plan.DestinationConfig), DestinationTable: plan.Name,
				Action: plan.Action, Mappings: cloneFlowStringMap(plan.Mappings),
				AuthorizationReference: plan.AuthorizationReference, ReadBackStream: readBack,
				MaxRetries: step.ActionCfg.MaxRetries, BatchSize: step.ActionCfg.BatchSize,
			}
			if err := a.ValidateAuthorizedFlowAction(ctx, app.FlowActionExecutionRequest{
				FlowName: manifest.Name, StepID: step.ID,
				SourceTable: cfg.SourceTable, SourceConnection: cfg.SourceConnection,
				DestinationTable: cfg.DestinationTable, DestinationConnector: cfg.DestinationConnector,
				DestinationCredential: cfg.DestinationCredential, DestinationConfig: cfg.DestinationConfig,
				Action: cfg.Action, Mappings: cfg.Mappings,
				AuthorizationReference: cfg.AuthorizationReference, ReadBackStream: cfg.ReadBackStream,
			}); err != nil {
				return flow.FlowManifest{}, flowJobReferenceError(manifest, *step, reference, flow.JobReferenceUnapproved, err)
			}
			step.ActionCfg = cfg
		}
	}
	return resolved, nil
}

func flowJobReferenceError(manifest flow.FlowManifest, step flow.FlowStep, reference string, reason flow.JobReferenceReason, err error) error {
	return &flow.JobReferenceError{Flow: manifest.Name, StepID: step.ID, Kind: step.Kind, Reference: reference, Reason: reason, Err: err}
}

func malformedFlowJobReference(reference string) bool {
	return reference == "" || reference == "." || reference == ".." || strings.ContainsAny(reference, `/\\`) || strings.IndexFunc(reference, unicode.IsControl) >= 0 || strings.TrimSpace(reference) != reference
}

func hasInlineActionScope(cfg flow.ActionConfig) bool {
	return cfg.SourceTable != "" || cfg.SourceConnection != "" || cfg.DestinationConnector != "" ||
		cfg.DestinationCredential != "" || len(cfg.DestinationConfig) != 0 || cfg.DestinationTable != "" ||
		cfg.Action != "" || len(cfg.Mappings) != 0 || cfg.AuthorizationReference != ""
}

func cloneFlowStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	clone := make(map[string]string, len(values))
	for key, value := range values {
		clone[key] = value
	}
	return clone
}

func saveFlowManifest(projectDir string, manifest flow.FlowManifest) error {
	dir := filepath.Join(projectDir, "flows")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("flow: create manifest directory: %w", err)
	}
	path := filepath.Join(dir, manifest.Name+".json")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("flow %q already exists", manifest.Name)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".flow-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	directory, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}
