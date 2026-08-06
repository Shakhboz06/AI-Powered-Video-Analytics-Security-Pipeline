FROM node:22-alpine AS builder
WORKDIR /app
COPY nuxt-dashboard/package.json nuxt-dashboard/package-lock.json ./
RUN npm ci
COPY nuxt-dashboard/ ./
RUN npm run build

FROM node:22-alpine
WORKDIR /app
COPY --from=builder /app/.output ./.output
ENV PORT=3000
ENV HOST=0.0.0.0
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
