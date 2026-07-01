FROM node:22-alpine AS deps

WORKDIR /app
COPY package*.json ./
RUN npm ci

# ── Build ────────────────────────────────────────────────────────────────────

FROM node:22-alpine AS build

WORKDIR /app

# Baked into routes-manifest.json at build time — sets the proxy destination
# for Next.js rewrites. Docker passes http://api:8080; local dev falls back to
# http://localhost:8080 via the default in next.config.ts.
ARG BACKEND_URL=http://localhost:8080
ENV BACKEND_URL=$BACKEND_URL

COPY --from=deps /app/node_modules ./node_modules
# .dockerignore excludes node_modules and .next from the build context, so
# COPY . . only sends source files — no Windows binaries to trigger SIGBUS.
COPY . .
RUN npm run build

# ── Runner ───────────────────────────────────────────────────────────────────
# next.config.ts output:"standalone" emits a self-contained server under
# .next/standalone that needs no node_modules — image drops from ~1 GB to
# ~200 MB.

FROM node:22-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

COPY --from=build /app/public ./public
COPY --from=build /app/.next/standalone ./
COPY --from=build /app/.next/static ./.next/static

EXPOSE 3000

CMD ["node", "server.js"]
