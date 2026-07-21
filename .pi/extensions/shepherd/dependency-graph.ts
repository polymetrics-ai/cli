import type { RepositoryBlocker } from "./autonomy-policy.ts";

const MAX_WORK_ITEMS = 64;
const MAX_DEPENDENCIES = 64;
const MAX_SCOPES = 64;
const MAX_TEXT_LENGTH = 512;
const MAX_CONFLICT_COMPONENT_SIZE = 12;
const MAX_SCOPE_ALIASES = 16;

export type WorkItemStatus = "pending" | "running" | "succeeded" | "failed" | "blocked";
export type WorkAccess = "read_only" | "mutating";

export interface DependencyWorkItem {
	id: string;
	dependsOn: string[];
	status: WorkItemStatus;
	access: WorkAccess;
	writeScopes: string[];
}

export type DependencyGraphErrorCode =
	| "invalid_item"
	| "duplicate_id"
	| "unknown_dependency"
	| "cycle"
	| "ambiguous_scope"
	| "incoherent_status"
	| "conflict_component_too_large";

export class DependencyGraphError extends Error {
	readonly code: DependencyGraphErrorCode;

	constructor(code: DependencyGraphErrorCode, message: string) {
		super(message);
		this.name = "DependencyGraphError";
		this.code = code;
	}
}

export interface ValidatedDependencyGraph {
	items: DependencyWorkItem[];
	topologicalOrder: string[];
}

export interface ReadyQueueOptions {
	maxConcurrency: number;
	allowMutating?: boolean;
}

export type ReadyQueueSelection =
	| { kind: "selected"; itemIds: string[] }
	| { kind: "blocked"; blocker: Extract<RepositoryBlocker, "not_spawned_dependency_blocked" | "not_spawned_write_scope_collision"> }
	| { kind: "at_capacity" }
	| { kind: "complete" };

function validText(value: unknown): value is string {
	return typeof value === "string"
		&& value.length > 0
		&& value.length <= MAX_TEXT_LENGTH
		&& value.trim() === value
		&& !/[\u0000-\u001f\u007f-\u009f]/.test(value);
}

function isExactRecord(value: unknown, keys: readonly string[]): value is Record<string, unknown> {
	if (typeof value !== "object" || value === null || Array.isArray(value)) return false;
	const actual = Object.keys(value);
	return actual.length === keys.length && actual.every((key) => keys.includes(key));
}

function compareText(left: string, right: string): number {
	return left < right ? -1 : left > right ? 1 : 0;
}

function scopeAliases(scope: string): string[] {
	// A bounded conservative alias closure for filesystem/Git collision safety. This intentionally
	// does not claim Unicode Default Caseless Matching or complete filesystem equivalence.
	const aliases = new Set<string>();
	const queue = [scope.normalize("NFKC")];
	while (queue.length > 0 && aliases.size < MAX_SCOPE_ALIASES) {
		const current = queue.shift();
		if (current === undefined) break;
		const normalized = current.normalize("NFKC").normalize("NFC");
		if (aliases.has(normalized)) continue;
		aliases.add(normalized);
		queue.push(normalized.toLowerCase(), normalized.toUpperCase());
	}
	return [...aliases].sort(compareText);
}

function validateScope(scope: unknown): asserts scope is string {
	if (!validText(scope)
		|| scope === "."
		|| scope.startsWith("/")
		|| scope.endsWith("/")
		|| scope.includes("\\")
		|| /[*?\[\]{}]/.test(scope)) {
		throw new DependencyGraphError("ambiguous_scope", "ambiguous write scope");
	}
	const segments = scope.split("/");
	if (segments.some((segment) => segment === "" || segment === "." || segment === ".." || /^[A-Za-z]:$/.test(segment))) {
		throw new DependencyGraphError("ambiguous_scope", "ambiguous write scope");
	}
}

function scopeContainsCanonical(left: string, right: string): boolean {
	return left === right || right.startsWith(`${left}/`);
}

function canonicalScopesCollide(left: readonly (readonly string[])[], right: readonly (readonly string[])[]): boolean {
	return left.some((leftAliases) => right.some((rightAliases) =>
		leftAliases.some((leftScope) => rightAliases.some((rightScope) =>
			scopeContainsCanonical(leftScope, rightScope) || scopeContainsCanonical(rightScope, leftScope),
		)),
	));
}

function scopeContains(left: string, right: string): boolean {
	return canonicalScopesCollide([scopeAliases(left)], [scopeAliases(right)]);
}

function compareIds(left: Pick<DependencyWorkItem, "id">, right: Pick<DependencyWorkItem, "id">): number {
	return compareText(left.id, right.id);
}

export function scopesCollide(left: readonly string[], right: readonly string[]): boolean {
	for (const scope of left) validateScope(scope);
	for (const scope of right) validateScope(scope);
	return canonicalScopesCollide(left.map(scopeAliases), right.map(scopeAliases));
}

export function validateDependencyGraph(input: unknown): ValidatedDependencyGraph {
	if (!Array.isArray(input) || input.length > MAX_WORK_ITEMS) {
		throw new DependencyGraphError("invalid_item", `dependency graph must contain at most ${MAX_WORK_ITEMS} items`);
	}
	const ids = new Set<string>();
	const canonical: DependencyWorkItem[] = [];
	for (const value of input) {
		if (!isExactRecord(value, ["id", "dependsOn", "status", "access", "writeScopes"])) {
			throw new DependencyGraphError("invalid_item", "work item must have a bounded canonical id");
		}
		const candidate = value;
		const id = candidate.id;
		if (!validText(id)) throw new DependencyGraphError("invalid_item", "work item must have a bounded canonical id");
		if (ids.has(id)) throw new DependencyGraphError("duplicate_id", `duplicate work item ${id}`);
		ids.add(id);
		const rawDependencies = candidate.dependsOn;
		if (!Array.isArray(rawDependencies) || rawDependencies.length > MAX_DEPENDENCIES
			|| !rawDependencies.every(validText)) {
			throw new DependencyGraphError("invalid_item", `invalid dependencies for ${id}`);
		}
		if (candidate.status !== "pending" && candidate.status !== "running" && candidate.status !== "succeeded"
			&& candidate.status !== "failed" && candidate.status !== "blocked") {
			throw new DependencyGraphError("invalid_item", `invalid status for ${id}`);
		}
		if (candidate.access !== "read_only" && candidate.access !== "mutating") {
			throw new DependencyGraphError("invalid_item", `invalid access for ${id}`);
		}
		const rawScopes = candidate.writeScopes;
		if (!Array.isArray(rawScopes) || rawScopes.length > MAX_SCOPES || !rawScopes.every(validText)) {
			throw new DependencyGraphError("ambiguous_scope", `invalid write scopes for ${id}`);
		}
		if (candidate.access === "read_only" && rawScopes.length !== 0) {
			throw new DependencyGraphError("ambiguous_scope", `read-only item ${id} declares a write scope`);
		}
		if (candidate.access === "mutating" && rawScopes.length === 0) {
			throw new DependencyGraphError("ambiguous_scope", `mutating item ${id} has no write scope`);
		}
		for (const scope of rawScopes) validateScope(scope);
		const dependencies = [...rawDependencies].sort(compareText);
		if (new Set(dependencies).size !== dependencies.length) {
			throw new DependencyGraphError("invalid_item", `duplicate dependency for ${id}`);
		}
		const scopes = [...rawScopes].sort(compareText);
		if (new Set(scopes).size !== scopes.length) {
			throw new DependencyGraphError("ambiguous_scope", `duplicate write scope for ${id}`);
		}
		for (let index = 0; index < scopes.length; index += 1) {
			for (let other = index + 1; other < scopes.length; other += 1) {
				if (scopeContains(scopes[index], scopes[other]) || scopeContains(scopes[other], scopes[index])) {
					throw new DependencyGraphError("ambiguous_scope", `redundant write scopes for ${id}`);
				}
			}
		}
		canonical.push({
			id,
			dependsOn: dependencies,
			status: candidate.status,
			access: candidate.access,
			writeScopes: scopes,
		});
	}

	canonical.sort(compareIds);
	const byId = new Map(canonical.map((candidate) => [candidate.id, candidate]));
	for (const candidate of canonical) {
		for (const dependency of candidate.dependsOn) {
			if (dependency === candidate.id) throw new DependencyGraphError("cycle", `${candidate.id} depends on itself`);
			if (!byId.has(dependency)) {
				throw new DependencyGraphError("unknown_dependency", `${candidate.id} depends on unknown item ${dependency}`);
			}
		}
	}

	const indegree = new Map(canonical.map((candidate) => [candidate.id, candidate.dependsOn.length]));
	const dependents = new Map(canonical.map((candidate) => [candidate.id, [] as string[]]));
	for (const candidate of canonical) {
		for (const dependency of candidate.dependsOn) dependents.get(dependency)?.push(candidate.id);
	}
	for (const values of dependents.values()) values.sort(compareText);
	const ready = canonical.filter((candidate) => candidate.dependsOn.length === 0).map((candidate) => candidate.id);
	const topologicalOrder: string[] = [];
	while (ready.length > 0) {
		const id = ready.shift();
		if (id === undefined) break;
		topologicalOrder.push(id);
		for (const dependent of dependents.get(id) ?? []) {
			const remaining = (indegree.get(dependent) ?? 0) - 1;
			indegree.set(dependent, remaining);
			if (remaining === 0) {
				ready.push(dependent);
				ready.sort(compareText);
			}
		}
	}
	if (topologicalOrder.length !== canonical.length) {
		throw new DependencyGraphError("cycle", "dependency graph contains a cycle");
	}
	for (const candidate of canonical) {
		if ((candidate.status === "running" || candidate.status === "succeeded" || candidate.status === "failed")
			&& candidate.dependsOn.some((dependency) => byId.get(dependency)?.status !== "succeeded")) {
			throw new DependencyGraphError(
				"incoherent_status",
				`${candidate.status} item ${candidate.id} has an unsatisfied dependency`,
			);
		}
	}
	return { items: canonical, topologicalOrder };
}

function itemsCollide(left: DependencyWorkItem, right: DependencyWorkItem): boolean {
	return left.access === "mutating"
		&& right.access === "mutating"
		&& scopesCollide(left.writeScopes, right.writeScopes);
}

function lexicographicallyEarlier(left: readonly DependencyWorkItem[], right: readonly DependencyWorkItem[]): boolean {
	for (let index = 0; index < Math.min(left.length, right.length); index += 1) {
		const comparison = compareText(left[index].id, right[index].id);
		if (comparison !== 0) return comparison < 0;
	}
	return left.length < right.length;
}

function maximumSafeSet(
	candidates: readonly DependencyWorkItem[],
	limit: number,
	conflicts: ReadonlyMap<string, ReadonlySet<string>>,
): DependencyWorkItem[] {
	let best: DependencyWorkItem[] = [];
	const chosen: DependencyWorkItem[] = [];
	const search = (index: number): boolean => {
		if (chosen.length > best.length || (chosen.length === best.length && lexicographicallyEarlier(chosen, best))) {
			best = [...chosen];
		}
		if (best.length === limit) return true;
		if (index >= candidates.length || chosen.length + (candidates.length - index) <= best.length) return false;
		const candidate = candidates[index];
		if (!chosen.some((selected) => conflicts.get(selected.id)?.has(candidate.id) === true)) {
			chosen.push(candidate);
			if (search(index + 1)) return true;
			chosen.pop();
		}
		return search(index + 1);
	};
	search(0);
	return best;
}

interface ConflictGraph {
	components: DependencyWorkItem[][];
	conflicts: Map<string, Set<string>>;
}

function buildConflictGraph(candidates: readonly DependencyWorkItem[]): ConflictGraph {
	const remaining = new Set(candidates.map((candidate) => candidate.id));
	const byId = new Map(candidates.map((candidate) => [candidate.id, candidate]));
	const canonicalScopes = new Map(candidates.map((candidate) => [
		candidate.id,
		candidate.writeScopes.map(scopeAliases),
	]));
	const conflicts = new Map(candidates.map((candidate) => [candidate.id, new Set<string>()]));
	for (let index = 0; index < candidates.length; index += 1) {
		for (let other = index + 1; other < candidates.length; other += 1) {
			const left = candidates[index];
			const right = candidates[other];
			if (left.access === "mutating" && right.access === "mutating"
				&& canonicalScopesCollide(canonicalScopes.get(left.id) ?? [], canonicalScopes.get(right.id) ?? [])) {
				conflicts.get(left.id)?.add(right.id);
				conflicts.get(right.id)?.add(left.id);
			}
		}
	}
	const components: DependencyWorkItem[][] = [];
	for (const root of candidates) {
		if (!remaining.delete(root.id)) continue;
		const component: DependencyWorkItem[] = [];
		const queue = [root];
		while (queue.length > 0) {
			const current = queue.shift();
			if (current === undefined) break;
			component.push(current);
			for (const id of conflicts.get(current.id) ?? []) {
				if (!remaining.delete(id)) continue;
				const candidate = byId.get(id);
				if (candidate !== undefined) queue.push(candidate);
			}
		}
		component.sort(compareIds);
		if (component.length > MAX_CONFLICT_COMPONENT_SIZE) {
			throw new DependencyGraphError(
				"conflict_component_too_large",
				`write-scope conflict component exceeds ${MAX_CONFLICT_COMPONENT_SIZE} items`,
			);
		}
		components.push(component);
	}
	return { components, conflicts };
}

function maximumSafeSelection(candidates: readonly DependencyWorkItem[], limit: number): DependencyWorkItem[] {
	const graph = buildConflictGraph(candidates);
	const selected = graph.components.flatMap((component) =>
		maximumSafeSet(component, Math.min(limit, component.length), graph.conflicts),
	);
	selected.sort(compareIds);
	return selected.slice(0, limit);
}

export function selectReadyWork(input: readonly DependencyWorkItem[], options: ReadyQueueOptions): ReadyQueueSelection {
	const graph = validateDependencyGraph(input);
	const items = graph.items;
	if (items.every((candidate) => candidate.status === "succeeded")) return { kind: "complete" };
	if (!Number.isSafeInteger(options.maxConcurrency) || options.maxConcurrency < 1 || options.maxConcurrency > MAX_WORK_ITEMS) {
		throw new RangeError(`maxConcurrency must be a positive safe integer no greater than ${MAX_WORK_ITEMS}`);
	}
	const running = items.filter((candidate) => candidate.status === "running");
	const available = options.maxConcurrency - running.length;
	if (available <= 0 && running.length > 0) return { kind: "at_capacity" };
	const byId = new Map(items.map((candidate) => [candidate.id, candidate]));
	const dependencyReady = items.filter((candidate) =>
		candidate.status === "pending"
		&& (options.allowMutating !== false || candidate.access === "read_only")
		&& candidate.dependsOn.every((dependency) => byId.get(dependency)?.status === "succeeded"),
	);
	const collisionFree = dependencyReady.filter((candidate) =>
		!running.some((active) => itemsCollide(candidate, active)),
	);
	if (collisionFree.length === 0) {
		return dependencyReady.length > 0
			? { kind: "blocked", blocker: "not_spawned_write_scope_collision" }
			: { kind: "blocked", blocker: "not_spawned_dependency_blocked" };
	}
	const selected = maximumSafeSelection(collisionFree, available);
	return { kind: "selected", itemIds: selected.map((candidate) => candidate.id) };
}
