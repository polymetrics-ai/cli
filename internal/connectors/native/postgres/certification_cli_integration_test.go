//go:build databaseintegration

package postgres_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors/certify"
	"polymetrics.ai/internal/connectors/native/dbtest"
	native "polymetrics.ai/internal/connectors/native/postgres"
)

// TestPostgresCertificationProfileRunsBuiltBinaryLive proves that the
// definition-owned PostgreSQL profile reaches the shipped CLI's live catalog
// and typed-read stages against a container the test owns. The source count is
// checked independently through pgx after certification; the report's own
// records_read value alone is not accepted as proof.
func TestPostgresCertificationProfileRunsBuiltBinaryLive(t *testing.T) {
	if os.Getenv(postgresCatalogIntegrationEnabledEnv) != "1" {
		t.Skipf("database integration skipped: set %s=1 to run the PostgreSQL Docker or Podman proof", postgresCatalogIntegrationEnabledEnv)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	harness, err := newPostgresCatalogContainerHarness(
		dbtest.Runtime(strings.TrimSpace(os.Getenv(postgresCatalogIntegrationRuntimeEnv))),
		strings.TrimSpace(os.Getenv(postgresCatalogIntegrationEndpointEnv)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Error("PostgreSQL certification container cleanup failed")
		}
	}()

	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatal("PostgreSQL certification container did not start")
	}
	connector := native.New()
	waitForPostgresCatalog(t, ctx, connector, postgresCatalogConfig(t, endpoint, postgresCatalogAlphaSchema))
	source := openPostgresCatalogSource(t, ctx, endpoint)
	defer func() { _ = source.Close(context.WithoutCancel(ctx)) }()
	seedPostgresCatalogs(t, ctx, source)
	seedPostgresCertificationRelation(t, ctx, source)

	var sourceRowsBefore int
	if err := source.QueryRow(ctx, "SELECT COUNT(*) FROM catalog_alpha.certification_events").Scan(&sourceRowsBefore); err != nil {
		t.Fatal("could not count seeded PostgreSQL source rows before certification")
	}
	if sourceRowsBefore != 5 {
		t.Fatalf("seeded PostgreSQL source rows before certification = %d, want 5", sourceRowsBefore)
	}

	binary := buildPostgresCertificationBinary(t, ctx)
	projectRoot := t.TempDir()
	passwordEnv := "PM_POSTGRES_CERTIFICATION_PROFILE_PASSWORD"
	t.Setenv(passwordEnv, "ephemeral-postgres-certification-password")

	args := []string{
		"connectors", "certify", "postgres",
		"--full",
		"--write",
		"--stream", postgresCatalogAlphaSchema + ".certification_events",
		"--config", "host=" + endpoint.Host,
		"--config", "port=" + strconv.Itoa(endpoint.Port),
		"--config", "database=" + postgresCatalogIntegrationDatabase,
		"--config", "username=" + postgresCatalogIntegrationUser,
		"--config", "schema=" + postgresCatalogAlphaSchema,
		"--config", "cursor_field=sequence",
		"--from-env", "password=" + passwordEnv,
		"--json",
		"--root", projectRoot,
	}
	var stdout, stderr bytes.Buffer
	command := exec.CommandContext(ctx, binary, args...)
	command.Stdout = &stdout
	command.Stderr = &stderr
	runErr := command.Run()

	var envelope struct {
		Kind   string         `json:"kind"`
		Report certify.Report `json:"report"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("decode built PostgreSQL certification report: %v", err)
	}
	if runErr != nil {
		t.Fatalf("built PostgreSQL certification binary exit=%d failed_stages=%s stdout_bytes=%d stderr_bytes=%d", exitCode(runErr), failedPostgresCertificationStages(envelope.Report), stdout.Len(), stderr.Len())
	}
	if envelope.Kind != "ConnectorCertification" || !envelope.Report.Passed {
		t.Fatalf("built PostgreSQL certification report = kind %q passed %t, want ConnectorCertification pass", envelope.Kind, envelope.Report.Passed)
	}
	if envelope.Report.Capabilities.Catalog.Result != "pass" || envelope.Report.Capabilities.Catalog.Streams < 1 {
		t.Fatalf("built PostgreSQL catalog capability = %#v, want live pass with streams", envelope.Report.Capabilities.Catalog)
	}
	if envelope.Report.Capabilities.Read.Result != "pass" || envelope.Report.Capabilities.Read.Records != sourceRowsBefore {
		t.Fatalf("built PostgreSQL read capability = %#v, want pass with %d records", envelope.Report.Capabilities.Read, sourceRowsBefore)
	}
	transport := postgresCertificationStage(t, envelope.Report, "declared_transport_pair")
	if !transport.Passed || transport.Status != "pass" {
		t.Fatalf("PostgreSQL declared transport stage = %#v, want live pass", transport)
	}
	if envelope.Report.Capabilities.DeclaredTransport == nil || envelope.Report.Capabilities.DeclaredTransport.Result != "pass" {
		t.Fatalf("PostgreSQL declared transport capability = %#v, want live pass", envelope.Report.Capabilities.DeclaredTransport)
	}
	if got := envelope.Report.Capabilities.DeclaredTransport.Modes; len(got) != 6 {
		t.Fatalf("PostgreSQL declared transport modes = %#v, want all six declared modes", got)
	} else {
		for _, mode := range got {
			if mode.Mode == "" || mode.ApplyStrategy == "" || mode.RecordsRead != sourceRowsBefore || mode.RecordsLoaded != sourceRowsBefore || !mode.CheckpointCommitted {
				t.Fatalf("PostgreSQL transport mode result = %#v, want completed source/read/checkpoint evidence", mode)
			}
			var targetRows int
			statement := "SELECT COUNT(*) FROM " + pgx.Identifier{mode.TargetNamespace, mode.TargetRelation}.Sanitize()
			if err := source.QueryRow(ctx, statement).Scan(&targetRows); err != nil {
				t.Fatalf("independent PostgreSQL target read-back for mode %q: %v", mode.Mode, err)
			}
			if targetRows != mode.RecordsLoaded {
				t.Fatalf("independent PostgreSQL target rows for mode %q = %d, want loaded %d", mode.Mode, targetRows, mode.RecordsLoaded)
			}
		}
	}

	var sourceRowsAfter int
	if err := source.QueryRow(ctx, "SELECT COUNT(*) FROM catalog_alpha.certification_events").Scan(&sourceRowsAfter); err != nil {
		t.Fatal("could not independently count PostgreSQL source rows after certification")
	}
	if sourceRowsAfter != sourceRowsBefore {
		t.Fatalf("PostgreSQL source rows after certification = %d, want unchanged %d", sourceRowsAfter, sourceRowsBefore)
	}
	writePostgresCertificationEvidence(t, ctx, binary, passwordEnv, stdout.Bytes())
	writePostgresTransportCapabilityEvidence(t, ctx, binary, passwordEnv, stdout.Bytes())
}

// seedPostgresCertificationRelation keeps the certification proof focused on
// the declared polling/managed-target pair: a primary key, an explicit
// integer watermark, and ordinary typed payload values. The broader dynamic
// catalog suite independently covers PostgreSQL UUID, JSON, temporal, and
// numerical edge types.
func seedPostgresCertificationRelation(t *testing.T, ctx context.Context, source *pgx.Conn) {
	t.Helper()
	if _, err := source.Exec(ctx, "CREATE TABLE catalog_alpha.certification_events (id bigint NOT NULL, sequence bigint NOT NULL, label text NOT NULL, PRIMARY KEY (id))"); err != nil {
		t.Fatalf("create PostgreSQL certification relation: %v", err)
	}
	if _, err := source.Exec(ctx, "INSERT INTO catalog_alpha.certification_events (id, sequence, label) VALUES (1, 10, 'alpha'), (2, 10, 'bravo'), (3, 11, 'charlie'), (4, 12, 'delta'), (5, 12, 'echo')"); err != nil {
		t.Fatalf("seed PostgreSQL certification relation: %v", err)
	}
}

func buildPostgresCertificationBinary(t *testing.T, ctx context.Context) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate PostgreSQL certification integration test")
	}
	repoRoot := postgresCertificationRepositoryRoot(t, thisFile)
	binary := filepath.Join(t.TempDir(), "pm")
	command := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/pm")
	command.Dir = repoRoot
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("build PostgreSQL certification binary: %v (stderr_bytes=%d)", err, stderr.Len())
	}
	return binary
}

func postgresCertificationRepositoryRoot(t *testing.T, thisFile string) string {
	t.Helper()
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", "..", "..", ".."))
	if _, err := os.Stat(filepath.Join(repoRoot, "go.mod")); err != nil {
		t.Fatalf("PostgreSQL certification integration test could not find repository root: %v", err)
	}
	return repoRoot
}

// writePostgresCertificationEvidence is opt-in because it creates immutable,
// source-controlled proof records from the report this test has already
// independently checked.  The default live test remains read-only apart from
// its container and temporary project.  A fresh evidence run deliberately
// fails rather than replacing an existing record.
func writePostgresCertificationEvidence(t *testing.T, ctx context.Context, binary, passwordEnv string, reportJSON []byte) {
	t.Helper()
	if os.Getenv("POLYMETRICS_WRITE_POSTGRES_CERTIFICATION_EVIDENCE") != "1" {
		return
	}
	binaryBytes, err := os.ReadFile(binary)
	if err != nil {
		t.Fatalf("read built PostgreSQL certification binary for evidence: %v", err)
	}
	reportPath := filepath.Join(t.TempDir(), "postgres-certification-report.json")
	if err := os.WriteFile(reportPath, reportJSON, 0o600); err != nil {
		t.Fatalf("write PostgreSQL certification report for evidence: %v", err)
	}
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate PostgreSQL certification integration test for evidence")
	}
	repoRoot := postgresCertificationRepositoryRoot(t, thisFile)
	checksum := sha256.Sum256(binaryBytes)
	command := exec.CommandContext(ctx, "go", "run", "./cmd/connectorgen",
		"certification-evidence", "transport", "--connector", "postgres",
		"--report", reportPath,
		"--binary-sha", fmt.Sprintf("%x", checksum),
		"--from-env", "password="+passwordEnv,
		"--run-id", "postgres_transport_r1",
		"--record-prefix", "postgres_transport_r1",
		"--repo-root", repoRoot,
	)
	command.Dir = repoRoot
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("write PostgreSQL certification evidence: %v (stdout_bytes=%d stderr_bytes=%d)", err, stdout.Len(), stderr.Len())
	}
	if stdout.String() != "wrote declared transport evidence records: 12\n" {
		t.Fatalf("write PostgreSQL certification evidence stdout=%q", stdout.String())
	}
}

// writePostgresTransportCapabilityEvidence is separately opt-in from the
// per-mode writer. The aggregate report has independently checked read-back
// for every declared target mode, so it can prove capability:write at that
// broader scope. It never relabels the per-mode records.
func writePostgresTransportCapabilityEvidence(t *testing.T, ctx context.Context, binary, passwordEnv string, reportJSON []byte) {
	t.Helper()
	if os.Getenv("POLYMETRICS_WRITE_POSTGRES_CAPABILITY_EVIDENCE") != "1" {
		return
	}
	binaryBytes, err := os.ReadFile(binary)
	if err != nil {
		t.Fatalf("read built PostgreSQL certification binary for capability evidence: %v", err)
	}
	reportPath := filepath.Join(t.TempDir(), "postgres-certification-report.json")
	if err := os.WriteFile(reportPath, reportJSON, 0o600); err != nil {
		t.Fatalf("write PostgreSQL certification report for capability evidence: %v", err)
	}
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate PostgreSQL certification integration test for capability evidence")
	}
	repoRoot := postgresCertificationRepositoryRoot(t, thisFile)
	checksum := sha256.Sum256(binaryBytes)
	command := exec.CommandContext(ctx, "go", "run", "./cmd/connectorgen",
		"certification-evidence", "transport-capability-write", "--connector", "postgres",
		"--report", reportPath,
		"--binary-sha", fmt.Sprintf("%x", checksum),
		"--from-env", "password="+passwordEnv,
		"--run-id", "postgres_transport_r1",
		"--record-prefix", "postgres_transport_r1",
		"--repo-root", repoRoot,
	)
	command.Dir = repoRoot
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("write PostgreSQL transport capability evidence: %v (stdout_bytes=%d stderr_bytes=%d)", err, stdout.Len(), stderr.Len())
	}
	if stdout.String() != "wrote declared transport capability evidence records: 1\n" {
		t.Fatalf("write PostgreSQL transport capability evidence stdout=%q", stdout.String())
	}
}

func postgresCertificationStage(t *testing.T, report certify.Report, name string) certify.StageResult {
	t.Helper()
	for _, stage := range report.Stages {
		if stage.Name == name {
			return stage
		}
	}
	t.Fatalf("PostgreSQL certification report omitted stage %q", name)
	return certify.StageResult{}
}

func exitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

func failedPostgresCertificationStages(report certify.Report) string {
	var names []string
	for _, stage := range report.Stages {
		if !stage.Passed {
			names = append(names, stage.Name+"="+stage.Status)
		}
	}
	return strings.Join(names, ",")
}
