package main

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func TestEnabledContractFinalLaneResults(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	bundle := loadGitLabEnabledContractBundle(t, os.DirFS(defsRoot))

	results := enabledContractFinalLaneResults(os.DirFS(defsRoot), bundle)
	want := map[string]enabledContractFinalLaneStatus{
		"direct_read":     enabledContractFinalLanePartial,
		"direct_write":    enabledContractFinalLanePartial,
		"binary_download": enabledContractFinalLanePartial,
		"binary_upload":   enabledContractFinalLanePartial,
		"etl":             enabledContractFinalLaneComplete,
		"reverse_etl":     enabledContractFinalLanePartial,
		"sync_transport":  enabledContractFinalLaneComplete,
	}
	if len(results) != len(want) {
		t.Fatalf("final lane results = %d, want %d", len(results), len(want))
	}
	for _, result := range results {
		if got, ok := want[result.Name]; !ok || result.Status != got {
			t.Fatalf("final lane result %+v, want status %q", result, got)
		}
		if strings.TrimSpace(result.Reason) == "" || len(result.Citations) == 0 {
			t.Fatalf("final lane result %+v must retain reason and citations", result)
		}
	}
	if findings := checkEnabledConnectorContract(os.DirFS(defsRoot), bundle); len(findings) != 0 {
		t.Fatalf("complete valid contract final gate findings = %+v", findings)
	}
}

func TestEnabledConnectorContractsLoadThroughNormalValidation(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	for _, connector := range []string{"asana", "github", "gitlab"} {
		t.Run(connector, func(t *testing.T) {
			bundle, err := engine.Load(os.DirFS(defsRoot), connector)
			if err != nil {
				t.Fatalf("load %s bundle: %v", connector, err)
			}
			if bundle.EnabledContract == nil {
				t.Fatalf("%s must declare enabled_connector_contract.json", connector)
			}
			if findings := checkEnabledConnectorContract(os.DirFS(defsRoot), bundle); len(findings) != 0 {
				t.Fatalf("%s normal final-contract validation findings = %+v", connector, findings)
			}
		})
	}
}

func TestEnabledConnectorContractsKeepExecutableLanesImplementedWhenSourceMappingIsPartial(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	tests := []struct {
		connector string
		lanes     map[string]string
	}{
		{connector: "asana", lanes: map[string]string{"etl": connectors.EnabledCoveragePartial, "reverse_etl": connectors.EnabledCoverageComplete, "sync_transport": connectors.EnabledCoverageComplete}},
		{connector: "github", lanes: map[string]string{"direct_write": connectors.EnabledCoveragePartial, "etl": connectors.EnabledCoveragePartial, "reverse_etl": connectors.EnabledCoveragePartial, "sync_transport": connectors.EnabledCoveragePartial}},
	}
	for _, test := range tests {
		t.Run(test.connector, func(t *testing.T) {
			bundle, err := engine.Load(os.DirFS(defsRoot), test.connector)
			if err != nil {
				t.Fatalf("load %s bundle: %v", test.connector, err)
			}
			wrapped := enabledContractBundle{bundle: bundle}
			for laneName, coverage := range test.lanes {
				lane := wrapped.lane(laneName)
				if lane.State != connectors.EnabledLaneImplemented {
					t.Fatalf("%s %s state = %q, want implemented: executable runtime proof is not a missing foundation", test.connector, laneName, lane.State)
				}
				if lane.Source.Coverage != coverage {
					t.Fatalf("%s %s coverage = %q, want %q", test.connector, laneName, lane.Source.Coverage, coverage)
				}
				if coverage == connectors.EnabledCoveragePartial && (lane.Source.UnmappedMapping == 0 && lane.Source.MappedUnproven == 0 || lane.Source.DeferredFoundation != 0) {
					t.Fatalf("%s %s source coverage = %+v, want a retained mapping-only partial coverage", test.connector, laneName, lane.Source)
				}
			}
		})
	}

	gitlab, err := engine.Load(os.DirFS(defsRoot), "gitlab")
	if err != nil {
		t.Fatalf("load gitlab bundle: %v", err)
	}
	wrapped := enabledContractBundle{bundle: gitlab}
	for _, laneName := range []string{"binary_download", "binary_upload"} {
		lane := wrapped.lane(laneName)
		if lane.State != connectors.EnabledLaneDeferred || lane.Source.DeferredFoundation == 0 || lane.Source.UnmappedMapping != 0 {
			t.Fatalf("gitlab %s source coverage = %+v, want a named missing execution foundation", laneName, lane.Source)
		}
	}
}

func TestEnabledContractFinalBuildGateRejectsIncompleteLaneEvidence(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	for _, test := range []struct {
		name    string
		fsys    fs.FS
		mutate  func(*enabledContractBundle)
		want    string
		missing string
	}{
		{
			name:    "missing artifact",
			fsys:    hiddenEnabledContractArtifactFS{FS: os.DirFS(defsRoot), hidden: "gitlab/sources/artifacts/53244a720b8509536290e0058c946a246817c775c797df36f4c9aa1225fdf0a4.artifact"},
			want:    "artifact is unavailable",
			missing: "binary_download",
		},
		{
			name: "missing reason",
			fsys: os.DirFS(defsRoot),
			mutate: func(bundle *enabledContractBundle) {
				bundle.lane("direct_write").Reason = ""
			},
			want: "reason is required",
		},
		{
			name: "missing citation",
			fsys: os.DirFS(defsRoot),
			mutate: func(bundle *enabledContractBundle) {
				bundle.lane("direct_write").Citations = nil
			},
			want: "citations and artifacts are required",
		},
		{
			name: "unmatched source row",
			fsys: os.DirFS(defsRoot),
			mutate: func(bundle *enabledContractBundle) {
				lane := bundle.lane("direct_read")
				lane.Source.Expected--
				lane.Source.Implemented--
			},
			want: "partition lane \"direct_read\" reconciled",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			bundle := loadGitLabEnabledContractBundle(t, test.fsys)
			wrapped := enabledContractBundle{bundle: bundle}
			if test.mutate != nil {
				wrapped.bundle.EnabledContract = bundle.EnabledContract.Clone()
				test.mutate(&wrapped)
			}
			if test.missing != "" {
				for _, result := range enabledContractFinalLaneResults(test.fsys, wrapped.bundle) {
					if result.Name == test.missing && result.Status == enabledContractFinalLaneMissing {
						break
					}
					if result.Name == test.missing {
						t.Fatalf("%s lane status = %q, want missing", test.missing, result.Status)
					}
				}
			}
			findings := checkEnabledConnectorContract(test.fsys, wrapped.bundle)
			for _, finding := range findings {
				if strings.Contains(finding.Message, test.want) {
					return
				}
			}
			t.Fatalf("final build findings = %+v, want %q", findings, test.want)
		})
	}
}

func TestEnabledConnectorContractBindsPrimaryV3RetainedEvidence(t *testing.T) {
	const defsRoot = "../../internal/connectors/defs"
	bundle, err := engine.Load(os.DirFS(defsRoot), "asana")
	if err != nil {
		t.Fatalf("load Asana enabled contract: %v", err)
	}
	bundle.EnabledContract = bundle.EnabledContract.Clone()
	bundle.EnabledContract.SourceLock.SHA256 = strings.Repeat("0", 64)
	findings := checkEnabledConnectorContract(os.DirFS(defsRoot), bundle)
	for _, finding := range findings {
		if strings.Contains(finding.Message, "primary source lock identity does not match") {
			return
		}
	}
	t.Fatalf("primary v3 source evidence findings = %+v, want source identity mismatch", findings)
}

func TestEnabledConnectorContractBindsAnyRetainedPrimaryV3Document(t *testing.T) {
	first := []byte("first retained source document")
	second := []byte("second retained source document")
	firstSHA := fmt.Sprintf("%x", sha256.Sum256(first))
	secondSHA := fmt.Sprintf("%x", sha256.Sum256(second))
	lock := sourceImportLock{SchemaVersion: 3}
	lock.Rest.SourceDocuments = []sourceImportRESTDocument{
		{ID: "first", Artifact: sourceImportArtifact{SHA256: firstSHA, Bytes: int64(len(first))}},
		{ID: "second", Artifact: sourceImportArtifact{SHA256: secondSHA, Bytes: int64(len(second))}},
	}
	fsys := fstest.MapFS{
		"asana/sources/artifacts/" + firstSHA + sourceImportRetainedArtifactExtension:  &fstest.MapFile{Data: first},
		"asana/sources/artifacts/" + secondSHA + sourceImportRetainedArtifactExtension: &fstest.MapFile{Data: second},
	}
	err := checkEnabledContractPrimarySourceEvidence(fsys, "asana", lock, connectors.EnabledContractSourceLock{SHA256: secondSHA, Bytes: int64(len(second))})
	if err != nil {
		t.Fatalf("multi-document primary source evidence = %v, want the selected retained identity accepted", err)
	}
}

type enabledContractBundle struct {
	bundle engine.Bundle
}

func (b *enabledContractBundle) lane(name string) *connectors.EnabledConnectorLane {
	for index := range b.bundle.EnabledContract.Lanes {
		if b.bundle.EnabledContract.Lanes[index].Name == name {
			return &b.bundle.EnabledContract.Lanes[index]
		}
	}
	panic("missing lane " + name)
}

func loadGitLabEnabledContractBundle(t *testing.T, fsys fs.FS) engine.Bundle {
	t.Helper()
	bundle, err := engine.Load(fsys, "gitlab")
	if err != nil {
		t.Fatalf("load GitLab enabled contract: %v", err)
	}
	return bundle
}
