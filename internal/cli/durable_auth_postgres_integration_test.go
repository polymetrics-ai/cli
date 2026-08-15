//go:build databaseintegration

package cli_test

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors/native/dbtest"
	"polymetrics.ai/internal/coordination"
)

const durableAuthPostgresHelperEnv = "POLYMETRICS_DURABLE_AUTH_POSTGRES_HELPER"

func TestCLIDurableAuthenticationFenceAndRepairLivePostgres(t *testing.T) {
	if mode := os.Getenv(durableAuthPostgresHelperEnv); mode != "" {
		runDurableAuthPostgresCLIHelper(t, mode)
		return
	}
	if os.Getenv("POLYMETRICS_DATABASE_INTEGRATION") != "1" {
		t.Skip("database integration skipped: set POLYMETRICS_DATABASE_INTEGRATION=1")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	password := durableAuthRandomSecret(t)
	wrongPassword := durableAuthRandomSecret(t)
	harness, err := dbtest.New(dbtest.Config{
		Engine: "postgres-auth-fencing", ContainerRuntime: dbtest.Runtime(strings.TrimSpace(os.Getenv("POLYMETRICS_CONTAINER_RUNTIME"))),
		Image: "docker.io/library/postgres:16.10", ContainerPort: 5432, DataVolumePath: "/var/lib/postgresql/data",
		ContainerEndpoint: strings.TrimSpace(os.Getenv("POLYMETRICS_CONTAINER_ENDPOINT")), ExpectedImageBytes: 420 << 20,
		DockerCapacityProbeImage: "docker.io/library/busybox:1.37.0",
		ContainerArgs:            []string{"--env", "POSTGRES_DB=pm_auth", "--env", "POSTGRES_USER=pm_auth", "--env", "POSTGRES_PASSWORD=" + password},
	})
	if err != nil {
		t.Fatal("could not construct the explicit PostgreSQL container harness")
	}
	defer func() {
		cleanup, cleanupCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cleanupCancel()
		if err := harness.Close(cleanup); err != nil {
			t.Error("PostgreSQL authentication harness cleanup failed")
		}
	}()
	endpoint, err := harness.Start(ctx)
	if err != nil {
		t.Fatal("PostgreSQL authentication container did not start")
	}
	observer := durableAuthOpenPostgres(t, ctx, endpoint, password)
	defer func() { _ = observer.Close(context.WithoutCancel(ctx)) }()

	root := t.TempDir()
	setup := durableAuthPostgresCommand(root, endpoint, "setup", password, wrongPassword)
	stdout, err := setup.StdoutPipe()
	if err != nil {
		t.Fatal("could not capture setup readiness")
	}
	if err := setup.Start(); err != nil {
		t.Fatal("could not start authentication setup process")
	}
	ready, readErr := bufio.NewReader(stdout).ReadString('\n')
	if readErr != nil || ready != "fenced\n" {
		_ = setup.Process.Kill()
		t.Fatal("authentication setup process did not persist its fence")
	}
	if err := setup.Process.Kill(); err != nil {
		t.Fatal("could not kill authentication setup process")
	}
	if err := setup.Wait(); err == nil {
		t.Fatal("killed authentication setup process unexpectedly succeeded")
	}
	beforeHealth := durableAuthHealth(t, root)
	if !beforeHealth.Fenced || beforeHealth.LastFencedEpoch != beforeHealth.Epoch {
		t.Fatalf("durable fence health = %+v, want fenced current epoch", beforeHealth)
	}
	beforeSessions := durableAuthSessions(t, ctx, observer)

	commands := []*exec.Cmd{
		durableAuthPostgresCommand(root, endpoint, "admission", password, wrongPassword),
		durableAuthPostgresCommand(root, endpoint, "admission", password, wrongPassword),
	}
	for _, command := range commands {
		if err := command.Start(); err != nil {
			t.Fatal("could not start concurrent fenced admission process")
		}
	}
	for _, command := range commands {
		if err := command.Wait(); err != nil {
			t.Fatal("fenced admission process did not observe the typed refusal")
		}
	}
	if got := durableAuthSessions(t, ctx, observer); got != beforeSessions {
		t.Fatalf("PostgreSQL sessions after fenced concurrent checks = %d, want %d (zero sends)", got, beforeSessions)
	}
	if after := durableAuthHealth(t, root); after != beforeHealth {
		t.Fatalf("fenced refusal mutated health: before=%+v after=%+v", beforeHealth, after)
	}

	repair := durableAuthPostgresCommand(root, endpoint, "repair", password, wrongPassword)
	repairOutput, repairErr := repair.CombinedOutput()
	if bytes.Contains(repairOutput, []byte(password)) || bytes.Contains(repairOutput, []byte(wrongPassword)) {
		t.Fatal("verified credential repair process exposed a PostgreSQL credential")
	}
	if repairErr != nil {
		t.Fatalf("verified credential repair process failed: %s", repairOutput)
	}
	afterRepair := durableAuthHealth(t, root)
	if afterRepair.Fenced || afterRepair.Epoch <= beforeHealth.Epoch || afterRepair.LastFencedEpoch != beforeHealth.Epoch {
		t.Fatalf("durable repair health = %+v, before=%+v", afterRepair, beforeHealth)
	}
	if got := durableAuthSessions(t, ctx, observer); got <= beforeSessions {
		t.Fatalf("PostgreSQL sessions after current-generation repair and check = %d, want greater than %d", got, beforeSessions)
	}

	// Hold an epoch-two runtime while a second verified repair publishes epoch
	// three. The stale runtime is a real app.Open-resolved PostgreSQL connector,
	// not a hand-built coordinator call.
	staleApp, err := app.Open(root)
	if err != nil {
		t.Fatal("could not open the stale-generation production app")
	}
	staleConnector, staleRuntime, err := staleApp.ResolveConnectorCredential(ctx, "postgres", "postgres-correct", nil)
	if err != nil {
		t.Fatal("could not resolve the stale-generation production runtime")
	}
	secondRepair := durableAuthPostgresCommand(root, endpoint, "repair", password, wrongPassword)
	secondOutput, secondErr := secondRepair.CombinedOutput()
	if bytes.Contains(secondOutput, []byte(password)) || bytes.Contains(secondOutput, []byte(wrongPassword)) {
		t.Fatal("second repair process exposed a PostgreSQL credential")
	}
	if secondErr != nil {
		t.Fatal("second verified repair process failed")
	}
	latestHealth := durableAuthHealth(t, root)
	if latestHealth.Epoch <= afterRepair.Epoch || latestHealth.Fenced {
		t.Fatalf("second repair did not publish current ownership: before=%+v after=%+v", afterRepair, latestHealth)
	}
	beforeStaleSessions := durableAuthSessions(t, ctx, observer)
	checkpointBeforeStale := durableAuthStreamStates(t, root)
	if err := staleConnector.Check(ctx, staleRuntime); !errors.Is(err, coordination.ErrAuthCohortEpochMismatch) {
		t.Fatalf("stale production runtime = %v, want ErrAuthCohortEpochMismatch", err)
	}
	if got := durableAuthSessions(t, ctx, observer); got != beforeStaleSessions {
		t.Fatalf("PostgreSQL sessions after stale runtime = %d, want %d (zero sends)", got, beforeStaleSessions)
	}
	if after := durableAuthStreamStates(t, root); !bytes.Equal(after, checkpointBeforeStale) {
		t.Fatal("stale production runtime advanced a checkpoint")
	}
}

func runDurableAuthPostgresCLIHelper(t *testing.T, mode string) {
	root := os.Getenv("POLYMETRICS_DURABLE_AUTH_ROOT")
	config := []string{
		"--config", "host=" + os.Getenv("POLYMETRICS_DURABLE_AUTH_HOST"),
		"--config", "port=" + os.Getenv("POLYMETRICS_DURABLE_AUTH_PORT"),
		"--config", "database=pm_auth", "--config", "username=pm_auth", "--config", "schema=public", "--config", "sslmode=disable",
	}
	var lastOutput string
	run := func(args ...string) int {
		var stdout, stderr bytes.Buffer
		code := cli.Run(args, &stdout, &stderr)
		combined := stdout.String() + stderr.String()
		lastOutput = combined
		if strings.Contains(combined, os.Getenv("PM_DURABLE_AUTH_PASSWORD")) || strings.Contains(combined, os.Getenv("PM_DURABLE_AUTH_WRONG_PASSWORD")) {
			t.Fatal("CLI output exposed a PostgreSQL credential")
		}
		return code
	}
	switch mode {
	case "setup":
		if code := run("init", "--root", root, "--json"); code != 0 {
			t.Fatal("pm init failed")
		}
		correct := append([]string{"credentials", "add", "postgres-correct", "--connector", "postgres"}, config...)
		correct = append(correct, "--from-env", "password=PM_DURABLE_AUTH_PASSWORD", "--root", root, "--json")
		if code := run(correct...); code != 0 {
			t.Fatal("adding the live PostgreSQL credential failed")
		}
		wrong := append([]string{"credentials", "add", "postgres-wrong", "--connector", "postgres", "--link-credential", "postgres-correct"}, config...)
		wrong = append(wrong, "--from-env", "password=PM_DURABLE_AUTH_WRONG_PASSWORD", "--root", root, "--json")
		if code := run(wrong...); code != 0 {
			t.Fatal("adding the linked PostgreSQL credential failed")
		}
		if code := run("credentials", "test", "postgres-wrong", "--root", root, "--json"); code == 0 {
			t.Fatal("wrong PostgreSQL authentication unexpectedly succeeded")
		}
		fmt.Fprintln(os.Stdout, "fenced")
		select {}
	case "admission":
		a, err := app.Open(root)
		if err != nil {
			t.Fatal("could not open the production app for fenced admission")
		}
		if _, _, err := a.ResolveConnectorCredential(context.Background(), "postgres", "postgres-correct", nil); !errors.Is(err, coordination.ErrAuthCohortFenced) {
			t.Fatalf("unrepaired production admission = %v, want ErrAuthCohortFenced", err)
		}
		if code := run("etl", "check", "--connector", "postgres", "--credential", "postgres-correct", "--root", root, "--json"); code == 0 {
			t.Fatal("fenced credential reached PostgreSQL")
		}
	case "repair":
		if code := run("credentials", "test", "postgres-correct", "--root", root, "--json"); code != 0 {
			t.Fatalf("verified healthy PostgreSQL repair failed: %s", lastOutput)
		}
		if code := run("etl", "check", "--connector", "postgres", "--credential", "postgres-correct", "--root", root, "--json"); code != 0 {
			t.Fatalf("repaired PostgreSQL cohort remained fenced: %s", lastOutput)
		}
	default:
		t.Fatalf("unknown durable auth helper mode %q", mode)
	}
}

func durableAuthPostgresCommand(root string, endpoint dbtest.Endpoint, mode, password, wrongPassword string) *exec.Cmd {
	command := exec.Command(os.Args[0], "-test.run=^TestCLIDurableAuthenticationFenceAndRepairLivePostgres$")
	command.Env = append(os.Environ(),
		durableAuthPostgresHelperEnv+"="+mode,
		"POLYMETRICS_DURABLE_AUTH_ROOT="+root,
		"POLYMETRICS_DURABLE_AUTH_HOST="+endpoint.Host,
		"POLYMETRICS_DURABLE_AUTH_PORT="+strconv.Itoa(endpoint.Port),
		"PM_DURABLE_AUTH_PASSWORD="+password,
		"PM_DURABLE_AUTH_WRONG_PASSWORD="+wrongPassword,
	)
	return command
}

func durableAuthRandomSecret(t *testing.T) string {
	t.Helper()
	var data [32]byte
	if _, err := rand.Read(data[:]); err != nil {
		t.Fatal("could not generate an ephemeral database credential")
	}
	return base64.RawURLEncoding.EncodeToString(data[:])
}

func durableAuthOpenPostgres(t *testing.T, ctx context.Context, endpoint dbtest.Endpoint, password string) *pgx.Conn {
	t.Helper()
	config, err := pgx.ParseConfig("host=" + endpoint.Host + " port=" + strconv.Itoa(endpoint.Port) + " dbname=pm_auth user=pm_auth sslmode=disable")
	if err != nil {
		t.Fatal("could not construct the PostgreSQL observer configuration")
	}
	config.Password = password
	for {
		connection, connectErr := pgx.ConnectConfig(ctx, config)
		if connectErr == nil {
			return connection
		}
		select {
		case <-ctx.Done():
			t.Fatal("PostgreSQL did not become reachable")
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func durableAuthSessions(t *testing.T, ctx context.Context, connection *pgx.Conn) int64 {
	t.Helper()
	var sessions int64
	if err := connection.QueryRow(ctx, "SELECT sessions FROM pg_stat_database WHERE datname = current_database()").Scan(&sessions); err != nil {
		t.Fatal("could not observe PostgreSQL session count")
	}
	return sessions
}

func durableAuthHealth(t *testing.T, root string) coordination.AuthCohortHealth {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "auth-cohorts.json"))
	if err != nil {
		t.Fatal("could not read durable authentication health")
	}
	var envelope struct {
		Records map[string]coordination.AuthCohortHealth `json:"records"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil || len(envelope.Records) != 1 {
		t.Fatal("durable authentication health had an unexpected shape")
	}
	for _, health := range envelope.Records {
		return health
	}
	return coordination.AuthCohortHealth{}
}

func durableAuthStreamStates(t *testing.T, root string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatal("could not read durable stream state")
	}
	var envelope struct {
		StreamStates json.RawMessage `json:"stream_states"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatal("could not decode durable stream state")
	}
	return append([]byte(nil), envelope.StreamStates...)
}
