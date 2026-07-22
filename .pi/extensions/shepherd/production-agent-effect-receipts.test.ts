import assert from "node:assert/strict";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import type { ProductionWorkspaceBinding } from "./autonomous-production-contract.ts";
import {
	ProductionAgentEffectReceiptRepository,
	productionAgentEffectResultDigest,
} from "./production-agent-effect-receipts.ts";

const EFFECT = "1".repeat(64);
const CLAIM = "2".repeat(64);
const SHA_A = "a".repeat(40);
const SHA_B = "b".repeat(40);

function binding(root: string, overrides: Partial<ProductionWorkspaceBinding> = {}): ProductionWorkspaceBinding {
	return {
		claimId: CLAIM,
		ownershipId: `production:${"3".repeat(64)}`,
		repositoryIdentity: "4".repeat(64),
		worktreeIdentity: "5".repeat(64),
		cwd: join(root, "issue-501-child"),
		branch: "feat/501-child",
		baseBranch: "feat/479-parent",
		baseHead: SHA_A,
		head: SHA_A,
		writeScopes: [".pi/extensions/shepherd"],
		...overrides,
	};
}

async function fixture(t: test.TestContext) {
	const root = await mkdtemp(join(tmpdir(), "shepherd-agent-effect-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	return { root, receipts: new ProductionAgentEffectReceiptRepository(root) };
}

test("start-only receipt is durable ambiguous evidence, not a fabricated completion", async (t) => {
	const { root, receipts } = await fixture(t);
	const start = {
		schemaVersion: 1 as const,
		effectKey: EFFECT,
		claimId: CLAIM,
		role: "implementation" as const,
		binding: binding(root),
	};
	assert.equal(await receipts.find(EFFECT), undefined);
	await receipts.begin(start);
	assert.deepEqual(await receipts.find(EFFECT), { start });
	assert.equal((await receipts.find(EFFECT))?.completion, undefined);
});

test("a valid no-op AgentSession has exact durable completion evidence", async (t) => {
	const { root, receipts } = await fixture(t);
	const initial = binding(root);
	const start = {
		schemaVersion: 1 as const,
		effectKey: EFFECT,
		claimId: CLAIM,
		role: "implementation" as const,
		binding: initial,
	};
	await receipts.begin(start);
	await receipts.complete({
		...start,
		resultDigest: productionAgentEffectResultDigest(initial),
		completedBinding: initial,
	});
	const recovered = await receipts.find(EFFECT);
	assert.deepEqual(recovered?.completion?.completedBinding, initial);
	assert.equal(recovered?.completion?.resultDigest, productionAgentEffectResultDigest(initial));
	await receipts.complete(recovered!.completion!);
});

test("completed mutation binds the exact resulting workspace and rejects false digests", async (t) => {
	const { root, receipts } = await fixture(t);
	const initial = binding(root);
	const completed = binding(root, { head: SHA_B });
	const start = {
		schemaVersion: 1 as const,
		effectKey: EFFECT,
		claimId: CLAIM,
		role: "correction" as const,
		binding: initial,
	};
	await receipts.begin(start);
	await assert.rejects(receipts.complete({
		...start,
		resultDigest: "9".repeat(64),
		completedBinding: completed,
	}), /result digest/i);
	await receipts.complete({
		...start,
		resultDigest: productionAgentEffectResultDigest(completed),
		completedBinding: completed,
	});
	assert.equal((await receipts.find(EFFECT))?.completion?.completedBinding.head, SHA_B);
});

test("same-key retries and malformed authority bindings fail closed", async (t) => {
	const { root, receipts } = await fixture(t);
	const start = {
		schemaVersion: 1 as const,
		effectKey: EFFECT,
		claimId: CLAIM,
		role: "implementation" as const,
		binding: binding(root),
	};
	await receipts.begin(start);
	await assert.rejects(receipts.begin({
		...start,
		binding: binding(root, { head: SHA_B }),
	}), /conflicts/i);
	await assert.rejects(receipts.begin({
		...start,
		effectKey: "../escape",
	}), /malformed|invalid/i);
	await assert.rejects(receipts.begin({
		...start,
		effectKey: "6".repeat(64),
		binding: { ...binding(root), writeScopes: ["../escape"] },
	}), /binding.*malformed/i);
});
