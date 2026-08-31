package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestValidateSourceDescriptorCorrectionProvenance(t *testing.T) {
	t.Run("validates GitLab frozen lock evidence against the declared correction", func(t *testing.T) {
		root, err := repoRoot()
		if err != nil {
			t.Fatalf("repo root: %v", err)
		}
		defsRoot := filepath.Join(root, "internal", "connectors", "defs")
		lockRaw, err := os.ReadFile(filepath.Join(defsRoot, "gitlab", "sources", "gitlab-operation-source-lock.json"))
		if err != nil {
			t.Fatalf("read GitLab source lock: %v", err)
		}
		descriptorRaw, err := os.ReadFile(filepath.Join(defsRoot, "gitlab", "sources", "gitlab-operation-descriptor.json"))
		if err != nil {
			t.Fatalf("read GitLab descriptor: %v", err)
		}
		if err := validateSourceDescriptorCorrectionProvenance(os.DirFS(defsRoot), "gitlab", lockRaw, descriptorRaw); err != nil {
			t.Fatalf("validate GitLab correction provenance: %v", err)
		}
	})

	t.Run("accepts frozen lock-backed correction evidence", func(t *testing.T) {
		fixture := newSourceDescriptorCorrectionFixture(t)
		if err := validateSourceDescriptorCorrectionProvenance(fixture.files, fixture.connector, fixture.lockRaw, fixture.descriptorRaw); err != nil {
			t.Fatalf("validate correction provenance: %v", err)
		}
	})

	t.Run("rejects derived matrix without provenance", func(t *testing.T) {
		fixture := newSourceDescriptorCorrectionFixture(t)
		snapshot := fixture.matrix["source_snapshot"].(map[string]any)
		delete(snapshot, "descriptor_correction_provenance")
		fixture.writeMatrix(t)
		err := validateSourceDescriptorCorrectionProvenance(fixture.files, fixture.connector, fixture.lockRaw, fixture.descriptorRaw)
		if err == nil || !strings.Contains(err.Error(), "declares no correction provenance") {
			t.Fatalf("missing provenance error = %v, want declared provenance rejection", err)
		}
	})

	t.Run("rejects invalid provenance binding", func(t *testing.T) {
		fixture := newSourceDescriptorCorrectionFixture(t)
		fixture.mutateCorrection(t, func(correction map[string]any) {
			correction["kind"] = "invalid"
		})
		err := validateSourceDescriptorCorrectionProvenance(fixture.files, fixture.connector, fixture.lockRaw, fixture.descriptorRaw)
		if err == nil || !strings.Contains(err.Error(), "unsupported identity") {
			t.Fatalf("invalid provenance error = %v, want identity rejection", err)
		}
	})

	t.Run("rejects wrong source operation", func(t *testing.T) {
		fixture := newSourceDescriptorCorrectionFixture(t)
		fixture.mutateCorrection(t, func(correction map[string]any) {
			correction["corrections"].([]any)[0].(map[string]any)["operation_id"] = "missing-operation"
		})
		err := validateSourceDescriptorCorrectionProvenance(fixture.files, fixture.connector, fixture.lockRaw, fixture.descriptorRaw)
		if err == nil || !strings.Contains(err.Error(), "source lock omits operation") {
			t.Fatalf("wrong operation error = %v, want source-lock rejection", err)
		}
	})

	t.Run("rejects wrong descriptor pointer", func(t *testing.T) {
		fixture := newSourceDescriptorCorrectionFixture(t)
		fixture.mutateCorrection(t, func(correction map[string]any) {
			correction["corrections"].([]any)[0].(map[string]any)["descriptor_pointer"] = "/request/body/schema/properties/missing"
		})
		err := validateSourceDescriptorCorrectionProvenance(fixture.files, fixture.connector, fixture.lockRaw, fixture.descriptorRaw)
		if err == nil || !strings.Contains(err.Error(), "descriptor pointer") {
			t.Fatalf("wrong pointer error = %v, want pointer rejection", err)
		}
	})
}

type sourceDescriptorCorrectionFixture struct {
	connector     string
	lockRaw       []byte
	descriptorRaw []byte
	files         fstest.MapFS
	matrix        map[string]any
	correction    map[string]any
}

func newSourceDescriptorCorrectionFixture(t *testing.T) *sourceDescriptorCorrectionFixture {
	t.Helper()
	const connector = "example"
	const lockPath = "sources/example-operation-source-lock.json"
	const descriptorPath = "sources/example-operation-descriptor.json"
	const correctionPath = "sources/example-operation-descriptor-correction-provenance.json"
	const sourceURL = "https://provider.example/openapi.json"
	const sourceSHA256 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	lock := map[string]any{
		"connector": connector,
		"rest": map[string]any{
			"source_url": sourceURL,
			"sha256":     sourceSHA256,
			"bytes":      123,
			"operations": []any{map[string]any{
				"id":              "createWidget",
				"operation_id":    "createWidget",
				"method":          "POST",
				"path":            "/widgets",
				"source_location": "paths[\"/widgets\"].post",
			}},
		},
	}
	lockRaw := marshalSourceDescriptorCorrectionFixture(t, lock)
	descriptor := map[string]any{
		"operations": []any{map[string]any{
			"source_id":    "createWidget",
			"operation_id": "createWidget",
			"method":       "POST",
			"path":         "/widgets",
			"source": map[string]any{
				"url":      sourceURL,
				"sha256":   sourceSHA256,
				"bytes":    123,
				"location": "paths[\"/widgets\"].post",
			},
			"request": map[string]any{
				"body": map[string]any{
					"schema": map[string]any{
						"properties": map[string]any{
							"origin": map[string]any{"type": "string", "minLength": 1},
						},
					},
				},
			},
			"runtime": map[string]any{
				"gaps": []any{map[string]any{
					"foundation": "foundation.retained",
					"location":   "request body",
					"reason":     "dynamic root remains intentionally deferred",
				}},
			},
		}},
		"gaps": []any{},
	}
	descriptorRaw := marshalSourceDescriptorCorrectionFixture(t, descriptor)

	derived := sourceDescriptorCorrectionIdentity{
		Path:        descriptorPath,
		GitBlobSHA1: sourceDescriptorCorrectionGitBlobSHA1(descriptorRaw),
		Bytes:       int64(len(descriptorRaw)),
	}
	lockIdentity := sourceDescriptorCorrectionIdentity{
		Path:        lockPath,
		GitBlobSHA1: sourceDescriptorCorrectionGitBlobSHA1(lockRaw),
		Bytes:       int64(len(lockRaw)),
	}
	removedReason := "the old compiler limitation is no longer present"
	retainedReason := "dynamic root remains intentionally deferred"
	correction := map[string]any{
		"schema_version": 1,
		"connector":      connector,
		"kind":           sourceDescriptorCorrectionKind,
		"base_snapshot": map[string]any{
			"ref":             "example/base",
			"commit":          "base-commit",
			"materialization": "git_archive_byte_identical",
		},
		"target": map[string]any{
			"path": descriptorPath,
			"original": map[string]any{
				"path":          descriptorPath,
				"git_blob_sha1": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				"bytes":         999,
			},
			"derived": map[string]any{
				"path":          derived.Path,
				"git_blob_sha1": derived.GitBlobSHA1,
				"bytes":         derived.Bytes,
			},
		},
		"source_lock": map[string]any{
			"path":       lockPath,
			"source_url": sourceURL,
			"sha256":     sourceSHA256,
			"bytes":      123,
		},
		"corrections": []any{map[string]any{
			"source_id":          "createWidget",
			"operation_id":       "createWidget",
			"method":             "POST",
			"path":               "/widgets",
			"source_location":    "paths[\"/widgets\"].post",
			"descriptor_pointer": "/request/body/schema/properties/origin",
			"source_schema":      map[string]any{"type": "string", "minLength": 1},
			"removed_gaps": []any{map[string]any{
				"scope":         "operation_runtime",
				"foundation":    "foundation.removed",
				"location":      "request body property origin",
				"reason_sha256": sourceDescriptorCorrectionSHA256(removedReason),
			}},
			"retained_gaps": []any{map[string]any{
				"scope":         "operation_runtime",
				"foundation":    "foundation.retained",
				"location":      "request body",
				"reason_sha256": sourceDescriptorCorrectionSHA256(retainedReason),
			}},
			"rationale": "The frozen source lock proves this exact closed schema correction.",
		}},
	}
	correctionRaw := marshalSourceDescriptorCorrectionFixture(t, correction)
	correctionIdentity := sourceDescriptorCorrectionIdentity{
		Path:        correctionPath,
		GitBlobSHA1: sourceDescriptorCorrectionGitBlobSHA1(correctionRaw),
		Bytes:       int64(len(correctionRaw)),
	}
	matrix := map[string]any{
		"connector": connector,
		"source_snapshot": map[string]any{
			"source_snapshot_ref":    "example/base",
			"source_snapshot_commit": "base-commit",
			"materialization":        sourceDescriptorCorrectionKind,
			"base_materialization":   "git_archive_byte_identical",
			"descriptor_correction_provenance": map[string]any{
				"path":          correctionIdentity.Path,
				"git_blob_sha1": correctionIdentity.GitBlobSHA1,
				"bytes":         correctionIdentity.Bytes,
			},
			"base_retained_files": []any{
				map[string]any{"path": lockIdentity.Path, "git_blob_sha1": lockIdentity.GitBlobSHA1, "bytes": lockIdentity.Bytes},
				map[string]any{"path": descriptorPath, "git_blob_sha1": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "bytes": 999},
			},
			"retained_files": []any{
				map[string]any{"path": lockIdentity.Path, "git_blob_sha1": lockIdentity.GitBlobSHA1, "bytes": lockIdentity.Bytes},
				map[string]any{"path": derived.Path, "git_blob_sha1": derived.GitBlobSHA1, "bytes": derived.Bytes},
			},
		},
	}
	fixture := &sourceDescriptorCorrectionFixture{
		connector:     connector,
		lockRaw:       lockRaw,
		descriptorRaw: descriptorRaw,
		files: fstest.MapFS{
			connector + "/" + lockPath:       &fstest.MapFile{Data: lockRaw},
			connector + "/" + descriptorPath: &fstest.MapFile{Data: descriptorRaw},
			connector + "/" + correctionPath: &fstest.MapFile{Data: correctionRaw},
		},
		matrix:     matrix,
		correction: correction,
	}
	fixture.writeMatrix(t)
	return fixture
}

func (fixture *sourceDescriptorCorrectionFixture) mutateCorrection(t *testing.T, mutate func(map[string]any)) {
	t.Helper()
	mutate(fixture.correction)
	correctionRaw := marshalSourceDescriptorCorrectionFixture(t, fixture.correction)
	correctionPath := fixture.connector + "/sources/" + fixture.connector + "-operation-descriptor-correction-provenance.json"
	fixture.files[correctionPath] = &fstest.MapFile{Data: correctionRaw}
	snapshot := fixture.matrix["source_snapshot"].(map[string]any)
	identity := snapshot["descriptor_correction_provenance"].(map[string]any)
	identity["git_blob_sha1"] = sourceDescriptorCorrectionGitBlobSHA1(correctionRaw)
	identity["bytes"] = int64(len(correctionRaw))
	fixture.writeMatrix(t)
}

func (fixture *sourceDescriptorCorrectionFixture) writeMatrix(t *testing.T) {
	t.Helper()
	matrixRaw := marshalSourceDescriptorCorrectionFixture(t, fixture.matrix)
	matrixPath := fixture.connector + "/sources/" + fixture.connector + "-source-lane-matrix.json"
	fixture.files[matrixPath] = &fstest.MapFile{Data: matrixRaw}
}

func marshalSourceDescriptorCorrectionFixture(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	return raw
}

func sourceDescriptorCorrectionSHA256(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
