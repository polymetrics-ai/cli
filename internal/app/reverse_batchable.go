package app

import (
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// NonBatchableActionError reports that a write action declared
// "batchable": false was targeted by a SourceTable-driven bulk reverse ETL
// plan. It is a typed error rather than a sentinel because callers need the
// fields: the CLI renders them, and an agent can branch on them without parsing
// prose. It mirrors commandrunner.BlockedCommandError.
//
// The action is not broken and not forbidden — it is only unavailable in bulk.
// Command names the individual `pm ...` invocation that still runs it, when the
// connector's command surface declares one.
type NonBatchableActionError struct {
	Connector   string
	Action      string
	SourceTable string
	Command     string
	Risk        string
}

func (e *NonBatchableActionError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "write action %q on connector %q is declared non-batchable and cannot run from a bulk reverse ETL plan", e.Action, e.Connector)
	if e.SourceTable != "" {
		fmt.Fprintf(&b, " over source table %q", e.SourceTable)
	}
	b.WriteString(": the connector declares this action must be invoked one record at a time")
	if e.Risk != "" {
		fmt.Fprintf(&b, " (%s)", e.Risk)
	}
	if e.Command != "" {
		fmt.Fprintf(&b, "; run it individually with `%s` instead", e.Command)
	} else {
		fmt.Fprintf(&b, "; run it individually as its own `pm %s <command>` instead", e.Connector)
	}
	return b.String()
}

// guardBatchableAction refuses a bulk reverse ETL plan whose action is declared
// non-batchable.
//
// It resolves the action from the LIVE connector manifest rather than from the
// stored plan, matching the reasoning at confirmationChallengeForPlan: a local
// state edit must not be able to strip a gate off an already-created plan. An
// unknown connector or an unknown action is not this guard's business — the
// existing resolve/validate paths report those — so it stays permissive there
// and refuses only on an explicit declaration.
func (a *App) guardBatchableAction(connectorName, actionName, sourceTable string) error {
	connector, ok := a.registry.Get(connectorName)
	if !ok {
		return nil
	}
	for _, action := range connectors.ManifestOf(connector).WriteActions {
		if action.Name != actionName || action.IsBatchable() {
			continue
		}
		return &NonBatchableActionError{
			Connector:   connectorName,
			Action:      actionName,
			SourceTable: sourceTable,
			Command:     individualCommandFor(connector, actionName),
			Risk:        strings.TrimSpace(action.Risk),
		}
	}
	return nil
}

// individualCommandFor finds the `pm <connector> <command>` that still executes
// the action, so the refusal can point at it. Returns "" when the connector
// publishes no command surface or no implemented command maps to the action.
func individualCommandFor(connector connectors.Connector, actionName string) string {
	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok {
		return ""
	}
	surface := provider.CommandSurface()
	if surface == nil {
		return ""
	}
	for _, command := range surface.Commands {
		if command.Write != actionName || command.Availability != "implemented" {
			continue
		}
		return strings.TrimSpace("pm " + connector.Name() + " " + command.Path)
	}
	return ""
}
