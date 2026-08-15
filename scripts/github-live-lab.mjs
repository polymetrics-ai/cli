#!/usr/bin/env node

import { appendFile, mkdir, readFile } from "node:fs/promises";
import { spawn } from "node:child_process";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import {
  assertPersistedArtifactSafe,
  assertSafePersistedScalar,
  redactPersistedText,
} from "./github-live-artifact-guard.mjs";

const SCHEMA_VERSION = 1;
const RESOURCE_TYPES = new Set(["repository", "organization"]);
const TERMINAL_STATES = new Set([
  "proven",
  "failed",
  "credential_blocker",
  "entitlement_blocker",
]);
const CLEANUP_ACTIONS = new Set([
  "initialized",
  "created",
  "read_back",
  "neutralized",
  "cleanup_failed",
  "cleanup_completed",
  "cleanup_already_absent",
  "retained",
]);
const LEDGER_FIELDS = new Set([
  "schema_version",
  "run_id",
  "fixture_id",
  "action",
  "target",
  "pm_command",
  "provider_id",
  "residual_state",
  "retention_reason",
  "note",
]);
const SENSITIVE_KEY = /(?:approval|authorization|credential|grant|password|private[_-]?key|secret|token|value)/i;
const SAFE_BOUNDARY_IDENTIFIER = /^[A-Za-z0-9][A-Za-z0-9_.:-]{0,255}$/u;
const GITHUB_SLUG = /^[A-Za-z0-9](?:[A-Za-z0-9-]{0,98})$/u;
const BOUNDARY_FIELDS = new Set([
  "schema_version",
  "run_id",
  "default_deny",
  "protected_owners",
  "protected_repositories",
  "working_repositories",
  "allowed_targets",
  "historical_runs",
  "bootstrap_principals",
]);
const OUTPUT_LIMIT_BYTES = 2 * 1024 * 1024;
const ISSUE_READBACK_MAX_ATTEMPTS = 6;
const ISSUE_READBACK_RETRY_DELAY_MS = 1000;
const TARGETLESS_ACCOUNT_PROBE_ERROR = "targetless account bootstrap probes are untestable; only the App installation repository preflight may use an installation credential";

function isPlainObject(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function requirePlainObject(value, label) {
  if (!isPlainObject(value)) throw new Error(`${label} must be an object`);
  return value;
}

function requireString(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  return value.trim();
}

function requireSafeString(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  assertSafePersistedScalar(value, label);
  return value.trim();
}

function requireBoundaryIdentifier(value, label) {
  const text = requireSafeString(value, label);
  if (!SAFE_BOUNDARY_IDENTIFIER.test(text)) {
    throw new Error(`${label} must be a well-formed immutable identifier`);
  }
  return text;
}

function requireKnownFields(value, fields, label) {
  const object = requirePlainObject(value, label);
  if (Object.keys(object).some((key) => !fields.has(key))) {
    throw new Error(`${label} contains an unsupported field`);
  }
  return object;
}

function requireProviderID(value, label) {
  if (typeof value === "string") return requireBoundaryIdentifier(value, label);
  if (typeof value === "number" && Number.isSafeInteger(value) && value > 0) return String(value);
  throw new Error(`${label} must be a non-empty immutable provider ID`);
}

function requireStringArray(value, label) {
  if (!Array.isArray(value) || value.some((item) => typeof item !== "string" || item.trim() === "")) {
    throw new Error(`${label} must be an array of non-empty strings`);
  }
  const normalized = value.map((item) => requireSafeString(item, label));
  if (new Set(normalized.map((item) => item.toLowerCase())).size !== normalized.length) {
    throw new Error(`${label} must not contain duplicate values`);
  }
  return normalized;
}

function normalizedSlug(value) {
  const slug = requireSafeString(value, "target slug");
  if (!GITHUB_SLUG.test(slug)) {
    throw new Error("target slug must be a well-formed GitHub slug");
  }
  return slug.toLowerCase();
}

function ensureNoSensitiveData(value, label = "lab value") {
  assertPersistedArtifactSafe(value, label);
  if (typeof value === "string") {
    requireSafeString(value, label);
    return;
  }
  if (Array.isArray(value)) {
    value.forEach((item, index) => ensureNoSensitiveData(item, `${label}[${index}]`));
    return;
  }
  if (!isPlainObject(value)) return;
  for (const [key, nested] of Object.entries(value)) {
    if (SENSITIVE_KEY.test(key)) {
      throw new Error(`${label} contains a forbidden sensitive field`);
    }
    ensureNoSensitiveData(nested, `${label}.${key}`);
  }
}

function normalizeRepositoryReference(value, label) {
  const target = requireKnownFields(value, new Set(["owner_slug", "repo_slug"]), label);
  return {
    owner_slug: normalizedSlug(target.owner_slug),
    repo_slug: normalizedSlug(target.repo_slug),
  };
}

function normalizeLabTarget(value, label = "lab target", { allowKey = false } = {}) {
  ensureNoSensitiveData(value, label);
  const target = requirePlainObject(value, label);
  const resourceType = requireSafeString(target.resource_type, `${label}.resource_type`);
  if (!RESOURCE_TYPES.has(resourceType)) {
    throw new Error(`${label}.resource_type must be repository or organization`);
  }
  const fields = resourceType === "repository"
    ? new Set(["resource_type", "owner_slug", "owner_id", "repo_slug", "repo_id", "run_owned", ...(allowKey ? ["key"] : [])])
    : new Set(["resource_type", "org_slug", "org_id", "run_owned", ...(allowKey ? ["key"] : [])]);
  requireKnownFields(target, fields, label);
  if (target.run_owned !== undefined && target.run_owned !== true) {
    throw new Error(`${label}.run_owned must be true when present`);
  }
  if (resourceType === "repository") {
    return {
      resource_type: resourceType,
      owner_slug: normalizedSlug(target.owner_slug),
      owner_id: requireProviderID(target.owner_id, `${label}.owner_id (immutable ID)`),
      repo_slug: normalizedSlug(target.repo_slug),
      repo_id: requireProviderID(target.repo_id, `${label}.repo_id (immutable ID)`),
      ...(target.run_owned === true ? { run_owned: true } : {}),
    };
  }
  return {
    resource_type: resourceType,
    org_slug: normalizedSlug(target.org_slug),
    org_id: requireProviderID(target.org_id, `${label}.org_id (immutable ID)`),
    ...(target.run_owned === true ? { run_owned: true } : {}),
  };
}

/** Historical runs are audit-only and never enter the executable allowlist. */
function normalizeHistoricalRun(value, label) {
  const run = requireKnownFields(value, new Set(["run_id", "target"]), label);
  const target = normalizeLabTarget(run.target, `${label}.target`, { allowKey: true });
  if (target.run_owned !== true) {
    throw new Error(`${label}.target must be marked run_owned`);
  }
  return {
    run_id: requireBoundaryIdentifier(run.run_id, `${label}.run_id`),
    target,
  };
}

function targetIdentity(target) {
  if (target.resource_type === "repository") {
    return [target.resource_type, target.owner_slug, target.owner_id, target.repo_slug, target.repo_id].join(":");
  }
  return [target.resource_type, target.org_slug, target.org_id].join(":");
}

function sameSlug(target, allowed) {
  if (target.resource_type !== allowed.resource_type) return false;
  if (target.resource_type === "repository") {
    return target.owner_slug === allowed.owner_slug && target.repo_slug === allowed.repo_slug;
  }
  return target.org_slug === allowed.org_slug;
}

function targetLabel(target) {
  if (target.resource_type === "repository") return `${target.owner_slug}/${target.repo_slug}`;
  return target.org_slug;
}

function normalizeBootstrapPrincipal(value, label) {
  const principal = requireKnownFields(value, new Set([
    "key",
    "resource_type",
    "user_slug",
    "user_id",
    "allowed_command",
    "requested_repo_slug",
    "required_private",
    "required_auto_init",
    "purpose",
  ]), label);
  if (requireSafeString(principal.resource_type, `${label}.resource_type`) !== "authenticated_user") {
    throw new Error(`${label}.resource_type must be authenticated_user`);
  }
  if (requireSafeString(principal.allowed_command, `${label}.allowed_command`) !== "repo create") {
    throw new Error(`${label}.allowed_command must be repo create`);
  }
  const requestedRepo = normalizedSlug(principal.requested_repo_slug);
  if (!requestedRepo.startsWith("pm-live-lab-")) {
    throw new Error(`${label}.requested_repo_slug must use the pm-live-lab- prefix`);
  }
  if (principal.required_private !== true || principal.required_auto_init !== true) {
    throw new Error(`${label} must require a private auto-initialized repository`);
  }
  return {
    key: requireBoundaryIdentifier(principal.key, `${label}.key`),
    resource_type: "authenticated_user",
    user_slug: normalizedSlug(principal.user_slug),
    user_id: requireProviderID(principal.user_id, `${label}.user_id (immutable ID)`),
    allowed_command: "repo create",
    requested_repo_slug: requestedRepo,
    required_private: true,
    required_auto_init: true,
    purpose: requireSafeString(principal.purpose, `${label}.purpose`),
  };
}

function isProtectedTarget(boundary, target) {
  if (target.resource_type === "organization") {
    return boundary.protected_owners.includes(target.org_slug);
  }
  if (boundary.protected_owners.includes(target.owner_slug)) return true;
  const reference = { owner_slug: target.owner_slug, repo_slug: target.repo_slug };
  return [...boundary.protected_repositories, ...boundary.working_repositories].some(
    (protectedRepository) =>
      protectedRepository.owner_slug === reference.owner_slug &&
      protectedRepository.repo_slug === reference.repo_slug,
  );
}

/**
 * The boundary is deliberately independent of credential metadata. A credential
 * may be valid but still not be allowed to address a target. This validation
 * happens before a PM command or provider request is constructed.
 */
export function validateLabBoundary(candidate) {
  const boundary = requireKnownFields(candidate, BOUNDARY_FIELDS, "lab boundary");
  ensureNoSensitiveData(boundary, "lab boundary");
  if (boundary.schema_version !== SCHEMA_VERSION) {
    throw new Error(`lab boundary schema_version must be ${SCHEMA_VERSION}`);
  }
  const runID = requireBoundaryIdentifier(boundary.run_id, "lab boundary.run_id");
  if (boundary.default_deny !== true) {
    throw new Error("lab boundary must set default_deny=true");
  }
  const protectedOwners = requireStringArray(boundary.protected_owners, "lab boundary.protected_owners")
    .map((owner) => normalizedSlug(owner));
  if (!protectedOwners.includes("polymetrics-ai")) {
    throw new Error("lab boundary must explicitly deny the polymetrics-ai owner");
  }
  const protectedRepositories = (boundary.protected_repositories || [])
    .map((item, index) => normalizeRepositoryReference(item, `lab boundary.protected_repositories[${index}]`));
  const workingRepositories = (boundary.working_repositories || [])
    .map((item, index) => normalizeRepositoryReference(item, `lab boundary.working_repositories[${index}]`));
  if (workingRepositories.length === 0) {
    throw new Error("lab boundary must list every working repository it denies");
  }
  for (const repository of workingRepositories) {
    const protectedByOwner = protectedOwners.includes(repository.owner_slug);
    const protectedByRepository = protectedRepositories.some(
      (entry) => entry.owner_slug === repository.owner_slug && entry.repo_slug === repository.repo_slug,
    );
    if (!protectedByOwner && !protectedByRepository) {
      throw new Error(`working repository ${repository.owner_slug}/${repository.repo_slug} is not denied`);
    }
  }
  if (!Array.isArray(boundary.allowed_targets)) {
    throw new Error("lab boundary.allowed_targets must be an array");
  }
  const allowedTargets = boundary.allowed_targets.map((entry, index) => {
    const normalized = normalizeLabTarget(entry, `lab boundary.allowed_targets[${index}]`, { allowKey: true });
    if (normalized.run_owned !== true) {
      throw new Error(`lab boundary allowed target ${targetLabel(normalized)} is not marked run_owned`);
    }
    return { key: requireBoundaryIdentifier(entry.key, `lab boundary.allowed_targets[${index}].key`), ...normalized };
  });
  const byIdentity = new Set();
  const keys = new Set();
  for (const target of allowedTargets) {
    const identity = targetIdentity(target);
    if (byIdentity.has(identity) || keys.has(target.key)) {
      throw new Error(`lab boundary has ambiguous allowlist target ${targetLabel(target)}`);
    }
    if (isProtectedTarget({ protected_owners: protectedOwners, protected_repositories: protectedRepositories, working_repositories: workingRepositories }, target)) {
      throw new Error(`lab boundary allowlist includes protected target ${targetLabel(target)}`);
    }
    byIdentity.add(identity);
    keys.add(target.key);
  }
  if (boundary.historical_runs !== undefined && !Array.isArray(boundary.historical_runs)) {
    throw new Error("lab boundary.historical_runs must be an array when present");
  }
  const historicalRuns = (boundary.historical_runs || []).map((entry, index) =>
    normalizeHistoricalRun(entry, `lab boundary.historical_runs[${index}]`),
  );
  const historicalRunIDs = new Set();
  for (const historical of historicalRuns) {
    if (historical.run_id === runID || historicalRunIDs.has(historical.run_id)) {
      throw new Error(`lab boundary has an ambiguous historical run ${historical.run_id}`);
    }
    historicalRunIDs.add(historical.run_id);
    if (isProtectedTarget({ protected_owners: protectedOwners, protected_repositories: protectedRepositories, working_repositories: workingRepositories }, historical.target)) {
      throw new Error(`lab boundary historical run includes protected target ${targetLabel(historical.target)}`);
    }
    if (allowedTargets.some((target) => targetIdentity(target) === targetIdentity(historical.target))) {
      throw new Error(`lab boundary historical target must not be executable ${targetLabel(historical.target)}`);
    }
  }
  if (boundary.bootstrap_principals !== undefined && !Array.isArray(boundary.bootstrap_principals)) {
    throw new Error("lab boundary.bootstrap_principals must be an array when present");
  }
  const bootstrapPrincipals = (boundary.bootstrap_principals || []).map((entry, index) =>
    normalizeBootstrapPrincipal(entry, `lab boundary.bootstrap_principals[${index}]`),
  );
  const bootstrapKeys = new Set();
  const bootstrapIdentities = new Set();
  for (const principal of bootstrapPrincipals) {
    if (protectedOwners.includes(principal.user_slug)) {
      throw new Error(`lab bootstrap principal ${principal.user_slug} is protected`);
    }
    const identity = `${principal.user_slug}:${principal.user_id}:${principal.requested_repo_slug}`;
    if (bootstrapKeys.has(principal.key) || bootstrapIdentities.has(identity)) {
      throw new Error(`lab boundary has ambiguous bootstrap principal ${principal.user_slug}`);
    }
    bootstrapKeys.add(principal.key);
    bootstrapIdentities.add(identity);
  }
  return {
    schema_version: SCHEMA_VERSION,
    run_id: runID,
    default_deny: true,
    protected_owners: protectedOwners,
    protected_repositories: protectedRepositories,
    working_repositories: workingRepositories,
    allowed_targets: allowedTargets,
    historical_runs: historicalRuns,
    bootstrap_principals: bootstrapPrincipals,
  };
}

/** Refuse a target unless both its human slug and immutable provider identity match. */
export function authorizeLabTarget(boundaryCandidate, targetCandidate) {
  const boundary = validateLabBoundary(boundaryCandidate);
  const target = normalizeLabTarget(targetCandidate, "lab target", { allowKey: true });
  if (isProtectedTarget(boundary, target)) {
    throw new Error(`lab target denied: ${targetLabel(target)} is protected`);
  }
  const exactMatches = boundary.allowed_targets.filter((allowed) => targetIdentity(allowed) === targetIdentity(target));
  if (exactMatches.length === 1) return exactMatches[0];
  const slugMatches = boundary.allowed_targets.filter((allowed) => sameSlug(allowed, target));
  if (slugMatches.length > 0) {
    throw new Error(`lab target denied: immutable provider ID does not match allowlist slug ${targetLabel(target)}`);
  }
  throw new Error(`lab target denied: no exact run-owned allowlist match for ${targetLabel(target)}`);
}

/** Validate the captain's PM-only tool boundary for one fixture operation. */
export function assertPMOnly(invocation) {
  const command = requireString(invocation, "fixture invocation");
  if (/\r|\n|\0/.test(command) || !/^pm\s+github\s+/u.test(command)) {
    throw new Error("PM-only fixture invocation must start with 'pm github '");
  }
  ensureNoSensitiveData(command, "fixture invocation");
  return command;
}

/**
 * Preserve a diagnostic PM JSON Error envelope without turning an unclassified
 * provider failure into a credential or setup result.  Evidence is admissible
 * only when the entire envelope is JSON, bounded, and contains no sensitive
 * field or credential-shaped value.  A missing HTTP status remains `null`;
 * callers must not infer one.
 */
export function capturePMErrorEnvelope({ invocation, envelope }) {
  const command = assertPMOnly(invocation);
  const errorEnvelope = requirePlainObject(envelope, "PM Error envelope");
  if (errorEnvelope.kind !== "Error") {
    throw new Error("PM diagnostic requires a kind: Error envelope");
  }
  ensureNoSensitiveData(errorEnvelope, "PM Error envelope");

  let serialized;
  try {
    serialized = JSON.stringify(errorEnvelope);
  } catch {
    throw new Error("PM Error envelope must be JSON serializable");
  }
  if (Buffer.byteLength(serialized, "utf8") > OUTPUT_LIMIT_BYTES) {
    throw new Error("PM Error envelope exceeds bounded output");
  }

  const statusFields = ["status", "http_status", "provider_status"].filter(
    (field) => errorEnvelope[field] !== undefined,
  );
  if (statusFields.length > 1) {
    throw new Error("PM Error envelope has ambiguous provider status");
  }
  let providerStatus = null;
  if (statusFields.length === 1) {
    const status = errorEnvelope[statusFields[0]];
    if (!Number.isInteger(status) || status < 100 || status > 599) {
      throw new Error("PM Error envelope provider status must be an HTTP status");
    }
    providerStatus = status;
  }

  return {
    invocation: command,
    envelope: JSON.parse(serialized),
    provider_status: providerStatus,
  };
}

export function authorizeAccountBootstrapProbe(candidate) {
  const probe = requirePlainObject(candidate, "account bootstrap probe");
  for (const key of Object.keys(probe)) {
    if (key !== "command") throw new Error(`account bootstrap probe forbids ${JSON.stringify(key)}`);
  }
  requireString(probe.command, "account bootstrap probe.command");
  throw new Error(TARGETLESS_ACCOUNT_PROBE_ERROR);
}

export function assertAccountBootstrapProbeInvocation(invocation) {
  assertPMOnly(invocation);
  throw new Error(TARGETLESS_ACCOUNT_PROBE_ERROR);
}

/**
 * The executor is injected so tests can prove that target validation completes
 * before any subprocess launch. Live callers pass a bounded PM-only executor.
 */
export async function executePMFixtureStep({ boundary, target, invocation, execute }) {
  authorizeLabTarget(boundary, target);
  const command = assertPMOnly(invocation);
  if (typeof execute !== "function") throw new Error("PM fixture executor must be a function");
  return execute(command);
}

function invocationFlagValue(invocation, name) {
  const fields = invocation.trim().split(/\s+/u);
  for (let index = 0; index < fields.length; index += 1) {
    const field = fields[index];
    if (field === `--${name}`) return fields[index + 1] || "";
    if (field.startsWith(`--${name}=`)) return field.slice(name.length + 3);
  }
  return "";
}

function invocationHasTrueFlag(invocation, name) {
  return invocation.trim().split(/\s+/u).some((field) => field === `--${name}` || field === `--${name}=true`);
}

/**
 * A bootstrap is not a broad owner exception. It is one exact authenticated
 * principal creating one exact private lab slug; after create, the repository
 * must be resolved and added to allowed_targets with its immutable ID.
 */
export function authorizeBootstrapRepoCreate(boundaryCandidate, requestCandidate) {
  const boundary = validateLabBoundary(boundaryCandidate);
  const request = requirePlainObject(requestCandidate, "bootstrap request");
  if (requireString(request.command, "bootstrap request.command") !== "repo create") {
    throw new Error("bootstrap request denied: only repo create is allowed");
  }
  const principal = requirePlainObject(request.principal, "bootstrap request.principal");
  const userSlug = normalizedSlug(principal.user_slug);
  const userID = requireString(principal.user_id, "bootstrap request.principal.user_id (immutable ID)");
  const record = requirePlainObject(request.record, "bootstrap request.record");
  const requestedRepo = normalizedSlug(record.name);
  if (record.private !== true || record.auto_init !== true) {
    throw new Error("bootstrap request denied: repository must be private and auto-initialized");
  }
  const matches = boundary.bootstrap_principals.filter(
    (entry) => entry.user_slug === userSlug && entry.user_id === userID && entry.requested_repo_slug === requestedRepo,
  );
  if (matches.length !== 1) {
    throw new Error("bootstrap request denied: principal, immutable ID, or requested repository does not match boundary");
  }
  return matches[0];
}

function assertBootstrapInvocation(invocation, principal) {
  const command = assertPMOnly(invocation);
  if (!command.startsWith("pm github repo create ")) {
    throw new Error("bootstrap invocation must be pm github repo create");
  }
  if (normalizedSlug(invocationFlagValue(command, "name")) !== principal.requested_repo_slug) {
    throw new Error("bootstrap invocation denied: repository name does not match boundary");
  }
  if (!invocationHasTrueFlag(command, "private") || !invocationHasTrueFlag(command, "auto-init")) {
    throw new Error("bootstrap invocation denied: repository must be private and auto-initialized");
  }
  return command;
}

export async function executeBootstrapPMFixtureStep({ boundary, request, invocation, execute }) {
  const principal = authorizeBootstrapRepoCreate(boundary, request);
  const command = assertBootstrapInvocation(invocation, principal);
  if (typeof execute !== "function") throw new Error("PM bootstrap executor must be a function");
  return execute(command);
}

/**
 * `repo view` is intentionally credential-pinned in the preserved 124-case
 * live proof, so it cannot safely resolve a new repository before a target ID
 * exists. The captain-approved bootstrap resolver is the PM-only authenticated
 * repository listing; it may select only the one exact private lab slug under
 * the already authenticated immutable user.
 */
export function authorizeBootstrapRepoDiscovery(boundaryCandidate, requestCandidate) {
  const boundary = validateLabBoundary(boundaryCandidate);
  const request = requirePlainObject(requestCandidate, "bootstrap discovery request");
  if (requireString(request.command, "bootstrap discovery request.command") !== "repos list-for-authenticated-user") {
    throw new Error("bootstrap discovery denied: only repos list-for-authenticated-user is allowed");
  }
  const principal = requirePlainObject(request.principal, "bootstrap discovery request.principal");
  const userSlug = normalizedSlug(principal.user_slug);
  const userID = requireString(principal.user_id, "bootstrap discovery request.principal.user_id (immutable ID)");
  const repository = requirePlainObject(request.repository, "bootstrap discovery request.repository");
  const ownerSlug = normalizedSlug(repository.owner_slug);
  const repoSlug = normalizedSlug(repository.repo_slug);
  if (ownerSlug !== userSlug) {
    throw new Error("bootstrap discovery denied: repository owner must match the authenticated user");
  }
  const matches = boundary.bootstrap_principals.filter(
    (entry) => entry.user_slug === userSlug && entry.user_id === userID && entry.requested_repo_slug === repoSlug,
  );
  if (matches.length !== 1) {
    throw new Error("bootstrap discovery denied: principal, immutable ID, or requested repository does not match boundary");
  }
  const alreadyResolved = boundary.allowed_targets.some(
    (entry) => entry.resource_type === "repository" && entry.owner_slug === ownerSlug && entry.repo_slug === repoSlug,
  );
  if (alreadyResolved) {
    throw new Error("bootstrap discovery denied: repository is already bound to an immutable allowlist target");
  }
  return matches[0];
}

function assertBootstrapDiscoveryInvocation(invocation) {
  const command = assertPMOnly(invocation);
  const fields = command.trim().split(/\s+/u);
  const prefix = ["pm", "github", "repos", "list-for-authenticated-user"];
  if (fields.length < prefix.length || prefix.some((part, index) => fields[index] !== part)) {
    throw new Error("bootstrap discovery invocation must be pm github repos list-for-authenticated-user");
  }
  const values = { credential: [], root: [], json: 0 };
  for (let index = prefix.length; index < fields.length; index += 1) {
    const field = fields[index];
    if (field === "--json" || field === "--json=true") {
      values.json += 1;
      continue;
    }
    const match = /^--(credential|root)=(.+)$/u.exec(field);
    if (match) {
      values[match[1]].push(match[2]);
      continue;
    }
    if (field === "--credential" || field === "--root") {
      const value = fields[index + 1];
      if (!value || value.startsWith("--")) {
        throw new Error(`bootstrap discovery invocation ${field} requires a value`);
      }
      values[field.slice(2)].push(value);
      index += 1;
      continue;
    }
    throw new Error(`bootstrap discovery invocation forbids ${field}; use the preserved credential-pinned PM shape`);
  }
  if (values.credential.length !== 1 || values.root.length !== 1 || values.json !== 1) {
    throw new Error("bootstrap discovery invocation requires exactly one --credential, --root, and --json");
  }
  safeCredentialName(values.credential[0]);
  requireString(values.root[0], "bootstrap discovery invocation root");
  return command;
}

/** Refuse a broad listing unless the boundary resolves it to one exact bootstrap principal first. */
export async function executeBootstrapPMDiscoveryStep({ boundary, request, invocation, execute }) {
  authorizeBootstrapRepoDiscovery(boundary, request);
  const command = assertBootstrapDiscoveryInvocation(invocation);
  if (typeof execute !== "function") throw new Error("PM bootstrap discovery executor must be a function");
  return execute(command);
}

/**
 * Filter an authenticated-repository PM response only in memory. The output
 * deliberately contains the known slug and immutable IDs, never a provider
 * response body or unrelated account records.
 */
export function resolveBootstrapRepositoryTarget({ envelope, principal, repository }) {
  const safeEnvelope = requirePlainObject(envelope, "bootstrap discovery envelope");
  if (
    safeEnvelope.kind !== "ConnectorCommandDirectRead" ||
    safeEnvelope.connector !== "github" ||
    safeEnvelope.command !== "repos list-for-authenticated-user" ||
    !Number.isInteger(safeEnvelope.status) ||
    safeEnvelope.status < 200 ||
    safeEnvelope.status >= 300
  ) {
    throw new Error("bootstrap discovery did not return a successful authenticated repository PM response");
  }
  const authenticated = requirePlainObject(principal, "bootstrap discovery principal");
  const ownerSlug = normalizedSlug(authenticated.user_slug);
  const ownerID = requireString(authenticated.user_id, "bootstrap discovery principal.user_id (immutable ID)");
  const requested = requirePlainObject(repository, "bootstrap discovery repository");
  const requestedOwner = normalizedSlug(requested.owner_slug);
  const repoSlug = normalizedSlug(requested.repo_slug);
  if (requestedOwner !== ownerSlug) {
    throw new Error("bootstrap discovery repository owner must match the authenticated user");
  }
  if (!Array.isArray(safeEnvelope.response)) {
    throw new Error("bootstrap discovery response must be one JSON repository page");
  }
  const matches = safeEnvelope.response.filter((record) => {
    if (!isPlainObject(record) || record.private !== true || !isPlainObject(record.owner)) return false;
    if (typeof record.name !== "string" || typeof record.owner.login !== "string") return false;
    const recordOwnerID = record.owner.id;
    const sameOwnerID = String(recordOwnerID) === ownerID;
    return normalizedSlug(record.name) === repoSlug && normalizedSlug(record.owner.login) === ownerSlug && sameOwnerID;
  });
  if (matches.length !== 1) {
    throw new Error("bootstrap discovery must find exactly one private repository under the authenticated immutable user");
  }
  const repoID = requireProviderID(matches[0].id, "bootstrap discovery repository.id");
  return normalizeLabTarget({
    resource_type: "repository",
    owner_slug: ownerSlug,
    owner_id: ownerID,
    repo_slug: repoSlug,
    repo_id: repoID,
    run_owned: true,
  });
}

function matchingBoundLabLabels(envelopeCandidate, nameCandidate) {
  const envelope = requirePlainObject(envelopeCandidate, "label-list envelope");
  const name = requireString(nameCandidate, "lab label name");
  ensureNoSensitiveData(name, "lab label name");
  if (envelope.kind !== "ConnectorCommandRead" || envelope.connector !== "github" || envelope.command !== "label list") {
    throw new Error("label read-back requires a PM github label list envelope");
  }
  if (!Array.isArray(envelope.records)) {
    throw new Error("label read-back label list envelope is malformed");
  }
  const expected = name.toLowerCase();
  return {
    name,
    matches: envelope.records.filter((record) => isPlainObject(record) && typeof record.name === "string" && record.name.toLowerCase() === expected),
  };
}

function matchingBoundLabDeployKeys(envelopeCandidate, titleCandidate) {
  const envelope = requirePlainObject(envelopeCandidate, "deploy-key-list envelope");
  const title = requireString(titleCandidate, "lab deploy-key title");
  ensureNoSensitiveData(title, "lab deploy-key title");
  if (envelope.kind !== "ConnectorCommandRead" || envelope.connector !== "github" || envelope.command !== "repo deploy-key list") {
    throw new Error("deploy-key read-back requires a PM github repo deploy-key list envelope");
  }
  if (!Array.isArray(envelope.records)) {
    throw new Error("deploy-key read-back deploy-key list envelope is malformed");
  }
  return {
    title,
    matches: envelope.records.filter((record) => isPlainObject(record) && record.title === title),
  };
}

/** Return only generated deploy-key identity facts; public-key material stays process-local. */
export function resolveBoundLabDeployKey({ envelope, title, readOnly = true }) {
  if (typeof readOnly !== "boolean") throw new Error("expected generated deploy key read-only state must be boolean");
  const result = matchingBoundLabDeployKeys(envelope, title);
  if (result.matches.length !== 1) {
    throw new Error("deploy-key read-back must find exactly one generated deploy key");
  }
  const record = result.matches[0];
  if (record.read_only !== readOnly) {
    throw new Error("generated deploy key read-only state does not match the expected value");
  }
  return {
    id: requireProviderID(record.id, "generated deploy-key immutable provider ID"),
    title: result.title,
  };
}

/** Refuse to create a deploy key when its generated title already exists in the bound repository. */
export function assertBoundLabDeployKeyAbsent({ envelope, title }) {
  const result = matchingBoundLabDeployKeys(envelope, title);
  if (result.matches.length !== 0) {
    throw new Error("generated lab deploy key already exists; provenance is not safe to assume");
  }
}

/**
 * Retry only a successful PM deploy-key-list assertion while GitHub makes an
 * accepted fixture visible. PM read failures propagate immediately so a scope
 * or credential divergence is never hidden by the visibility wait.
 */
export async function waitForBoundLabDeployKey({ read, title, readOnly = true, maxAttempts = ISSUE_READBACK_MAX_ATTEMPTS, sleep }) {
  if (typeof read !== "function") throw new Error("PM deploy-key read-back requires a read function");
  if (!Number.isSafeInteger(maxAttempts) || maxAttempts < 1 || maxAttempts > ISSUE_READBACK_MAX_ATTEMPTS) {
    throw new Error(`PM deploy-key read-back attempts must be between 1 and ${ISSUE_READBACK_MAX_ATTEMPTS}`);
  }
  const pause = sleep ?? ((milliseconds) => new Promise((resolve) => setTimeout(resolve, milliseconds)));
  if (typeof pause !== "function") throw new Error("PM deploy-key read-back sleep function must be a function");

  for (let attempt = 1; attempt <= maxAttempts; attempt += 1) {
    const envelope = await read();
    try {
      return { ...resolveBoundLabDeployKey({ envelope, title, readOnly }), attempts: attempt };
    } catch {
      if (attempt === maxAttempts) break;
      await pause(ISSUE_READBACK_RETRY_DELAY_MS);
    }
  }
  throw new Error(`PM deploy-key read-back did not converge within ${maxAttempts} PM-only attempts`);
}

/** Retry only stale successful PM list results until the generated deploy-key title is absent. */
export async function waitForBoundLabDeployKeyAbsent({ read, title, maxAttempts = ISSUE_READBACK_MAX_ATTEMPTS, sleep }) {
  if (typeof read !== "function") throw new Error("PM deploy-key absence read-back requires a read function");
  if (!Number.isSafeInteger(maxAttempts) || maxAttempts < 1 || maxAttempts > ISSUE_READBACK_MAX_ATTEMPTS) {
    throw new Error(`PM deploy-key absence attempts must be between 1 and ${ISSUE_READBACK_MAX_ATTEMPTS}`);
  }
  const pause = sleep ?? ((milliseconds) => new Promise((resolve) => setTimeout(resolve, milliseconds)));
  if (typeof pause !== "function") throw new Error("PM deploy-key absence sleep function must be a function");

  for (let attempt = 1; attempt <= maxAttempts; attempt += 1) {
    const envelope = await read();
    try {
      assertBoundLabDeployKeyAbsent({ envelope, title });
      return { attempts: attempt };
    } catch {
      if (attempt === maxAttempts) break;
      await pause(ISSUE_READBACK_RETRY_DELAY_MS);
    }
  }
  throw new Error(`PM deploy-key absence read-back did not converge within ${maxAttempts} PM-only attempts`);
}

/** Return only the known label name and its immutable provider ID from a PM list response. */
export function resolveBoundLabLabel({ envelope, name }) {
  const result = matchingBoundLabLabels(envelope, name);
  if (result.matches.length !== 1) {
    throw new Error("label read-back must find exactly one generated label");
  }
  return {
    id: requireProviderID(result.matches[0].id, "generated label immutable provider ID"),
    name: result.name,
  };
}

function normalizeLabLabelColor(value, label) {
  const color = requireString(value, label).replace(/^#/u, "").toLowerCase();
  if (!/^[0-9a-f]{6}$/u.test(color)) throw new Error(`${label} must be a six-digit hexadecimal label color`);
  return color;
}

/** Assert only caller-declared, non-secret label properties from the PM list response. */
export function assertBoundLabLabelProperties({ envelope, name, color, description }) {
  const result = matchingBoundLabLabels(envelope, name);
  if (result.matches.length !== 1) {
    throw new Error("label property read-back must find exactly one generated label");
  }
  const record = result.matches[0];
  if (color !== undefined) {
    const expectedColor = normalizeLabLabelColor(color, "expected generated label color");
    const actualColor = normalizeLabLabelColor(record.color, "returned generated label color");
    if (actualColor !== expectedColor) throw new Error("generated label color does not match the expected value");
  }
  if (description !== undefined) {
    const expectedDescription = requireString(description, "expected generated label description");
    ensureNoSensitiveData(expectedDescription, "expected generated label description");
    if (record.description !== expectedDescription) {
      throw new Error("generated label description does not match the expected value");
    }
  }
  return {
    id: requireProviderID(record.id, "generated label immutable provider ID"),
    name: result.name,
  };
}

/** Refuse to create a label that a PM read already found in the bound repository. */
export function assertBoundLabLabelAbsent({ envelope, name }) {
  const result = matchingBoundLabLabels(envelope, name);
  if (result.matches.length !== 0) {
    throw new Error("generated lab label already exists; provenance is not safe to assume");
  }
}

function matchingBoundLabIssues(envelopeCandidate, titleCandidate) {
  const envelope = requirePlainObject(envelopeCandidate, "issue-list envelope");
  const title = requireString(titleCandidate, "lab issue title");
  ensureNoSensitiveData(title, "lab issue title");
  if (envelope.kind !== "ConnectorCommandRead" || envelope.connector !== "github" || envelope.command !== "issue list") {
    throw new Error("issue read-back requires a PM github issue list envelope");
  }
  if (!Array.isArray(envelope.records)) {
    throw new Error("issue read-back issue list envelope is malformed");
  }
  return {
    title,
    matches: envelope.records.filter((record) => isPlainObject(record) && record.title === title),
  };
}

/** Return only the generated issue's immutable ID and issue number after bounded property checks. */
export function resolveBoundLabIssue({ envelope, title, body, state, expectedComments, minComments }) {
  const result = matchingBoundLabIssues(envelope, title);
  if (result.matches.length !== 1) {
    throw new Error("issue read-back must find exactly one generated issue");
  }
  const record = result.matches[0];
  if (body !== undefined) {
    const expectedBody = requireString(body, "expected generated issue body");
    ensureNoSensitiveData(expectedBody, "expected generated issue body");
    if (record.body !== expectedBody) throw new Error("generated issue body does not match the expected value");
  }
  if (state !== undefined) {
    const expectedState = requireString(state, "expected generated issue state");
    if (record.state !== expectedState) throw new Error("generated issue state does not match the expected value");
  }
  if (expectedComments !== undefined) {
    if (!Number.isSafeInteger(expectedComments) || expectedComments < 0) {
      throw new Error("expected generated issue comment count must be a non-negative integer");
    }
    if (!Number.isSafeInteger(record.comments) || record.comments !== expectedComments) {
      throw new Error("generated issue comment count does not match the expected value");
    }
  }
  if (minComments !== undefined) {
    if (!Number.isSafeInteger(minComments) || minComments < 0) {
      throw new Error("expected generated issue minimum comment count must be a non-negative integer");
    }
    if (!Number.isSafeInteger(record.comments) || record.comments < minComments) {
      throw new Error("generated issue comment count does not meet the expected minimum");
    }
  }
  if (!Number.isSafeInteger(record.number) || record.number < 1) {
    throw new Error("generated issue read-back is missing a positive issue number");
  }
  return {
    id: requireProviderID(record.node_id ?? record.id, "generated issue immutable provider ID"),
    number: record.number,
  };
}

/** Refuse to create an issue that a PM list already found in the bound repository. */
export function assertBoundLabIssueAbsent({ envelope, title }) {
  const result = matchingBoundLabIssues(envelope, title);
  if (result.matches.length !== 0) {
    throw new Error("generated lab issue already exists; provenance is not safe to assume");
  }
}

/**
 * Retry only a successful PM issue-list assertion while GitHub makes an
 * accepted write visible. A PM read failure propagates immediately so a
 * credential, entitlement, or scope divergence cannot be masked as a delay.
 */
export async function waitForBoundLabIssue({ read, title, body, state, expectedComments, minComments, maxAttempts = ISSUE_READBACK_MAX_ATTEMPTS, sleep }) {
  if (typeof read !== "function") throw new Error("PM issue read-back requires a read function");
  if (!Number.isSafeInteger(maxAttempts) || maxAttempts < 1 || maxAttempts > ISSUE_READBACK_MAX_ATTEMPTS) {
    throw new Error(`PM issue read-back attempts must be between 1 and ${ISSUE_READBACK_MAX_ATTEMPTS}`);
  }
  const pause = sleep ?? ((milliseconds) => new Promise((resolve) => setTimeout(resolve, milliseconds)));
  if (typeof pause !== "function") throw new Error("PM issue read-back sleep function must be a function");

  for (let attempt = 1; attempt <= maxAttempts; attempt += 1) {
    const envelope = await read();
    try {
      return { ...resolveBoundLabIssue({ envelope, title, body, state, expectedComments, minComments }), attempts: attempt };
    } catch {
      if (attempt === maxAttempts) break;
      await pause(ISSUE_READBACK_RETRY_DELAY_MS);
    }
  }
  throw new Error(`PM issue read-back did not converge within ${maxAttempts} PM-only attempts`);
}

function commandParts(command) {
  const value = requireString(command, "PM command");
  const parts = value.split(" ");
  if (parts.some((part) => !/^[a-z0-9_-]+$/iu.test(part))) {
    throw new Error("PM command must be a registered GitHub command path");
  }
  return parts;
}

function safeCredentialName(value) {
  const name = requireString(value, "credential name");
  if (!/^[A-Za-z0-9._-]+$/u.test(name)) {
    throw new Error("credential name contains unsupported characters");
  }
  ensureNoSensitiveData(name, "credential name");
  return name;
}

function validateRecordArgs(args) {
  if (!Array.isArray(args) || args.some((argument) => typeof argument !== "string" || /\r|\n|\0/u.test(argument))) {
    throw new Error("planned PM write record arguments must be a string array");
  }
  const forbidden = new Set(["--approve", "--approval-token-stdin", "--confirm", "--connection", "--credential", "--plan", "--root"]);
  if (args.some((argument) => forbidden.has(argument) || [...forbidden].some((flag) => argument.startsWith(`${flag}=`)))) {
    throw new Error("planned PM write record arguments may not override lifecycle or credential flags");
  }
  args.forEach((argument, index) => ensureNoSensitiveData(argument, `planned PM write record argument ${index}`));
  return [...args];
}

function flagValuesFromRecordArgs(args, name) {
  const values = [];
  const flag = `--${name}`;
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    if (argument === flag) {
      const value = args[index + 1];
      if (!value || value.startsWith("--")) throw new Error(`${flag} requires a value in a planned PM write`);
      values.push(value);
      index += 1;
      continue;
    }
    if (argument.startsWith(`${flag}=`)) {
      const value = argument.slice(flag.length + 1);
      if (value === "") throw new Error(`${flag} requires a value in a planned PM write`);
      values.push(value);
    }
  }
  return values;
}

function scopedWriteArguments(args) {
  const config = new Map();
  for (const raw of flagValuesFromRecordArgs(args, "config")) {
    const separator = raw.indexOf("=");
    if (separator < 1) throw new Error("--config requires key=value in a planned PM write");
    const key = raw.slice(0, separator).trim();
    const value = raw.slice(separator + 1);
    if (SENSITIVE_KEY.test(key)) throw new Error("planned PM write configuration is sensitive");
    const values = config.get(key) || [];
    values.push(value);
    config.set(key, values);
  }
  return {
    config,
    owner: flagValuesFromRecordArgs(args, "owner"),
    repo: flagValuesFromRecordArgs(args, "repo"),
    repoOwner: flagValuesFromRecordArgs(args, "repo-owner"),
    repoName: flagValuesFromRecordArgs(args, "repo-name"),
    repository: flagValuesFromRecordArgs(args, "repository"),
  };
}

function requireExactScopeValues(values, expected, label, { required = false } = {}) {
  if (required && values.length !== 1) {
    throw new Error(`normal repository write requires exactly one ${label} matching the ID-bound target`);
  }
  if (values.length > 1) throw new Error(`normal repository write has ambiguous ${label} values`);
  if (values.length === 1 && normalizedSlug(values[0]) !== expected) {
    throw new Error(`normal repository write ${label} does not match the ID-bound target`);
  }
}

function assertRepositoryWriteScope(target, args) {
  if (target.resource_type !== "repository") {
    throw new Error(`normal planned write does not yet support target type ${target.resource_type}`);
  }
  const scope = scopedWriteArguments(args);
  requireExactScopeValues(scope.config.get("owner") || [], target.owner_slug, "--config owner", { required: true });
  requireExactScopeValues(scope.config.get("repo") || [], target.repo_slug, "--config repo", { required: true });
  requireExactScopeValues(scope.owner, target.owner_slug, "--owner");
  requireExactScopeValues(scope.repo, target.repo_slug, "--repo");
  requireExactScopeValues(scope.repoOwner, target.owner_slug, "--repo-owner");
  requireExactScopeValues(scope.repoName, target.repo_slug, "--repo-name");
  requireExactScopeValues(scope.repository, `${target.owner_slug}/${target.repo_slug}`, "--repository");
}

function redactText(value) {
  let text = String(value ?? "");
  return redactPersistedText(text);
}

export async function runPMProcess(binary, args, stdin = "") {
  return new Promise((resolve, reject) => {
    const child = spawn(binary, args, { stdio: ["pipe", "pipe", "pipe"] });
    let stdout = "";
    let stderr = "";
    let bytes = 0;
    let overflow = false;
    const consume = (target, chunk) => {
      bytes += chunk.length;
      if (bytes > OUTPUT_LIMIT_BYTES) {
        overflow = true;
        child.kill("SIGTERM");
        return;
      }
      if (target === "stdout") stdout += chunk.toString("utf8");
      else stderr += chunk.toString("utf8");
    };
    child.stdout.on("data", (chunk) => consume("stdout", chunk));
    child.stderr.on("data", (chunk) => consume("stderr", chunk));
    child.stdin.on("error", reject);
    child.on("error", reject);
    child.on("close", (code, signal) => resolve({ code, signal, stdout, stderr, overflow }));
    child.stdin.end(stdin);
  });
}

function normalizeProcessResult(result, step) {
  if (!isPlainObject(result) || typeof result.code !== "number" || typeof result.stdout !== "string" || typeof result.stderr !== "string") {
    throw new Error(`PM ${step} executor returned an invalid result`);
  }
  if (result.overflow === true) throw new Error(`PM ${step} exceeded bounded output`);
  if (result.code !== 0) {
    const status = /\b(?:http|status)\s+(\d{3})\b/iu.exec(`${result.stdout}\n${result.stderr}`)?.[1];
    throw new Error(status ? `PM ${step} failed with provider status ${status}` : `PM ${step} failed without a safe provider status`);
  }
  return result;
}

export async function runPMAccountBootstrapProbe({ probe }) {
  authorizeAccountBootstrapProbe(probe);
}

function planIdentity(stdout) {
  const planID = /Created connector command plan\s+(\S+)/u.exec(stdout)?.[1] || "";
  if (planID === "") throw new Error("PM write plan omitted plan identity");
  return planID;
}

function approvalMaterial(stdout) {
  return {
    grant: /Approval token:\s*(\S+)/u.exec(stdout)?.[1] || "",
    challenge: /Confirmation required:\s+--confirm\s+(\S+)/u.exec(stdout)?.[1] || "",
  };
}

/**
 * Execute the existing PM reverse-ETL lifecycle without serializing transient
 * plan/approval material. Callers may inject `run` for deterministic tests.
 */
export async function runPMPlannedWrite({ binary, root, credentialName, command, recordArgs, boundary, target, bootstrapRequest, run }) {
  const pmBinary = requireString(binary, "PM binary");
  const projectRoot = requireString(root, "PM project root");
  const parts = commandParts(command);
  const args = validateRecordArgs(recordArgs);
  if (bootstrapRequest !== undefined && target !== undefined) {
    throw new Error("planned PM write cannot combine bootstrap and normal ID-bound targets");
  }
  if (bootstrapRequest !== undefined) {
    authorizeBootstrapRepoCreate(boundary, bootstrapRequest);
  } else {
    if (target === undefined) throw new Error("normal planned PM write requires an ID-bound lab target");
    const allowedTarget = authorizeLabTarget(boundary, target);
    assertRepositoryWriteScope(allowedTarget, args);
  }
  const credential = safeCredentialName(credentialName);
  const runner = run || ((processArgs, stdin) => runPMProcess(pmBinary, processArgs, stdin));
  if (typeof runner !== "function") throw new Error("PM write executor must be a function");

  const planArgs = ["github", ...parts, ...args, "--credential", credential, "--root", projectRoot];
  if (bootstrapRequest !== undefined) {
    const principal = authorizeBootstrapRepoCreate(boundary, bootstrapRequest);
    assertBootstrapInvocation(["pm", ...planArgs].join(" "), principal);
  }
  const plan = normalizeProcessResult(await runner(planArgs), "write plan");
  const planID = planIdentity(plan.stdout);
  const plannedMaterial = approvalMaterial(plan.stdout);

  // PM intentionally withholds declared-sensitive record fields from a persisted
  // reverse plan. Re-supply the validated, process-local arguments at both later
  // lifecycle stages so those fields can be hash-bound for preview and execution.
  const previewArgs = ["github", ...parts, "--plan", planID, "--preview", ...args, "--root", projectRoot];
  const preview = normalizeProcessResult(await runner(previewArgs), "write preview");
  const previewMaterial = approvalMaterial(preview.stdout);
  const grant = previewMaterial.grant || plannedMaterial.grant;
  if (grant === "") throw new Error("PM write preview omitted approval grant");
  const challenge = previewMaterial.challenge || plannedMaterial.challenge;

  const executeArgs = [
    "github",
    ...parts,
    "--plan",
    planID,
    "--approval-token-stdin",
    ...(challenge ? ["--confirm", challenge] : []),
    ...args,
    "--root",
    projectRoot,
    "--json",
  ];
  const execution = normalizeProcessResult(await runner(executeArgs, grant + "\n"), "write execution");
  let envelope;
  try {
    envelope = JSON.parse(execution.stdout);
  } catch {
    throw new Error("PM write execution did not produce machine-readable JSON");
  }
  const runRecord = envelope?.run;
  if (envelope?.kind !== "ReverseRun" || runRecord?.status !== "completed" || runRecord?.records_succeeded !== 1 || runRecord?.records_failed !== 0) {
    throw new Error("PM write execution did not report one completed provider mutation");
  }
  const status = runRecord?.operation_direct_write?.status;
  return {
    command: parts.join(" "),
    ...(Number.isInteger(status) ? { http_status: status } : {}),
    records_succeeded: runRecord.records_succeeded,
    records_failed: runRecord.records_failed,
  };
}

/**
 * Execute one non-mutating PM connector command only after the same immutable
 * repository binding used for writes has accepted its exact `--config` scope.
 * The provider envelope stays in memory for the caller's narrow assertion and
 * must never be persisted verbatim.
 */
export async function runPMScopedRead({ binary, root, credentialName, command, commandArgs, boundary, target, run }) {
  const pmBinary = requireString(binary, "PM binary");
  const projectRoot = requireString(root, "PM project root");
  const parts = commandParts(command);
  const args = validateRecordArgs(commandArgs);
  const allowedTarget = authorizeLabTarget(boundary, target);
  assertRepositoryWriteScope(allowedTarget, args);
  const credential = safeCredentialName(credentialName);
  const processArgs = ["github", ...parts, ...args, "--credential", credential, "--root", projectRoot, "--json"];
  assertPMOnly(["pm", ...processArgs].join(" "));
  const runner = run || ((argsForProcess) => runPMProcess(pmBinary, argsForProcess));
  if (typeof runner !== "function") throw new Error("PM scoped read executor must be a function");
  const result = normalizeProcessResult(await runner(processArgs), "scoped read");
  let envelope;
  try {
    envelope = JSON.parse(result.stdout);
  } catch {
    throw new Error("PM scoped read did not produce machine-readable JSON");
  }
  if (!isPlainObject(envelope) || envelope.connector !== "github" || envelope.command !== parts.join(" ")) {
    throw new Error("PM scoped read response does not identify the requested GitHub command");
  }
  return envelope;
}

function sanitizeCleanupEntry(candidate) {
  const entry = requirePlainObject(candidate, "cleanup ledger entry");
  ensureNoSensitiveData(entry, "cleanup ledger entry");
  for (const key of Object.keys(entry)) {
    if (!LEDGER_FIELDS.has(key)) {
      throw new Error(`cleanup ledger entry contains forbidden field ${JSON.stringify(key)}`);
    }
  }
  if (entry.schema_version !== SCHEMA_VERSION) {
    throw new Error(`cleanup ledger entry schema_version must be ${SCHEMA_VERSION}`);
  }
  const action = requireString(entry.action, "cleanup ledger entry.action");
  if (!CLEANUP_ACTIONS.has(action)) {
    throw new Error(`cleanup ledger entry action ${JSON.stringify(action)} is not allowed`);
  }
  const out = {
    schema_version: SCHEMA_VERSION,
    run_id: requireString(entry.run_id, "cleanup ledger entry.run_id"),
    action,
  };
  if (action === "initialized") {
    if (entry.note !== undefined) out.note = requireString(entry.note, "cleanup ledger entry.note");
    return out;
  }
  out.fixture_id = requireString(entry.fixture_id, "cleanup ledger entry.fixture_id");
  out.target = normalizeLabTarget(entry.target, "cleanup ledger entry.target", { allowKey: true });
  out.pm_command = assertPMOnly(entry.pm_command);
  if (entry.provider_id !== undefined) out.provider_id = requireString(entry.provider_id, "cleanup ledger entry.provider_id");
  if (entry.residual_state !== undefined) out.residual_state = requireString(entry.residual_state, "cleanup ledger entry.residual_state");
  if (entry.retention_reason !== undefined) out.retention_reason = requireString(entry.retention_reason, "cleanup ledger entry.retention_reason");
  if (action === "retained" && !out.retention_reason) {
    throw new Error("retained cleanup entry requires retention_reason");
  }
  if (action === "cleanup_failed" && !out.residual_state) {
    throw new Error("cleanup_failed cleanup entry requires residual_state");
  }
  return out;
}

/** Append one sanitized JSONL event; never rewrite the existing cleanup ledger. */
export async function appendCleanupEntry(ledgerPath, candidate) {
  const entry = sanitizeCleanupEntry(candidate);
  await mkdir(path.dirname(ledgerPath), { recursive: true, mode: 0o700 });
  await appendFile(ledgerPath, `${JSON.stringify(entry)}\n`, { encoding: "utf8", mode: 0o600, flag: "a" });
  return entry;
}

export async function readCleanupLedger(ledgerPath) {
  let text;
  try {
    text = await readFile(ledgerPath, "utf8");
  } catch (error) {
    if (error && error.code === "ENOENT") return [];
    throw error;
  }
  const entries = [];
  for (const [index, line] of text.split(/\r?\n/u).entries()) {
    if (line.trim() === "") continue;
    try {
      entries.push(JSON.parse(line));
    } catch {
      throw new Error(`cleanup ledger line ${index + 1} is not JSON`);
    }
  }
  return entries;
}

function authorizeHistoricalLedgerTarget(boundary, runID, targetCandidate) {
  const historical = boundary.historical_runs.find((entry) => entry.run_id === runID);
  if (!historical) return undefined;
  const target = normalizeLabTarget(targetCandidate, "historical cleanup ledger target", { allowKey: true });
  if (targetIdentity(target) !== targetIdentity(historical.target)) {
    throw new Error(`historical cleanup ledger target does not match archived run ${runID}`);
  }
  return historical.target;
}

/** Validate ledger order and allow repeated idempotent cleanup observations. */
export function validateCleanupLedger({ entries, boundary }) {
  if (!Array.isArray(entries)) throw new Error("cleanup ledger entries must be an array");
  const normalizedBoundary = validateLabBoundary(boundary);
  const fixtures = new Map();
  let initialized = false;
  for (const [index, rawEntry] of entries.entries()) {
    const entry = sanitizeCleanupEntry(rawEntry);
    const historical = entry.run_id === normalizedBoundary.run_id
      ? undefined
      : normalizedBoundary.historical_runs.find((candidate) => candidate.run_id === entry.run_id);
    if (entry.run_id !== normalizedBoundary.run_id && !historical) {
      throw new Error(`cleanup ledger entry ${index + 1} has a different run_id`);
    }
    if (entry.action === "initialized") {
      if (initialized || index !== 0) throw new Error("cleanup ledger initialized event must appear once at the beginning");
      initialized = true;
      continue;
    }
    const authorizedTarget = historical
      ? authorizeHistoricalLedgerTarget(normalizedBoundary, entry.run_id, entry.target)
      : authorizeLabTarget(normalizedBoundary, entry.target);
    const fixtureKey = `${entry.run_id}\u0000${entry.fixture_id}`;
    const prior = fixtures.get(fixtureKey);
    if (entry.action === "created") {
      if (prior) throw new Error(`cleanup ledger fixture ${JSON.stringify(entry.fixture_id)} was created twice`);
      fixtures.set(fixtureKey, { target: authorizedTarget, terminal: false });
      continue;
    }
    if (!prior) throw new Error(`cleanup ledger action ${entry.action} has no created fixture ${JSON.stringify(entry.fixture_id)}`);
    if (targetIdentity(prior.target) !== targetIdentity(authorizedTarget)) {
      throw new Error(`cleanup ledger fixture ${JSON.stringify(entry.fixture_id)} changed target identity`);
    }
    if (["cleanup_completed", "cleanup_already_absent", "retained"].includes(entry.action)) {
      prior.terminal = true;
    }
  }
  return { entries: entries.length, fixtures: fixtures.size };
}

/** Lab rows cannot silently create multiple terminal outcomes for the same command. */
export function validateLabTerminalRecords(records) {
  if (!Array.isArray(records)) throw new Error("lab terminal records must be an array");
  const seen = new Set();
  for (const record of records) {
    requirePlainObject(record, "lab terminal record");
    ensureNoSensitiveData(record, "lab terminal record");
    const command = requireString(record.command, "lab terminal record.command");
    if (seen.has(command)) throw new Error(`duplicate terminal result for ${JSON.stringify(command)}`);
    seen.add(command);
    const state = requireString(record.state, "lab terminal record.state");
    if (!TERMINAL_STATES.has(state)) throw new Error(`invalid lab terminal state ${JSON.stringify(state)}`);
    if (state === "proven") {
      requireString(record.assertion, "proven lab terminal record.assertion");
    } else {
      const reason = requireString(record.reason, "blocked or failed lab terminal record.reason");
      if (reason.length < 24) throw new Error("blocked or failed lab terminal reason must be concrete");
    }
  }
  return { terminal_records: records.length };
}

function parseArgs(args) {
  const options = {};
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    if (argument === "--check-boundary") {
      options.checkBoundary = true;
      continue;
    }
    if (argument === "--boundary") {
      const value = args[index + 1];
      if (!value || value.startsWith("--")) throw new Error("--boundary requires a path");
      options.boundary = value;
      index += 1;
      continue;
    }
    throw new Error(`unexpected argument ${argument}`);
  }
  return options;
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  if (!options.checkBoundary || !options.boundary) {
    throw new Error("usage: github-live-lab.mjs --check-boundary --boundary <path>");
  }
  const boundary = JSON.parse(await readFile(path.resolve(options.boundary), "utf8"));
  const normalized = validateLabBoundary(boundary);
  process.stdout.write(`github live lab boundary: ok allowed_targets=${normalized.allowed_targets.length}\n`);
}

if (path.resolve(process.argv[1] || "") === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    process.stderr.write(`github live lab: ${error instanceof Error ? error.message : "validation failed"}\n`);
    process.exitCode = 1;
  });
}
