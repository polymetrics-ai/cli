package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFoundationEvidenceRejectsEveryStaleGraphIdentityAndArtifact(t *testing.T) {
	fixture := newFoundationEvidenceFixture(t)
	if err := validateFoundationEvidenceWithContext(fixture.manifest, fixture.tdd, fixture.review, fixture.context()); err != nil {
		t.Fatalf("valid evidence closure: %v", err)
	}

	cases := []struct {
		name string
		edit func(map[string]any)
		want string
	}{
		{
			name: "implementation SHA",
			edit: func(document map[string]any) { document["code_sha"] = sha40('9') },
			want: "implementation SHA",
		},
		{
			name: "diff base SHA",
			edit: func(document map[string]any) { document["base_sha"] = sha40('0') },
			want: "diff base",
		},
		{
			name: "self reference",
			edit: func(document map[string]any) { document["code_sha"] = fixture.head },
			want: "self-reference",
		},
	}
	for componentIndex := range fixture.components {
		index := componentIndex
		cases = append(cases,
			struct {
				name string
				edit func(map[string]any)
				want string
			}{
				name: fmt.Sprintf("component %d head", index+1),
				edit: func(document map[string]any) {
					components := document["component_inputs"].([]any)
					components[index].(map[string]any)["sha"] = sha40(byte('5' + index))
				},
				want: "component head",
			},
			struct {
				name string
				edit func(map[string]any)
				want string
			}{
				name: fmt.Sprintf("component %d preserving merge", index+1),
				edit: func(document map[string]any) {
					components := document["component_inputs"].([]any)
					components[index].(map[string]any)["preserving_merge"] = sha40(byte('0' + index))
				},
				want: "preserving merge",
			},
		)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := fixture.mutatedManifest(t, tc.edit)
			err := validateFoundationEvidenceWithContext(raw, fixture.tdd, fixture.review, fixture.context())
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}

	t.Run("closure artifact byte", func(t *testing.T) {
		context := fixture.context()
		context.ReadArtifact = func(path string) ([]byte, error) {
			if path == "website/lib/docs.generated.ts" {
				return []byte("mutated"), nil
			}
			return fixture.artifacts[path], nil
		}
		err := validateFoundationEvidenceWithContext(fixture.manifest, fixture.tdd, fixture.review, context)
		if err == nil || !strings.Contains(err.Error(), "artifact digest") {
			t.Fatalf("error = %v, want artifact digest rejection", err)
		}
	})

	t.Run("closure category", func(t *testing.T) {
		raw := fixture.mutatedManifest(t, func(document map[string]any) {
			closure := document["artifact_closure"].(map[string]any)
			artifacts := closure["artifacts"].([]any)
			closure["artifacts"] = artifacts[:len(artifacts)-1]
		})
		err := validateFoundationEvidenceWithContext(raw, fixture.tdd, fixture.review, fixture.context())
		if err == nil || !strings.Contains(err.Error(), "required artifact category") {
			t.Fatalf("error = %v, want required category rejection", err)
		}
	})

	t.Run("closure subject", func(t *testing.T) {
		raw := fixture.mutatedManifest(t, func(document map[string]any) {
			closure := document["artifact_closure"].(map[string]any)
			closure["subject_fingerprint"] = sha256hex("other subject")
		})
		err := validateFoundationEvidenceWithContext(raw, fixture.tdd, fixture.review, fixture.context())
		if err == nil || !strings.Contains(err.Error(), "subject_fingerprint") {
			t.Fatalf("error = %v, want subject-fingerprint rejection", err)
		}
	})

	t.Run("trailing manifest value", func(t *testing.T) {
		raw := append(append([]byte(nil), fixture.manifest...), []byte("\n{}")...)
		err := validateFoundationEvidenceWithContext(raw, fixture.tdd, fixture.review, fixture.context())
		if err == nil || !strings.Contains(err.Error(), "trailing JSON value") {
			t.Fatalf("error = %v, want trailing-value rejection", err)
		}
	})
}

func TestCertificationSubjectFingerprintIncludesEveryRepositoryComponent(t *testing.T) {
	base, err := newCertificationSubject(certificationSubjectComponents{
		DeclarationsSHA256:      sha256hex("declarations"),
		SourceProjectionSHA256:  sha256hex("source projection"),
		CLICommandMappingSHA256: sha256hex("CLI command mapping"),
		RelevantConfigSHA256:    sha256hex("relevant config"),
		ProofProtocol:           "foundation-certification-proof-v1",
	})
	if err != nil {
		t.Fatalf("newCertificationSubject() = %v", err)
	}

	cases := []struct {
		name string
		edit func(*certificationSubjectComponents)
	}{
		{"declarations", func(value *certificationSubjectComponents) {
			value.DeclarationsSHA256 = sha256hex("changed declarations")
		}},
		{"source projection", func(value *certificationSubjectComponents) {
			value.SourceProjectionSHA256 = sha256hex("changed source projection")
		}},
		{"CLI command mapping", func(value *certificationSubjectComponents) {
			value.CLICommandMappingSHA256 = sha256hex("changed CLI command mapping")
		}},
		{"relevant config", func(value *certificationSubjectComponents) {
			value.RelevantConfigSHA256 = sha256hex("changed relevant config")
		}},
		{"proof protocol", func(value *certificationSubjectComponents) { value.ProofProtocol = "foundation-certification-proof-v2" }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			components := base.Components()
			tc.edit(&components)
			changed, err := newCertificationSubject(components)
			if err != nil {
				t.Fatalf("newCertificationSubject() = %v", err)
			}
			if changed.Fingerprint == base.Fingerprint {
				t.Fatalf("%s mutation retained subject fingerprint %q", tc.name, base.Fingerprint)
			}
		})
	}
}

func TestCertificationSubjectFingerprintExcludesProofTimeBuildProvenance(t *testing.T) {
	base, err := newCertificationSubject(certificationSubjectComponents{
		PMBinarySHA256:          sha256hex("pm binary"),
		PMBuildSHA256:           sha256hex("pm build"),
		DeclarationsSHA256:      sha256hex("declarations"),
		SourceProjectionSHA256:  sha256hex("source projection"),
		CLICommandMappingSHA256: sha256hex("CLI command mapping"),
		RelevantConfigSHA256:    sha256hex("relevant config"),
		ProofProtocol:           certificationSubjectProofProtocol,
	})
	if err != nil {
		t.Fatalf("newCertificationSubject() = %v", err)
	}

	for _, tc := range []struct {
		name string
		edit func(*certificationSubjectComponents)
	}{
		{"pm binary", func(value *certificationSubjectComponents) { value.PMBinarySHA256 = sha256hex("other pm binary") }},
		{"pm build", func(value *certificationSubjectComponents) { value.PMBuildSHA256 = sha256hex("other pm build") }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			components := base.Components()
			tc.edit(&components)
			changed, err := newCertificationSubject(components)
			if err != nil {
				t.Fatalf("newCertificationSubject() = %v", err)
			}
			if changed.Fingerprint != base.Fingerprint {
				t.Fatalf("%s changed deterministic subject fingerprint: got %q want %q", tc.name, changed.Fingerprint, base.Fingerprint)
			}
		})
	}
}

func TestCertificationSubjectCheckRejectsEveryRepositoryInput(t *testing.T) {
	root := t.TempDir()
	inputs := map[string]string{
		"internal/connectors/defs/example/operations.json":       `{"operations":[]}`,
		"internal/connectors/defs/example/sources/contract.json": `{"source":"locked"}`,
		"internal/connectors/defs/example/cli_surface.json":      `{"commands":[]}`,
		"internal/connectors/defs/example/rate_limits.json":      `{"scopes":[]}`,
	}
	for path, contents := range inputs {
		fullPath := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(contents), 0o600); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	var stdout, stderr bytes.Buffer
	if code := runCertificationSubject([]string{"certification-subject", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("write subject exit=%d stderr=%q", code, stderr.String())
	}
	for path := range inputs {
		t.Run(path, func(t *testing.T) {
			fullPath := filepath.Join(root, filepath.FromSlash(path))
			if err := os.WriteFile(fullPath, []byte(`{"changed":true}`), 0o600); err != nil {
				t.Fatalf("mutate %s: %v", path, err)
			}
			stdout.Reset()
			stderr.Reset()
			if code := runCertificationSubject([]string{"certification-subject", root, "--check"}, &stdout, &stderr); code != 1 {
				t.Fatalf("check after %s mutation exit=%d stdout=%q stderr=%q, want stale failure", path, code, stdout.String(), stderr.String())
			}
			if err := os.WriteFile(fullPath, []byte(inputs[path]), 0o600); err != nil {
				t.Fatalf("restore %s: %v", path, err)
			}
		})
	}
}

func TestCertificationSubjectCheckAdvisesValidStaleArtifactWithoutMutatingIt(t *testing.T) {
	root := t.TempDir()
	inputs := map[string]string{
		"internal/connectors/defs/example/operations.json":       `{"operations":[]}`,
		"internal/connectors/defs/example/sources/contract.json": `{"source":"locked"}`,
		"internal/connectors/defs/example/cli_surface.json":      `{"commands":[]}`,
		"internal/connectors/defs/example/rate_limits.json":      `{"scopes":[]}`,
	}
	for path, contents := range inputs {
		fullPath := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(contents), 0o600); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	var stdout, stderr bytes.Buffer
	if code := runCertificationSubject([]string{"certification-subject", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("write subject exit=%d stderr=%q", code, stderr.String())
	}
	artifactPath := filepath.Join(root, filepath.FromSlash(certificationSubjectArtifactPath))
	before, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("read initial subject: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal/connectors/defs/example/rate_limits.json"), []byte(`{"changed":true}`), 0o600); err != nil {
		t.Fatalf("mutate relevant config: %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	if code := runCertificationSubject([]string{"certification-subject", root, "--check"}, &stdout, &stderr); code != 1 {
		t.Fatalf("strict stale check exit=%d stdout=%q stderr=%q, want 1", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "current subject is stale") {
		t.Fatalf("strict stale check stderr=%q, want stale diagnostic", stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := runCertificationSubject([]string{"certification-subject", root, "--check", "--advisory-stale"}, &stdout, &stderr); code != 0 {
		t.Fatalf("advisory stale check exit=%d stdout=%q stderr=%q, want 0", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "advisory: current subject is stale; provenance retained") {
		t.Fatalf("advisory stale check stdout=%q, want provenance diagnostic", stdout.String())
	}
	after, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Fatalf("read advisory subject: %v", err)
	}
	if !bytes.Equal(after, before) {
		t.Fatalf("advisory stale check rewrote subject artifact")
	}
}

func TestCertificationSubjectAdvisoryStaleRejectsInvalidArtifact(t *testing.T) {
	root := t.TempDir()
	for path, contents := range map[string]string{
		"internal/connectors/defs/example/operations.json":       `{"operations":[]}`,
		"internal/connectors/defs/example/sources/contract.json": `{"source":"locked"}`,
		"internal/connectors/defs/example/cli_surface.json":      `{"commands":[]}`,
		"internal/connectors/defs/example/rate_limits.json":      `{"scopes":[]}`,
	} {
		fullPath := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(contents), 0o600); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	artifactPath := filepath.Join(root, filepath.FromSlash(certificationSubjectArtifactPath))
	if err := os.MkdirAll(filepath.Dir(artifactPath), 0o755); err != nil {
		t.Fatalf("mkdir artifact parent: %v", err)
	}
	if err := os.WriteFile(artifactPath, []byte(`not json`), 0o600); err != nil {
		t.Fatalf("write invalid subject: %v", err)
	}

	var stdout, stderr bytes.Buffer
	if code := runCertificationSubject([]string{"certification-subject", root, "--check", "--advisory-stale"}, &stdout, &stderr); code != 1 {
		t.Fatalf("invalid advisory subject exit=%d stdout=%q stderr=%q, want 1", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "current subject artifact must remain valid") {
		t.Fatalf("invalid advisory subject stderr=%q, want validity diagnostic", stderr.String())
	}
}

func TestCertificationSubjectForBinarySeparatesProofTimeProvenance(t *testing.T) {
	root := t.TempDir()
	for path, contents := range map[string]string{
		"internal/connectors/defs/example/operations.json":       `{"operations":[]}`,
		"internal/connectors/defs/example/sources/contract.json": `{"source":"locked"}`,
		"internal/connectors/defs/example/cli_surface.json":      `{"commands":[]}`,
		"internal/connectors/defs/example/rate_limits.json":      `{"scopes":[]}`,
	} {
		fullPath := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(contents), 0o600); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatalf("locate test binary: %v", err)
	}
	subject, provenance, err := certificationSubjectForBinary(root, binary)
	if err != nil {
		t.Fatalf("certificationSubjectForBinary() = %v", err)
	}
	if !evidenceSHA256.MatchString(provenance.PMBinarySHA256) || !evidenceSHA256.MatchString(provenance.PMBuildSHA256) {
		t.Fatalf("proof-time provenance = %#v, want two SHA-256 digests", provenance)
	}
	raw, err := json.Marshal(subject)
	if err != nil {
		t.Fatalf("marshal deterministic subject: %v", err)
	}
	for _, digest := range []string{provenance.PMBinarySHA256, provenance.PMBuildSHA256} {
		if bytes.Contains(raw, []byte(digest)) {
			t.Fatalf("checked-in subject contains proof-time digest %q: %s", digest, raw)
		}
	}
}

func TestCertificationEvidenceBecomesStaleWhenSubjectChanges(t *testing.T) {
	base, err := newCertificationSubject(certificationSubjectComponents{
		DeclarationsSHA256:      sha256hex("declarations"),
		SourceProjectionSHA256:  sha256hex("source projection"),
		CLICommandMappingSHA256: sha256hex("CLI command mapping"),
		RelevantConfigSHA256:    sha256hex("relevant config"),
		ProofProtocol:           "foundation-certification-proof-v1",
	})
	if err != nil {
		t.Fatal(err)
	}
	evidence := acceptedEvidence{SchemaVersion: acceptedEvidenceSchemaVersion, Proof: embeddedEvidenceProof{CertificationSubject: base}}
	live, historical := classifyEvidenceForCertificationSubject([]acceptedEvidence{evidence}, base)
	if len(live) != 1 || len(historical) != 0 {
		t.Fatalf("exact subject classified live=%d historical=%d, want 1/0", len(live), len(historical))
	}

	for _, tc := range []struct {
		name string
		edit func(*certificationSubjectComponents)
	}{
		{"declarations", func(value *certificationSubjectComponents) {
			value.DeclarationsSHA256 = sha256hex("changed declarations")
		}},
		{"source projection", func(value *certificationSubjectComponents) {
			value.SourceProjectionSHA256 = sha256hex("changed source projection")
		}},
		{"CLI command mapping", func(value *certificationSubjectComponents) {
			value.CLICommandMappingSHA256 = sha256hex("changed CLI command mapping")
		}},
		{"relevant config", func(value *certificationSubjectComponents) {
			value.RelevantConfigSHA256 = sha256hex("changed relevant config")
		}},
		{"proof protocol", func(value *certificationSubjectComponents) { value.ProofProtocol = "foundation-certification-proof-v2" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			components := base.Components()
			tc.edit(&components)
			changed, err := newCertificationSubject(components)
			if err != nil {
				t.Fatal(err)
			}
			live, historical := classifyEvidenceForCertificationSubject([]acceptedEvidence{evidence}, changed)
			if len(live) != 0 || len(historical) != 1 {
				t.Fatalf("changed %s classified live=%d historical=%d, want 0/1", tc.name, len(live), len(historical))
			}
		})
	}
}

type foundationEvidenceFixture struct {
	manifest   []byte
	tdd        []byte
	review     []byte
	input      []byte
	head       string
	code       string
	components []map[string]any
	artifacts  map[string][]byte
	ancestry   map[string]bool
}

func newFoundationEvidenceFixture(t *testing.T) foundationEvidenceFixture {
	t.Helper()
	base, code, head, reviewed := sha40('a'), sha40('b'), sha40('c'), sha40('d')
	components := []map[string]any{
		{"issue": 4302, "sha": sha40('e'), "preserving_merge": sha40('4')},
		{"issue": 4303, "sha": sha40('f'), "preserving_merge": sha40('5')},
		{"issue": 4305, "sha": sha40('1'), "preserving_merge": sha40('6')},
		{"issue": 4306, "sha": sha40('2'), "preserving_merge": sha40('7')},
		{"issue": 4307, "sha": sha40('3'), "preserving_merge": sha40('8')},
	}
	artifacts := map[string][]byte{
		certificationSubjectArtifactPath:                                            []byte{},
		"internal/connectors/defs/github/sources/github-operation-source-lock.json": []byte("source"),
		"internal/connectors/defs/github/cli_surface.json":                          []byte("cli"),
		"docs/cli/etl.md":               []byte("docs"),
		"website/lib/docs.generated.ts": []byte("website"),
		"docs/skills/pm-etl/SKILL.md":   []byte("skills"),
		".planning/phases/cli-current-foundations-main-integration-r1/POSTFIX-TDD-LEDGER.md": []byte("ledger"),
		"internal/connectors/defs/github/certification-matrix.json":                          []byte("matrix"),
		"internal/connectors/defs/github/certification-mutation-candidates.json":             []byte("candidate"),
		"internal/connectors/defs/github/certification-sweep.json":                           []byte("sweep"),
		"data/cli-current-foundations-main-integration-r1/report.md":                         []byte("foundation evidence"),
	}
	subject, err := newCertificationSubject(certificationSubjectComponents{
		DeclarationsSHA256:      sha256hex("declarations"),
		SourceProjectionSHA256:  sha256hex("source projection"),
		CLICommandMappingSHA256: sha256hex("CLI command mapping"),
		RelevantConfigSHA256:    sha256hex("relevant config"),
		ProofProtocol:           certificationSubjectProofProtocol,
	})
	if err != nil {
		t.Fatal(err)
	}
	subjectPayload, err := marshalGeneratedJSON(currentCertificationSubjectArtifact{
		SchemaVersion:    certificationSubjectSchemaVersion,
		GeneratedCommand: "go run ./cmd/connectorgen certification-subject",
		Subject:          subject,
	})
	if err != nil {
		t.Fatal(err)
	}
	artifacts[certificationSubjectArtifactPath] = subjectPayload
	categories := []string{"certification_subject", "source", "cli", "docs", "website", "skills", "ledger", "matrix", "candidate", "sweep", "foundation_evidence"}
	paths := []string{
		certificationSubjectArtifactPath,
		"internal/connectors/defs/github/sources/github-operation-source-lock.json",
		"internal/connectors/defs/github/cli_surface.json",
		"docs/cli/etl.md",
		"website/lib/docs.generated.ts",
		"docs/skills/pm-etl/SKILL.md",
		".planning/phases/cli-current-foundations-main-integration-r1/POSTFIX-TDD-LEDGER.md",
		"internal/connectors/defs/github/certification-matrix.json",
		"internal/connectors/defs/github/certification-mutation-candidates.json",
		"internal/connectors/defs/github/certification-sweep.json",
		"data/cli-current-foundations-main-integration-r1/report.md",
	}
	closure := make([]map[string]any, 0, len(paths))
	for index, path := range paths {
		closure = append(closure, map[string]any{"category": categories[index], "path": path, "sha256": sha256hex(string(artifacts[path]))})
	}
	manifest := map[string]any{
		"schema_version": 3, "integration": "foundation", "checkpoint": "final-evidence-closure",
		"code_sha": code, "reviewed_sha": reviewed, "base_sha": base,
		"component_inputs": components, "input_method": "immutable graph",
		"focused_checks": []any{map[string]any{"command": "go test ./cmd/connectorgen", "status": "passed", "assertion": "exact evidence graph", "tests": []string{"TestFoundationEvidenceRejectsEveryStaleGraphIdentityAndArtifact"}, "modes": []string{"foundation_evidence"}}},
		"gate_ledger":    []any{}, "deferred_gates": []string{"CI"}, "credential_handling": "none", "temporary_material": map[string]any{},
		"artifact_closure": map[string]any{"implementation_sha": code, "subject_fingerprint": subject.Fingerprint, "artifacts": closure},
	}
	inputComponents := make([]map[string]any, 0, len(components))
	for _, component := range components {
		inputComponents = append(inputComponents, map[string]any{
			"issue":                    component["issue"],
			"sha":                      component["sha"],
			"integration_merge_commit": component["preserving_merge"],
		})
	}
	input := map[string]any{"inputs": inputComponents}
	manifestRaw, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	inputRaw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	ancestry := map[string]bool{base + "\x00" + code: true, code + "\x00" + head: true}
	for _, component := range components {
		componentSHA := component["sha"].(string)
		merge := component["preserving_merge"].(string)
		ancestry[componentSHA+"\x00"+merge] = true
		ancestry[merge+"\x00"+code] = true
	}
	return foundationEvidenceFixture{
		manifest: manifestRaw, tdd: []byte("# TDD ledger\n"), review: []byte("source_sha: " + reviewed + "\n"), input: inputRaw,
		head: head, code: code, components: components, artifacts: artifacts, ancestry: ancestry,
	}
}

func (fixture foundationEvidenceFixture) context() foundationEvidenceValidationContext {
	return foundationEvidenceValidationContext{
		HeadSHA:       fixture.head,
		ParentSHA:     fixture.code,
		InputManifest: fixture.input,
		WorktreeClean: true,
		IsAncestor: func(ancestor, descendant string) bool {
			return ancestor == descendant || fixture.ancestry[ancestor+"\x00"+descendant]
		},
		ReadArtifact: func(path string) ([]byte, error) {
			value, ok := fixture.artifacts[path]
			if !ok {
				return nil, fmt.Errorf("unknown artifact %q", path)
			}
			return value, nil
		},
	}
}

func (fixture foundationEvidenceFixture) mutatedManifest(t *testing.T, edit func(map[string]any)) []byte {
	t.Helper()
	var document map[string]any
	if err := json.Unmarshal(fixture.manifest, &document); err != nil {
		t.Fatal(err)
	}
	edit(document)
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func sha40(value byte) string { return strings.Repeat(string(value), 40) }

func sha256hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
