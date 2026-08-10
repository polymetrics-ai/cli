#!/usr/bin/env node

import { createHash } from "node:crypto";
import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(SCRIPT_DIR, "..");
const CONNECTOR_DIR = path.join(ROOT, "internal/connectors/defs/github");
const PHASE_DIR = path.join(ROOT, ".planning/phases/github-parity-extract-r1");
const SOURCE_FILES = {
  surface: "internal/connectors/defs/github/api_surface.json",
  streams: "internal/connectors/defs/github/streams.json",
  writes: "internal/connectors/defs/github/writes.json",
  operations: "internal/connectors/defs/github/operations.json",
  cli: "internal/connectors/defs/github/cli_surface.json",
};

const EXPECTED_ARTIFACTS = {
  operation: "OPERATION-PROOF-LEDGER.json",
  command: "COMMAND-PROOF-LEDGER.json",
};

async function readJSON(file) {
  return JSON.parse(await readFile(file, "utf8"));
}

function sourcePath(name) {
  return path.join(ROOT, SOURCE_FILES[name]);
}

function normalizePath(value) {
  return String(value || "")
    .replace(/\{\{\s*(?:config|record)\.([A-Za-z0-9_]+)\s*\}\}/g, "{$1}")
    .replace(/\{\{\s*([A-Za-z0-9_]+)\s*\}\}/g, "{$1}")
    .replace(/\s+/g, " ")
    .trim();
}

function endpointKey(method, value) {
  return `${String(method || "").toUpperCase()} ${normalizePath(value)}`.trim();
}

function sortedUnique(values) {
  return [...new Set(values.filter((value) => String(value || "").trim() !== ""))].sort((a, b) =>
    a.localeCompare(b),
  );
}

function sha256(value) {
  return createHash("sha256").update(JSON.stringify(value)).digest("hex");
}

function coverageOf(endpoint) {
  const covered = endpoint.covered_by;
  if (!covered) {
    return undefined;
  }
  const bindings = [];
  if (covered.stream) {
    bindings.push({ kind: "stream", targets: [covered.stream] });
  }
  if (covered.write) {
    bindings.push({ kind: "write", targets: [covered.write] });
  }
  if (Array.isArray(covered.writes)) {
    bindings.push({ kind: "write", targets: covered.writes });
  }
  if (covered.direct_read) {
    bindings.push({ kind: "direct_read", targets: [covered.direct_read] });
  }
  if (Array.isArray(covered.direct_reads)) {
    bindings.push({ kind: "direct_read", targets: covered.direct_reads });
  }
  if (Array.isArray(covered.operations)) {
    bindings.push({ kind: "operation", targets: covered.operations });
  }
  if (bindings.length !== 1) {
    throw new Error(
      `endpoint ${endpointKey(endpoint.method, endpoint.path)} has ${bindings.length} covered_by bindings; expected one kind`,
    );
  }
  const binding = bindings[0];
  return { kind: binding.kind, targets: sortedUnique(binding.targets) };
}

function graphQLOperationName(operation) {
  return operation.graphql?.operation_name || "";
}

function operationEndpointKeys(operation) {
  if (operation.rest?.method && operation.rest?.path) {
    return [endpointKey(operation.rest.method, operation.rest.path)];
  }
  if (graphQLOperationName(operation)) {
    return [endpointKey("GRAPHQL", graphQLOperationName(operation))];
  }
  return [];
}

function commandEndpointKeys(command) {
  return (command.api_surface || []).map((ref) => endpointKey(ref.method, ref.path));
}

function actionEndpointKey(action) {
  return endpointKey(action.method, action.path);
}

function operationEndpointLinks(operation, endpointsByKey) {
  return sortedUnique(operationEndpointKeys(operation).flatMap((key) => endpointsByKey.get(key) || []));
}

function commandLinksForEndpoint(endpoint, coverage, commandsByPath, operationsByID, operationsByEndpoint) {
  const endpointKeyValue = endpointKey(endpoint.method, endpoint.path);
  const paths = new Set();
  const operations = new Set(operationsByEndpoint.get(endpointKeyValue) || []);

  if (coverage?.kind === "operation") {
    for (const operationID of coverage.targets) operations.add(operationID);
  }

  if (coverage) {
    for (const target of coverage.targets) {
      if (coverage.kind === "direct_read") {
        paths.add(target);
      }
      if (coverage.kind === "stream") {
        for (const command of commandsByPath.values()) {
          if (command.stream === target) paths.add(command.path);
        }
      }
      if (coverage.kind === "write") {
        for (const command of commandsByPath.values()) {
          if (command.write === target) paths.add(command.path);
        }
      }
    }
  }

  for (const operationID of operations) {
    for (const command of commandsByPath.values()) {
      if (command.operation === operationID) paths.add(command.path);
    }
  }

  // A GraphQL operation can be represented by a stream command rather than an
  // operation field; match its declared API operation name as well.
  for (const command of commandsByPath.values()) {
    if (commandEndpointKeys(command).includes(endpointKeyValue)) paths.add(command.path);
  }

  // Keep the argument in this helper explicit: it makes an unknown operation
  // fail in model validation rather than silently disappearing from evidence.
  void operationsByID;
  return sortedUnique([...paths]);
}

function genericOnly(streams, actions, commands) {
  const streamRefs = new Map(streams.map((stream) => [stream.name, []]));
  const actionRefs = new Map(actions.map((action) => [action.name, []]));
  for (const command of commands) {
    if (command.stream && streamRefs.has(command.stream)) streamRefs.get(command.stream).push(command.path);
    if (command.write && actionRefs.has(command.write)) actionRefs.get(command.write).push(command.path);
  }
  return {
    streams: [...streamRefs]
      .filter(([, paths]) => paths.length === 0)
      .map(([name]) => ({ name, route: "generic_etl", evidence: "pm etl" }))
      .sort((a, b) => a.name.localeCompare(b.name)),
    writeActions: [...actionRefs]
      .filter(([, paths]) => paths.length === 0)
      .map(([name]) => ({ name, route: "generic_reverse_etl", evidence: "pm reverse" }))
      .sort((a, b) => a.name.localeCompare(b.name)),
  };
}

function makeEndpointLedger(bundle, indexes) {
  return bundle.surface.endpoints.map((endpoint, index) => {
    const coverage = coverageOf(endpoint);
    const blocked = endpoint.operation;
    const operationIDs = sortedUnique([
      ...(indexes.operationsByEndpoint.get(endpointKey(endpoint.method, endpoint.path)) || []),
      ...(coverage?.kind === "operation" ? coverage.targets : []),
    ]);
    const commands = commandLinksForEndpoint(
      endpoint,
      coverage,
      indexes.commandsByPath,
      indexes.operationsByID,
      indexes.operationsByEndpoint,
    );
    const row = {
      endpoint_id: endpointKey(endpoint.method, endpoint.path),
      source: { file: SOURCE_FILES.surface, index },
      method: String(endpoint.method || "").toUpperCase(),
      path: normalizePath(endpoint.path),
      disposition: coverage ? "covered" : "blocked",
      links: {
        streams: coverage?.kind === "stream" ? coverage.targets : [],
        write_actions: coverage?.kind === "write" ? coverage.targets : [],
        operations: operationIDs,
        commands,
      },
      evidence: {
        static: "source-derived covered_by/operation contract validated by github-parity-proof",
        deterministic: "required; see provider-double proof ledger",
      },
    };
    if (coverage) {
      row.coverage = coverage;
    } else if (blocked) {
      row.block = {
        model: blocked.model,
        status: blocked.status,
        reason: blocked.reason,
        source_url: blocked.source_url || "",
        duplicate_of: blocked.duplicate_of || "",
      };
    } else {
      row.block = { model: "invalid", status: "missing", reason: "endpoint has no coverage or typed block" };
    }
    return row;
  });
}

function commandRoute(command, indexes) {
  if (command.stream) return { kind: "etl", target: command.stream, executor: "connector stream" };
  if (command.write) return { kind: "reverse_etl", target: command.write, executor: "generic reverse ETL" };
  if (command.operation) return { kind: "operation", target: command.operation, executor: "declared operation" };
  if (command.api_surface?.length) {
    return { kind: "direct_read", target: commandEndpointKeys(command), executor: "declared endpoint" };
  }
  if (command.intent === "etl") return { kind: "etl", target: "unbound", executor: "generic ETL" };
  if (command.intent === "reverse_etl") return { kind: "reverse_etl", target: "unbound", executor: "generic reverse ETL" };
  void indexes;
  return { kind: "declared_block_or_local", target: [], executor: "no connector executor" };
}

function makeCommandLedger(bundle, indexes, reachability) {
  return bundle.cli.commands.map((command, index) => {
    const route = commandRoute(command, indexes);
    const refs = commandEndpointKeys(command);
    const evidence = reachability?.records?.[command.path];
    const row = {
      command: command.path,
      source: { file: SOURCE_FILES.cli, index },
      intent: command.intent,
      availability: command.availability,
      route,
      api_surface: refs,
      operation: command.operation || "",
      write_action: command.write || "",
      stream: command.stream || "",
      evidence: {
        static: "source-derived command contract validated by github-parity-proof",
        binary: evidence || {
          state: "not_run",
          reason: "current-head binary sweep has not been attached",
        },
      },
    };
    if (command.notes) row.notes = command.notes;
    return row;
  });
}

function normalizeReachability(reachability) {
  if (!reachability) return undefined;
  if (!Array.isArray(reachability.records)) return reachability;
  return {
    ...reachability,
    records: Object.fromEntries(reachability.records.map((record) => [record.command, record])),
  };
}

function makeIndexes(bundle) {
  const endpointsByKey = new Map();
  for (const endpoint of bundle.surface.endpoints) {
    const key = endpointKey(endpoint.method, endpoint.path);
    const list = endpointsByKey.get(key) || [];
    list.push(key);
    endpointsByKey.set(key, list);
  }
  const commandsByPath = new Map(bundle.cli.commands.map((command) => [command.path, command]));
  const operationsByID = new Map(bundle.operations.operations.map((operation) => [operation.id, operation]));
  const operationsByEndpoint = new Map();
  for (const operation of bundle.operations.operations) {
    for (const key of operationEndpointKeys(operation)) {
      const list = operationsByEndpoint.get(key) || [];
      list.push(operation.id);
      operationsByEndpoint.set(key, list);
    }
  }
  return { endpointsByKey, commandsByPath, operationsByID, operationsByEndpoint };
}

function buildBundleFromFiles(files) {
  return {
    surface: files.surface,
    streams: files.streams,
    writes: files.writes,
    operations: files.operations,
    cli: files.cli,
  };
}

function countBundle(bundle) {
  const endpoints = bundle.surface.endpoints.length;
  const coveredEndpoints = bundle.surface.endpoints.filter((endpoint) => coverageOf(endpoint)).length;
  const blockedEndpoints = bundle.surface.endpoints.filter((endpoint) => endpoint.operation).length;
  return {
    endpoints,
    coveredEndpoints,
    blockedEndpoints,
    streams: bundle.streams.streams.length,
    writeActions: bundle.writes.actions.length,
    operations: bundle.operations.operations.length,
    commands: bundle.cli.commands.length,
    implementedCommands: bundle.cli.commands.filter((command) => command.availability === "implemented").length,
    partialCommands: bundle.cli.commands.filter((command) => command.availability === "partial").length,
  };
}

export function validateProofModel(model) {
  const { bundle, endpointLedger, commandLedger, operationLedger, genericOnly, counts } = model;
  if (!bundle?.surface?.endpoints?.length) throw new Error("source bundle has no endpoints");
  if (endpointLedger.length !== counts.endpoints) {
    throw new Error(`endpoint ledger has ${endpointLedger.length} rows, want ${counts.endpoints}`);
  }
  if (commandLedger.length !== counts.commands) {
    throw new Error(`command ledger has ${commandLedger.length} rows, want ${counts.commands}`);
  }
  if (operationLedger.length !== counts.operations) {
    throw new Error(`operation ledger has ${operationLedger.length} rows, want ${counts.operations}`);
  }

  const streams = new Set(bundle.streams.streams.map((stream) => stream.name));
  const actions = new Set(bundle.writes.actions.map((action) => action.name));
  const operations = new Set(bundle.operations.operations.map((operation) => operation.id));
  const commands = new Set(bundle.cli.commands.map((command) => command.path));
  const endpoints = new Set(endpointLedger.map((row) => row.endpoint_id));
  if (endpoints.size !== endpointLedger.length) throw new Error("endpoint ledger contains duplicate endpoint IDs");

  for (const row of endpointLedger) {
    for (const name of row.links.streams) if (!streams.has(name)) throw new Error(`unknown stream ${JSON.stringify(name)}`);
    for (const name of row.links.write_actions) if (!actions.has(name)) throw new Error(`unknown write action ${JSON.stringify(name)}`);
    for (const id of row.links.operations) if (!operations.has(id)) throw new Error(`unknown operation ${JSON.stringify(id)}`);
    for (const command of row.links.commands) if (!commands.has(command)) throw new Error(`unknown command ${JSON.stringify(command)}`);
    if (row.disposition === "covered" && !row.coverage?.targets?.length) throw new Error(`covered endpoint ${row.endpoint_id} has no targets`);
    if (row.disposition === "blocked" && !row.block?.reason) throw new Error(`blocked endpoint ${row.endpoint_id} has no reason`);
  }

  for (const row of commandLedger) {
    if (row.operation && !operations.has(row.operation)) throw new Error(`unknown operation ${JSON.stringify(row.operation)}`);
    if (row.stream && !streams.has(row.stream)) throw new Error(`unknown stream ${JSON.stringify(row.stream)}`);
    if (row.write_action && !actions.has(row.write_action)) throw new Error(`unknown write action ${JSON.stringify(row.write_action)}`);
    for (const key of row.api_surface) if (!endpoints.has(key)) throw new Error(`unknown endpoint ${JSON.stringify(key)}`);
  }

  for (const row of operationLedger) {
    if (!operations.has(row.operation_id)) throw new Error(`operation ledger references unknown operation ${JSON.stringify(row.operation_id)}`);
    for (const key of row.endpoint_refs) if (!endpoints.has(key)) throw new Error(`operation ${row.operation_id} references unknown endpoint ${JSON.stringify(key)}`);
    for (const command of row.command_refs) if (!commands.has(command)) throw new Error(`operation ${row.operation_id} references unknown command ${JSON.stringify(command)}`);
    if (!row.endpoint_refs.length && !row.command_refs.length) throw new Error(`operation ${row.operation_id} has no executable or endpoint binding`);
  }

  if (genericOnly.streams.length !== counts.streams - new Set(commandLedger.filter((row) => row.stream).map((row) => row.stream)).size) {
    throw new Error("generic-only stream accounting does not match declared stream references");
  }
  if (genericOnly.writeActions.length !== counts.writeActions - new Set(commandLedger.filter((row) => row.write_action).map((row) => row.write_action)).size) {
    throw new Error("generic-only write accounting does not match declared write references");
  }
  return true;
}

export async function loadBundle() {
  const files = {};
  for (const name of Object.keys(SOURCE_FILES)) files[name] = await readJSON(sourcePath(name));
  return buildBundleFromFiles(files);
}

export async function buildProofModel(bundle = undefined, reachability = undefined) {
  const source = bundle || (await loadBundle());
  const binaryEvidence = normalizeReachability(reachability);
  const indexes = makeIndexes(source);
  const endpointLedger = makeEndpointLedger(source, indexes);
  const operationLedger = source.operations.operations.map((operation) => ({
    operation_id: operation.id,
    kind: operation.kind,
    source: { file: SOURCE_FILES.operations },
    endpoint_refs: operationEndpointLinks(operation, indexes.endpointsByKey),
    command_refs: source.cli.commands.filter((command) => command.operation === operation.id).map((command) => command.path).sort(),
    exercise_route: operation.kind.startsWith("graphql") ? "graphql operation executor" : "declared operation executor",
    evidence: "source-derived operation binding validated by github-parity-proof",
  }));
  const generic = genericOnly(source.streams.streams, source.writes.actions, source.cli.commands);
  const model = {
    schema_version: 1,
    connector: "github",
    bundle_source: SOURCE_FILES,
    source_sha256: sha256(source),
    bundle: source,
    counts: countBundle(source),
    endpointLedger,
    operationLedger,
    commandLedger: makeCommandLedger(source, indexes, binaryEvidence),
    genericOnly: generic,
  };
  return model;
}

function publicLedger(model) {
  return {
    schema_version: model.schema_version,
    connector: model.connector,
    source_sha256: model.source_sha256,
    counts: model.counts,
    generic_only: model.genericOnly,
    endpoints: model.endpointLedger,
    operations: model.operationLedger,
  };
}

function publicCommands(model) {
  return {
    schema_version: model.schema_version,
    connector: model.connector,
    source_sha256: model.source_sha256,
    counts: model.counts,
    generic_only: model.genericOnly,
    commands: model.commandLedger,
  };
}

export async function writeProofArtifacts(model, outputDir = PHASE_DIR) {
  await writeFile(path.join(outputDir, EXPECTED_ARTIFACTS.operation), `${JSON.stringify(publicLedger(model), null, 2)}\n`);
  await writeFile(path.join(outputDir, EXPECTED_ARTIFACTS.command), `${JSON.stringify(publicCommands(model), null, 2)}\n`);
}

async function main() {
  const args = new Set(process.argv.slice(2));
  if (!["--check", "--write"].some((flag) => args.has(flag))) {
    throw new Error("use --check or --write");
  }
  let reachability;
  try {
    reachability = await readJSON(path.join(PHASE_DIR, "COMMAND-REACHABILITY.json"));
  } catch (error) {
    if (error?.code !== "ENOENT") throw error;
  }
  const model = await buildProofModel(undefined, reachability);
  validateProofModel(model);
  if (args.has("--write")) await writeProofArtifacts(model);
  process.stdout.write(`${JSON.stringify(model.counts)} generic_streams=${model.genericOnly.streams.length} generic_writes=${model.genericOnly.writeActions.length}\n`);
}

if (path.resolve(process.argv[1] || "") === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    process.stderr.write(`github parity proof: ${error instanceof Error ? error.message : String(error)}\n`);
    process.exitCode = 1;
  });
}
