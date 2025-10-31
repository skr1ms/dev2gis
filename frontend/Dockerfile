# syntax=docker/dockerfile:1.7

ARG NODE_MAJOR=24

############################
# STAGE 1: deps
############################
FROM node:${NODE_MAJOR}-alpine AS deps
WORKDIR /app
RUN corepack enable || true
RUN apk add --no-cache python3 make g++ bash
COPY package.json package-lock.json* ./
RUN --mount=type=cache,target=/root/.npm npm install --no-audit --no-fund

############################
# STAGE 2: build
############################
FROM node:${NODE_MAJOR}-alpine AS build
WORKDIR /app
ENV NODE_ENV=production

ARG NEXT_PUBLIC_API_URL
ARG NEXT_PUBLIC_ENVIRONMENT
ARG NEXT_TELEMETRY_DISABLED=1
ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL}
ENV NEXT_PUBLIC_ENVIRONMENT=${NEXT_PUBLIC_ENVIRONMENT}
ENV NEXT_TELEMETRY_DISABLED=${NEXT_TELEMETRY_DISABLED}

COPY --from=deps /app/node_modules ./node_modules
COPY package.json package-lock.json* ./
COPY tsconfig.json postcss.config.js tailwind.config.js next.config.js .eslintrc.json ./
COPY src ./src
COPY public ./public

RUN --mount=type=cache,target=/root/.npm npm run build

############################
# STAGE 3: prune (prod deps)
############################
FROM node:${NODE_MAJOR}-alpine AS prune
WORKDIR /app
ENV NODE_ENV=production
COPY package.json package-lock.json* ./
COPY --from=deps /app/node_modules ./node_modules
RUN --mount=type=cache,target=/root/.npm npm prune --omit=dev --no-audit --no-fund

############################
# STAGE 4: runner (final runtime)
############################
FROM node:${NODE_MAJOR}-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup -S nodejs && adduser -S nextjs -G nodejs
RUN apk add --no-cache dumb-init wget && rm -rf /var/cache/apk/*

COPY --chown=nextjs:nodejs --from=prune /app/node_modules ./node_modules
COPY --chown=nextjs:nodejs --from=build /app/.next/standalone ./
COPY --chown=nextjs:nodejs --from=build /app/.next/static ./.next/static
COPY --chown=nextjs:nodejs --from=build /app/public ./public
COPY --chown=nextjs:nodejs package.json ./

ENV PORT=3000
ENV HOSTNAME=0.0.0.0
EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:3000/health || exit 1

USER nextjs
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["node", "server.js"]

############################
# STAGE 5: dev (hot reload)
############################
FROM node:${NODE_MAJOR}-alpine AS dev
WORKDIR /app
ENV NODE_ENV=development
RUN addgroup -g 101 -S nodejs && adduser -u 100 -S nextjs -G nodejs
RUN chown -R nextjs:nodejs /app
COPY --chown=nextjs:nodejs package.json package-lock.json* ./
USER nextjs
RUN --mount=type=cache,target=/home/nextjs/.npm,uid=100,gid=101 npm install --no-audit --no-fund
COPY --chown=nextjs:nodejs . .
EXPOSE 3000
ENV PORT=3000
CMD ["npm", "run", "dev"]

FROM runner AS production

############################
# STAGE 6: dev-nginx (dev with nginx)
############################
FROM nginx:alpine AS dev-nginx

RUN apk add --no-cache nodejs npm

WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY package.json package-lock.json* ./
COPY . .

ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL}
ENV PORT=3000
ENV HOSTNAME=0.0.0.0
ENV NODE_ENV=production

RUN npm run build

COPY --from=build /app/.next/standalone /app/
COPY --from=build /app/.next/static /app/.next/static
COPY --from=build /app/public /app/public

EXPOSE 80

CMD sh -c "node server.js & sleep 3 && nginx -g 'daemon off;'"