import { Buffer } from "node:buffer";

const CONTROL_OR_SEPARATOR = /[\u0000-\u001F\u007F-\u009F\u2028\u2029]/u;
const UNSAFE_TEXT_CONTROL = /[\u0000-\u0008\u000B\u000C\u000E-\u001F\u007F-\u009F\u2028\u2029]/u;
const PEM_ARMOR = /-----BEGIN [A-Z0-9][A-Z0-9 -]{0,80}-----/iu;
const PEM_BLOCK = /-----BEGIN [\s\S]{0,8192}?-----END [^-]{1,80}-----/gu;
const BASE64_TEXT = /^[A-Za-z0-9+/_=-]{32,}$/u;
const BASE64_URL = /^[A-Za-z0-9_-]+$/u;
const SENSITIVE_FIELD_NAME = /(?:^|[-_])(?:approval|authorization|grant|password|private[-_]?key|secret|token)(?:$|[-_])/iu;
const SENSITIVE_ASSIGNMENT = /^([\t ]*(?:approval|authorization|grant|password|private[ _-]?key|secret|token)(?:[ _-][A-Za-z0-9_.-]+)?[\t ]*[:=][\t ]*)[^\r\n]*$/gimu;
const SAFE_METADATA_FIELD_NAMES = new Set(["secret_material"]);
const TOKEN_PATTERNS = Object.freeze([
  /gh[pousr]_[A-Za-z0-9_-]+/iu,
  /github_pat_[A-Za-z0-9_-]+/iu,
  /\b(?:bearer|token)\s+[A-Za-z0-9._~+\/-]{12,}\b/iu,
]);
const TOKEN_REDACTION_PATTERNS = Object.freeze([
  /gh[pousr]_[A-Za-z0-9_-]+/giu,
  /github_pat_[A-Za-z0-9_-]+/giu,
  /\b(?:bearer|token)\s+[A-Za-z0-9._~+\/-]{12,}\b/giu,
]);

function isPlainObject(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function decodeBase64(value) {
  if (!BASE64_TEXT.test(value)) return null;
  try {
    return Buffer.from(value.replaceAll("-", "+").replaceAll("_", "/"), "base64").toString("utf8");
  } catch {
    return null;
  }
}

function hasEncodedPrivateKeyArmor(value) {
  const decoded = decodeBase64(value);
  return decoded !== null && PEM_ARMOR.test(decoded);
}

function parseCompactJOSEHeader(segment) {
  if (!BASE64_URL.test(segment)) return null;
  try {
    const padding = "=".repeat((4 - (segment.length % 4)) % 4);
    const decoded = Buffer.from(`${segment}${padding}`, "base64url").toString("utf8");
    const parsed = JSON.parse(decoded);
    return isPlainObject(parsed) ? parsed : null;
  } catch {
    return null;
  }
}

function isCompactJOSE(value) {
  const parts = value.split(".");
  if (parts.length !== 3 && parts.length !== 5) {
    return false;
  }
  if (!BASE64_URL.test(parts[0]) ||
      parts.slice(1).some((part) => part !== "" && !BASE64_URL.test(part))) {
    return false;
  }
  const header = parseCompactJOSEHeader(parts[0]);
  return typeof header?.alg === "string" && header.alg.trim() !== "";
}

export function containsUnsafePersistedText(value) {
  const text = String(value ?? "");
  return CONTROL_OR_SEPARATOR.test(text) ||
    PEM_ARMOR.test(text) ||
    hasEncodedPrivateKeyArmor(text) ||
    isCompactJOSE(text) ||
    TOKEN_PATTERNS.some((pattern) => pattern.test(text));
}

export function assertSafePersistedScalar(value, label = "persisted scalar") {
  if (typeof value !== "string") {
    throw new Error(`${label} must be a string`);
  }
  if (containsUnsafePersistedText(value)) {
    throw new Error(`${label} contains unsafe credential-like material`);
  }
  return value;
}

export function assertPersistedArtifactSafe(value, label = "persisted artifact") {
  const visit = (candidate) => {
    if (typeof candidate === "string") {
      assertSafePersistedScalar(candidate, label);
      return;
    }
    if (candidate === undefined || candidate === null || typeof candidate === "boolean") return;
    if (typeof candidate === "number") {
      if (!Number.isFinite(candidate)) throw new Error(`${label} contains an invalid scalar`);
      return;
    }
    if (Array.isArray(candidate)) {
      candidate.forEach(visit);
      return;
    }
    if (!isPlainObject(candidate)) {
      throw new Error(`${label} contains an unsupported value`);
    }
    for (const [key, nested] of Object.entries(candidate)) {
      assertSafePersistedScalar(key, `${label} field name`);
      if (SENSITIVE_FIELD_NAME.test(key) && !SAFE_METADATA_FIELD_NAMES.has(key)) {
        throw new Error(`${label} contains a forbidden sensitive field`);
      }
      visit(nested);
    }
  };
  visit(value);
  return value;
}

export function stableJSONString(value) {
  if (Array.isArray(value)) return `[${value.map((item) => stableJSONString(item)).join(",")}]`;
  if (isPlainObject(value)) {
    return `{${Object.keys(value).sort().map((key) => `${JSON.stringify(key)}:${stableJSONString(value[key])}`).join(",")}}`;
  }
  return JSON.stringify(value);
}

export function redactPersistedText(value) {
	let text = String(value ?? "").replace(PEM_BLOCK, "<redacted>");
	for (const pattern of TOKEN_REDACTION_PATTERNS) {
		text = text.replace(pattern, "<redacted>");
	}
	text = text.replace(SENSITIVE_ASSIGNMENT, "$1<redacted>");
	text = text.replace(/[A-Za-z0-9+/_=-]{32,}/gu, (candidate) =>
		hasEncodedPrivateKeyArmor(candidate) ? "<redacted>" : candidate,
	);
	text = text.replace(/[A-Za-z0-9_-]+(?:\.[A-Za-z0-9_-]+){2,4}/gu, (candidate) =>
		isCompactJOSE(candidate) ? "<redacted>" : candidate,
	);
	if (UNSAFE_TEXT_CONTROL.test(text) ||
		PEM_ARMOR.test(text) ||
		hasEncodedPrivateKeyArmor(text) ||
		isCompactJOSE(text) ||
		TOKEN_PATTERNS.some((pattern) => pattern.test(text))) {
		throw new Error("refusing to emit text that still contains unsafe credential-like material");
	}
	return text;
}
