#!/usr/bin/env node

import { createHash } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(SCRIPT_DIR, "..");
const GITHUB_DEFS = path.join(ROOT, "internal", "connectors", "defs", "github");
const DEFAULT_LOCK = path.join(GITHUB_DEFS, "sources", "github-operation-source-lock.json");
const DEFAULT_LEDGER = path.join(ROOT, ".planning", "phases", "github-parity-extract-r1", "GITHUB-COMBINED-OPERATION-LEDGER.json");
const REST_COMMIT = "b26c240ded1c8b79cb0fb09dee4a21239061fa23";
const REST_URL = `https://raw.githubusercontent.com/github/rest-api-description/${REST_COMMIT}/descriptions/api.github.com/api.github.com.json`;
const GRAPHQL_URL = "https://docs.github.com/public/ghec/schema.docs.graphql";
const HTTP_METHODS = ["get", "post", "put", "patch", "delete"];
const REQUIRED_ROW_FIELDS = ["id", "protocol", "source", "pm", "implementation", "auth", "fixture", "safety", "assertion", "cleanup", "terminal_evidence"];

function isPlainObject(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function requireObject(value, label) {
  if (!isPlainObject(value)) throw new Error(`${label} must be an object`);
  return value;
}

function requireString(value, label) {
  if (typeof value !== "string" || value.trim() === "") throw new Error(`${label} must be a non-empty string`);
  return value.trim();
}

function requireDate(value, label) {
  const date = requireString(value, label);
  if (!/^\d{4}-\d{2}-\d{2}$/u.test(date)) throw new Error(`${label} must be YYYY-MM-DD`);
  return date;
}

function sha256(value) {
  return createHash("sha256").update(value).digest("hex");
}

function byteLength(value) {
  return Buffer.byteLength(value, "utf8");
}

function normalizePath(value) {
  return String(value || "")
    .replace(/\{\{\s*(?:config|record)\.([A-Za-z0-9_]+)\s*\}\}/g, "{$1}")
    .replace(/\s+/g, " ")
    .trim();
}

function endpointKey(method, resourcePath) {
  return `${String(method || "").toUpperCase()} ${normalizePath(resourcePath)}`;
}

function sourceLocation(resourcePath, method) {
  return `paths[${JSON.stringify(resourcePath)}].${method}`;
}

function sorted(values) {
  return [...values].sort((left, right) => left.localeCompare(right));
}

function uniqueSorted(values) {
  return sorted(new Set(values.filter((value) => String(value || "").trim() !== "")));
}

/** Enumerate the provider's documented HTTP operation keys, not bundle coverage. */
export function enumerateRESTOperations(restDocument) {
  const document = requireObject(restDocument, "REST OpenAPI document");
  const paths = requireObject(document.paths, "REST OpenAPI document.paths");
  const operations = [];
  for (const resourcePath of sorted(Object.keys(paths))) {
    const item = requireObject(paths[resourcePath], `REST path ${resourcePath}`);
    for (const method of HTTP_METHODS) {
      const operation = item[method];
      if (operation === undefined) continue;
      requireObject(operation, `REST operation ${method.toUpperCase()} ${resourcePath}`);
      const operationID = requireString(operation.operationId, `REST operation ${method.toUpperCase()} ${resourcePath}.operationId`);
      operations.push({
        id: `github.rest.${operationID}`,
        protocol: "rest",
        method: method.toUpperCase(),
        path: normalizePath(resourcePath),
        operation_id: operationID,
        deprecated: operation.deprecated === true,
        source_location: sourceLocation(resourcePath, method),
      });
    }
  }
  const ids = new Set();
  for (const operation of operations) {
    if (ids.has(operation.id)) throw new Error(`REST source has duplicate stable operation ID ${operation.id}`);
    ids.add(operation.id);
  }
  return operations;
}

function advancePosition(state, character) {
  if (character === "\n") {
    state.line += 1;
    state.column = 1;
  } else {
    state.column += 1;
  }
  state.index += 1;
}

/** A deliberately small SDL lexer: enough to locate schema declarations, never execute GraphQL. */
function tokenizeGraphQL(schema) {
  const text = requireString(schema, "GraphQL schema");
  const state = { index: 0, line: 1, column: 1 };
  const tokens = [];
  const isNameStart = (character) => /[_A-Za-z]/u.test(character);
  const isNameContinue = (character) => /[_0-9A-Za-z]/u.test(character);
  const skipQuoted = (triple) => {
    const delimiter = triple ? '"""' : '"';
    for (let consumed = 0; consumed < delimiter.length; consumed += 1) advancePosition(state, text[state.index]);
    while (state.index < text.length) {
      if (text.startsWith(delimiter, state.index)) {
        for (let consumed = 0; consumed < delimiter.length; consumed += 1) advancePosition(state, text[state.index]);
        return;
      }
      if (!triple && text[state.index] === "\\" && state.index + 1 < text.length) {
        advancePosition(state, text[state.index]);
        advancePosition(state, text[state.index]);
        continue;
      }
      advancePosition(state, text[state.index]);
    }
    throw new Error("GraphQL schema has an unterminated quoted string");
  };

  while (state.index < text.length) {
    const character = text[state.index];
    if (/\s/u.test(character)) {
      advancePosition(state, character);
      continue;
    }
    if (character === "#") {
      while (state.index < text.length && text[state.index] !== "\n") advancePosition(state, text[state.index]);
      continue;
    }
    if (text.startsWith('"""', state.index)) {
      skipQuoted(true);
      continue;
    }
    if (character === '"') {
      skipQuoted(false);
      continue;
    }
    const line = state.line;
    const column = state.column;
    if (isNameStart(character)) {
      let value = "";
      while (state.index < text.length && isNameContinue(text[state.index])) {
        value += text[state.index];
        advancePosition(state, text[state.index]);
      }
      tokens.push({ kind: "name", value, line, column });
      continue;
    }
    if ("!$():=[]@&|{}".includes(character)) {
      tokens.push({ kind: "punct", value: character, line, column });
      advancePosition(state, character);
      continue;
    }
    // Numeric defaults and comma separators do not influence root-field boundaries.
    advancePosition(state, character);
  }
  return tokens;
}

function findBalanced(tokens, start, open, close) {
  if (tokens[start]?.value !== open) throw new Error(`expected ${open} while parsing GraphQL schema`);
  let depth = 0;
  for (let index = start; index < tokens.length; index += 1) {
    if (tokens[index].value === open) depth += 1;
    if (tokens[index].value === close) {
      depth -= 1;
      if (depth === 0) return { inner: tokens.slice(start + 1, index), next: index + 1 };
    }
  }
  throw new Error(`unterminated ${open}${close} while parsing GraphQL schema`);
}

function parseTypeReference(tokens, start) {
  let index = start;
  const output = [];
  if (tokens[index]?.value === "[") {
    output.push(tokens[index]);
    const nested = parseTypeReference(tokens, index + 1);
    output.push(...nested.tokens);
    index = nested.next;
    if (tokens[index]?.value !== "]") throw new Error("GraphQL type reference is missing ]");
    output.push(tokens[index]);
    index += 1;
  } else if (tokens[index]?.kind === "name") {
    output.push(tokens[index]);
    index += 1;
  } else {
    throw new Error("GraphQL field is missing its return type");
  }
  if (tokens[index]?.value === "!") {
    output.push(tokens[index]);
    index += 1;
  }
  return { tokens: output, next: index };
}

function formatTokens(tokens) {
  let output = "";
  for (const token of tokens) {
    const value = token.value;
    if (token.kind === "name") {
      if (output !== "" && !/[\s(\[$]$/u.test(output)) output += " ";
      output += value;
      continue;
    }
    switch (value) {
      case "(":
      case "[":
      case "$":
        output += value;
        break;
      case ")":
      case "]":
      case "!":
        output = output.trimEnd() + value;
        break;
      case ":":
        output = output.trimEnd() + ": ";
        break;
      case "=":
      case "|":
        output = output.trimEnd() + ` ${value} `;
        break;
      default:
        output += value;
        break;
    }
  }
  return output.trim();
}

function findRootType(tokens, root) {
  for (let index = 0; index + 1 < tokens.length; index += 1) {
    if (tokens[index].value !== "type" || tokens[index + 1].value !== root) continue;
    for (let cursor = index + 2; cursor < tokens.length; cursor += 1) {
      if (tokens[cursor].value !== "{") continue;
      const body = findBalanced(tokens, cursor, "{", "}");
      return { start: cursor, end: body.next - 1 };
    }
  }
  throw new Error(`GraphQL schema is missing root type ${root}`);
}

function parseRootFields(tokens, root) {
  const body = findRootType(tokens, root);
  const fields = [];
  let index = body.start + 1;
  while (index < body.end) {
    const field = tokens[index];
    if (field.kind !== "name") {
      index += 1;
      continue;
    }
    const name = field.value;
    index += 1;
    let argumentsTokens = [];
    if (tokens[index]?.value === "(") {
      const argumentsBlock = findBalanced(tokens, index, "(", ")");
      argumentsTokens = argumentsBlock.inner;
      index = argumentsBlock.next;
    }
    if (tokens[index]?.value !== ":") throw new Error(`GraphQL ${root}.${name} is missing : before its return type`);
    index += 1;
    const returnType = parseTypeReference(tokens, index);
    index = returnType.next;
    const directives = [];
    while (tokens[index]?.value === "@") {
      const directive = tokens[index + 1];
      if (directive?.kind !== "name") throw new Error(`GraphQL ${root}.${name} has an unnamed directive`);
      directives.push(directive.value);
      index += 2;
      if (tokens[index]?.value === "(") index = findBalanced(tokens, index, "(", ")").next;
    }
    const argumentsText = formatTokens(argumentsTokens);
    const returnText = formatTokens(returnType.tokens);
    fields.push({
      root,
      name,
      line: field.line,
      signature: `${name}${argumentsTokens.length > 0 ? `(${argumentsText})` : ""}: ${returnText}`,
      deprecated: directives.includes("deprecated"),
      preview: directives.includes("preview"),
    });
  }
  return fields;
}

/** Parse just the schema root fields, treating nested types as projections rather than operations. */
export function parseGraphQLRootOperations(graphqlSchema) {
  const tokens = tokenizeGraphQL(graphqlSchema);
  const fields = [...parseRootFields(tokens, "Mutation"), ...parseRootFields(tokens, "Query")];
  const seen = new Set();
  for (const field of fields) {
    const key = `${field.root}.${field.name}`;
    if (seen.has(key)) throw new Error(`GraphQL schema has duplicate root field ${key}`);
    seen.add(key);
  }
  return fields;
}

function countLock(restOperations, rootFields) {
  const queries = rootFields.filter((field) => field.root === "Query").length;
  const mutations = rootFields.filter((field) => field.root === "Mutation").length;
  return { rest: restOperations.length, graphql_query: queries, graphql_mutation: mutations, total: restOperations.length + queries + mutations };
}

/** Build a compact, hermetic source lock from explicit official-source content. */
export function buildSourceLock({ restDocument, restText, graphqlSchema, capturedAt, restSource, graphqlSource }) {
  const document = requireObject(restDocument, "REST OpenAPI document");
  const rest = enumerateRESTOperations(document);
  const graphql = parseGraphQLRootOperations(graphqlSchema);
  if (!graphql.some((field) => field.root === "Mutation" && field.name === "createEnterpriseOrganization")) {
    throw new Error("GitHub GraphQL schema source is missing mandatory createEnterpriseOrganization mutation canary");
  }
  const restMetadata = requireObject(restSource, "REST source metadata");
  const graphqlMetadata = requireObject(graphqlSource, "GraphQL source metadata");
  const restRaw = typeof restText === "string" ? restText : JSON.stringify(document);
  if (typeof graphqlSchema !== "string" || graphqlSchema.trim() === "") {
    throw new Error("GraphQL schema must be a non-empty string");
  }
  // Preserve the artifact byte-for-byte for provenance. Parsing may trim
  // surrounding whitespace, but a source hash must include the final newline.
  const graphqlRaw = graphqlSchema;
  return {
    schema_version: 1,
    connector: "github",
    captured_at: requireDate(capturedAt, "source capture date"),
    rest: {
      source_url: requireString(restMetadata.url, "REST source URL"),
      commit: requireString(restMetadata.commit, "REST source commit"),
      sha256: sha256(restRaw),
      bytes: byteLength(restRaw),
      openapi: requireString(document.openapi, "REST OpenAPI version"),
      info_version: requireString(document.info?.version, "REST OpenAPI info.version"),
      operations: rest,
    },
    graphql: {
      source_url: requireString(graphqlMetadata.url, "GraphQL source URL"),
      sha256: sha256(graphqlRaw),
      bytes: byteLength(graphqlRaw),
      query_fields: graphql.filter((field) => field.root === "Query"),
      mutation_fields: graphql.filter((field) => field.root === "Mutation"),
    },
    counts: countLock(rest, graphql),
  };
}

function validateSourceLock(lockCandidate) {
  const lock = requireObject(lockCandidate, "GitHub source lock");
  if (lock.schema_version !== 1 || lock.connector !== "github") throw new Error("GitHub source lock has unsupported schema or connector");
  requireDate(lock.captured_at, "GitHub source lock captured_at");
  const rest = requireObject(lock.rest, "GitHub source lock rest");
  const graphql = requireObject(lock.graphql, "GitHub source lock graphql");
  for (const [part, value] of [["rest source URL", rest.source_url], ["rest commit", rest.commit], ["rest sha256", rest.sha256], ["graphql source URL", graphql.source_url], ["graphql sha256", graphql.sha256]]) {
    requireString(value, part);
  }
  if (!Array.isArray(rest.operations) || !Array.isArray(graphql.query_fields) || !Array.isArray(graphql.mutation_fields)) {
    throw new Error("GitHub source lock is missing source operation arrays");
  }
  const rootFields = [...graphql.query_fields, ...graphql.mutation_fields];
  if (!rootFields.some((field) => field.root === "Mutation" && field.name === "createEnterpriseOrganization")) {
    throw new Error("GitHub source lock is missing createEnterpriseOrganization mutation canary");
  }
  const counts = countLock(rest.operations, rootFields);
  if (JSON.stringify(counts) !== JSON.stringify(lock.counts)) {
    throw new Error(`GitHub source lock counts ${JSON.stringify(lock.counts)} do not match derived ${JSON.stringify(counts)}`);
  }
  return lock;
}

function graphQLRootSelection(document) {
  if (typeof document !== "string" || document.trim() === "") return "";
  const tokens = tokenizeGraphQL(document);
  const opening = tokens.findIndex((token) => token.value === "{");
  if (opening < 0) return "";
  let candidate = tokens[opening + 1];
  if (candidate?.kind !== "name") return "";
  if (tokens[opening + 2]?.value === ":") candidate = tokens[opening + 3];
  return candidate?.kind === "name" ? candidate.value : "";
}

function buildBundleIndexes(bundleCandidate) {
  const bundle = requireObject(bundleCandidate, "GitHub bundle");
  const surface = requireObject(bundle.surface, "GitHub bundle surface");
  const operations = requireObject(bundle.operations, "GitHub bundle operations");
  const cli = requireObject(bundle.cli, "GitHub bundle cli");
  if (!Array.isArray(surface.endpoints) || !Array.isArray(operations.operations) || !Array.isArray(cli.commands)) {
    throw new Error("GitHub bundle is missing endpoint, operation, or command arrays");
  }
  const surfaceByKey = new Map(surface.endpoints.map((endpoint) => [endpointKey(endpoint.method, endpoint.path), endpoint]));
  const commandsByOperation = new Map();
  const commandsBySurface = new Map();
  const commandsByStream = new Map();
  const commandsByWrite = new Map();
  for (const command of cli.commands) {
    if (command.operation) commandsByOperation.set(command.operation, [...(commandsByOperation.get(command.operation) || []), command]);
    if (command.stream) commandsByStream.set(command.stream, [...(commandsByStream.get(command.stream) || []), command]);
    if (command.write) commandsByWrite.set(command.write, [...(commandsByWrite.get(command.write) || []), command]);
    for (const reference of command.api_surface || []) {
      const key = endpointKey(reference.method, reference.path);
      commandsBySurface.set(key, [...(commandsBySurface.get(key) || []), command]);
    }
  }
  const restOperationCommands = new Map();
  const graphQLBindings = new Map();
  for (const operation of operations.operations) {
    if (operation.rest?.method && operation.rest?.path) {
      const key = endpointKey(operation.rest.method, operation.rest.path);
      restOperationCommands.set(key, [...(restOperationCommands.get(key) || []), ...(commandsByOperation.get(operation.id) || [])]);
    }
    if (!String(operation.kind || "").startsWith("graphql")) continue;
    const root = graphQLRootSelection(operation.graphql?.document);
    if (root === "") continue;
    const protocol = operation.kind === "graphql_mutation" ? "Mutation" : "Query";
    const key = `${protocol}.${root}`;
    const bindings = graphQLBindings.get(key) || { operations: [], commands: [], documents: [] };
    bindings.operations.push(operation.id);
    bindings.documents.push(operation.graphql?.document || "");
    bindings.commands.push(...(commandsByOperation.get(operation.id) || []));
    const operationName = operation.graphql?.operation_name;
    for (const command of cli.commands) {
      if ((command.api_surface || []).some((reference) => reference.method === "GRAPHQL" && reference.path === operationName)) {
        bindings.commands.push(command);
      }
    }
    graphQLBindings.set(key, bindings);
  }
  return { bundle, surfaceByKey, commandsBySurface, commandsByStream, commandsByWrite, restOperationCommands, graphQLBindings };
}

function coverageTargets(endpoint) {
  const coverage = endpoint?.covered_by;
  if (!isPlainObject(coverage)) return { streams: [], writes: [], directReads: [] };
  return {
    streams: coverage.stream ? [coverage.stream] : [],
    writes: uniqueSorted([coverage.write, ...(coverage.writes || [])]),
    directReads: uniqueSorted([coverage.direct_read, ...(coverage.direct_reads || [])]),
  };
}

function commandPaths(commands) {
  return uniqueSorted(commands.map((command) => command.path));
}

function restBindings(operation, indexes) {
  const key = endpointKey(operation.method, operation.path);
  const endpoint = indexes.surfaceByKey.get(key);
  const commands = [
    ...(indexes.commandsBySurface.get(key) || []),
    ...(indexes.restOperationCommands.get(key) || []),
  ];
  const targets = coverageTargets(endpoint);
  for (const stream of targets.streams) commands.push(...(indexes.commandsByStream.get(stream) || []));
  for (const write of targets.writes) commands.push(...(indexes.commandsByWrite.get(write) || []));
  for (const directRead of targets.directReads) {
    const matched = indexes.bundle.cli.commands.find((command) => command.path === directRead);
    if (matched) commands.push(matched);
  }
  return { endpoint, commands: [...new Map(commands.map((command) => [command.path, command])).values()] };
}

function implementationState({ protocol, commands, endpoint }) {
  const availabilities = commands.map((command) => command.availability).filter(Boolean);
  if (availabilities.length > 0) {
    if (protocol.startsWith("graphql")) {
      if (availabilities.some((availability) => availability === "implemented" || availability === "partial")) return "partially_implemented";
      return "declared_not_executable";
    }
    if (availabilities.every((availability) => availability === "implemented")) return "implemented";
    if (availabilities.some((availability) => availability === "implemented" || availability === "partial")) return "partially_implemented";
    return "declared_not_executable";
  }
  if (endpoint?.covered_by) return "implemented_generic_only";
  if (endpoint?.operation?.reason) return "blocked";
  return "not_implemented";
}

function lifecycleFields(protocol) {
  const isRead = protocol === "rest" ? undefined : protocol === "graphql_query";
  return {
    auth: {
      persona: "requires_operation_contract",
      scopes: "requires_operation_contract",
      entitlement: "requires_operation_contract",
    },
    fixture: { dependency: "requires_operation_contract" },
    safety: {
      mutability: isRead === true ? "read" : isRead === false ? "write" : "derived_from_rest_method",
      approval_class: isRead === true ? "not_applicable" : "requires_operation_contract",
    },
    assertion: { strategy: isRead === true ? "independent response assertion required" : "independent read-back required" },
    cleanup: { strategy: isRead === true ? "not_applicable" : "requires_operation_contract" },
    terminal_evidence: { state: "not_attempted" },
  };
}

function blockerFor({ state, endpoint, protocol, commands = [] }) {
  if (["implemented", "partially_implemented", "implemented_generic_only"].includes(state)) return undefined;
  if (state === "blocked" && endpoint?.operation?.reason) {
    return {
      category: "existing_bundle_block",
      reason: endpoint.operation.reason,
      unblocking_condition: "Implement the named dependency recorded by the existing GitHub bundle contract.",
    };
  }
  const operationLabel = protocol.startsWith("graphql") ? "typed GraphQL operation contract" : "typed REST operation contract";
  if (state === "declared_not_executable" && commands.length > 0) {
    const commandDetails = uniqueSorted(commands.map((command) => `${command.path} (${command.availability || "unspecified"})`));
    return {
      category: "mapped_command_not_executable",
      reason: `PM maps this source operation to ${commandDetails.join(", ")}, but the declared command availability is not executable by the ${operationLabel} runtime.`,
      unblocking_condition: `Add an executable ${operationLabel} for the mapped PM command with its declared safety policy, typed inputs, auth/scope classification, fixture/read-back, and cleanup contract.`,
    };
  }
  return {
    category: state,
    reason: `No ${operationLabel} is currently mapped to this source operation.`,
    unblocking_condition: `Add a ${operationLabel} with a declared PM command, auth/scope classification, safety class, fixture/read-back, and cleanup contract.`,
  };
}

function graphQLRow(field, lock, indexes) {
  const protocol = field.root === "Mutation" ? "graphql_mutation" : "graphql_query";
  const key = `${field.root}.${field.name}`;
  const bindings = indexes.graphQLBindings.get(key) || { operations: [], commands: [], documents: [] };
  const commands = [...new Map(bindings.commands.map((command) => [command.path, command])).values()];
  const state = implementationState({ protocol, commands });
  const row = {
    id: `github.graphql.${field.root.toLowerCase()}.${field.name}`,
    protocol,
    source: {
      artifact: "GitHub Docs public GraphQL schema",
      source_url: lock.graphql.source_url,
      sha256: lock.graphql.sha256,
      location: `type ${field.root}:${field.line}`,
      signature: field.signature,
      deprecated: field.deprecated,
      preview: field.preview,
    },
    pm: {
      commands: commandPaths(commands),
      operation_ids: uniqueSorted(bindings.operations),
      mapping: commands.length > 0 ? "fixed-document binding" : "no typed PM command",
    },
    implementation: { state },
    ...lifecycleFields(protocol),
  };
  const blocker = blockerFor({ state, protocol, commands });
  if (blocker) row.blocker = blocker;
  if (field.name === "node" || field.name === "nodes") {
    const supportedObjectTypes = uniqueSorted(bindings.documents.flatMap((document) =>
      [...String(document || "").matchAll(/\.\.\.\s+on\s+([_A-Za-z][_0-9A-Za-z]*)/gu)].map((match) => match[1]),
    ));
    row.projection_matrix = {
      root_field: field.name,
      policy: "fixed declared documents only; no caller-supplied GraphQL selection",
      supported_object_types: supportedObjectTypes,
      state: supportedObjectTypes.length > 0 ? "fixed_projection_only" : "no_supported_projection",
    };
  }
  return row;
}

function restRow(operation, lock, indexes) {
  const bindings = restBindings(operation, indexes);
  const state = implementationState({ protocol: "rest", commands: bindings.commands, endpoint: bindings.endpoint });
  const isRead = operation.method === "GET";
  const row = {
    id: operation.id,
    protocol: "rest",
    source: {
      artifact: "GitHub REST OpenAPI description",
      source_url: lock.rest.source_url,
      commit: lock.rest.commit,
      sha256: lock.rest.sha256,
      location: operation.source_location,
      method: operation.method,
      path: operation.path,
      operation_id: operation.operation_id,
      deprecated: operation.deprecated,
    },
    pm: {
      commands: commandPaths(bindings.commands),
      mapping: bindings.commands.length > 0 ? "declared PM command binding" : bindings.endpoint?.covered_by ? "generic connector binding" : "no typed PM command",
    },
    implementation: { state },
    ...lifecycleFields(isRead ? "graphql_query" : "graphql_mutation"),
  };
  row.safety.mutability = isRead ? "read" : "write";
  const blocker = blockerFor({ state, endpoint: bindings.endpoint, protocol: "rest", commands: bindings.commands });
  if (blocker) row.blocker = blocker;
  return row;
}

/** Generate a protocol-separated operation ledger from a source lock and current bundle bindings. */
export function buildCombinedOperationLedger({ lock: lockCandidate, bundle: bundleCandidate }) {
  const lock = validateSourceLock(lockCandidate);
  const indexes = buildBundleIndexes(bundleCandidate);
  const rootFields = [...lock.graphql.mutation_fields, ...lock.graphql.query_fields];
  const rows = [
    ...lock.rest.operations.map((operation) => restRow(operation, lock, indexes)),
    ...rootFields.map((field) => graphQLRow(field, lock, indexes)),
  ].sort((left, right) => left.id.localeCompare(right.id));
  return {
    schema_version: 1,
    connector: "github",
    generated_from: {
      source_lock_sha256: sha256(JSON.stringify(lock)),
      captured_at: lock.captured_at,
    },
    sources: {
      rest: { source_url: lock.rest.source_url, commit: lock.rest.commit, sha256: lock.rest.sha256, bytes: lock.rest.bytes },
      graphql: { source_url: lock.graphql.source_url, sha256: lock.graphql.sha256, bytes: lock.graphql.bytes },
    },
    counts: lock.counts,
    rows,
  };
}

/** Fail closed for source omissions, duplicate IDs, stale source hashes, or UNTESTABLE wording. */
export function validateCombinedOperationLedger({ lock: lockCandidate, ledger: ledgerCandidate }) {
  const lock = validateSourceLock(lockCandidate);
  const ledger = requireObject(ledgerCandidate, "combined GitHub operation ledger");
  if (ledger.schema_version !== 1 || ledger.connector !== "github") throw new Error("combined ledger has unsupported schema or connector");
  if (!Array.isArray(ledger.rows)) throw new Error("combined ledger rows must be an array");
  if (JSON.stringify(ledger.counts) !== JSON.stringify(lock.counts)) throw new Error("combined ledger counts do not match source lock");
  const expectedIDs = new Set([
    ...lock.rest.operations.map((operation) => operation.id),
    ...lock.graphql.query_fields.map((field) => `github.graphql.query.${field.name}`),
    ...lock.graphql.mutation_fields.map((field) => `github.graphql.mutation.${field.name}`),
  ]);
  const actualIDs = new Set();
  const protocolCounts = { rest: 0, graphql_query: 0, graphql_mutation: 0 };
  for (const row of ledger.rows) {
    requireObject(row, "combined ledger row");
    for (const field of REQUIRED_ROW_FIELDS) {
      if (!Object.hasOwn(row, field)) throw new Error(`combined ledger row ${row.id || "<unknown>"} is missing ${field}`);
    }
    const id = requireString(row.id, "combined ledger row.id");
    if (actualIDs.has(id)) throw new Error(`combined ledger has duplicate stable operation ID ${id}`);
    actualIDs.add(id);
    if (!Object.hasOwn(protocolCounts, row.protocol)) throw new Error(`combined ledger row ${id} has invalid protocol ${row.protocol}`);
    protocolCounts[row.protocol] += 1;
    if (JSON.stringify(row).toUpperCase().includes("UNTESTABLE")) throw new Error(`combined ledger row ${id} uses forbidden UNTESTABLE classification`);
    const sourceHash = row.protocol === "rest" ? lock.rest.sha256 : lock.graphql.sha256;
    if (row.source?.sha256 !== sourceHash) throw new Error(`combined ledger row ${id} does not retain its source hash`);
    if (row.implementation?.state === "not_implemented" && !row.blocker?.unblocking_condition) {
      throw new Error(`combined ledger row ${id} is not implemented without an explicit unblocking condition`);
    }
  }
  if (actualIDs.size !== expectedIDs.size || [...expectedIDs].some((id) => !actualIDs.has(id))) {
    throw new Error("combined ledger does not contain every source operation exactly once");
  }
  if (protocolCounts.rest !== lock.counts.rest || protocolCounts.graphql_query !== lock.counts.graphql_query || protocolCounts.graphql_mutation !== lock.counts.graphql_mutation) {
    throw new Error(`combined ledger protocol counts ${JSON.stringify(protocolCounts)} do not match source lock`);
  }
  const canary = ledger.rows.find((row) => row.id === "github.graphql.mutation.createEnterpriseOrganization");
  if (!canary || !canary.implementation?.state) throw new Error("combined ledger is missing classified createEnterpriseOrganization canary");
  return true;
}

async function loadBundle() {
  const files = await Promise.all(["api_surface.json", "operations.json", "cli_surface.json"].map((name) => readFile(path.join(GITHUB_DEFS, name), "utf8")));
  return {
    surface: JSON.parse(files[0]),
    operations: JSON.parse(files[1]),
    cli: JSON.parse(files[2]),
  };
}

function parseArgs(args) {
  const options = {};
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    if (argument === "--write" || argument === "--check") {
      options.mode = argument.slice(2);
      continue;
    }
    if (argument === "--help") {
      options.help = true;
      continue;
    }
    if (!argument.startsWith("--")) throw new Error(`unknown argument ${argument}`);
    const key = argument.slice(2);
    const value = args[index + 1];
    if (!value || value.startsWith("--")) throw new Error(`${argument} requires a value`);
    options[key] = value;
    index += 1;
  }
  return options;
}

function usage() {
  return [
    "usage:",
    "  github-combined-operation-ledger --write --rest <openapi.json> --graphql <schema.graphql> [--lock <path>] [--ledger <path>] [--captured-at YYYY-MM-DD]",
    "  github-combined-operation-ledger --check [--lock <path>] [--ledger <path>]",
  ].join("\n");
}

async function writeJSON(file, value) {
  await mkdir(path.dirname(file), { recursive: true, mode: 0o700 });
  await writeFile(file, `${JSON.stringify(value, null, 2)}\n`, "utf8");
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  if (options.help) {
    process.stdout.write(`${usage()}\n`);
    return;
  }
  const lockPath = path.resolve(options.lock || DEFAULT_LOCK);
  const ledgerPath = path.resolve(options.ledger || DEFAULT_LEDGER);
  const bundle = await loadBundle();
  if (options.mode === "write") {
    if (!options.rest || !options.graphql) throw new Error("--write requires --rest and --graphql source artifact paths");
    const [restText, graphqlSchema] = await Promise.all([readFile(path.resolve(options.rest), "utf8"), readFile(path.resolve(options.graphql), "utf8")]);
    const lock = buildSourceLock({
      restDocument: JSON.parse(restText),
      restText,
      graphqlSchema,
      capturedAt: options["captured-at"] || new Date().toISOString().slice(0, 10),
      restSource: { url: options["rest-url"] || REST_URL, commit: options["rest-commit"] || REST_COMMIT },
      graphqlSource: { url: options["graphql-url"] || GRAPHQL_URL },
    });
    const ledger = buildCombinedOperationLedger({ lock, bundle });
    validateCombinedOperationLedger({ lock, ledger });
    await Promise.all([writeJSON(lockPath, lock), writeJSON(ledgerPath, ledger)]);
    process.stdout.write(`github combined operation ledger: rest=${lock.counts.rest} graphql_query=${lock.counts.graphql_query} graphql_mutation=${lock.counts.graphql_mutation} total=${lock.counts.total}\n`);
    return;
  }
  if (options.mode === "check") {
    const [lockText, ledgerText] = await Promise.all([readFile(lockPath, "utf8"), readFile(ledgerPath, "utf8")]);
    const lock = JSON.parse(lockText);
    const expected = buildCombinedOperationLedger({ lock, bundle });
    const ledger = JSON.parse(ledgerText);
    validateCombinedOperationLedger({ lock, ledger });
    if (JSON.stringify(expected) !== JSON.stringify(ledger)) throw new Error("combined operation ledger drift; regenerate from the checked-in source lock");
    process.stdout.write(`github combined operation ledger: ok rest=${lock.counts.rest} graphql_query=${lock.counts.graphql_query} graphql_mutation=${lock.counts.graphql_mutation} total=${lock.counts.total}\n`);
    return;
  }
  throw new Error(usage());
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    process.stderr.write(`github-combined-operation-ledger: ${error.message}\n`);
    process.exitCode = 1;
  });
}
