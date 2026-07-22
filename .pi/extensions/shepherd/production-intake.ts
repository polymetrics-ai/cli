import { constants } from "node:fs";
import { lstat, open, realpath } from "node:fs/promises";
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

/** Reads one canonical, bounded, non-symlink repository plan without exposing a generic file port. */
export class ProductionRepositoryPlanIntake {
	readonly #cwd: string;

	constructor(cwd: string) {
		this.#cwd = resolve(cwd);
	}

	async load(issue: number, signal: AbortSignal): Promise<ProductionPlanSnapshot> {
		if (!Number.isSafeInteger(issue) || issue < 1) throw new Error("issue must be a positive integer");
		if (signal.aborted) throw signal.reason ?? new Error("production plan intake cancelled");
		const repositoryRoot = await realpath(this.#cwd);
		const path = resolve(repositoryRoot, ".planning", "shepherd", `issue-${issue}.json`);
		const parent = await realpath(dirname(path));
		if (!isWithin(repositoryRoot, parent)) throw new Error("production plan directory escapes the repository");
		const metadata = await lstat(path);
		if (!metadata.isFile() || metadata.isSymbolicLink()) throw new Error("production plan must be a regular non-symlink file");
		if (metadata.size < 2 || metadata.size > MAX_PLAN_BYTES) throw new Error("production plan exceeds its bounded size");
		const noFollow = "O_NOFOLLOW" in constants ? constants.O_NOFOLLOW : 0;
		const handle = await open(path, constants.O_RDONLY | noFollow);
		try {
			const opened = await handle.stat();
			if (!opened.isFile() || opened.size !== metadata.size || opened.size > MAX_PLAN_BYTES) {
				throw new Error("production plan changed during intake");
			}
			const raw = await handle.readFile({ encoding: "utf8", signal });
			if (signal.aborted) throw signal.reason ?? new Error("production plan intake cancelled");
			let parsed: unknown;
			try {
				parsed = JSON.parse(raw);
			} catch (error) {
				throw new Error("production plan is not valid JSON", { cause: error });
			}
			const plan = validateProductionParentPlan(parsed, issue);
			return { plan, digest: productionPlanDigest(plan), path };
		} finally {
			await handle.close();
		}
	}
}
