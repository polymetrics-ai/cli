/**
 * In-memory sliding-window rate limiter. The site runs as a single
 * replica (one Node process), so process-local state is authoritative.
 */
type Window = { limit: number; windowMs: number };

const WINDOWS: Record<string, Window[]> = {
  'comment:create': [
    { limit: 5, windowMs: 60_000 },
    { limit: 100, windowMs: 86_400_000 },
  ],
  'bookmark:create': [{ limit: 30, windowMs: 60_000 }],
};

const hits = new Map<string, number[]>();

export type RateResult = { allowed: true } | { allowed: false; retryAfterSeconds: number };

export function checkRateLimit(action: keyof typeof WINDOWS, userId: string, now = Date.now()): RateResult {
  const windows = WINDOWS[action];
  const key = `${action}:${userId}`;
  const maxWindow = Math.max(...windows.map((w) => w.windowMs));
  const timestamps = (hits.get(key) ?? []).filter((t) => now - t < maxWindow);

  for (const { limit, windowMs } of windows) {
    const inWindow = timestamps.filter((t) => now - t < windowMs);
    if (inWindow.length >= limit) {
      const oldest = Math.min(...inWindow);
      return { allowed: false, retryAfterSeconds: Math.max(1, Math.ceil((oldest + windowMs - now) / 1000)) };
    }
  }

  timestamps.push(now);
  hits.set(key, timestamps);
  return { allowed: true };
}

/** Test hook. */
export function resetRateLimits(): void {
  hits.clear();
}
