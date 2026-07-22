import { selectReadyWork, type DependencyWorkItem, type WorkItemStatus } from "./dependency-graph.ts";
import type { ProductionAutonomousState, ProductionChildRuntimeState } from "./autonomous-production-state.ts";

export type ProductionScheduleDecision =
	| { kind: "dispatch"; childIds: string[] }
	| { kind: "idle"; reason: "capacity" | "dependencies" | "write_scope_collision" }
	| { kind: "complete" };

function schedulerStatus(child: ProductionChildRuntimeState): WorkItemStatus {
	if (child.status === "cancelled") return "pending";
	return child.status;
}

function item(child: ProductionChildRuntimeState): DependencyWorkItem {
	return {
		id: child.id,
		dependsOn: [...child.dependsOn],
		status: schedulerStatus(child),
		access: "mutating",
		writeScopes: [...child.writeScopes],
	};
}

/** Pure deterministic scheduler adapter used before any worktree or network effect. */
export function selectProductionChildren(state: ProductionAutonomousState): ProductionScheduleDecision {
	const selection = selectReadyWork(state.children.map(item), {
		maxConcurrency: state.maxConcurrency,
		allowMutating: true,
	});
	if (selection.kind === "selected") return { kind: "dispatch", childIds: selection.itemIds };
	if (selection.kind === "complete") return { kind: "complete" };
	if (selection.kind === "at_capacity") return { kind: "idle", reason: "capacity" };
	return {
		kind: "idle",
		reason: selection.blocker === "not_spawned_write_scope_collision"
			? "write_scope_collision"
			: "dependencies",
	};
}
