#!/usr/bin/env node

// Produces a source-locked, portfolio-scoped record of genuine shared-runtime
// gaps. It does not turn a gap into a disabled/N/A operation: every row retains
// the exact provider source operation and says that it cannot contribute to a
// merge-ready verdict until its generic owner closes it.
import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";

const root = process.cwd();
const phase = path.join(root, ".planning", "phases", "issue-4289-parity-map-batches-2-3-r1");
const output = path.join(phase, "traces", "missing-foundation-gaps.json");
const generatedAt = "2026-08-20T00:00:00Z";
const batches = [
  { id: "batch-2", connectors: ["grafana", "trello", "slack", "n8n", "google-calendar", "gmail", "twilio", "amazon-sqs", "elasticsearch"] },
  { id: "batch-3", connectors: ["gong", "google-ads", "facebook-marketing", "linkedin-ads", "aircall", "xero", "paypal-transaction", "gocardless", "amazon-seller-partner", "miro"] }
];

const owners = {
  "closed-operation-runtime-f2-binary-download": {
    lane: "cli-closed-operation-runtime-r1",
    issue: "https://github.com/polymetrics-ai/cli/issues/4307",
    issue_number: 4307,
    status: "open",
    capability: "closed declaration-owned binary download with bounded transfer, exact operation binding, and complete non-secret result preservation"
  },
  "closed-operation-runtime-f4-binary-upload-approval-digest": {
    lane: "cli-closed-operation-runtime-r1",
    issue: "https://github.com/polymetrics-ai/cli/issues/4307",
    issue_number: 4307,
    status: "open",
    capability: "closed declaration-owned binary or multipart upload that binds every approved file payload digest before plan, preview, explicit approval, and execute"
  }
};

function definitionFile(connector, name) {
  return path.join(root, "internal", "connectors", "defs", connector, name);
}

async function readJSON(file) {
  return JSON.parse(await readFile(file, "utf8"));
}

function groupBy(values, key) {
  const groups = new Map();
  for (const value of values) {
    const group = key(value);
    const entries = groups.get(group) || [];
    entries.push(value);
    groups.set(group, entries);
  }
  return groups;
}

function sourceReference(connector, lock, source) {
  return {
    source_lock: `internal/connectors/defs/${connector}/sources/${connector}-operation-source-lock.json`,
    url: source.source_url,
    source_location: source.source_location,
    captured_at: lock.captured_at,
    revision: `sha256:${lock.rest.sha256}`,
    sha256: lock.rest.sha256,
    bytes: lock.rest.bytes,
    provider_revision: null,
    provider_revision_note: "The published source has no immutable provider revision; this source-lock SHA-256 is the exact revision identity."
  };
}

function providerOperation(disposition) {
  const source = disposition.source;
  if (!disposition.method || !disposition.path) throw new Error(`${source?.source_id || "unknown"}: disposition lacks exact method/path`);
  return {
    source_id: source.source_id,
    method: disposition.method,
    path: disposition.path,
    operation_id: source.operation_id,
    deprecated: source.deprecated
  };
}

function closureVerification(gapID, connector, operation) {
  const generic = [
    "go test -count=1 -timeout 20m ./internal/connectors/engine -run 'Test.*(Binary|Multipart)'",
    "go test -count=1 -timeout 20m ./internal/connectors/commandrunner -run 'TestEveryImplementedCommandPassesRuntimePreflight'",
    "go run ./cmd/connectorgen validate --json",
    "go run ./cmd/connectorgen surface-sync --check",
    "go build ./cmd/pm"
  ];
  return {
    required_owner_head: {
      issue: owners[gapID].issue,
      lane: owners[gapID].lane,
      requirement: owners[gapID].capability
    },
    commands: generic,
    exact_operation_assertion: `${connector} ${operation.method} ${operation.path} is source-bound to one implemented closed operation and the built CLI reaches credential preflight rather than unknown-command, disabled, or an unbound-surface error.`,
    live_assertion: "With the approved disposable credential reference, use the persisted App path and record bounded non-secret request/result fingerprints plus provider readback where the provider operation mutates state."
  };
}

function row({ batch, connector, lock, disposition, gapID, affectedSurfaces, evidence, declaration }) {
  const source = disposition.source;
  const operation = providerOperation(disposition);
  return {
    gap_row_id: `${gapID}:${connector}:${operation.source_id}`,
    gap_id: gapID,
    connector,
    batch,
    provider_operation: operation,
    source: sourceReference(connector, lock, source),
    affected_surfaces: affectedSurfaces,
    failing_evidence: evidence,
    missing_provider_neutral_capability: owners[gapID].capability,
    owner: owners[gapID],
    status: "open",
    enabled: false,
    merge_ready_eligible: false,
    declaration,
    fanout_ref: gapID,
    closure_verification: closureVerification(gapID, connector, operation)
  };
}

function reasonGap(operation) {
  const reason = operation?.reason || "";
  if (/shared binary download runner/i.test(reason)) return "closed-operation-runtime-f2-binary-download";
  if (/shared file upload runner/i.test(reason)) return "closed-operation-runtime-f4-binary-upload-approval-digest";
  return null;
}

function evidenceFromOperation(operation) {
  return {
    kind: "declaration_runtime_gate",
    location: "api_surface.json operation metadata",
    validator: "commandrunner preflight cannot promote a blocked operation to an executable command",
    status: operation.status,
    model: operation.model,
    detail: operation.reason
  };
}

async function rowsForConnector(batch, connector) {
  const lock = await readJSON(definitionFile(connector, `sources/${connector}-operation-source-lock.json`));
  const map = await readJSON(definitionFile(connector, `sources/${connector}-declaration-disposition.json`));
  const rows = [];
  for (const disposition of map.ledger_dispositions) {
    const gapID = reasonGap(disposition.api_surface?.operation);
    if (!gapID) continue;
    const binaryUpload = gapID === "closed-operation-runtime-f4-binary-upload-approval-digest";
    rows.push(row({
      batch,
      connector,
      lock,
      disposition,
      gapID,
      affectedSurfaces: binaryUpload ? ["binary_upload", "direct_write", "reverse_etl"] : ["binary_download", "direct_read"],
      evidence: evidenceFromOperation(disposition.api_surface.operation),
      declaration: {
        api_surface_method: disposition.api_surface.method,
        api_surface_path: disposition.api_surface.path,
        parity_class: disposition.parity_class,
        state: disposition.state
      }
    }));
  }

  return { rows, operations: map.summary.operations_found };
}

function rollup(batch, sourceOperations, rows) {
  const byGap = [...groupBy(rows, (entry) => entry.gap_id).entries()].map(([gapID, entries]) => ({ gap_id: gapID, rows: entries.length })).sort((a, b) => a.gap_id.localeCompare(b.gap_id));
  const bySurface = [...groupBy(rows.flatMap((entry) => entry.affected_surfaces), (surface) => surface).entries()].map(([surface, entries]) => ({ surface, rows: entries.length })).sort((a, b) => a.surface.localeCompare(b.surface));
  return {
    ...(batch ? { batch } : {}),
    source_operations: sourceOperations,
    open_gap_rows: rows.length,
    operations_not_enabled_for_merge_ready: rows.length,
    distinct_open_gap_ids: byGap.map((entry) => entry.gap_id),
    by_gap: byGap,
    by_affected_surface: bySurface,
    open_foundation_gaps_block_merge_ready: rows.length > 0
  };
}

async function build() {
  const rows = [];
  const batchRollups = [];
  let totalOperations = 0;
  for (const batch of batches) {
    const batchRows = [];
    let batchOperations = 0;
    for (const connector of batch.connectors) {
      const result = await rowsForConnector(batch.id, connector);
      batchRows.push(...result.rows);
      batchOperations += result.operations;
    }
    rows.push(...batchRows);
    totalOperations += batchOperations;
    batchRollups.push({ connectors: batch.connectors, ...rollup(batch.id, batchOperations, batchRows) });
  }
  rows.sort((left, right) => left.gap_row_id.localeCompare(right.gap_row_id));
  const fanout = [...groupBy(rows, (entry) => entry.gap_id).entries()].map(([gapID, entries]) => ({
    gap_id: gapID,
    owner: owners[gapID],
    affected_connectors: [...new Set(entries.map((entry) => entry.connector))].sort(),
    affected_operations: entries.map((entry) => ({
      gap_row_id: entry.gap_row_id,
      connector: entry.connector,
      source_id: entry.provider_operation.source_id,
      method: entry.provider_operation.method,
      path: entry.provider_operation.path
    }))
  })).sort((a, b) => a.gap_id.localeCompare(b.gap_id));
  return {
    schema_version: 1,
    generated_at: generatedAt,
    scope: {
      batches: batches.map((batch) => batch.id),
      connector_count: batches.flatMap((batch) => batch.connectors).length,
      source_of_truth: "captured provider operation locks and their exact declaration-disposition rows",
      rule: "An open foundation gap is not enabled and cannot contribute to a connector, batch, portfolio, or merge-ready verdict. Safety, scope, tier, and destructive metadata add runtime controls; they do not hide a gap as disabled or not_applicable."
    },
    gap_rows: rows,
    fanout,
    batch_rollups: batchRollups,
    portfolio_rollup: rollup(null, totalOperations, rows)
  };
}

const expected = `${JSON.stringify(await build(), null, 2)}\n`;
if (process.argv.includes("--check")) {
  const actual = await readFile(output, "utf8");
  if (actual !== expected) throw new Error(`missing foundation gap ledger drift: run ${path.relative(root, import.meta.url.replace("file://", ""))}`);
  console.log("missing foundation gap ledger verified");
} else {
  await writeFile(output, expected);
  console.log(`wrote ${path.relative(root, output)}`);
}
