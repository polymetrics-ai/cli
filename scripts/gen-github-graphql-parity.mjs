#!/usr/bin/env node

// Generate the fixed, typed GitHub GraphQL root contracts from the checked-in
// source lock. This script intentionally has no network code: source capture
// is a separate, reviewed process and the generated commands never accept a
// caller-controlled document, selection, endpoint, or cursor.

import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(SCRIPT_DIR, "..");
const GITHUB_DEFS = path.join(ROOT, "internal", "connectors", "defs", "github");
const DEFAULT_LOCK = path.join(GITHUB_DEFS, "sources", "github-operation-source-lock.json");
const DEFAULT_OPERATIONS = path.join(GITHUB_DEFS, "operations.json");
const DEFAULT_CLI = path.join(GITHUB_DEFS, "cli_surface.json");
const DEFAULT_SURFACE = path.join(GITHUB_DEFS, "api_surface.json");

const GRAPHQL_TRANSPORT = { method: "POST", path: "/graphql" };
const MAX_GRAPHQL_BYTES = 1024 * 1024;
// This pre-generator metadata-only operation used to back the blocked
// `issue delete` alias. The source-pinned DeleteIssue mutation now owns that
// alias with an executable, typed contract, so retaining the old unbound
// operation would create a second, non-executable capability record.
const OBSOLETE_GRAPHQL_OPERATION_IDS = new Set(["github.issue.delete"]);
const MAX_GRAPHQL_ARRAY_ITEMS = 100;
const MAX_INPUT_DEPTH = 8;
const MAX_MUTATION_OUTPUT_DEPTH = 3;
const MAX_MUTATION_OUTPUT_FIELDS = 64;
const BUILTIN_SCALARS = new Set(["Boolean", "Float", "ID", "Int", "String"]);
const SENSITIVE_NAME = /(?:secret|token|password|private(?:_|-)?key|encrypted(?:_|-)?value|credential|access(?:_|-)?key)/iu;
const NON_SECRET_SENSITIVE_NAME = /(?:token|secret)count$/iu;
const IDENTITY_OUTPUT_FIELD_ORDER = ["id", "databaseId", "fullDatabaseId", "number", "url", "resourcePath", "clientMutationId"];

function isSourceSecretName(name) {
  const value = String(name || "");
  return SENSITIVE_NAME.test(value) && !NON_SECRET_SENSITIVE_NAME.test(value);
}

// These reviewed verb families receive the ordinary plan/preview/approval
// write lifecycle without a destructive typed acknowledgement. Every other
// GraphQL mutation fails closed as destructive. Secret-bearing mutations use
// the separately declared environment-only input channel and still require
// typed confirmation at the shared write gate.
const APPROVAL_ONLY_MUTATION_PREFIXES = [
  "accept",
  "add",
  "archive",
  "close",
  "convert",
  "create",
  "disable",
  "enable",
  "link",
  "lock",
  "mark",
  "move",
  "pin",
  "reopen",
  "set",
  "transfer",
  "unarchive",
  "unlink",
  "unlock",
  "unmark",
  "unpin",
  "update",
];

// A repository-to-repository issue transfer changes the target resource even
// though its root name starts with an otherwise approval-only verb. Keep this
// concrete high-impact mutation behind the typed destructive acknowledgement.
const TYPED_CONFIRMATION_MUTATION_NAMES = new Set(["transferIssue"]);

function isPlainObject(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function requireObject(value, label) {
  if (!isPlainObject(value)) throw new Error(label + " must be an object");
  return value;
}

function requireArray(value, label) {
  if (!Array.isArray(value)) throw new Error(label + " must be an array");
  return value;
}

function requireString(value, label) {
  if (typeof value !== "string" || value.trim() === "") throw new Error(label + " must be a non-empty string");
  return value.trim();
}

function sorted(values) {
  return [...values].sort((left, right) => String(left).localeCompare(String(right)));
}

function unique(values) {
  return [...new Set(values)];
}

function uniqueSorted(values) {
  return sorted(unique(values));
}

function kebabCase(value) {
  return requireString(value, "GraphQL root name")
    .replace(/([a-z0-9])([A-Z])/gu, "$1-$2")
    .replace(/([A-Z])([A-Z][a-z])/gu, "$1-$2")
    .replace(/_/gu, "-")
    .toLowerCase();
}

function pascalCase(value) {
  return requireString(value, "GraphQL root name")
    .replace(/[^A-Za-z0-9]+(.)/gu, (_, character) => character.toUpperCase())
    .replace(/^./u, (character) => character.toUpperCase());
}

function variableType(ref) {
  const node = requireObject(ref, "GraphQL type reference");
  let rendered;
  if (node.kind === "named") {
    rendered = requireString(node.name, "GraphQL named type");
  } else if (node.kind === "list") {
    rendered = "[" + variableType(node.of_type) + "]";
  } else {
    throw new Error("GraphQL type reference has unsupported kind " + JSON.stringify(node.kind));
  }
  return rendered + (node.non_null === true ? "!" : "");
}

function namedType(ref) {
  const node = requireObject(ref, "GraphQL type reference");
  if (node.kind === "named") return requireString(node.name, "GraphQL named type");
  if (node.kind === "list") return namedType(node.of_type);
  throw new Error("GraphQL type reference has unsupported kind " + JSON.stringify(node.kind));
}

function typeIndexes(typeSystemCandidate) {
  const typeSystem = requireObject(typeSystemCandidate, "GitHub GraphQL type_system");
  const indexes = {
    inputObjects: new Map(),
    enums: new Map(),
    objects: new Map(),
    interfaces: new Map(),
    unions: new Map(),
    scalars: new Set(requireArray(typeSystem.scalars || [], "GraphQL scalars")),
  };
  for (const [property, target] of [
    ["input_objects", indexes.inputObjects],
    ["enums", indexes.enums],
    ["objects", indexes.objects],
    ["interfaces", indexes.interfaces],
    ["unions", indexes.unions],
  ]) {
    for (const entry of requireArray(typeSystem[property] || [], "GraphQL " + property)) {
      const object = requireObject(entry, "GraphQL " + property + " entry");
      const name = requireString(object.name, "GraphQL " + property + " name");
      if (target.has(name)) throw new Error("GraphQL " + property + " repeats " + name);
      target.set(name, object);
    }
  }
  return indexes;
}

function scalarSchema(name) {
  switch (name) {
    case "Boolean":
      return { type: "boolean" };
    case "Int":
      // GraphQL Int is a signed 32-bit integer, even on a 64-bit Go host.
      // The same exact range is projected into the command flag and checked
      // again by the closed engine schema before any operation reaches I/O.
      return { type: "integer", minimum: -2147483648, maximum: 2147483647 };
    case "Float":
      return { type: "number" };
    default:
      // GitHub custom scalars (DateTime, URI, GitObjectID, and so on) cross
      // the GraphQL wire as strings. Keeping them strings is closed and avoids
      // accidentally accepting caller-owned JSON objects for an opaque scalar.
      return { type: "string" };
  }
}

function inputSchema(ref, indexes, ancestry = [], depth = 0) {
  if (depth > MAX_INPUT_DEPTH) throw new Error("GraphQL input nesting exceeds bounded generator depth");
  const node = requireObject(ref, "GraphQL input type reference");
  if (node.kind === "list") {
    return {
      type: "array",
      maxItems: MAX_GRAPHQL_ARRAY_ITEMS,
      items: inputSchema(node.of_type, indexes, ancestry, depth + 1),
    };
  }
  if (node.kind !== "named") throw new Error("GraphQL input type has unsupported kind " + JSON.stringify(node.kind));
  const name = requireString(node.name, "GraphQL input named type");
  if (BUILTIN_SCALARS.has(name) || indexes.scalars.has(name)) return scalarSchema(name);
  const enumEntry = indexes.enums.get(name);
  if (enumEntry) return { type: "string", enum: sorted(requireArray(enumEntry.values, "GraphQL enum " + name + " values")) };
  const input = indexes.inputObjects.get(name);
  if (!input) throw new Error("GraphQL input type " + name + " is not a declared scalar, enum, or input object");
  if (ancestry.includes(name)) throw new Error("GraphQL input object cycle reaches " + [...ancestry, name].join(" -> "));
  const properties = {};
  const required = [];
  const fields = requireArray(input.fields, "GraphQL input object " + name + " fields");
  for (const field of [...fields].sort((left, right) => String(left.name).localeCompare(String(right.name)))) {
    const fieldName = requireString(field?.name, "GraphQL input field name");
    properties[fieldName] = inputSchema(field.type, indexes, [...ancestry, name], depth + 1);
    if (field.type?.non_null === true) required.push(fieldName);
  }
  const schema = { type: "object", additionalProperties: false, properties };
  if (required.length > 0) schema.required = required;
  return schema;
}

function rootVariablesSchema(field, indexes, { paginated = false } = {}) {
  const properties = {};
  const required = [];
  const sourceArguments = requireArray(field.arguments, "GraphQL root " + field.name + " arguments");
  const argumentsByName = new Map(sourceArguments.map((argument) => [argument.name, argument]));
  for (const argument of sourceArguments) {
    const name = requireString(argument?.name, "GraphQL root argument name");
    properties[name] = inputSchema(argument.type, indexes);
    if (argument.type?.non_null === true) required.push(name);
  }
  if (paginated) {
    const after = argumentsByName.get("after");
    const first = argumentsByName.get("first");
    const before = argumentsByName.get("before");
    const last = argumentsByName.get("last");
    if (!after || !first || !before || !last) throw new Error("GraphQL connection root " + field.name + " must declare forward and backward pagination arguments");
    if (!properties.after) properties.after = inputSchema(after.type, indexes);
    if (!properties.first) properties.first = inputSchema(first.type, indexes);
    if (!properties.before) properties.before = inputSchema(before.type, indexes);
    if (!properties.last) properties.last = inputSchema(last.type, indexes);
  }
  const schema = { type: "object", additionalProperties: false, properties };
  if (required.length > 0) schema.required = sorted(required);
  return schema;
}

function scalarLeafSelection(typeName, indexes) {
  const object = indexes.objects.get(typeName);
  if (!object) return [];
  return requireArray(object.fields || [], "GraphQL object " + typeName + " fields")
    .filter((field) => requireArray(field.arguments || [], "GraphQL field arguments").length === 0)
	.filter((field) => !isSourceSecretName(field.name))
    .filter((field) => {
      const name = namedType(field.type);
      return BUILTIN_SCALARS.has(name) || indexes.scalars.has(name) || indexes.enums.has(name);
    })
    .map((field) => requireString(field.name, "GraphQL output field name"));
}

function concreteTypeSelection(typeName, indexes) {
  const leaves = scalarLeafSelection(typeName, indexes);
  return leaves.length > 0 ? uniqueSorted(["__typename", ...leaves]).join(" ") : "__typename";
}

function abstractTypeSelection(typeName, indexes) {
  const abstract = indexes.interfaces.get(typeName) || indexes.unions.get(typeName);
  if (!abstract) return "";
  const fragments = sorted(requireArray(abstract.possible_types || [], "GraphQL abstract possible_types"))
    .map((concrete) => "... on " + concrete + " { " + concreteTypeSelection(concrete, indexes) + " }");
  return ["__typename", ...fragments].join(" ");
}

function outputFieldOrder(left, right) {
	const leftIndex = IDENTITY_OUTPUT_FIELD_ORDER.indexOf(left);
	const rightIndex = IDENTITY_OUTPUT_FIELD_ORDER.indexOf(right);
	if (leftIndex >= 0 || rightIndex >= 0) {
		if (leftIndex < 0) return 1;
		if (rightIndex < 0) return -1;
		return leftIndex - rightIndex;
	}
	return String(left).localeCompare(String(right));
}

function isScalarOutputField(field, indexes) {
	const name = namedType(field.type);
	return BUILTIN_SCALARS.has(name) || indexes.scalars.has(name) || indexes.enums.has(name);
}

// Mutation payloads are source-derived acknowledgements, not caller-selected
// documents. Select bounded nested identity/status information so callers can
// persist a provider result, while keeping selection depth and total field
// count finite and excluding source-classified secret outputs.
function boundedMutationObjectSelection(typeName, indexes, budget, ancestry = [], depth = 0) {
	const object = indexes.objects.get(typeName);
	if (!object) return "";
	if (depth > MAX_MUTATION_OUTPUT_DEPTH || ancestry.includes(typeName) || budget.remaining <= 0) return "__typename";
	const fields = requireArray(object.fields || [], "GraphQL object " + typeName + " fields")
		.filter((field) => requireArray(field.arguments || [], "GraphQL field arguments").length === 0)
		.filter((field) => !isSourceSecretName(field.name));
	const scalar = fields
		.filter((field) => isScalarOutputField(field, indexes))
		.map((field) => requireString(field.name, "GraphQL output field name"))
		.sort(outputFieldOrder);
	const selected = ["__typename"];
	for (const name of scalar) {
		if (budget.remaining <= 0) break;
		selected.push(name);
		budget.remaining--;
	}
	if (depth === MAX_MUTATION_OUTPUT_DEPTH || budget.remaining <= 0) return selected.join(" ");
	const nested = fields
		.filter((field) => !isScalarOutputField(field, indexes))
		.sort((left, right) => outputFieldOrder(String(left.name), String(right.name)));
	for (const field of nested) {
		if (budget.remaining <= 0) break;
		// The parent field is part of the declared selection too; reserve its
		// one slot before recursing so descendants cannot exceed the ceiling.
		budget.remaining--;
		const nestedName = namedType(field.type);
		let child = boundedMutationObjectSelection(nestedName, indexes, budget, [...ancestry, typeName], depth + 1);
		if (child === "") {
			child = boundedMutationAbstractSelection(nestedName, indexes, budget, [...ancestry, typeName], depth + 1);
		}
		if (child === "") {
			budget.remaining++;
			continue;
		}
		selected.push(requireString(field.name, "GraphQL output field name") + " { " + child + " }");
	}
	return selected.join(" ");
}

function boundedMutationAbstractSelection(typeName, indexes, budget, ancestry, depth) {
	const abstract = indexes.interfaces.get(typeName) || indexes.unions.get(typeName);
	if (!abstract || depth > MAX_MUTATION_OUTPUT_DEPTH || budget.remaining <= 0) return "";
	const fragments = [];
	for (const concrete of sorted(requireArray(abstract.possible_types || [], "GraphQL abstract possible_types"))) {
		if (budget.remaining <= 0) break;
		const child = boundedMutationObjectSelection(concrete, indexes, budget, ancestry, depth);
		if (child !== "") fragments.push("... on " + concrete + " { " + child + " }");
	}
	return fragments.length > 0 ? ["__typename", ...fragments].join(" ") : "";
}

function boundedMutationSelection(typeName, indexes) {
	const budget = { remaining: MAX_MUTATION_OUTPUT_FIELDS };
	return boundedMutationObjectSelection(typeName, indexes, budget) || boundedMutationAbstractSelection(typeName, indexes, budget, [], 0);
}

function outputSelection(field, indexes, { paginated = false } = {}) {
  if (field.root === "Query" && field.name === "rateLimit") return "limit cost remaining resetAt";
  const name = namedType(field.return_type);
  if (field.root === "Mutation") return boundedMutationSelection(name, indexes);
  const abstractSelection = abstractTypeSelection(name, indexes);
  if (abstractSelection !== "") return abstractSelection;
  if (paginated) {
    const connection = requireObject(indexes.objects.get(name), "GraphQL connection " + name);
    const fields = new Map(requireArray(connection.fields || [], "GraphQL connection fields").map((entry) => [entry.name, entry]));
    const nodes = requireObject(fields.get("nodes"), "GraphQL connection nodes field");
    const nodeType = namedType(nodes.type);
    const nodeSelection = abstractTypeSelection(nodeType, indexes) || concreteTypeSelection(nodeType, indexes);
    const connectionScalars = scalarLeafSelection(name, indexes).filter((fieldName) => fieldName !== "pageInfo");
    return uniqueSorted(["__typename", ...connectionScalars]).join(" ") + " nodes { " + nodeSelection + " } pageInfo { hasNextPage hasPreviousPage startCursor endCursor }";
  }
  if (indexes.objects.has(name)) return concreteTypeSelection(name, indexes);
  return "";
}

function isConnectionRoot(field, indexes) {
  if (field.root !== "Query") return false;
  const args = new Set(requireArray(field.arguments, "GraphQL root arguments").map((argument) => argument.name));
  if (!args.has("after") || !args.has("before") || !args.has("first") || !args.has("last")) return false;
  const returnObject = indexes.objects.get(namedType(field.return_type));
  if (!returnObject) return false;
  const fields = new Map(requireArray(returnObject.fields || [], "GraphQL return fields").map((entry) => [entry.name, entry]));
  return fields.has("nodes") && fields.has("pageInfo");
}

function documentFor(field, indexes, { paginated }) {
  const rootArguments = requireArray(field.arguments, "GraphQL root " + field.name + " arguments");
  const declarations = rootArguments.map((argument) => "$" + argument.name + ": " + variableType(argument.type));
  const invocation = rootArguments.length === 0
    ? field.name
    : field.name + "(" + rootArguments.map((argument) => argument.name + ": $" + argument.name).join(", ") + ")";
  const selection = outputSelection(field, indexes, { paginated });
  const rootSelection = selection === "" ? invocation : invocation + " { " + selection + " }";
  const rateLimit = field.root === "Query" && field.name !== "rateLimit"
    ? " rateLimit { limit cost remaining resetAt }"
    : "";
  const kind = field.root === "Query" ? "query" : "mutation";
  const operationName = "GitHub" + (kind === "query" ? "Query" : "Mutation") + pascalCase(field.name);
  const variableDeclaration = declarations.length === 0 ? "" : "(" + declarations.join(", ") + ")";
  return {
    operationName,
    document: kind + " " + operationName + variableDeclaration + " { " + rootSelection + rateLimit + " }",
  };
}

function containsSensitiveInput(ref, indexes, seen = new Set()) {
	const node = requireObject(ref, "GraphQL input type reference");
	if (node.kind === "list") return containsSensitiveInput(node.of_type, indexes, seen);
	if (node.kind !== "named") return false;
	const name = requireString(node.name, "GraphQL input named type");
	const input = indexes.inputObjects.get(name);
	// A GraphQL input *type* called RegenerateVerifiableDomainTokenInput is not
	// itself a secret input. Classify the actual recursively declared field
	// names instead, so generated tokens retain their ordinary destructive
	// write contract while access-token-bearing inputs get the env-only path.
	if (!input || seen.has(name)) return false;
	return requireArray(input.fields || [], "GraphQL input object fields").some((field) =>
		isSourceSecretName(field.name) || containsSensitiveInput(field.type, indexes, new Set([...seen, name])),
	);
}

function sourceSensitiveArguments(field, indexes) {
	return requireArray(field.arguments, "GraphQL root arguments")
		.filter((argument) => isSourceSecretName(argument.name) || containsSensitiveInput(argument.type, indexes));
}

function mutationPolicy(field, indexes) {
	const sensitiveArguments = sourceSensitiveArguments(field, indexes);
	if (sensitiveArguments.length > 0) {
		return {
			availability: "implemented",
			mutationClass: "secret",
			secretSensitive: true,
			sensitiveArguments: sensitiveArguments.map((argument) => requireString(argument.name, "GraphQL sensitive argument name")),
			risk: "high",
			approval: "plan, preview, approval, execute (typed destructive confirmation; input via --from-env)",
			notes: "The source-derived input contains a secret field. Supply its complete typed JSON value only through --from-env input=ENV; it is withheld from persisted plans.",
		};
	}
	if (TYPED_CONFIRMATION_MUTATION_NAMES.has(field.name)) {
		return {
			availability: "implemented",
			mutationClass: "destructive",
			destructive: true,
			risk: "critical",
			approval: "plan, preview, approval, execute (typed destructive confirmation)",
		};
	}
	const approvalOnly = APPROVAL_ONLY_MUTATION_PREFIXES.some((prefix) => field.name.startsWith(prefix));
  return approvalOnly
    ? {
        availability: "implemented",
        mutationClass: field.name.startsWith("create") ? "create" : field.name.startsWith("update") ? "update" : "admin",
        destructive: false,
        risk: "high",
        approval: "plan, preview, approval, execute",
      }
    : {
        availability: "implemented",
        mutationClass: "destructive",
        destructive: true,
        risk: "critical",
        approval: "plan, preview, approval, execute (typed destructive confirmation)",
      };
}

function commandFlagForArgument(argument, indexes, { paginated = false } = {}) {
  const name = requireString(argument?.name, "GraphQL root argument name");
  if (paginated && (name === "after" || name === "before")) return undefined;
  const base = { name: kebabCase(name), maps_to: "body." + name };
  const type = requireObject(argument.type, "GraphQL root argument type");
  const named = namedType(type);
  if (type.kind === "list" || indexes.inputObjects.has(named)) {
    return { ...base, type: "json", ...(type.non_null === true ? { required: true } : {}) };
  }
  const enumEntry = indexes.enums.get(named);
  if (enumEntry) {
    return {
      ...base,
      type: "enum",
      values: sorted(requireArray(enumEntry.values, "GraphQL enum values")),
      ...(type.non_null === true ? { required: true } : {}),
    };
  }
	let flagType = "string";
	if (named === "Boolean") flagType = "boolean";
	if (named === "Int") return { ...base, type: "integer", minimum: -2147483648, maximum: 2147483647, ...(type.non_null === true ? { required: true } : {}) };
	if (named === "Float") flagType = "number";
  return { ...base, type: flagType, ...(type.non_null === true ? { required: true } : {}) };
}

function generatedOperation(field, indexes) {
  const isQuery = field.root === "Query";
  const paginated = isConnectionRoot(field, indexes);
  const document = documentFor(field, indexes, { paginated });
  const operation = {
    id: "github.graphql." + (isQuery ? "query." : "mutation.") + kebabCase(field.name),
    kind: isQuery ? "graphql_query" : "graphql_mutation",
    summary: (isQuery ? "Query" : "Mutate") + " GitHub GraphQL root " + field.name + " using a fixed typed document.",
    source_url: requireString(field.source_url || "", "GraphQL source URL"),
    risk: isQuery ? "low" : mutationPolicy(field, indexes).risk,
    approval: isQuery ? "none" : mutationPolicy(field, indexes).approval,
    output_policy: "json_redacted",
    graphql: {
      operation_name: document.operationName,
      document: document.document,
      path: GRAPHQL_TRANSPORT.path,
      max_bytes: MAX_GRAPHQL_BYTES,
      variables_schema: rootVariablesSchema(field, indexes, { paginated }),
    },
  };
	if (paginated) {
    operation.graphql.pagination = {
      connection_path: field.name,
      cursor_variable: "after",
      page_size_variable: "first",
      backward_cursor_variable: "before",
      backward_page_size_variable: "last",
      max_page_size: MAX_GRAPHQL_ARRAY_ITEMS,
    };
  }
	if (!isQuery) {
		const policy = mutationPolicy(field, indexes);
		operation.mutation_class = policy.mutationClass;
		if (policy.secretSensitive) {
			operation.secret_sensitive = true;
			operation.sensitive_policy = {
				input_mode: "env",
				redact_fields: policy.sensitiveArguments.map((argument) => "body." + argument),
				transform: "none",
				approval_mode: "typed_confirmation",
			};
		}
		if (policy.destructive) operation.destructive = true;
	}
	if (isQuery) {
		const sensitiveArguments = sourceSensitiveArguments(field, indexes);
		if (sensitiveArguments.length > 0) {
			operation.sensitive_policy = {
				input_mode: "env",
				redact_fields: sensitiveArguments.map((argument) => "body." + argument.name),
				transform: "none",
			};
		}
	}
  return operation;
}

function generatedCommand(field, indexes, operation) {
  const isQuery = field.root === "Query";
  const paginated = isConnectionRoot(field, indexes);
	const querySensitiveArguments = isQuery
		? sourceSensitiveArguments(field, indexes).map((argument) => requireString(argument.name, "GraphQL sensitive argument name"))
		: [];
	const policy = isQuery
		? { availability: "implemented", sensitiveArguments: querySensitiveArguments, secretSensitive: querySensitiveArguments.length > 0 }
		: mutationPolicy(field, indexes);
	const flags = requireArray(field.arguments, "GraphQL root arguments")
		.map((argument) => {
			const flag = commandFlagForArgument(argument, indexes, { paginated });
			if (flag && policy.secretSensitive && policy.sensitiveArguments.includes(argument.name)) flag.env_only = true;
			return flag;
		})
    .filter(Boolean);
  const command = {
    path: "graphql " + (isQuery ? "query " : "mutation ") + kebabCase(field.name),
    summary: (isQuery ? "Run fixed GitHub GraphQL query " : "Prepare fixed GitHub GraphQL mutation ") + field.name + ".",
    intent: isQuery ? "direct_read" : "direct_write",
    availability: policy.availability,
    operation: operation.id,
    source_url: requireString(field.source_url || "", "GraphQL source URL"),
    flags,
    api_surface: [{ ...GRAPHQL_TRANSPORT }],
    output_policy: "json_redacted",
	};
	if (paginated) {
		command.constraints = [{
			kind: "exactly_one",
			fields: ["body.first", "body.last"],
			message: "exactly one GraphQL pagination direction must be provided",
		}];
	}
  if (!isQuery) {
    command.risk = mutationPolicy(field, indexes).risk;
    command.approval = mutationPolicy(field, indexes).approval;
    if (policy.notes) command.notes = policy.notes;
  }
  return command;
}

function sourceFields(lock) {
  const graphql = requireObject(lock?.graphql, "GitHub GraphQL source lock graphql");
  const queries = requireArray(graphql.query_fields, "GitHub GraphQL query_fields");
  const mutations = requireArray(graphql.mutation_fields, "GitHub GraphQL mutation_fields");
  const roots = [...queries, ...mutations].map((field) => ({ ...requireObject(field, "GitHub GraphQL root"), source_url: graphql.source_url }));
  const seen = new Set();
  for (const field of roots) {
    const identity = requireString(field.root, "GraphQL root kind") + "." + requireString(field.name, "GraphQL root name");
    if (field.root !== "Query" && field.root !== "Mutation") throw new Error("GraphQL root " + identity + " has unsupported root");
    if (seen.has(identity)) throw new Error("GitHub source lock duplicates GraphQL root " + identity);
    seen.add(identity);
    requireArray(field.arguments, "GraphQL root " + identity + " arguments");
  }
  const canary = mutations.find((field) => field.name === "createEnterpriseOrganization");
  if (!canary) throw new Error("GitHub GraphQL source lock is missing mandatory createEnterpriseOrganization mutation canary");
  return roots;
}

function assertSourceLock(lock) {
  const candidate = requireObject(lock, "GitHub source lock");
  if (candidate.schema_version !== 2 || candidate.connector !== "github") {
    throw new Error("GitHub GraphQL parity generator requires source-lock schema_version 2 for connector github");
  }
  const fields = sourceFields(candidate);
  const counts = requireObject(candidate.counts, "GitHub source lock counts");
  const queryCount = fields.filter((field) => field.root === "Query").length;
  const mutationCount = fields.filter((field) => field.root === "Mutation").length;
  if (counts.graphql_query !== queryCount || counts.graphql_mutation !== mutationCount) {
    throw new Error("GitHub GraphQL source lock count drift: roots do not match counts");
  }
  return { lock: candidate, fields, indexes: typeIndexes(candidate.graphql.type_system) };
}

function operationIDSet(operations, label) {
  const seen = new Set();
  for (const operation of operations) {
    const id = requireString(operation?.id, label + " operation id");
    if (seen.has(id)) throw new Error(label + " has duplicate operation id " + id);
    seen.add(id);
  }
  return seen;
}

function commandPathSet(commands, label) {
  const seen = new Set();
  for (const command of commands) {
    const commandPath = requireString(command?.path, label + " command path");
    if (seen.has(commandPath)) throw new Error(label + " has duplicate command path " + commandPath);
    seen.add(commandPath);
  }
  return seen;
}

function validateClosedBoundedSchema(schema, label) {
  const node = requireObject(schema, label);
  if (node.type === "array") {
    if (!Number.isInteger(node.maxItems) || node.maxItems <= 0) throw new Error(label + " array requires positive maxItems");
    validateClosedBoundedSchema(node.items, label + "/items");
  }
  if (node.type !== "object") return;
  if (node.additionalProperties !== false) throw new Error(label + " object requires additionalProperties: false");
  const properties = requireObject(node.properties, label + " properties");
  for (const [name, child] of Object.entries(properties)) validateClosedBoundedSchema(child, label + "/" + name);
}

function verifyFixedDocument(operation, field) {
  const graphql = requireObject(operation.graphql, "generated GraphQL operation");
  const document = requireString(graphql.document, "generated GraphQL document");
  const kind = field.root === "Query" ? "query" : "mutation";
  if (!document.startsWith(kind + " ")) throw new Error(operation.id + " fixed document must start with " + kind);
  if (document.includes("callerSelection") || document.includes("callerDocument") || document.includes("$selection")) {
    throw new Error(operation.id + " exposes caller-controlled GraphQL selection or document");
  }
  if (!document.includes(field.name + (field.arguments.length === 0 ? "" : "("))) {
    throw new Error(operation.id + " fixed document does not invoke source root " + field.name);
  }
  if (graphql.path !== GRAPHQL_TRANSPORT.path || graphql.max_bytes !== MAX_GRAPHQL_BYTES) {
    throw new Error(operation.id + " does not use the bounded fixed GraphQL transport");
  }
  validateClosedBoundedSchema(graphql.variables_schema, operation.id + " variables_schema");
}

/**
 * Build the generated source-derived artifact fragments. This pure function is
 * used by both hermetic tests and the --write/--check CLI below.
 */
export function buildGitHubGraphQLParityArtifacts({ lock, bundle }) {
  requireObject(bundle, "GitHub bundle");
  const source = assertSourceLock(lock);
  const operations = source.fields.map((field) => generatedOperation(field, source.indexes));
  const commands = source.fields.map((field, index) => generatedCommand(field, source.indexes, operations[index]));
  const transport = {
    ...GRAPHQL_TRANSPORT,
    covered_by: { operations: operations.map((operation) => operation.id) },
  };
  const generated = { operations, commands, transport };
  validateGitHubGraphQLParityArtifacts({ lock, generated });
  return generated;
}

/**
 * Fail closed before writing an incomplete or widened generated GraphQL
 * catalog. The checks are deliberately structural and source-derived; they do
 * not treat command reachability or old live proof as source coverage.
 */
export function validateGitHubGraphQLParityArtifacts({ lock, generated }) {
  const source = assertSourceLock(lock);
  const artifact = requireObject(generated, "generated GitHub GraphQL parity artifacts");
  const operations = requireArray(artifact.operations, "generated GraphQL operations");
  const commands = requireArray(artifact.commands, "generated GraphQL commands");
  const transport = requireObject(artifact.transport, "generated GraphQL transport");
  const operationIDs = operationIDSet(operations, "generated GraphQL artifacts");
  commandPathSet(commands, "generated GraphQL artifacts");
  if (operations.length !== source.fields.length || commands.length !== source.fields.length) {
    throw new Error("generated GraphQL operation and command counts must equal source root count");
  }
  if (transport.method !== GRAPHQL_TRANSPORT.method || transport.path !== GRAPHQL_TRANSPORT.path) {
    throw new Error("generated GraphQL artifacts must use POST /graphql transport");
  }
  const covered = requireArray(transport.covered_by?.operations, "generated GraphQL transport covered_by.operations");
  if (covered.length !== operations.length || new Set(covered).size !== covered.length || covered.some((id) => !operationIDs.has(id))) {
    throw new Error("generated GraphQL transport must cover every fixed operation exactly once");
  }

  for (const field of source.fields) {
    const prefix = field.root === "Query" ? "query" : "mutation";
    const id = "github.graphql." + prefix + "." + kebabCase(field.name);
    const operation = operations.find((candidate) => candidate.id === id);
    if (!operation) throw new Error("generated GraphQL operation missing source root " + field.root + "." + field.name);
    if (operation.kind !== "graphql_" + prefix) throw new Error(id + " has wrong GraphQL operation kind");
    verifyFixedDocument(operation, field);
    const expectedPath = "graphql " + prefix + " " + kebabCase(field.name);
    const command = commands.find((candidate) => candidate.path === expectedPath);
    if (!command || command.operation !== id) throw new Error(id + " has no generated command binding");
    if (command.intent !== (prefix === "query" ? "direct_read" : "direct_write")) {
      throw new Error(id + " generated command has wrong intent");
    }
    if (JSON.stringify(command.api_surface) !== JSON.stringify([{ ...GRAPHQL_TRANSPORT }])) {
      throw new Error(id + " generated command must bind exactly POST /graphql");
    }
  }

  const canary = operations.find((operation) => operation.id === "github.graphql.mutation.create-enterprise-organization");
  if (!canary) throw new Error("generated GraphQL artifacts are missing createEnterpriseOrganization canary");
  const canaryInput = canary.graphql?.variables_schema?.properties?.input;
  if (!isPlainObject(canaryInput) || canary.graphql?.variables_schema?.required?.includes("input") !== true) {
    throw new Error("createEnterpriseOrganization canary has no required typed input");
  }
	const deleteIssue = source.fields.find((field) => field.root === "Mutation" && field.name === "deleteIssue");
	if (deleteIssue) {
		const command = commands.find((candidate) => candidate.path === "graphql mutation delete-issue");
		if (command?.availability !== "implemented" || command?.approval !== "plan, preview, approval, execute (typed destructive confirmation)") {
			throw new Error("deleteIssue must remain an implemented typed-confirmation destructive mutation");
		}
	}
	for (const field of source.fields.filter((candidate) => candidate.root === "Mutation" && candidate.name !== "deleteIssue")) {
		const policy = mutationPolicy(field, source.indexes);
		if (!policy.secretSensitive) continue;
		const command = commands.find((candidate) => candidate.path === "graphql mutation " + kebabCase(field.name));
		const operation = operations.find((candidate) => candidate.id === "github.graphql.mutation." + kebabCase(field.name));
		if (!command || !operation || command.availability !== "implemented") {
			throw new Error(field.name + " secret mutation must generate an implemented command and operation");
		}
		for (const argument of policy.sensitiveArguments) {
			if (!command.flags.some((flag) => flag.maps_to === "body." + argument && flag.type === "json" && flag.required === true && flag.env_only === true)) {
				throw new Error(field.name + " secret mutation must expose required env_only JSON flag for " + argument);
			}
			if (!operation.sensitive_policy?.redact_fields?.includes("body." + argument)) {
				throw new Error(field.name + " secret mutation must redact body." + argument);
			}
		}
	}
}

export function mergeGitHubGraphQLParityArtifacts(bundle, generated) {
  const operations = requireObject(bundle.operations, "GitHub operations artifact");
  const cli = requireObject(bundle.cli, "GitHub CLI artifact");
  const surface = requireObject(bundle.surface, "GitHub API surface artifact");
  const generatedOperation = (operation) => /^github\.graphql\.(?:query|mutation)\./u.test(String(operation?.id || ""));
  const generatedCommand = (command) => /^graphql (?:query|mutation) /u.test(String(command?.path || ""));
  const generatedTransport = (endpoint) =>
    endpoint?.method === GRAPHQL_TRANSPORT.method &&
    endpoint?.path === GRAPHQL_TRANSPORT.path &&
    Array.isArray(endpoint?.covered_by?.operations);
  const supplementalGraphQLOperations = new Set(
    requireArray(operations.operations, "GitHub operations")
      .filter((operation) => !generatedOperation(operation) && /^graphql_(?:query|mutation)$/u.test(String(operation?.kind || "")))
      .map((operation) => operation.id),
  );
  const supplementalTransportOperations = requireArray(surface.endpoints, "GitHub API endpoints")
    .filter(generatedTransport)
    .flatMap((endpoint) => endpoint.covered_by.operations)
    .filter((id) => supplementalGraphQLOperations.has(id));
  const transport = {
    ...generated.transport,
    covered_by: {
      operations: uniqueSorted([...supplementalTransportOperations, ...generated.transport.covered_by.operations]),
    },
  };
  const by = (key) => (left, right) => String(key(left)).localeCompare(String(key(right)));
  const operationOrder = by((operation) => operation?.id || "");
  const commandOrder = by((command) => `${command?.path || ""}\u0000${command?.operation || command?.write || ""}`);
  const endpointOrder = by((endpoint) => `${String(endpoint?.method || "").toUpperCase()}\u0000${endpoint?.path || ""}\u0000${JSON.stringify(endpoint?.covered_by || {})}`);
  return {
    operations: {
      ...operations,
      operations: [
        ...requireArray(operations.operations, "GitHub operations").filter(
          (operation) => !generatedOperation(operation) && !OBSOLETE_GRAPHQL_OPERATION_IDS.has(operation?.id),
        ),
        ...generated.operations,
      ].sort(operationOrder),
    },
    cli: {
      ...cli,
      commands: [...requireArray(cli.commands, "GitHub CLI commands").filter((command) => !generatedCommand(command)), ...generated.commands].sort(commandOrder),
    },
    surface: {
      ...surface,
      endpoints: [...requireArray(surface.endpoints, "GitHub API endpoints").filter((endpoint) => !generatedTransport(endpoint)), transport].sort(endpointOrder),
    },
  };
}

function renderJSON(value) {
  return JSON.stringify(value, null, 2) + "\n";
}

async function readJSON(filename) {
  return JSON.parse(await readFile(filename, "utf8"));
}

async function loadDefaultBundle() {
  const [lock, operations, cli, surface] = await Promise.all([
    readJSON(DEFAULT_LOCK),
    readJSON(DEFAULT_OPERATIONS),
    readJSON(DEFAULT_CLI),
    readJSON(DEFAULT_SURFACE),
  ]);
  return { lock, bundle: { operations, cli, surface } };
}

async function generatedFiles() {
  const { lock, bundle } = await loadDefaultBundle();
  const generated = buildGitHubGraphQLParityArtifacts({ lock, bundle });
  return mergeGitHubGraphQLParityArtifacts(bundle, generated);
}

async function checkGeneratedFiles() {
  const expected = await generatedFiles();
  const [operations, cli, surface] = await Promise.all([
    readFile(DEFAULT_OPERATIONS, "utf8"),
    readFile(DEFAULT_CLI, "utf8"),
    readFile(DEFAULT_SURFACE, "utf8"),
  ]);
  return operations === renderJSON(expected.operations) &&
    cli === renderJSON(expected.cli) &&
    surface === renderJSON(expected.surface);
}

async function writeGeneratedFiles() {
  const expected = await generatedFiles();
  await Promise.all([
    writeFile(DEFAULT_OPERATIONS, renderJSON(expected.operations)),
    writeFile(DEFAULT_CLI, renderJSON(expected.cli)),
    writeFile(DEFAULT_SURFACE, renderJSON(expected.surface)),
  ]);
}

async function main(argv) {
  const args = new Set(argv);
  if (args.has("--help") || args.has("-h")) {
    process.stdout.write("usage: node scripts/gen-github-graphql-parity.mjs [--check|--write]\n");
    return;
  }
  if (args.size !== 1 || (!args.has("--check") && !args.has("--write"))) {
    throw new Error("expected exactly one of --check or --write");
  }
  if (args.has("--check")) {
    if (!(await checkGeneratedFiles())) throw new Error("GitHub GraphQL generated artifacts are stale; run scripts/gen-github-graphql-parity.mjs --write");
    process.stdout.write("github graphql parity artifacts: ok\n");
    return;
  }
  await writeGeneratedFiles();
  process.stdout.write("github graphql parity artifacts: generated\n");
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main(process.argv.slice(2)).catch((error) => {
    process.stderr.write("gen-github-graphql-parity: " + (error?.stack || error) + "\n");
    process.exitCode = 1;
  });
}
