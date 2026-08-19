package postgres

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"

	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
)

// postgresManagedTargetHistoryColumns is fixed target-owned metadata. It is
// intentionally outside MappingContractV1: mappings describe source business
// fields, while validity windows are the managed history-target contract.
func postgresManagedTargetHistoryColumns() []postgresManagedTargetColumn {
	return []postgresManagedTargetColumn{
		{name: synccontract.HistoryValidFromColumn, typeSQL: "TIMESTAMP WITH TIME ZONE", nullable: false},
		{name: synccontract.HistoryValidToColumn, typeSQL: "TIMESTAMP WITH TIME ZONE", nullable: true},
		{name: synccontract.HistoryIsCurrentColumn, typeSQL: "BOOLEAN", nullable: false},
	}
}

// postgresEnsureManagedTargetHistoryLayout adds the complete history layout
// only to an empty, already-owned target and only in the first history write
// transaction. A partially altered or populated non-history target is refused
// rather than backfilled with invented validity windows.
func postgresEnsureManagedTargetHistoryLayout(ctx context.Context, tx pgx.Tx, plan database.DatabaseWritePlan) error {
	target := plan.Control().Target()
	if err := postgresAssertMappedRelation(ctx, tx, target, plan.Mapping(), true); err != nil {
		return errPostgresWriteTargetUnverified
	}
	qualified := postgresManagedTargetQualifiedRelation(plan.Control())
	var existing int64
	if err := tx.QueryRow(ctx, `
		SELECT count(*)
		FROM pg_catalog.pg_attribute AS a
		JOIN pg_catalog.pg_class AS c ON c.oid = a.attrelid
		JOIN pg_catalog.pg_namespace AS n ON n.oid = c.relnamespace
		WHERE n.nspname = $1 AND c.relname = $2
			AND a.attnum > 0 AND NOT a.attisdropped
			AND a.attname IN ($3, $4, $5)`,
		target.Namespace(), target.Relation(),
		synccontract.HistoryValidFromColumn,
		synccontract.HistoryValidToColumn,
		synccontract.HistoryIsCurrentColumn,
	).Scan(&existing); err != nil {
		return errPostgresWriteTargetUnverified
	}
	if existing == int64(len(postgresManagedTargetHistoryColumns())) {
		return nil
	}
	if existing != 0 {
		return errPostgresWriteTargetUnverified
	}
	var populated bool
	if err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM "+qualified+" LIMIT 1)").Scan(&populated); err != nil || populated {
		return errPostgresWriteTargetUnverified
	}
	definitions := make([]string, 0, len(postgresManagedTargetHistoryColumns()))
	for _, column := range postgresManagedTargetHistoryColumns() {
		definition := "ADD COLUMN " + quoteIdentifier(column.name) + " " + column.typeSQL
		if !column.nullable {
			definition += " NOT NULL"
		}
		definitions = append(definitions, definition)
	}
	if _, err := tx.Exec(ctx, "ALTER TABLE "+qualified+" "+strings.Join(definitions, ", ")); err != nil {
		return errPostgresWriteTargetUnverified
	}
	if err := postgresAssertMappedRelation(ctx, tx, target, plan.Mapping(), true); err != nil {
		return errPostgresWriteTargetUnverified
	}
	return nil
}
