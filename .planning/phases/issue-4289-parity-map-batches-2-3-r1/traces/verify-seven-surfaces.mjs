#!/usr/bin/env node

import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";

const root = process.cwd();
const phase = path.join(root, ".planning", "phases", "issue-4289-parity-map-batches-2-3-r1");
const connectors = ["grafana", "trello", "slack", "n8n", "google-calendar", "gmail", "twilio", "amazon-sqs", "elasticsearch", "gong", "google-ads", "facebook-marketing", "linkedin-ads", "aircall", "xero", "paypal-transaction", "gocardless", "amazon-seller-partner", "miro"];
const surfaces = ["binary_download", "binary_upload", "direct_read", "direct_write", "etl", "reverse_etl"];
const sourceClassSurface = {
  binary_read: "binary_download",
  binary_write: "binary_upload",
  direct_read: "direct_read",
  direct_write: "direct_write",
  etl: "etl",
};

function batchForConnector(connector) {
  return connectors.indexOf(connector) < 9 ? "batch-2" : "batch-3";
}

async function readJSON(file, fallback) {
  try {
    return JSON.parse(await readFile(file, "utf8"));
  } catch (error) {
    if (error.code === "ENOENT") return fallback;
    throw error;
  }
}

async function fileExists(file) {
  try {
    await readFile(file, "utf8");
    return true;
  } catch (error) {
    if (error.code === "ENOENT") return false;
    throw error;
  }
}

function endpointKey(method, pathname) {
  return `${String(method).toUpperCase()} ${pathname}`;
}

function endpointMatches(endpoint, expected) {
  return endpoint && endpointKey(endpoint.method, endpoint.path) === expected;
}

function commandEvidence(commands, expected, intents = null) {
  const candidates = commands.filter((command) =>
    (!intents || intents.includes(command.intent)) &&
    command.api_surface?.some((endpoint) => endpointMatches(endpoint, expected)),
  );
  const implemented = candidates.filter((command) => command.availability === "implemented");
  const partial = candidates.filter((command) => command.availability === "partial");
  if (implemented.length > 0) {
    return {
      status: "implemented",
      commands: implemented.map((command) => ({ path: command.path, intent: command.intent, availability: command.availability })),
    };
  }
  if (partial.length > 0) {
    return {
      status: "partial",
      commands: partial.map((command) => ({ path: command.path, intent: command.intent, availability: command.availability })),
    };
  }
  return { status: "declaration-pending", commands: [] };
}

function primaryRuntimeEvidence(operations, commands, transport, row) {
  const expected = endpointKey(row.method, row.api_surface.path);
  switch (row.parity_class) {
    case "direct_read": {
      const operation = operations.find((candidate) =>
        candidate.kind === "rest_read" && endpointMatches(candidate.rest, expected),
      );
      const cli = commandEvidence(commands, expected, ["direct_read"]);
      return {
        status: operation && cli.status === "implemented" ? "implemented" : "declaration-pending",
        operation: operation?.id ?? null,
        cli,
      };
    }
    case "direct_write": {
      const operation = operations.find((candidate) =>
        candidate.kind === "rest_write" && endpointMatches(candidate.rest, expected),
      );
      const cli = commandEvidence(commands, expected, ["direct_write"]);
      return {
        status: operation && cli.status === "implemented" ? "implemented" : "declaration-pending",
        operation: operation?.id ?? null,
        cli,
      };
    }
    case "etl": {
      const cli = commandEvidence(commands, expected, ["etl"]);
      const source = transport?.source_transport;
      const sourceMatches = source && endpointMatches(source.api_surface, expected);
      return {
        status: sourceMatches && cli.status === "implemented" ? "implemented" : "declaration-pending",
        source_transport: sourceMatches ? source.stream : null,
        cli,
      };
    }
    case "binary_read":
    case "binary_write": {
      const operation = operations.find((candidate) =>
        endpointMatches(candidate.binary || candidate.file, expected),
      );
      const intent = row.parity_class === "binary_read" ? "binary_read" : "binary_write";
      const cli = commandEvidence(commands, expected, [intent]);
      return {
        status: operation && cli.status === "implemented" ? "implemented" : "declaration-pending",
        operation: operation?.id ?? null,
        cli,
      };
    }
    default:
      throw new Error(`unsupported parity class ${row.parity_class}`);
  }
}

function collectSourceArtifacts(value, output = []) {
  if (Array.isArray(value)) {
    for (const item of value) collectSourceArtifacts(item, output);
    return output;
  }
  if (!value || typeof value !== "object") return output;
  if (typeof value.source_url === "string") {
    output.push({
      source_url: value.source_url,
      sha256: value.sha256 ?? value.hash ?? null,
      document_version: value.info_version ?? value.version ?? value.api_version ?? value.source_version ?? null,
      format_version: value.openapi ?? value.format ?? null,
    });
  }
  for (const item of Object.values(value)) collectSourceArtifacts(item, output);
  return output;
}

function sourceDocumentForRow(lock, row) {
  const sourceDocuments = lock.rest?.source_documents;
  if (!Array.isArray(sourceDocuments)) return null;
  return sourceDocuments.find((document) => (Array.isArray(document.operations) ? document.operations : []).some((operation) => operation.id === row.source?.source_id && operation.source_location === row.source?.source_location && (!operation.citation_url || operation.citation_url === row.source?.source_url)));
}

function versionFromURL(sourceURL) {
  try {
    const url = new URL(sourceURL);
    return url.searchParams.get("version");
  } catch {
    return null;
  }
}

function sourceTrace(lock, row) {
  const lockRoot = lock.rest ?? lock;
  const artifacts = collectSourceArtifacts(lockRoot);
  const sourceDocument = sourceDocumentForRow(lock, row);
  const documentArtifact = sourceDocument?.artifact ? {
    source_url: sourceDocument.artifact.source_url ?? null,
    sha256: sourceDocument.artifact.sha256 ?? null,
    document_version: sourceDocument.artifact.info_version ?? sourceDocument.artifact.version ?? sourceDocument.artifact.api_version ?? sourceDocument.artifact.source_version ?? null,
    format_version: sourceDocument.artifact.openapi ?? sourceDocument.artifact.format ?? null,
  } : null;
  const exact = artifacts.filter((artifact) => artifact.source_url === row.source.source_url);
  const rootArtifact = artifacts.find((artifact) => artifact.source_url === lockRoot.source_url) ?? null;
  const hashed = documentArtifact?.sha256 ? documentArtifact : exact.find((artifact) => artifact.sha256) ?? rootArtifact;
  const version = documentArtifact?.document_version ?? exact.find((artifact) => artifact.document_version)?.document_version ??
    rootArtifact?.document_version ??
    versionFromURL(row.source.source_url);
  return {
    status: row.source.source_url && hashed?.sha256 ? "pinned" : "missing",
    source_url: row.source.source_url ?? null,
    source_lock: row.source.source_lock ?? null,
    source_id: row.source.source_id ?? null,
    source_location: row.source.source_location ?? null,
    sha256: hashed?.sha256 ?? null,
    document_version: version ?? null,
    version_status: version ? "published" : "not-published-in-pinned-source",
    format_version: documentArtifact?.format_version ?? exact.find((artifact) => artifact.format_version)?.format_version ?? rootArtifact?.format_version ?? null,
  };
}

function nonPrimarySurface(primary, target) {
  return {
    status: "not-applicable",
    provider_evidence: `The source-locked canonical mapping classifies this operation as ${primary}; it does not document a ${target} capability for this operation.`,
  };
}

function reverseETLEvidence(commands, row) {
  const expected = endpointKey(row.method, row.api_surface.path);
  const write = row.api_surface?.covered_by?.write ?? null;
  const eligibility = row.declaration?.reverse_etl_eligibility ?? null;
  if (!write) return nonPrimarySurface(row.parity_class, "reverse_etl");
  const cli = commandEvidence(commands, expected, ["reverse_etl"]);
  return {
    status: cli.status === "implemented" ? "foundation-pending" : cli.status,
    action: write,
    eligibility,
    cli,
    runtime_dependency: "generic typed destination persisted App/CLI dispatch and exact multi-action selection remain foundation dependencies",
  };
}

function increment(counter, status) {
  counter[status] = (counter[status] ?? 0) + 1;
}

const connectorLedger = [];
const operationLedger = [];
const operationBySourceID = new Map();
const actionDispositionEvidence = [];
const aggregate = {
  provider_operations: 0,
  source_trace: { pinned: 0, missing: 0, version_not_published: 0 },
  canonical_mapping: { mapped: 0, missing: 0 },
  primary_runtime_reachability: {},
  generated_cli_command: {},
  website_row: {},
  fixture_conformance: {},
};
let incomplete = false;

for (const connector of connectors) {
  const dir = path.join(root, "internal", "connectors", "defs", connector);
  const sourceDir = path.join(dir, "sources");
  const map = await readJSON(path.join(sourceDir, `${connector}-declaration-disposition.json`), { ledger_dispositions: [] });
  const lock = await readJSON(path.join(sourceDir, `${connector}-operation-source-lock.json`), {});
  const apiSurface = await readJSON(path.join(dir, "api_surface.json"), { endpoints: [] });
  const operations = (await readJSON(path.join(dir, "operations.json"), { operations: [] })).operations;
  const commands = (await readJSON(path.join(dir, "cli_surface.json"), { commands: [] })).commands;
  const writes = (await readJSON(path.join(dir, "writes.json"), { actions: [] })).actions;
  const transport = await readJSON(path.join(dir, "sync_transport.json"), null);
  const hasFixtureCheck = await fileExists(path.join(dir, "fixtures", "check.json"));
  const missing = { binary_download: 0, binary_upload: 0, direct_read: 0, direct_write: 0, etl: 0 };
  const represented = { binary_download: 0, binary_upload: 0, direct_read: 0, direct_write: 0, etl: 0 };
  const connectorGate = {
    provider_operations: map.ledger_dispositions.length,
    source_trace: { pinned: 0, missing: 0, version_not_published: 0 },
    canonical_mapping: { mapped: 0, missing: 0 },
    primary_runtime_reachability: {},
    generated_cli_command: {},
    website_row: { "declaration-pending": 0 },
    fixture_conformance: { "connector-level-only": 0 },
  };

  for (const row of map.ledger_dispositions) {
    const expected = endpointKey(row.method, row.api_surface.path);
    const primary = sourceClassSurface[row.parity_class];
    if (!primary) throw new Error(`${connector}: unsupported parity class ${row.parity_class}`);
    const trace = sourceTrace(lock, row);
    const canonical = apiSurface.endpoints.some((endpoint) => endpointMatches(endpoint, expected));
    const runtime = primaryRuntimeEvidence(operations, commands, transport, row);
    const anyCLI = commandEvidence(commands, expected);
    const surfaceEvidence = Object.fromEntries(surfaces.map((surface) => [
      surface,
      surface === "reverse_etl"
        ? reverseETLEvidence(commands, row)
        : surface === primary
          ? {
              status: runtime.status,
              provider_evidence: `source disposition parity_class=${row.parity_class}`,
              runtime,
            }
          : nonPrimarySurface(row.parity_class, surface),
    ]));
    const website = {
      status: "declaration-pending",
      row: null,
      reason: "Static CLI and website docs describe dynamic connector command surfaces but no generated provider-operation website row exists yet.",
    };
    const fixture = {
      status: "connector-level-only",
      fixture_check: hasFixtureCheck ? `internal/connectors/defs/${connector}/fixtures/check.json` : null,
      conformance_test: `go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/${connector}$' -count=1`,
      per_operation_binding: "declaration-pending",
      reason: "The connector fixture/conformance gate exists, but no source-operation-to-fixture binding is generated yet; connector-level conformance is not credited as per-operation proof.",
    };
    const record = {
      connector,
      operation: { method: row.method, path: row.path, parity_class: row.parity_class },
      source_trace: trace,
      canonical_mapping: { status: canonical ? "mapped" : "missing", api_surface: row.api_surface },
      surfaces: surfaceEvidence,
      generated_cli_command: anyCLI,
      generated_website_row: website,
      fixture_conformance: fixture,
      constraints: {
        declaration_state: row.state,
        declaration_status: row.declaration?.status ?? null,
        foundation: row.foundation ?? null,
        rejection: row.rejection ?? null,
        rule: "Scope, tier, destructive, and safety constraints must remain typed runtime metadata/confirmation and never substitute for a reachability disposition.",
      },
    };
    operationLedger.push(record);
    operationBySourceID.set(`${connector}:${trace.source_id}`, record);
    aggregate.provider_operations++;
    connectorGate.source_trace[trace.status]++;
    aggregate.source_trace[trace.status]++;
    if (trace.version_status === "not-published-in-pinned-source") {
      connectorGate.source_trace.version_not_published++;
      aggregate.source_trace.version_not_published++;
    }
    const canonicalStatus = canonical ? "mapped" : "missing";
    connectorGate.canonical_mapping[canonicalStatus]++;
    aggregate.canonical_mapping[canonicalStatus]++;
    increment(connectorGate.primary_runtime_reachability, runtime.status);
    increment(aggregate.primary_runtime_reachability, runtime.status);
    increment(connectorGate.generated_cli_command, anyCLI.status);
    increment(aggregate.generated_cli_command, anyCLI.status);
    connectorGate.website_row[website.status]++;
    increment(aggregate.website_row, website.status);
    connectorGate.fixture_conformance[fixture.status]++;
    increment(aggregate.fixture_conformance, fixture.status);
    if (runtime.status === "implemented") represented[primary]++;
    else missing[primary]++;
    if (trace.status !== "pinned" || !canonical || runtime.status !== "implemented" || anyCLI.status !== "implemented") incomplete = true;
  }

  const actionDispositions = map.summary?.reverse_etl_eligibility?.action_dispositions || [];
  actionDispositionEvidence.push(...actionDispositions.map((action) => ({ connector, action })));
  const actionDispositionByName = new Map(actionDispositions.map((disposition) => [disposition.action, disposition]));
  const destinationActions = transport?.destination_transport?.eligible_actions || [];
  const missingDestinationActions = writes.map((action) => action.name).filter((action) => !destinationActions.includes(action));
  const sourceInputPending = writes.map((action) => actionDispositionByName.get(action.name)).filter((disposition) => disposition?.source_input_binding?.state !== "source-bound").length;
  const reverseCommands = commands.filter((command) =>
    (command.availability === "implemented" || command.availability === "partial") &&
    command.intent === "reverse_etl" &&
    typeof command.write === "string",
  );
  const partialReverseCommandActions = new Set(reverseCommands
    .filter((command) => command.availability === "partial")
    .map((command) => command.write));
  const reverseCommandActions = new Set(reverseCommands.map((command) => command.write));
  const missingReverseCommandActions = writes.map((action) => action.name).filter((action) => !reverseCommandActions.has(action));
  const duplicateReverseCommandActions = reverseCommands.length - reverseCommandActions.size;
  const writeActionNames = new Set(writes.map((action) => action.name));
  const orphanReverseCommandActions = [...reverseCommandActions].filter((action) => !writeActionNames.has(action));
  const reverseETL = writes.length === 0
    ? {
        declaration: "not-applicable",
        provider_evidence: "No connector-owned typed write action exists in writes.json.",
        eligible_actions: 0,
        source_input_bindings_pending: 0,
        installed_command_actions: 0,
        implemented_installed_command_actions: 0,
        partial_installed_command_actions: 0,
        missing_installed_command_actions: 0,
        duplicate_installed_command_actions: 0,
        orphan_installed_command_actions: 0,
        application_dispatch: "not-applicable",
      }
    : {
        declaration: transport?.destination_transport ? "present" : "missing",
        eligible_actions: writes.length,
        bound_actions: destinationActions.length,
        missing_actions: missingDestinationActions.length,
        source_input_bindings_pending: sourceInputPending,
        installed_command_actions: reverseCommandActions.size,
        implemented_installed_command_actions: reverseCommandActions.size - partialReverseCommandActions.size,
        partial_installed_command_actions: partialReverseCommandActions.size,
        missing_installed_command_actions: missingReverseCommandActions.length,
        duplicate_installed_command_actions: duplicateReverseCommandActions,
        orphan_installed_command_actions: orphanReverseCommandActions.length,
        application_dispatch: "foundation-pending",
        action_selection: "foundation-pending",
      };
  if (writes.length > 0 && (
    !transport?.destination_transport ||
    missingDestinationActions.length > 0 ||
    sourceInputPending > 0 ||
    missingReverseCommandActions.length > 0 ||
    duplicateReverseCommandActions > 0 ||
    orphanReverseCommandActions.length > 0
  )) incomplete = true;
  connectorLedger.push({
    connector,
    source_operations: map.ledger_dispositions.length,
    represented,
    missing,
    reverse_etl: reverseETL,
    executable_commands: commands.filter((command) => command.availability === "implemented").length,
    captain_premerge_gate: connectorGate,
  });
}

function newFoundationGap(specification) {
  return {
    ...specification,
    affected_operations: new Map(),
    provenance_pending_actions: [],
  };
}

function addGapOperation(gap, record, affectedSurfaces, failingEvidence, action = null) {
  const sourceID = record.source_trace.source_id;
  // A provider operation identifier can cover a route family (for example Trello's
  // notification settings variants), so source id alone is not an operation key.
  const key = `${record.connector}:${sourceID}:${record.operation.method}:${record.operation.path}`;
  let operation = gap.affected_operations.get(key);
  if (!operation) {
    operation = {
      batch: batchForConnector(record.connector),
      connector: record.connector,
      provider_operation: {
        source_id: sourceID,
        method: record.operation.method,
        path: record.operation.path,
        parity_class: record.operation.parity_class,
      },
      source_trace: {
        source_url: record.source_trace.source_url,
        document_version: record.source_trace.document_version,
        version_status: record.source_trace.version_status,
        sha256: record.source_trace.sha256,
      },
      affected_surfaces: new Set(),
      failing_evidence: new Set(),
      status: "not-enabled-open-foundation-gap",
      actions: new Set(),
    };
    gap.affected_operations.set(key, operation);
  }
  for (const surface of affectedSurfaces) operation.affected_surfaces.add(surface);
  for (const evidence of failingEvidence.filter(Boolean)) operation.failing_evidence.add(evidence);
  if (action) operation.actions.add(action);
}

function pendingAction(gap, connector, action, reason) {
  gap.provenance_pending_actions.push({
    batch: batchForConnector(connector),
    connector,
    action,
    status: "declaration-pending-no-exact-provider-operation",
    reason,
  });
}

function serializeGap(gap) {
  const operations = [...gap.affected_operations.values()]
    .map((operation) => ({
      ...operation,
      affected_surfaces: [...operation.affected_surfaces].sort(),
      failing_evidence: [...operation.failing_evidence].sort(),
      actions: [...operation.actions].sort(),
    }))
    .sort((left, right) => `${left.connector}:${left.provider_operation.source_id}`.localeCompare(`${right.connector}:${right.provider_operation.source_id}`));
  const connectorFanOut = [...new Set(operations.map((operation) => operation.connector))].sort();
  return {
    id: gap.id,
    status: gap.status,
    missing_provider_neutral_capability: gap.missing_provider_neutral_capability,
    owning: gap.owning,
    validator_runtime_evidence: gap.validator_runtime_evidence,
    exact_closure_verification: gap.exact_closure_verification,
    fan_out: {
      connectors: connectorFanOut,
      connectors_count: connectorFanOut.length,
      provider_operations_count: operations.length,
      provenance_pending_actions_count: gap.provenance_pending_actions.length,
    },
    affected_operations: operations,
    provenance_pending_actions: gap.provenance_pending_actions,
  };
}

const foundationGaps = new Map([
  ["generic-typed-destination-app-dispatch", newFoundationGap({
    id: "generic-typed-destination-app-dispatch",
    status: "open-awaiting-foundation-head",
    missing_provider_neutral_capability: "Persisted App/CLI dispatch must select the exact preflighted declaration-owned generic typed destination without widening the selected executor, action, source binding, approval, or authorization route.",
    owning: { issue: "#4304", lane: "fm/cli-reverse-etl-destination-r1", last_observed_head: "c6f03c937c1f" },
    validator_runtime_evidence: ["internal/app/transport_dispatch.go:53-67 refuses to select declarative_typed_destination through the persisted App/CLI path."],
    exact_closure_verification: {
      commands: [
        "git fetch origin fm/cli-reverse-etl-destination-r1",
        "git merge --no-ff origin/fm/cli-reverse-etl-destination-r1",
        "git merge-base --is-ancestor origin/fm/cli-reverse-etl-destination-r1 HEAD",
        "go test -timeout 20m ./internal/app -run '^TestDefinitionTransportFactoriesRunTypedDestinationFromDefinition$' -count=1",
        "go build ./cmd/pm",
      ],
      observable: "An initialized isolated-project installed pm App/CLI run selects the declaration-owned generic destination and reaches the normal credential/approval boundary; it must not report a missing destination dispatcher or unknown command.",
    },
  })],
  ["declarative-typed-destination-action-multiplicity", newFoundationGap({
    id: "declarative-typed-destination-action-multiplicity",
    status: "open-awaiting-foundation-head",
    missing_provider_neutral_capability: "The closed destination contract must select one exact declaration-owned action and its exact input_fields binding per approved route, without accepting a caller-supplied operation, method, URL, or body.",
    owning: { issue: "#4304", lane: "fm/cli-reverse-etl-destination-r1", last_observed_head: "c6f03c937c1f" },
    validator_runtime_evidence: ["internal/connectors/sync_transport.go:388-415 rejects duplicate apply strategies and ApplyStrategyFor at :471-480 resolves only one action per mode."],
    exact_closure_verification: {
      commands: [
        "go test -timeout 20m ./internal/connectors -run '^TestSyncTransportDescriptorResolvesDeclaredApplyStrategy$' -count=1",
        "go test -timeout 20m ./internal/app -run '^TestDefinitionTransportFactoriesSelectDistinctTypedDestinationEvidence$' -count=1",
        "node .planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/verify-seven-surfaces.mjs",
      ],
      observable: "Every source-bound action has one exact declared selection route and the seven-surface assertion reports zero missing or duplicate destination actions.",
    },
  })],
  ["declaration-bound-rest-structured-body-cli", newFoundationGap({
    id: "declaration-bound-rest-structured-body-cli",
    status: "open-awaiting-engine-head",
    missing_provider_neutral_capability: "A closed declaration-bound REST structured-body CLI projection must validate only the exact pinned operation schema; it must not expose a raw body, URL, method, or action selector.",
    owning: { issue: "#4305", lane: "engine declaration-bound structured REST bodies" },
    validator_runtime_evidence: ["internal/connectors/commandrunner/runner.go:462-502 admits structured JSON only for a fixed GraphQL variable or reverse-ETL record field.", "internal/connectors/commandrunner/runner.go:1425-1455 rejects a direct-write raw body mapping."],
    exact_closure_verification: {
      commands: [
        "go test -timeout 20m ./internal/connectors/commandrunner -run '^TestBuildOperationDirectWriteCommandUsesTypedInputsAndPlanLifecycle$' -count=1",
        "go run ./cmd/connectorgen validate internal/connectors/defs --json",
        "go run ./cmd/connectorgen surface-sync --check",
        "node .planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/verify-seven-surfaces.mjs",
      ],
      observable: "Each affected REST write has a typed generated command that validates the declaration-owned body contract and reaches the normal plan/preview/approval execution boundary without a raw-body fallback.",
    },
  })],
  ["declaration-bound-bounded-binary-transfer", newFoundationGap({
    id: "declaration-bound-bounded-binary-transfer",
    status: "open-awaiting-engine-head",
    missing_provider_neutral_capability: "A declaration-owned binary upload/download contract must bind headers, provider path parameters, bounded transfer policy, output destination, and confirmation metadata without treating binary payloads as JSON REST bodies.",
    owning: { issue: "#4307", lane: "engine declaration-owned headers and bounded transfer operations" },
    validator_runtime_evidence: ["Affected source dispositions identify binary/document payload operations as not representable by the current declarative JSON path; no connector-local HTTP or raw-body workaround is permitted."],
    exact_closure_verification: {
      commands: [
        "go test -timeout 20m ./internal/connectors/commandrunner -run '^TestRunBinaryDownload(ReachesOperationDeclaredCap|PassesDestinationThrough|RequiresDestinationRoot)$' -count=1",
        "go run ./cmd/connectorgen validate internal/connectors/defs --json",
        "node .planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/verify-seven-surfaces.mjs",
      ],
      observable: "Every affected binary download/upload is represented by a bounded declaration-owned transfer command; no binary source operation remains disabled merely for payload shape, headers, privilege, or risk metadata.",
    },
  })],
  ["generated-provider-operation-website-row", newFoundationGap({
    id: "generated-provider-operation-website-row",
    status: "open-routing-required",
    missing_provider_neutral_capability: "Generate a website documentation row from each declaration-owned provider operation and its installed CLI command, retaining the source trace and current reachability state.",
    owning: { issue: null, lane: "unassigned platform/docs generator; firstmate routing required" },
    validator_runtime_evidence: ["The current static docs describe dynamic connector surfaces but contain no generated provider-operation website row for any of the 5,127 source-locked operations."],
    exact_closure_verification: {
      commands: [
        "make docs-check",
        "node .planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/verify-seven-surfaces.mjs",
      ],
      observable: "The operation evidence ledger records one generated website row identifier per provider operation and the docs check passes.",
    },
  })],
  ["source-operation-fixture-conformance-binding", newFoundationGap({
    id: "source-operation-fixture-conformance-binding",
    status: "open-routing-required",
    missing_provider_neutral_capability: "Generate an exact source-operation-to-fixture/conformance binding so connector-level conformance is not misreported as executable proof for every provider operation.",
    owning: { issue: null, lane: "unassigned connector certification generator; firstmate routing required" },
    validator_runtime_evidence: ["fixtures/check.json and TestConformance/<connector> are connector-level evidence only; no operation-specific source-to-fixture binding exists."],
    exact_closure_verification: {
      commands: [
        "go test -timeout 20m ./internal/connectors/conformance -run 'TestConformance/(grafana|trello|slack|n8n|google-calendar|gmail|twilio|amazon-sqs|elasticsearch|gong|google-ads|facebook-marketing|linkedin-ads|aircall|xero|paypal-transaction|gocardless|amazon-seller-partner|miro)$' -count=1",
        "node .planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/verify-seven-surfaces.mjs",
      ],
      observable: "Every provider operation names a concrete fixture or conformance case and the target conformance suite passes without relying on a connector-wide implicit credit.",
    },
  })],
  ["declaration-bound-scalar-union-cli", newFoundationGap({
    id: "declaration-bound-scalar-union-cli",
    status: "open-routing-required",
    missing_provider_neutral_capability: "A closed declaration-bound CLI scalar-union value must encode only the exact documented field alternatives; it cannot accept an untyped body or bypass normal command validation.",
    owning: { issue: null, lane: "unassigned engine CLI projection; firstmate routing required" },
    validator_runtime_evidence: ["internal/connectors/commandrunner/runner.go:1737-1816 accepts only concrete scalar flag types and cmd/connectorgen/batch_materialize.go:1552-1581 rejects non-single-type flags."],
    exact_closure_verification: {
      commands: [
        "go test -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1",
        "go run ./cmd/connectorgen surface-sync --check",
        "node .planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/verify-seven-surfaces.mjs",
      ],
      observable: "The 82 currently partial actions become declaration-bound implemented commands with no raw request body and no partial availability left for a scalar-union field.",
    },
  })],
]);

for (const record of operationLedger) {
  const reverseGap = record.surfaces.reverse_etl.eligibility?.foundation_gap;
  if (reverseGap?.id && foundationGaps.has(reverseGap.id)) {
    addGapOperation(
      foundationGaps.get(reverseGap.id),
      record,
      ["reverse_etl"],
      [reverseGap.evidence, record.surfaces.reverse_etl.runtime_dependency],
      record.surfaces.reverse_etl.action,
    );
  }
  if (record.operation.parity_class === "binary_read" || record.operation.parity_class === "binary_write") {
    const surface = sourceClassSurface[record.operation.parity_class];
    addGapOperation(
      foundationGaps.get("declaration-bound-bounded-binary-transfer"),
      record,
      [surface, "generated_cli_command"],
      [record.canonical_mapping.api_surface?.operation?.reason, `seven-surface primary runtime=${record.surfaces[surface].status}`],
    );
  }
  addGapOperation(
    foundationGaps.get("generated-provider-operation-website-row"),
    record,
    ["generated_website_row"],
    [record.generated_website_row.reason],
  );
  addGapOperation(
    foundationGaps.get("source-operation-fixture-conformance-binding"),
    record,
    ["fixture_conformance"],
    [record.fixture_conformance.reason],
  );
}

const writeCommandReport = await readJSON(path.join(phase, "INSTALLED-WRITE-COMMAND-GENERATION.json"), { connectors: [] });
const partialReasons = new Map();
for (const report of writeCommandReport.connectors) {
  for (const partial of report.partial ?? []) partialReasons.set(`${report.connector}:${partial.action}`, partial.reason);
}
for (const entry of actionDispositionEvidence) {
  const action = entry.action;
  const sourceIDs = action.provider_operation_binding?.source_ids ?? [];
  const multiplicityGap = action.destination_binding?.foundation_gap;
  const partialReason = partialReasons.get(`${entry.connector}:${action.action}`);
  for (const sourceID of sourceIDs) {
    const record = operationBySourceID.get(`${entry.connector}:${sourceID}`);
    if (!record) throw new Error(`${entry.connector}: action ${action.action} references missing source operation ${sourceID}`);
    if (multiplicityGap?.id === "declarative-typed-destination-action-multiplicity") {
      addGapOperation(
        foundationGaps.get(multiplicityGap.id),
        record,
        ["reverse_etl"],
        [multiplicityGap.evidence],
        action.action,
      );
    }
    addGapOperation(
      foundationGaps.get("declaration-bound-rest-structured-body-cli"),
      record,
      ["direct_write", "generated_cli_command"],
      ["The source-bound typed action has no enabled declaration-bound REST direct-write command body projection."],
      action.action,
    );
    if (partialReason) {
      addGapOperation(
        foundationGaps.get("declaration-bound-scalar-union-cli"),
        record,
        ["reverse_etl", "generated_cli_command"],
        [`${action.action}: ${partialReason}`],
        action.action,
      );
    }
  }
  if (sourceIDs.length === 0 && multiplicityGap?.id === "declarative-typed-destination-action-multiplicity") {
    pendingAction(
      foundationGaps.get(multiplicityGap.id),
      entry.connector,
      action.action,
      action.provider_operation_binding?.detail ?? "No exact provider-operation source binding is available.",
    );
  }
}

const serializedFoundationGaps = [...foundationGaps.values()].map(serializeGap);
const openGapIDsByOperation = new Map();
for (const gap of serializedFoundationGaps) {
  for (const operation of gap.affected_operations) {
    const key = `${operation.connector}:${operation.provider_operation.source_id}:${operation.provider_operation.method}:${operation.provider_operation.path}`;
    const ids = openGapIDsByOperation.get(key) ?? [];
    ids.push(gap.id);
    openGapIDsByOperation.set(key, ids);
  }
}
for (const record of operationLedger) {
  const key = `${record.connector}:${record.source_trace.source_id}:${record.operation.method}:${record.operation.path}`;
  const openGapIDs = (openGapIDsByOperation.get(key) ?? []).sort();
  record.merge_readiness = openGapIDs.length === 0
    ? { status: "not-merge-ready-pending-other-evidence", open_foundation_gap_ids: [] }
    : { status: "not-enabled-open-foundation-gap", open_foundation_gap_ids: openGapIDs };
}
const batchRollup = Object.fromEntries(["batch-2", "batch-3"].map((batch) => {
  const operationKeys = new Set();
  const gapIDs = [];
  const connectorNames = new Set();
  for (const gap of serializedFoundationGaps) {
    const affected = gap.affected_operations.filter((operation) => operation.batch === batch);
    if (affected.length === 0) continue;
    gapIDs.push(gap.id);
    for (const operation of affected) {
      operationKeys.add(`${operation.connector}:${operation.provider_operation.source_id}:${operation.provider_operation.method}:${operation.provider_operation.path}`);
      connectorNames.add(operation.connector);
    }
  }
  return [batch, {
    status: "not-merge-ready-open-foundation-gaps",
    open_gap_ids: gapIDs.sort(),
    affected_connectors: [...connectorNames].sort(),
    unique_provider_operations: operationKeys.size,
  }];
}));
const portfolioOperationKeys = new Set(serializedFoundationGaps.flatMap((gap) => gap.affected_operations.map((operation) => `${operation.connector}:${operation.provider_operation.source_id}:${operation.provider_operation.method}:${operation.provider_operation.path}`)));
const foundationGapLedger = {
  schema_version: 1,
  issue: 4289,
  policy: "A shared gap is deduplicated by stable id. Every affected provider operation remains not enabled for the affected surface and cannot contribute to a merge-ready verdict while that gap is open. Unauthored connector declarations remain declaration-pending and are not relabelled as foundation gaps.",
  status: "not-merge-ready-open-foundation-gaps",
  gaps: serializedFoundationGaps,
  rollups: {
    batches: batchRollup,
    portfolio: {
      status: "not-merge-ready-open-foundation-gaps",
      open_gap_ids: serializedFoundationGaps.map((gap) => gap.id),
      open_gap_count: serializedFoundationGaps.length,
      unique_affected_provider_operations: portfolioOperationKeys.size,
      total_provider_operations: aggregate.provider_operations,
      zero_silent_omissions: aggregate.provider_operations === operationLedger.length,
    },
  },
};
if (serializedFoundationGaps.some((gap) => gap.status.startsWith("open"))) incomplete = true;

const zeroSilentOmissions = aggregate.provider_operations === operationLedger.length;
if (!zeroSilentOmissions) incomplete = true;
const captainPremergeGate = {
  status: incomplete ? "incomplete" : "complete",
  required_evidence: aggregate,
  foundation_gaps: {
    status: foundationGapLedger.status,
    open_gap_ids: foundationGapLedger.rollups.portfolio.open_gap_ids,
    unique_affected_provider_operations: foundationGapLedger.rollups.portfolio.unique_affected_provider_operations,
  },
  zero_silent_omissions: {
    status: zeroSilentOmissions ? "proved" : "missing",
    provider_operations: aggregate.provider_operations,
    operation_records: operationLedger.length,
  },
  hold: "Website operation rows, source-operation fixture bindings, complete primary-surface reachability, and current typed-destination foundations are required before this gate can pass.",
};

await writeFile(path.join(phase, "FOUNDATION-GAP-LEDGER.json"), `${JSON.stringify(foundationGapLedger, null, 2)}\n`);
await writeFile(path.join(phase, "SEVEN-SURFACE-LEDGER.json"), `${JSON.stringify({
  schema_version: 2,
  foundation_dispatch: "pending latest fm/cli-reverse-etl-destination-r1",
  captain_premerge_gate: captainPremergeGate,
  connectors: connectorLedger,
}, null, 2)}\n`);
await writeFile(path.join(phase, "OPERATION-EVIDENCE-LEDGER.json"), `${JSON.stringify({
  schema_version: 1,
  issue: 4289,
  rule: "Each provider-defined source operation is represented exactly once. A non-primary surface is not-applicable only when the source-locked canonical parity class proves that capability absent for that operation; safety and scope are preserved as constraints, not disablement.",
  captain_premerge_gate: captainPremergeGate,
  operations: operationLedger,
}, null, 2)}\n`);
if (incomplete) {
  throw new Error("captain pre-merge and seven-surface reconciliation are incomplete; see SEVEN-SURFACE-LEDGER.json, OPERATION-EVIDENCE-LEDGER.json, and FOUNDATION-GAP-LEDGER.json");
}
console.log(`verified ${connectorLedger.length} connectors / ${operationLedger.length} provider operations across all required surfaces and evidence gates`);
