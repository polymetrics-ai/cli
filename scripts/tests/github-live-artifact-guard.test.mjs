import assert from "node:assert/strict";
import test from "node:test";

import { redactPersistedText } from "../github-live-artifact-guard.mjs";

test("redacts sensitive assignment values before they can reach a persisted report", () => {
  const redacted = redactPersistedText("password = synthetic-test-value\nGITHUB_TOKEN: synthetic-test-value");

  assert.equal(redacted.includes("synthetic-test-value"), false);
  assert.match(redacted, /password\s*=\s*<redacted>/i);
  assert.match(redacted, /GITHUB_TOKEN:\s*<redacted>/i);
});

test("fails closed instead of returning text that cannot be made safe", () => {
  assert.throws(
    () => redactPersistedText("token\u0000= synthetic-test-value"),
    /refusing to emit text that still contains unsafe credential-like material/,
  );
});
