# syntax=docker/dockerfile:1

ARG PNPM_VERSION=11.7.0

# ── deps ──────────────────────────────────────────────────────────────────────
FROM node:26-alpine AS deps
ARG PNPM_VERSION
RUN npm install -g "pnpm@${PNPM_VERSION}"
WORKDIR /app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --ignore-scripts --reporter=silent

# ── builder ───────────────────────────────────────────────────────────────────
FROM node:26-alpine AS builder
ARG PNPM_VERSION
RUN npm install -g "pnpm@${PNPM_VERSION}"
WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY . .

ENV NEXT_TELEMETRY_DISABLED=1
RUN pnpm build

# ── runner ────────────────────────────────────────────────────────────────────
FROM node:26-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

RUN addgroup --system --gid 1001 nodejs && \
    adduser  --system --uid 1001 nextjs

COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

CMD ["node", "server.js"]
