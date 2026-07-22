import { createHash, randomUUID } from "node:crypto";
import { constants } from "node:fs";
import { link, lstat, mkdir, open, realpath, rm } from "node:fs/promises";
import { basename, dirname, isAbsolute, join, resolve } from "node:path";

import type { ProductionWorkspaceBinding } from "./autonomous-production-contract.ts";

const DIGEST = /^[0-9a-f]{64}$/u;
const SAFE_EFFECT = /^[A-Za-z0-9](?:[A-Za-z0-9._:-]*[A-Za-z0-9])?$/u;
const MAX_BYTES = 128 * 1024;
const SHA = /^[0-9a-f]{40}$/u;
const SAFE_BRANCH = /^(?!\/|.*(?:\.\.|\s|[~^:?*\\\[\]])|.*\/$)[A-Za-z0-9][A-Za-z0-9._\/-]{0,239}$/u;
const BINDING_FIELDS = [
	"claimId", "ownershipId", "repositoryIdentity", "worktreeIdentity", "cwd", "branch", "baseBranch",
	"baseHead", "head", "writeScopes",
] as const;

export interface ProductionAgentEffectStart {
	schemaVersion: 1;
	effectKey: string;
	claimId: string;
	role: "implementation" | "correction";
	binding: ProductionWorkspaceBinding;
}

export interface ProductionAgentEffectCompletion extends ProductionAgentEffectStart {
	resultDigest: string;
	completedBinding: ProductionWorkspaceBinding;
}

export interface ProductionAgentEffectReceipt {
	start: ProductionAgentEffectStart;
	completion?: ProductionAgentEffectCompletion;
}

function key(effectKey: string): string {
	if (!SAFE_EFFECT.test(effectKey)) throw new Error("agent effect key is invalid");
	return createHash("sha256").update(effectKey).digest("hex");
}

function canonicalValue(value: unknown): unknown {
	if (Array.isArray(value)) return value.map(canonicalValue);
	if (typeof value !== "object" || value === null) return value;
	const output: Record<string, unknown> = {};
	for (const field of Object.keys(value as Record<string, unknown>).sort()) {
		const item = (value as Record<string, unknown>)[field];
		if (item !== undefined) output[field] = canonicalValue(item);
	}
	return output;
}

function canonical(value: unknown): string { return JSON.stringify(canonicalValue(value)); }

function safeText(value: unknown, maximum: number): value is string {
	return typeof value === "string" && value.length > 0 && Buffer.byteLength(value) <= maximum
		&& !/[\u0000-\u001f\u007f-\u009f]/u.test(value);
}

function validateBinding(value: unknown): ProductionWorkspaceBinding {
	if (typeof value !== "object" || value === null || Array.isArray(value)) {
		throw new Error("agent effect workspace binding is malformed");
	}
	const record = value as Record<string, unknown>;
	if (Object.keys(record).length !== BINDING_FIELDS.length
		|| BINDING_FIELDS.some((field) => !Object.hasOwn(record, field))
		|| !safeText(record.claimId, 256) || !safeText(record.ownershipId, 256)
		|| !safeText(record.repositoryIdentity, 256) || !safeText(record.worktreeIdentity, 4_096)
		|| !safeText(record.cwd, 4_096) || !isAbsolute(record.cwd)
		|| typeof record.branch !== "string" || !SAFE_BRANCH.test(record.branch)
		|| typeof record.baseBranch !== "string" || !SAFE_BRANCH.test(record.baseBranch)
		|| typeof record.baseHead !== "string" || !SHA.test(record.baseHead)
		|| typeof record.head !== "string" || !SHA.test(record.head)
		|| !Array.isArray(record.writeScopes) || record.writeScopes.length < 1 || record.writeScopes.length > 64
		|| record.writeScopes.some((scope) => !safeText(scope, 4_096) || scope === "."
			|| scope.startsWith("-") || scope.startsWith("/") || scope.includes("\\")
			|| scope.split("/").some((part) => part.length === 0 || part === "." || part === ".."))
		|| new Set(record.writeScopes).size !== record.writeScopes.length) {
		throw new Error("agent effect workspace binding is malformed");
	}
	return structuredClone(record) as unknown as ProductionWorkspaceBinding;
}

/** The exact stable result journaled by implementation and correction effects. */
export function productionAgentEffectResultDigest(binding: ProductionWorkspaceBinding): string {
	return createHash("sha256").update(canonical({ workspace: validateBinding(binding) })).digest("hex");
}

async function safeRoot(trustedRoot: string): Promise<string> {
	if (!isAbsolute(trustedRoot)) throw new Error("agent effect receipt root must be absolute");
	const root = await realpath(trustedRoot);
	const metadata = await lstat(root);
	if (!metadata.isDirectory() || metadata.isSymbolicLink()) throw new Error("agent effect receipt root is unsafe");
	return root;
}

async function publish(path: string, value: unknown): Promise<void> {
	const payload = `${canonical(value)}\n`;
	if (Buffer.byteLength(payload) > MAX_BYTES) throw new Error("agent effect receipt exceeds its bound");
	await mkdir(dirname(path), { recursive: true, mode: 0o700 });
	const parent = await lstat(dirname(path));
	if (!parent.isDirectory() || parent.isSymbolicLink() || (parent.mode & 0o077) !== 0) {
		throw new Error("agent effect receipt directory is unsafe");
	}
	const temporary = join(dirname(path), `.${basename(path)}.${randomUUID()}.tmp`);
	const handle = await open(temporary, "wx", 0o600);
	try { await handle.writeFile(payload, "utf8"); await handle.sync(); }
	finally { await handle.close(); }
	try {
		try { await link(temporary, path); }
		catch (error) {
			if (!(typeof error === "object" && error !== null && "code" in error && error.code === "EEXIST")) throw error;
		}
		const existing = await read(path);
		if (canonical(existing) !== canonical(value)) throw new Error("agent effect receipt conflicts with its exact retry");
	} finally { await rm(temporary, { force: true }); }
}

async function read(path: string): Promise<unknown> {
	const handle = await open(path, constants.O_RDONLY | (constants.O_NOFOLLOW ?? 0));
	try {
		const metadata = await handle.stat();
		if (!metadata.isFile() || metadata.size < 1 || metadata.size > MAX_BYTES || (metadata.mode & 0o777) !== 0o600) {
			throw new Error("agent effect receipt is not a bounded mode-0600 file");
		}
		return JSON.parse(await handle.readFile("utf8")) as unknown;
	} finally { await handle.close(); }
}

function validateStart(value: unknown): ProductionAgentEffectStart {
	if (typeof value !== "object" || value === null || Array.isArray(value)) throw new Error("agent effect start receipt is malformed");
	const record = value as Record<string, unknown>;
	if (Object.keys(record).some((field) => !["schemaVersion", "effectKey", "claimId", "role", "binding"].includes(field))
		|| record.schemaVersion !== 1 || typeof record.effectKey !== "string" || !SAFE_EFFECT.test(record.effectKey)
		|| typeof record.claimId !== "string" || !DIGEST.test(record.claimId)
		|| (record.role !== "implementation" && record.role !== "correction")) {
		throw new Error("agent effect start receipt is malformed");
	}
	return {
		schemaVersion: 1,
		effectKey: record.effectKey,
		claimId: record.claimId,
		role: record.role,
		binding: validateBinding(record.binding),
	};
}

function validateCompletion(value: unknown): ProductionAgentEffectCompletion {
	if (typeof value !== "object" || value === null || Array.isArray(value)) throw new Error("agent effect completion receipt is malformed");
	const record = value as Record<string, unknown>;
	if (Object.keys(record).some((field) => ![
		"schemaVersion", "effectKey", "claimId", "role", "binding", "resultDigest", "completedBinding",
	].includes(field)) || typeof record.resultDigest !== "string" || !DIGEST.test(record.resultDigest)
		|| typeof record.completedBinding !== "object" || record.completedBinding === null || Array.isArray(record.completedBinding)) {
		throw new Error("agent effect completion receipt is malformed");
	}
	const start = validateStart({
		schemaVersion: record.schemaVersion,
		effectKey: record.effectKey,
		claimId: record.claimId,
		role: record.role,
		binding: record.binding,
	});
	const completedBinding = validateBinding(record.completedBinding);
	if (record.resultDigest !== productionAgentEffectResultDigest(completedBinding)) {
		throw new Error("agent effect completion result digest is invalid");
	}
	return {
		...start,
		resultDigest: record.resultDigest,
		completedBinding,
	};
}

export class ProductionAgentEffectReceiptRepository {
	readonly #trustedRoot: string;

	constructor(trustedRoot: string) {
		if (!isAbsolute(trustedRoot)) throw new Error("agent effect receipt root must be absolute");
		this.#trustedRoot = resolve(trustedRoot);
	}

	async #paths(effectKey: string) {
		const root = await safeRoot(this.#trustedRoot);
		const directory = join(root, ".shepherd-agent-effects");
		const identity = key(effectKey);
		return { start: join(directory, `${identity}.start.json`), complete: join(directory, `${identity}.complete.json`) };
	}

	async begin(value: ProductionAgentEffectStart): Promise<void> {
		const start = validateStart(value);
		const paths = await this.#paths(start.effectKey);
		await publish(paths.start, start);
	}

	async complete(value: ProductionAgentEffectCompletion): Promise<void> {
		const completion = validateCompletion(value);
		const paths = await this.#paths(completion.effectKey);
		const start = validateStart(await read(paths.start));
		if (canonical(start) !== canonical({
			schemaVersion: completion.schemaVersion,
			effectKey: completion.effectKey,
			claimId: completion.claimId,
			role: completion.role,
			binding: completion.binding,
		})) throw new Error("agent effect completion moved from its exact start receipt");
		await publish(paths.complete, completion);
	}

	async find(effectKey: string): Promise<ProductionAgentEffectReceipt | undefined> {
		const paths = await this.#paths(effectKey);
		let start: ProductionAgentEffectStart;
		try { start = validateStart(await read(paths.start)); }
		catch (error) {
			if (typeof error === "object" && error !== null && "code" in error && error.code === "ENOENT") return undefined;
			throw error;
		}
		let completion: ProductionAgentEffectCompletion | undefined;
		try { completion = validateCompletion(await read(paths.complete)); }
		catch (error) {
			if (!(typeof error === "object" && error !== null && "code" in error && error.code === "ENOENT")) throw error;
		}
		if (completion !== undefined && (completion.effectKey !== start.effectKey || completion.claimId !== start.claimId
			|| completion.role !== start.role || canonical(completion.binding) !== canonical(start.binding))) {
			throw new Error("agent effect completion conflicts with its start receipt");
		}
		return { start, ...(completion === undefined ? {} : { completion }) };
	}
}
