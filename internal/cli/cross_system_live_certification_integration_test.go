//go:build databaseintegration

package cli

import (
	"bytes"
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	crossSystemProofEnv = "POLYMETRICS_CROSS_SYSTEM_LIVE_PROOF"
	crossSystemTokenEnv = "POLYMETRICS_GITHUB_TOKEN"
	crossSystemOwner    = "Polymetrics-Cert"
	crossSystemRepo     = "pm-cert-3993-20260810-wz0fru"
)

// TestPMBinaryExecutesLiveCrossSystemPipelines proves the three warehouse-
// mediated routes that the PostgreSQL control lane did not attempt, then
// executes the API-to-API route a second time through the exact payload written
// by schedule install. Every provider assertion uses a separate bounded GitHub
// client, and every task-created provider object is deleted and read back as
// HTTP 404 during unconditional cleanup.
func TestPMBinaryExecutesLiveCrossSystemPipelines(t *testing.T) {
	if os.Getenv("POLYMETRICS_DATABASE_INTEGRATION") != "1" || os.Getenv(crossSystemProofEnv) != "1" {
		t.Skip("cross-system live certification is opt-in")
	}
	token := strings.TrimSpace(os.Getenv(crossSystemTokenEnv))
	if token == "" {
		t.Fatalf("%s=1 requires %s", crossSystemProofEnv, crossSystemTokenEnv)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	fixtureID := crossSystemFixtureID(t)
	labelName := "pm-cert-db-api-" + fixtureID
	commentBefore := "pm-cert-api-api-before-" + fixtureID
	commentAfter := crossSystemOwner + "/" + crossSystemRepo
	github := newCrossSystemGitHubClient(token)
	var commentID int64
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cleanupCancel()
		if commentID != 0 {
			github.cleanupIssueComment(t, cleanupCtx, commentID)
		}
		github.cleanupLabel(t, cleanupCtx, labelName)
	})

	harness := newPostgresTransportHarness(t)
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanupCtx); err != nil {
			t.Errorf("close cross-system PostgreSQL harness: %v", err)
		}
	}()
	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatalf("start cross-system PostgreSQL harness: %v", err)
	}
	admin := waitForPostgresTransport(t, ctx, endpoint, postgresTransportSourceDB, postgresTransportUser)
	defer func() { _ = admin.Close(context.Background()) }()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+postgresTransportTargetDB); err != nil {
		t.Fatalf("create cross-system target database: %v", err)
	}
	if _, err := admin.Exec(ctx, `CREATE TABLE public.label_updates (
		id bigint PRIMARY KEY,
		sequence bigint NOT NULL,
		name text NOT NULL,
		new_name text NOT NULL,
		color text NOT NULL,
		description text NOT NULL
	)`); err != nil {
		t.Fatalf("create database-to-API source table: %v", err)
	}
	if _, err := admin.Exec(ctx, `INSERT INTO public.label_updates
		(id, sequence, name, new_name, color, description)
		VALUES (1, 10, $1, $1, '1d76db', $2)`, labelName, "pm-cert-db-api-updated-"+fixtureID); err != nil {
		t.Fatalf("seed database-to-API source row: %v", err)
	}
	target := waitForPostgresTransport(t, ctx, endpoint, postgresTransportTargetDB, postgresTransportUser)
	defer func() { _ = target.Close(context.Background()) }()

	binary := buildTransportPM(t)
	sha, size := transportBinaryIdentity(t, binary)
	t.Logf("cross-system fresh pm binary sha256=%s size_bytes=%d", sha, size)
	root := filepath.Join(t.TempDir(), "cross-system-project")
	crontab := filepath.Join(root, "pm-cert-crontab")
	t.Setenv("PM_CRONTAB_FILE", crontab)
	mustPostgresTransportPM(t, binary, "", "init", "--root", root, "--json")
	addPostgresTransportCredential(t, binary, root, endpoint, "pg-source", postgresTransportSourceDB)
	addPostgresTransportCredential(t, binary, root, endpoint, "pg-target", postgresTransportTargetDB)
	mustPostgresTransportPM(t, binary, "",
		"credentials", "add", "warehouse-cross-system", "--connector", "warehouse",
		"--config", "path="+filepath.Join(root, ".polymetrics", "warehouse"),
		"--root", root, "--json")
	mustPostgresTransportPM(t, binary, "",
		"credentials", "add", "github-cross-system", "--connector", "github",
		"--config", "owner="+crossSystemOwner,
		"--config", "repo="+crossSystemRepo,
		"--config", "auth_type=token",
		"--config", "max_pages=10",
		"--config", "rate_limit_account=pm-cert-cross-system",
		"--from-env", "token="+crossSystemTokenEnv,
		"--root", root, "--json")

	github.createLabel(t, ctx, crossSystemLabel{
		Name: labelName, Color: "b60205", Description: "pm-cert-db-api-before-" + fixtureID,
	})

	// Route 1: PostgreSQL -> owned warehouse -> typed GitHub update_label.
	mustPostgresTransportPM(t, binary, "",
		"connections", "create", "postgres-to-warehouse",
		"--source", "postgres:pg-source", "--destination", "warehouse:warehouse-cross-system",
		"--stream", "public.label_updates", "--sync-mode", "incremental_upsert",
		"--cursor", "sequence", "--primary-key", "id", "--table", "db_api_label",
		"--root", root, "--json")
	firstDatabaseExtract := runCrossSystemETL(t, binary, root, "postgres-to-warehouse", "public.label_updates", 1)
	assertCrossSystemRun(t, "PostgreSQL to warehouse first run", firstDatabaseExtract.Status, firstDatabaseExtract.RecordsRead, firstDatabaseExtract.RecordsLoaded, 1, 1)
	assertCrossSystemWarehouseRows(t, binary, root, "postgres-to-warehouse", "db_api_label", 1, "name", labelName)
	firstLabelPlan := prepareCrossSystemReversePlan(t, binary, root, []string{
		"reverse", "plan", "pm-cert-db-api-first-" + fixtureID,
		"--source-table", "db_api_label", "--connection", "postgres-to-warehouse",
		"--destination", "github:github-cross-system", "--action", "update_label",
		"--map", "name:name", "--map", "new_name:new_name", "--map", "color:color", "--map", "description:description",
	})
	runCrossSystemReversePlan(t, binary, root, firstLabelPlan)
	label := github.getLabel(t, ctx, labelName, http.StatusOK)
	if label.Name != labelName || label.Color != "1d76db" || label.Description != "pm-cert-db-api-updated-"+fixtureID {
		t.Fatalf("independent database-to-API label read-back = %+v", label)
	}
	secondDatabaseExtract := runCrossSystemETL(t, binary, root, "postgres-to-warehouse", "public.label_updates", 1)
	assertCrossSystemRun(t, "PostgreSQL to warehouse incremental replay", secondDatabaseExtract.Status, secondDatabaseExtract.RecordsRead, secondDatabaseExtract.RecordsLoaded, 0, 0)
	secondLabelPlan := prepareCrossSystemReversePlan(t, binary, root, []string{
		"reverse", "plan", "pm-cert-db-api-second-" + fixtureID,
		"--source-table", "db_api_label", "--connection", "postgres-to-warehouse",
		"--destination", "github:github-cross-system", "--action", "update_label",
		"--map", "name:name", "--map", "new_name:new_name", "--map", "color:color", "--map", "description:description",
	})
	runCrossSystemReversePlan(t, binary, root, secondLabelPlan)
	labelsAfterDatabaseReplay := github.listLabels(t, ctx)
	if got := crossSystemNamedLabelCount(labelsAfterDatabaseReplay, labelName); got != 1 {
		t.Fatalf("database-to-API replay left %d labels named %q, want exactly one", got, labelName)
	}
	t.Logf("ROUTE postgres->github command=pm etl run + pm reverse plan/preview/run first=1/1 replay=0/0 destination_count=1 sample=%s", labelName)

	// Route 2: GitHub labels -> owned warehouse -> PostgreSQL managed target.
	labelsBeforePostgres := github.listLabels(t, ctx)
	mustPostgresTransportPM(t, binary, "",
		"connections", "create", "github-to-postgres",
		"--source", "github:github-cross-system", "--destination", "postgres:pg-target",
		"--stream", "labels", "--sync-mode", "full_overwrite",
		"--primary-key", "name", "--table", "github_labels",
		"--root", root, "--json")
	firstGitHubPostgres := runApprovedPostgresTransportBinary(t, binary, root, "github-to-postgres", "labels", 100)
	assertCrossSystemRun(t, "GitHub to PostgreSQL first run", firstGitHubPostgres.Run.Status, firstGitHubPostgres.Run.RecordsRead, firstGitHubPostgres.Run.RecordsLoaded, len(labelsBeforePostgres), len(labelsBeforePostgres))
	assertCrossSystemPostgresLabel(t, ctx, target, len(labelsBeforePostgres), label)
	unattendedOutput := mustPostgresTransportPM(t, binary, "",
		"etl", "run", "--connection", "github-to-postgres", "--stream", "labels", "--batch-size", "100",
		"--approval-plan", firstGitHubPostgres.PlanID, "--confirm", "destructive", "--root", root, "--json")
	secondGitHubPostgres := decodeCrossSystemETLRun(t, unattendedOutput)
	if secondGitHubPostgres.Status != "completed" {
		t.Fatalf("GitHub to PostgreSQL full overwrite replay status = %q, want completed", secondGitHubPostgres.Status)
	}
	if secondGitHubPostgres.RecordsRead != len(labelsBeforePostgres) || secondGitHubPostgres.RecordsLoaded != len(labelsBeforePostgres) {
		// Keep executing the remaining independent routes after recording this
		// release finding. full_overwrite promises a complete source read and
		// replacement on every run; a zero-row checkpoint skip is not that mode.
		t.Errorf("DEFECT GitHub to PostgreSQL full_overwrite replay read/loaded=%d/%d, want declared full replacement %d/%d",
			secondGitHubPostgres.RecordsRead, secondGitHubPostgres.RecordsLoaded, len(labelsBeforePostgres), len(labelsBeforePostgres))
	}
	replayCount, replaySampleFound := crossSystemPostgresLabelState(t, ctx, target, labelName)
	if replayCount != len(labelsBeforePostgres) || !replaySampleFound {
		t.Errorf("DEFECT independent PostgreSQL full_overwrite replay read-back rows=%d sample_present=%t, want rows=%d sample_present=true",
			replayCount, replaySampleFound, len(labelsBeforePostgres))
	}
	t.Logf("ROUTE github->postgres command=pm etl transport postgres-managed-target plan/preview + pm etl run first=%d/%d replay=%d/%d destination_count=%d sample=%s",
		firstGitHubPostgres.Run.RecordsRead, firstGitHubPostgres.Run.RecordsLoaded,
		secondGitHubPostgres.RecordsRead, secondGitHubPostgres.RecordsLoaded, len(labelsBeforePostgres), labelName)

	// Route 3: GitHub issue comment -> owned warehouse -> typed GitHub
	// update_issue_comment. The credential's exact since boundary makes this a
	// one-record source rather than relying on provider list ordering.
	issueNumber := github.firstIssueNumber(t, ctx)
	commentSince := time.Now().UTC().Add(-2 * time.Second).Format(time.RFC3339)
	comment := github.createIssueComment(t, ctx, issueNumber, commentBefore)
	commentID = comment.ID
	commentsBeforeAPIReplay := github.listIssueCommentsSince(t, ctx, commentSince)
	if len(commentsBeforeAPIReplay) != 1 || commentsBeforeAPIReplay[0].ID != commentID {
		t.Fatalf("bounded GitHub issue-comment source = %+v, want only comment_id=%d", commentsBeforeAPIReplay, commentID)
	}
	mustPostgresTransportPM(t, binary, "",
		"credentials", "add", "github-comment-source", "--connector", "github",
		"--config", "owner="+crossSystemOwner,
		"--config", "repo="+crossSystemRepo,
		"--config", "auth_type=token",
		"--config", "max_pages=1",
		"--config", "since="+commentSince,
		"--config", "rate_limit_account=pm-cert-cross-system-comments",
		"--from-env", "token="+crossSystemTokenEnv,
		"--root", root, "--json")
	mustPostgresTransportPM(t, binary, "",
		"connections", "create", "github-to-warehouse",
		"--source", "github:github-comment-source", "--destination", "warehouse:warehouse-cross-system",
		"--stream", "issue_comments", "--sync-mode", "full_refresh_overwrite",
		"--cursor", "updated_at", "--primary-key", "id", "--table", "api_api_comments",
		"--root", root, "--json")
	firstGitHubExtract := runCrossSystemETL(t, binary, root, "github-to-warehouse", "issue_comments", 10)
	assertCrossSystemRun(t, "GitHub to warehouse first run", firstGitHubExtract.Status, firstGitHubExtract.RecordsRead, firstGitHubExtract.RecordsLoaded, 1, 1)
	assertCrossSystemWarehouseRows(t, binary, root, "github-to-warehouse", "api_api_comments", 1, "id", commentID)
	commentPlan := prepareCrossSystemReversePlan(t, binary, root, []string{
		"reverse", "plan", "pm-cert-api-api-" + fixtureID,
		"--source-table", "api_api_comments", "--connection", "github-to-warehouse",
		"--destination", "github:github-cross-system", "--action", "update_issue_comment", "--limit", "1",
		"--map", "id:comment_id", "--map", "repository:body",
	})
	runCrossSystemReversePlan(t, binary, root, commentPlan)
	updatedComment := github.getIssueComment(t, ctx, commentID, http.StatusOK)
	if updatedComment.ID != commentID || updatedComment.Body != commentAfter {
		t.Fatalf("independent API-to-API comment read-back = %+v", updatedComment)
	}
	if got := crossSystemNamedCommentCount(github.listIssueCommentsSince(t, ctx, commentSince), commentID); got != 1 {
		t.Fatalf("API-to-API first run left %d comments with id %d, want exactly one", got, commentID)
	}
	t.Logf("ROUTE github->github command=pm etl run + pm reverse plan/preview/run source_count=1 destination_count=1 sample_comment_id=%d body=%s", commentID, commentAfter)

	// Route 4 and route 3 replay: persisted ETL + approved reverse job through
	// the exact command installed by the crontab backend.
	flowName := "pm-cert-cross-system-" + fixtureID
	flowFile := writeCrossSystemFlow(t, flowName, commentPlan.ID)
	mustPostgresTransportPM(t, binary, "", "flow", "plan", "--file", flowFile, "--root", root, "--json")
	mustPostgresTransportPM(t, binary, "", "flow", "preview", "--file", flowFile, "--root", root, "--json")
	mustPostgresTransportPM(t, binary, "", "flow", "create", "--file", flowFile, "--root", root, "--json")
	scheduleName := "pm-cert-schedule-" + fixtureID
	mustPostgresTransportPM(t, binary, "",
		"schedule", "create", "--name", scheduleName, "--cron", "0 2 * * *", "--flow", flowName,
		"--root", root, "--json")
	mustPostgresTransportPM(t, binary, "", "schedule", "install", scheduleName, "--crontab", "--root", root, "--json")
	installed, err := os.ReadFile(crontab)
	if err != nil {
		t.Fatalf("read installed pm-cert crontab: %v", err)
	}
	wantEntryPoint := binary + " --root " + root + " flow run " + flowName + " --json"
	if !strings.Contains(string(installed), wantEntryPoint) {
		t.Fatalf("installed schedule omitted exact flow entry point %q", wantEntryPoint)
	}
	for _, forbidden := range []string{token, commentPlan.Token, "approval_token", "--authorization"} {
		if forbidden != "" && strings.Contains(string(installed), forbidden) {
			t.Fatal("installed schedule retained credential or approval material")
		}
	}
	firedOutput := mustPostgresTransportPM(t, binary, "", "--root", root, "flow", "run", flowName, "--json")
	flowResult := decodeCrossSystemFlowResult(t, firedOutput)
	if flowResult.Status != "ok" || len(flowResult.Steps) != 2 || flowResult.Steps[0].RecordsRead != 1 || flowResult.Steps[1].RecordsWritten != 1 || flowResult.Steps[1].PreparedExecutionIdentity == "" {
		t.Fatalf("installed scheduled flow result = %+v", flowResult)
	}
	scheduleOutput := mustPostgresTransportPM(t, binary, "", "schedule", "inspect", scheduleName, "--root", root, "--json")
	assertCrossSystemScheduleState(t, scheduleOutput, flowResult.Steps[1].PreparedExecutionIdentity)
	replayedComment := github.getIssueComment(t, ctx, commentID, http.StatusOK)
	if replayedComment.Body != commentAfter || crossSystemNamedCommentCount(github.listIssueCommentsSince(t, ctx, commentSince), commentID) != 1 {
		t.Fatalf("scheduled API-to-API replay changed destination incorrectly: %+v", replayedComment)
	}
	assertNoCredentialMaterialInTree(t, filepath.Join(root, ".polymetrics"), token)
	t.Logf("SCHEDULE entry_point=%q flow=%s status=ok source_full_overwrite=%d action_updated=1 destination_count=1 sample_comment_id=%d", wantEntryPoint, flowName, flowResult.Steps[0].RecordsRead, commentID)
}

type crossSystemRun struct {
	Status        string `json:"status"`
	RecordsRead   int    `json:"records_read"`
	RecordsLoaded int    `json:"records_loaded"`
}

func runCrossSystemETL(t *testing.T, binary, root, connection, stream string, batchSize int) crossSystemRun {
	t.Helper()
	output := mustPostgresTransportPM(t, binary, "",
		"etl", "run", "--connection", connection, "--stream", stream,
		"--batch-size", strconv.Itoa(batchSize), "--root", root, "--json")
	return decodeCrossSystemETLRun(t, output)
}

func decodeCrossSystemETLRun(t *testing.T, output string) crossSystemRun {
	t.Helper()
	var envelope struct {
		Run crossSystemRun `json:"run"`
	}
	if err := json.Unmarshal([]byte(output), &envelope); err != nil {
		t.Fatalf("decode cross-system ETL run: %v", err)
	}
	return envelope.Run
}

func assertCrossSystemRun(t *testing.T, stage, status string, recordsRead, recordsLoaded, wantRead, wantLoaded int) {
	t.Helper()
	if status != "completed" || recordsRead != wantRead || recordsLoaded != wantLoaded {
		t.Fatalf("%s = status:%s read:%d loaded:%d, want completed/%d/%d", stage, status, recordsRead, recordsLoaded, wantRead, wantLoaded)
	}
}

type crossSystemReversePlan struct {
	ID    string
	Token string
}

func prepareCrossSystemReversePlan(t *testing.T, binary, root string, args []string) crossSystemReversePlan {
	t.Helper()
	args = append(append([]string(nil), args...), "--root", root)
	output := mustPostgresTransportPM(t, binary, "", args...)
	match := regexp.MustCompile(`Created reverse plan (\S+)`).FindStringSubmatch(output)
	if len(match) != 2 {
		t.Fatal("cross-system reverse plan did not return a plan id")
	}
	token := binaryApprovalToken(output)
	if token == "" {
		t.Fatal("cross-system reverse plan did not issue a bounded approval token")
	}
	mustPostgresTransportPM(t, binary, "", "reverse", "preview", match[1], "--root", root, "--json")
	return crossSystemReversePlan{ID: match[1], Token: token}
}

func runCrossSystemReversePlan(t *testing.T, binary, root string, plan crossSystemReversePlan) {
	t.Helper()
	mustPostgresTransportPM(t, binary, plan.Token+"\n",
		"reverse", "run", plan.ID, "--approval-token-stdin", "--root", root, "--json")
}

func assertCrossSystemWarehouseRows(t *testing.T, binary, root, connection, table string, want int, sampleField string, sample any) {
	t.Helper()
	output := mustPostgresTransportPM(t, binary, "",
		"query", "run", "--table", table, "--connection", connection,
		"--limit", strconv.Itoa(want+10), "--root", root, "--json")
	var result struct {
		Count int              `json:"count"`
		Rows  []map[string]any `json:"rows"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("decode cross-system warehouse query: %v", err)
	}
	if result.Count != want || len(result.Rows) != want {
		t.Fatalf("independent warehouse rows for %s/%s = %d/%d, want %d", connection, table, result.Count, len(result.Rows), want)
	}
	for _, row := range result.Rows {
		if crossSystemValuesEqual(row[sampleField], sample) {
			return
		}
	}
	t.Fatalf("independent warehouse rows for %s/%s omitted %s=%v", connection, table, sampleField, sample)
}

func crossSystemValuesEqual(got, want any) bool {
	if fmt.Sprint(got) == fmt.Sprint(want) {
		return true
	}
	gotNumber, gotOK := got.(float64)
	if !gotOK {
		return false
	}
	switch wantNumber := want.(type) {
	case int64:
		return gotNumber == float64(wantNumber)
	case int:
		return gotNumber == float64(wantNumber)
	default:
		return false
	}
}

func assertCrossSystemPostgresLabel(t *testing.T, ctx context.Context, target *pgx.Conn, wantCount int, want crossSystemLabel) {
	t.Helper()
	schema, relation, count := postgresTransportTargetRelation(t, ctx, target)
	if count != wantCount {
		t.Fatalf("independent PostgreSQL label count = %d, want %d", count, wantCount)
	}
	var got crossSystemLabel
	query := "SELECT name, color, COALESCE(description, '') FROM " + pgx.Identifier{schema, relation}.Sanitize() + " WHERE name = $1"
	if err := target.QueryRow(ctx, query, want.Name).Scan(&got.Name, &got.Color, &got.Description); err != nil {
		t.Fatalf("independently read named PostgreSQL label %q: %v", want.Name, err)
	}
	if got != want {
		t.Fatalf("independent PostgreSQL named sample = %+v, want %+v", got, want)
	}
}

func crossSystemPostgresLabelState(t *testing.T, ctx context.Context, target *pgx.Conn, name string) (int, bool) {
	t.Helper()
	schema, relation, count := postgresTransportTargetRelation(t, ctx, target)
	var found bool
	query := "SELECT EXISTS (SELECT 1 FROM " + pgx.Identifier{schema, relation}.Sanitize() + " WHERE name = $1)"
	if err := target.QueryRow(ctx, query, name).Scan(&found); err != nil {
		t.Fatalf("independently inspect replayed PostgreSQL label %q: %v", name, err)
	}
	return count, found
}

func writeCrossSystemFlow(t *testing.T, name, job string) string {
	t.Helper()
	manifest := map[string]any{
		"version":     1,
		"name":        name,
		"description": "pm-cert GitHub issue-comment warehouse round trip",
		"steps": []any{
			map[string]any{
				"id": "extract-comment", "kind": "sync", "connection": "github-to-warehouse",
				"streams": []string{"issue_comments"}, "in": []string{}, "out": []string{"api_api_comments"},
			},
			map[string]any{
				"id": "update-comment", "kind": "action", "job": job,
				"in": []string{"api_api_comments"}, "out": []string{},
				"action_cfg": map[string]any{"read_back_stream": "issue_comments"},
			},
		},
	}
	payload, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("encode cross-system flow: %v", err)
	}
	path := filepath.Join(t.TempDir(), name+".json")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatalf("write cross-system flow: %v", err)
	}
	return path
}

type crossSystemFlowResult struct {
	Status string `json:"status"`
	Steps  []struct {
		RecordsRead               int    `json:"records_read"`
		RecordsWritten            int    `json:"records_written"`
		PreparedExecutionIdentity string `json:"prepared_execution_identity"`
	} `json:"steps"`
}

func decodeCrossSystemFlowResult(t *testing.T, output string) crossSystemFlowResult {
	t.Helper()
	var result crossSystemFlowResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("decode installed cross-system flow: %v", err)
	}
	return result
}

func assertCrossSystemScheduleState(t *testing.T, output, wantIdentity string) {
	t.Helper()
	var inspected struct {
		Status struct {
			Status   string `json:"status"`
			LastFire struct {
				FlowStatus                  string   `json:"flow_status"`
				PreparedExecutionIdentities []string `json:"prepared_execution_identities"`
			} `json:"last_fire"`
		} `json:"status"`
	}
	if err := json.Unmarshal([]byte(output), &inspected); err != nil {
		t.Fatalf("decode cross-system schedule inspection: %v", err)
	}
	identities := inspected.Status.LastFire.PreparedExecutionIdentities
	if inspected.Status.Status != "succeeded" || inspected.Status.LastFire.FlowStatus != "ok" || len(identities) != 1 || identities[0] != wantIdentity {
		t.Fatalf("durable cross-system schedule state = %+v", inspected.Status)
	}
}

func crossSystemFixtureID(t *testing.T) string {
	t.Helper()
	var raw [6]byte
	if _, err := cryptorand.Read(raw[:]); err != nil {
		t.Fatalf("generate cross-system fixture id: %v", err)
	}
	return hex.EncodeToString(raw[:])
}

type crossSystemGitHubClient struct {
	token  string
	client *http.Client
}

type crossSystemLabel struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

type crossSystemIssueComment struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	UpdatedAt string `json:"updated_at"`
}

func newCrossSystemGitHubClient(token string) *crossSystemGitHubClient {
	return &crossSystemGitHubClient{token: token, client: &http.Client{Timeout: 30 * time.Second}}
}

func (c *crossSystemGitHubClient) request(ctx context.Context, method, path string, body any) (*http.Response, []byte, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, nil, fmt.Errorf("encode bounded GitHub fixture request: %w", err)
		}
		reader = bytes.NewReader(payload)
	}
	request, err := http.NewRequestWithContext(ctx, method, "https://api.github.com"+path, reader)
	if err != nil {
		return nil, nil, fmt.Errorf("build bounded GitHub fixture request: %w", err)
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	request.Header.Set("User-Agent", "polymetrics-cross-system-certification")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := c.client.Do(request)
	if err != nil {
		return nil, nil, fmt.Errorf("execute bounded GitHub fixture request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	payload, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return nil, nil, fmt.Errorf("read bounded GitHub fixture response: %w", err)
	}
	return response, payload, nil
}

func (c *crossSystemGitHubClient) createLabel(t *testing.T, ctx context.Context, label crossSystemLabel) {
	t.Helper()
	response, _, err := c.request(ctx, http.MethodPost, crossSystemRepoPath("/labels"), label)
	if err != nil || response.StatusCode != http.StatusCreated {
		t.Fatalf("create pm-cert label: status=%s err=%v", crossSystemStatus(response), err)
	}
}

func (c *crossSystemGitHubClient) getLabel(t *testing.T, ctx context.Context, name string, wantStatus int) crossSystemLabel {
	t.Helper()
	response, payload, err := c.request(ctx, http.MethodGet, crossSystemRepoPath("/labels/"+url.PathEscape(name)), nil)
	if err != nil || response.StatusCode != wantStatus {
		t.Fatalf("read pm-cert label %q: status=%s err=%v", name, crossSystemStatus(response), err)
	}
	if wantStatus == http.StatusNotFound {
		return crossSystemLabel{}
	}
	var label crossSystemLabel
	if err := json.Unmarshal(payload, &label); err != nil {
		t.Fatalf("decode pm-cert label %q: %v", name, err)
	}
	return label
}

func (c *crossSystemGitHubClient) listLabels(t *testing.T, ctx context.Context) []crossSystemLabel {
	t.Helper()
	var labels []crossSystemLabel
	for page := 1; page <= 10; page++ {
		path := crossSystemRepoPath("/labels?per_page=100&page=" + strconv.Itoa(page))
		response, payload, err := c.request(ctx, http.MethodGet, path, nil)
		if err != nil || response.StatusCode != http.StatusOK {
			t.Fatalf("list pm-cert repository labels page %d: status=%s err=%v", page, crossSystemStatus(response), err)
		}
		var batch []crossSystemLabel
		if err := json.Unmarshal(payload, &batch); err != nil {
			t.Fatalf("decode pm-cert repository labels page %d: %v", page, err)
		}
		labels = append(labels, batch...)
		if len(batch) < 100 {
			return labels
		}
	}
	t.Fatal("pm-cert repository labels exceed bounded ten-page proof")
	return nil
}

func (c *crossSystemGitHubClient) firstIssueNumber(t *testing.T, ctx context.Context) int {
	t.Helper()
	response, payload, err := c.request(ctx, http.MethodGet, crossSystemRepoPath("/issues?state=all&per_page=100"), nil)
	if err != nil || response.StatusCode != http.StatusOK {
		t.Fatalf("list pm-cert fixture issues: status=%s err=%v", crossSystemStatus(response), err)
	}
	var issues []struct {
		Number      int             `json:"number"`
		PullRequest json.RawMessage `json:"pull_request"`
	}
	if err := json.Unmarshal(payload, &issues); err != nil {
		t.Fatalf("decode pm-cert fixture issues: %v", err)
	}
	for _, issue := range issues {
		if issue.Number > 0 && len(issue.PullRequest) == 0 {
			return issue.Number
		}
	}
	t.Fatal("pm-cert fixture repository has no issue available for a deletable comment")
	return 0
}

func (c *crossSystemGitHubClient) createIssueComment(t *testing.T, ctx context.Context, issue int, body string) crossSystemIssueComment {
	t.Helper()
	response, payload, err := c.request(ctx, http.MethodPost,
		crossSystemRepoPath("/issues/"+strconv.Itoa(issue)+"/comments"), map[string]string{"body": body})
	if err != nil || response.StatusCode != http.StatusCreated {
		t.Fatalf("create pm-cert issue comment: status=%s err=%v", crossSystemStatus(response), err)
	}
	var comment crossSystemIssueComment
	if err := json.Unmarshal(payload, &comment); err != nil || comment.ID == 0 {
		t.Fatalf("decode created pm-cert issue comment: %v", err)
	}
	return comment
}

func (c *crossSystemGitHubClient) getIssueComment(t *testing.T, ctx context.Context, id int64, wantStatus int) crossSystemIssueComment {
	t.Helper()
	response, payload, err := c.request(ctx, http.MethodGet,
		crossSystemRepoPath("/issues/comments/"+strconv.FormatInt(id, 10)), nil)
	if err != nil || response.StatusCode != wantStatus {
		t.Fatalf("read pm-cert issue comment %d: status=%s err=%v", id, crossSystemStatus(response), err)
	}
	if wantStatus == http.StatusNotFound {
		return crossSystemIssueComment{}
	}
	var comment crossSystemIssueComment
	if err := json.Unmarshal(payload, &comment); err != nil {
		t.Fatalf("decode pm-cert issue comment %d: %v", id, err)
	}
	return comment
}

func (c *crossSystemGitHubClient) listIssueCommentsSince(t *testing.T, ctx context.Context, since string) []crossSystemIssueComment {
	t.Helper()
	path := crossSystemRepoPath("/issues/comments?per_page=100&since=" + url.QueryEscape(since))
	response, payload, err := c.request(ctx, http.MethodGet, path, nil)
	if err != nil || response.StatusCode != http.StatusOK {
		t.Fatalf("list bounded pm-cert issue comments: status=%s err=%v", crossSystemStatus(response), err)
	}
	var comments []crossSystemIssueComment
	if err := json.Unmarshal(payload, &comments); err != nil {
		t.Fatalf("decode bounded pm-cert issue comments: %v", err)
	}
	if len(comments) == 100 {
		t.Fatal("bounded pm-cert issue-comment proof reached one full page")
	}
	return comments
}

func (c *crossSystemGitHubClient) cleanupLabel(t *testing.T, ctx context.Context, name string) {
	t.Helper()
	response, _, err := c.request(ctx, http.MethodDelete, crossSystemRepoPath("/labels/"+url.PathEscape(name)), nil)
	if err != nil || (response.StatusCode != http.StatusNoContent && response.StatusCode != http.StatusNotFound) {
		t.Errorf("delete pm-cert label %q: status=%s err=%v", name, crossSystemStatus(response), err)
		return
	}
	c.getLabel(t, ctx, name, http.StatusNotFound)
	t.Logf("CLEANUP label=%s independent_status=404", name)
}

func (c *crossSystemGitHubClient) cleanupIssueComment(t *testing.T, ctx context.Context, id int64) {
	t.Helper()
	path := crossSystemRepoPath("/issues/comments/" + strconv.FormatInt(id, 10))
	response, _, err := c.request(ctx, http.MethodDelete, path, nil)
	if err != nil || (response.StatusCode != http.StatusNoContent && response.StatusCode != http.StatusNotFound) {
		t.Errorf("delete pm-cert issue comment %d: status=%s err=%v", id, crossSystemStatus(response), err)
		return
	}
	c.getIssueComment(t, ctx, id, http.StatusNotFound)
	t.Logf("CLEANUP issue_comment_id=%d independent_status=404", id)
}

func crossSystemRepoPath(suffix string) string {
	return "/repos/" + crossSystemOwner + "/" + crossSystemRepo + suffix
}

func crossSystemStatus(response *http.Response) string {
	if response == nil {
		return "no-response"
	}
	return response.Status
}

func crossSystemNamedLabelCount(labels []crossSystemLabel, name string) int {
	count := 0
	for _, label := range labels {
		if label.Name == name {
			count++
		}
	}
	return count
}

func crossSystemNamedCommentCount(comments []crossSystemIssueComment, id int64) int {
	count := 0
	for _, comment := range comments {
		if comment.ID == id {
			count++
		}
	}
	return count
}
