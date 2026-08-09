#!/usr/bin/env node

// Scheduled-only GitHub source drift detector. Hermetic CI validates the
// checked-in lock; this script intentionally reads the two public upstream
// artifacts and reports any byte or operation-inventory difference without
// changing the worktree or making a provider mutation.

import { readFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { buildSourceLock } from "./github-combined-operation-ledger.mjs";

const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(SCRIPT_DIR, "..");
const DEFAULT_LOCK = path.join(ROOT, "internal", "connectors", "defs", "github", "sources", "github-operation-source-lock.json");
export const DEFAULT_REST_URL = "https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.json";
export const DEFAULT_GRAPHQL_URL = "https://docs.github.com/public/ghec/schema.docs.graphql";
const MAX_SOURCE_BYTES = 20 * 1024 * 1024;

function requireNonEmptyString(value, label) {
  if (typeof value !== "string" || value.trim() === "") throw new Error(`${label} must be a non-empty string`);
  return value.trim();
}

function requireCounts(counts, label) {
  if (!counts || typeof counts !== "object") throw new Error(`${label} counts are missing`);
  for (const key of ["rest", "graphql_query", "graphql_mutation", "total"]) {
    if (!Number.isInteger(counts[key]) || counts[key] < 0) throw new Error(`${label} counts.${key} must be a non-negative integer`);
  }
  if (counts.rest + counts.graphql_query + counts.graphql_mutation !== counts.total) {
    throw new Error(`${label} counts do not sum to total`);
  }
  return counts;
}

function rootOperationIDs(lock) {
  return {
    rest: lock.rest.operations.map((operation) => operation.id).sort((left, right) => left.localeCompare(right)),
    graphql: [
      ...lock.graphql.query_fields.map((field) => `github.graphql.query.${field.name}`),
      ...lock.graphql.mutation_fields.map((field) => `github.graphql.mutation.${field.name}`),
    ].sort((left, right) => left.localeCompare(right)),
  };
}

function operationDifference(pinned, observed) {
  const pinnedSet = new Set(pinned);
  const observedSet = new Set(observed);
  return {
    added_operation_ids: observed.filter((id) => !pinnedSet.has(id)),
    removed_operation_ids: pinned.filter((id) => !observedSet.has(id)),
  };
}

function sourceDifference({ pinned, observed, pinnedIDs, observedIDs }) {
  const operationDiff = operationDifference(pinnedIDs, observedIDs);
  return {
    pinned_sha256: pinned.sha256,
    observed_sha256: observed.sha256,
    pinned_bytes: pinned.bytes,
    observed_bytes: observed.bytes,
    content_changed: pinned.sha256 !== observed.sha256,
    ...operationDiff,
  };
}

function hasSourceDifference(difference) {
  return difference.content_changed || difference.added_operation_ids.length > 0 || difference.removed_operation_ids.length > 0;
}

function differenceSummary(label, difference) {
  const unchanged = !hasSourceDifference(difference);
  if (unchanged) return `${label}: unchanged`;
  return `${label}: ${difference.added_operation_ids.length} added, ${difference.removed_operation_ids.length} removed, content ${difference.content_changed ? "changed" : "unchanged"}`;
}

/**
 * Compares public upstream source content to the checked-in source lock. It
 * deliberately exposes only hashes, counts, and public operation IDs—never
 * fetched source content, credentials, or provider response bodies.
 */
export function compareGitHubSourceDrift({ lock, restText, graphqlSchema, restSource, graphqlSource, capturedAt }) {
  if (!lock || typeof lock !== "object") throw new Error("checked-in GitHub source lock is missing");
  requireCounts(lock.counts, "checked-in GitHub source lock");
  if (!lock.rest || !lock.graphql || !Array.isArray(lock.rest.operations) || !Array.isArray(lock.graphql.query_fields) || !Array.isArray(lock.graphql.mutation_fields)) {
    throw new Error("checked-in GitHub source lock has no complete REST and GraphQL inventories");
  }
  requireNonEmptyString(restText, "upstream REST source");
  requireNonEmptyString(graphqlSchema, "upstream GraphQL source");
  const observed = buildSourceLock({
    restDocument: JSON.parse(restText),
    restText,
    graphqlSchema,
    capturedAt: requireNonEmptyString(capturedAt, "source drift capture date"),
    restSource: {
      url: requireNonEmptyString(restSource?.url, "upstream REST source URL"),
      commit: requireNonEmptyString(restSource?.commit, "upstream REST source revision"),
    },
    graphqlSource: { url: requireNonEmptyString(graphqlSource?.url, "upstream GraphQL source URL") },
  });
  const pinnedIDs = rootOperationIDs(lock);
  const observedIDs = rootOperationIDs(observed);
  const rest = sourceDifference({ pinned: lock.rest, observed: observed.rest, pinnedIDs: pinnedIDs.rest, observedIDs: observedIDs.rest });
  const graphql = sourceDifference({ pinned: lock.graphql, observed: observed.graphql, pinnedIDs: pinnedIDs.graphql, observedIDs: observedIDs.graphql });
  const hasDrift = hasSourceDifference(rest) || hasSourceDifference(graphql);
  return {
    schema_version: 1,
    connector: "github",
    has_drift: hasDrift,
    pinned: {
      captured_at: lock.captured_at,
      counts: lock.counts,
    },
    observed: {
      captured_at: observed.captured_at,
      counts: observed.counts,
    },
    rest,
    graphql,
    summary: `${differenceSummary("REST", rest)}; ${differenceSummary("GraphQL", graphql)}`,
  };
}

export async function fetchPublicSourceText(url, { fetchFn = fetch, maxBytes = MAX_SOURCE_BYTES } = {}) {
  const response = await fetchFn(url, {
    headers: {
      accept: "application/json, application/graphql, text/plain;q=0.9",
      "user-agent": "polymetrics-github-source-drift/1",
    },
    signal: AbortSignal.timeout(30_000),
  });
  if (!response?.ok) throw new Error(`public source fetch failed for ${url}: HTTP ${response?.status ?? "unknown"}`);
  const advertisedLength = Number(response.headers?.get?.("content-length"));
  if (Number.isFinite(advertisedLength) && advertisedLength > maxBytes) {
    throw new Error(`public source fetch exceeds ${maxBytes} byte limit for ${url}`);
  }
  const text = await response.text();
  if (Buffer.byteLength(text, "utf8") > maxBytes) throw new Error(`public source fetch exceeds ${maxBytes} byte limit for ${url}`);
  return text;
}

function parseArgs(argumentsList) {
  const options = {};
  for (let index = 0; index < argumentsList.length; index += 1) {
    const argument = argumentsList[index];
    if (argument === "--help") {
      options.help = true;
      continue;
    }
    if (!argument.startsWith("--")) throw new Error(`unknown argument ${argument}`);
    const key = argument.slice(2);
    const value = argumentsList[index + 1];
    if (!value || value.startsWith("--")) throw new Error(`${argument} requires a value`);
    options[key] = value;
    index += 1;
  }
  return options;
}

function usage() {
  return [
    "usage:",
    "  github-source-drift [--lock <source-lock.json>] [--rest-url <official-openapi-url>] [--graphql-url <official-schema-url>]",
    "",
    "Fetches public official GitHub schema artifacts, compares only hashes/counts/operation IDs to the checked-in lock, and exits 1 when source drift is detected.",
  ].join("\n");
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  if (options.help) {
    process.stdout.write(`${usage()}\n`);
    return;
  }
  const lockPath = path.resolve(options.lock || DEFAULT_LOCK);
  const restURL = options["rest-url"] || DEFAULT_REST_URL;
  const graphqlURL = options["graphql-url"] || DEFAULT_GRAPHQL_URL;
  const [lockText, restText, graphqlSchema] = await Promise.all([
    readFile(lockPath, "utf8"),
    fetchPublicSourceText(restURL),
    fetchPublicSourceText(graphqlURL),
  ]);
  const report = compareGitHubSourceDrift({
    lock: JSON.parse(lockText),
    restText,
    graphqlSchema,
    restSource: { url: restURL, commit: "upstream-default-branch" },
    graphqlSource: { url: graphqlURL },
    capturedAt: new Date().toISOString().slice(0, 10),
  });
  process.stdout.write(`${JSON.stringify(report, null, 2)}\n`);
  if (report.has_drift) process.exitCode = 1;
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    process.stderr.write(`github source drift: ${error instanceof Error ? error.message : String(error)}\n`);
    process.exitCode = 2;
  });
}
