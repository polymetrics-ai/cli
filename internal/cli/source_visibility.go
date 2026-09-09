package cli

import (
	"context"
	"fmt"
	"io"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
)

func hasSourceSelectionFlags(flags parsedFlags) bool {
	for _, name := range []string{"sources", "inventory", "source-id", "lane", "preflight"} {
		if len(flags.values[name]) > 0 {
			return true
		}
	}
	return false
}
func runConnectorSourceInspection(ctx context.Context, connector string, flags parsedFlags, stdout io.Writer, jsonOut bool, registry *connectors.Registry) error {
	stdout = sourceOutputWriter{Writer: stdout}
	allowed := map[string]bool{"sources": true, "inventory": true, "source-id": true, "lane": true, "preflight": true}
	for name, values := range flags.values {
		if !allowed[name] || len(values) != 1 {
			return usageErrorf("source inspection requires one occurrence of each permitted selector flag")
		}
	}
	discovery := flags.first("sources") != ""
	preflight := flags.first("preflight") != ""
	for _, name := range []string{"sources", "preflight"} {
		if len(flags.values[name]) > 0 && (!flags.isBare(name) || flags.first(name) != "true") {
			return usageErrorf("--%s is a presence flag", name)
		}
	}
	if discovery {
		if len(flags.values) != 1 {
			return usageErrorf("--sources cannot be combined with source selection")
		}
		v, err := registry.SourceVisibility(ctx, connector)
		if err != nil {
			return err
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ConnectorSourceCatalog", "api_version": apiVersion, "connector": connector, "catalog": v, "execution_checked": false})
		}
		if _, err := fmt.Fprintf(stdout, "SOURCE CATALOG %s coverage=%s operations=%d cells=%d (inspection only)\n", connector, v.Coverage, v.OperationCount, v.CellCount); err != nil {
			return err
		}
		for _, o := range v.Operations {
			for _, c := range o.Cells {
				view, e := connectors.InspectSourceCell(v, connectors.SourceCellSelection{Source: o.Source, Lane: c.Lane})
				if e != nil {
					return e
				}
				if err := writeSourceCellText(stdout, view); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, name := range []string{"inventory", "source-id", "lane"} {
		if flags.first(name) == "" || flags.isBare(name) {
			return usageErrorf("exact source selection requires --inventory, --source-id and --lane values")
		}
	}
	s := connectors.SourceCellSelection{Source: connectors.SourceOperationKey{Connector: connector, Inventory: flags.first("inventory"), ID: flags.first("source-id")}, Lane: connectors.SourceLane(flags.first("lane"))}
	if preflight {
		result, err := app.PreflightConnectorSource(ctx, registry, s)
		if err != nil {
			return err
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "ConnectorSourcePreflight", "api_version": apiVersion, "preflight": result, "execution_checked": false})
		}
		return writeSourceCellText(stdout, result.Cell)
	}
	v, err := registry.SourceVisibility(ctx, connector)
	if err != nil {
		return err
	}
	view, err := connectors.InspectSourceCell(v, s)
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ConnectorSourceCell", "api_version": apiVersion, "cell": view, "execution_checked": false})
	}
	return writeSourceCellText(stdout, view)
}

// A source report must not return success after a short write, including a
// writer that violates io.Writer's requirement to pair short counts with errors.
type sourceOutputWriter struct{ io.Writer }

func (w sourceOutputWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	if err == nil && n != len(p) {
		err = io.ErrShortWrite
	}
	return n, err
}

func writeSourceCellText(w io.Writer, view connectors.SourceCellView) error {
	s, c := view.Selection, view.Cell
	if _, err := fmt.Fprintf(w, "%s/%s/%s lane=%s %s %s state=%s capability=%s reason=%s\n", s.Source.Connector, s.Source.Inventory, s.Source.ID, s.Lane, view.Operation.Method, view.Operation.Path, c.SourceState, c.Capability.Name, c.SourceReason.Code); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  evidence=%s identity-citation=%d lane-citations=%v\n", view.LaneEvidenceScope, view.Operation.IdentityCitation, c.LaneFacts); err != nil {
		return err
	}
	for _, i := range append([]int{view.Operation.IdentityCitation}, c.LaneFacts...) {
		ref := view.Citations[i]
		if _, err := fmt.Fprintf(w, "  %s#%s sha256=%s\n", ref.DocumentID, ref.Pointer, ref.ValueSHA256); err != nil {
			return err
		}
	}
	return nil
}
