package database_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
)

func TestIncrementalDedupeHistoryRefusesEachNonPostgresRouteBeforeSessionMutation(t *testing.T) {
	definition := loadTestDefinition(t, strings.Replace(
		validDefinitionJSON,
		`"admitted_modes": []`,
		`"admitted_modes": ["incremental_dedupe_history"]`,
		1,
	))
	postgres := definition.Driver()
	api := database.DriverDeclaration{ID: "api", Protocol: "https", APIVersion: 1}
	mysql := database.DriverDeclaration{ID: "mysql", Protocol: "mysql-wire", APIVersion: 1}

	tests := []struct {
		name  string
		route database.DatabaseWriteHistoryRoute
		want  database.DatabaseWriteHistoryRouteReason
	}{
		{
			name:  "api source to postgres destination",
			route: database.DatabaseWriteHistoryRoute{Source: api, Destination: postgres},
			want:  database.DatabaseWriteHistoryRouteSourceUnsupported,
		},
		{
			name:  "postgres source to mysql destination",
			route: database.DatabaseWriteHistoryRoute{Source: postgres, Destination: mysql},
			want:  database.DatabaseWriteHistoryRouteDestinationUnsupported,
		},
		{
			name:  "api source to mysql destination",
			route: database.DatabaseWriteHistoryRoute{Source: api, Destination: mysql},
			want:  database.DatabaseWriteHistoryRouteUnsupported,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			driver := &databaseWriteDriverFake{atomicOverwrite: true, targetRows: map[string]bool{"retained": true}}
			executor, ledger := testDatabaseWriteExecutorWithStore(t, driver)
			_, err := database.NewDatabaseWritePlan(context.Background(), database.DatabaseWritePlanRequest{
				Definition:     definition,
				Control:        testDatabaseWriteControl(t, "history_orders", "history-orders", 1),
				Mode:           synccontract.ModeIncrementalDedupeHistory,
				Strategy:       connectors.ApplyStrategyDedupeHistory,
				Mapping:        testDatabaseWriteMapping(t, "source_id", "id"),
				Keys:           []string{"id"},
				RecordCount:    1,
				BatchSize:      1,
				HistoryRoute:   testCase.route,
				Destructive:    false,
			})
			var routeErr *database.DatabaseWriteHistoryRouteError
			if !errors.As(err, &routeErr) || routeErr.Reason != testCase.want {
				t.Fatalf("NewDatabaseWritePlan(history route) error = %T %v, want typed reason %q", err, err, testCase.want)
			}
			if driver.beginCalls != 0 || driver.batchCalls != 0 || driver.commitCalls != 0 || driver.rollbackCalls != 0 || ledger.writeCount() != 0 || !driver.targetRows["retained"] {
				t.Fatalf("rejected route mutated state: begin/batch/commit/rollback/ledger/retained = %d/%d/%d/%d/%d/%t", driver.beginCalls, driver.batchCalls, driver.commitCalls, driver.rollbackCalls, ledger.writeCount(), driver.targetRows["retained"])
			}
			_ = executor
		})
	}
}
