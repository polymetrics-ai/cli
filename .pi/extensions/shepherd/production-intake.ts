import { randomUUID } from "node:crypto";
import { constants } from "node:fs";
import { link, lstat, mkdir, open, realpath, unlink } from "node:fs/promises";
import { dirname, relative, resolve, sep } from "node:path";

import {
	productionPlanDigest,
	validateProductionParentPlan,
	type ProductionParentPlanDocument,
} from "./autonomous-production-contract.ts";

const MAX_PLAN_BYTES = 1024 * 1024;

export interface ProductionPlanSnapshot {
	plan: ProductionParentPlanDocument;
	digest: string;
	path: string;
}

function isWithin(root: string, candidate: string): boolean {
	const path = relative(root, candidate);
	return path === "" || (path !== ".." && !path.startsWith(`..${sep}`));
}

function isMissing(error: unknown): boolean {
	return typeof error === "object" && error !== null && "code" in error && error.code === "ENOENT";
}

function assertActive(signal: AbortSignal): void {
	if (!(signal instanceof AbortSignal)) throw new Error("production plan AbortSignal is invalid");
	if (signal.aborted) throw signal.reason ?? new Error("production plan intake cancelled");
}

/** Reads and atomically publishes one canonical, bounded, non-symlink repository plan. */
export class ProductionRepositoryPlanIntake {
	readonly #cwd: string;

	constructor(cwd: string) {
		this.#cwd = resolve(cwd);
	}

	async tryLoad(issue: number, signal: AbortSignal): Promise<ProductionPlanSnapshot | undefined> {
		if (!Number.isSafeInteger(issue) || issue < 1) throw new Error("issue must be a positive integer");
		assertActive(signal);
		const location = await this.#location(issue, false);
		if (location === undefined) return undefined;
		const { repositoryRoot, path } = location;
		let metadata: Awaited<ReturnType<typeof lstat>>;
		try {
			metadata = await lstat(path);
		} catch (error) {
			if (isMissing(error)) return undefined;
			throw error;
		}
		if (!metadata.isFile() || metadata.isSymbolicLink() || metadata.nlink !== 1) {
			throw new Error("production plan must be a regular non-symlink single-link file");
		}
		if (metadata.size < 2 || metadata.size > MAX_PLAN_BYTES) {
			throw new Error("production plan exceeds its bounded size");
		}
		const noFollow = "O_NOFOLLOW" in constants ? constants.O_NOFOLLOW : 0;
		const handle = await open(path, constants.O_RDONLY | noFollow);
		try {
			const opened = await handle.stat();
			if (!opened.isFile() || opened.nlink !== 1 || opened.ino !== metadata.ino
				|| opened.dev !== metadata.dev || opened.size !== metadata.size
				|| opened.size > MAX_PLAN_BYTES) {
				throw new Error("production plan changed during intake");
			}
			const raw = await handle.readFile({ encoding: "utf8", signal });
			assertActive(signal);
			let parsed: unknown;
			try {
				parsed = JSON.parse(raw);
			} catch (error) {
				throw new Error("production plan is not valid JSON", { cause: error });
			}
			const plan = validateProductionParentPlan(parsed, issue);
			if (!isWithin(repositoryRoot, path)) throw new Error("production plan escapes the repository");
			return { plan, digest: productionPlanDigest(plan), path };
		} finally {
			await handle.close();
		}
	}

	async load(issue: number, signal: AbortSignal): Promise<ProductionPlanSnapshot> {
		const snapshot = await this.tryLoad(issue, signal);
		if (snapshot === undefined) {
			throw new Error(`production plan is unavailable; create .planning/shepherd/issue-${issue}.json`);
		}
		return snapshot;
	}

	async publish(
		issue: number,
		value: ProductionParentPlanDocument,
		signal: AbortSignal,
	): Promise<ProductionPlanSnapshot> {
		if (!Number.isSafeInteger(issue) || issue < 1) throw new Error("issue must be a positive integer");
		assertActive(signal);
		const plan = validateProductionParentPlan(value, issue);
		const digest = productionPlanDigest(plan);
		const existing = await this.tryLoad(issue, signal);
		if (existing !== undefined) {
			if (existing.digest !== digest) throw new Error("existing production plan conflicts with generated plan");
			return existing;
		}
		const location = await this.#location(issue, true);
		if (location === undefined) throw new Error("production plan directory could not be created");
		const { parent, path } = location;
		const temporary = resolve(parent, `.${issue}.${process.pid}.${randomUUID()}.tmp`);
		const serialized = `${JSON.stringify(plan)}\n`;
		if (Buffer.byteLength(serialized) > MAX_PLAN_BYTES) throw new Error("production plan exceeds its bounded size");
		const handle = await open(
			temporary,
			constants.O_WRONLY | constants.O_CREAT | constants.O_EXCL,
			0o600,
		);
		try {
			await handle.writeFile(serialized, { encoding: "utf8", signal });
			await handle.sync();
		} finally {
			await handle.close();
		}
		assertActive(signal);
		try {
			await link(temporary, path);
		} catch (error) {
			if (!isMissing(error) && !(typeof error === "object" && error !== null && "code" in error && error.code === "EEXIST")) {
				throw error;
			}
			if (isMissing(error)) throw error;
			const raced = await this.load(issue, signal);
			if (raced.digest !== digest) throw new Error("concurrent production plan conflicts with generated plan");
			return raced;
		} finally {
			await unlink(temporary).catch((error) => { if (!isMissing(error)) throw error; });
		}
		const directory = await open(parent, constants.O_RDONLY);
		try { await directory.sync(); } finally { await directory.close(); }
		const published = await this.load(issue, signal);
		if (published.digest !== digest) throw new Error("published production plan digest changed");
		return published;
	}

	async #location(
		issue: number,
		create: boolean,
	): Promise<{ repositoryRoot: string; parent: string; path: string } | undefined> {
		const repositoryRoot = await realpath(this.#cwd);
		let current = repositoryRoot;
		for (const component of [".planning", "shepherd"]) {
			current = resolve(current, component);
			let metadata: Awaited<ReturnType<typeof lstat>>;
			try {
				metadata = await lstat(current);
			} catch (error) {
				if (!isMissing(error)) throw error;
				if (!create) return undefined;
				await mkdir(current, { mode: 0o700 });
				metadata = await lstat(current);
			}
			if (!metadata.isDirectory() || metadata.isSymbolicLink()) {
				throw new Error("production plan directory must contain only real directories");
			}
			if (!isWithin(repositoryRoot, await realpath(current))) {
				throw new Error("production plan directory escapes the repository");
			}
		}
		const parent = await realpath(current);
		const path = resolve(parent, `issue-${issue}.json`);
		if (!isWithin(repositoryRoot, path) || dirname(path) !== parent) {
			throw new Error("production plan path escapes the repository");
		}
		return { repositoryRoot, parent, path };
	}
}
