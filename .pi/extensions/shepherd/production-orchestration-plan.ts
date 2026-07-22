import type { ProductionParentPlanDocument, ProductionVerificationCommand } from "./autonomous-production-contract.ts";
import { validateProductionParentPlan } from "./autonomous-production-contract.ts";
import { createParentOrchestrationPlan, type ParentOrchestrationPlan } from "./github-orchestrator.ts";
import type { RequiredGitHubCheckPolicy } from "./github-evidence.ts";

function verificationKind(command: ProductionVerificationCommand): "test" | "typecheck" | "offline_rpc" | "diff_scope" {
	const coordinate = `${command.id}\u0000${command.executable}\u0000${command.args.join("\u0000")}`.toLowerCase();
	if (coordinate.includes("typecheck") || coordinate.includes("tsc")) return "typecheck";
	if (coordinate.includes("rpc") || coordinate.includes("list-extensions")) return "offline_rpc";
	if (coordinate.includes("diff") || coordinate.includes("scope")) return "diff_scope";
	return "test";
}

export function productionOrchestrationObjective(planValue: ProductionParentPlanDocument, generation: number): unknown {
	const plan = validateProductionParentPlan(planValue, planValue.parentIssue);
	if (!Number.isSafeInteger(generation) || generation < 1) throw new Error("orchestration generation must be positive");
	return {
		repository: plan.repository,
		parentIssue: plan.parentIssue,
		generation,
		title: plan.title,
		objective: plan.objective,
		parentBranch: plan.parentBranch,
		parentBaseBranch: plan.parentBaseBranch,
		children: plan.children.map((child) => ({
			id: child.id,
			title: child.title,
			objective: child.task,
			slug: child.slug,
			dependsOn: [...child.dependsOn],
			access: "mutating" as const,
			writeScopes: [...child.writeScopes],
			requiredSkills: [...child.requiredSkills],
			verification: child.verification.map((command) => ({
				id: command.id,
				kind: verificationKind(command),
				description: `Run the bounded ${command.id} verification command.`,
			})),
			humanGates: [...child.humanGates],
		})),
	};
}

/** Converts the executable plan to the existing authoritative GitHub orchestration contract. */
export function createProductionOrchestrationPlan(
	planValue: ProductionParentPlanDocument,
	generation: number,
	requiredCheckPolicies: readonly RequiredGitHubCheckPolicy[],
): ParentOrchestrationPlan {
	return createParentOrchestrationPlan(productionOrchestrationObjective(planValue, generation), {
		schemaVersion: 1,
		requiredCheckPolicies,
	});
}
