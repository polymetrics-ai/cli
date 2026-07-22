import assert from "node:assert/strict";
import { mkdtemp, readFile, readdir, rm, symlink, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import {
	FileHumanDecisionRepository,
	consumeHumanDecision,
	createHumanDecisionRecord,
	persistHumanDecisionRequest,
	recordHumanDecision,
	recordHumanDecisionRequestComment,
	routeHumanDecisionTarget,
	validateHumanDecisionRecord,
	validateHumanDecisionRequestComment,
	type HumanDecisionBinding,
	type HumanDecisionRecord,
	type HumanDecisionRequestSpec,
} from "./human-decision.ts";

const head = "a".repeat(40);
const issueBinding: HumanDecisionBinding = {
	repository: "polymetrics-ai/cli",
	target: { kind: "issue", number: 471 },
	generation: 3,
};
const prBinding: HumanDecisionBinding = {
	repository: "polymetrics-ai/cli",
	target: { kind: "pull_request", number: 477 },
	generation: 3,
	headSha: head,
};

function spec(overrides: {
	requestId?: string;
	gate?: HumanDecisionRequestSpec["gate"];
	binding?: HumanDecisionBinding;
	allowedOptions?: string[];
	actorAllowlist?: string[];
	expiresAt?: string;
	question?: string;
} = {}): HumanDecisionRequestSpec {
	return {
		requestId: "req-477",
		gate: "requirements",
		binding: issueBinding,
		allowedOptions: ["approve", "reject"],
		actorAllowlist: ["maintainer-one", "Maintainer-Two"],
		expiresAt: "2026-07-22T10:00:00.000Z",
		question: "Approve the exact requirements for issue #471?",
		...overrides,
	} as HumanDecisionRequestSpec;
}

test("routes requirements and scope only to the parent issue", () => {
	for (const gate of ["requirements", "scope"] as const) {
		assert.deepEqual(routeHumanDecisionTarget(gate, 471, 477), { kind: "issue", number: 471 });
	}
});

test("routes review, head, merge, and distinct parent merge gates only to the PR", () => {
	for (const gate of ["review", "head", "merge", "parent_merge"] as const) {
		assert.deepEqual(routeHumanDecisionTarget(gate, 471, 477), { kind: "pull_request", number: 477 });
	}
});

test("requires exact head binding for every PR gate and rejects head binding for issue gates", () => {
	assert.throws(() => createHumanDecisionRecord(spec({ gate: "review", binding: { ...prBinding, headSha: undefined } })), /head/i);
	assert.throws(() => createHumanDecisionRecord(spec({ gate: "requirements", binding: { ...issueBinding, headSha: head } })), /head/i);
	assert.throws(() => createHumanDecisionRecord(spec({ gate: "parent_merge", binding: { ...prBinding, headSha: "b".repeat(39) } })), /head/i);
});

test("parent merge remains a distinct exact-head request", () => {
	assert.throws(() => createHumanDecisionRecord(spec({
		gate: "parent_merge",
		binding: prBinding,
		allowedOptions: ["approve", "reject"],
	})), /approve-merge/i);
	const parentMerge = createHumanDecisionRecord(spec({
		gate: "parent_merge",
		binding: prBinding,
		allowedOptions: ["approve-merge", "reject"],
	}));
	const merge = createHumanDecisionRecord(spec({ gate: "merge", binding: prBinding }));
	assert.equal(parentMerge.gate, "parent_merge");
	assert.deepEqual(parentMerge.allowedOptions, ["approve-merge", "reject"]);
	assert.notEqual(parentMerge.idempotencyMarker, merge.idempotencyMarker);
});

test("accepts digits in GitHub repositories and human logins", () => {
	const record = createHumanDecisionRecord(spec({
		binding: { ...issueBinding, repository: "Owner2/Repo3" },
		actorAllowlist: ["Maintainer2"],
	}));
	assert.equal(record.binding.repository, "owner2/repo3");
	assert.deepEqual(record.actorAllowlist, ["maintainer2"]);
});

test("normalizes GitHub second-resolution timestamps and safe-integer comment IDs with bounded skew", () => {
	const record = createHumanDecisionRecord(spec({
		expiresAt: "2026-07-22T10:00:00Z",
	}), new Date("2026-07-21T12:30:12.750Z"));
	assert.equal(record.expiresAt, "2026-07-22T10:00:00.000Z");
	const comment = validateHumanDecisionRequestComment(record, {
		id: 5_034_006_493,
		url: "https://github.com/polymetrics-ai/cli/issues/471#issuecomment-5034006493",
		actor: "shepherd-host",
		createdAt: "2026-07-21T12:30:12Z",
	});
	assert.deepEqual(comment, {
		id: 5_034_006_493,
		url: "https://github.com/polymetrics-ai/cli/issues/471#issuecomment-5034006493",
		actor: "shepherd-host",
		createdAt: "2026-07-21T12:30:12.000Z",
	});
	assert.throws(() => validateHumanDecisionRequestComment(record, {
		...comment,
		id: 5_034_006_494,
		url: "https://github.com/polymetrics-ai/cli/issues/471#issuecomment-5034006494",
		createdAt: "2026-07-21T12:30:11Z",
	}), /lifetime|timestamp/i);
});

test("canonicalizes actors/options and rejects duplicate, malformed, expired, or secret-bearing requests", () => {
	const record = createHumanDecisionRecord(spec());
	assert.deepEqual(record.actorAllowlist, ["maintainer-one", "maintainer-two"]);
	assert.deepEqual(record.allowedOptions, ["approve", "reject"]);
	for (const candidate of [
		spec({ allowedOptions: ["approve", "approve"] }),
		spec({ actorAllowlist: ["same", "SAME"] }),
		spec({ requestId: "../escape" }),
		spec({ expiresAt: "2026-07-21T09:59:59.000Z" }),
		spec({ question: "Authorization: Bearer secret-value" }),
		spec({ question: "token github_pat_1234567890123456789012" }),
	]) {
		assert.throws(() => createHumanDecisionRecord(candidate, new Date("2026-07-21T10:00:00.000Z")));
	}
});

test("rejects centralized credential forms without reflecting their values", () => {
	const credentialMarker = "synthetic-credential-marker";
	const candidates = [
		["token", credentialMarker].join("="),
		`{"api_key":"${credentialMarker}"}`,
		`{"password":${credentialMarker}}`,
		["OPENAI_API_KEY", credentialMarker].join("="),
		["AWS_SECRET_ACCESS_KEY", credentialMarker].join("="),
		`https://user:${credentialMarker}@example.invalid/path`,
		`https://example.invalid/path?access_token=${credentialMarker}`,
		["sk", "live", credentialMarker].join("_"),
	];
	for (const question of candidates) {
		assert.throws(
			() => createHumanDecisionRecord(spec({ question })),
			(error: unknown) => error instanceof Error
				&& /credential|secret/i.test(error.message)
				&& !error.message.includes(credentialMarker),
			question,
		);
	}
});

test("rejects invisible bidi controls and untrusted mentions in decision questions", () => {
	for (const question of [
		"Approve this exact gate?\u202Edetarapes",
		"Approve this exact gate?\u2028Spoofed line",
		"Ask @attacker to approve this gate",
	]) {
		assert.throws(() => createHumanDecisionRecord(spec({ question })), /question|format|mention/i);
	}
});

test("durably round-trips and rejects a changed retry specification after restart", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-477-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const first = new FileHumanDecisionRepository(root, { lockRetryMs: 1 });
	const created = await persistHumanDecisionRequest(first, spec(), new Date("2026-07-21T10:00:00.000Z"));
	const restarted = new FileHumanDecisionRepository(root, { lockRetryMs: 1 });
	assert.deepEqual(await restarted.load("req-477"), created);
	await assert.rejects(
		persistHumanDecisionRequest(restarted, spec({ allowedOptions: ["approve", "defer"] })),
		/conflict|differs/i,
	);
	const serialized = await readFile(join(root, "req-477.json"), "utf8");
	assert.equal(serialized.includes("github_pat_"), false);
});

test("persists minimal accepted evidence and consumes it exactly once across repository instances", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-477-consume-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const first = new FileHumanDecisionRepository(root, { lockRetryMs: 1 });
	const second = new FileHumanDecisionRepository(root, { lockRetryMs: 1 });
	await persistHumanDecisionRequest(first, spec(), new Date("2026-07-21T10:00:00.000Z"));
	await recordHumanDecisionRequestComment(first, "req-477", issueBinding, {
		id: 1001,
		url: "https://github.com/polymetrics-ai/cli/issues/471#issuecomment-1001",
		actor: "shepherd-host",
		createdAt: "2026-07-21T10:00:10.000Z",
	}, new Date("2026-07-21T10:00:10.000Z"));
	await recordHumanDecision(first, "req-477", issueBinding, {
		option: "approve",
		actor: "maintainer-one",
		sourceUrl: "https://github.com/polymetrics-ai/cli/issues/471#issuecomment-2001",
		decidedAt: "2026-07-21T10:01:00.000Z",
	});

	const attempts = await Promise.allSettled([
		consumeHumanDecision(first, "req-477", issueBinding, new Date("2026-07-21T10:02:00.000Z")),
		consumeHumanDecision(second, "req-477", issueBinding, new Date("2026-07-21T10:02:00.000Z")),
	]);
	assert.equal(attempts.filter((attempt) => attempt.status === "fulfilled").length, 1);
	assert.equal(attempts.filter((attempt) => attempt.status === "rejected").length, 1);
	const accepted = attempts.find((attempt) => attempt.status === "fulfilled");
	assert.equal(accepted?.status === "fulfilled" && accepted.value.option, "approve");
	assert.deepEqual(Object.keys(accepted?.status === "fulfilled" ? accepted.value : {}).sort(), [
		"actor", "decidedAt", "option", "sourceUrl",
	]);
	await assert.rejects(consumeHumanDecision(first, "req-477", issueBinding), /consum/i);
});

test("fails closed when generation, repository, target, or head differs at decision time", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-477-binding-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const repository = new FileHumanDecisionRepository(root, { lockRetryMs: 1 });
	await persistHumanDecisionRequest(repository, spec({ gate: "review", binding: prBinding }));
	for (const binding of [
		{ ...prBinding, generation: 4 },
		{ ...prBinding, repository: "other/repo" },
		{ ...prBinding, target: { kind: "pull_request" as const, number: 478 } },
		{ ...prBinding, headSha: "b".repeat(40) },
	]) {
		await assert.rejects(recordHumanDecision(repository, "req-477", binding, {
			option: "approve",
			actor: "maintainer-one",
			sourceUrl: "https://github.com/polymetrics-ai/cli/pull/477#issuecomment-1",
			decidedAt: "2026-07-21T10:01:00.000Z",
		}), /binding|stale|target/i);
	}
});

test("rejects unknown persisted fields and symbolic-link state after restart", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-477-tamper-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const repository = new FileHumanDecisionRepository(root, { lockRetryMs: 1 });
	await persistHumanDecisionRequest(repository, spec(), new Date("2026-07-21T10:00:00.000Z"));
	const path = join(root, "req-477.json");
	const persisted = JSON.parse(await readFile(path, "utf8"));
	persisted.untrusted = "must-not-load";
	await writeFile(path, JSON.stringify(persisted), { mode: 0o600 });
	await assert.rejects(repository.load("req-477"), /unknown|field/i);
	await symlink(path, join(root, "req-link.json"));
	await assert.rejects(repository.load("req-link"), /ELOOP|symbolic|link/i);
});

test("publishes a complete atomic lock owner record before entering a transaction", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-477-atomic-lock-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const repository = new FileHumanDecisionRepository(root, { lockRetryMs: 1 });
	const value = await repository.transact("req-477", async () => {
		const locks = (await readdir(root)).filter((name) => /^req-477\.lock\.[0-9a-f-]{36}\.active$/.test(name));
		assert.equal(locks.length, 1);
		const owner = JSON.parse(await readFile(join(root, locks[0]), "utf8"));
		assert.equal(owner.schemaVersion, 1);
		assert.equal(owner.pid, process.pid);
		assert.match(owner.token, /^[0-9a-f-]{36}$/);
		return { state: null, value: owner.token as string };
	});
	assert.match(value, /^[0-9a-f-]{36}$/);
});

test("rescans candidate transitions and active releases under high contention", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-477-lock-snapshot-race-"));
	t.after(() => rm(root, { recursive: true, force: true }));

	for (let round = 0; round < 3; round += 1) {
		let active = 0;
		let maximumActive = 0;
		const contenders = Array.from({ length: 32 }, (_, index) => {
			const repository = new FileHumanDecisionRepository(root, {
				lockRetryMs: 1,
				lockMaxAttempts: 2_000,
			});
			return repository.transact(`req-race-${round}`, async () => {
				active += 1;
				maximumActive = Math.max(maximumActive, active);
				try {
					await new Promise((resolve) => setTimeout(resolve, 2));
					return { state: null, value: index };
				} finally {
					active -= 1;
				}
			});
		});
		const settled = await Promise.allSettled(contenders);
		const failureKinds = [...new Set(settled
			.filter((result): result is PromiseRejectedResult => result.status === "rejected")
			.map((result) => {
				if (!(result.reason instanceof Error)) return "non-error rejection";
				const code = (result.reason as NodeJS.ErrnoException).code;
				if (code) return code;
				return /lock owner record is invalid/i.test(result.reason.message)
					? "invalid-owner"
					: result.reason.name;
			}))];
		assert.deepEqual(failureKinds, [], `round ${round} leaked benign lock-snapshot races`);
		assert.equal(maximumActive, 1, `round ${round} admitted overlapping lock owners`);
		assert.deepEqual(
			settled.map((result) => result.status === "fulfilled" ? result.value : -1).sort((left, right) => left - right),
			Array.from({ length: 32 }, (_, index) => index),
		);
		assert.deepEqual(
			(await readdir(root)).filter((name) => name.startsWith(`req-race-${round}.lock.`)),
			[],
			`round ${round} leaked a lock after bounded completion`,
		);
	}
});

test("a stable malformed live lock still fails closed", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-477-malformed-live-lock-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const token = "44444444-4444-4444-8444-444444444444";
	const lockPath = join(root, `req-477.lock.${token}.active`);
	await writeFile(lockPath, '{"schemaVersion":1,"pid":', { mode: 0o600 });
	const repository = new FileHumanDecisionRepository(root, {
		lockRetryMs: 1,
		lockMaxAttempts: 3,
		lockStaleMs: 60_000,
	});
	await assert.rejects(
		repository.transact("req-477", () => ({ state: null, value: undefined })),
		/lock owner record is invalid/i,
	);
	assert.equal(await readFile(lockPath, "utf8"), '{"schemaVersion":1,"pid":');
});

test("an obsolete lock owner cannot delete a replacement lock during release", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-477-fenced-release-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const replacementToken = "11111111-1111-4111-8111-111111111111";
	const repository = new FileHumanDecisionRepository(root, { lockRetryMs: 1 });
	await assert.rejects(repository.transact("req-477", async () => {
		const locks = (await readdir(root)).filter((name) => /^req-477\.lock\.[0-9a-f-]{36}\.active$/.test(name));
		assert.equal(locks.length, 1);
		const lockPath = join(root, locks[0]);
		await rm(lockPath, { recursive: true, force: true });
		await writeFile(lockPath, JSON.stringify({
			schemaVersion: 1,
			pid: process.pid,
			token: replacementToken,
			createdAt: "2026-07-21T12:30:12Z",
		}), { mode: 0o600, flag: "wx" });
		return { state: null, value: undefined };
	}), /ownership|token|replacement/i);
	const [replacementPath] = (await readdir(root)).filter((name) => name.endsWith(".active"));
	const replacement = JSON.parse(await readFile(join(root, replacementPath), "utf8"));
	assert.equal(replacement.token, replacementToken);
});

test("reclaims a dead-process transaction lock immediately after restart", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-477-orphan-lock-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const token = "00000000-0000-4000-8000-000000000000";
	const lock = join(root, `req-477.lock.${token}.active`);
	await writeFile(lock, JSON.stringify({
		schemaVersion: 1,
		pid: 2_147_483_647,
		token,
		createdAt: "2026-07-21T10:00:00Z",
	}), { mode: 0o600 });
	const repository = new FileHumanDecisionRepository(root, { lockRetryMs: 1, lockMaxAttempts: 3 });
	const persisted = await persistHumanDecisionRequest(repository, spec(), new Date("2026-07-21T10:00:00.000Z"));
	assert.equal(persisted.requestId, "req-477");
});

test("reclaim removes only a dead token path and preserves a live replacement owner", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-477-reclaim-fence-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const deadToken = "22222222-2222-4222-8222-222222222222";
	const liveToken = "33333333-3333-4333-8333-333333333333";
	const deadPath = join(root, `req-477.lock.${deadToken}.active`);
	const livePath = join(root, `req-477.lock.${liveToken}.active`);
	for (const [path, pid, token] of [
		[deadPath, 2_147_483_647, deadToken],
		[livePath, process.pid, liveToken],
	] as const) {
		await writeFile(path, JSON.stringify({
			schemaVersion: 1,
			pid,
			token,
			createdAt: "2026-07-21T10:00:00Z",
		}), { mode: 0o600 });
	}
	const repository = new FileHumanDecisionRepository(root, { lockRetryMs: 1, lockMaxAttempts: 1 });
	await assert.rejects(repository.transact("req-477", () => ({ state: null, value: undefined })), /timed out/i);
	await assert.rejects(readFile(deadPath, "utf8"), /ENOENT/);
	const liveOwner = JSON.parse(await readFile(livePath, "utf8"));
	assert.equal(liveOwner.token, liveToken);
});

test("cycle 6 rejects the shared credential grammar at the native human-decision boundary", async (t) => {
	const samples = [
		"//registry.invalid/:_authToken=SYNTHETIC_NPM_MARKER",
		"machine github.com login maintainer password SYNTHETIC_NETRC_MARKER",
		"aws_secret_access_key = SYNTHETIC_AWS_MARKER",
		"azure_client_secret=SYNTHETIC_AZURE_MARKER",
		"credentials_file = /tmp/SYNTHETIC_CREDENTIAL_FILE",
	];
	for (const question of samples) {
		await t.test(question.split("SYNTHETIC_")[0].trim(), () => assert.throws(
			() => createHumanDecisionRecord(spec({ question })),
			(error: unknown) => error instanceof Error
				&& /credential|secret|sensitive/i.test(error.message)
				&& !error.message.includes("SYNTHETIC_"),
		));
	}
});

test("cycle 7 rejects finite Kubernetes Docker and AWS schemas before persistence or comment rendering", async (t) => {
	const samples = [
		"client-key-data: SYNTHETIC_KUBERNETES_KEY_DATA",
		"token: SYNTHETIC_KUBERNETES_TOKEN",
		'{"auth":"SYNTHETIC_DOCKER_AUTH"}',
		'{"identitytoken":"SYNTHETIC_DOCKER_IDENTITY_TOKEN"}',
		"aws_access_key_id = SYNTHETIC_AWS_ACCESS_KEY_ID",
		"aws_secret_access_key = SYNTHETIC_AWS_SECRET_ACCESS_KEY",
		"aws_session_token = SYNTHETIC_AWS_SESSION_TOKEN",
		"ASIAABCDEFGHIJKLMNOP",
	];
	for (const [index, question] of samples.entries()) {
		await t.test(`schema form ${index + 1}`, () => {
			let rejection: unknown;
			try {
				createHumanDecisionRecord(spec({ question }));
			} catch (error) {
				rejection = error;
			}
			assert.ok(rejection instanceof Error);
			assert.match(rejection.message, /credential|secret|sensitive/i);
			assert.doesNotMatch(rejection.message, /SYNTHETIC_/u);
		});
	}
});

test("cycle 6 closes canonical decision records before reading hostile values", async (t) => {
	const pending = createHumanDecisionRecord(spec(), new Date("2026-07-21T10:00:00.000Z"));
	const requestComment = {
		id: 1001,
		url: "https://github.com/polymetrics-ai/cli/issues/471#issuecomment-1001",
		actor: "shepherd-host",
		createdAt: "2026-07-21T10:00:10.000Z",
	};
	const decision = {
		option: "approve",
		actor: "maintainer-one",
		sourceUrl: "https://github.com/polymetrics-ai/cli/issues/471#issuecomment-1002",
		decidedAt: "2026-07-21T10:01:00.000Z",
	};
	const decided = {
		...pending,
		requestComment,
		status: "decided",
		decision,
		updatedAt: decision.decidedAt,
	};
	const consumed = {
		...decided,
		status: "consumed",
		consumedAt: "2026-07-21T10:02:00.000Z",
		updatedAt: "2026-07-21T10:02:00.000Z",
	};

	await t.test("decided requires request-comment provenance", () => assert.throws(
		() => validateHumanDecisionRecord({ ...decided, requestComment: undefined }),
		/request.?comment|provenance|coherence/i,
	));
	await t.test("decision follows request comment", () => assert.throws(
		() => validateHumanDecisionRecord({
			...decided,
			requestComment: { ...requestComment, createdAt: "2026-07-21T10:01:30.000Z" },
		}),
		/chronology|timestamp|request.?comment/i,
	));
	await t.test("updatedAt covers the decision", () => assert.throws(
		() => validateHumanDecisionRecord({ ...decided, updatedAt: "2026-07-21T10:00:30.000Z" }),
		/chronology|update|decision/i,
	));
	await t.test("updatedAt covers consumption", () => assert.throws(
		() => validateHumanDecisionRecord({ ...consumed, updatedAt: "2026-07-21T10:01:30.000Z" }),
		/chronology|update|consum/i,
	));

	await t.test("wide records reject before accessors", () => {
		let accessed = false;
		const wide: Record<string, unknown> = {};
		Object.defineProperty(wide, "schemaVersion", {
			enumerable: true,
			get() {
				accessed = true;
				throw new Error("SYNTHETIC_CYCLE6_ACCESSOR_MARKER");
			},
		});
		for (let index = 0; index < 300; index += 1) wide[`field${index}`] = index;
		assert.throws(() => validateHumanDecisionRecord(wide), /bounded|field|shape|record/i);
		assert.equal(accessed, false);
	});

	await t.test("normal and revoked proxies reject without traps or host text", () => {
		let trapped = false;
		const proxied = new Proxy(consumed, {
			ownKeys() {
				trapped = true;
				throw new Error("SYNTHETIC_CYCLE6_PROXY_MARKER");
			},
		});
		assert.throws(() => validateHumanDecisionRecord(proxied), /proxy|shape|record|invalid/i);
		assert.equal(trapped, false);

		const revoked = Proxy.revocable(consumed, {});
		revoked.revoke();
		let rejection: unknown;
		try {
			validateHumanDecisionRecord(revoked.proxy);
		} catch (error) {
			rejection = error;
		}
		assert.ok(rejection instanceof Error);
		assert.match(String(rejection), /proxy|shape|record|invalid/i);
		assert.doesNotMatch(String(rejection), /Cannot perform|revoked/i);
	});
});

test("cycle 7 bounds every canonical decision event to a controller-owned observation clock", async (t) => {
	const observedAt = new Date("2026-07-21T10:05:00.000Z");
	const pending = createHumanDecisionRecord(spec({ expiresAt: "2027-07-22T10:00:00.000Z" }), new Date("2026-07-21T10:00:00.000Z"));
	const requestComment = {
		id: 1001,
		url: "https://github.com/polymetrics-ai/cli/issues/471#issuecomment-1001",
		actor: "shepherd-host",
		createdAt: "2026-07-21T10:00:10.000Z",
	};
	const decision = {
		option: "approve",
		actor: "maintainer-one",
		sourceUrl: "https://github.com/polymetrics-ai/cli/issues/471#issuecomment-1002",
		decidedAt: "2026-07-21T10:01:00.000Z",
	};
	const consumed = {
		...pending,
		requestComment,
		status: "consumed" as const,
		decision,
		consumedAt: "2026-07-21T10:02:00.000Z",
		updatedAt: "2026-07-21T10:02:00.000Z",
	};
	const future = "2026-07-21T10:05:02.000Z";
	const cases: Array<[string, HumanDecisionRecord]> = [
		["creation", { ...pending, createdAt: future, updatedAt: future }],
		["request comment", {
			...pending,
			requestComment: { ...requestComment, createdAt: future },
			updatedAt: future,
		}],
		["decision", {
			...consumed,
			decision: { ...decision, decidedAt: future },
			consumedAt: future,
			updatedAt: future,
		}],
		["consumption", { ...consumed, consumedAt: future, updatedAt: future }],
		["update", { ...consumed, updatedAt: future }],
		["all events", {
			...consumed,
			createdAt: "2026-07-21T10:05:02.000Z",
			requestComment: { ...requestComment, createdAt: "2026-07-21T10:05:03.000Z" },
			decision: { ...decision, decidedAt: "2026-07-21T10:05:04.000Z" },
			consumedAt: "2026-07-21T10:05:05.000Z",
			updatedAt: "2026-07-21T10:05:05.000Z",
		}],
	];
	const validateAt = validateHumanDecisionRecord as unknown as (value: unknown, observedAt: Date) => HumanDecisionRecord;
	for (const [name, record] of cases) {
		await t.test(name, () => assert.throws(
			() => validateAt(record, observedAt),
			/future|observation|clock|chronology|timestamp/i,
		));
	}
});
