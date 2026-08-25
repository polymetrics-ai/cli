#!/usr/bin/env node

// Generate a reviewable declaration-first map without attempting to infer an
// operation contract. It intentionally reads only the committed source map,
// disposition, existing typed-action schema, and existing CLI surface.
import { readFile, writeFile } from "node:fs/promises";
import { existsSync } from "node:fs";
import { resolve } from "node:path";

const root = process.cwd();
const phase = resolve(root, ".planning/phases/issue-4292-parity-batches-8-10-r1");
const defs = resolve(root, "internal/connectors/defs");
const cohortDefinitions = {
  "01": [
    { batch: 8, connector: "brex" },
    { batch: 8, connector: "zoho-books" },
    { batch: 8, connector: "testrail" },
    { batch: 8, connector: "amplitude" },
    { batch: 8, connector: "posthog" },
  ],
  "02": [
    { batch: 8, connector: "metabase" },
    { batch: 8, connector: "dbt" },
    { batch: 8, connector: "looker" },
    { batch: 8, connector: "mode" },
    { batch: 8, connector: "dremio" },
  ],
  "03": [
    { batch: 9, connector: "coda" },
    { batch: 9, connector: "clickup-api" },
    { batch: 9, connector: "calendly" },
    { batch: 9, connector: "greenhouse" },
    { batch: 9, connector: "lever-hiring" },
  ],
  "04": [
    { batch: 9, connector: "ashby" },
    { batch: 9, connector: "workable" },
    { batch: 9, connector: "recruitee" },
    { batch: 9, connector: "hibob" },
    { batch: 9, connector: "factorial" },
  ],
  "05": [
    { batch: 10, connector: "datadog" },
    { batch: 10, connector: "pagerduty" },
    { batch: 10, connector: "auth0" },
    { batch: 10, connector: "okta" },
    { batch: 10, connector: "firehydrant" },
  ],
  "06": [
    { batch: 10, connector: "adobe-commerce-magento" },
    { batch: 10, connector: "commercetools" },
    { batch: 10, connector: "recharge" },
    { batch: 10, connector: "docuseal" },
    { batch: 10, connector: "eventbrite" },
  ],
};
const cohortArgument = process.argv.indexOf("--cohort");
const cohortID = cohortArgument < 0 ? "01" : process.argv[cohortArgument + 1];
if (!cohortID || !cohortDefinitions[cohortID]) {
  throw new Error(`unknown or missing --cohort value ${JSON.stringify(cohortID)}`);
}
const cohort = cohortDefinitions[cohortID];
const output = resolve(phase, `DECLARATION-FIRST-COHORT-${cohortID}.json`);
const summaryOutput = resolve(phase, `DECLARATION-FIRST-COHORT-${cohortID}.md`);

const sourceProjectionGap = {
  id: "source-lock-projection-gap",
  scope: "shared source-certification only",
  evidence: [
    "cmd/connectorgen/sourceprojection.go:2244-2268 requires a canonical *-operation-descriptor.json and has no v3 source_documents projector.",
    "cmd/connectorgen/sourceimport.go:809-813 rejects imported OpenAPI metadata until aggregate inventory derivation exists.",
  ],
  minimal_change: "Project v3 source_documents and retained captured-document evidence into canonical operation descriptors and aggregate OpenAPI inventory; do not fabricate a connector-local descriptor.",
};

const array = (value) => Array.isArray(value) ? value : [];
const pretty = (value) => `${JSON.stringify(value, null, 2)}\n`;
const assert = (condition, message) => {
  if (!condition) throw new Error(message);
};
async function readJSON(path, fallback) {
  if (!existsSync(path)) {
    if (fallback !== undefined) return fallback;
    throw new Error(`required JSON artifact is missing: ${path}`);
  }
  return JSON.parse(await readFile(path, "utf8"));
}
const connectorFile = (connector, name) => resolve(defs, connector, name);
const sourceFile = (connector, name) => connectorFile(connector, `sources/${name}`);

function sourceIdentity(operation) {
  return {
    source_lock: operation.source_lock,
    source_id: operation.source_id,
    source_url: operation.source_url,
    source_location: operation.source_location,
    operation_id: operation.operation_id,
    method: operation.method,
    path: operation.path,
  };
}

function exactCLIPaths(commands, actionID) {
  return commands
    .filter((command) => command.write === actionID)
    .map((command) => ({
      path: command.path,
      intent: command.intent,
      availability: command.availability,
    }))
    .sort((left, right) => left.path.localeCompare(right.path));
}

function unavailableConnector(batch, connector, disposition) {
  const summary = disposition.summary ?? {};
  assert(summary.state === "skipped" || summary.state === "dynamic", `${connector}: expected explicit unavailable/dynamic source state`);
  assert(summary.operations_found === null, `${connector}: unavailable/dynamic source must retain null operation count`);
  return {
    batch,
    connector,
    source_status: summary.state,
    direct_write_operations: null,
    unavailable_reason: summary.reason,
    coverage_confidence: summary.coverage_confidence,
    intended_cli_paths: [],
    lane: {
      id: "public-source-unavailable",
      state: "not-enumerable",
      evidence: summary.coverage_confidence?.basis,
    },
    deferred_component: null,
  };
}

async function mapConnector({ batch, connector }) {
  const disposition = await readJSON(sourceFile(connector, `${connector}-declaration-disposition.json`));
  if (disposition.summary?.state === "skipped" || disposition.summary?.state === "dynamic") {
    return unavailableConnector(batch, connector, disposition);
  }

  const [crosswalk, writes, cli] = await Promise.all([
    readJSON(sourceFile(connector, `${connector}-operation-crosswalk.json`)),
    readJSON(connectorFile(connector, "writes.json"), { actions: [] }),
    readJSON(connectorFile(connector, "cli_surface.json")),
  ]);
  const dispositionBySourceID = new Map(array(disposition.ledger_dispositions).map((row) => [row.source?.source_id, row]));
  const writeByID = new Map(array(writes.actions).map((action) => [action.name, action]));
  const rows = [];

  for (const operation of array(crosswalk.source_operations)) {
    const mapped = dispositionBySourceID.get(operation.source_id);
    assert(mapped, `${connector}: missing disposition for ${operation.source_id}`);
    if (mapped.parity_class !== "direct_write") continue;

    const inventory = operation.crosswalk?.inventory ?? {};
    const actionIDs = inventory.state === "materialized" && inventory.kind === "direct_write"
      ? array(inventory.id)
      : [];
    const actionPaths = actionIDs.flatMap((actionID) => {
      assert(writeByID.has(actionID), `${connector}: crosswalk action ${actionID} is absent from writes.json`);
      const paths = exactCLIPaths(array(cli.commands), actionID);
      assert(paths.length > 0, `${connector}: materialized action ${actionID} has no declared CLI path`);
      return paths.map((path) => ({ action: actionID, ...path }));
    }).sort((left, right) => left.path.localeCompare(right.path));

    const actionBacked = actionIDs.length > 0;
    assert(actionBacked === (mapped.state === "enabled"), `${connector}: disposition state disagrees with crosswalk materialization for ${operation.source_id}`);
    rows.push({
      source: sourceIdentity(operation),
      existing_typed_action_ids: actionIDs,
      intended_cli_paths: actionPaths,
      lane: actionBacked
        ? {
            id: "connector-local-existing-schema",
            state: "already-materialized",
            evidence: `writes.json action(s) ${JSON.stringify(actionIDs)} and their existing CLI binding(s) are exact crosswalk targets.`,
          }
        : {
            id: "connector-local-typed-operation-contract",
            state: "declaration-pending",
            evidence: mapped.rejection?.evidence ?? mapped.declaration?.status,
          },
      deferred_component: sourceProjectionGap,
      local_declaration: actionBacked
        ? { state: "enabled", evidence: mapped.declaration?.status }
        : {
            state: "declaration-pending",
            reason: mapped.rejection?.reason,
            detail: mapped.rejection?.detail,
            minimal_change: mapped.foundation?.declaration_pending?.minimal_change,
          },
    });
  }
  rows.sort((left, right) => left.source.source_id.localeCompare(right.source.source_id));
  return {
    batch,
    connector,
    source_status: "enumerable",
    direct_write_operations: rows.length,
    rows,
  };
}

function summary(document) {
  const lines = [
    `# Issue #4292 declaration-first direct-write cohort ${cohortID}`,
    "",
    "This mechanical cohort preserves source identities and existing action CLI paths. It does not create or infer a provider request, response, pagination, body schema, or CLI spelling. `source-lock-projection-gap` blocks source certification uniformly; it does not turn connector-local declaration work into an engine gap.",
    "",
    "| Batch | Connector | Direct-write source operations | Existing-schema CLI-bound | Declaration-pending | Source status |",
    "| --- | --- | ---: | ---: | ---: | --- |",
  ];
  for (const connector of document.connectors) {
    if (connector.direct_write_operations === null) {
      lines.push(`| ${connector.batch} | ${connector.connector} | — | — | — | ${connector.source_status}: ${connector.unavailable_reason} |`);
      continue;
    }
    const actionBacked = connector.rows.filter((row) => row.lane.state === "already-materialized").length;
    lines.push(`| ${connector.batch} | ${connector.connector} | ${connector.direct_write_operations} | ${actionBacked} | ${connector.direct_write_operations - actionBacked} | ${connector.source_status} |`);
  }
  lines.push("", "All enumerable rows name the same deferred source-certification component: `source-lock-projection-gap`. Existing command paths are reported verbatim from `cli_surface.json`; missing rows deliberately contain no intended path until a bounded connector-owned typed contract exists.", "");
  return lines.join("\n");
}

const connectors = await Promise.all(cohort.map(mapConnector));
const document = {
  schema_version: 1,
  issue: 4292,
  cohort: `batch-${cohort[0].batch}-declaration-first-${cohortID}`,
  purpose: "mechanical direct-write promotion inventory from committed connector evidence",
  source_certification: {
    state: "blocked",
    deferred_component: sourceProjectionGap,
  },
  connectors,
};
const rendered = pretty(document);
const renderedSummary = summary(document);

if (process.argv.includes("--check")) {
  assert(existsSync(output), "cohort JSON is missing; run the generator without --check");
  assert(existsSync(summaryOutput), "cohort summary is missing; run the generator without --check");
  assert(await readFile(output, "utf8") === rendered, "cohort JSON drift; regenerate it");
  assert(await readFile(summaryOutput, "utf8") === renderedSummary, "cohort summary drift; regenerate it");
  const total = connectors.reduce((sum, connector) => sum + (connector.direct_write_operations ?? 0), 0);
  process.stdout.write(`verified declaration-first cohort for ${connectors.length} connector(s), ${total} direct-write source operation(s)\n`);
} else {
  await writeFile(output, rendered);
  await writeFile(summaryOutput, renderedSummary);
  const total = connectors.reduce((sum, connector) => sum + (connector.direct_write_operations ?? 0), 0);
  process.stdout.write(`generated declaration-first cohort for ${connectors.length} connector(s), ${total} direct-write source operation(s)\n`);
}
